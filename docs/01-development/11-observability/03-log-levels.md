# 日志级别 (Log Levels)

## 概述

日志级别用于区分日志的重要性，便于过滤和分析。

```
┌──────────────────────────────────────────────────────────────────┐
│                         日志级别                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  DEBUG  ──▶  INFO  ──▶  WARN  ──▶  ERROR  ──▶  FATAL            │
│   调试       信息       警告       错误       致命               │
│                                                                  │
│  设置级别为 INFO 时，DEBUG 日志不会输出                           │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 日志级别定义

| 级别 | 说明 | 适用场景 |
|------|------|----------|
| DEBUG | 详细调试信息 | 开发调试、问题排查 |
| INFO | 一般信息 | 业务事件、状态变更 |
| WARN | 警告信息 | 潜在问题、非预期情况 |
| ERROR | 错误信息 | 需要关注的错误 |
| FATAL | 致命错误 | 程序无法继续运行 |

## Zap 日志级别

```go
package main

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // 设置日志级别
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

    logger, _ := config.Build()
    defer logger.Sync()

    // 这些级别的日志都会输出
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")

    // DEBUG 级别日志不会输出（因为设置了 INFO 级别）
    logger.Debug("Debug message") // 不输出
}
```

### 动态调整级别

```go
package main

import (
    "net/http"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // 创建可动态调整的级别
    atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)

    config := zap.NewProductionConfig()
    config.Level = atomicLevel

    logger, _ := config.Build()
    defer logger.Sync()

    // HTTP 端点动态调整级别
    http.HandleFunc("/log/level", atomicLevel.ServeHTTP)

    // GET /log/level         - 获取当前级别
    // PUT /log/level -d "debug" - 设置为 debug

    go http.ListenAndServe(":8080", nil)

    // 代码中调整级别
    logger.Info("Current level: INFO")

    atomicLevel.SetLevel(zapcore.DebugLevel)
    logger.Debug("Now DEBUG is visible")

    atomicLevel.SetLevel(zapcore.ErrorLevel)
    logger.Info("This won't show")  // 不输出
    logger.Error("Only errors now")  // 输出
}
```

## 日志上下文

### 请求上下文

```go
package main

import (
    "context"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// contextKey 上下文键类型
type contextKey string

const loggerKey contextKey = "logger"

// WithLogger 将 logger 添加到上下文
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

// FromContext 从上下文获取 logger
func FromContext(ctx context.Context) *zap.Logger {
    if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
        return logger
    }
    return zap.NewNop() // 返回空 logger
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(baseLogger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 为每个请求创建带上下文的 logger
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }

        logger := baseLogger.With(
            zap.String("request_id", requestID),
            zap.String("path", c.Request.URL.Path),
            zap.String("method", c.Request.Method),
        )

        // 添加到 Gin 上下文
        c.Set("logger", logger)

        // 添加到 Go 上下文
        ctx := WithLogger(c.Request.Context(), logger)
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}

// 在 Handler 中使用
func UserHandler(c *gin.Context) {
    logger := c.MustGet("logger").(*zap.Logger)

    logger.Info("Processing user request",
        zap.String("user_id", c.Param("id")),
    )

    // 或从上下文获取
    ctx := c.Request.Context()
    ctxLogger := FromContext(ctx)
    ctxLogger.Info("From context")
}

// 在 Service 中使用
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    logger := FromContext(ctx)

    logger.Debug("Fetching user from database",
        zap.String("user_id", id),
    )

    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        logger.Error("Failed to fetch user",
            zap.String("user_id", id),
            zap.Error(err),
        )
        return nil, err
    }

    logger.Info("User fetched successfully",
        zap.String("user_id", id),
    )

    return user, nil
}
```

### 追踪上下文

```go
package logger

import (
    "context"

    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
)

// WithTraceContext 添加追踪上下文到日志
func WithTraceContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
    span := trace.SpanFromContext(ctx)
    if !span.SpanContext().IsValid() {
        return logger
    }

    return logger.With(
        zap.String("trace_id", span.SpanContext().TraceID().String()),
        zap.String("span_id", span.SpanContext().SpanID().String()),
    )
}

// 使用示例
func ProcessOrder(ctx context.Context, orderID string) error {
    logger := WithTraceContext(ctx, baseLogger)

    logger.Info("Processing order",
        zap.String("order_id", orderID),
    )
    // Output: {"msg":"Processing order","order_id":"123","trace_id":"abc...","span_id":"def..."}

    return nil
}
```

## 日志字段规范

### 字段命名规范

```go
// 推荐：snake_case
logger.Info("User logged in",
    zap.String("user_id", "123"),
    zap.String("session_id", "abc"),
    zap.String("client_ip", "192.168.1.1"),
)

// 不推荐：各种风格混用
logger.Info("User logged in",
    zap.String("userId", "123"),     // camelCase
    zap.String("SessionID", "abc"),  // PascalCase
    zap.String("client-ip", "192.168.1.1"), // kebab-case
)
```

### 常用字段

```go
// 请求相关
zap.String("request_id", requestID)
zap.String("method", "GET")
zap.String("path", "/api/users")
zap.Int("status", 200)
zap.Duration("latency", latency)
zap.String("client_ip", ip)
zap.String("user_agent", userAgent)

