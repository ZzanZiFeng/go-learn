# Step 1: 项目初始化

## 目标

搭建项目基础结构，配置数据库连接。

## 1.1 项目结构

```bash
# 项目已在 Phase 2 创建，确认结构
cd projects/todo-api
ls -la

# 目录结构
# ├── cmd/api/
# ├── internal/
# │   ├── config/
# │   ├── handlers/
# │   ├── middleware/
# │   ├── models/
# │   ├── repositories/
# │   └── services/
# ├── configs/
# └── go.mod
```

## 1.2 配置管理

创建 `configs/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug  # debug, release, test

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: todo_api
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-super-secret-key-change-in-production
  expire_hours: 24
```

创建 `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int
	Mode string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr 返回 Redis 地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string
	ExpireHours int
}

// ExpireDuration 返回过期时间
func (c *JWTConfig) ExpireDuration() time.Duration {
	return time.Duration(c.ExpireHours) * time.Hour
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 环境变量覆盖
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}

// LoadDefault 加载默认配置
func LoadDefault() (*Config, error) {
	return Load("configs/config.yaml")
}
```

## 1.3 数据库连接

创建 `internal/database/database.go`:

```go
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/user/todo-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 配置 GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	return db, nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
```

## 1.4 Redis 连接

创建 `internal/cache/redis.go`:

```go
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/user/todo-api/internal/config"
)

// Client Redis 客户端
var Client *redis.Client

// Init 初始化 Redis 连接
func Init(cfg *config.RedisConfig) (*redis.Client, error) {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return Client, nil
}

// Close 关闭 Redis 连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
```

## 1.5 主程序入口

创建 `cmd/api/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/todo-api/internal/cache"
	"github.com/user/todo-api/internal/config"
	"github.com/user/todo-api/internal/database"
	"github.com/user/todo-api/internal/handlers"
	"github.com/user/todo-api/internal/middleware"
	"github.com/user/todo-api/internal/models"
	"github.com/user/todo-api/internal/repositories"
	"github.com/user/todo-api/internal/services"
)

func main() {
	// 加载配置
	cfg, err := config.LoadDefault()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db, err := database.Init(&cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 自动迁移
	if err := db.AutoMigrate(&models.User{}, &models.Todo{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化 Redis
	_, err = cache.Init(&cfg.Redis)
	if err != nil {
		log.Printf("警告: Redis 连接失败: %v", err)
		// Redis 连接失败不阻止启动
	}
	defer cache.Close()

	// 初始化依赖
	userRepo := repositories.NewUserRepository(db)
	todoRepo := repositories.NewTodoRepository(db)

	authService := services.NewAuthService(userRepo, &cfg.JWT)
	todoService := services.NewTodoService(todoRepo)

	authHandler := handlers.NewAuthHandler(authService)
	todoHandler := handlers.NewTodoHandler(todoService)

	// 设置 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 路由
	setupRoutes(r, authHandler, todoHandler, authService)

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// 优雅关闭
	go func() {
		log.Printf("服务器启动在 http://localhost:%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("服务器已关闭")
}

func setupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, todoHandler *handlers.TodoHandler, authService *services.AuthService) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API 路由
	api := r.Group("/api")
	{
		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 待办路由（需要认证）
		todos := api.Group("/todos")
		todos.Use(middleware.Auth(authService))
		{
			todos.GET("", todoHandler.List)
			todos.POST("", todoHandler.Create)
			todos.GET("/:id", todoHandler.Get)
			todos.PUT("/:id", todoHandler.Update)
			todos.DELETE("/:id", todoHandler.Delete)
			todos.PATCH("/:id/complete", todoHandler.Complete)
		}
	}
}
```

## 1.6 go.mod 依赖

更新 `go.mod`:

```go
module github.com/user/todo-api

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/spf13/viper v1.18.1
	golang.org/x/crypto v0.17.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)
```

安装依赖:

```bash
go mod tidy
```

## 1.7 验证配置

```bash
# 启动基础设施
cd infra
docker-compose up -d postgres redis

# 测试运行
cd projects/todo-api
go run cmd/api/main.go

# 测试健康检查
curl http://localhost:8080/health
```

## 检查点

完成本步骤后：

- [x] 项目结构已搭建
- [x] 配置管理已实现
- [x] 数据库连接已配置
- [x] Redis 连接已配置
- [x] 基础路由已设置

## 下一步

[Step 2: 路由与处理器](./step-02-routes.md) - 实现完整的 CRUD API。
