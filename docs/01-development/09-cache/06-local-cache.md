# 本地缓存 (Local Cache)

## 概述

本地缓存将数据存储在应用进程内存中，访问速度极快（纳秒级），但容量受限且多实例不共享。

```
┌─────────────────────────────────────────────────────────────┐
│                     本地缓存 vs 分布式缓存                    │
├─────────────────────────────────────────────────────────────┤
│  特性        │  本地缓存         │  分布式缓存 (Redis)      │
├─────────────────────────────────────────────────────────────┤
│  访问延迟    │  < 1μs            │  ~1ms                    │
│  容量        │  受限于进程内存    │  可水平扩展              │
│  多实例共享  │  否               │  是                      │
│  网络开销    │  无               │  有                      │
│  持久化      │  否               │  支持                    │
│  适用场景    │  热点数据、配置    │  会话、分布式数据        │
└─────────────────────────────────────────────────────────────┘
```

## go-cache

`go-cache` 是 Go 语言最常用的本地缓存库，简单易用，支持过期时间。

### 安装

```bash
go get github.com/patrickmn/go-cache
```

### 基本使用

```go
package main

import (
    "fmt"
    "time"

    "github.com/patrickmn/go-cache"
)

func main() {
    // 创建缓存：默认过期时间 5 分钟，清理间隔 10 分钟
    c := cache.New(5*time.Minute, 10*time.Minute)

    // 设置值（使用默认过期时间）
    c.Set("name", "Alice", cache.DefaultExpiration)

    // 设置值（指定过期时间）
    c.Set("session", "abc123", 30*time.Minute)

    // 设置永不过期的值
    c.Set("config", map[string]string{"env": "prod"}, cache.NoExpiration)

    // 获取值
    if val, found := c.Get("name"); found {
        fmt.Println("name:", val.(string))  // name: Alice
    }

    // 获取带类型的值
    if val, found := c.Get("session"); found {
        session := val.(string)
        fmt.Println("session:", session)
    }

    // 删除值
    c.Delete("name")

    // 检查是否存在
    _, found := c.Get("name")
    fmt.Println("name exists:", found)  // false

    // 获取所有项
    items := c.Items()
    fmt.Printf("Total items: %d\n", len(items))
}
```

### 与 JavaScript 对比

```javascript
// Node.js 本地缓存 (node-cache)
const NodeCache = require('node-cache');

// 创建缓存：默认 TTL 300 秒，检查周期 120 秒
const cache = new NodeCache({ stdTTL: 300, checkperiod: 120 });

// 设置
cache.set('name', 'Alice');
cache.set('session', 'abc123', 1800);  // 30 分钟

// 获取
const name = cache.get('name');
if (name !== undefined) {
    console.log('name:', name);
}

// 删除
cache.del('name');
```

### 高级操作

```go
// SetDefault - 使用默认过期时间
c.SetDefault("key", "value")

// Add - 仅当 key 不存在时设置
err := c.Add("key", "value", cache.DefaultExpiration)
if err != nil {
    // key 已存在
}

// Replace - 仅当 key 存在时替换
err = c.Replace("key", "new_value", cache.DefaultExpiration)
if err != nil {
    // key 不存在
}

// Increment/Decrement - 数值操作
c.Set("counter", 0, cache.NoExpiration)
c.Increment("counter", 1)
c.Decrement("counter", 1)

// IncrementFloat
c.Set("balance", 100.0, cache.NoExpiration)
c.IncrementFloat("balance", 10.5)

// DeleteExpired - 手动清理过期项
c.DeleteExpired()

// Flush - 清空所有
c.Flush()

// ItemCount - 获取项数
count := c.ItemCount()
```

### 持久化

```go
// 保存到文件
if err := c.SaveFile("cache.gob"); err != nil {
    log.Fatal(err)
}

// 从文件加载
if err := c.LoadFile("cache.gob"); err != nil {
    log.Fatal(err)
}

// 保存到 io.Writer
var buf bytes.Buffer
if err := c.Save(&buf); err != nil {
    log.Fatal(err)
}

// 从 io.Reader 加载
if err := c.Load(&buf); err != nil {
    log.Fatal(err)
}
```

## 封装本地缓存

### 泛型缓存封装

