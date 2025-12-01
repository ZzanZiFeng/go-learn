# 项目 2: Todo API

## 项目概述

将 Todo CLI 升级为 RESTful API 服务，学习 Web 开发的完整流程。

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Todo API 架构                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│    客户端              Web 层               业务层              数据层    │
│  ┌─────────┐       ┌──────────┐       ┌──────────┐       ┌──────────┐  │
│  │ HTTP    │──────▶│ Handlers │──────▶│ Services │──────▶│PostgreSQL│  │
│  │ Client  │       │ (Gin)    │       │          │       │          │  │
│  └─────────┘       └──────────┘       └──────────┘       └──────────┘  │
│                          │                  │                  │        │
│                          │                  │                  ▼        │
│                    ┌──────────┐       ┌──────────┐       ┌──────────┐  │
│                    │Middleware│       │   Cache  │──────▶│  Redis   │  │
│                    │ (Auth)   │       │          │       │          │  │
│                    └──────────┘       └──────────┘       └──────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## API 端点

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | /api/auth/register | 用户注册 | 否 |
| POST | /api/auth/login | 用户登录 | 否 |
| GET | /api/todos | 获取待办列表 | 是 |
| POST | /api/todos | 创建待办 | 是 |
| GET | /api/todos/:id | 获取单个待办 | 是 |
| PUT | /api/todos/:id | 更新待办 | 是 |
| DELETE | /api/todos/:id | 删除待办 | 是 |
| PATCH | /api/todos/:id/complete | 完成待办 | 是 |

## 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| Web 框架 | Gin | 高性能 HTTP 框架 |
| ORM | GORM | Go ORM 库 |
| 数据库 | PostgreSQL | 关系型数据库 |
| 缓存 | Redis | 内存缓存 |
| 认证 | JWT | JSON Web Token |
| 文档 | Swagger | API 文档 |

## 项目结构

```
projects/todo-api/
├── cmd/
│   └── api/
│       └── main.go           # 入口
├── internal/
│   ├── config/
│   │   └── config.go         # 配置管理
│   ├── models/
│   │   ├── user.go           # 用户模型
│   │   └── todo.go           # 待办模型
│   ├── handlers/
│   │   ├── auth.go           # 认证处理器
│   │   └── todo.go           # 待办处理器
│   ├── services/
│   │   ├── auth.go           # 认证服务
│   │   └── todo.go           # 待办服务
│   ├── repositories/
│   │   ├── user.go           # 用户仓储
│   │   └── todo.go           # 待办仓储
│   ├── middleware/
│   │   ├── auth.go           # JWT 中间件
│   │   └── logger.go         # 日志中间件
│   └── cache/
│       └── redis.go          # Redis 缓存
├── api/
│   └── swagger/              # Swagger 文档
├── configs/
│   └── config.yaml           # 配置文件
├── migrations/
│   └── 001_init.sql          # 数据库迁移
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

## 学习步骤

### Step 1: 项目初始化

[▶ step-01-setup.md](./step-01-setup.md)

- 项目结构搭建
- 配置管理（Viper）
- 数据库连接

### Step 2: 路由与处理器

[▶ step-02-routes.md](./step-02-routes.md)

- Gin 路由设置
- CRUD 处理器
- 请求验证

### Step 3: 数据库集成

[▶ step-03-database.md](./step-03-database.md)

- GORM 模型定义
- Repository 模式
- 数据库迁移

### Step 4: JWT 认证

[▶ step-04-auth.md](./step-04-auth.md)

- 用户注册/登录
- JWT 生成与验证
- 认证中间件

### Step 5: Redis 缓存

[▶ step-05-cache.md](./step-05-cache.md)

- Redis 连接
- 缓存策略
- 缓存失效

## 预期产出

### 用户注册

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com","password":"secret123"}'
```

```json
{
  "message": "注册成功",
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com"
  }
}
```

### 用户登录

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john",
    "email": "john@example.com"
  }
}
```

### 创建待办

```bash
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -d '{"title":"学习 Go 语言","description":"完成基础语法学习"}'
```

```json
{
  "id": 1,
  "title": "学习 Go 语言",
  "description": "完成基础语法学习",
  "completed": false,
  "created_at": "2024-01-01T10:00:00Z"
}
```

### 获取待办列表

```bash
curl http://localhost:8080/api/todos \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

```json
{
  "todos": [
    {
      "id": 1,
      "title": "学习 Go 语言",
      "description": "完成基础语法学习",
      "completed": false,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

## 运行项目

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f api

# 停止服务
docker-compose down
```

### 本地开发

```bash
# 启动基础设施
cd infra && docker-compose up -d postgres redis

# 运行 API
cd projects/todo-api
go run cmd/api/main.go

# 使用 air 热重载
air
```

## 前置知识

开始本项目前，请确保已完成：

- [x] DEV-00: 环境搭建
- [x] DEV-01: 语法基础
- [x] DEV-02: 结构体与接口
- [x] DEV-04: 后端架构
- [x] DEV-05: Web API 开发
- [x] DEV-06: 认证授权
- [x] DEV-07: 数据库基础
- [x] DEV-08: GORM 框架

## 下一步

准备好了吗？让我们开始：

**[Step 1: 项目初始化](./step-01-setup.md)**
