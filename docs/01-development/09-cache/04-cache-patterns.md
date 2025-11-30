# 缓存模式 (Cache Patterns)

## 概述

缓存模式定义了应用程序如何与缓存和数据库交互，选择合适的模式对性能和数据一致性至关重要。

## Cache-Aside (旁路缓存)

最常用的缓存模式，应用程序直接管理缓存。

### 读取流程

```
┌─────────┐     ┌─────────┐     ┌──────────┐
│ Client  │────▶│  Cache  │     │ Database │
└─────────┘     └────┬────┘     └────┬─────┘
                     │               │
              1. 查询缓存            │
                     │               │
              ┌──────┴──────┐        │
              │             │        │
         命中(Hit)     未命中(Miss)  │
              │             │        │
              ▼             ▼        │
         返回数据    2. 查询数据库 ──┘
                           │
                    3. 写入缓存
                           │
                    4. 返回数据
```

### 实现

```go
type CacheAside struct {
    rdb *redis.Client
    db  *gorm.DB
}

// 读取
func (c *CacheAside) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 查询缓存
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            return &user, nil  // 缓存命中
        }
    }

    // 2. 缓存未命中，查询数据库
    var user User
    if err := c.db.First(&user, id).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存
    data, _ := json.Marshal(user)
    c.rdb.Set(ctx, key, data, time.Hour)

    return &user, nil
}

// 更新
func (c *CacheAside) UpdateUser(ctx context.Context, user *User) error {
    // 1. 更新数据库
    if err := c.db.Save(user).Error; err != nil {
        return err
    }

    // 2. 删除缓存（而非更新）
    key := fmt.Sprintf("user:%d", user.ID)
    c.rdb.Del(ctx, key)

    return nil
}

// 删除
func (c *CacheAside) DeleteUser(ctx context.Context, id uint) error {
    // 1. 删除数据库
    if err := c.db.Delete(&User{}, id).Error; err != nil {
        return err
    }

    // 2. 删除缓存
    key := fmt.Sprintf("user:%d", id)
    c.rdb.Del(ctx, key)

    return nil
}
```

### 与 JavaScript 对比

```javascript
// Node.js Cache-Aside
class CacheAside {
    constructor(redis, db) {
        this.redis = redis;
        this.db = db;
    }

    async getUser(id) {
        const key = `user:${id}`;

        // 1. 查询缓存
        const cached = await this.redis.get(key);
        if (cached) {
            return JSON.parse(cached);
        }

        // 2. 查询数据库
        const user = await this.db.findUnique({ where: { id } });
        if (!user) return null;

        // 3. 写入缓存
        await this.redis.set(key, JSON.stringify(user), 'EX', 3600);

        return user;
    }

    async updateUser(user) {
        // 1. 更新数据库
        await this.db.update({ where: { id: user.id }, data: user });

        // 2. 删除缓存
        await this.redis.del(`user:${user.id}`);
    }
}
```

### 优缺点

| 优点 | 缺点 |
|------|------|
| 简单直观 | 首次访问必然 Miss |
| 只缓存热点数据 | 需要处理缓存失效 |
| 缓存故障不影响业务 | 数据可能短暂不一致 |

## Read-Through (读穿透)

缓存层负责从数据库加载数据，应用只与缓存交互。

### 读取流程

```
┌─────────┐     ┌─────────────────────────────┐
│ Client  │────▶│           Cache             │
└─────────┘     │  ┌─────────┐  ┌──────────┐ │
                │  │  读取   │──▶│ Database │ │
                │  └─────────┘  └──────────┘ │
                └─────────────────────────────┘
```

### 实现

