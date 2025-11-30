# 第九章：缓存策略 (Caching Strategies)

## 章节概述

缓存是提升应用性能的关键技术，本章学习 Redis 和本地缓存的使用。

## 学习目标

完成本章后，你将能够：

- 理解缓存的核心概念和使用场景
- 使用 Redis 实现分布式缓存
- 使用 go-cache 实现本地缓存
- 掌握常见缓存模式和策略
- 解决缓存穿透、击穿、雪崩问题
- 实现多级缓存架构

## 与 JavaScript 对比

### Node.js + Redis

```javascript
import Redis from 'ioredis';

const redis = new Redis();

// 设置缓存
await redis.set('user:1', JSON.stringify(user), 'EX', 3600);

// 获取缓存
const cached = await redis.get('user:1');
if (cached) {
    return JSON.parse(cached);
}

// 删除缓存
await redis.del('user:1');
```

### Go + Redis

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 设置缓存
err := rdb.Set(ctx, "user:1", userJSON, time.Hour).Err()

// 获取缓存
val, err := rdb.Get(ctx, "user:1").Result()
if err == redis.Nil {
    // 缓存不存在
}

// 删除缓存
rdb.Del(ctx, "user:1")
```

## 章节内容

| 文档 | 主题 | 描述 |
|------|------|------|
| [01-cache-intro.md](./01-cache-intro.md) | 缓存概述 | 缓存概念和使用场景 |
| [02-redis-basics.md](./02-redis-basics.md) | Redis 基础 | 安装和基本命令 |
| [03-go-redis.md](./03-go-redis.md) | go-redis | Go Redis 客户端 |
| [04-cache-patterns.md](./04-cache-patterns.md) | 缓存模式 | Cache-Aside, Write-Through |
| [05-cache-problems.md](./05-cache-problems.md) | 缓存问题 | 穿透、击穿、雪崩 |
| [06-local-cache.md](./06-local-cache.md) | 本地缓存 | go-cache 使用 |
| [07-multi-level.md](./07-multi-level.md) | 多级缓存 | L1/L2 缓存架构 |
| [08-cache-consistency.md](./08-cache-consistency.md) | 缓存一致性 | 数据库与缓存同步 |
| [exercises.md](./exercises.md) | 练习 | 实践练习 |

## 快速开始

### 安装 Redis

```bash
# macOS
brew install redis
brew services start redis

# Docker
docker run -d --name redis -p 6379:6379 redis:7-alpine

# 验证
redis-cli ping
# PONG
```

### 安装 Go Redis 客户端

```bash
go get github.com/redis/go-redis/v9
```

### 基本示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type User struct {
    ID    uint   `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    ctx := context.Background()

    // 连接 Redis
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    // 测试连接
    if err := rdb.Ping(ctx).Err(); err != nil {
        panic(err)
    }

    // 设置字符串
    rdb.Set(ctx, "greeting", "Hello, Redis!", time.Hour)

    // 获取字符串
    val, _ := rdb.Get(ctx, "greeting").Result()
    fmt.Println(val)  // Hello, Redis!

    // 缓存 JSON 对象
    user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
    userJSON, _ := json.Marshal(user)
    rdb.Set(ctx, "user:1", userJSON, time.Hour)

    // 获取 JSON 对象
    cached, _ := rdb.Get(ctx, "user:1").Result()
    var cachedUser User
    json.Unmarshal([]byte(cached), &cachedUser)
    fmt.Printf("Cached user: %+v\n", cachedUser)

    // 删除
    rdb.Del(ctx, "greeting", "user:1")
}
```

## 缓存核心概念

### 为什么使用缓存

```
┌─────────┐     ┌─────────┐     ┌──────────┐
│ Client  │────▶│  Cache  │────▶│ Database │
└─────────┘     └─────────┘     └──────────┘
                    │
              ┌─────┴─────┐
              │           │
         Cache Hit   Cache Miss
         (< 1ms)     (10-100ms)
```

| 场景 | 无缓存 | 有缓存 |
|------|--------|--------|
| 数据库查询 | 10-100ms | < 1ms |
| API 响应 | 100-500ms | 10-50ms |
| 并发能力 | 100 QPS | 10000+ QPS |

### 缓存类型

| 类型 | 位置 | 速度 | 容量 | 共享 |
|------|------|------|------|------|
| CPU Cache | CPU | 极快 | 很小 | 否 |
| 本地缓存 | 进程内存 | 很快 | 受限 | 否 |
| 分布式缓存 | Redis/Memcached | 快 | 大 | 是 |
| CDN | 边缘节点 | 快 | 大 | 是 |

### 缓存策略

| 策略 | 描述 | 适用场景 |
|------|------|----------|
| Cache-Aside | 应用管理缓存 | 通用场景 |
| Read-Through | 缓存管理读取 | 读多写少 |
| Write-Through | 同步写入缓存和数据库 | 数据一致性要求高 |
| Write-Behind | 异步写入数据库 | 写多读少 |
| Refresh-Ahead | 提前刷新热点数据 | 热点数据 |

## 常见问题

### 缓存穿透

```go
// 问题：查询不存在的数据，每次都打到数据库
// 解决：缓存空值或使用布隆过滤器

func GetUser(id uint) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    // 查缓存
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        if val == "null" {
            return nil, nil  // 空值缓存
        }
        var user User
        json.Unmarshal([]byte(val), &user)
        return &user, nil
    }

    // 查数据库
    user, err := db.FindUser(id)
    if err != nil {
        return nil, err
    }

    if user == nil {
        // 缓存空值，短过期时间
        rdb.Set(ctx, key, "null", 5*time.Minute)
        return nil, nil
    }

    // 缓存结果
    userJSON, _ := json.Marshal(user)
    rdb.Set(ctx, key, userJSON, time.Hour)
    return user, nil
}
```

### 缓存击穿

```go
// 问题：热点 key 过期，大量请求同时打到数据库
// 解决：互斥锁或永不过期

func GetHotData(key string) (string, error) {
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        return val, nil
    }

    // 获取分布式锁
    lockKey := key + ":lock"
    ok, _ := rdb.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
    if ok {
        defer rdb.Del(ctx, lockKey)

        // 查询数据库并更新缓存
        data := fetchFromDB(key)
        rdb.Set(ctx, key, data, time.Hour)
        return data, nil
    }

    // 等待并重试
    time.Sleep(100 * time.Millisecond)
    return GetHotData(key)
}
```

### 缓存雪崩

```go
// 问题：大量 key 同时过期
// 解决：随机过期时间

func SetWithRandomExpiry(key string, value interface{}) error {
    // 基础过期时间 1 小时，加上 0-10 分钟随机
    baseExpiry := time.Hour
    randomExpiry := time.Duration(rand.Intn(600)) * time.Second
    expiry := baseExpiry + randomExpiry

    data, _ := json.Marshal(value)
    return rdb.Set(ctx, key, data, expiry).Err()
}
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 设置过期时间 | 避免数据永久占用内存 |
| 使用连接池 | 复用连接，提高性能 |
| 序列化选择 | JSON 可读，Protobuf 高效 |
| Key 命名规范 | 使用 `模块:实体:ID` 格式 |
| 监控指标 | 命中率、内存使用、延迟 |
| 容量规划 | 预估数据量，避免 OOM |

**下一节**：[缓存概述](./01-cache-intro.md) - 深入理解缓存概念