```go
type LocalCache[T any] struct {
    cache *cache.Cache
}

func NewLocalCache[T any](defaultExpiration, cleanupInterval time.Duration) *LocalCache[T] {
    return &LocalCache[T]{
        cache: cache.New(defaultExpiration, cleanupInterval),
    }
}

func (c *LocalCache[T]) Set(key string, value T, expiration time.Duration) {
    c.cache.Set(key, value, expiration)
}

func (c *LocalCache[T]) Get(key string) (T, bool) {
    var zero T
    val, found := c.cache.Get(key)
    if !found {
        return zero, false
    }
    return val.(T), true
}

func (c *LocalCache[T]) Delete(key string) {
    c.cache.Delete(key)
}

func (c *LocalCache[T]) GetOrSet(key string, loader func() (T, error), expiration time.Duration) (T, error) {
    if val, found := c.Get(key); found {
        return val, nil
    }

    value, err := loader()
    if err != nil {
        var zero T
        return zero, err
    }

    c.Set(key, value, expiration)
    return value, nil
}

// 使用示例
func main() {
    userCache := NewLocalCache[*User](5*time.Minute, 10*time.Minute)

    // 设置
    userCache.Set("user:1", &User{ID: 1, Name: "Alice"}, cache.DefaultExpiration)

    // 获取（无需类型断言）
    if user, found := userCache.Get("user:1"); found {
        fmt.Println(user.Name)
    }

    // 获取或加载
    user, err := userCache.GetOrSet("user:2", func() (*User, error) {
        // 从数据库加载
        return &User{ID: 2, Name: "Bob"}, nil
    }, 5*time.Minute)
}
```

### 带锁的缓存封装

```go
type SafeLocalCache struct {
    cache *cache.Cache
    locks sync.Map  // key -> *sync.Mutex
}

func NewSafeLocalCache(defaultExpiration, cleanupInterval time.Duration) *SafeLocalCache {
    return &SafeLocalCache{
        cache: cache.New(defaultExpiration, cleanupInterval),
    }
}

func (c *SafeLocalCache) getLock(key string) *sync.Mutex {
    lock, _ := c.locks.LoadOrStore(key, &sync.Mutex{})
    return lock.(*sync.Mutex)
}

func (c *SafeLocalCache) GetOrLoad(key string, loader func() (interface{}, error), expiration time.Duration) (interface{}, error) {
    // 1. 尝试从缓存获取
    if val, found := c.cache.Get(key); found {
        return val, nil
    }

    // 2. 获取锁
    lock := c.getLock(key)
    lock.Lock()
    defer lock.Unlock()

    // 3. 双重检查
    if val, found := c.cache.Get(key); found {
        return val, nil
    }

    // 4. 加载数据
    val, err := loader()
    if err != nil {
        return nil, err
    }

    // 5. 写入缓存
    c.cache.Set(key, val, expiration)
    return val, nil
}
```

## bigcache

`bigcache` 是一个高性能本地缓存库，专为大数据量和高并发场景设计。

### 安装

```bash
go get github.com/allegro/bigcache/v3
```

### 基本使用

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/allegro/bigcache/v3"
)

func main() {
    ctx := context.Background()

    // 创建缓存
    cache, err := bigcache.New(ctx, bigcache.DefaultConfig(10*time.Minute))
    if err != nil {
        panic(err)
    }
    defer cache.Close()

    // 设置（只支持 []byte）
    user := User{ID: 1, Name: "Alice"}
    data, _ := json.Marshal(user)
    cache.Set("user:1", data)

    // 获取
    data, err = cache.Get("user:1")
    if err == nil {
        var u User
        json.Unmarshal(data, &u)
        fmt.Printf("User: %+v\n", u)
    }

    // 删除
    cache.Delete("user:1")
}

type User struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}
```

### bigcache 配置

```go
config := bigcache.Config{
    // 分片数（必须是 2 的幂）
    Shards: 1024,

    // 条目生存时间
    LifeWindow: 10 * time.Minute,

    // 清理间隔
    CleanWindow: 5 * time.Minute,

    // 每个分片的最大条目数
    MaxEntriesInWindow: 1000 * 10 * 60,

    // 最大条目大小（字节）
    MaxEntrySize: 500,

    // 详细日志
    Verbose: false,

    // 初始分片大小
    HardMaxCacheSize: 256,  // MB

    // 条目被删除时的回调
    OnRemove: func(key string, entry []byte) {
        fmt.Printf("Key %s removed\n", key)
    },

    // 条目被删除的原因回调
    OnRemoveWithReason: func(key string, entry []byte, reason bigcache.RemoveReason) {
        fmt.Printf("Key %s removed, reason: %v\n", key, reason)
    },
}

