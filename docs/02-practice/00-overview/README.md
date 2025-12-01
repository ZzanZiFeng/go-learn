# 实战教程概览

## 概述

实战教程（轨道二）通过 4 个递进式项目，将开发教程中学到的知识应用到实际项目中。每个项目都比前一个更复杂，帮助你逐步掌握 Go 后端开发的完整技能。

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         实战项目学习路径                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌────────────────┐  │
│  │   Todo CLI   │───▶│   Todo API   │───▶│ Auth Service │───▶│ Fullstack Demo │  │
│  │    (入门)    │    │    (进阶)    │    │    (进阶)     │    │     (高级)      │  │
│  └──────────────┘    └──────────────┘    └──────────────┘    └────────────────┘  │
│        │                   │                   │                    │           │
│        ▼                   ▼                   ▼                    ▼           │
│   • 基础语法             • Gin框架          • JWT认证           • 系统设计        │
│   • 文件I/O              • GORM ORM         • Session管理       • 前后端集成      │
│   • CLI开发              • REST API         • OAuth2            • Docker部署     │
│                         • Redis缓存        • RBAC权限          • 完整项目        │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## 项目列表

### 项目 1: Todo CLI（入门）

**难度**: ⭐

**目标**: 构建一个命令行待办事项应用

**涉及技术**:
- Go 基础语法
- 文件 I/O（JSON 存储）
- 命令行参数解析
- 项目结构组织

**产出**:
```bash
# 使用示例
todo add "学习 Go 语言"
todo list
todo complete 1
todo delete 1
```

**预计时间**: 2-3 小时

---

### 项目 2: Todo API（进阶）

**难度**: ⭐⭐⭐

**目标**: 将 CLI 升级为 RESTful API 服务

**涉及技术**:
- Gin Web 框架
- GORM + PostgreSQL
- JWT 认证
- Redis 缓存
- API 文档（Swagger）

**产出**:
```bash
# API 端点
GET    /api/todos          # 获取待办列表
POST   /api/todos          # 创建待办
PUT    /api/todos/:id      # 更新待办
DELETE /api/todos/:id      # 删除待办
POST   /api/auth/login     # 用户登录
POST   /api/auth/register  # 用户注册
```

**预计时间**: 6-8 小时

---

### 项目 3: Auth Service（进阶）

**难度**: ⭐⭐⭐

**目标**: 构建完整的认证授权服务

**涉及技术**:
- JWT 生成与验证
- Session 管理
- OAuth2（GitHub 登录）
- RBAC 权限控制

**产出**:
```bash
# API 端点
POST   /api/auth/register      # 注册
POST   /api/auth/login         # 登录
POST   /api/auth/refresh       # 刷新令牌
POST   /api/auth/logout        # 登出
GET    /api/auth/github        # GitHub OAuth2
GET    /api/auth/github/callback
GET    /api/users/me           # 当前用户
PUT    /api/users/me           # 更新资料
```

**预计时间**: 6-8 小时

---

### 项目 4: Fullstack Demo（高级）

**难度**: ⭐⭐⭐⭐⭐

**目标**: 构建完整的全栈应用

**涉及技术**:
- 系统设计与架构
- 微服务通信
- Docker 容器化
- 前后端集成
- CI/CD 部署

**产出**:
```
fullstack-demo/
├── backend/               # Go 后端服务
│   ├── api-gateway/       # API 网关
│   ├── user-service/      # 用户服务
│   ├── order-service/     # 订单服务
│   └── notification/      # 通知服务
├── frontend/              # 前端（可选）
├── docker-compose.yml     # 容器编排
└── Makefile               # 构建脚本
```

**预计时间**: 12-16 小时

---

## 学习路径

### 推荐顺序

```
完成 DEV-00 ~ DEV-01 ──▶ Todo CLI ──▶ 完成 DEV-02 ~ DEV-08 ──▶ Todo API
                                                                  │
                                                                  ▼
                        完成 DEV-09 ~ DEV-11 ◀──────────── Auth Service
                                                                  │
                                                                  ▼
                                                        Fullstack Demo
```

