# 认证概念 (Authentication Concepts)

## 认证 vs 授权

### 定义

```
认证 (Authentication): 验证 "你是谁"
- 用户名/密码登录
- Token 验证
- 生物识别
- 多因素认证 (MFA)

授权 (Authorization): 验证 "你能做什么"
- 角色检查
- 权限验证
- 资源访问控制
- 操作权限
```

### 流程

```
┌──────────────────────────────────────────────────────────────┐
│                        请求流程                                │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│   用户请求 ──> 认证 ──> 授权 ──> 资源访问                      │
│      │         │        │          │                         │
│      │         │        │          └─ 返回请求的资源           │
│      │         │        │                                    │
│      │         │        └─ 检查权限                           │
│      │         │           - 有权限: 继续                      │
│      │         │           - 无权限: 403 Forbidden            │
│      │         │                                             │
│      │         └─ 验证身份                                    │
│      │            - 成功: 继续                                │
│      │            - 失败: 401 Unauthorized                   │
│      │                                                       │
│      └─ 携带凭据 (Token/Session)                             │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## 认证方式

### 1. 基于 Token 的认证

```go
// JWT Token 流程
1. 用户登录，服务器验证凭据
2. 服务器生成 JWT Token
3. 客户端存储 Token
4. 后续请求携带 Token
5. 服务器验证 Token

// 示例请求
GET /api/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**优点：**
- 无状态，易于扩展
- 可跨域使用
- 适合微服务

**缺点：**
- Token 一旦签发，无法撤销
- Token 可能较大
- 需要处理过期/刷新

### 2. 基于 Session 的认证

```go
// Session 流程
1. 用户登录，服务器验证凭据
2. 服务器创建 Session，存入内存/数据库/Redis
3. 返回 Session ID (通常通过 Cookie)
4. 后续请求自动携带 Cookie
5. 服务器查找 Session 验证

// 示例请求
GET /api/profile
Cookie: session_id=abc123...
```

**优点：**
- 可以随时撤销
- 数据存在服务端，更安全
- Cookie 自动携带

**缺点：**
- 需要服务端存储
- 分布式环境需要共享 Session
- 跨域支持复杂

### 3. API Key 认证

```go
// API Key 流程
GET /api/data
X-API-Key: sk_live_xxxxx

// 或通过 Query
GET /api/data?api_key=sk_live_xxxxx
```

**优点：**
- 实现简单
- 适合服务间调用

**缺点：**
- 安全性较低
- 难以区分用户
- 无过期机制

### 4. OAuth2 认证

```go
// OAuth2 流程 (Authorization Code)
1. 用户点击 "使用 GitHub 登录"
2. 重定向到 GitHub 授权页面
3. 用户授权
4. GitHub 回调，携带 code
5. 服务器用 code 换取 access_token
6. 使用 access_token 获取用户信息
```

## 认证架构

### 单体应用

```
┌─────────────────────────────────────────┐
│              单体应用                     │
├─────────────────────────────────────────┤
│                                         │
│   ┌─────────┐    ┌──────────────────┐  │
│   │  用户   │───>│    认证模块       │  │
│   └─────────┘    │  - 登录/注册      │  │
│                  │  - Token 验证     │  │
│                  │  - Session 管理   │  │
│                  └──────────────────┘  │
│                          │             │
│                          ▼             │
│                  ┌──────────────────┐  │
│                  │    业务模块       │  │
│                  └──────────────────┘  │
│                                         │
└─────────────────────────────────────────┘
```

### 微服务

```
┌─────────────────────────────────────────────────────────────┐
│                        微服务架构                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌─────────┐    ┌─────────────┐    ┌─────────────────┐   │
│   │  用户   │───>│  API Gateway │───>│   认证服务       │   │
│   └─────────┘    └─────────────┘    │  - 用户管理      │   │
│                         │            │  - Token 签发    │   │
│                         │            │  - Token 验证    │   │
│                         ▼            └─────────────────┘   │
│                  ┌─────────────┐                           │
│                  │  业务服务    │                           │
│                  └─────────────┘                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 安全考虑

### 常见攻击

| 攻击类型 | 描述 | 防护措施 |
|---------|------|---------|
| 暴力破解 | 尝试大量密码 | 登录限流，账户锁定 |
| CSRF | 跨站请求伪造 | CSRF Token，SameSite Cookie |
| XSS | 跨站脚本攻击 | HttpOnly Cookie，CSP |
| 中间人攻击 | 截获通信 | HTTPS，证书验证 |
| Session 劫持 | 窃取 Session | Secure Cookie，定期更新 |

### 密码安全

```go
// ✅ 正确做法
- 使用 bcrypt/argon2 哈希
- 添加随机盐 (bcrypt 自动处理)
- 强制密码复杂度
- 限制登录尝试

