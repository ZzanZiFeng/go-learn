# 缓存概述 (Cache Introduction)

## 什么是缓存

缓存是一种将数据存储在快速存储介质中的技术，用于加速后续访问。

```
┌─────────────────────────────────────────────────────────────┐
│                      数据访问层次                              │
├─────────────────────────────────────────────────────────────┤
│  L1 Cache (CPU)     │  ~1ns      │  64KB         │  最快    │
│  L2 Cache (CPU)     │  ~4ns      │  256KB        │          │
│  L3 Cache (CPU)     │  ~10ns     │  8MB          │          │
│  RAM                │  ~100ns    │  16-128GB     │          │
│  本地缓存 (进程内)   │  ~1μs      │  MB-GB        │          │
│  Redis/Memcached    │  ~1ms      │  GB-TB        │          │
│  SSD                │  ~100μs    │  TB           │          │
│  HDD                │  ~10ms     │  TB           │          │
│  网络数据库         │  ~10-100ms │  无限         │  最慢    │
└─────────────────────────────────────────────────────────────┘
```

## 为什么需要缓存

### 性能提升

```go
// 无缓存：每次查询数据库
func GetUserWithoutCache(id uint) (*User, error) {
    var user User
    err := db.First(&user, id).Error  // 10-100ms
    return &user, err
}

// 有缓存：优先查缓存
func GetUserWithCache(id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 1. 查缓存 (~1ms)
    if cached, err := rdb.Get(ctx, key).Result(); err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 2. 查数据库 (10-100ms)
    var user User
    if err := db.First(&user, id).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存
    data, _ := json.Marshal(user)
    rdb.Set(ctx, key, data, time.Hour)

    return &user, nil
}
```

### 减轻数据库压力

```
无缓存:
┌────────┐     ┌──────────┐
│ 1000   │────▶│ Database │  1000 QPS
│ 请求/秒 │     │ (压力大) │
└────────┘     └──────────┘

有缓存 (90% 命中率):
┌────────┐     ┌─────────┐     ┌──────────┐
│ 1000   │────▶│  Cache  │────▶│ Database │  100 QPS
│ 请求/秒 │     │ (900次) │     │ (压力小) │
└────────┘     └─────────┘     └──────────┘
```

### 成本节省

| 存储类型 | 每 GB 成本 | 读取速度 |
|----------|-----------|----------|
| Redis 内存 | $$$$ | 极快 |
| SSD | $$ | 快 |
| HDD | $ | 慢 |
| 网络存储 | $ | 最慢 |

合理使用缓存可以用少量高速存储替代大量慢速存储。

## 缓存使用场景

### 适合缓存的数据

| 场景 | 特点 | 示例 |
|------|------|------|
| 读多写少 | 写入少，读取频繁 | 用户信息、商品详情 |
| 计算结果 | 计算耗时，结果可复用 | 排行榜、统计数据 |
| 会话数据 | 需要快速访问 | 用户登录状态 |
| 热点数据 | 访问量大 | 首页推荐、热门文章 |
| 配置数据 | 很少变化 | 系统配置、字典数据 |

### 不适合缓存的数据

| 场景 | 原因 |
|------|------|
| 写多读少 | 缓存频繁失效，收益低 |
| 实时性要求高 | 缓存有延迟，无法保证实时 |
| 数据量大且访问分散 | 缓存命中率低 |
| 敏感数据 | 缓存安全风险 |

## 缓存类型

### 本地缓存 (Local Cache)

```go
import "github.com/patrickmn/go-cache"

// 创建本地缓存
c := cache.New(5*time.Minute, 10*time.Minute)

// 设置
c.Set("key", "value", cache.DefaultExpiration)

// 获取
if val, found := c.Get("key"); found {
    fmt.Println(val)
}

// 删除
c.Delete("key")
```

**优点**：
- 速度极快（纳秒级）
- 无网络开销
- 简单易用

**缺点**：
- 容量受限于进程内存
- 多实例不共享
- 进程重启数据丢失

### 分布式缓存 (Distributed Cache)

```go
import "github.com/redis/go-redis/v9"

// 创建 Redis 客户端
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 设置
rdb.Set(ctx, "key", "value", time.Hour)

// 获取
val, err := rdb.Get(ctx, "key").Result()

// 删除
rdb.Del(ctx, "key")
```

