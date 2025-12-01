# 结构化日志 (Structured Logging)

## 概述

结构化日志使用键值对或 JSON 格式记录日志，便于搜索、过滤和分析。

```
传统日志:
2024-01-01 10:00:00 INFO User 123 logged in from 192.168.1.1

结构化日志 (JSON):
{"level":"info","ts":"2024-01-01T10:00:00Z","msg":"User logged in","user_id":123,"ip":"192.168.1.1"}
```

## Zap 日志库

Zap 是 Uber 开发的高性能日志库，特点是零分配和极快的速度。

### 安装

```bash
go get go.uber.org/zap
```

### 基本使用

```go
package main

import (
    "go.uber.org/zap"
)

func main() {
    // 1. 预设配置 - 生产环境
    logger, _ := zap.NewProduction()
    defer logger.Sync() // 刷新缓冲区

    logger.Info("Hello Zap",
        zap.String("key", "value"),
        zap.Int("count", 42),
    )
    // Output: {"level":"info","ts":1704096000,"caller":"main.go:12","msg":"Hello Zap","key":"value","count":42}

    // 2. 预设配置 - 开发环境
    devLogger, _ := zap.NewDevelopment()
    devLogger.Info("Development mode",
        zap.String("env", "dev"),
    )
    // Output: 2024-01-01T10:00:00.000+0800	INFO	main.go:18	Development mode	{"env": "dev"}

    // 3. 示例配置
    exampleLogger := zap.NewExample()
    exampleLogger.Info("Example logger")
    // Output: {"level":"info","msg":"Example logger"}
}
```

### 日志级别

```go
package main

import (
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()

    // 从低到高
    logger.Debug("Debug message")   // 调试信息
    logger.Info("Info message")     // 一般信息
    logger.Warn("Warning message")  // 警告
    logger.Error("Error message")   // 错误
    // logger.DPanic("DPanic message") // 开发环境 panic，生产环境 error
    // logger.Panic("Panic message")   // 记录后 panic
    // logger.Fatal("Fatal message")   // 记录后 os.Exit(1)
}
```

### 字段类型

```go
package main

import (
    "time"
    "errors"

    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 基本类型
    logger.Info("Basic types",
        zap.String("string", "hello"),
        zap.Int("int", 42),
        zap.Int64("int64", 123456789),
        zap.Float64("float64", 3.14),
        zap.Bool("bool", true),
    )

    // 复杂类型
    logger.Info("Complex types",
        zap.Time("time", time.Now()),
        zap.Duration("duration", 5*time.Second),
        zap.Error(errors.New("something went wrong")),
        zap.Strings("strings", []string{"a", "b", "c"}),
        zap.Ints("ints", []int{1, 2, 3}),
    )

    // 任意类型
    logger.Info("Any type",
        zap.Any("map", map[string]int{"a": 1, "b": 2}),
        zap.Any("struct", struct{ Name string }{"John"}),
    )

    // 反射（性能较低）
    logger.Info("Reflect",
        zap.Reflect("data", map[string]interface{}{"key": "value"}),
    )

    // 命名空间
    logger.Info("Namespaced",
        zap.Namespace("user"),
        zap.String("id", "123"),
        zap.String("name", "John"),
    )
    // {"level":"info","msg":"Namespaced","user":{"id":"123","name":"John"}}
}
```

### Sugar Logger

Sugar Logger 提供更简洁的 API，但性能略低。

```go
package main

import (
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    sugar := logger.Sugar()

    // Printf 风格
    sugar.Infof("User %s logged in", "John")

    // 键值对风格
    sugar.Infow("User logged in",
        "user", "John",
        "role", "admin",
    )

    // 简单消息
    sugar.Info("Simple message")

    // 从 Sugar 转回 Logger
    logger = sugar.Desugar()
}
```

### 自定义配置

```go
package main

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // 1. 使用 Config 结构
    config := zap.Config{
        Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
        Development: false,
        Encoding:    "json", // 或 "console"
        EncoderConfig: zapcore.EncoderConfig{
            TimeKey:        "ts",
            LevelKey:       "level",
            NameKey:        "logger",
            CallerKey:      "caller",
            MessageKey:     "msg",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
        },
        OutputPaths:      []string{"stdout", "/var/log/app.log"},
        ErrorOutputPaths: []string{"stderr"},
        InitialFields: map[string]interface{}{
            "service": "my-app",
            "version": "1.0.0",
        },
    }

    logger, _ := config.Build()
    defer logger.Sync()

    logger.Info("Custom config logger")

    // 2. 手动构建 Core
    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(os.Stdout),
        zap.InfoLevel,
    )

    logger2 := zap.New(core, zap.AddCaller())
    defer logger2.Sync()

    logger2.Info("Manual core logger")
}
```

### 多输出

```go
package main

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // 创建多个输出
    consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
    jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

    // 文件输出
    file, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

    core := zapcore.NewTee(
        // 控制台输出 (Debug 及以上)
        zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel),
        // 文件输出 (Info 及以上，JSON 格式)
        zapcore.NewCore(jsonEncoder, zapcore.AddSync(file), zap.InfoLevel),
    )

    logger := zap.New(core, zap.AddCaller())
    defer logger.Sync()

    logger.Debug("Only console") // 只输出到控制台
    logger.Info("Both outputs")  // 输出到控制台和文件
}
```

