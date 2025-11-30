# 缓存一致性 (Cache Consistency)

## 概述

缓存一致性是指确保缓存中的数据与数据库中的数据保持同步。由于缓存和数据库是两个独立的存储系统，在更新数据时可能出现不一致。

```
┌─────────────────────────────────────────────────────────────┐
│                      一致性问题场景                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  时刻 T1: 缓存 = "Alice", 数据库 = "Alice"  ✓ 一致          │
│                                                             │
│  时刻 T2: 用户更新名字为 "Bob"                              │
│                                                             │
│  时刻 T3: 数据库 = "Bob", 缓存 = "Alice"   ✗ 不一致！       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 一致性策略对比

| 策略 | 一致性 | 性能 | 复杂度 | 适用场景 |
|------|--------|------|--------|----------|
| 先更新数据库，再删除缓存 | 中 | 高 | 低 | 通用场景 |
| 先删除缓存，再更新数据库 | 低 | 高 | 低 | 读少写多 |
| 延迟双删 | 中高 | 中 | 中 | 读多写多 |
| 订阅 Binlog | 高 | 中 | 高 | 强一致性要求 |
| 分布式事务 | 最高 | 低 | 最高 | 金融级场景 |

## 策略一：先更新数据库，再删除缓存

最常用的策略，也叫 Cache-Aside 写入模式。

### 基本实现

```go
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    // 1. 更新数据库
    if err := s.db.Save(user).Error; err != nil {
        return err
    }

    // 2. 删除缓存
    key := fmt.Sprintf("user:%d", user.ID)
    if err := s.rdb.Del(ctx, key).Err(); err != nil {
        // 记录日志，但不影响主流程
        log.Printf("Failed to delete cache: %v", err)
    }

    return nil
}
```

### 问题：并发读写

```
线程 A (读)                    线程 B (写)
   │                              │
   │ 1. 缓存未命中                │
   │                              │ 2. 更新数据库 (v2)
   │                              │ 3. 删除缓存
   │ 4. 读数据库 (v2)             │
   │ 5. 写入缓存 (v2)             │
   │                              │
   ✓ 正常情况下没有问题
```

但极端情况：

```
线程 A (读)                    线程 B (写)
   │                              │
   │ 1. 缓存未命中                │
   │ 2. 读数据库 (v1)             │
   │                              │ 3. 更新数据库 (v2)
   │                              │ 4. 删除缓存
   │ 5. 写入缓存 (v1)             │  ← 旧数据写入缓存！
   │                              │
   ✗ 缓存中是旧数据
```

### 解决：设置较短过期时间

```go
func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 查缓存
    val, err := s.rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(val), &user)
        return &user, nil
    }

    // 查数据库
    var user User
    if err := s.db.First(&user, id).Error; err != nil {
        return nil, err
    }

    // 写入缓存，设置较短过期时间
    data, _ := json.Marshal(user)
    s.rdb.Set(ctx, key, data, 5*time.Minute)  // 5 分钟过期

    return &user, nil
}
```

## 策略二：先删除缓存，再更新数据库

```go
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    key := fmt.Sprintf("user:%d", user.ID)

    // 1. 先删除缓存
    s.rdb.Del(ctx, key)

    // 2. 再更新数据库
    return s.db.Save(user).Error
}
```

### 问题：并发读写导致脏数据

```
线程 A (写)                    线程 B (读)
   │                              │
   │ 1. 删除缓存                  │
   │                              │ 2. 缓存未命中
   │                              │ 3. 读数据库 (v1)
   │                              │ 4. 写入缓存 (v1)
   │ 5. 更新数据库 (v2)           │
   │                              │
   ✗ 缓存中是旧数据 v1，数据库是新数据 v2
```

这种策略在高并发下容易出问题，**不推荐单独使用**。

## 策略三：延迟双删

先删除缓存，更新数据库后，延迟一段时间再删除一次缓存。

### 实现

```go
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    key := fmt.Sprintf("user:%d", user.ID)

    // 1. 先删除缓存
    s.rdb.Del(ctx, key)

    // 2. 更新数据库
    if err := s.db.Save(user).Error; err != nil {
        return err
    }

    // 3. 延迟再删除一次缓存
    go func() {
        time.Sleep(500 * time.Millisecond)  // 延迟 500ms
        s.rdb.Del(context.Background(), key)
    }()

    return nil
}
```

### 使用消息队列实现可靠延迟删除

```go
type CacheInvalidation struct {
    Key       string    `json:"key"`
    DeleteAt  time.Time `json:"delete_at"`
}

func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    key := fmt.Sprintf("user:%d", user.ID)

    // 1. 先删除缓存
    s.rdb.Del(ctx, key)

    // 2. 更新数据库
    if err := s.db.Save(user).Error; err != nil {
        return err
    }

    // 3. 发送延迟删除消息到队列
    msg := CacheInvalidation{
        Key:      key,
        DeleteAt: time.Now().Add(500 * time.Millisecond),
    }
    s.mq.Publish("cache:invalidate", msg)

    return nil
}

