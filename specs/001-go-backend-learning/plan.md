# Implementation Plan: Go后端学习教程（四轨并行结构）

**Branch**: `001-go-backend-learning` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-go-backend-learning/spec.md`

## Summary

本教程为前端工程师设计，以四条独立学习轨道组织内容：**开发教程**、**实战教程**、**部署教程**、**测试教程**。每条轨道可独立学习，也可交叉学习，适应不同学习者的需求和学习风格。

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Gin (Web框架), GORM v2 (ORM), Viper (配置), Zap (日志), jwt-go (认证)
**Storage**: PostgreSQL 15+, Redis 7+, RabbitMQ 3.x
**Testing**: Go testing, testify, mockery, go-sqlmock
**Target Platform**: Linux server (production), macOS/Windows/Linux (development)
**Project Type**: Multi-track tutorial (4 独立轨道)
**Performance Goals**: API响应 <100ms (p95), 支持1000 req/s
**Constraints**: 教程代码需在入门级硬件上运行
**Scale/Scope**: 14个User Stories, 4条学习轨道, 预计学习时间40-60小时

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. 渐进式学习 | ✅ PASS | 四轨结构保持每轨内部的渐进性，明确列出先决条件 |
| II. 实践驱动 | ✅ PASS | 实战教程独立成轨，每章包含可运行的完整项目 |
| III. 可验证性 | ✅ PASS | 测试教程独立成轨，提供验证标准和自测检查点 |
| IV. 问题解决导向 | ✅ PASS | 从真实场景出发，说明"为什么"和"何时使用" |

## 四轨结构设计

### 轨道概览

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

### 轨道依赖关系

```
开发教程 ──────────────────────────────────────────────────────────────────►
    │
    │ 基础完成后
    ▼
实战教程 ─────────────────────────────────────────────────────────────────►
    │                        │
    │ 项目完成后              │ 项目可测试时
    ▼                        ▼
