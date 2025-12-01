# 项目 3: Auth Service

## 项目概述

构建一个完整的认证授权服务，支持 OAuth2、RBAC 权限管理。

| 难度 | 预计时间 | 前置项目 |
|------|----------|----------|
| 中高级 | 8-10 小时 | Todo API |

## 学习目标

- OAuth2 授权流程
- 角色和权限管理（RBAC）
- 刷新令牌机制
- 密码加密和安全最佳实践

## 技术栈

```yaml
框架: Gin
数据库: PostgreSQL
缓存: Redis
认证: JWT + OAuth2
加密: bcrypt, argon2
```

## API 设计

### 认证端点

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |
| POST | /api/auth/refresh | 刷新令牌 |
| POST | /api/auth/logout | 用户登出 |
| POST | /api/auth/forgot-password | 忘记密码 |
| POST | /api/auth/reset-password | 重置密码 |

### OAuth2 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/oauth/authorize | OAuth2 授权 |
| POST | /api/oauth/token | 获取令牌 |
| GET | /api/oauth/callback/:provider | 第三方回调 |

### 用户管理端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/users/me | 获取当前用户 |
| PUT | /api/users/me | 更新当前用户 |
| PUT | /api/users/me/password | 修改密码 |

### 角色权限端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/roles | 获取角色列表 |
| POST | /api/roles | 创建角色 |
| PUT | /api/roles/:id | 更新角色 |
| DELETE | /api/roles/:id | 删除角色 |
| POST | /api/users/:id/roles | 分配角色 |

## 项目结构

```
projects/auth-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   ├── user.go
│   │   ├── role.go
│   │   ├── permission.go
│   │   └── token.go
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── role.go
│   │   └── oauth.go
│   ├── services/
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── token.go
│   │   └── oauth.go
│   ├── repositories/
│   │   ├── user.go
│   │   ├── role.go
│   │   └── token.go
│   └── middleware/
│       ├── auth.go
│       └── rbac.go
├── configs/
│   └── config.yaml
├── migrations/
│   └── 001_init.sql
└── go.mod
```

## 核心概念

### RBAC 模型

```
┌──────────┐     ┌──────────┐     ┌─────────────┐
│   User   │────▶│   Role   │────▶│ Permission  │
└──────────┘     └──────────┘     └─────────────┘
     │                │                   │
     │ has_many       │ has_many          │
     ▼                ▼                   ▼
  用户表          角色表            权限表
```

### 令牌流程

```
┌────────────┐     ┌────────────┐     ┌────────────┐
│   Login    │────▶│Access Token│────▶│  API Call  │
└────────────┘     └────────────┘     └────────────┘
      │                  │
      │                  │ 过期
      ▼                  ▼
┌────────────┐     ┌────────────┐
│Refresh Token│────▶│ New Tokens │
└────────────┘     └────────────┘
```

## 学习步骤

1. **[Step 1: 项目初始化](./step-01-setup.md)** - 搭建项目结构
2. **[Step 2: 用户认证](./step-02-auth.md)** - 实现注册、登录
3. **[Step 3: 令牌管理](./step-03-token.md)** - Access Token + Refresh Token
4. **[Step 4: RBAC 权限](./step-04-rbac.md)** - 角色和权限管理
5. **[Step 5: OAuth2 集成](./step-05-oauth.md)** - 第三方登录

## 知识点映射

| 步骤 | Go 知识点 | 开发轨道章节 |
|------|-----------|--------------|
| Step 1 | 项目结构、配置管理 | DEV-04, DEV-05 |
| Step 2 | 密码加密、表单验证 | DEV-05, DEV-06 |
| Step 3 | JWT、Redis | DEV-06, DEV-09 |
| Step 4 | 中间件、权限检查 | DEV-05 |
| Step 5 | HTTP 客户端、OAuth2 | DEV-05 |

## 安全要点

### 密码安全

```go
// 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// 验证密码
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### 令牌安全

- Access Token: 短期有效（15-30分钟）
- Refresh Token: 长期有效（7-30天），存储在 Redis
- 支持令牌撤销（黑名单机制）

### RBAC 检查

```go
// 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("user_id")
        if !rbacService.HasPermission(userID, permission) {
            c.AbortWithStatusJSON(403, gin.H{"error": "权限不足"})
            return
        }
        c.Next()
    }
}
```

## 下一步

完成 Auth Service 后，你将掌握企业级认证系统的完整实现。

继续学习: **[项目 4: Fullstack Demo](../04-fullstack-demo/README.md)** - 整合所有项目！