### 前置知识要求

| 项目 | 必须完成的章节 |
|------|----------------|
| Todo CLI | DEV-00, DEV-01 |
| Todo API | DEV-00 ~ DEV-08 |
| Auth Service | DEV-00 ~ DEV-08 |
| Fullstack Demo | DEV-00 ~ DEV-11 |

### 技能对应

| 技能 | Todo CLI | Todo API | Auth Service | Fullstack |
|------|----------|----------|--------------|-----------|
| 基础语法 | ✅ | ✅ | ✅ | ✅ |
| 结构体/接口 | ⬜ | ✅ | ✅ | ✅ |
| 并发编程 | ⬜ | ⬜ | ⬜ | ✅ |
| Web API | ⬜ | ✅ | ✅ | ✅ |
| 认证授权 | ⬜ | ✅ | ✅ | ✅ |
| 数据库 | ⬜ | ✅ | ✅ | ✅ |
| 缓存 | ⬜ | ✅ | ✅ | ✅ |
| 消息队列 | ⬜ | ⬜ | ⬜ | ✅ |
| 可观测性 | ⬜ | ⬜ | ⬜ | ✅ |

---

## 项目结构规范

### 目录布局

```
project-name/
├── cmd/                    # 主程序入口
│   └── api/
│       └── main.go
├── internal/               # 私有代码
│   ├── handlers/           # HTTP 处理器
│   ├── services/           # 业务逻辑
│   ├── repositories/       # 数据访问
│   ├── models/             # 数据模型
│   ├── middleware/         # 中间件
│   └── config/             # 配置
├── pkg/                    # 公共代码
├── api/                    # API 定义（Swagger）
├── configs/                # 配置文件
├── migrations/             # 数据库迁移
├── scripts/                # 脚本
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

### 代码规范

```go
// 1. 文件命名: snake_case
user_handler.go
user_service.go

// 2. 包命名: 小写单词
package handlers
package services

// 3. 接口命名: 动词+er
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id int64) (*User, error)
}

// 4. 错误处理: 包装上下文
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// 5. 配置: 使用结构体
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
}
```

---

## 开发环境

### 基础设施

所有项目共享基础设施，在 `infra/` 目录下：

```bash
# 启动所有服务
cd infra
docker-compose up -d

# 服务访问
PostgreSQL: localhost:5432
Redis:      localhost:6379
RabbitMQ:   localhost:5672 (管理UI: 15672)
Prometheus: localhost:9090
Grafana:    localhost:3000
```

### 开发工具

```bash
# 安装开发工具
go install github.com/cosmtrek/air@latest           # 热重载
go install github.com/swaggo/swag/cmd/swag@latest   # Swagger
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest  # 迁移

# 使用 air 开发
air init
air
```

---

## 验收标准

每个项目完成后，应该能够：

### Todo CLI

- [ ] 添加/列出/完成/删除待办事项
- [ ] 数据持久化到 JSON 文件
- [ ] 友好的命令行提示
- [ ] 代码结构清晰

### Todo API

- [ ] 完整的 CRUD API
- [ ] JWT 认证
- [ ] 数据持久化到 PostgreSQL
- [ ] Redis 缓存
- [ ] Swagger 文档
- [ ] 单元测试覆盖

### Auth Service

- [ ] 用户注册/登录
- [ ] JWT 令牌管理
- [ ] GitHub OAuth2 登录
- [ ] RBAC 权限控制
- [ ] 完善的错误处理

### Fullstack Demo

- [ ] 完整的业务流程
- [ ] Docker 容器化
- [ ] 服务间通信
- [ ] 可观测性集成
- [ ] 部署文档

---

## 下一步

开始你的第一个实战项目：

**[项目 1: Todo CLI](../01-todo-cli/README.md)** - 从简单的命令行应用开始你的 Go 实战之旅！
