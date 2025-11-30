# Go后端学习教程（四轨并行结构）

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 一套专为前端工程师设计的Go语言后端开发教程，采用**四轨并行结构**，支持灵活的学习路径选择。

## 四轨结构

本教程采用独特的**四轨并行学习结构**，允许你根据自身需求选择学习路径：

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Go后端学习教程                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  │  开发教程      │  │  实战教程      │  │  部署教程      │  │  测试教程      │
│  │  (Development) │  │  (Practice)   │  │  (Deployment) │  │  (Testing)    │
│  ├───────────────┤  ├───────────────┤  ├───────────────┤  ├───────────────┤
│  │ 语言基础       │  │ Todo CLI     │  │ 编译打包       │  │ 单元测试       │
│  │ 并发模型       │  │ Todo API     │  │ Docker化      │  │ 集成测试       │
│  │ 架构模式       │  │ 认证服务      │  │ CI/CD        │  │ E2E测试       │
│  │ 框架使用       │  │ 全栈应用      │  │ 监控运维      │  │ 性能测试       │
│  └───────────────┘  └───────────────┘  └───────────────┘  └───────────────┘
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## 轨道说明

### 轨道一：开发教程 (Development Track)

Go语言和后端开发的**核心知识体系**，12个章节，约52小时学习内容：

| 章节 | 内容 | 难度 | 时长 |
|------|------|------|------|
| [00-环境搭建](./docs/01-development/00-environment/) | Go安装、VS Code配置 | 🟢 | 2h |
| [01-语法基础](./docs/01-development/01-syntax/) | 变量、类型、函数、控制流 | 🟢 | 6h |
| [02-结构体接口](./docs/01-development/02-struct-interface/) | 结构体、方法、接口、泛型 | 🟡 | 4h |
| [03-并发编程](./docs/01-development/03-concurrency/) | goroutine、channel、context | 🟡 | 5h |
| [04-后端架构](./docs/01-development/04-architecture/) | 分层架构、DI、配置管理 | 🟡 | 3h |
| [05-Web API](./docs/01-development/05-web-api/) | net/http、Gin、中间件 | 🟡 | 6h |
| [06-认证授权](./docs/01-development/06-authentication/) | JWT、Session、OAuth2、RBAC | 🟡 | 5h |
| [07-数据库基础](./docs/01-development/07-database/) | database/sql、pgx、事务 | 🟡 | 4h |
| [08-GORM框架](./docs/01-development/08-gorm/) | ORM操作、关联关系、迁移 | 🟡 | 4h |
| [09-缓存策略](./docs/01-development/09-cache/) | Redis、本地缓存、一致性 | 🟡 | 4h |
| [10-消息队列](./docs/01-development/10-message-queue/) | RabbitMQ、任务队列、死信 | 🔴 | 4h |
| [11-可观测性](./docs/01-development/11-observability/) | 日志、指标、追踪 | 🔴 | 5h |

### 轨道二：实战教程 (Practice Track)

通过**完整项目**巩固知识，4个项目，约30小时实战：

| 项目 | 说明 | 难度 | 涉及知识 |
|------|------|------|----------|
| [todo-cli](./docs/02-practice/01-todo-cli/) | 命令行待办应用 | 🟢 | 语法基础、文件I/O |
| [todo-api](./docs/02-practice/02-todo-api/) | RESTful API服务 | 🟡 | Gin、GORM、JWT、Redis |
| [auth-service](./docs/02-practice/03-auth-service/) | 认证服务 | 🟡 | JWT、Session、OAuth2、RBAC |
| [fullstack-demo](./docs/02-practice/04-fullstack-demo/) | 全栈应用 | 🔴 | 全部技术栈 |

### 轨道三：部署教程 (Deployment Track)

将应用**部署到生产环境**的完整流程，5个章节：