// ❌ 错误做法
- 明文存储
- MD5/SHA1 哈希
- 固定盐值
- 无登录限制
```

### Token 安全

```go
// ✅ 正确做法
- 使用 HTTPS
- 设置合理过期时间
- 实现刷新机制
- 敏感操作重新认证

// ❌ 错误做法
- HTTP 传输
- Token 永不过期
- 在 URL 中传递 Token
- 存储敏感信息在 Token
```

## Go 实现示例

### 认证中间件架构

```go
package auth

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

// Authenticator 认证器接口
type Authenticator interface {
    Authenticate(c *gin.Context) (*User, error)
}

// JWTAuthenticator JWT 认证器
type JWTAuthenticator struct {
    secretKey []byte
}

func (a *JWTAuthenticator) Authenticate(c *gin.Context) (*User, error) {
    token := extractToken(c)
    if token == "" {
        return nil, ErrMissingToken
    }

    claims, err := validateJWT(token, a.secretKey)
    if err != nil {
        return nil, ErrInvalidToken
    }

    return &User{
        ID:   claims.UserID,
        Role: claims.Role,
    }, nil
}

// SessionAuthenticator Session 认证器
type SessionAuthenticator struct {
    store SessionStore
}

func (a *SessionAuthenticator) Authenticate(c *gin.Context) (*User, error) {
    sessionID, err := c.Cookie("session_id")
    if err != nil {
        return nil, ErrMissingSession
    }

    session, err := a.store.Get(sessionID)
    if err != nil {
        return nil, ErrInvalidSession
    }

    return session.User, nil
}

// AuthMiddleware 认证中间件
func AuthMiddleware(auth Authenticator) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, err := auth.Authenticate(c)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": err.Error(),
            })
            return
        }

        c.Set("user", user)
        c.Next()
    }
}

func extractToken(c *gin.Context) string {
    // 从 Header 获取
    auth := c.GetHeader("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return auth[7:]
    }

    // 从 Query 获取
    if token := c.Query("token"); token != "" {
        return token
    }

    return ""
}
```

### 授权中间件

```go
package auth

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// Authorizer 授权器接口
type Authorizer interface {
    Authorize(user *User, resource, action string) bool
}

// RBACAuthorizer 基于角色的授权
type RBACAuthorizer struct {
    permissions map[string][]string // role -> permissions
}

func (a *RBACAuthorizer) Authorize(user *User, resource, action string) bool {
    perms, ok := a.permissions[user.Role]
    if !ok {
        return false
    }

    required := resource + ":" + action
    for _, p := range perms {
        if p == required || p == "*" {
            return true
        }
    }
    return false
}

// RequirePermission 权限检查中间件
func RequirePermission(auth Authorizer, resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Not authenticated",
            })
            return
        }

        if !auth.Authorize(user.(*User), resource, action) {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "Permission denied",
            })
            return
        }

        c.Next()
    }
}

// RequireRole 角色检查中间件
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Not authenticated",
            })
            return
        }

        userRole := user.(*User).Role
        for _, role := range roles {
            if userRole == role {
                c.Next()
                return
            }
        }

        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
            "error": "Insufficient role",
        })
    }
}
```

## 总结

| 概念 | 描述 |
|------|------|
| 认证 | 验证用户身份 |
| 授权 | 验证用户权限 |
| JWT | 无状态 Token 认证 |
| Session | 有状态会话认证 |
| RBAC | 基于角色的访问控制 |

**下一节**：[JWT 认证](./02-jwt-auth.md) - 学习 JWT 实现
