# go-redis 客户端 (go-redis Client)

## 概述

go-redis 是 Go 语言最流行的 Redis 客户端，支持 Redis 7.x 及所有特性。

## 安装

```bash
go get github.com/redis/go-redis/v9
```

## 连接

### 单机连接

```go
import (
    "context"
    "github.com/redis/go-redis/v9"
)

func main() {
    ctx := context.Background()

    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",  // 无密码
        DB:       0,   // 默认数据库
    })

    // 测试连接
    pong, err := rdb.Ping(ctx).Result()
    if err != nil {
        panic(err)
    }
    fmt.Println(pong) // PONG
}
```

### 连接选项

```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "password",
    DB:       0,

    // 连接池配置
    PoolSize:     10,               // 连接池大小
    MinIdleConns: 5,                // 最小空闲连接数
    PoolTimeout:  4 * time.Second,  // 获取连接超时

    // 超时配置
    DialTimeout:  5 * time.Second,  // 建立连接超时
    ReadTimeout:  3 * time.Second,  // 读取超时
    WriteTimeout: 3 * time.Second,  // 写入超时

    // 重试
    MaxRetries:      3,
    MinRetryBackoff: 8 * time.Millisecond,
    MaxRetryBackoff: 512 * time.Millisecond,

    // TLS
    // TLSConfig: &tls.Config{...},
})
```

### 使用 URL 连接

```go
// redis://user:password@localhost:6379/0?dial_timeout=3s
opt, err := redis.ParseURL("redis://localhost:6379/0")
if err != nil {
    panic(err)
}

rdb := redis.NewClient(opt)
```

### 集群连接

```go
rdb := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "node1:6379",
        "node2:6379",
        "node3:6379",
    },
    Password: "password",

    // 读取模式
    ReadOnly:       true,  // 只读从节点
    RouteRandomly:  true,  // 随机路由
    RouteByLatency: true,  // 按延迟路由
})
```

### Sentinel 连接

```go
rdb := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{
        "sentinel1:26379",
        "sentinel2:26379",
        "sentinel3:26379",
    },
    Password:         "password",
    SentinelPassword: "sentinel-password",
})
```

## 基本操作

### String 操作

```go
ctx := context.Background()

// Set
err := rdb.Set(ctx, "key", "value", 0).Err()  // 永不过期
err = rdb.Set(ctx, "key", "value", time.Hour).Err()  // 1小时过期

// Get
val, err := rdb.Get(ctx, "key").Result()
if err == redis.Nil {
    fmt.Println("key does not exist")
} else if err != nil {
    panic(err)
} else {
    fmt.Println("key:", val)
}

// SetNX (仅当不存在时设置)
ok, err := rdb.SetNX(ctx, "key", "value", time.Hour).Result()
if ok {
    fmt.Println("key was set")
}

// SetXX (仅当存在时设置)
ok, err = rdb.SetXX(ctx, "key", "new_value", 0).Result()

// GetSet (设置并返回旧值)
oldVal, err := rdb.GetSet(ctx, "key", "new_value").Result()

// MSet/MGet
err = rdb.MSet(ctx, "key1", "val1", "key2", "val2").Err()
vals, err := rdb.MGet(ctx, "key1", "key2").Result()

// Incr/Decr
n, err := rdb.Incr(ctx, "counter").Result()
n, err = rdb.IncrBy(ctx, "counter", 10).Result()
f, err := rdb.IncrByFloat(ctx, "counter", 1.5).Result()
n, err = rdb.Decr(ctx, "counter").Result()
```

### Hash 操作

```go
// HSet
err := rdb.HSet(ctx, "user:1", "name", "Alice").Err()
err = rdb.HSet(ctx, "user:1", map[string]interface{}{
    "name":  "Alice",
    "age":   25,
    "email": "alice@example.com",
}).Err()

// HGet
name, err := rdb.HGet(ctx, "user:1", "name").Result()

// HGetAll
fields, err := rdb.HGetAll(ctx, "user:1").Result()
for key, val := range fields {
    fmt.Printf("%s: %s\n", key, val)
}

// HMGet
vals, err := rdb.HMGet(ctx, "user:1", "name", "age").Result()

// HExists
exists, err := rdb.HExists(ctx, "user:1", "name").Result()

// HDel
err = rdb.HDel(ctx, "user:1", "email").Err()

// HIncrBy
n, err := rdb.HIncrBy(ctx, "user:1", "age", 1).Result()

// HScan (扫描大 Hash)
var cursor uint64
for {
    keys, cursor, err := rdb.HScan(ctx, "user:1", cursor, "*", 10).Result()
    if err != nil {
        panic(err)
    }
    for i := 0; i < len(keys); i += 2 {
        fmt.Printf("%s: %s\n", keys[i], keys[i+1])
    }
    if cursor == 0 {
        break
    }
}
```

### List 操作

```go
// LPush/RPush
err := rdb.LPush(ctx, "tasks", "task1", "task2").Err()
err = rdb.RPush(ctx, "tasks", "task3").Err()

// LPop/RPop
val, err := rdb.LPop(ctx, "tasks").Result()
val, err = rdb.RPop(ctx, "tasks").Result()

// BLPop/BRPop (阻塞)
result, err := rdb.BLPop(ctx, 10*time.Second, "tasks").Result()
// result[0] = key, result[1] = value

// LRange
vals, err := rdb.LRange(ctx, "tasks", 0, -1).Result()

// LLen
length, err := rdb.LLen(ctx, "tasks").Result()

// LIndex
val, err = rdb.LIndex(ctx, "tasks", 0).Result()

// LTrim
err = rdb.LTrim(ctx, "tasks", 0, 99).Err()
```