// 消费者处理延迟删除
func (s *UserService) ProcessCacheInvalidation() {
    for msg := range s.mq.Subscribe("cache:invalidate") {
        var inv CacheInvalidation
        json.Unmarshal(msg, &inv)

        // 等待到指定时间
        delay := time.Until(inv.DeleteAt)
        if delay > 0 {
            time.Sleep(delay)
        }

        // 删除缓存
        s.rdb.Del(context.Background(), inv.Key)
    }
}
```

### 延迟时间计算

```
延迟时间 > 读操作耗时 + 写缓存耗时

假设：
- 读数据库：50ms
- 写缓存：5ms

延迟时间 > 50 + 5 = 55ms

建议设置 100-500ms 的延迟
```

## 策略四：订阅数据库 Binlog

通过订阅 MySQL Binlog，在数据库数据变更时自动更新缓存。

### 架构

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌─────────┐
│  应用    │────▶│ Database │────▶│  Canal   │────▶│  Redis  │
│          │     │ (MySQL)  │     │ (Binlog) │     │         │
└──────────┘     └──────────┘     └──────────┘     └─────────┘
```

### 使用 Canal 订阅

```go
// Canal 消息处理
type CanalMessage struct {
    Database string            `json:"database"`
    Table    string            `json:"table"`
    Type     string            `json:"type"` // INSERT, UPDATE, DELETE
    Data     []json.RawMessage `json:"data"`
    Old      []json.RawMessage `json:"old"`
}

type BinlogProcessor struct {
    rdb *redis.Client
}

func (p *BinlogProcessor) Process(msg *CanalMessage) {
    switch msg.Table {
    case "users":
        p.processUserChange(msg)
    case "products":
        p.processProductChange(msg)
    }
}

func (p *BinlogProcessor) processUserChange(msg *CanalMessage) {
    ctx := context.Background()

    for _, data := range msg.Data {
        var user struct {
            ID uint `json:"id"`
        }
        json.Unmarshal(data, &user)

        key := fmt.Sprintf("user:%d", user.ID)

        switch msg.Type {
        case "INSERT", "UPDATE":
            // 删除缓存，让下次访问重新加载
            p.rdb.Del(ctx, key)
        case "DELETE":
            p.rdb.Del(ctx, key)
        }

        log.Printf("Processed %s on user:%d", msg.Type, user.ID)
    }
}
```

### 优缺点

| 优点 | 缺点 |
|------|------|
| 应用代码简单 | 需要部署 Canal |
| 强一致性 | 增加系统复杂度 |
| 解耦应用和缓存 | 有一定延迟 |

## 策略五：分布式事务

使用分布式事务确保数据库和缓存的原子性更新。

### 基于 TCC 的实现

```go
type TCCTransaction struct {
    db  *gorm.DB
    rdb *redis.Client
}

type UserUpdateTCC struct {
    UserID   uint
    OldData  *User
    NewData  *User
    CacheKey string
}

// Try 阶段：预留资源
func (t *TCCTransaction) Try(ctx context.Context, tcc *UserUpdateTCC) error {
    // 1. 加锁
    lockKey := fmt.Sprintf("lock:%s", tcc.CacheKey)
    ok, err := t.rdb.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
    if err != nil || !ok {
        return errors.New("failed to acquire lock")
    }

    // 2. 记录旧数据
    var user User
    if err := t.db.First(&user, tcc.UserID).Error; err != nil {
        return err
    }
    tcc.OldData = &user

    return nil
}

// Confirm 阶段：确认执行
func (t *TCCTransaction) Confirm(ctx context.Context, tcc *UserUpdateTCC) error {
    // 1. 更新数据库
    if err := t.db.Save(tcc.NewData).Error; err != nil {
        return err
    }

    // 2. 删除缓存
    t.rdb.Del(ctx, tcc.CacheKey)

    // 3. 释放锁
    lockKey := fmt.Sprintf("lock:%s", tcc.CacheKey)
    t.rdb.Del(ctx, lockKey)

    return nil
}

// Cancel 阶段：取消回滚
func (t *TCCTransaction) Cancel(ctx context.Context, tcc *UserUpdateTCC) error {
    // 1. 恢复数据库（如果已修改）
    if tcc.OldData != nil {
        t.db.Save(tcc.OldData)
    }

    // 2. 释放锁
    lockKey := fmt.Sprintf("lock:%s", tcc.CacheKey)
    t.rdb.Del(ctx, lockKey)

    return nil
}

// 使用
func (s *UserService) UpdateUserTCC(ctx context.Context, user *User) error {
    tcc := &UserUpdateTCC{
        UserID:   user.ID,
        NewData:  user,
        CacheKey: fmt.Sprintf("user:%d", user.ID),
    }

    tx := &TCCTransaction{db: s.db, rdb: s.rdb}

    // Try
    if err := tx.Try(ctx, tcc); err != nil {
        return err
    }

    // Confirm
    if err := tx.Confirm(ctx, tcc); err != nil {
        // Cancel
        tx.Cancel(ctx, tcc)
        return err
    }

    return nil
}
```

## 实际应用

