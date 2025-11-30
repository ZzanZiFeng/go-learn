# 中间件 (Middleware)

## 概述

中间件是在请求处理前后执行的函数，用于处理通用逻辑如日志、认证、CORS 等。

## 与 Express.js 对比

### Express.js

```javascript
// 中间件函数
const logger = (req, res, next) => {
  console.log(`${req.method} ${req.url}`);
  next();
};

app.use(logger);
app.use(express.json());
app.use(cors());

// 路由级中间件
app.get('/protected', authMiddleware, handler);
```

### Gin

```go
// 中间件函数
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        fmt.Printf("%s %s\n", c.Request.Method, c.Request.URL)
        c.Next()
    }
}

r := gin.New()
r.Use(Logger())
r.Use(gin.Recovery())

// 路由级中间件
r.GET("/protected", AuthMiddleware(), handler)
```

## 基础概念

### 中间件执行流程

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 请求前处理
        fmt.Println("Before handler")

        // 2. 调用下一个处理器
        c.Next()

        // 3. 响应后处理
        fmt.Println("After handler")
    }
}

// 执行顺序：
// Middleware1 Before -> Middleware2 Before -> Handler
// -> Middleware2 After -> Middleware1 After
```

### 中间件控制

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        if token == "" {
            // Abort: 停止后续中间件和处理器
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }

        // Next: 继续执行
        c.Next()
    }
}
```

## 内置中间件

### gin.Default()

```go
// gin.Default() 包含两个中间件
r := gin.Default()  // = gin.New() + Logger() + Recovery()

// 等价于
r := gin.New()
r.Use(gin.Logger())
r.Use(gin.Recovery())
```

### Logger 中间件

```go
// 默认日志格式
r.Use(gin.Logger())

// 自定义日志格式
r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    return fmt.Sprintf("[%s] %s %s %d %s\n",
        param.TimeStamp.Format(time.RFC3339),
        param.Method,
        param.Path,
        param.StatusCode,
        param.Latency,
    )
}))

// 写入文件
f, _ := os.Create("gin.log")
gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
r.Use(gin.Logger())

// 跳过某些路径
r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
}))
```

### Recovery 中间件

```go
// 默认 Recovery
r.Use(gin.Recovery())

// 自定义 Recovery
r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
    if err, ok := recovered.(string); ok {
        c.JSON(500, gin.H{"error": err})
    }
    c.AbortWithStatus(500)
}))
```

## 常用自定义中间件

### 日志中间件

```go
import (
    "log/slog"
    "time"
)

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        // 处理请求
        c.Next()

        // 记录日志
        latency := time.Since(start)
        status := c.Writer.Status()

        logger.Info("request",
            "method", c.Request.Method,
            "path", path,
            "query", query,
            "status", status,
            "latency", latency,
            "ip", c.ClientIP(),
            "user_agent", c.Request.UserAgent(),
            "errors", c.Errors.String(),
        )
    }
}
```

### 请求 ID 中间件

```go
import "github.com/google/uuid"

func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从请求头获取或生成新的
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // 设置到上下文和响应头
        c.Set("RequestID", requestID)
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}

// 使用
r.Use(RequestIDMiddleware())

r.GET("/test", func(c *gin.Context) {
    requestID := c.GetString("RequestID")
    c.JSON(200, gin.H{"request_id": requestID})
})
```

### CORS 中间件

```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        c.Header("Access-Control-Expose-Headers", "Content-Length")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

// 或使用 gin-contrib/cors
import "github.com/gin-contrib/cors"

r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

### 认证中间件

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{
                "code":    "UNAUTHORIZED",
                "message": "Missing authorization token",
            })
            return
        }

        // 移除 "Bearer " 前缀
        if strings.HasPrefix(token, "Bearer ") {
            token = token[7:]
        }

        // 验证 token
        claims, err := validateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{
                "code":    "INVALID_TOKEN",
                "message": "Invalid or expired token",
            })
            return
        }

        // 将用户信息存入上下文
        c.Set("userID", claims.UserID)
        c.Set("role", claims.Role)

        c.Next()
    }
}

// 使用
authorized := r.Group("/api")
authorized.Use(AuthMiddleware())
{
    authorized.GET("/profile", getProfile)
    authorized.PUT("/profile", updateProfile)
}
```