```go
// ReadThroughCache 抽象缓存层，自动处理数据加载
type ReadThroughCache struct {
    rdb    *redis.Client
    loader func(ctx context.Context, key string) (interface{}, error)
    ttl    time.Duration
}

func NewReadThroughCache(rdb *redis.Client, loader func(ctx context.Context, key string) (interface{}, error), ttl time.Duration) *ReadThroughCache {
    return &ReadThroughCache{
        rdb:    rdb,
        loader: loader,
        ttl:    ttl,
    }
}

func (c *ReadThroughCache) Get(ctx context.Context, key string, dest interface{}) error {
    // 1. 尝试从缓存获取
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        return json.Unmarshal([]byte(val), dest)
    }

    if err != redis.Nil {
        return err
    }

    // 2. 缓存未命中，通过 loader 加载
    data, err := c.loader(ctx, key)
    if err != nil {
        return err
    }

    // 3. 写入缓存
    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }
    c.rdb.Set(ctx, key, jsonData, c.ttl)

    // 4. 设置到目标
    return json.Unmarshal(jsonData, dest)
}

// 使用示例
func main() {
    ctx := context.Background()

    // 创建 Read-Through 缓存
    cache := NewReadThroughCache(rdb, func(ctx context.Context, key string) (interface{}, error) {
        // 从 key 解析 ID
        var id uint
        fmt.Sscanf(key, "user:%d", &id)

        // 从数据库加载
        var user User
        if err := db.First(&user, id).Error; err != nil {
            return nil, err
        }
        return user, nil
    }, time.Hour)

    // 使用缓存
    var user User
    if err := cache.Get(ctx, "user:1", &user); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User: %+v\n", user)
}
```

### 泛型版本 (Go 1.18+)

```go
type ReadThroughCache[T any] struct {
    rdb    *redis.Client
    loader func(ctx context.Context, key string) (T, error)
    ttl    time.Duration
}

func NewReadThroughCache[T any](
    rdb *redis.Client,
    loader func(ctx context.Context, key string) (T, error),
    ttl time.Duration,
) *ReadThroughCache[T] {
    return &ReadThroughCache[T]{
        rdb:    rdb,
        loader: loader,
        ttl:    ttl,
    }
}

func (c *ReadThroughCache[T]) Get(ctx context.Context, key string) (T, error) {
    var zero T

    // 1. 尝试从缓存获取
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        var result T
        if err := json.Unmarshal([]byte(val), &result); err != nil {
            return zero, err
        }
        return result, nil
    }

    if err != redis.Nil {
        return zero, err
    }

    // 2. 缓存未命中，通过 loader 加载
    data, err := c.loader(ctx, key)
    if err != nil {
        return zero, err
    }

    // 3. 写入缓存
    jsonData, _ := json.Marshal(data)
    c.rdb.Set(ctx, key, jsonData, c.ttl)

    return data, nil
}

// 使用
userCache := NewReadThroughCache[User](rdb, loadUserFromDB, time.Hour)
user, err := userCache.Get(ctx, "user:1")
```

## Write-Through (写穿透)

写入时同时更新缓存和数据库，保证一致性。

### 写入流程

```
┌─────────┐     ┌─────────────────────────────┐
│ Client  │────▶│           Cache             │
└─────────┘     │  ┌─────────┐  ┌──────────┐ │
                │  │  写入   │──▶│ Database │ │
                │  │ (同步)  │  └──────────┘ │
                │  └─────────┘               │
                └─────────────────────────────┘
```

### 实现

```go
type WriteThroughCache struct {
    rdb *redis.Client
    db  *gorm.DB
    ttl time.Duration
}

// 写入 - 同时更新缓存和数据库
func (c *WriteThroughCache) Set(ctx context.Context, key string, value interface{}) error {
    // 开启事务
    tx := c.db.Begin()
    if tx.Error != nil {
        return tx.Error
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 1. 写入数据库
    if err := tx.Save(value).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 2. 写入缓存
    data, err := json.Marshal(value)
    if err != nil {
        tx.Rollback()
        return err
    }

    if err := c.rdb.Set(ctx, key, data, c.ttl).Err(); err != nil {
        tx.Rollback()
        return err
    }

    // 3. 提交事务
    return tx.Commit().Error
}

// 读取
func (c *WriteThroughCache) Get(ctx context.Context, key string, dest interface{}) error {
    // 直接从缓存读取
    val, err := c.rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        return errors.New("key not found")
    }
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

// 使用示例
func (c *WriteThroughCache) SaveUser(ctx context.Context, user *User) error {
    key := fmt.Sprintf("user:%d", user.ID)
    return c.Set(ctx, key, user)
}
```

### 优缺点

| 优点 | 缺点 |
|------|------|
| 数据一致性好 | 写入延迟增加 |
| 读取始终命中 | 写入失败需要回滚 |
| 简化读取逻辑 | 可能缓存不热门数据 |

## Write-Behind (Write-Back，异步写回)

先写缓存，异步批量写入数据库。

### 写入流程

```
┌─────────┐     ┌─────────┐     ┌──────────┐
│ Client  │────▶│  Cache  │- - -│ Database │
└─────────┘     └────┬────┘     └──────────┘
                     │               ▲
              立即返回               │
                                异步写入
                              (批量/延迟)
```

