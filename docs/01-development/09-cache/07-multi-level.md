# 多级缓存 (Multi-Level Cache)

## 概述

多级缓存通过组合不同类型的缓存，在性能和一致性之间取得平衡。

```
┌─────────────────────────────────────────────────────────────┐
│                      多级缓存架构                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌──────┐  │
│  │ Client  │────▶│   L1    │────▶│   L2    │────▶│  DB  │  │
│  └─────────┘     │ (本地)  │     │ (Redis) │     └──────┘  │
│                  └─────────┘     └─────────┘               │
│                       │               │                    │
│                   < 1μs           ~1ms                     │
│                   热点数据        共享数据                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 为什么需要多级缓存

### 单级缓存的问题

| 缓存类型 | 问题 |
|----------|------|
| 仅本地缓存 | 多实例数据不一致，重启丢失 |
| 仅 Redis | 网络延迟，热点数据压力大 |

### 多级缓存的优势

| 优势 | 说明 |
|------|------|
| 性能最优 | 热点数据在本地，极速响应 |
| 减轻 Redis 压力 | 大量请求被 L1 拦截 |
| 高可用 | Redis 故障时 L1 可用 |
| 成本效益 | 用少量内存获得极致性能 |

## 基础实现

### 双层缓存结构

```go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/patrickmn/go-cache"
    "github.com/redis/go-redis/v9"
)

type MultiLevelCache struct {
    l1      *cache.Cache   // 本地缓存
    l2      *redis.Client  // Redis
    l1TTL   time.Duration
    l2TTL   time.Duration
}

func NewMultiLevelCache(redisAddr string, l1TTL, l2TTL time.Duration) *MultiLevelCache {
    return &MultiLevelCache{
        l1:    cache.New(l1TTL, l1TTL*2),
        l2:    redis.NewClient(&redis.Options{Addr: redisAddr}),
        l1TTL: l1TTL,
        l2TTL: l2TTL,
    }
}

// Get 多级查询
func (c *MultiLevelCache) Get(ctx context.Context, key string, dest interface{}) error {
    // 1. L1 本地缓存
    if val, found := c.l1.Get(key); found {
        // 类型断言
        if data, ok := val.([]byte); ok {
            return json.Unmarshal(data, dest)
        }
    }

    // 2. L2 Redis
    val, err := c.l2.Get(ctx, key).Result()
    if err == redis.Nil {
        return ErrNotFound
    }
    if err != nil {
        return err
    }

    // 回填 L1
    c.l1.Set(key, []byte(val), c.l1TTL)

    return json.Unmarshal([]byte(val), dest)
}

// Set 多级写入
func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    // 同时写入 L1 和 L2
    c.l1.Set(key, data, c.l1TTL)
    return c.l2.Set(ctx, key, data, c.l2TTL).Err()
}

// Delete 多级删除
func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
    c.l1.Delete(key)
    return c.l2.Del(ctx, key).Err()
}

var ErrNotFound = errors.New("cache: key not found")
```

### 泛型版本

```go
type MultiLevelCache[T any] struct {
    l1      *cache.Cache
    l2      *redis.Client
    l1TTL   time.Duration
    l2TTL   time.Duration
}

func NewMultiLevelCache[T any](redisAddr string, l1TTL, l2TTL time.Duration) *MultiLevelCache[T] {
    return &MultiLevelCache[T]{
        l1:    cache.New(l1TTL, l1TTL*2),
        l2:    redis.NewClient(&redis.Options{Addr: redisAddr}),
        l1TTL: l1TTL,
        l2TTL: l2TTL,
    }
}

func (c *MultiLevelCache[T]) Get(ctx context.Context, key string) (T, error) {
    var zero T

    // L1
    if val, found := c.l1.Get(key); found {
        if data, ok := val.([]byte); ok {
            var result T
            if err := json.Unmarshal(data, &result); err == nil {
                return result, nil
            }
        }
    }

    // L2
    val, err := c.l2.Get(ctx, key).Result()
    if err == redis.Nil {
        return zero, ErrNotFound
    }
    if err != nil {
        return zero, err
    }

    // 回填 L1
    c.l1.Set(key, []byte(val), c.l1TTL)

    var result T
    if err := json.Unmarshal([]byte(val), &result); err != nil {
        return zero, err
    }
    return result, nil
}