### 角色权限中间件

```go
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }

        roleStr := role.(string)
        allowed := false
        for _, r := range allowedRoles {
            if r == roleStr {
                allowed = true
                break
            }
        }

        if !allowed {
            c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
            return
        }

        c.Next()
    }
}

// 使用
admin := r.Group("/admin")
admin.Use(AuthMiddleware(), RoleMiddleware("admin"))
{
    admin.GET("/users", listAllUsers)
    admin.DELETE("/users/:id", deleteUser)
}
```

### 限流中间件

```go
import (
    "sync"
    "time"
)

type RateLimiter struct {
    visitors map[string]*Visitor
    mu       sync.RWMutex
    rate     int           // 每分钟请求数
    burst    int           // 突发请求数
}

type Visitor struct {
    tokens    int
    lastVisit time.Time
}

func NewRateLimiter(rate, burst int) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*Visitor),
        rate:     rate,
        burst:    burst,
    }
    go rl.cleanup()
    return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    v, exists := rl.visitors[ip]
    if !exists {
        rl.visitors[ip] = &Visitor{tokens: rl.burst - 1, lastVisit: time.Now()}
        return true
    }

    // 补充 tokens
    elapsed := time.Since(v.lastVisit)
    v.tokens += int(elapsed.Minutes()) * rl.rate
    if v.tokens > rl.burst {
        v.tokens = rl.burst
    }
    v.lastVisit = time.Now()

    if v.tokens > 0 {
        v.tokens--
        return true
    }

    return false
}

func (rl *RateLimiter) cleanup() {
    for {
        time.Sleep(time.Minute)
        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastVisit) > 3*time.Minute {
                delete(rl.visitors, ip)
            }
        }
        rl.mu.Unlock()
    }
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()

        if !limiter.Allow(ip) {
            c.AbortWithStatusJSON(429, gin.H{
                "code":    "RATE_LIMIT_EXCEEDED",
                "message": "Too many requests",
            })
            return
        }

        c.Next()
    }
}

// 使用
limiter := NewRateLimiter(60, 10)  // 每分钟 60 次，突发 10 次
r.Use(RateLimitMiddleware(limiter))
```

### 超时中间件

```go
import (
    "context"
    "time"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()

        c.Request = c.Request.WithContext(ctx)

        finished := make(chan struct{})
        go func() {
            c.Next()
            close(finished)
        }()

        select {
        case <-finished:
            // 正常完成
        case <-ctx.Done():
            c.AbortWithStatusJSON(504, gin.H{
                "code":    "TIMEOUT",
                "message": "Request timeout",
            })
        }
    }
}

// 使用
r.Use(TimeoutMiddleware(30 * time.Second))
```

### 响应压缩中间件

```go
import "github.com/gin-contrib/gzip"

// 使用 gzip 压缩
r.Use(gzip.Gzip(gzip.DefaultCompression))

// 排除某些路径
r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api/upload"})))
```

## 中间件作用域

### 全局中间件

```go
r := gin.New()

// 所有路由都使用
r.Use(LoggingMiddleware())
r.Use(RecoveryMiddleware())
r.Use(CORSMiddleware())
```

### 路由组中间件

```go
// 公开路由
public := r.Group("/api")
{
    public.POST("/login", login)
    public.POST("/register", register)
}

// 需要认证的路由
auth := r.Group("/api")
auth.Use(AuthMiddleware())
{
    auth.GET("/profile", getProfile)
    auth.PUT("/profile", updateProfile)
}

// 管理员路由
admin := r.Group("/api/admin")
admin.Use(AuthMiddleware(), RoleMiddleware("admin"))
{
    admin.GET("/users", listUsers)
    admin.DELETE("/users/:id", deleteUser)
}
```