**优点**：
- 容量大（可扩展）
- 多实例共享
- 持久化支持

**缺点**：
- 网络延迟（毫秒级）
- 需要额外部署维护
- 序列化/反序列化开销

### 多级缓存 (Multi-Level Cache)

```go
// L1: 本地缓存 (最快，容量小)
// L2: Redis (快，容量大)
// L3: 数据库 (慢，容量无限)

func Get(key string) (string, error) {
    // L1: 本地缓存
    if val, found := localCache.Get(key); found {
        return val.(string), nil
    }

    // L2: Redis
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        localCache.Set(key, val, time.Minute)  // 回填 L1
        return val, nil
    }

    // L3: 数据库
    data, err := db.Query(key)
    if err != nil {
        return "", err
    }

    // 回填 L2 和 L1
    rdb.Set(ctx, key, data, time.Hour)
    localCache.Set(key, data, time.Minute)

    return data, nil
}
```

## 缓存指标

### 命中率 (Hit Rate)

```go
type CacheStats struct {
    Hits   int64
    Misses int64
}

func (s *CacheStats) HitRate() float64 {
    total := s.Hits + s.Misses
    if total == 0 {
        return 0
    }
    return float64(s.Hits) / float64(total) * 100
}

// 命中率目标:
// - 用户信息: 95%+
// - 商品详情: 90%+
// - 搜索结果: 80%+
```

### 缓存大小

```go
// Redis 内存使用
info, _ := rdb.Info(ctx, "memory").Result()
// used_memory: 1234567
// used_memory_human: 1.18M

// 监控内存使用，避免 OOM
```

### 延迟

```go
// 测量缓存延迟
start := time.Now()
rdb.Get(ctx, "key")
latency := time.Since(start)

// 正常延迟:
// - 本地缓存: < 1μs
// - Redis 本地: < 1ms
// - Redis 远程: < 10ms
```

## Key 设计规范

### 命名规范

```go
// 格式: 业务:实体:ID[:字段]

// 用户
"user:1"              // 用户 1 的完整信息
"user:1:profile"      // 用户 1 的 profile
"user:email:alice@x"  // 按邮箱索引

// 文章
"post:123"            // 文章 123
"post:123:views"      // 文章 123 的浏览量
"post:list:page:1"    // 文章列表第 1 页

// 会话
"session:abc123"      // 会话 ID

// 分布式锁
"lock:order:456"      // 订单 456 的锁
```

### Key 长度

```go
// ✅ 好的 key 设计
"u:1"        // 简短但不清晰
"user:1"     // 适中，推荐
"user:id:1"  // 稍长但清晰

// ❌ 过长的 key
"application:module:user:entity:id:1:field:profile"
```

### 避免大 Key

```go
// ❌ 大 Key (单个值过大)
// 缓存整个列表
rdb.Set(ctx, "all_users", hugeUserList, time.Hour)

// ✅ 分片存储
for i, chunk := range chunks(users, 100) {
    key := fmt.Sprintf("users:chunk:%d", i)
    rdb.Set(ctx, key, chunk, time.Hour)
}

// ✅ 或使用 Hash
for _, user := range users {
    rdb.HSet(ctx, "users", strconv.Itoa(user.ID), userJSON)
}
```

## 序列化选择

### JSON

```go
// 优点: 可读性好，调试方便
// 缺点: 体积大，性能一般

user := User{ID: 1, Name: "Alice"}
data, _ := json.Marshal(user)
// {"id":1,"name":"Alice"}
```

### MessagePack

```go
import "github.com/vmihailenco/msgpack/v5"

// 优点: 体积小，性能好
// 缺点: 不可读

data, _ := msgpack.Marshal(user)
```

### Protocol Buffers

```go
// 优点: 体积最小，性能最好，强类型
// 缺点: 需要定义 .proto 文件

data, _ := proto.Marshal(user)
```

### 选择建议

| 场景 | 推荐 |
|------|------|
| 开发调试 | JSON |
| 生产环境 | MessagePack 或 Protobuf |
| 跨语言 | JSON 或 Protobuf |
| 极致性能 | Protobuf |

**下一节**：[Redis 基础](./02-redis-basics.md) - 学习 Redis 安装和基本命令
