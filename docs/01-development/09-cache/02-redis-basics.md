# Redis 基础 (Redis Basics)

## 概述

Redis (Remote Dictionary Server) 是一个开源的内存数据结构存储，可用作数据库、缓存和消息代理。

## 安装

### macOS

```bash
# Homebrew 安装
brew install redis

# 启动服务
brew services start redis

# 停止服务
brew services stop redis

# 验证
redis-cli ping
# PONG
```

### Linux (Ubuntu/Debian)

```bash
# 安装
sudo apt update
sudo apt install redis-server

# 启动
sudo systemctl start redis
sudo systemctl enable redis

# 验证
redis-cli ping
```

### Docker

```bash
# 启动 Redis
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7-alpine

# 带持久化
docker run -d \
  --name redis \
  -p 6379:6379 \
  -v redis-data:/data \
  redis:7-alpine redis-server --appendonly yes

# 带密码
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7-alpine redis-server --requirepass yourpassword

# 连接
docker exec -it redis redis-cli
```

### Docker Compose

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD:-}
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  redis-data:
```

## Redis CLI 基础

### 连接

```bash
# 本地连接
redis-cli

# 远程连接
redis-cli -h hostname -p 6379

# 带密码
redis-cli -a password
redis-cli -h hostname -p 6379 -a password

# 选择数据库 (0-15)
redis-cli -n 1
```

### 基本命令

```bash
# 测试连接
127.0.0.1:6379> PING
PONG

# 查看所有 key
127.0.0.1:6379> KEYS *

# 查看 key 数量
127.0.0.1:6379> DBSIZE

# 清空当前数据库
127.0.0.1:6379> FLUSHDB

# 清空所有数据库
127.0.0.1:6379> FLUSHALL

# 查看服务器信息
127.0.0.1:6379> INFO

# 查看内存使用
127.0.0.1:6379> INFO memory
```

## 数据类型

### String (字符串)

```bash
# 设置
SET name "Alice"
SET counter 100

# 获取
GET name
# "Alice"

# 设置过期时间 (秒)
SET session "abc123" EX 3600
SETEX session 3600 "abc123"

# 设置过期时间 (毫秒)
SET session "abc123" PX 3600000

# 仅当 key 不存在时设置 (分布式锁)
SETNX lock "1"
SET lock "1" NX EX 10

# 仅当 key 存在时设置
SET name "Bob" XX

# 数值操作
INCR counter      # 101
INCRBY counter 5  # 106
DECR counter      # 105
DECRBY counter 5  # 100

# 浮点数
INCRBYFLOAT price 1.5

# 批量操作
MSET name "Alice" age "25" city "NYC"
MGET name age city
# 1) "Alice"
# 2) "25"
# 3) "NYC"

# 追加
APPEND name " Smith"
GET name
# "Alice Smith"

# 获取长度
STRLEN name
# 11
```

### Hash (哈希)

```bash
# 设置单个字段
HSET user:1 name "Alice"
HSET user:1 age 25
HSET user:1 email "alice@example.com"

# 设置多个字段
HMSET user:2 name "Bob" age 30 email "bob@example.com"
HSET user:3 name "Charlie" age 35 email "charlie@example.com"

# 获取单个字段
HGET user:1 name
# "Alice"

# 获取多个字段
HMGET user:1 name age
# 1) "Alice"
# 2) "25"

# 获取所有字段
HGETALL user:1
# 1) "name"
# 2) "Alice"
# 3) "age"
# 4) "25"
# 5) "email"
# 6) "alice@example.com"

# 获取所有字段名
HKEYS user:1

# 获取所有值
HVALS user:1

# 字段是否存在
HEXISTS user:1 name
# (integer) 1

# 删除字段
HDEL user:1 email

# 字段数量
HLEN user:1

# 数值操作
HINCRBY user:1 age 1
HINCRBYFLOAT user:1 balance 10.5
```

### List (列表)

```bash
# 左侧插入
LPUSH tasks "task1" "task2" "task3"
# 列表: task3, task2, task1

# 右侧插入
RPUSH tasks "task4"
# 列表: task3, task2, task1, task4

# 获取范围 (0 到 -1 表示全部)
LRANGE tasks 0 -1
# 1) "task3"
# 2) "task2"
# 3) "task1"
# 4) "task4"

# 获取指定索引
LINDEX tasks 0
# "task3"

# 列表长度
LLEN tasks

# 左侧弹出
LPOP tasks
# "task3"

# 右侧弹出
RPOP tasks
# "task4"

# 阻塞弹出 (队列)
BLPOP tasks 10  # 等待 10 秒
BRPOP tasks 10

# 修剪列表 (保留指定范围)
LTRIM tasks 0 99  # 保留前 100 个

# 设置指定索引的值
LSET tasks 0 "new_task"
```

### Set (集合)

```bash
# 添加
SADD tags "go" "redis" "cache"

