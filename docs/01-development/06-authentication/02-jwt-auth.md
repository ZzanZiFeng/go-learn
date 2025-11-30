# JWT 认证 (JWT Authentication)

## 什么是 JWT？

JWT (JSON Web Token) 是一种开放标准 (RFC 7519)，用于在各方之间安全传输信息。

## JWT 结构

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJyb2xlIjoiYWRtaW4ifQ.signature
└───────────────┬──────────────┘.└─────────────┬─────────────┘.└───┬────┘
              Header                        Payload              Signature
```

### Header

```json
{
  "alg": "HS256",  // 算法：HS256, RS256 等
  "typ": "JWT"     // 类型
}
```

### Payload (Claims)

```json
{
  "user_id": 1,           // 自定义数据
  "role": "admin",        // 自定义数据
  "exp": 1704067200,      // 过期时间 (标准)
  "iat": 1703980800,      // 签发时间 (标准)
  "iss": "my-app",        // 签发者 (标准)
  "sub": "user@example.com" // 主题 (标准)
}
```

### Signature

```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret
)
```

## 安装

```bash
go get -u github.com/golang-jwt/jwt/v5
```

## 基础实现

### 定义 Claims

```go
package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT Claims
type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

// TokenPair Access Token 和 Refresh Token
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}
```

### JWT 服务

```go
package auth

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrExpiredToken = errors.New("token expired")
)

type JWTService struct {
    secretKey       []byte
    accessTokenTTL  time.Duration
    refreshTokenTTL time.Duration
    issuer          string
}

func NewJWTService(secret string) *JWTService {
    return &JWTService{
        secretKey:       []byte(secret),
        accessTokenTTL:  15 * time.Minute,
        refreshTokenTTL: 7 * 24 * time.Hour,
        issuer:          "my-app",
    }
}

// GenerateTokenPair 生成 Token 对
func (s *JWTService) GenerateTokenPair(userID uint, email, role string) (*TokenPair, error) {
    // Access Token
    accessToken, err := s.generateToken(userID, email, role, s.accessTokenTTL)
    if err != nil {
        return nil, err
    }

    // Refresh Token
    refreshToken, err := s.generateToken(userID, email, role, s.refreshTokenTTL)
    if err != nil {
        return nil, err
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
    }, nil
}

func (s *JWTService) generateToken(userID uint, email, role string, ttl time.Duration) (string, error) {
    now := time.Now()

    claims := Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            Issuer:    s.issuer,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secretKey)
}

// ValidateToken 验证 Token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // 验证签名算法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.secretKey, nil
    })

    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }

    return claims, nil
}

// RefreshToken 刷新 Token
func (s *JWTService) RefreshToken(refreshToken string) (*TokenPair, error) {
    claims, err := s.ValidateToken(refreshToken)
    if err != nil {
        return nil, err
    }

    return s.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
}
```

## Gin 集成

### 认证中间件

```go
package middleware

import (
    "net/http"
    "strings"

    "myapp/internal/auth"

    "github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取 Token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "code":    "MISSING_TOKEN",
                "message": "Authorization header required",
            })
            return
        }

        // 解析 Bearer Token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "code":    "INVALID_TOKEN_FORMAT",
                "message": "Invalid authorization format",
            })
            return
        }

        // 验证 Token
        claims, err := jwtService.ValidateToken(parts[1])
        if err != nil {
            code := "INVALID_TOKEN"
            if err == auth.ErrExpiredToken {
                code = "TOKEN_EXPIRED"
            }
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "code":    code,
                "message": err.Error(),
            })
            return
        }

        // 将用户信息存入 Context
        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)
        c.Set("claims", claims)

        c.Next()
    }
}

// 辅助函数
func GetUserID(c *gin.Context) uint {
    return c.GetUint("userID")
}

func GetUserRole(c *gin.Context) string {
    return c.GetString("role")
}

func GetClaims(c *gin.Context) *auth.Claims {
    claims, _ := c.Get("claims")
    return claims.(*auth.Claims)
}
```

### 认证处理器

```go
package handlers

import (
    "net/http"

    "myapp/internal/auth"
    "myapp/internal/services"

    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    userService *services.UserService
    jwtService  *auth.JWTService
}