func (c *MultiLevelCache[T]) Set(ctx context.Context, key string, value T) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    c.l1.Set(key, data, c.l1TTL)
    return c.l2.Set(ctx, key, data, c.l2TTL).Err()
}

// 使用示例
func main() {
    userCache := NewMultiLevelCache[*User]("localhost:6379", 5*time.Minute, time.Hour)

    ctx := context.Background()

    // 设置
    userCache.Set(ctx, "user:1", &User{ID: 1, Name: "Alice"})

    // 获取
    user, err := userCache.Get(ctx, "user:1")
    if err == nil {
        fmt.Println(user.Name)
    }
}
```

## 高级特性

### 带加载器的多级缓存

```go
type LoadingMultiLevelCache[T any] struct {
    l1     *cache.Cache
    l2     *redis.Client
    loader func(ctx context.Context, key string) (T, error)
    l1TTL  time.Duration
    l2TTL  time.Duration
    sf     singleflight.Group
}

func NewLoadingMultiLevelCache[T any](
    redisAddr string,
    loader func(ctx context.Context, key string) (T, error),
    l1TTL, l2TTL time.Duration,
) *LoadingMultiLevelCache[T] {
    return &LoadingMultiLevelCache[T]{
        l1:     cache.New(l1TTL, l1TTL*2),
        l2:     redis.NewClient(&redis.Options{Addr: redisAddr}),
        loader: loader,
        l1TTL:  l1TTL,
        l2TTL:  l2TTL,
    }
}

func (c *LoadingMultiLevelCache[T]) Get(ctx context.Context, key string) (T, error) {
    var zero T

    // 1. L1 本地缓存
    if val, found := c.l1.Get(key); found {
        if data, ok := val.([]byte); ok {
            var result T
            if err := json.Unmarshal(data, &result); err == nil {
                return result, nil
            }
        }
    }

    // 2. L2 Redis
    val, err := c.l2.Get(ctx, key).Result()
    if err == nil {
        c.l1.Set(key, []byte(val), c.l1TTL)
        var result T
        if err := json.Unmarshal([]byte(val), &result); err == nil {
            return result, nil
        }
    }

    // 3. 使用 singleflight 加载数据
    result, err, _ := c.sf.Do(key, func() (interface{}, error) {
        // 再次检查 L2（其他 goroutine 可能已加载）
        val, err := c.l2.Get(ctx, key).Result()
        if err == nil {
            var result T
            if err := json.Unmarshal([]byte(val), &result); err == nil {
                c.l1.Set(key, []byte(val), c.l1TTL)
                return result, nil
            }
        }

        // 从数据源加载
        data, err := c.loader(ctx, key)
        if err != nil {
            return zero, err
        }

        // 回填 L1 和 L2
        jsonData, _ := json.Marshal(data)
        c.l1.Set(key, jsonData, c.l1TTL)
        c.l2.Set(ctx, key, jsonData, c.l2TTL)

        return data, nil
    })

    if err != nil {
        return zero, err
    }
    return result.(T), nil
}

// 使用示例
func main() {
    cache := NewLoadingMultiLevelCache[*User](
        "localhost:6379",
        func(ctx context.Context, key string) (*User, error) {
            // 从数据库加载
            var id uint
            fmt.Sscanf(key, "user:%d", &id)
            var user User
            if err := db.First(&user, id).Error; err != nil {
                return nil, err
            }
            return &user, nil
        },
        5*time.Minute,
        time.Hour,
    )

    // 自动加载
    user, err := cache.Get(ctx, "user:1")
}
```

### 带统计的多级缓存

```go
type CacheStats struct {
    L1Hits   int64
    L1Misses int64
    L2Hits   int64
    L2Misses int64
    Loads    int64
}

func (s *CacheStats) L1HitRate() float64 {
    total := s.L1Hits + s.L1Misses
    if total == 0 {
        return 0
    }
    return float64(s.L1Hits) / float64(total) * 100
}

func (s *CacheStats) L2HitRate() float64 {
    total := s.L2Hits + s.L2Misses
    if total == 0 {
        return 0
    }
    return float64(s.L2Hits) / float64(total) * 100
}