部署教程 ──────────►      测试教程 ─────────────────────────────────────────►
```

---

## 轨道一：开发教程 (Development Track)

### 目录结构

```
docs/
├── 01-development/           # 开发教程
│   ├── 00-environment/       # 环境搭建 (US1)
│   │   ├── README.md         # 章节概览
│   │   ├── install-guide.md  # Go安装指南(Win/Mac/Linux)
│   │   ├── gopath-modules.md # GOPATH vs Go Modules
│   │   ├── vscode-setup.md   # VS Code配置
│   │   └── first-program.md  # 第一个Go程序
│   │
│   ├── 01-syntax/            # 语法基础 (US2)
│   │   ├── README.md
│   │   ├── 01-variables.md       # 变量声明(var, :=, const)
│   │   ├── 02-basic-types.md     # 基本类型(int, string, bool, float)
│   │   ├── 03-composite-types.md # 复合类型(array, slice, map)
│   │   ├── 04-control-flow.md    # 控制流(if, for, switch)
│   │   ├── 05-functions.md       # 函数(参数、返回值、多返回值)
│   │   ├── 06-pointers.md        # 指针基础
│   │   ├── 07-error-handling.md  # 错误处理(error, panic, recover)
│   │   ├── 08-packages.md        # 包管理(import, go mod)
│   │   └── js-comparison.md      # JS/TS对比总结
│   │
│   ├── 02-struct-interface/  # 结构体与接口 (US3)
│   │   ├── README.md
│   │   ├── 01-structs.md         # 结构体定义与初始化
│   │   ├── 02-struct-tags.md     # 结构体标签(json, db)
│   │   ├── 03-methods.md         # 方法(值接收者vs指针接收者)
│   │   ├── 04-interfaces.md      # 接口定义与隐式实现
│   │   ├── 05-type-assertion.md  # 类型断言与类型转换
│   │   ├── 06-embedding.md       # 结构体嵌入(组合)
│   │   └── 07-generics.md        # 泛型基础(Go 1.18+)
│   │
│   ├── 03-concurrency/       # 并发编程 (US4)
│   │   ├── README.md
│   │   ├── 01-goroutines.md      # Goroutine基础
│   │   ├── 02-channels.md        # Channel(创建、发送、接收)
│   │   ├── 03-buffered-chan.md   # 缓冲Channel
│   │   ├── 04-select.md          # Select多路复用
│   │   ├── 05-sync-package.md    # sync包(WaitGroup, Mutex, RWMutex)
│   │   ├── 06-context.md         # Context上下文控制
│   │   ├── 07-patterns.md        # 并发模式(worker pool, fan-out/fan-in)
│   │   └── 08-race-detection.md  # 竞态检测与避免
│   │
│   ├── 04-architecture/      # 后端架构 (US5)
│   │   ├── README.md
│   │   ├── 01-project-layout.md  # 项目目录结构
│   │   ├── 02-layered-arch.md    # 分层架构(Handler/Service/Repository)
│   │   ├── 03-dependency-injection.md # 依赖注入
│   │   ├── 04-config-management.md    # 配置管理(Viper)
│   │   ├── 05-error-design.md    # 错误设计与处理策略
│   │   └── 06-microservices-intro.md  # 微服务概念与API网关
│   │
│   ├── 05-web-api/           # Web API开发 (US6)
│   │   ├── README.md
│   │   ├── 01-http-basics.md     # HTTP协议与net/http包
│   │   ├── 02-gin-intro.md       # Gin框架入门
│   │   ├── 03-routing.md         # 路由(路径参数、查询参数)
│   │   ├── 04-request-binding.md # 请求绑定(JSON, Form, URI)
│   │   ├── 05-response.md        # 响应处理(JSON, XML, 文件)
│   │   ├── 06-validation.md      # 请求验证(binding tags)
│   │   ├── 07-middleware.md      # 中间件(日志、恢复、CORS)
│   │   ├── 08-error-handling.md  # API错误处理
│   │   ├── 09-file-upload.md     # 文件上传
│   │   └── 10-swagger.md         # API文档(Swagger/OpenAPI)
│   │
│   ├── 06-authentication/    # 认证授权 (US7)
│   │   ├── README.md
│   │   ├── 01-auth-intro.md      # 认证vs授权概念
│   │   ├── 02-password.md        # 密码哈希(bcrypt)
│   │   ├── 03-jwt-basics.md      # JWT原理与结构
│   │   ├── 04-jwt-impl.md        # JWT实现(生成、验证、刷新)
│   │   ├── 05-jwt-middleware.md  # JWT认证中间件
│   │   ├── 06-session.md         # Session认证
│   │   ├── 07-jwt-vs-session.md  # JWT vs Session对比
│   │   ├── 08-oauth2-intro.md    # OAuth2原理
│   │   ├── 09-oauth2-github.md   # GitHub OAuth2登录
│   │   └── 10-rbac.md            # RBAC权限控制
│   │
│   ├── 07-database/          # 数据库基础 (US8)
│   │   ├── README.md
│   │   ├── 01-sql-review.md      # SQL回顾(CRUD, JOIN)
│   │   ├── 02-postgres-setup.md  # PostgreSQL安装与配置
│   │   ├── 03-database-sql.md    # database/sql标准库
│   │   ├── 04-pgx-driver.md      # pgx驱动使用
│   │   ├── 05-connection-pool.md # 连接池配置
│   │   ├── 06-prepared-stmt.md   # 预处理语句(防SQL注入)
│   │   ├── 07-transactions.md    # 事务处理
│   │   └── 08-migrations.md      # 数据库迁移
│   │
│   ├── 08-gorm/              # GORM框架 (US9)
│   │   ├── README.md
│   │   ├── 01-gorm-intro.md      # GORM简介与安装
│   │   ├── 02-model-definition.md # 模型定义与约定
│   │   ├── 03-crud.md            # CRUD操作
│   │   ├── 04-query-builder.md   # 查询构建器
│   │   ├── 05-associations.md    # 关联关系(1:1, 1:N, M:N)
│   │   ├── 06-preload.md         # 预加载与懒加载
│   │   ├── 07-hooks.md           # 钩子函数
│   │   ├── 08-transactions.md    # GORM事务
│   │   ├── 09-raw-sql.md         # 原生SQL与GORM混用
│   │   └── 10-best-practices.md  # 最佳实践与性能优化
│   │
│   ├── 09-cache/             # 缓存策略 (US10)
│   │   ├── README.md
│   │   ├── 01-cache-intro.md     # 缓存概念与作用
│   │   ├── 02-redis-basics.md    # Redis安装与基本命令
│   │   ├── 03-go-redis.md        # go-redis客户端使用
│   │   ├── 04-cache-patterns.md  # 缓存模式(Cache-Aside, Write-Through)
│   │   ├── 05-cache-problems.md  # 缓存问题(穿透、击穿、雪崩)
│   │   ├── 06-local-cache.md     # 本地缓存(go-cache)
│   │   ├── 07-multi-level.md     # 多级缓存架构
│   │   └── 08-cache-consistency.md # 缓存一致性策略
│   │
│   ├── 10-message-queue/     # 消息队列 (US11)
│   │   ├── README.md
│   │   ├── 01-mq-intro.md        # 消息队列概念
│   │   ├── 02-rabbitmq-setup.md  # RabbitMQ安装与管理界面
│   │   ├── 03-rabbitmq-go.md     # Go操作RabbitMQ
│   │   ├── 04-work-queues.md     # 工作队列模式
│   │   ├── 05-pubsub.md          # 发布订阅模式
│   │   ├── 06-delayed-queue.md   # 延迟队列
│   │   ├── 07-dead-letter.md     # 死信队列
│   │   └── 08-reliability.md     # 消息可靠性(确认、持久化)
│   │
│   └── 11-observability/     # 可观测性 (US12)
│       ├── README.md
│       ├── 01-observability-intro.md # 可观测性三支柱
│       ├── 02-structured-logging.md  # 结构化日志(Zap)
│       ├── 03-log-levels.md      # 日志级别与上下文
│       ├── 04-prometheus.md      # Prometheus指标收集
│       ├── 05-custom-metrics.md  # 自定义指标
│       ├── 06-grafana.md         # Grafana看板配置
│       ├── 07-alerting.md        # 告警规则
│       ├── 08-tracing-intro.md   # 分布式追踪概念
│       ├── 09-jaeger.md          # Jaeger集成
│       └── 10-elk.md             # ELK日志聚合(可选)
```

### 章节内容详情

#### 00-environment 环境搭建
- **学习目标**: 搭建Go开发环境，运行第一个程序
- **核心内容**: Go安装、环境变量、Go Modules、IDE配置
- **动手实践**: Hello World程序、简单计算器

#### 01-syntax 语法基础
- **学习目标**: 掌握Go基础语法，理解与JS/TS的差异
- **核心内容**:
  - 变量声明（var vs := vs const）
  - 基本类型（int系列、string、bool、浮点数）
  - 复合类型（数组定长、切片动态、map哈希表）
  - 控制流（if无括号、for统一、switch自动break）
  - 函数（多返回值、命名返回值、可变参数）
  - 指针（&取地址、*解引用）
  - 错误处理（error接口、panic/recover）
- **JS对比重点**: 静态类型vs动态类型、var作用域、nil vs null/undefined

#### 02-struct-interface 结构体与接口
- **学习目标**: 理解Go的面向对象方式
- **核心内容**:
  - 结构体定义与嵌入（组合优于继承）
  - 结构体标签（json、db、validate）
  - 方法与接收者（值vs指针）
  - 接口与隐式实现
  - 类型断言与类型开关
  - 泛型基础（Go 1.18+）
- **JS对比重点**: struct vs class、隐式接口vs显式implements

#### 03-concurrency 并发编程
- **学习目标**: 掌握Go并发模型，理解与async/await的差异
- **核心内容**:
  - Goroutine（轻量级线程）
  - Channel（通信机制）
  - Select（多路复用）
  - sync包（WaitGroup、Mutex、Once）
  - Context（取消、超时、传值）
  - 并发模式（worker pool、pipeline）
- **JS对比重点**: goroutine vs async/await、channel vs Promise

#### 05-web-api Web API开发
- **学习目标**: 开发RESTful API服务
- **核心内容**:
  - net/http标准库（底层理解）
  - Gin框架（路由、分组、参数绑定）
  - 请求处理（JSON/Form/URI绑定）
  - 响应格式（统一响应结构）
  - 请求验证（binding标签）
  - 中间件（日志、认证、CORS、限流）
  - 错误处理（业务错误、系统错误）
  - 文件上传下载
  - Swagger文档生成

### 学习路线

| 阶段 | 章节 | 预计时间 | 前置条件 | 对应US |
|------|------|----------|----------|--------|
| 入门 | 00-environment | 2h | 无 | US1 |
| 入门 | 01-syntax | 6h | 00完成 | US2 |
| 基础 | 02-struct-interface | 4h | 01完成 | US3 |
| 基础 | 03-concurrency | 5h | 02完成 | US4 |
| 进阶 | 04-architecture | 3h | 03完成 | US5 |
| 进阶 | 05-web-api | 6h | 04完成 | US6 |
| 进阶 | 06-authentication | 5h | 05完成 | US7 |
| 进阶 | 07-database | 4h | 05完成 | US8 |
| 进阶 | 08-gorm | 4h | 07完成 | US9 |
| 高级 | 09-cache | 4h | 08完成 | US10 |
| 高级 | 10-message-queue | 4h | 09完成 | US11 |
| 高级 | 11-observability | 5h | 10完成 | US12 |

**总计**: 约52小时

---

## 轨道二：实战教程 (Practice Track)

### 目录结构

```
docs/
├── 02-practice/              # 实战教程
│   ├── 00-overview/          # 项目概览
│   │   └── README.md
│   ├── 01-todo-cli/          # CLI应用实战
│   │   ├── README.md
│   │   ├── step-01-init.md
│   │   ├── step-02-crud.md
│   │   ├── step-03-storage.md
│   │   └── step-04-polish.md
│   ├── 02-todo-api/          # API服务实战
│   │   ├── README.md
│   │   ├── step-01-setup.md
│   │   ├── step-02-routes.md
│   │   ├── step-03-database.md
│   │   ├── step-04-auth.md
│   │   └── step-05-cache.md
│   ├── 03-auth-service/      # 认证服务实战
│   │   ├── README.md
│   │   ├── step-01-jwt.md
│   │   ├── step-02-session.md
│   │   ├── step-03-oauth2.md
│   │   └── step-04-rbac.md
│   └── 04-fullstack-demo/    # 全栈项目实战
│       ├── README.md
│       ├── step-01-planning.md
│       ├── step-02-backend.md
│       ├── step-03-frontend.md
│       ├── step-04-integration.md
│       └── step-05-deploy.md