func NewAuthHandler(userService *services.UserService, jwtService *auth.JWTService) *AuthHandler {
    return &AuthHandler{
        userService: userService,
        jwtService:  jwtService,
    }
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 验证用户
    user, err := h.userService.ValidateCredentials(req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "code":    "INVALID_CREDENTIALS",
            "message": "Invalid email or password",
        })
        return
    }

    // 生成 Token
    tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "user":   user,
        "tokens": tokens,
    })
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 创建用户
    user, err := h.userService.Create(req.Name, req.Email, req.Password)
    if err != nil {
        if err == services.ErrEmailExists {
            c.JSON(http.StatusConflict, gin.H{
                "code":    "EMAIL_EXISTS",
                "message": "Email already registered",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    // 生成 Token
    tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "user":   user,
        "tokens": tokens,
    })
}

// Refresh 刷新 Token
func (h *AuthHandler) Refresh(c *gin.Context) {
    var req RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    tokens, err := h.jwtService.RefreshToken(req.RefreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "code":    "INVALID_REFRESH_TOKEN",
            "message": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, tokens)
}

// Me 获取当前用户
func (h *AuthHandler) Me(c *gin.Context) {
    userID := c.GetUint("userID")

    user, err := h.userService.GetByID(userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, user)
}
```

## 高级功能

### Token 黑名单

```go
package auth

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type TokenBlacklist struct {
    redis *redis.Client
}

func NewTokenBlacklist(redis *redis.Client) *TokenBlacklist {
    return &TokenBlacklist{redis: redis}
}

// Add 将 Token 加入黑名单
func (b *TokenBlacklist) Add(ctx context.Context, token string, expiration time.Duration) error {
    return b.redis.Set(ctx, "blacklist:"+token, "1", expiration).Err()
}

// IsBlacklisted 检查 Token 是否在黑名单
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) bool {
    exists, _ := b.redis.Exists(ctx, "blacklist:"+token).Result()
    return exists > 0
}

// ValidateToken 带黑名单检查的验证
func (s *JWTService) ValidateTokenWithBlacklist(ctx context.Context, tokenString string, blacklist *TokenBlacklist) (*Claims, error) {
    // 检查黑名单
    if blacklist.IsBlacklisted(ctx, tokenString) {
        return nil, ErrInvalidToken
    }

    return s.ValidateToken(tokenString)
}
```

### Logout 实现

```go
func (h *AuthHandler) Logout(c *gin.Context) {
    // 获取 Token
    authHeader := c.GetHeader("Authorization")
    token := strings.TrimPrefix(authHeader, "Bearer ")

    // 获取 Token 剩余有效期
    claims := middleware.GetClaims(c)
    expiration := time.Until(claims.ExpiresAt.Time)

    // 加入黑名单
    if err := h.blacklist.Add(c.Request.Context(), token, expiration); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
```

### RS256 签名

```go
import (
    "crypto/rsa"
    "os"

    "github.com/golang-jwt/jwt/v5"
)

type JWTServiceRS256 struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
}

func NewJWTServiceRS256(privateKeyPath, publicKeyPath string) (*JWTServiceRS256, error) {
    // 读取私钥
    privateKeyData, err := os.ReadFile(privateKeyPath)
    if err != nil {
        return nil, err
    }
    privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
    if err != nil {
        return nil, err
    }

    // 读取公钥
    publicKeyData, err := os.ReadFile(publicKeyPath)
    if err != nil {
        return nil, err
    }
    publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
    if err != nil {
        return nil, err
    }

    return &JWTServiceRS256{
        privateKey: privateKey,
        publicKey:  publicKey,
    }, nil
}

func (s *JWTServiceRS256) GenerateToken(claims Claims) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(s.privateKey)
}

func (s *JWTServiceRS256) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.publicKey, nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }

    return claims, nil
}
```

## 完整示例

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// ... (之前的代码)

func main() {
    r := gin.Default()

    jwtService := NewJWTService("your-secret-key")
    userService := services.NewUserService()
    authHandler := NewAuthHandler(userService, jwtService)

    // 公开路由
    r.POST("/auth/register", authHandler.Register)
    r.POST("/auth/login", authHandler.Login)
    r.POST("/auth/refresh", authHandler.Refresh)

    // 需要认证的路由
    auth := r.Group("/api")
    auth.Use(AuthMiddleware(jwtService))
    {
        auth.GET("/me", authHandler.Me)
        auth.POST("/auth/logout", authHandler.Logout)

        // 其他受保护的路由
        auth.GET("/users", listUsers)
        auth.GET("/users/:id", getUser)
    }

    r.Run(":8080")
}
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 使用短期 Access Token | 15-30 分钟 |
| 使用长期 Refresh Token | 7-30 天 |
| 存储在 HttpOnly Cookie | 防止 XSS |
| 实现 Token 刷新 | 无感续期 |
| 敏感操作重新认证 | 支付、修改密码 |
| 实现 Token 撤销 | 黑名单机制 |
| 使用强密钥 | 至少 256 位 |
| 验证 Token 签发者 | 防止跨应用使用 |

**下一节**：[密码安全](./03-password-hashing.md) - 学习密码哈希