### 实现

```go
type WriteBehindCache struct {
    rdb      *redis.Client
    db       *gorm.DB
    ttl      time.Duration
    buffer   chan writeOp
    stopCh   chan struct{}
}

type writeOp struct {
    key   string
    value interface{}
}

func NewWriteBehindCache(rdb *redis.Client, db *gorm.DB, ttl time.Duration, bufferSize int) *WriteBehindCache {
    c := &WriteBehindCache{
        rdb:    rdb,
        db:     db,
        ttl:    ttl,
        buffer: make(chan writeOp, bufferSize),
        stopCh: make(chan struct{}),
    }

    // 启动后台写入 goroutine
    go c.backgroundWriter()

    return c
}

// 后台写入
func (c *WriteBehindCache) backgroundWriter() {
    batch := make([]writeOp, 0, 100)
    ticker := time.NewTicker(time.Second) // 每秒批量写入
    defer ticker.Stop()

    for {
        select {
        case op := <-c.buffer:
            batch = append(batch, op)
            // 达到批量大小，立即写入
            if len(batch) >= 100 {
                c.flushBatch(batch)
                batch = batch[:0]
            }

        case <-ticker.C:
            // 定时写入
            if len(batch) > 0 {
                c.flushBatch(batch)
                batch = batch[:0]
            }

        case <-c.stopCh:
            // 停止前写入剩余数据
            if len(batch) > 0 {
                c.flushBatch(batch)
            }
            return
        }
    }
}

// 批量写入数据库
func (c *WriteBehindCache) flushBatch(batch []writeOp) {
    ctx := context.Background()

    for _, op := range batch {
        if err := c.db.Save(op.value).Error; err != nil {
            log.Printf("Write-behind failed for key %s: %v", op.key, err)
            // 可以加入重试队列
        }
    }

    log.Printf("Flushed %d items to database", len(batch))
}

// 写入 - 先写缓存，异步写数据库
func (c *WriteBehindCache) Set(ctx context.Context, key string, value interface{}) error {
    // 1. 写入缓存
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    if err := c.rdb.Set(ctx, key, data, c.ttl).Err(); err != nil {
        return err
    }

    // 2. 发送到写入队列
    select {
    case c.buffer <- writeOp{key: key, value: value}:
        return nil
    default:
        // 队列满，同步写入
        return c.db.Save(value).Error
    }
}

// 停止
func (c *WriteBehindCache) Stop() {
    close(c.stopCh)
}
```

### 优缺点

| 优点 | 缺点 |
|------|------|
| 写入延迟极低 | 数据可能丢失 |
| 支持批量写入 | 一致性较差 |
| 减轻数据库压力 | 实现复杂 |

## Refresh-Ahead (预刷新)

在缓存过期前提前刷新，避免缓存 Miss。

### 实现

```go
type RefreshAheadCache struct {
    rdb           *redis.Client
    db            *gorm.DB
    ttl           time.Duration
    refreshBefore time.Duration // 过期前多久刷新
    refreshing    sync.Map      // 正在刷新的 key
}

func (c *RefreshAheadCache) Get(ctx context.Context, key string) (*User, error) {
    // 1. 获取值和 TTL
    pipe := c.rdb.Pipeline()
    getCmd := pipe.Get(ctx, key)
    ttlCmd := pipe.TTL(ctx, key)
    _, err := pipe.Exec(ctx)

    if err != nil && err != redis.Nil {
        return nil, err
    }

    val, err := getCmd.Result()
    if err == redis.Nil {
        // 缓存未命中，同步加载
        return c.loadAndCache(ctx, key)
    }

    ttl, _ := ttlCmd.Result()

    // 2. 检查是否需要提前刷新
    if ttl > 0 && ttl < c.refreshBefore {
        // 异步刷新
        go c.refreshAsync(context.Background(), key)
    }

    // 3. 返回当前值
    var user User
    if err := json.Unmarshal([]byte(val), &user); err != nil {
        return nil, err
    }
    return &user, nil
}

// 异步刷新
func (c *RefreshAheadCache) refreshAsync(ctx context.Context, key string) {
    // 防止重复刷新
    if _, loaded := c.refreshing.LoadOrStore(key, true); loaded {
        return
    }
    defer c.refreshing.Delete(key)

    // 重新加载数据
    c.loadAndCache(ctx, key)
}

// 加载并缓存
func (c *RefreshAheadCache) loadAndCache(ctx context.Context, key string) (*User, error) {
    var id uint
    fmt.Sscanf(key, "user:%d", &id)

    var user User
    if err := c.db.First(&user, id).Error; err != nil {
        return nil, err
    }

    data, _ := json.Marshal(user)
    c.rdb.Set(ctx, key, data, c.ttl)

    return &user, nil
}
```