### Set 操作

```go
// SAdd
err := rdb.SAdd(ctx, "tags", "go", "redis", "cache").Err()

// SMembers
members, err := rdb.SMembers(ctx, "tags").Result()

// SIsMember
isMember, err := rdb.SIsMember(ctx, "tags", "go").Result()

// SCard
count, err := rdb.SCard(ctx, "tags").Result()

// SRem
err = rdb.SRem(ctx, "tags", "cache").Err()

// SInter/SUnion/SDiff
inter, err := rdb.SInter(ctx, "set1", "set2").Result()
union, err := rdb.SUnion(ctx, "set1", "set2").Result()
diff, err := rdb.SDiff(ctx, "set1", "set2").Result()
```

### Sorted Set 操作

```go
// ZAdd
err := rdb.ZAdd(ctx, "leaderboard", redis.Z{
    Score:  100,
    Member: "Alice",
}).Err()

err = rdb.ZAdd(ctx, "leaderboard",
    redis.Z{Score: 200, Member: "Bob"},
    redis.Z{Score: 150, Member: "Charlie"},
).Err()

// ZRange (升序)
members, err := rdb.ZRange(ctx, "leaderboard", 0, -1).Result()

// ZRevRange (降序)
members, err = rdb.ZRevRange(ctx, "leaderboard", 0, -1).Result()

// ZRangeWithScores
results, err := rdb.ZRevRangeWithScores(ctx, "leaderboard", 0, -1).Result()
for _, z := range results {
    fmt.Printf("%s: %.0f\n", z.Member, z.Score)
}

// ZScore
score, err := rdb.ZScore(ctx, "leaderboard", "Alice").Result()

// ZRank/ZRevRank
rank, err := rdb.ZRevRank(ctx, "leaderboard", "Alice").Result()

// ZIncrBy
newScore, err := rdb.ZIncrBy(ctx, "leaderboard", 50, "Alice").Result()

// ZRangeByScore
members, err = rdb.ZRangeByScore(ctx, "leaderboard", &redis.ZRangeBy{
    Min: "100",
    Max: "200",
}).Result()
```

## Key 操作

```go
// Exists
n, err := rdb.Exists(ctx, "key1", "key2").Result()

// Del
n, err = rdb.Del(ctx, "key1", "key2").Result()

// Expire
ok, err := rdb.Expire(ctx, "key", time.Hour).Result()

// TTL
ttl, err := rdb.TTL(ctx, "key").Result()
if ttl == -1 {
    fmt.Println("key has no expiration")
} else if ttl == -2 {
    fmt.Println("key does not exist")
}

// Keys (生产环境慎用)
keys, err := rdb.Keys(ctx, "user:*").Result()

// Scan (安全扫描)
var cursor uint64
var keys []string
for {
    var batch []string
    batch, cursor, err = rdb.Scan(ctx, cursor, "user:*", 100).Result()
    if err != nil {
        panic(err)
    }
    keys = append(keys, batch...)
    if cursor == 0 {
        break
    }
}

// Type
keyType, err := rdb.Type(ctx, "key").Result()
```

## Pipeline

```go
// Pipeline 批量执行（不保证原子性）
pipe := rdb.Pipeline()

incr := pipe.Incr(ctx, "counter")
pipe.Expire(ctx, "counter", time.Hour)

_, err := pipe.Exec(ctx)
if err != nil {
    panic(err)
}

fmt.Println(incr.Val())
```

## 事务

```go
// TxPipeline (事务，保证原子性)
tx := rdb.TxPipeline()

tx.Incr(ctx, "counter")
tx.Expire(ctx, "counter", time.Hour)

_, err := tx.Exec(ctx)
if err != nil {
    panic(err)
}

// Watch (乐观锁)
err = rdb.Watch(ctx, func(tx *redis.Tx) error {
    val, err := tx.Get(ctx, "balance").Int()
    if err != nil && err != redis.Nil {
        return err
    }

    _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.Set(ctx, "balance", val+100, 0)
        return nil
    })
    return err
}, "balance")

if err == redis.TxFailedErr {
    // 重试
}
```

## Pub/Sub

```go
// 订阅
pubsub := rdb.Subscribe(ctx, "channel1", "channel2")
defer pubsub.Close()

// 接收消息
ch := pubsub.Channel()
for msg := range ch {
    fmt.Printf("Channel: %s, Message: %s\n", msg.Channel, msg.Payload)
}

// 发布
err := rdb.Publish(ctx, "channel1", "hello").Err()

// 模式订阅
pubsub = rdb.PSubscribe(ctx, "news:*")
```

## 分布式锁

```go
// 简单锁
func AcquireLock(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (bool, error) {
    return rdb.SetNX(ctx, key, "1", ttl).Result()
}

func ReleaseLock(ctx context.Context, rdb *redis.Client, key string) error {
    return rdb.Del(ctx, key).Err()
}

// 使用 Lua 脚本实现安全释放
var releaseLockScript = redis.NewScript(`
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
`)

func ReleaseLockSafe(ctx context.Context, rdb *redis.Client, key, value string) error {
    return releaseLockScript.Run(ctx, rdb, []string{key}, value).Err()
}
```

## 封装示例

```go
type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    return &RedisCache{client: rdb}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
    return c.client.Del(ctx, keys...).Err()
}

func (c *RedisCache) Close() error {
    return c.client.Close()
}
```

**下一节**：[缓存模式](./04-cache-patterns.md) - 学习常见缓存模式