projects/
├── todo-cli/                 # CLI项目源码
│   ├── main.go
│   ├── cmd/
│   ├── internal/
│   └── go.mod
├── todo-api/                 # API项目源码
│   ├── main.go
│   ├── cmd/
│   ├── internal/
│   │   ├── handlers/
│   │   ├── services/
│   │   ├── repositories/
│   │   └── models/
│   └── go.mod
├── auth-service/             # 认证服务源码
│   ├── main.go
│   ├── internal/
│   └── go.mod
└── fullstack-demo/           # 全栈项目源码
    ├── backend/
    ├── frontend/
    └── docker-compose.yml
```

### 项目列表

| 项目 | 难度 | 预计时间 | 开发教程前置 | 学习重点 |
|------|------|----------|-------------|----------|
| todo-cli | 🟢 入门 | 4h | 01-syntax | CLI开发、文件I/O |
| todo-api | 🟡 进阶 | 8h | 05-web-api, 08-gorm | RESTful API、数据库 |
| auth-service | 🟡 进阶 | 6h | 06-authentication | JWT、OAuth2、RBAC |
| fullstack-demo | 🔴 高级 | 12h | 全部开发教程 | 全栈整合、微服务 |

---

## 轨道三：部署教程 (Deployment Track)

### 目录结构

```
docs/
├── 03-deployment/            # 部署教程
│   ├── 00-overview/          # 部署概览
│   │   └── README.md
│   ├── 01-compilation/       # 编译与打包 (US13)
│   │   ├── README.md
│   │   ├── basic-build.md
│   │   ├── cross-compile.md
│   │   └── build-flags.md
│   ├── 02-docker/            # Docker容器化 (US13)
│   │   ├── README.md
│   │   ├── dockerfile.md
│   │   ├── multi-stage.md
│   │   └── compose.md
│   ├── 03-config/            # 配置管理
│   │   ├── README.md
│   │   ├── env-vars.md
│   │   ├── config-files.md
│   │   └── secrets.md
│   ├── 04-cicd/              # CI/CD流水线
│   │   ├── README.md
│   │   ├── github-actions.md
│   │   └── gitlab-ci.md
│   └── 05-operations/        # 运维监控
│       ├── README.md
│       ├── health-check.md
│       ├── log-management.md
│       └── alerting.md