// 用户相关
zap.String("user_id", userID)
zap.String("username", username)
zap.String("role", role)

// 业务相关
zap.String("order_id", orderID)
zap.Float64("amount", amount)
zap.String("action", "create")

// 错误相关
zap.Error(err)
zap.String("error_code", "E001")
zap.Stack("stacktrace")

// 追踪相关
zap.String("trace_id", traceID)
zap.String("span_id", spanID)
zap.String("parent_span_id", parentSpanID)
```

## 日志分类

### 按功能分类

```go
package logger

import "go.uber.org/zap"

var (
    // 访问日志
    Access *zap.Logger

    // 业务日志
    Business *zap.Logger

    // 错误日志
    Error *zap.Logger

    // 审计日志
    Audit *zap.Logger
)

func Init() {
    baseLogger, _ := zap.NewProduction()

    Access = baseLogger.Named("access")
    Business = baseLogger.Named("business")
    Error = baseLogger.Named("error")
    Audit = baseLogger.Named("audit")
}

// 使用
func main() {
    logger.Init()

    // 访问日志
    logger.Access.Info("Request received",
        zap.String("path", "/api/users"),
    )
    // {"logger":"access","msg":"Request received",...}

    // 业务日志
    logger.Business.Info("Order created",
        zap.String("order_id", "123"),
    )

    // 审计日志
    logger.Audit.Info("User permission changed",
        zap.String("user_id", "123"),
        zap.String("old_role", "user"),
        zap.String("new_role", "admin"),
    )
}
```

### 不同输出

```go
package logger

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func InitWithSeparateOutputs() (*zap.Logger, *zap.Logger, *zap.Logger) {
    encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

    // 访问日志 -> access.log
    accessFile, _ := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    accessCore := zapcore.NewCore(encoder, zapcore.AddSync(accessFile), zap.InfoLevel)
    accessLogger := zap.New(accessCore).Named("access")

    // 业务日志 -> app.log
    appFile, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    appCore := zapcore.NewCore(encoder, zapcore.AddSync(appFile), zap.InfoLevel)
    appLogger := zap.New(appCore).Named("app")

    // 错误日志 -> error.log (只记录 Error 及以上)
    errorFile, _ := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    errorCore := zapcore.NewCore(encoder, zapcore.AddSync(errorFile), zap.ErrorLevel)
    errorLogger := zap.New(errorCore).Named("error")

    return accessLogger, appLogger, errorLogger
}
```

## 敏感信息处理

```go
package logger

import (
    "strings"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// 敏感字段列表
var sensitiveFields = []string{"password", "token", "secret", "credit_card"}

// SensitiveField 敏感字段（自动脱敏）
func SensitiveField(key string, value string) zap.Field {
    return zap.String(key, maskSensitive(key, value))
}

func maskSensitive(key, value string) string {
    keyLower := strings.ToLower(key)
    for _, sensitive := range sensitiveFields {
        if strings.Contains(keyLower, sensitive) {
            if len(value) <= 4 {
                return "****"
            }
            return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
        }
    }
    return value
}

// 使用
func main() {
    logger, _ := zap.NewProduction()

    // 自动脱敏
    logger.Info("User login",
        SensitiveField("password", "mysecretpass"),
        SensitiveField("token", "abc123xyz789"),
    )
    // {"msg":"User login","password":"my******ss","token":"ab******89"}
}
```

## 日志轮转

使用 `lumberjack` 实现日志轮转：

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

func InitWithRotation() *zap.Logger {
    // 日志轮转配置
    writer := &lumberjack.Logger{
        Filename:   "/var/log/app/app.log",
        MaxSize:    100, // MB
        MaxBackups: 3,   // 保留旧文件数
        MaxAge:     28,  // 保留天数
        Compress:   true,
    }

    encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
    core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zap.InfoLevel)

    return zap.New(core, zap.AddCaller())
}
```

安装：

```bash
go get gopkg.in/natefinch/lumberjack.v2
```

## 与 Node.js 对比

### Node.js (Winston)

```javascript
import winston from 'winston';

const logger = winston.createLogger({
    level: process.env.LOG_LEVEL || 'info',
    format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.json()
    ),
    transports: [
        new winston.transports.Console(),
        new winston.transports.File({ filename: 'error.log', level: 'error' }),
        new winston.transports.File({ filename: 'combined.log' }),
    ],
});

// 子 logger
const userLogger = logger.child({ service: 'user-service' });
userLogger.info('User created', { userId: 123 });
```

### Go (Zap)

```go
config := zap.NewProductionConfig()
config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

logger, _ := config.Build()
defer logger.Sync()

// 子 logger
userLogger := logger.With(zap.String("service", "user-service"))
userLogger.Info("User created", zap.Int("user_id", 123))
```

**下一节**：[Prometheus](./04-prometheus.md) - 学习指标收集
