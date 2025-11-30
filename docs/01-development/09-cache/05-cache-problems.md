# 缓存问题 (Cache Problems)

## 概述

缓存虽然能显著提升性能，但也会带来一些经典问题。本节介绍三个最常见的缓存问题及其解决方案。

```
┌─────────────────────────────────────────────────────────────┐
│                      缓存问题概览                            │
├─────────────────────────────────────────────────────────────┤
│  缓存穿透 (Penetration)  │  查询不存在的数据               │
│  缓存击穿 (Breakdown)    │  热点 key 过期瞬间大量请求       │
│  缓存雪崩 (Avalanche)    │  大量 key 同时过期              │
└─────────────────────────────────────────────────────────────┘
```

## 缓存穿透 (Cache Penetration)

### 问题描述

查询不存在的数据，每次请求都穿透缓存打到数据库。

```
攻击者/异常请求: GET /user/9999999999
                    │
              ┌─────┴─────┐
              │   Cache   │ ← 无数据
              └─────┬─────┘
                    │
              ┌─────┴─────┐
              │  Database │ ← 也无数据，但每次都要查
              └───────────┘
```

### 解决方案 1：缓存空值

```go
type CachePenetrationHandler struct {
    rdb *redis.Client
    db  *gorm.DB
}

func (h *CachePenetrationHandler) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 查询缓存
    val, err := h.rdb.Get(ctx, key).Result()
    if err == nil {
        // 检查是否是空值标记
        if val == "NULL" {
            return nil, nil  // 数据不存在
        }
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            return &user, nil
        }
    }

    // 2. 查询数据库
    var user User
    err = h.db.First(&user, id).Error

    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 缓存空值，设置较短过期时间
        h.rdb.Set(ctx, key, "NULL", 5*time.Minute)
        return nil, nil
    }

    if err != nil {
        return nil, err
    }

    // 3. 缓存正常数据
    data, _ := json.Marshal(user)
    h.rdb.Set(ctx, key, data, time.Hour)

    return &user, nil
}
```

### 解决方案 2：布隆过滤器

布隆过滤器可以快速判断数据是否**可能存在**。

```go
import "github.com/bits-and-blooms/bloom/v3"

type BloomFilterCache struct {
    rdb    *redis.Client
    db     *gorm.DB
    filter *bloom.BloomFilter
    mu     sync.RWMutex
}

func NewBloomFilterCache(rdb *redis.Client, db *gorm.DB) *BloomFilterCache {
    // 创建布隆过滤器：预计 100 万条数据，误判率 0.01%
    filter := bloom.NewWithEstimates(1_000_000, 0.0001)

    cache := &BloomFilterCache{
        rdb:    rdb,
        db:     db,
        filter: filter,
    }

    // 初始化时加载所有 ID 到布隆过滤器
    cache.initFilter()

    return cache
}

func (c *BloomFilterCache) initFilter() {
    var ids []uint
    c.db.Model(&User{}).Pluck("id", &ids)

    c.mu.Lock()
    defer c.mu.Unlock()

    for _, id := range ids {
        c.filter.AddString(fmt.Sprintf("%d", id))
    }
}

func (c *BloomFilterCache) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 布隆过滤器检查
    c.mu.RLock()
    exists := c.filter.TestString(fmt.Sprintf("%d", id))
    c.mu.RUnlock()

    if !exists {
        // 一定不存在
        return nil, nil
    }

    // 2. 查询缓存
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            return &user, nil
        }
    }

    // 3. 查询数据库（可能是布隆过滤器误判）
    var user User
    if err := c.db.First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 误判，缓存空值
            c.rdb.Set(ctx, key, "NULL", 5*time.Minute)
            return nil, nil
        }
        return nil, err
    }

    // 4. 缓存数据
    data, _ := json.Marshal(user)
    c.rdb.Set(ctx, key, data, time.Hour)

    return &user, nil
}

// 新增用户时更新布隆过滤器
func (c *BloomFilterCache) AddUser(ctx context.Context, user *User) error {
    if err := c.db.Create(user).Error; err != nil {
        return err
    }

    c.mu.Lock()
    c.filter.AddString(fmt.Sprintf("%d", user.ID))
    c.mu.Unlock()

    return nil
}
```

### Redis 布隆过滤器