## 日志封装

### 全局日志

```go
package logger

import (
    "os"
    "sync"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var (
    Log   *zap.Logger
    Sugar *zap.SugaredLogger
    once  sync.Once
)

// Config 日志配置
type Config struct {
    Level      string // debug, info, warn, error
    Encoding   string // json, console
    OutputPath string // stdout, file path
    Debug      bool   // 开发模式
}

// Init 初始化日志
func Init(cfg Config) error {
    var err error
    once.Do(func() {
        err = initLogger(cfg)
    })
    return err
}

func initLogger(cfg Config) error {
    // 解析日志级别
    var level zapcore.Level
    if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
        level = zap.InfoLevel
    }

    // 编码器配置
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "ts",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }

    // 开发模式使用更友好的格式
    if cfg.Debug {
        encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }

    // 创建编码器
    var encoder zapcore.Encoder
    if cfg.Encoding == "console" || cfg.Debug {
        encoder = zapcore.NewConsoleEncoder(encoderConfig)
    } else {
        encoder = zapcore.NewJSONEncoder(encoderConfig)
    }

    // 创建输出
    var writer zapcore.WriteSyncer
    if cfg.OutputPath == "" || cfg.OutputPath == "stdout" {
        writer = zapcore.AddSync(os.Stdout)
    } else {
        file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            return err
        }
        writer = zapcore.AddSync(file)
    }

    // 创建 Core
    core := zapcore.NewCore(encoder, writer, level)

    // 创建 Logger
    options := []zap.Option{zap.AddCaller()}
    if cfg.Debug {
        options = append(options, zap.Development())
    }

    Log = zap.New(core, options...)
    Sugar = Log.Sugar()

    return nil
}

// Sync 刷新日志缓冲
func Sync() {
    if Log != nil {
        Log.Sync()
    }
}

// 便捷方法
func Debug(msg string, fields ...zap.Field) { Log.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { Log.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { Log.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { Log.Error(msg, fields...) }

// With 返回带有预设字段的 Logger
func With(fields ...zap.Field) *zap.Logger {
    return Log.With(fields...)
}
```

### 使用示例

```go
package main

import (
    "myapp/logger"
    "go.uber.org/zap"
)

func main() {
    // 初始化
    logger.Init(logger.Config{
        Level:    "debug",
        Encoding: "console",
        Debug:    true,
    })
    defer logger.Sync()

    // 使用全局日志
    logger.Info("Application started",
        zap.String("version", "1.0.0"),
    )

    // 带上下文的日志
    userLogger := logger.With(
        zap.String("user_id", "123"),
        zap.String("session_id", "abc"),
    )
    userLogger.Info("User action")
}
```

## Gin 集成

```go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Logger Gin 日志中间件
func Logger(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        // 处理请求
        c.Next()

        // 请求完成后记录日志
        end := time.Now()
        latency := end.Sub(start)

        fields := []zap.Field{
            zap.Int("status", c.Writer.Status()),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.String("ip", c.ClientIP()),
            zap.String("user-agent", c.Request.UserAgent()),
            zap.Duration("latency", latency),
        }

        // 添加请求 ID
        if requestID := c.GetString("request_id"); requestID != "" {
            fields = append(fields, zap.String("request_id", requestID))
        }

        // 添加错误信息
        if len(c.Errors) > 0 {
            fields = append(fields, zap.String("errors", c.Errors.String()))
        }

        // 根据状态码选择日志级别
        switch {
        case c.Writer.Status() >= 500:
            logger.Error("Server error", fields...)
        case c.Writer.Status() >= 400:
            logger.Warn("Client error", fields...)
        default:
            logger.Info("Request", fields...)
        }
    }
}

// RequestID 请求 ID 中间件
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// 使用示例
func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    r := gin.New()
    r.Use(RequestID())
    r.Use(Logger(logger))
    r.Use(gin.Recovery())

    r.GET("/api/hello", func(c *gin.Context) {
        // 获取带请求上下文的日志
        reqLogger := logger.With(
            zap.String("request_id", c.GetString("request_id")),
        )
        reqLogger.Info("Processing request")

        c.JSON(200, gin.H{"message": "hello"})
    })

    r.Run(":8080")
}
```

## 与 Node.js 对比

### Node.js (Winston)

```javascript
import winston from 'winston';

const logger = winston.createLogger({
    level: 'info',
    format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.json()
    ),
    transports: [
        new winston.transports.Console(),
        new winston.transports.File({ filename: 'app.log' }),
    ],
});

logger.info('User logged in', { userId: 123, ip: '192.168.1.1' });
```

### Go (Zap)

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("User logged in",
    zap.Int("userId", 123),
    zap.String("ip", "192.168.1.1"),
)
```

## 性能对比

| 库 | 禁用日志 | 简单日志 | 10 字段 | 分配 |
|------|---------|---------|---------|------|
| Zap | 0.5ns | 120ns | 350ns | 0 |
| Zerolog | 0.3ns | 85ns | 300ns | 0 |
| Logrus | 1000ns | 2000ns | 5000ns | 23 |
| log (标准库) | - | 800ns | 2500ns | 12 |

Zap 和 Zerolog 通过零分配实现极高性能。

**下一节**：[日志级别](./03-log-levels.md) - 学习日志级别和上下文