infra/
├── docker-compose.yml        # 开发环境
├── docker-compose.prod.yml   # 生产环境
├── postgres/
│   └── init.sql
├── redis/
│   └── redis.conf
├── monitoring/
│   ├── prometheus.yml
│   └── grafana/
├── ci/
│   ├── Dockerfile
│   └── .github/
│       └── workflows/
│           ├── test.yml
│           └── deploy.yml
└── k8s/                      # (可选) Kubernetes配置
    ├── deployment.yaml
    ├── service.yaml
    └── configmap.yaml
```

### 学习路线

| 阶段 | 章节 | 预计时间 | 前置条件 | 学习重点 |
|------|------|----------|----------|----------|
| 入门 | 01-compilation | 2h | 开发教程01 | go build, 交叉编译 |
| 基础 | 02-docker | 3h | 实战项目完成 | Dockerfile, Compose |
| 基础 | 03-config | 2h | 02完成 | 环境变量, Viper |
| 进阶 | 04-cicd | 3h | 03完成 | GitHub Actions |
| 进阶 | 05-operations | 3h | 04完成 | 健康检查, 监控 |

---

## 轨道四：测试教程 (Testing Track)

### 目录结构

```
docs/
├── 04-testing/               # 测试教程
│   ├── 00-overview/          # 测试概览
│   │   └── README.md
│   ├── 01-unit-testing/      # 单元测试
│   │   ├── README.md
│   │   ├── basics.md
│   │   ├── table-driven.md
│   │   ├── mocking.md
│   │   └── coverage.md
│   ├── 02-integration/       # 集成测试
│   │   ├── README.md
│   │   ├── database.md
│   │   ├── api.md
│   │   └── testcontainers.md
│   ├── 03-e2e/               # 端到端测试
│   │   ├── README.md
│   │   ├── setup.md
│   │   └── scenarios.md
│   ├── 04-performance/       # 性能测试
│   │   ├── README.md
│   │   ├── benchmarks.md
│   │   └── profiling.md
│   └── 05-debugging/         # 调试技巧 (US14)
│       ├── README.md
│       ├── compile-errors.md
│       ├── runtime-errors.md
│       ├── delve.md
│       └── pprof.md