```go
// 使用 Redis 的布隆过滤器模块 (RedisBloom)
type RedisBloomFilter struct {
    rdb *redis.Client
}

func (f *RedisBloomFilter) Add(ctx context.Context, key, item string) error {
    return f.rdb.Do(ctx, "BF.ADD", key, item).Err()
}

func (f *RedisBloomFilter) Exists(ctx context.Context, key, item string) (bool, error) {
    result, err := f.rdb.Do(ctx, "BF.EXISTS", key, item).Int()
    return result == 1, err
}

func (f *RedisBloomFilter) MAdd(ctx context.Context, key string, items ...string) error {
    args := make([]interface{}, 0, len(items)+2)
    args = append(args, "BF.MADD", key)
    for _, item := range items {
        args = append(args, item)
    }
    return f.rdb.Do(ctx, args...).Err()
}
```

### 解决方案 3：接口校验

在请求到达缓存前进行参数校验。

```go
func (h *Handler) GetUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid id"})
        return
    }

    // 简单校验：ID 必须在合理范围内
    if id <= 0 || id > 100_000_000 {
        c.JSON(400, gin.H{"error": "invalid id range"})
        return
    }

    user, err := h.service.GetUser(c.Request.Context(), uint(id))
    // ...
}
```

## 缓存击穿 (Cache Breakdown)

### 问题描述

热点 key 过期的瞬间，大量请求同时打到数据库。

```
时间线:
─────────────────────────────────────────────────────▶
    │                    │
    热点 key 过期        大量请求同时到达
                         │
                    ┌────┴────┐
                    │ 1000个  │
                    │  请求   │
                    └────┬────┘
                         │
                    ┌────┴────┐
                    │ Database│ ← 瞬间高负载
                    └─────────┘
```

### 解决方案 1：互斥锁 (Mutex)

```go
type MutexCache struct {
    rdb *redis.Client
    db  *gorm.DB
}

func (c *MutexCache) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    lockKey := fmt.Sprintf("lock:user:%d", id)

    // 1. 查询缓存
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            return &user, nil
        }
    }

    // 2. 获取分布式锁
    acquired, err := c.rdb.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
    if err != nil {
        return nil, err
    }

    if acquired {
        // 获得锁，负责查询数据库
        defer c.rdb.Del(ctx, lockKey)

        // 再次检查缓存（double check）
        val, err = c.rdb.Get(ctx, key).Result()
        if err == nil {
            var user User
            if err := json.Unmarshal([]byte(val), &user); err == nil {
                return &user, nil
            }
        }

        // 查询数据库
        var user User
        if err := c.db.First(&user, id).Error; err != nil {
            return nil, err
        }

        // 写入缓存
        data, _ := json.Marshal(user)
        c.rdb.Set(ctx, key, data, time.Hour)

        return &user, nil
    }

    // 3. 未获得锁，等待后重试
    time.Sleep(50 * time.Millisecond)
    return c.GetUser(ctx, id)
}
```

### 解决方案 2：逻辑过期

不设置 Redis TTL，在数据中记录逻辑过期时间。

```go
type CacheValue struct {
    Data       json.RawMessage `json:"data"`
    ExpireAt   int64           `json:"expire_at"`
}

type LogicalExpireCache struct {
    rdb        *redis.Client
    db         *gorm.DB
    refreshing sync.Map
}

func (c *LogicalExpireCache) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 查询缓存
    val, err := c.rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        // 缓存不存在，同步加载
        return c.loadUser(ctx, id)
    }
    if err != nil {
        return nil, err
    }

    // 2. 解析缓存值
    var cacheVal CacheValue
    if err := json.Unmarshal([]byte(val), &cacheVal); err != nil {
        return c.loadUser(ctx, id)
    }

    // 3. 检查逻辑过期
    var user User
    if err := json.Unmarshal(cacheVal.Data, &user); err != nil {
        return c.loadUser(ctx, id)
    }

    if time.Now().Unix() > cacheVal.ExpireAt {
        // 已过期，异步刷新
        go c.refreshAsync(context.Background(), id)
    }

    // 4. 返回数据（即使过期也返回旧数据）
    return &user, nil
}

func (c *LogicalExpireCache) refreshAsync(ctx context.Context, id uint) {
    key := fmt.Sprintf("user:%d", id)

    // 防止重复刷新
    if _, loaded := c.refreshing.LoadOrStore(key, true); loaded {
        return
    }
    defer c.refreshing.Delete(key)

    c.loadUser(ctx, id)
}

func (c *LogicalExpireCache) loadUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    var user User
    if err := c.db.First(&user, id).Error; err != nil {
        return nil, err
    }

    // 设置逻辑过期时间
    data, _ := json.Marshal(user)
    cacheVal := CacheValue{
        Data:     data,
        ExpireAt: time.Now().Add(time.Hour).Unix(),
    }
    valData, _ := json.Marshal(cacheVal)

    // 不设置 Redis TTL
    c.rdb.Set(ctx, key, valData, 0)

    return &user, nil
}
```

