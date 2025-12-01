# 代码示例索引

本目录包含教程中的所有独立代码示例。每个示例都可以独立运行。

## 运行示例

```bash
# 进入示例目录
cd examples/syntax/hello

# 运行
go run main.go
```

## 示例目录

### 语法基础 (syntax/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [hello](./syntax/hello/) | Hello World 程序 | DEV-00 环境搭建 |
| [variables](./syntax/variables/) | 变量声明与使用 | DEV-01 语法基础 |
| [types](./syntax/types/) | 基本类型和类型转换 | DEV-01 语法基础 |
| [slices](./syntax/slices/) | 切片操作 | DEV-01 语法基础 |
| [maps](./syntax/maps/) | Map 操作 | DEV-01 语法基础 |
| [control](./syntax/control/) | 控制流语句 | DEV-01 语法基础 |
| [functions](./syntax/functions/) | 函数定义与调用 | DEV-01 语法基础 |
| [pointers](./syntax/pointers/) | 指针使用 | DEV-01 语法基础 |
| [errors](./syntax/errors/) | 错误处理 | DEV-01 语法基础 |
| [structs](./syntax/structs/) | 结构体定义 | DEV-02 结构体接口 |
| [methods](./syntax/methods/) | 方法定义 | DEV-02 结构体接口 |
| [interfaces](./syntax/interfaces/) | 接口实现 | DEV-02 结构体接口 |
| [embedding](./syntax/embedding/) | 结构体嵌入 | DEV-02 结构体接口 |

### 并发编程 (concurrency/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [goroutines](./concurrency/goroutines/) | Goroutine 基础 | DEV-03 并发编程 |
| [channels](./concurrency/channels/) | Channel 通信 | DEV-03 并发编程 |
| [select](./concurrency/select/) | Select 多路复用 | DEV-03 并发编程 |
| [sync](./concurrency/sync/) | 同步原语 (Mutex/WaitGroup) | DEV-03 并发编程 |
| [context](./concurrency/context/) | Context 使用 | DEV-03 并发编程 |
| [patterns](./concurrency/patterns/) | 并发模式 (Worker Pool/Fan-out) | DEV-03 并发编程 |

### 架构模式 (patterns/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [layered](./patterns/layered/) | 分层架构示例 | DEV-04 后端架构 |
| [config](./patterns/config/) | Viper 配置管理 | DEV-04 后端架构 |

### Web 开发 (web/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [basic-api](./web/basic-api/) | 基础 HTTP API | DEV-05 Web API |

### 数据库 (database/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [connect](./database/connect/) | 数据库连接 | DEV-07 数据库基础 |
| [query](./database/query/) | SQL 查询操作 | DEV-07 数据库基础 |
| [transaction](./database/transaction/) | 事务处理 | DEV-07 数据库基础 |

### GORM (gorm/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [basic](./gorm/basic/) | GORM 基础操作 | DEV-08 GORM框架 |
| [query](./gorm/query/) | GORM 查询 | DEV-08 GORM框架 |
| [associations](./gorm/associations/) | 关联关系 | DEV-08 GORM框架 |
| [transactions](./gorm/transactions/) | GORM 事务 | DEV-08 GORM框架 |

### 缓存 (cache/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [redis](./cache/redis/) | Redis 基础操作 | DEV-09 缓存策略 |
| [local](./cache/local/) | 本地缓存 | DEV-09 缓存策略 |
| [patterns](./cache/patterns/) | 缓存模式 | DEV-09 缓存策略 |

### 消息队列 (mq/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [basic](./mq/basic/) | 基础发布/消费 | DEV-10 消息队列 |
| [worker](./mq/worker/) | 工作队列模式 | DEV-10 消息队列 |
| [delayed](./mq/delayed/) | 延迟队列 | DEV-10 消息队列 |

### 可观测性 (observability/)

| 示例 | 说明 | 相关章节 |
|------|------|----------|
| [logging](./observability/logging/) | Zap 结构化日志 | DEV-11 可观测性 |
| [metrics](./observability/metrics/) | Prometheus 指标 | DEV-11 可观测性 |
| [tracing](./observability/tracing/) | Jaeger 分布式追踪 | DEV-11 可观测性 |

## 依赖服务

某些示例需要外部服务，可以使用项目根目录的 Docker Compose 启动：

```bash
# 启动所有服务
cd /path/to/go-learn
docker-compose -f infra/docker-compose.yml up -d

# 启动特定服务
docker-compose -f infra/docker-compose.yml up -d postgres redis rabbitmq
```

| 服务 | 端口 | 使用示例 |
|------|------|----------|
| PostgreSQL | 5432 | database/, gorm/ |
| Redis | 6379 | cache/ |
| RabbitMQ | 5672, 15672 | mq/ |
| Prometheus | 9090 | observability/metrics |
| Grafana | 3000 | observability/ |
| Jaeger | 16686 | observability/tracing |

## 示例结构

每个示例遵循统一结构：

```
examples/<category>/<name>/
├── main.go          # 主程序
├── go.mod           # 模块文件（如需要独立依赖）
└── README.md        # 示例说明（可选）
```

## 按难度分类

### 🟢 入门级

- syntax/hello
- syntax/variables
- syntax/types
- syntax/control
- syntax/functions

### 🟡 中级

- syntax/structs
- syntax/interfaces
- concurrency/goroutines
- concurrency/channels
- web/basic-api
- database/query

### 🔴 高级

- concurrency/patterns
- patterns/layered
- cache/patterns
- observability/tracing

## 贡献示例

添加新示例时请遵循：

1. 放在合适的分类目录下
2. 包含完整可运行的代码
3. 添加清晰的注释
4. 在此文件中更新索引

---

返回 [项目首页](../README.md) | [开发教程](../docs/01-development/) | [实战教程](../docs/02-practice/)
