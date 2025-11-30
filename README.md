# 前端工程师Go后端开发学习教程

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 一套专为前端工程师设计的Go语言后端开发教程，从零基础到生产级服务的完整学习路径。

## 为什么选择这个教程？

作为前端工程师，你已经熟悉了JavaScript/TypeScript的异步编程模式。Go语言以其出色的并发性能、简洁的语法和强大的标准库，成为后端开发的理想选择。本教程将帮助你：

- 🚀 **快速入门** - 利用你现有的编程知识，建立Go与JS/TS的概念映射
- 💡 **理解并发** - 掌握goroutine和channel，理解Go的并发优势
- 🛠️ **实战项目** - 通过完整的项目示例，学习后端开发的最佳实践
- 📊 **生产级技能** - 涵盖数据库、缓存、消息队列、监控等企业级技术栈

## 学习路径

### 🎯 阶段1: 基础入门 (P1)

| 章节 | 内容 | 预计时间 |
|------|------|----------|
| [01-环境搭建](./docs/01-environment-setup/) | Go安装、VS Code配置、Hello World | 1小时 |
| [02-语法基础](./docs/02-syntax-basics/) | 变量、类型、函数、控制流 | 3小时 |

### 🔧 阶段2: 核心技能 (P2)

| 章节 | 内容 | 预计时间 |
|------|------|----------|
| [03-结构体与接口](./docs/03-struct-interface/) | 结构体、方法、接口、组合 | 2小时 |
| [04-并发编程](./docs/04-concurrency/) | goroutine、channel、sync包 | 3小时 |
| [05-后端架构](./docs/05-architecture/) | 分层架构、依赖注入、配置管理 | 2小时 |
| [06-Web API开发](./docs/06-web-api/) | net/http、Gin框架、中间件 | 4小时 |
| [07-认证与授权](./docs/07-authentication/) | JWT、Session、OAuth2、RBAC | 3小时 |
| [08-PostgreSQL基础](./docs/08-database-sql/) | database/sql、pgx、事务处理 | 3小时 |
| [09-GORM框架](./docs/09-gorm/) | ORM操作、关联关系、迁移 | 3小时 |
| [10-Redis缓存](./docs/10-caching/) | go-redis、缓存策略、一致性 | 2小时 |

### 🚀 阶段3: 进阶实战 (P3)

| 章节 | 内容 | 预计时间 |
|------|------|----------|
| [11-消息队列](./docs/11-message-queue/) | RabbitMQ、任务队列、延迟队列 | 3小时 |
| [12-可观测性](./docs/12-observability/) | 日志、指标、追踪、Grafana | 3小时 |
| [13-项目部署](./docs/13-deployment/) | 编译、Docker、Docker Compose | 2小时 |
| [14-调试技能](./docs/14-debugging/) | delve、pprof、常见问题排查 | 2小时 |

## 综合实战项目

- [todo-api](./projects/todo-api/) - 基础CRUD API项目
- [user-auth-service](./projects/user-auth-service/) - 完整认证服务
- [full-stack-demo](./projects/full-stack-demo/) - 生产级后端服务演示

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- VS Code + Go插件
- Docker Desktop（用于运行数据库等服务）
- 基本的JavaScript/TypeScript知识

### 开发环境启动

```bash
# 克隆仓库
git clone https://github.com/your-username/go-learn.git
cd go-learn

# 启动基础设施服务（PostgreSQL, Redis, RabbitMQ等）
docker-compose -f infra/docker-compose.yml up -d

# 运行第一个示例
cd docs/01-environment-setup/examples/hello
go run main.go
```

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| Web框架 | net/http (标准库), Gin |
| ORM | GORM v2 |
| 数据库 | PostgreSQL 15+ |
| 缓存 | Redis 7+ |
| 消息队列 | RabbitMQ |
| 认证 | JWT (golang-jwt/jwt v5) |
| 配置 | Viper |
| 日志 | zap, slog |
| 监控 | Prometheus + Grafana |
| 追踪 | Jaeger |

## 特色

- ✅ **前端视角** - 每个概念都与JavaScript/TypeScript进行对比
- ✅ **完整示例** - 所有代码示例都可独立运行，附带预期输出
- ✅ **渐进式学习** - 从简单到复杂，循序渐进
- ✅ **实践驱动** - 每章包含练习题和实战项目
- ✅ **生产级实践** - 涵盖真实项目中需要的所有技能

## 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解如何参与。

## 许可证

MIT License - 详见 [LICENSE](./LICENSE) 文件
