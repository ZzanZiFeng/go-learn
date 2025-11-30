# 认证授权练习 (Exercises)

## 练习 1：JWT 认证系统

实现完整的 JWT 认证系统。

### 要求

1. 实现以下端点：
   - `POST /auth/register` - 用户注册
   - `POST /auth/login` - 用户登录
   - `POST /auth/refresh` - 刷新 Token
   - `POST /auth/logout` - 登出（Token 黑名单）
   - `GET /auth/me` - 获取当前用户

2. 功能：
   - 使用 bcrypt 哈希密码
   - Access Token 有效期 15 分钟
   - Refresh Token 有效期 7 天
   - 实现 Token 黑名单

### 提示

```go
type AuthService struct {
    userRepo   UserRepository
    jwtService *JWTService
    blacklist  *TokenBlacklist
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}
```

---

## 练习 2：密码安全

实现安全的密码处理。

### 要求

1. 密码哈希服务：
   - 使用 bcrypt，cost = 12
   - 验证密码强度

2. 密码策略：
   - 最少 8 字符
   - 必须包含大写、小写、数字、特殊字符
   - 禁止常见弱密码

3. 密码重置流程：
   - 生成重置令牌（1 小时有效）
   - 验证令牌
   - 更新密码

### 提示

```go
type PasswordPolicy struct {
    MinLength        int
    RequireUppercase bool
    RequireLowercase bool
    RequireNumber    bool
    RequireSpecial   bool
}

type PasswordResetToken struct {
    Token     string
    UserID    uint
    ExpiresAt time.Time
}
```

---

## 练习 3：Session 管理

实现基于 Redis 的 Session 管理。

### 要求

1. Session 功能：
   - 创建 Session
   - 获取 Session
   - 刷新 Session（滑动过期）
   - 销毁 Session
   - 销毁用户所有 Session

2. 安全配置：
   - HttpOnly Cookie
   - Secure (HTTPS)
   - SameSite Strict

3. CSRF 保护

### 提示

```go
type Session struct {
    ID        string
    UserID    uint
    Data      map[string]interface{}
    CreatedAt time.Time
    ExpiresAt time.Time
}

type SessionStore interface {
    Create(ctx context.Context, session *Session) error
    Get(ctx context.Context, id string) (*Session, error)
    Delete(ctx context.Context, id string) error
    DeleteByUserID(ctx context.Context, userID uint) error
}
```

---

## 练习 4：OAuth2 集成

实现 GitHub OAuth2 登录。

### 要求

1. 端点：
   - `GET /oauth/github` - 开始 GitHub 登录
   - `GET /oauth/github/callback` - GitHub 回调

2. 功能：
   - State 参数防止 CSRF
   - 获取用户邮箱和名称
   - 创建或关联用户账号
   - 返回 JWT Token

### 提示

```go
type OAuthHandler struct {
    config      *oauth2.Config
    userService *UserService
    jwtService  *JWTService
}

type GitHubUser struct {
    ID    int64  `json:"id"`
    Login string `json:"login"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

---

## 练习 5：RBAC 系统

实现基于角色的访问控制。

### 要求

1. 数据模型：
   - User -> Role (多对一)
   - Role -> Permission (多对多)

2. 预定义角色：
   - admin: 所有权限
   - moderator: 读取和管理内容
   - user: 基本操作

3. 中间件：
   - RequirePermission(resource, action)
   - RequireRole(roles...)

4. 管理 API：
   - 创建角色
   - 分配权限
   - 分配用户角色

### 提示

```go
type RBACService interface {
    HasPermission(ctx context.Context, userID uint, resource, action string) bool
    HasRole(ctx context.Context, userID uint, roleName string) bool
    GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
}
```

---

## 综合练习：认证服务

构建完整的认证微服务。

### 功能要求

1. **用户管理**
   - 注册、登录、登出
   - 个人信息查看和修改
   - 密码修改和重置

2. **认证方式**
   - 邮箱密码登录
   - JWT Token
   - OAuth2 (GitHub)

3. **授权管理**
   - 角色管理
   - 权限管理
   - RBAC 中间件

4. **安全功能**
   - 密码强度验证
   - 登录限流
   - Token 黑名单
   - 审计日志

### 技术要求

1. 使用分层架构
2. Redis 缓存权限
3. 完整错误处理
4. Swagger 文档

### API 设计

```
# 认证
POST   /auth/register
POST   /auth/login
POST   /auth/logout
POST   /auth/refresh
POST   /auth/forgot-password
POST   /auth/reset-password
GET    /auth/me
PUT    /auth/me

# OAuth
GET    /oauth/github
GET    /oauth/github/callback

# 用户管理 (Admin)
GET    /admin/users
GET    /admin/users/:id
PUT    /admin/users/:id
DELETE /admin/users/:id
PUT    /admin/users/:id/role

# 角色管理 (Admin)
GET    /admin/roles
POST   /admin/roles
GET    /admin/roles/:id
PUT    /admin/roles/:id
DELETE /admin/roles/:id
POST   /admin/roles/:id/permissions
DELETE /admin/roles/:id/permissions/:permId
```

### 数据模型

```go
type User struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    RoleID    uint      `json:"role_id"`
    Role      *Role     `json:"role,omitempty"`
    Status    string    `json:"status"` // active, inactive, banned
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Role struct {
    ID          uint         `json:"id"`
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Permissions []Permission `json:"permissions"`
}

type Permission struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`     // user:read
    Resource string `json:"resource"` // user
    Action   string `json:"action"`   // read
}

type OAuthAccount struct {
    ID       uint   `json:"id"`
    UserID   uint   `json:"user_id"`
    Provider string `json:"provider"` // github, google
    OAuthID  string `json:"oauth_id"`
}

type AuditLog struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    Action    string    `json:"action"`    // login, logout, password_change
    IP        string    `json:"ip"`
    UserAgent string    `json:"user_agent"`
    Details   string    `json:"details"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 项目结构

```
auth-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   ├── password.go
│   │   ├── session.go
│   │   └── rbac.go
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── oauth.go
│   │   ├── user.go
│   │   └── role.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── rbac.go
│   │   └── ratelimit.go
│   ├── models/
│   │   ├── user.go
│   │   ├── role.go
│   │   └── audit.go
│   ├── repositories/
│   │   ├── user.go
│   │   ├── role.go
│   │   └── audit.go
│   └── services/
│       ├── auth.go
│       ├── user.go
│       └── role.go
├── pkg/
│   └── response/
├── configs/
├── migrations/
└── go.mod
```

---

## 提交检查清单

### 基础功能
- [ ] 用户注册和登录正常工作
- [ ] JWT Token 生成和验证正确
- [ ] 密码使用 bcrypt 哈希
- [ ] Token 刷新机制工作
- [ ] 登出时 Token 失效

### 安全功能
- [ ] 密码强度验证
- [ ] 登录限流实现
- [ ] CSRF 保护（Session 模式）
- [ ] 敏感信息不泄露

### RBAC
- [ ] 角色权限模型正确
- [ ] 权限中间件工作
- [ ] 权限缓存实现

### 代码质量
- [ ] 代码结构清晰
- [ ] 错误处理完整
- [ ] 有必要的注释
- [ ] API 文档完整