# 获取所有成员
SMEMBERS tags
# 1) "go"
# 2) "redis"
# 3) "cache"

# 是否是成员
SISMEMBER tags "go"
# (integer) 1

# 成员数量
SCARD tags

# 删除成员
SREM tags "cache"

# 随机获取
SRANDMEMBER tags 2

# 随机弹出
SPOP tags

# 集合运算
SADD set1 "a" "b" "c"
SADD set2 "b" "c" "d"

# 交集
SINTER set1 set2
# 1) "b"
# 2) "c"

# 并集
SUNION set1 set2
# 1) "a"
# 2) "b"
# 3) "c"
# 4) "d"

# 差集
SDIFF set1 set2
# 1) "a"
```

### Sorted Set (有序集合)

```bash
# 添加 (score member)
ZADD leaderboard 100 "Alice"
ZADD leaderboard 200 "Bob"
ZADD leaderboard 150 "Charlie"

# 获取排名 (升序)
ZRANGE leaderboard 0 -1
# 1) "Alice"
# 2) "Charlie"
# 3) "Bob"

# 获取排名 (降序)
ZREVRANGE leaderboard 0 -1
# 1) "Bob"
# 2) "Charlie"
# 3) "Alice"

# 带分数
ZREVRANGE leaderboard 0 -1 WITHSCORES
# 1) "Bob"
# 2) "200"
# 3) "Charlie"
# 4) "150"
# 5) "Alice"
# 6) "100"

# 获取成员排名 (从 0 开始)
ZRANK leaderboard "Alice"    # 升序排名
ZREVRANK leaderboard "Alice" # 降序排名

# 获取分数
ZSCORE leaderboard "Alice"

# 增加分数
ZINCRBY leaderboard 50 "Alice"

# 按分数范围获取
ZRANGEBYSCORE leaderboard 100 200
ZRANGEBYSCORE leaderboard -inf +inf

# 成员数量
ZCARD leaderboard

# 删除成员
ZREM leaderboard "Alice"
```

## Key 操作

```bash
# 查找 key
KEYS *           # 所有 key (生产环境慎用)
KEYS user:*      # 匹配模式
SCAN 0 MATCH user:* COUNT 100  # 安全扫描

# 检查 key 是否存在
EXISTS user:1
# (integer) 1

# 删除
DEL user:1
DEL user:1 user:2 user:3  # 批量删除

# 设置过期时间
EXPIRE user:1 3600        # 秒
PEXPIRE user:1 3600000    # 毫秒
EXPIREAT user:1 1735689600  # Unix 时间戳

# 查看剩余时间
TTL user:1    # 秒
PTTL user:1   # 毫秒
# -1: 永不过期
# -2: key 不存在

# 移除过期时间
PERSIST user:1

# 重命名
RENAME oldkey newkey
RENAMENX oldkey newkey  # 仅当 newkey 不存在

# 查看类型
TYPE user:1
# string/hash/list/set/zset
```

## 事务

```bash
# 开始事务
MULTI

# 命令入队
SET balance 100
DECRBY balance 50
INCRBY savings 50

# 执行事务
EXEC

# 取消事务
DISCARD

# 乐观锁 (WATCH)
WATCH balance
val = GET balance
MULTI
SET balance (val - 50)
EXEC
# 如果 balance 在 WATCH 和 EXEC 之间被修改，EXEC 返回 nil
```

## 发布订阅

```bash
# 订阅频道 (终端 1)
SUBSCRIBE news

# 发布消息 (终端 2)
PUBLISH news "Hello, World!"

# 模式订阅
PSUBSCRIBE news:*
```

## 持久化

### RDB (快照)

```bash
# 手动触发
SAVE      # 同步，阻塞
BGSAVE    # 异步，后台

# 配置 (redis.conf)
save 900 1      # 900秒内至少1次写入
save 300 10     # 300秒内至少10次写入
save 60 10000   # 60秒内至少10000次写入
```

### AOF (追加日志)

```bash
# 配置 (redis.conf)
appendonly yes
appendfsync everysec  # 每秒同步一次

# 重写 AOF
BGREWRITEAOF
```

## 监控

```bash
# 实时监控命令
MONITOR

# 慢查询日志
SLOWLOG GET 10
SLOWLOG LEN

# 客户端列表
CLIENT LIST

# 内存分析
MEMORY DOCTOR
MEMORY STATS

# 性能指标
INFO stats
INFO clients
INFO memory
```

## 配置参考

```bash
# redis.conf 关键配置

# 绑定地址
bind 127.0.0.1

# 端口
port 6379

# 密码
requirepass yourpassword

# 最大内存
maxmemory 256mb

# 内存淘汰策略
maxmemory-policy allkeys-lru

# 持久化
save 900 1
appendonly yes

# 客户端连接
maxclients 10000
timeout 300
```

**下一节**：[go-redis](./03-go-redis.md) - 学习 Go Redis 客户端使用