### 解决方案 3：singleflight

使用 `singleflight` 合并并发请求。

```go
import "golang.org/x/sync/singleflight"

type SingleFlightCache struct {
    rdb *redis.Client
    db  *gorm.DB
    sf  singleflight.Group
}

func (c *SingleFlightCache) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 查询缓存
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            return &user, nil
        }
    }

    // 2. 使用 singleflight 防止并发查询
    result, err, _ := c.sf.Do(key, func() (interface{}, error) {
        // 再次检查缓存
        val, err := c.rdb.Get(ctx, key).Result()
        if err == nil {
            var user User
            if err := json.Unmarshal([]byte(val), &user); err == nil {
                return &user, nil
            }
        }

        // 查询数据库
        var user User
        if err := c.db.First(&user, id).Error; err != nil {
            return nil, err
        }

        // 写入缓存
        data, _ := json.Marshal(user)
        c.rdb.Set(ctx, key, data, time.Hour)

        return &user, nil
    })

    if err != nil {
        return nil, err
    }
    return result.(*User), nil
}
```

## 缓存雪崩 (Cache Avalanche)

### 问题描述

大量 key 同时过期，导致请求全部打到数据库。

```
时间线:
─────────────────────────────────────────────────────▶
    │
    同一时刻设置的 10000 个 key 同时过期
    │
    ┌─────────────────────────────┐
    │   10000 个请求同时到达       │
    └─────────────┬───────────────┘
                  │
            ┌─────┴─────┐
            │ Database  │ ← 崩溃
            └───────────┘
```

### 解决方案 1：随机过期时间

```go
func SetWithRandomExpiry(rdb *redis.Client, ctx context.Context, key string, value interface{}, baseExpiry time.Duration) error {
    // 基础过期时间 + 随机偏移
    // 例如：基础 1 小时，随机增加 0-10 分钟
    randomOffset := time.Duration(rand.Intn(600)) * time.Second
    expiry := baseExpiry + randomOffset

    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return rdb.Set(ctx, key, data, expiry).Err()
}

// 批量设置
func BatchSetWithRandomExpiry(rdb *redis.Client, ctx context.Context, items map[string]interface{}, baseExpiry time.Duration) error {
    pipe := rdb.Pipeline()

    for key, value := range items {
        randomOffset := time.Duration(rand.Intn(600)) * time.Second
        expiry := baseExpiry + randomOffset

        data, _ := json.Marshal(value)
        pipe.Set(ctx, key, data, expiry)
    }

    _, err := pipe.Exec(ctx)
    return err
}
```

### 解决方案 2：多级缓存

```go
type MultiLevelCache struct {
    l1  *cache.Cache  // 本地缓存
    l2  *redis.Client // Redis
    db  *gorm.DB
}

func (c *MultiLevelCache) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. L1 本地缓存
    if val, found := c.l1.Get(key); found {
        return val.(*User), nil
    }

    // 2. L2 Redis
    val, err := c.l2.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            // 回填 L1
            c.l1.Set(key, &user, cache.DefaultExpiration)
            return &user, nil
        }
    }

    // 3. 数据库
    var user User
    if err := c.db.First(&user, id).Error; err != nil {
        return nil, err
    }

    // 回填 L2 和 L1
    data, _ := json.Marshal(user)
    c.l2.Set(ctx, key, data, time.Hour+time.Duration(rand.Intn(600))*time.Second)
    c.l1.Set(key, &user, 5*time.Minute)

    return &user, nil
}
```

### 解决方案 3：缓存预热

```go
type CacheWarmer struct {
    rdb *redis.Client
    db  *gorm.DB
}

// 启动时预热缓存
func (w *CacheWarmer) WarmUp(ctx context.Context) error {
    // 预热热点数据
    var hotUsers []User
    if err := w.db.Order("view_count DESC").Limit(1000).Find(&hotUsers).Error; err != nil {
        return err
    }

    pipe := w.rdb.Pipeline()
    for _, user := range hotUsers {
        key := fmt.Sprintf("user:%d", user.ID)
        data, _ := json.Marshal(user)

        // 随机过期时间
        expiry := time.Hour + time.Duration(rand.Intn(3600))*time.Second
        pipe.Set(ctx, key, data, expiry)
    }

    _, err := pipe.Exec(ctx)
    return err
}

// 定时刷新热点数据
func (w *CacheWarmer) RefreshHotData(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.WarmUp(ctx)
        }
    }
}
```

