# 会话管理 (Session Management)

## 概述

Session 是一种有状态的认证方式，用户状态存储在服务端。

## Session vs JWT

| 特性 | Session | JWT |
|------|---------|-----|
| 状态 | 有状态 | 无状态 |
| 存储 | 服务端 | 客户端 |
| 撤销 | 即时 | 困难 |
| 扩展性 | 需要共享存储 | 天然分布式 |
| 安全性 | 更可控 | 依赖签名 |
| 适用场景 | 传统 Web | API/微服务 |

## 安装

```bash
# Gorilla Sessions
go get github.com/gorilla/sessions

# Gin Session (基于 Gorilla)
go get github.com/gin-contrib/sessions
go get github.com/gin-contrib/sessions/cookie
go get github.com/gin-contrib/sessions/redis
```

## 基础实现

### Cookie Store

```go
package main

import (
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 创建 Cookie Store
    store := cookie.NewStore([]byte("secret-key"))

    // 配置 Session
    store.Options(sessions.Options{
        MaxAge:   3600 * 24, // 24 小时
        Path:     "/",
        HttpOnly: true,
        Secure:   true,      // 仅 HTTPS
        SameSite: http.SameSiteStrictMode,
    })

    r.Use(sessions.Sessions("session", store))

    r.GET("/login", func(c *gin.Context) {
        session := sessions.Default(c)
        session.Set("userID", 123)
        session.Set("role", "admin")
        session.Save()
        c.JSON(200, gin.H{"message": "logged in"})
    })

    r.GET("/profile", func(c *gin.Context) {
        session := sessions.Default(c)
        userID := session.Get("userID")
        if userID == nil {
            c.JSON(401, gin.H{"error": "not logged in"})
            return
        }
        c.JSON(200, gin.H{"user_id": userID})
    })

    r.GET("/logout", func(c *gin.Context) {
        session := sessions.Default(c)
        session.Clear()
        session.Save()
        c.JSON(200, gin.H{"message": "logged out"})
    })

    r.Run(":8080")
}
```

### Redis Store

```go
package main

import (
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/redis"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 创建 Redis Store
    store, err := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
    if err != nil {
        panic(err)
    }

    store.Options(sessions.Options{
        MaxAge:   3600 * 24,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
    })

    r.Use(sessions.Sessions("session", store))

    // ... 路由
    r.Run(":8080")
}
```

## 自定义 Session 服务

### Session 接口

```go
package auth

import (
    "context"
    "time"
)

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
    Update(ctx context.Context, session *Session) error
    Delete(ctx context.Context, id string) error
    DeleteByUserID(ctx context.Context, userID uint) error
}
```

### Redis 实现

```go
package auth

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisSessionStore struct {
    client *redis.Client
    prefix string
    ttl    time.Duration
}

func NewRedisSessionStore(client *redis.Client, ttl time.Duration) *RedisSessionStore {
    return &RedisSessionStore{
        client: client,
        prefix: "session:",
        ttl:    ttl,
    }
}

func (s *RedisSessionStore) Create(ctx context.Context, session *Session) error {
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }

    // 存储 Session
    key := s.prefix + session.ID
    if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
        return err
    }

    // 建立 User -> Session 映射（支持多设备）
    userKey := s.prefix + "user:" + fmt.Sprint(session.UserID)
    s.client.SAdd(ctx, userKey, session.ID)
    s.client.Expire(ctx, userKey, s.ttl)

    return nil
}

func (s *RedisSessionStore) Get(ctx context.Context, id string) (*Session, error) {
    key := s.prefix + id
    data, err := s.client.Get(ctx, key).Bytes()
    if err != nil {
        if err == redis.Nil {
            return nil, ErrSessionNotFound
        }
        return nil, err
    }

    var session Session
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, err
    }

    return &session, nil
}

func (s *RedisSessionStore) Update(ctx context.Context, session *Session) error {
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }

    key := s.prefix + session.ID
    return s.client.Set(ctx, key, data, s.ttl).Err()
}

func (s *RedisSessionStore) Delete(ctx context.Context, id string) error {
    // 获取 Session 以移除用户映射
    session, err := s.Get(ctx, id)
    if err != nil {
        return err
    }

    // 删除 Session
    key := s.prefix + id
    s.client.Del(ctx, key)

    // 从用户映射中移除
    userKey := s.prefix + "user:" + fmt.Sprint(session.UserID)
    s.client.SRem(ctx, userKey, id)

    return nil
}

func (s *RedisSessionStore) DeleteByUserID(ctx context.Context, userID uint) error {
    userKey := s.prefix + "user:" + fmt.Sprint(userID)

    // 获取用户的所有 Session
    sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
    if err != nil {
        return err
    }

    // 删除所有 Session
    for _, sessionID := range sessionIDs {
        s.client.Del(ctx, s.prefix+sessionID)
    }

    // 删除用户映射
    s.client.Del(ctx, userKey)

    return nil
}
```

### Session 服务

