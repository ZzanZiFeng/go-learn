# 认证授权 (Authentication & Authorization)

## 学习目标

完成本章后，你将能够：
- 理解认证与授权的区别
- 实现 JWT 认证系统
- 安全地存储密码
- 实现基于角色的访问控制（RBAC）
- 理解 OAuth2 基本流程

## 与 JavaScript/Node.js 对比

| 功能 | Node.js | Go |
|------|---------|-----|
| JWT | jsonwebtoken | golang-jwt/jwt |
| 密码哈希 | bcrypt | golang.org/x/crypto/bcrypt |
| Session | express-session | gorilla/sessions |
| OAuth2 | passport | golang.org/x/oauth2 |

## 核心概念

### 认证 (Authentication) vs 授权 (Authorization)

```
认证 (Authentication): "你是谁？"
- 验证用户身份
- 登录/注册流程
- JWT/Session Token

授权 (Authorization): "你能做什么？"
- 验证用户权限
- 角色/权限检查
- RBAC/ABAC
```

### 常见认证方式

| 方式 | 优点 | 缺点 | 适用场景 |
|------|------|------|---------|
| JWT | 无状态，可扩展 | Token 无法撤销 | API，微服务 |
| Session | 可撤销，安全 | 需要存储 | 传统 Web |
| API Key | 简单 | 安全性较低 | 服务间调用 |
| OAuth2 | 标准化，第三方登录 | 复杂 | 第三方登录 |

## 章节内容

1. [认证概念](./01-auth-concepts.md) - 认证授权基础概念
2. [JWT 认证](./02-jwt-auth.md) - JWT 实现和最佳实践
3. [密码安全](./03-password-hashing.md) - 密码哈希和验证
4. [会话管理](./04-session-management.md) - Session 管理
5. [OAuth2](./05-oauth2.md) - OAuth2 和第三方登录
6. [RBAC](./06-rbac.md) - 基于角色的访问控制
7. [练习](./exercises.md) - 实践练习

## 快速开始

### JWT 认证示例

```go
package main

import (
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key")

type Claims struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

// 生成 JWT Token
func GenerateToken(userID uint, role string) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "my-app",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// 验证 JWT Token
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, jwt.ErrSignatureInvalid
}

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := ValidateToken(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        c.Set("userID", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}

func main() {
    r := gin.Default()

    // 公开路由
    r.POST("/login", func(c *gin.Context) {
        // 验证用户凭据...
        token, _ := GenerateToken(1, "user")
        c.JSON(http.StatusOK, gin.H{"token": token})
    })

    // 需要认证的路由
    auth := r.Group("/api")
    auth.Use(AuthMiddleware())
    {
        auth.GET("/profile", func(c *gin.Context) {
            userID := c.GetUint("userID")
            c.JSON(http.StatusOK, gin.H{"user_id": userID})
        })
    }

    r.Run(":8080")
}
```

### 密码哈希示例

```go
import "golang.org/x/crypto/bcrypt"

// 哈希密码
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

## 安全最佳实践

### JWT 安全

```go
// ✅ 好的做法
- 使用强密钥（至少 256 位）
- 设置合理的过期时间
- 使用 HTTPS 传输
- 敏感操作使用短期 token
- 实现 token 刷新机制

// ❌ 避免
- 在 token 中存储敏感信息
- 永不过期的 token
- 在 URL 中传递 token
- 客户端存储在 localStorage（XSS 风险）
```

### 密码安全

```go
// ✅ 好的做法
- 使用 bcrypt/argon2 哈希
- 强制密码复杂度
- 限制登录尝试次数
- 实现账户锁定机制

// ❌ 避免
- 明文存储密码
- 使用 MD5/SHA1 哈希密码
- 自定义加密算法
```

## 项目结构

```
internal/
├── auth/
│   ├── jwt.go           # JWT 相关
│   ├── password.go      # 密码处理
│   └── middleware.go    # 认证中间件
├── handlers/
│   └── auth.go          # 认证处理器
├── services/
│   └── auth.go          # 认证服务
└── models/
    └── user.go          # 用户模型
```

## 实践项目

本章的知识将在以下项目中应用：
- [Auth Service](../../02-practice/auth-service/) - 完整的认证服务

## 下一步

完成本章后，继续学习 [数据库操作](../07-database/)，学习如何持久化用户数据。
