# 缓存策略练习 (Cache Exercises)

## 练习 1：Redis 基础操作

实现一个 Redis 工具类，封装常用操作。

### 要求

1. 实现 String 类型的 Get/Set/Delete
2. 实现 Hash 类型的 HGet/HSet/HGetAll
3. 支持过期时间设置
4. 添加连接健康检查

### 参考代码

```go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
    // 实现连接创建
}

func (r *RedisClient) Ping(ctx context.Context) error {
    // 实现健康检查
}

func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
    // 实现获取并反序列化
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    // 实现序列化并设置
}

func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
    // 实现删除
}

func (r *RedisClient) HGet(ctx context.Context, key, field string, dest interface{}) error {
    // 实现 Hash 字段获取
}

func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
    // 实现 Hash 字段设置
}

func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
    // 实现获取所有字段
}
```

---

## 练习 2：Cache-Aside 模式

实现一个用户缓存服务，使用 Cache-Aside 模式。

### 要求

1. 实现 GetUser - 先查缓存，未命中查数据库
2. 实现 UpdateUser - 更新数据库后删除缓存
3. 实现 GetUsers - 批量获取，尽量减少数据库查询
4. 添加缓存命中率统计

### 参考代码

```go
package service

type UserCache struct {
    rdb   *redis.Client
    db    *gorm.DB
    stats CacheStats
}

type CacheStats struct {
    Hits   int64
    Misses int64
}

func (c *UserCache) GetUser(ctx context.Context, id uint) (*User, error) {
    // 1. 查询缓存
    // 2. 缓存未命中，查询数据库
    // 3. 写入缓存
    // 4. 更新统计
}

func (c *UserCache) UpdateUser(ctx context.Context, user *User) error {
    // 1. 更新数据库
    // 2. 删除缓存
}

func (c *UserCache) GetUsers(ctx context.Context, ids []uint) ([]*User, error) {
    // 1. 批量查询缓存 (MGET)
    // 2. 收集未命中的 ID
    // 3. 批量查询数据库
    // 4. 批量写入缓存 (Pipeline)
}

func (c *UserCache) HitRate() float64 {
    // 计算命中率
}
```

---

## 练习 3：缓存穿透防护

实现布隆过滤器防止缓存穿透。

### 要求

1. 使用布隆过滤器预加载所有有效 ID
2. 查询时先检查布隆过滤器
3. 对不存在的数据缓存空值
4. 实现过滤器的动态更新

### 参考代码

```go
package cache

import "github.com/bits-and-blooms/bloom/v3"

type AntiPenetrationCache struct {
    rdb    *redis.Client
    db     *gorm.DB
    filter *bloom.BloomFilter
    mu     sync.RWMutex
}

func NewAntiPenetrationCache(rdb *redis.Client, db *gorm.DB) *AntiPenetrationCache {
    // 初始化布隆过滤器
    // 预加载所有 ID
}

func (c *AntiPenetrationCache) GetProduct(ctx context.Context, id uint) (*Product, error) {
    // 1. 布隆过滤器检查
    // 2. 查询缓存
    // 3. 处理空值缓存
    // 4. 查询数据库
}

func (c *AntiPenetrationCache) AddProduct(ctx context.Context, product *Product) error {
    // 1. 写入数据库
    // 2. 更新布隆过滤器
}
```

---

## 练习 4：缓存击穿防护

使用 singleflight 防止缓存击穿。

### 要求

1. 使用 singleflight 合并并发请求
2. 实现热点数据的自动续期
3. 添加请求排队超时机制

### 参考代码

```go
package cache

import "golang.org/x/sync/singleflight"

type AntiBreakdownCache struct {
    rdb *redis.Client
    db  *gorm.DB
    sf  singleflight.Group
}

func (c *AntiBreakdownCache) GetHotProduct(ctx context.Context, id uint) (*Product, error) {
    key := fmt.Sprintf("product:%d", id)

    // 1. 查询缓存
    // 2. 使用 singleflight 加载
    // 3. 续期热点数据
}

func (c *AntiBreakdownCache) refreshIfNeeded(ctx context.Context, key string, ttl time.Duration) {
    // 如果 TTL 小于阈值，异步刷新
}
```