type StatsMultiLevelCache[T any] struct {
    l1     *cache.Cache
    l2     *redis.Client
    stats  CacheStats
    mu     sync.Mutex
    // ...
}

func (c *StatsMultiLevelCache[T]) Get(ctx context.Context, key string) (T, error) {
    var zero T

    // L1
    if val, found := c.l1.Get(key); found {
        atomic.AddInt64(&c.stats.L1Hits, 1)
        // ...
    } else {
        atomic.AddInt64(&c.stats.L1Misses, 1)
    }

    // L2
    val, err := c.l2.Get(ctx, key).Result()
    if err == nil {
        atomic.AddInt64(&c.stats.L2Hits, 1)
        // ...
    } else if err == redis.Nil {
        atomic.AddInt64(&c.stats.L2Misses, 1)
    }

    // Load
    atomic.AddInt64(&c.stats.Loads, 1)
    // ...
}

func (c *StatsMultiLevelCache[T]) Stats() CacheStats {
    return CacheStats{
        L1Hits:   atomic.LoadInt64(&c.stats.L1Hits),
        L1Misses: atomic.LoadInt64(&c.stats.L1Misses),
        L2Hits:   atomic.LoadInt64(&c.stats.L2Hits),
        L2Misses: atomic.LoadInt64(&c.stats.L2Misses),
        Loads:    atomic.LoadInt64(&c.stats.Loads),
    }
}
```

## 缓存同步

### 问题：多实例 L1 不一致

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│ 实例 A  │     │ 实例 B  │     │ 实例 C  │
│  L1: v1 │     │  L1: v1 │     │  L1: v1 │
└─────────┘     └─────────┘     └─────────┘

用户更新数据到实例 A:

┌─────────┐     ┌─────────┐     ┌─────────┐
│ 实例 A  │     │ 实例 B  │     │ 实例 C  │
│  L1: v2 │     │  L1: v1 │     │  L1: v1 │  ← B、C 的 L1 还是旧数据！
└─────────┘     └─────────┘     └─────────┘
```

### 解决方案 1：Redis Pub/Sub 同步

```go
type SyncMultiLevelCache struct {
    l1        *cache.Cache
    l2        *redis.Client
    pubsub    *redis.PubSub
    channel   string
    instanceID string
}

func NewSyncMultiLevelCache(redisAddr string) *SyncMultiLevelCache {
    rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
    instanceID := uuid.New().String()

    c := &SyncMultiLevelCache{
        l1:         cache.New(5*time.Minute, 10*time.Minute),
        l2:         rdb,
        channel:    "cache:invalidate",
        instanceID: instanceID,
    }

    // 订阅失效消息
    c.pubsub = rdb.Subscribe(context.Background(), c.channel)
    go c.subscribeInvalidation()

    return c
}

type InvalidationMessage struct {
    Key        string `json:"key"`
    InstanceID string `json:"instance_id"`
}

func (c *SyncMultiLevelCache) subscribeInvalidation() {
    ch := c.pubsub.Channel()
    for msg := range ch {
        var inv InvalidationMessage
        if err := json.Unmarshal([]byte(msg.Payload), &inv); err != nil {
            continue
        }

        // 忽略自己发送的消息
        if inv.InstanceID == c.instanceID {
            continue
        }

        // 删除本地缓存
        c.l1.Delete(inv.Key)
        log.Printf("Invalidated local cache: %s", inv.Key)
    }
}

func (c *SyncMultiLevelCache) Delete(ctx context.Context, key string) error {
    // 1. 删除本地缓存
    c.l1.Delete(key)

    // 2. 删除 Redis
    if err := c.l2.Del(ctx, key).Err(); err != nil {
        return err
    }

    // 3. 广播失效消息
    msg := InvalidationMessage{
        Key:        key,
        InstanceID: c.instanceID,
    }
    data, _ := json.Marshal(msg)
    return c.l2.Publish(ctx, c.channel, data).Err()
}

func (c *SyncMultiLevelCache) Set(ctx context.Context, key string, value interface{}) error {
    data, _ := json.Marshal(value)

    // 写入本地和 Redis
    c.l1.Set(key, data, cache.DefaultExpiration)
    if err := c.l2.Set(ctx, key, data, time.Hour).Err(); err != nil {
        return err
    }

    // 广播更新（让其他实例删除旧的本地缓存）
    msg := InvalidationMessage{
        Key:        key,
        InstanceID: c.instanceID,
    }
    msgData, _ := json.Marshal(msg)
    return c.l2.Publish(ctx, c.channel, msgData).Err()
}

func (c *SyncMultiLevelCache) Close() {
    c.pubsub.Close()
    c.l2.Close()
}
```

