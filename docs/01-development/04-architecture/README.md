# 后端架构 (Backend Architecture)

## 学习目标

完成本章后，你将能够：
- 理解 Go 标准项目布局
- 掌握分层架构模式 (Handler/Service/Repository)
- 实现依赖注入
- 使用 Viper 管理配置
- 设计良好的错误处理策略
- 了解微服务基础概念

## 与 JavaScript/Node.js 的对比

| 概念 | Node.js | Go |
|------|---------|-----|
| 项目结构 | 无标准（各框架不同） | [Standard Go Project Layout](https://github.com/golang-standards/project-layout) |
| 分层架构 | Controller/Service/Repository | Handler/Service/Repository |
| 依赖注入 | InversifyJS, NestJS DI | 手动注入 / Wire |
| 配置管理 | dotenv, config | Viper |
| 错误处理 | try/catch, 中间件 | 显式错误返回 |

## 核心概念

### Go 标准项目布局

```
myapp/
├── cmd/                    # 应用入口
│   ├── api/
│   │   └── main.go        # API 服务入口
│   └── worker/
│       └── main.go        # 后台工作进程入口
├── internal/              # 私有代码（不可被外部导入）
│   ├── handlers/          # HTTP 处理器
│   ├── services/          # 业务逻辑
│   ├── repositories/      # 数据访问层
│   ├── models/            # 数据模型
│   └── middleware/        # 中间件
├── pkg/                   # 可被外部导入的公共库
├── configs/               # 配置文件
├── scripts/               # 脚本
├── docs/                  # 文档
└── go.mod
```

### 分层架构

```
                    HTTP Request
                         │
                         ▼
┌─────────────────────────────────────────┐
│              Middleware                  │  认证、日志、恢复
└─────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────┐
│              Handlers                    │  请求解析、响应格式化
│         (Presentation Layer)             │
└─────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────┐
│              Services                    │  业务逻辑
│          (Business Layer)                │
└─────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────┐
│            Repositories                  │  数据访问
│            (Data Layer)                  │
└─────────────────────────────────────────┘
                         │
                         ▼
                   Database/Cache
```

## 章节内容

1. [项目布局](./01-project-layout.md) - Go 标准项目结构
2. [分层架构](./02-layered-arch.md) - Handler/Service/Repository 模式
3. [依赖注入](./03-dependency-injection.md) - 依赖注入模式和 Wire
4. [配置管理](./04-config-management.md) - Viper 配置管理
5. [错误设计](./05-error-design.md) - 错误处理策略
6. [微服务入门](./06-microservices-intro.md) - 微服务概念和 API 网关

## 为什么需要架构？

### Node.js 中的问题

```javascript
// 典型的 Express 代码 - 混乱的单文件
app.post('/users', async (req, res) => {
  // 验证
  if (!req.body.email) {
    return res.status(400).json({ error: 'Email required' });
  }

  // 业务逻辑（直接在 handler 中）
  const existingUser = await db.query('SELECT * FROM users WHERE email = ?', [req.body.email]);
  if (existingUser) {
    return res.status(409).json({ error: 'User exists' });
  }

  // 数据访问（直接 SQL）
  const result = await db.query('INSERT INTO users ...');

  // 发送邮件（副作用混在主逻辑中）
  await sendWelcomeEmail(req.body.email);

  res.json(result);
});
```

### Go 分层架构的解决方案

```go
// Handler - 只处理 HTTP
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.service.CreateUser(c.Request.Context(), req)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, user)
}

// Service - 业务逻辑
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // 检查用户是否存在
    exists, err := s.repo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrUserExists
    }

    // 创建用户
    user, err := s.repo.Create(ctx, &User{Email: req.Email, Name: req.Name})
    if err != nil {
        return nil, err
    }

    // 发送欢迎邮件（异步）
    go s.emailService.SendWelcome(user.Email)

    return user, nil
}

// Repository - 数据访问
func (r *UserRepository) Create(ctx context.Context, user *User) (*User, error) {
    return user, r.db.WithContext(ctx).Create(user).Error
}
```

## 架构选择指南

| 项目规模 | 推荐架构 | 理由 |
|---------|---------|------|
| 小型/原型 | 简单分层 | 快速开发，易于理解 |
| 中型 | 清洁架构 | 可测试性，关注点分离 |
| 大型/团队 | 领域驱动设计 (DDD) | 业务复杂性管理 |
| 微服务 | 六边形架构 | 端口和适配器，易于替换依赖 |

## 实践项目

本章的知识将在以下项目中应用：
- [Todo API](../../02-practice/todo-api/) - 完整的分层架构实践
- [Auth Service](../../02-practice/auth-service/) - 服务拆分和 API 网关

## 下一步

完成本章后，继续学习 [Web API 开发](../05-web-api/)，将架构知识应用到实际 HTTP 服务开发中。