cache, err := bigcache.New(context.Background(), config)
```

### bigcache vs go-cache

| 特性 | go-cache | bigcache |
|------|----------|----------|
| 数据类型 | interface{} | []byte |
| GC 影响 | 大 | 小 |
| 内存效率 | 一般 | 高 |
| 适用场景 | 小规模缓存 | 大规模缓存 |
| 易用性 | 简单 | 需要序列化 |

## ristretto

`ristretto` 是 Dgraph 开发的高性能缓存库，支持 LFU 淘汰策略。

### 安装

```bash
go get github.com/dgraph-io/ristretto
```

### 基本使用

```go
package main

import (
    "fmt"
    "time"

    "github.com/dgraph-io/ristretto"
)

func main() {
    // 创建缓存
    cache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: 1e7,     // 跟踪频率的 key 数量
        MaxCost:     1 << 30, // 最大缓存大小（字节）
        BufferItems: 64,      // 每个 Get buffer 的 key 数量
    })
    if err != nil {
        panic(err)
    }
    defer cache.Close()

    // 设置（cost 表示占用的容量）
    cache.Set("key", "value", 1)

    // 等待值被存储
    time.Sleep(10 * time.Millisecond)

    // 获取
    if val, found := cache.Get("key"); found {
        fmt.Println("value:", val)
    }

    // 设置带 TTL
    cache.SetWithTTL("session", "abc123", 1, time.Hour)

    // 删除
    cache.Del("key")

    // 清空
    cache.Clear()

    // 统计
    fmt.Printf("Hits: %d, Misses: %d\n", cache.Metrics.Hits(), cache.Metrics.Misses())
}
```

### ristretto 特性

```go
// 1. 基于 cost 的淘汰
cache.Set("small", []byte("hello"), 5)      // cost = 5
cache.Set("large", []byte("world"*1000), 5000)  // cost = 5000

// 2. 通过 OnEvict 监听淘汰事件
cache, _ := ristretto.NewCache(&ristretto.Config{
    NumCounters: 1e7,
    MaxCost:     1 << 30,
    BufferItems: 64,
    OnEvict: func(item *ristretto.Item) {
        fmt.Printf("Evicted key: %v\n", item.Key)
    },
})

// 3. 获取统计信息
metrics := cache.Metrics
fmt.Printf("Hits: %d\n", metrics.Hits())
fmt.Printf("Misses: %d\n", metrics.Misses())
fmt.Printf("Hit Rate: %.2f%%\n", metrics.Ratio()*100)
fmt.Printf("Keys Added: %d\n", metrics.KeysAdded())
fmt.Printf("Cost Added: %d\n", metrics.CostAdded())
```

## 缓存淘汰策略

### LRU (Least Recently Used)

淘汰最近最少使用的数据。

```go
import "github.com/hashicorp/golang-lru/v2"

// 创建 LRU 缓存
cache, _ := lru.New[string, *User](1000)

// 设置
cache.Add("user:1", &User{ID: 1, Name: "Alice"})

// 获取
if user, ok := cache.Get("user:1"); ok {
    fmt.Println(user.Name)
}

// 最近使用的 key 会被保留，最少使用的会被淘汰
```

### LFU (Least Frequently Used)

淘汰使用频率最低的数据。

```go
// ristretto 使用 TinyLFU 策略
cache, _ := ristretto.NewCache(&ristretto.Config{
    NumCounters: 1e7,     // 用于跟踪访问频率
    MaxCost:     1 << 30,
    BufferItems: 64,
})
```

### TTL (Time To Live)

按过期时间淘汰数据。

```go
// go-cache 使用 TTL 策略
c := cache.New(5*time.Minute, 10*time.Minute)
c.Set("key", "value", 30*time.Second)  // 30 秒后过期
```

## 实际应用

### 配置缓存

```go
type ConfigCache struct {
    cache *cache.Cache
    db    *gorm.DB
}

func NewConfigCache(db *gorm.DB) *ConfigCache {
    c := &ConfigCache{
        cache: cache.New(10*time.Minute, 30*time.Minute),
        db:    db,
    }

    // 启动时预加载配置
    c.preload()

    return c
}