### 解决方案 2：版本号控制

```go
type VersionedValue struct {
    Data    json.RawMessage `json:"data"`
    Version int64           `json:"version"`
}

type VersionedCache struct {
    l1 *cache.Cache
    l2 *redis.Client
}

func (c *VersionedCache) Get(ctx context.Context, key string, dest interface{}) error {
    // 获取 L2 版本
    l2Val, err := c.l2.Get(ctx, key).Result()
    if err == redis.Nil {
        return ErrNotFound
    }
    if err != nil {
        return err
    }

    var l2Data VersionedValue
    if err := json.Unmarshal([]byte(l2Val), &l2Data); err != nil {
        return err
    }

    // 检查 L1 版本
    if l1Val, found := c.l1.Get(key); found {
        var l1Data VersionedValue
        if err := json.Unmarshal(l1Val.([]byte), &l1Data); err == nil {
            if l1Data.Version >= l2Data.Version {
                // L1 版本不旧，使用 L1
                return json.Unmarshal(l1Data.Data, dest)
            }
        }
    }

    // 更新 L1
    c.l1.Set(key, []byte(l2Val), cache.DefaultExpiration)
    return json.Unmarshal(l2Data.Data, dest)
}

func (c *VersionedCache) Set(ctx context.Context, key string, value interface{}) error {
    // 获取新版本号
    version := time.Now().UnixNano()

    data, _ := json.Marshal(value)
    versionedData := VersionedValue{
        Data:    data,
        Version: version,
    }
    jsonData, _ := json.Marshal(versionedData)

    // 写入 L1 和 L2
    c.l1.Set(key, jsonData, cache.DefaultExpiration)
    return c.l2.Set(ctx, key, jsonData, time.Hour).Err()
}
```

## 实际应用

### 用户服务多级缓存

```go
type UserService struct {
    cache *LoadingMultiLevelCache[*User]
    db    *gorm.DB
}

func NewUserService(rdb *redis.Client, db *gorm.DB) *UserService {
    svc := &UserService{db: db}

    svc.cache = NewLoadingMultiLevelCache[*User](
        "localhost:6379",
        svc.loadUser,
        5*time.Minute,   // L1 TTL
        time.Hour,       // L2 TTL
    )

    return svc
}

func (s *UserService) loadUser(ctx context.Context, key string) (*User, error) {
    var id uint
    fmt.Sscanf(key, "user:%d", &id)

    var user User
    if err := s.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    return s.cache.Get(ctx, key)
}

func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    // 更新数据库
    if err := s.db.Save(user).Error; err != nil {
        return err
    }

    // 删除缓存（下次访问会重新加载）
    key := fmt.Sprintf("user:%d", user.ID)
    return s.cache.Delete(ctx, key)
}

func (s *UserService) GetUsers(ctx context.Context, ids []uint) ([]*User, error) {
    users := make([]*User, len(ids))
    var missedIDs []uint
    var missedIndexes []int

    // 批量从缓存获取
    for i, id := range ids {
        key := fmt.Sprintf("user:%d", id)
        user, err := s.cache.Get(ctx, key)
        if err == nil {
            users[i] = user
        } else {
            missedIDs = append(missedIDs, id)
            missedIndexes = append(missedIndexes, i)
        }
    }

    // 批量从数据库获取未命中的
    if len(missedIDs) > 0 {
        var dbUsers []User
        if err := s.db.Find(&dbUsers, missedIDs).Error; err != nil {
            return nil, err
        }

        userMap := make(map[uint]*User)
        for i := range dbUsers {
            userMap[dbUsers[i].ID] = &dbUsers[i]
            // 写入缓存
            key := fmt.Sprintf("user:%d", dbUsers[i].ID)
            s.cache.Set(ctx, key, &dbUsers[i])
        }

        for i, id := range missedIDs {
            users[missedIndexes[i]] = userMap[id]
        }
    }

    return users, nil
}
```