### 解决方案 4：熔断降级

```go
import "github.com/sony/gobreaker"

type CircuitBreakerCache struct {
    rdb *redis.Client
    db  *gorm.DB
    cb  *gobreaker.CircuitBreaker
}

func NewCircuitBreakerCache(rdb *redis.Client, db *gorm.DB) *CircuitBreakerCache {
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "cache",
        MaxRequests: 100,              // 半开状态允许的请求数
        Interval:    10 * time.Second, // 统计周期
        Timeout:     30 * time.Second, // 开启后多久尝试恢复
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // 失败率超过 50% 且请求数超过 10 则熔断
            return counts.Requests > 10 && counts.ConsecutiveFailures > 5
        },
    })

    return &CircuitBreakerCache{rdb: rdb, db: db, cb: cb}
}

func (c *CircuitBreakerCache) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 尝试从缓存获取
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            return &user, nil
        }
    }

    // 2. 通过熔断器访问数据库
    result, err := c.cb.Execute(func() (interface{}, error) {
        var user User
        if err := c.db.First(&user, id).Error; err != nil {
            return nil, err
        }

        // 写入缓存
        data, _ := json.Marshal(user)
        c.rdb.Set(ctx, key, data, time.Hour)

        return &user, nil
    })

    if err != nil {
        // 熔断器开启，返回降级数据
        if err == gobreaker.ErrOpenState {
            return c.getFallbackUser(ctx, id)
        }
        return nil, err
    }

    return result.(*User), nil
}

// 降级数据
func (c *CircuitBreakerCache) getFallbackUser(ctx context.Context, id uint) (*User, error) {
    // 返回默认数据或从备份缓存获取
    return &User{
        ID:   id,
        Name: "Unknown",
    }, nil
}
```

## 问题对比与解决方案选择

| 问题 | 原因 | 推荐方案 | 备选方案 |
|------|------|----------|----------|
| 缓存穿透 | 查询不存在的数据 | 布隆过滤器 | 缓存空值 + 参数校验 |
| 缓存击穿 | 热点 key 过期 | singleflight | 互斥锁 + 逻辑过期 |
| 缓存雪崩 | 大量 key 同时过期 | 随机过期时间 | 多级缓存 + 熔断 |

## 综合解决方案

```go
type RobustCache struct {
    rdb        *redis.Client
    db         *gorm.DB
    sf         singleflight.Group
    filter     *bloom.BloomFilter
    l1         *cache.Cache
    cb         *gobreaker.CircuitBreaker
    mu         sync.RWMutex
}

func (c *RobustCache) GetUser(ctx context.Context, id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 布隆过滤器检查（防穿透）
    c.mu.RLock()
    exists := c.filter.TestString(fmt.Sprintf("%d", id))
    c.mu.RUnlock()
    if !exists {
        return nil, nil
    }

    // 2. L1 本地缓存（防雪崩）
    if val, found := c.l1.Get(key); found {
        return val.(*User), nil
    }

    // 3. L2 Redis
    val, err := c.rdb.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            c.l1.Set(key, &user, 5*time.Minute)
            return &user, nil
        }
    }

    // 4. singleflight 防击穿
    result, err, _ := c.sf.Do(key, func() (interface{}, error) {
        // 熔断器保护数据库
        return c.cb.Execute(func() (interface{}, error) {
            var user User
            if err := c.db.First(&user, id).Error; err != nil {
                if errors.Is(err, gorm.ErrRecordNotFound) {
                    return nil, nil
                }
                return nil, err
            }

            // 写入缓存（随机过期时间）
            data, _ := json.Marshal(user)
            expiry := time.Hour + time.Duration(rand.Intn(600))*time.Second
            c.rdb.Set(ctx, key, data, expiry)
            c.l1.Set(key, &user, 5*time.Minute)

            return &user, nil
        })
    })

    if err != nil {
        return nil, err
    }
    if result == nil {
        return nil, nil
    }
    return result.(*User), nil
}
```

**下一节**：[本地缓存](./06-local-cache.md) - 学习 go-cache 本地缓存使用