| 章节 | 内容 | 难度 |
|------|------|------|
| [01-编译打包](./docs/03-deployment/01-compilation/) | go build、交叉编译、ldflags | 🟢 |
| [02-Docker容器化](./docs/03-deployment/02-docker/) | Dockerfile、多阶段构建、Compose | 🟡 |
| [03-配置管理](./docs/03-deployment/03-config/) | 环境变量、配置文件、Secrets | 🟡 |
| [04-CI/CD](./docs/03-deployment/04-cicd/) | GitHub Actions、GitLab CI | 🟡 |
| [05-运维监控](./docs/03-deployment/05-operations/) | 健康检查、日志管理、告警 | 🔴 |

### 轨道四：测试教程 (Testing Track)

编写**高质量测试**和**调试技巧**，5个章节：

| 章节 | 内容 | 难度 |
|------|------|------|
| [01-单元测试](./docs/04-testing/01-unit-testing/) | go test、testify、mocking | 🟢 |
| [02-集成测试](./docs/04-testing/02-integration/) | 数据库测试、API测试、testcontainers | 🟡 |
| [03-E2E测试](./docs/04-testing/03-e2e/) | 端到端测试场景 | 🟡 |
| [04-性能测试](./docs/04-testing/04-performance/) | benchmark、pprof | 🟡 |
| [05-调试技巧](./docs/04-testing/05-debugging/) | delve、常见Bug排查 | 🔴 |

## 学习路径推荐

### 路径A：系统学习（推荐新手）

```
开发教程(全部) → 实战教程(01-02) → 测试教程(01-02) → 部署教程(01-02) → 实战教程(03-04) → 完成
```

### 路径B：快速上手（有后端基础）

```
开发教程(00-01) → 实战教程(01) → 开发教程(05-08) → 实战教程(02) → 部署教程 → 测试教程
```

### 路径C：项目驱动（边做边学）

```
实战教程(01) ←→ 开发教程(按需) → 实战教程(02) ←→ 开发教程(按需) → 部署+测试
```

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

# 启动基础设施服务
docker-compose -f infra/docker-compose.yml up -d

# 验证服务
docker-compose -f infra/docker-compose.yml ps
```

服务访问地址：
- PostgreSQL: `localhost:5432` (user: golearn, pass: golearn123)
- Redis: `localhost:6379`
- RabbitMQ: `localhost:15672` (user: golearn, pass: golearn123)
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (user: admin, pass: admin123)
- Jaeger: `http://localhost:16686`

## 技术栈

| 类别 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.21+ |
| Web框架 | Gin | v1.9+ |
| ORM | GORM | v2 |
| 数据库 | PostgreSQL | 15+ |
| 缓存 | Redis | 7+ |
| 消息队列 | RabbitMQ | 3.x |
| 认证 | golang-jwt/jwt | v5 |
| 配置 | Viper | v1.17+ |
| 日志 | Zap | v1.26+ |
| 监控 | Prometheus + Grafana | latest |
| 追踪 | Jaeger | 1.50+ |

## 项目结构

```
go-learn/
├── docs/                         # 四轨教程文档
│   ├── 00-introduction/          # 总体介绍
│   ├── 01-development/           # 轨道一：开发教程
│   ├── 02-practice/              # 轨道二：实战教程
│   ├── 03-deployment/            # 轨道三：部署教程
│   └── 04-testing/               # 轨道四：测试教程
├── projects/                     # 实战项目源码
│   ├── todo-cli/
│   ├── todo-api/
│   ├── auth-service/
│   └── fullstack-demo/
├── tests/                        # 测试示例与验证
├── examples/                     # 独立代码示例
├── infra/                        # 基础设施配置
│   ├── docker-compose.yml
│   ├── postgres/
│   ├── redis/
│   └── monitoring/
├── go.mod                        # 根模块
└── go.work                       # 工作区配置
```

## 特色

- **前端视角** - 每个概念都与JavaScript/TypeScript进行对比
- **四轨并行** - 灵活选择学习路径，适应不同背景
- **完整示例** - 所有代码示例都可独立运行，附带预期输出
- **渐进式学习** - 从简单到复杂，循序渐进
- **实践驱动** - 每章包含练习题和实战项目
- **生产级实践** - 涵盖真实项目中需要的所有技能

## 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解如何参与。

## 许可证

MIT License - 详见 [LICENSE](./LICENSE) 文件