### 综合方案：延迟双删 + 重试

```go
type CacheConsistencyManager struct {
    db          *gorm.DB
    rdb         *redis.Client
    retryQueue  chan RetryTask
    maxRetries  int
}

type RetryTask struct {
    Key      string
    Attempts int
}

func NewCacheConsistencyManager(db *gorm.DB, rdb *redis.Client) *CacheConsistencyManager {
    m := &CacheConsistencyManager{
        db:         db,
        rdb:        rdb,
        retryQueue: make(chan RetryTask, 10000),
        maxRetries: 3,
    }

    // 启动重试处理
    go m.processRetries()

    return m
}

func (m *CacheConsistencyManager) UpdateWithConsistency(ctx context.Context, key string, updateFn func() error) error {
    // 1. 第一次删除缓存
    m.rdb.Del(ctx, key)

    // 2. 更新数据库
    if err := updateFn(); err != nil {
        return err
    }

    // 3. 延迟第二次删除
    go func() {
        time.Sleep(500 * time.Millisecond)
        if err := m.rdb.Del(context.Background(), key).Err(); err != nil {
            // 加入重试队列
            m.retryQueue <- RetryTask{Key: key, Attempts: 0}
        }
    }()

    return nil
}

func (m *CacheConsistencyManager) processRetries() {
    for task := range m.retryQueue {
        if task.Attempts >= m.maxRetries {
            log.Printf("Max retries exceeded for key: %s", task.Key)
            continue
        }

        // 指数退避
        delay := time.Duration(1<<task.Attempts) * 100 * time.Millisecond
        time.Sleep(delay)

        if err := m.rdb.Del(context.Background(), task.Key).Err(); err != nil {
            task.Attempts++
            m.retryQueue <- task
        }
    }
}

// 使用
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    key := fmt.Sprintf("user:%d", user.ID)

    return s.consistency.UpdateWithConsistency(ctx, key, func() error {
        return s.db.Save(user).Error
    })
}
```

### 批量操作的一致性

```go
func (s *UserService) BatchUpdate(ctx context.Context, users []*User) error {
    keys := make([]string, len(users))
    for i, u := range users {
        keys[i] = fmt.Sprintf("user:%d", u.ID)
    }

    // 1. 批量删除缓存
    s.rdb.Del(ctx, keys...)

    // 2. 批量更新数据库（事务）
    err := s.db.Transaction(func(tx *gorm.DB) error {
        for _, user := range users {
            if err := tx.Save(user).Error; err != nil {
                return err
            }
        }
        return nil
    })

    if err != nil {
        return err
    }

    // 3. 延迟再次批量删除
    go func() {
        time.Sleep(500 * time.Millisecond)
        s.rdb.Del(context.Background(), keys...)
    }()

    return nil
}
```

## 监控与告警

```go
type ConsistencyMonitor struct {
    inconsistentCount int64
    checkCount        int64
}

// 定期一致性检查
func (m *ConsistencyMonitor) Check(ctx context.Context, db *gorm.DB, rdb *redis.Client) {
    // 随机抽样检查
    var users []User
    db.Order("RANDOM()").Limit(100).Find(&users)

    for _, user := range users {
        atomic.AddInt64(&m.checkCount, 1)

        key := fmt.Sprintf("user:%d", user.ID)
        val, err := rdb.Get(ctx, key).Result()
        if err == redis.Nil {
            continue  // 缓存不存在是正常的
        }
        if err != nil {
            continue
        }

        var cachedUser User
        if err := json.Unmarshal([]byte(val), &cachedUser); err != nil {
            continue
        }

        // 比较关键字段
        if cachedUser.Name != user.Name || cachedUser.Email != user.Email {
            atomic.AddInt64(&m.inconsistentCount, 1)
            log.Printf("Inconsistency detected for user:%d", user.ID)

            // 删除不一致的缓存
            rdb.Del(ctx, key)
        }
    }

    // 上报指标
    inconsistentRate := float64(m.inconsistentCount) / float64(m.checkCount) * 100
    log.Printf("Consistency check: %.2f%% inconsistent", inconsistentRate)
}
```

## 策略选择建议

```
┌─────────────────────────────────────────────────────────────┐
│                      选择一致性策略                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  对一致性要求？                                              │
│     │                                                       │
│     ├── 最终一致即可 ──▶ 先更新DB再删缓存 + 短TTL            │
│     │                                                       │
│     ├── 较高 ──▶ 延迟双删                                   │
│     │                                                       │
│     ├── 很高 ──▶ 订阅 Binlog                                │
│     │                                                       │
│     └── 强一致 ──▶ 分布式事务 (TCC)                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 优先删除而非更新 | 删除缓存比更新更简单可靠 |
| 设置合理 TTL | 作为最后的一致性保障 |
| 实现重试机制 | 缓存删除失败时重试 |
| 定期一致性检查 | 监控并修复不一致数据 |
| 使用分布式锁 | 关键操作加锁防并发 |
| 记录操作日志 | 便于问题排查 |

**下一节**：[练习](./exercises.md) - 缓存实践练习