### 单路由中间件

```go
r.GET("/public", publicHandler)
r.GET("/protected", AuthMiddleware(), protectedHandler)
r.DELETE("/admin/users/:id", AuthMiddleware(), RoleMiddleware("admin"), deleteUserHandler)
```

## 中间件传递数据

### 使用 Context

```go
// 设置数据
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 设置数据到上下文
        c.Set("userID", 123)
        c.Set("user", User{ID: 123, Name: "Alice"})

        c.Next()
    }
}

// 获取数据
func handler(c *gin.Context) {
    // 获取并类型断言
    userID, exists := c.Get("userID")
    if exists {
        id := userID.(int)
        fmt.Println(id)
    }

    // 或使用 MustGet（不存在会 panic）
    user := c.MustGet("user").(User)

    // 类型安全的获取方法
    userIDStr := c.GetString("userID")   // 返回 ""
    userIDInt := c.GetInt("userID")      // 返回 123
}
```

### 自定义 Context Key

```go
type contextKey string

const (
    UserIDKey  contextKey = "userID"
    UserKey    contextKey = "user"
    RequestIDKey contextKey = "requestID"
)

func SetUserID(c *gin.Context, userID int) {
    c.Set(string(UserIDKey), userID)
}

func GetUserID(c *gin.Context) (int, bool) {
    if v, exists := c.Get(string(UserIDKey)); exists {
        if userID, ok := v.(int); ok {
            return userID, true
        }
    }
    return 0, false
}
```

## 完整示例

```go
package main

import (
    "log/slog"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// 日志中间件
func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        logger.Info("request",
            "request_id", c.GetString("RequestID"),
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status", c.Writer.Status(),
            "latency", time.Since(start),
            "ip", c.ClientIP(),
        )
    }
}

// 请求 ID 中间件
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("RequestID", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }

        if strings.HasPrefix(token, "Bearer ") {
            token = token[7:]
        }

        // 简化的 token 验证
        if token != "valid-token" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }

        c.Set("userID", 1)
        c.Set("role", "user")
        c.Next()
    }
}

// 角色中间件
func RoleMiddleware(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role := c.GetString("role")
        for _, r := range roles {
            if r == role {
                c.Next()
                return
            }
        }
        c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
    }
}

// Recovery 中间件
func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("panic recovered",
                    "request_id", c.GetString("RequestID"),
                    "error", err,
                )
                c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
            }
        }()
        c.Next()
    }
}

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    r := gin.New()

    // 全局中间件
    r.Use(RequestIDMiddleware())
    r.Use(LoggingMiddleware(logger))
    r.Use(RecoveryMiddleware(logger))
    r.Use(CORSMiddleware())

    // 公开路由
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.POST("/login", func(c *gin.Context) {
        c.JSON(200, gin.H{"token": "valid-token"})
    })

    // 认证路由
    auth := r.Group("/api")
    auth.Use(AuthMiddleware())
    {
        auth.GET("/profile", func(c *gin.Context) {
            userID := c.GetInt("userID")
            c.JSON(200, gin.H{"user_id": userID})
        })
    }

    // 管理员路由
    admin := r.Group("/api/admin")
    admin.Use(AuthMiddleware(), RoleMiddleware("admin"))
    {
        admin.GET("/users", func(c *gin.Context) {
            c.JSON(200, gin.H{"users": []string{"Alice", "Bob"}})
        })
    }

    r.Run(":8080")
}
```

## 总结

| 中间件类型 | 用途 |
|-----------|------|
| Logger | 请求日志 |
| Recovery | Panic 恢复 |
| CORS | 跨域支持 |
| Auth | 认证 |
| Role | 权限控制 |
| RateLimit | 限流 |
| Timeout | 超时控制 |
| RequestID | 请求追踪 |
| Gzip | 响应压缩 |

**下一节**：[错误处理](./08-error-handling.md) - 学习 API 错误处理