```go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "time"
)

type SessionService struct {
    store SessionStore
    ttl   time.Duration
}

func NewSessionService(store SessionStore, ttl time.Duration) *SessionService {
    return &SessionService{
        store: store,
        ttl:   ttl,
    }
}

func (s *SessionService) Create(ctx context.Context, userID uint) (*Session, error) {
    // 生成 Session ID
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return nil, err
    }
    sessionID := hex.EncodeToString(bytes)

    session := &Session{
        ID:        sessionID,
        UserID:    userID,
        Data:      make(map[string]interface{}),
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(s.ttl),
    }

    if err := s.store.Create(ctx, session); err != nil {
        return nil, err
    }

    return session, nil
}

func (s *SessionService) Get(ctx context.Context, sessionID string) (*Session, error) {
    session, err := s.store.Get(ctx, sessionID)
    if err != nil {
        return nil, err
    }

    // 检查过期
    if time.Now().After(session.ExpiresAt) {
        s.store.Delete(ctx, sessionID)
        return nil, ErrSessionExpired
    }

    return session, nil
}

func (s *SessionService) Refresh(ctx context.Context, sessionID string) (*Session, error) {
    session, err := s.Get(ctx, sessionID)
    if err != nil {
        return nil, err
    }

    session.ExpiresAt = time.Now().Add(s.ttl)
    if err := s.store.Update(ctx, session); err != nil {
        return nil, err
    }

    return session, nil
}

func (s *SessionService) Destroy(ctx context.Context, sessionID string) error {
    return s.store.Delete(ctx, sessionID)
}

func (s *SessionService) DestroyAllForUser(ctx context.Context, userID uint) error {
    return s.store.DeleteByUserID(ctx, userID)
}

func (s *SessionService) SetData(ctx context.Context, sessionID, key string, value interface{}) error {
    session, err := s.Get(ctx, sessionID)
    if err != nil {
        return err
    }

    session.Data[key] = value
    return s.store.Update(ctx, session)
}
```

## Gin 集成

### Session 中间件

```go
package middleware

import (
    "net/http"

    "myapp/internal/auth"

    "github.com/gin-gonic/gin"
)

func SessionMiddleware(sessionService *auth.SessionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Cookie 获取 Session ID
        sessionID, err := c.Cookie("session_id")
        if err != nil || sessionID == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "No session",
            })
            return
        }

        // 获取 Session
        session, err := sessionService.Get(c.Request.Context(), sessionID)
        if err != nil {
            // 清除无效 Cookie
            c.SetCookie("session_id", "", -1, "/", "", true, true)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid session",
            })
            return
        }

        // 存入 Context
        c.Set("session", session)
        c.Set("userID", session.UserID)

        // 刷新 Session（滑动过期）
        sessionService.Refresh(c.Request.Context(), sessionID)

        c.Next()
    }
}

func GetSession(c *gin.Context) *auth.Session {
    session, _ := c.Get("session")
    return session.(*auth.Session)
}
```

### 认证处理器

```go
package handlers

type AuthHandler struct {
    userService    *services.UserService
    sessionService *auth.SessionService
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 验证凭据
    user, err := h.userService.ValidateCredentials(req.Email, req.Password)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }

    // 创建 Session
    session, err := h.sessionService.Create(c.Request.Context(), user.ID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create session"})
        return
    }

    // 设置 Cookie
    c.SetSameSite(http.SameSiteStrictMode)
    c.SetCookie(
        "session_id",
        session.ID,
        int(session.ExpiresAt.Sub(time.Now()).Seconds()),
        "/",
        "",
        true,  // Secure
        true,  // HttpOnly
    )

    c.JSON(200, gin.H{
        "user":    user,
        "message": "Logged in",
    })
}

func (h *AuthHandler) Logout(c *gin.Context) {
    session := middleware.GetSession(c)

    // 销毁 Session
    h.sessionService.Destroy(c.Request.Context(), session.ID)

    // 清除 Cookie
    c.SetCookie("session_id", "", -1, "/", "", true, true)

    c.JSON(200, gin.H{"message": "Logged out"})
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
    session := middleware.GetSession(c)

    // 销毁用户所有 Session
    h.sessionService.DestroyAllForUser(c.Request.Context(), session.UserID)

    // 清除 Cookie
    c.SetCookie("session_id", "", -1, "/", "", true, true)

    c.JSON(200, gin.H{"message": "Logged out from all devices"})
}
```

## CSRF 保护

```go
package middleware

import (
    "crypto/rand"
    "encoding/hex"
    "net/http"

    "github.com/gin-gonic/gin"
)

func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 只对修改请求检查
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            c.Next()
            return
        }

        // 获取 Session
        session := GetSession(c)
        if session == nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No session"})
            return
        }

        // 获取 CSRF Token
        csrfToken := c.GetHeader("X-CSRF-Token")
        if csrfToken == "" {
            csrfToken = c.PostForm("csrf_token")
        }

        // 验证 Token
        storedToken, ok := session.Data["csrf_token"].(string)
        if !ok || csrfToken != storedToken {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
            return
        }

        c.Next()
    }
}

func GenerateCSRFToken(sessionService *auth.SessionService, session *auth.Session) (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    token := hex.EncodeToString(bytes)

    sessionService.SetData(context.Background(), session.ID, "csrf_token", token)
    return token, nil
}
```

## 安全配置

```go
// Cookie 安全配置
c.SetSameSite(http.SameSiteStrictMode)  // 防止 CSRF
c.SetCookie(
    "session_id",
    session.ID,
    maxAge,
    "/",       // Path
    "",        // Domain
    true,      // Secure (HTTPS only)
    true,      // HttpOnly (防止 XSS)
)

// Session 配置
store.Options(sessions.Options{
    MaxAge:   3600 * 24,                    // 24 小时
    Path:     "/",
    Domain:   "",
    Secure:   true,                         // 仅 HTTPS
    HttpOnly: true,                         // 防止 JS 访问
    SameSite: http.SameSiteStrictMode,      // 防止 CSRF
})
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| HttpOnly | 防止 XSS 获取 Cookie |
| Secure | 仅通过 HTTPS 传输 |
| SameSite | 防止 CSRF 攻击 |
| 随机 Session ID | 使用足够长的随机值 |
| 滑动过期 | 活跃用户自动延长 |
| 登出时销毁 | 不仅清除 Cookie |
| 使用 Redis | 分布式环境共享 Session |
| 定期轮换 | 重新生成 Session ID |

**下一节**：[OAuth2](./05-oauth2.md) - 学习第三方登录