---

## 练习 5：多级缓存

实现 L1 (本地) + L2 (Redis) 多级缓存。

### 要求

1. L1 使用 go-cache，L2 使用 Redis
2. 实现自动回填机制
3. 使用 Redis Pub/Sub 同步多实例的 L1 缓存
4. 分别统计 L1 和 L2 命中率

### 参考代码

```go
package cache

type MultiLevelCache struct {
    l1         *cache.Cache
    l2         *redis.Client
    pubsub     *redis.PubSub
    instanceID string
    stats      MultiLevelStats
}

type MultiLevelStats struct {
    L1Hits   int64
    L1Misses int64
    L2Hits   int64
    L2Misses int64
}

func NewMultiLevelCache(redisAddr string) *MultiLevelCache {
    // 初始化 L1 和 L2
    // 订阅失效消息
    // 启动消息处理 goroutine
}

func (c *MultiLevelCache) Get(ctx context.Context, key string, dest interface{}) error {
    // 1. 查询 L1
    // 2. 查询 L2
    // 3. 回填 L1
    // 4. 更新统计
}

func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, l1TTL, l2TTL time.Duration) error {
    // 1. 写入 L1 和 L2
    // 2. 广播失效消息
}

func (c *MultiLevelCache) handleInvalidation() {
    // 处理来自其他实例的失效消息
}
```

---

## 练习 6：延迟双删

实现延迟双删策略保证缓存一致性。

### 要求

1. 更新数据库前删除缓存
2. 更新成功后延迟删除缓存
3. 实现可靠的延迟删除（使用消息队列）
4. 添加删除失败重试机制

### 参考代码

```go
package cache

type DelayedDoubleDelete struct {
    rdb        *redis.Client
    db         *gorm.DB
    delayQueue chan DelayedTask
    retryQueue chan string
}

type DelayedTask struct {
    Key      string
    DeleteAt time.Time
}

func (d *DelayedDoubleDelete) UpdateWithDoubleDelete(ctx context.Context, key string, updateFn func() error) error {
    // 1. 第一次删除缓存
    // 2. 执行更新
    // 3. 发送延迟删除任务
}

func (d *DelayedDoubleDelete) processDelayedDeletions() {
    // 处理延迟删除任务
}

func (d *DelayedDoubleDelete) processRetries() {
    // 处理删除失败重试
}
```

---

## 综合项目：商品服务缓存系统

实现一个完整的商品服务缓存系统。

### 功能要求

1. **基础功能**
   - 商品信息缓存（支持 CRUD）
   - 商品列表缓存（分页）
   - 商品搜索结果缓存

2. **高级功能**
   - 多级缓存（L1 本地 + L2 Redis）
   - 布隆过滤器防穿透
   - singleflight 防击穿
   - 延迟双删保证一致性

3. **监控功能**
   - 缓存命中率统计
   - 缓存大小监控
   - 一致性检查

### 项目结构

```
product-cache/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── cache/
│   │   ├── redis.go
│   │   ├── local.go
│   │   ├── multi_level.go
│   │   └── bloom.go
│   ├── service/
│   │   └── product.go
│   ├── repository/
│   │   └── product.go
│   └── model/
│       └── product.go
├── pkg/
│   └── singleflight/
│       └── singleflight.go
└── go.mod
```

### 参考实现

