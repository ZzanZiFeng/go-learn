// Package main 演示使用 Zap 进行结构化日志
package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ==================== 上下文日志 ====================

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
	return zap.NewNop()
}

// ==================== 敏感字段脱敏 ====================

var sensitiveFields = []string{"password", "token", "secret", "credit_card"}

// SensitiveField 创建脱敏字段
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

// ==================== 日志初始化 ====================

// Config 日志配置
type Config struct {
	Level      string
	Encoding   string // json 或 console
	OutputPath string
	Debug      bool
}

// NewLogger 创建 logger
func NewLogger(cfg Config) (*zap.Logger, error) {
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

	// 开发模式使用彩色输出
	if cfg.Debug {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if cfg.Encoding == "console" || cfg.Debug {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 选择输出
	var writer zapcore.WriteSyncer
	if cfg.OutputPath == "" || cfg.OutputPath == "stdout" {
		writer = zapcore.AddSync(os.Stdout)
	} else {
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writer = zapcore.AddSync(file)
	}

	// 创建 core
	core := zapcore.NewCore(encoder, writer, level)

	// 创建 logger
	options := []zap.Option{zap.AddCaller()}
	if cfg.Debug {
		options = append(options, zap.Development())
	}

	return zap.New(core, options...), nil
}

// ==================== Gin 中间件 ====================

// LoggerMiddleware 日志中间件
func LoggerMiddleware(baseLogger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 生成请求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 创建带上下文的 logger
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

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		// 处理请求
		c.Next()

		// 记录请求日志
		duration := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.Duration("latency", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("Server error", fields...)
		case status >= 400:
			logger.Warn("Client error", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}

// ==================== 业务处理 ====================

// UserService 用户服务
type UserService struct {
	logger *zap.Logger
}

func NewUserService(logger *zap.Logger) *UserService {
	return &UserService{logger: logger.Named("user-service")}
}

func (s *UserService) Login(ctx context.Context, username, password string) error {
	logger := FromContext(ctx)

	// 记录登录尝试（密码已脱敏）
	logger.Info("User login attempt",
		zap.String("username", username),
		SensitiveField("password", password),
	)

	// 模拟验证
	time.Sleep(50 * time.Millisecond)

	logger.Info("User login successful",
		zap.String("username", username),
	)

	return nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (map[string]string, error) {
	logger := FromContext(ctx)

	logger.Debug("Fetching user from database",
		zap.String("user_id", id),
	)

	// 模拟数据库查询
	time.Sleep(30 * time.Millisecond)

	user := map[string]string{
		"id":    id,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	logger.Info("User fetched successfully",
		zap.String("user_id", id),
	)

	return user, nil
}

// ==================== 主程序 ====================

func main() {
	// 初始化 logger
	logger, err := NewLogger(Config{
		Level:    "debug",
		Encoding: "console",
		Debug:    true,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// 创建服务
	userService := NewUserService(logger)

	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(LoggerMiddleware(logger))

	// 路由
	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := userService.Login(c.Request.Context(), req.Username, req.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "login successful",
			"token":   "jwt-token-here",
		})
	})

	r.GET("/api/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		user, err := userService.GetUser(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动服务器
	logger.Info("Server starting", zap.Int("port", 8080))
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

/*
运行示例:
  go run main.go

测试:
  # 登录
  curl -X POST http://localhost:8080/api/login \
    -H "Content-Type: application/json" \
    -d '{"username":"john","password":"secret123"}'

  # 获取用户
  curl http://localhost:8080/api/users/123

预期输出:
  2024-01-01T10:00:00.000+0800    INFO    main.go:180 Server starting  {"port": 8080}
  2024-01-01T10:00:01.000+0800    INFO    main.go:145 User login attempt        {"request_id": "abc123", "path": "/api/login", "method": "POST", "username": "john", "password": "se****23"}
  2024-01-01T10:00:01.050+0800    INFO    main.go:152 User login successful     {"request_id": "abc123", "path": "/api/login", "method": "POST", "username": "john"}
  2024-01-01T10:00:01.051+0800    INFO    main.go:108 Request completed {"request_id": "abc123", "path": "/api/login", "method": "POST", "status": 200, "latency": "51ms", "client_ip": "127.0.0.1"}
*/