### 配置中心多级缓存

```go
type ConfigService struct {
    l1       *cache.Cache
    l2       *redis.Client
    db       *gorm.DB
    watchers map[string][]func(string, string)
    mu       sync.RWMutex
}

func NewConfigService(rdb *redis.Client, db *gorm.DB) *ConfigService {
    svc := &ConfigService{
        l1:       cache.New(time.Minute, 5*time.Minute),
        l2:       rdb,
        db:       db,
        watchers: make(map[string][]func(string, string)),
    }

    // 启动配置监听
    go svc.watchConfig()

    return svc
}

func (s *ConfigService) Get(ctx context.Context, key string) (string, error) {
    // L1
    if val, found := s.l1.Get(key); found {
        return val.(string), nil
    }

    // L2
    val, err := s.l2.Get(ctx, "config:"+key).Result()
    if err == nil {
        s.l1.Set(key, val, cache.DefaultExpiration)
        return val, nil
    }

    // DB
    var config Config
    if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
        return "", err
    }

    // 回填缓存
    s.l1.Set(key, config.Value, cache.DefaultExpiration)
    s.l2.Set(ctx, "config:"+key, config.Value, time.Hour)

    return config.Value, nil
}

func (s *ConfigService) Set(ctx context.Context, key, value string) error {
    // 更新数据库
    if err := s.db.Save(&Config{Key: key, Value: value}).Error; err != nil {
        return err
    }

    // 更新缓存
    s.l1.Set(key, value, cache.DefaultExpiration)
    s.l2.Set(ctx, "config:"+key, value, time.Hour)

    // 发布变更通知
    s.l2.Publish(ctx, "config:changed", key)

    // 通知本地监听者
    s.notifyWatchers(key, value)

    return nil
}

func (s *ConfigService) Watch(key string, callback func(key, value string)) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.watchers[key] = append(s.watchers[key], callback)
}

func (s *ConfigService) notifyWatchers(key, value string) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    if watchers, ok := s.watchers[key]; ok {
        for _, w := range watchers {
            go w(key, value)
        }
    }
}

func (s *ConfigService) watchConfig() {
    ctx := context.Background()
    pubsub := s.l2.Subscribe(ctx, "config:changed")
    defer pubsub.Close()

    ch := pubsub.Channel()
    for msg := range ch {
        key := msg.Payload
        // 删除本地缓存，下次访问会重新加载
        s.l1.Delete(key)

        // 获取新值并通知监听者
        if value, err := s.Get(ctx, key); err == nil {
            s.notifyWatchers(key, value)
        }
    }
}
```

## TTL 策略

### L1 和 L2 TTL 关系

```
L1 TTL < L2 TTL

原因：
1. L1 更新频繁，保持数据新鲜
2. L2 作为后备，减少数据库访问
3. 允许一定程度的数据延迟

推荐配置：
- L1 TTL: 1-5 分钟
- L2 TTL: 1-24 小时
```

### 动态 TTL

```go
func (c *MultiLevelCache) SetWithDynamicTTL(ctx context.Context, key string, value interface{}, accessFrequency int) error {
    data, _ := json.Marshal(value)

    // 根据访问频率调整 TTL
    var l1TTL, l2TTL time.Duration

    switch {
    case accessFrequency > 1000:
        // 热点数据，较长 TTL
        l1TTL = 10 * time.Minute
        l2TTL = 2 * time.Hour
    case accessFrequency > 100:
        l1TTL = 5 * time.Minute
        l2TTL = time.Hour
    default:
        // 冷数据，较短 TTL
        l1TTL = time.Minute
        l2TTL = 30 * time.Minute
    }

    c.l1.Set(key, data, l1TTL)
    return c.l2.Set(ctx, key, data, l2TTL).Err()
}
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| L1 TTL < L2 TTL | 保证数据一致性 |
| 使用 singleflight | 防止缓存击穿 |
| 实现缓存同步 | 多实例 L1 一致性 |
| 监控命中率 | 分别监控 L1、L2 命中率 |
| 合理设置容量 | L1 容量受限，只缓存热点 |
| 降级策略 | Redis 故障时仅用 L1 |

**下一节**：[缓存一致性](./08-cache-consistency.md) - 学习数据库与缓存同步策略