```go
// internal/cache/multi_level.go
package cache

type ProductCache struct {
    l1           *cache.Cache
    l2           *redis.Client
    db           *gorm.DB
    bloom        *bloom.BloomFilter
    sf           singleflight.Group
    delayQueue   chan string
    stats        CacheStats
    mu           sync.RWMutex
}

type CacheStats struct {
    L1Hits       int64
    L1Misses     int64
    L2Hits       int64
    L2Misses     int64
    BloomRejects int64
    DBLoads      int64
}

func NewProductCache(redisAddr string, db *gorm.DB) *ProductCache {
    c := &ProductCache{
        l1:         cache.New(5*time.Minute, 10*time.Minute),
        l2:         redis.NewClient(&redis.Options{Addr: redisAddr}),
        db:         db,
        bloom:      bloom.NewWithEstimates(1_000_000, 0.0001),
        delayQueue: make(chan string, 10000),
    }

    c.initBloomFilter()
    go c.processDelayedDeletions()

    return c
}

func (c *ProductCache) GetProduct(ctx context.Context, id uint) (*Product, error) {
    key := fmt.Sprintf("product:%d", id)

    // 1. 布隆过滤器检查
    if !c.bloomCheck(id) {
        atomic.AddInt64(&c.stats.BloomRejects, 1)
        return nil, ErrNotFound
    }

    // 2. L1 本地缓存
    if val, found := c.l1.Get(key); found {
        atomic.AddInt64(&c.stats.L1Hits, 1)
        return val.(*Product), nil
    }
    atomic.AddInt64(&c.stats.L1Misses, 1)

    // 3. L2 Redis
    val, err := c.l2.Get(ctx, key).Result()
    if err == nil {
        atomic.AddInt64(&c.stats.L2Hits, 1)
        var product Product
        json.Unmarshal([]byte(val), &product)
        c.l1.Set(key, &product, cache.DefaultExpiration)
        return &product, nil
    }
    atomic.AddInt64(&c.stats.L2Misses, 1)

    // 4. singleflight 加载
    result, err, _ := c.sf.Do(key, func() (interface{}, error) {
        atomic.AddInt64(&c.stats.DBLoads, 1)

        var product Product
        if err := c.db.First(&product, id).Error; err != nil {
            return nil, err
        }

        // 回填缓存
        data, _ := json.Marshal(product)
        c.l2.Set(ctx, key, data, time.Hour)
        c.l1.Set(key, &product, 5*time.Minute)

        return &product, nil
    })

    if err != nil {
        return nil, err
    }
    return result.(*Product), nil
}

func (c *ProductCache) UpdateProduct(ctx context.Context, product *Product) error {
    key := fmt.Sprintf("product:%d", product.ID)

    // 1. 第一次删除
    c.l1.Delete(key)
    c.l2.Del(ctx, key)

    // 2. 更新数据库
    if err := c.db.Save(product).Error; err != nil {
        return err
    }

    // 3. 延迟删除
    c.delayQueue <- key

    return nil
}

func (c *ProductCache) Stats() CacheStats {
    return CacheStats{
        L1Hits:       atomic.LoadInt64(&c.stats.L1Hits),
        L1Misses:     atomic.LoadInt64(&c.stats.L1Misses),
        L2Hits:       atomic.LoadInt64(&c.stats.L2Hits),
        L2Misses:     atomic.LoadInt64(&c.stats.L2Misses),
        BloomRejects: atomic.LoadInt64(&c.stats.BloomRejects),
        DBLoads:      atomic.LoadInt64(&c.stats.DBLoads),
    }
}

func (c *ProductCache) L1HitRate() float64 {
    hits := atomic.LoadInt64(&c.stats.L1Hits)
    misses := atomic.LoadInt64(&c.stats.L1Misses)
    total := hits + misses
    if total == 0 {
        return 0
    }
    return float64(hits) / float64(total) * 100
}

// ... 其他方法
```

### 验收标准

1. 所有功能正常运行
2. 并发测试通过（1000 并发无数据竞争）
3. L1 命中率 > 80%
4. L1 + L2 总命中率 > 95%
5. 无缓存穿透、击穿、雪崩问题
6. 更新后数据最终一致

### 测试用例

```go
func TestProductCache(t *testing.T) {
    // 1. 测试基本 CRUD
    // 2. 测试缓存命中
    // 3. 测试穿透防护
    // 4. 测试并发访问
    // 5. 测试一致性
}

func BenchmarkProductCache(b *testing.B) {
    // 性能测试
}
```