func (c *ConfigCache) preload() {
    var configs []Config
    c.db.Find(&configs)

    for _, cfg := range configs {
        c.cache.Set(cfg.Key, cfg.Value, cache.NoExpiration)
    }
}

func (c *ConfigCache) Get(key string) (string, bool) {
    if val, found := c.cache.Get(key); found {
        return val.(string), true
    }
    return "", false
}

func (c *ConfigCache) Reload() {
    c.cache.Flush()
    c.preload()
}
```

### 用户会话缓存

```go
type SessionCache struct {
    cache *cache.Cache
}

func NewSessionCache() *SessionCache {
    return &SessionCache{
        cache: cache.New(30*time.Minute, 5*time.Minute),
    }
}

type Session struct {
    UserID    uint
    Username  string
    CreatedAt time.Time
}

func (c *SessionCache) Set(token string, session *Session) {
    c.cache.Set(token, session, cache.DefaultExpiration)
}

func (c *SessionCache) Get(token string) (*Session, bool) {
    if val, found := c.cache.Get(token); found {
        return val.(*Session), true
    }
    return nil, false
}

func (c *SessionCache) Delete(token string) {
    c.cache.Delete(token)
}

func (c *SessionCache) Refresh(token string) {
    if session, found := c.Get(token); found {
        c.cache.Set(token, session, cache.DefaultExpiration)
    }
}
```

### API 响应缓存

```go
type ResponseCache struct {
    cache *cache.Cache
}

type CachedResponse struct {
    StatusCode int
    Headers    map[string]string
    Body       []byte
}

func (c *ResponseCache) Middleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // 只缓存 GET 请求
        if ctx.Request.Method != "GET" {
            ctx.Next()
            return
        }

        key := ctx.Request.URL.String()

        // 检查缓存
        if cached, found := c.cache.Get(key); found {
            resp := cached.(*CachedResponse)
            for k, v := range resp.Headers {
                ctx.Header(k, v)
            }
            ctx.Header("X-Cache", "HIT")
            ctx.Data(resp.StatusCode, "application/json", resp.Body)
            ctx.Abort()
            return
        }

        // 使用 ResponseWriter 包装器
        w := &responseWriter{ResponseWriter: ctx.Writer}
        ctx.Writer = w

        ctx.Next()

        // 缓存成功响应
        if ctx.Writer.Status() == 200 {
            cached := &CachedResponse{
                StatusCode: ctx.Writer.Status(),
                Body:       w.body.Bytes(),
            }
            c.cache.Set(key, cached, 5*time.Minute)
        }
    }
}
```

## 最佳实践

### 1. 选择合适的缓存库

| 场景 | 推荐库 |
|------|--------|
| 小规模、简单场景 | go-cache |
| 大规模、低 GC | bigcache |
| 需要 LFU 淘汰 | ristretto |
| 需要精确 LRU | golang-lru |

### 2. 设置合理的容量

```go
// 根据数据大小估算
// 假设每条数据平均 1KB，要缓存 10000 条
// 需要约 10MB 内存

// bigcache 配置
config := bigcache.DefaultConfig(10 * time.Minute)
config.HardMaxCacheSize = 16  // 16MB

// ristretto 配置
cache, _ := ristretto.NewCache(&ristretto.Config{
    NumCounters: 1e5,        // 10万个计数器
    MaxCost:     16 << 20,   // 16MB
    BufferItems: 64,
})
```

### 3. 监控缓存指标

```go
func (c *MonitoredCache) recordMetrics() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        // 记录缓存大小
        prometheus.CacheSize.Set(float64(c.cache.ItemCount()))

        // 记录命中率（如果缓存库支持）
        // prometheus.CacheHitRate.Set(cache.Metrics.Ratio())
    }
}
```

### 4. 处理并发

```go
// go-cache 本身是并发安全的
c := cache.New(5*time.Minute, 10*time.Minute)

// 但复杂操作需要额外同步
var mu sync.Mutex

func updateCounter(key string) {
    mu.Lock()
    defer mu.Unlock()

    val, _ := c.Get(key)
    count := val.(int) + 1
    c.Set(key, count, cache.DefaultExpiration)
}

// 更好的方式：使用 Increment
c.Set("counter", 0, cache.NoExpiration)
c.Increment("counter", 1)
```

**下一节**：[多级缓存](./07-multi-level.md) - 学习 L1/L2 缓存架构