## 模式对比

| 模式 | 读性能 | 写性能 | 一致性 | 复杂度 | 适用场景 |
|------|--------|--------|--------|--------|----------|
| Cache-Aside | 中 | 高 | 中 | 低 | 通用场景 |
| Read-Through | 高 | 高 | 中 | 中 | 读多写少 |
| Write-Through | 高 | 低 | 高 | 中 | 一致性要求高 |
| Write-Behind | 高 | 极高 | 低 | 高 | 写多读少 |
| Refresh-Ahead | 极高 | 中 | 中 | 高 | 热点数据 |

## 组合使用

### Cache-Aside + Write-Behind

```go
type HybridCache struct {
    rdb         *redis.Client
    db          *gorm.DB
    ttl         time.Duration
    writeBuffer chan interface{}
}

// 读取使用 Cache-Aside
func (c *HybridCache) Get(ctx context.Context, key string) (*User, error) {
    // Cache-Aside 逻辑...
}

// 写入使用 Write-Behind
func (c *HybridCache) Set(ctx context.Context, key string, value *User) error {
    // 1. 立即写入缓存
    data, _ := json.Marshal(value)
    c.rdb.Set(ctx, key, data, c.ttl)

    // 2. 异步写入数据库
    c.writeBuffer <- value

    return nil
}
```

## 选择建议

```
┌─────────────────────────────────────────────────────────────┐
│                      选择缓存模式                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  是否允许数据短暂不一致？                                    │
│     │                                                       │
│     ├── 否 ──▶ Write-Through                               │
│     │                                                       │
│     └── 是 ──▶ 写入量大吗？                                 │
│                   │                                         │
│                   ├── 是 ──▶ Write-Behind                  │
│                   │                                         │
│                   └── 否 ──▶ 热点数据吗？                   │
│                                 │                           │
│                                 ├── 是 ──▶ Refresh-Ahead   │
│                                 │                           │
│                                 └── 否 ──▶ Cache-Aside     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 实际应用示例

### 用户服务缓存

```go
type UserService struct {
    cache *CacheAside
}

func NewUserService(rdb *redis.Client, db *gorm.DB) *UserService {
    return &UserService{
        cache: &CacheAside{rdb: rdb, db: db},
    }
}

func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    return s.cache.GetUser(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    return s.cache.UpdateUser(ctx, user)
}

func (s *UserService) GetUsers(ctx context.Context, ids []uint) ([]*User, error) {
    users := make([]*User, len(ids))
    var missedIDs []uint
    var missedIndexes []int

    // 1. 批量查询缓存
    keys := make([]string, len(ids))
    for i, id := range ids {
        keys[i] = fmt.Sprintf("user:%d", id)
    }

    vals, err := s.cache.rdb.MGet(ctx, keys...).Result()
    if err != nil {
        return nil, err
    }

    // 2. 处理结果
    for i, val := range vals {
        if val == nil {
            missedIDs = append(missedIDs, ids[i])
            missedIndexes = append(missedIndexes, i)
        } else {
            var user User
            if err := json.Unmarshal([]byte(val.(string)), &user); err == nil {
                users[i] = &user
            }
        }
    }

    // 3. 查询未命中的数据
    if len(missedIDs) > 0 {
        var dbUsers []User
        if err := s.cache.db.Find(&dbUsers, missedIDs).Error; err != nil {
            return nil, err
        }

        // 写入缓存
        pipe := s.cache.rdb.Pipeline()
        for _, u := range dbUsers {
            data, _ := json.Marshal(u)
            key := fmt.Sprintf("user:%d", u.ID)
            pipe.Set(ctx, key, data, time.Hour)
        }
        pipe.Exec(ctx)

        // 填充结果
        userMap := make(map[uint]*User)
        for i := range dbUsers {
            userMap[dbUsers[i].ID] = &dbUsers[i]
        }
        for i, id := range missedIDs {
            users[missedIndexes[i]] = userMap[id]
        }
    }

    return users, nil
}
```

**下一节**：[缓存问题](./05-cache-problems.md) - 学习缓存穿透、击穿、雪崩问题