tests/
├── go.mod
├── go.sum
├── unit/                     # 单元测试示例
│   ├── calculator_test.go
│   ├── user_service_test.go
│   └── mock_examples_test.go
├── integration/              # 集成测试示例
│   ├── database_test.go
│   ├── api_test.go
│   └── testdata/
├── e2e/                      # E2E测试示例
│   ├── user_flow_test.go
│   └── fixtures/
├── benchmark/                # 性能测试示例
│   ├── json_bench_test.go
│   └── concurrent_bench_test.go
├── debugging/                # 调试练习
│   ├── bug-01-nil-pointer/
│   ├── bug-02-race-condition/
│   ├── bug-03-deadlock/
│   └── bug-04-memory-leak/
└── validation/               # 教程验证测试
    └── validation_test.go
```

### 学习路线

| 阶段 | 章节 | 预计时间 | 前置条件 | 学习重点 |
|------|------|----------|----------|----------|
| 入门 | 01-unit-testing | 3h | 开发教程02 | go test, testify |
| 基础 | 02-integration | 3h | 实战项目todo-api | 数据库测试, API测试 |
| 进阶 | 03-e2e | 2h | 02完成 | 场景测试 |
| 进阶 | 04-performance | 2h | 03完成 | benchmark, pprof |
| 综合 | 05-debugging | 3h | 04完成 | delve, 问题排查 |

---

## Project Structure

### Documentation (this feature)

```text
specs/001-go-backend-learning/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
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
│   ├── unit/
│   ├── integration/
│   ├── e2e/
│   ├── benchmark/
│   ├── debugging/
│   └── validation/
├── infra/                        # 基础设施配置
│   ├── docker-compose.yml
│   ├── postgres/
│   ├── redis/
│   ├── monitoring/
│   └── ci/
├── examples/                     # 独立代码示例
│   ├── syntax/
│   ├── concurrency/
│   └── patterns/
├── go.mod                        # 根模块
├── go.sum
├── README.md                     # 项目说明
├── .gitignore
└── .dockerignore
```

**Structure Decision**: 采用四轨并行结构，将开发、实战、部署、测试分离为独立学习轨道，允许学习者根据自身需求选择学习路径，同时保持轨道间的清晰依赖关系。

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

## Complexity Tracking

> 四轨结构增加了组织复杂度，但提供了更好的模块化和学习灵活性。

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 四轨独立结构 | 允许不同背景学习者选择适合的学习路径 | 单一线性结构无法满足不同学习风格，且测试/部署内容与开发内容交织会增加认知负担 |
| 多项目结构 | 渐进式实战项目展示完整开发流程 | 单一项目无法覆盖从CLI到全栈的完整学习曲线 |
