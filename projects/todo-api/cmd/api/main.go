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
	"github.com/your-username/go-learn/projects/todo-api/internal/cache"
	"github.com/your-username/go-learn/projects/todo-api/internal/config"
	"github.com/your-username/go-learn/projects/todo-api/internal/database"
	"github.com/your-username/go-learn/projects/todo-api/internal/handlers"
	"github.com/your-username/go-learn/projects/todo-api/internal/middleware"
	"github.com/your-username/go-learn/projects/todo-api/internal/models"
	"github.com/your-username/go-learn/projects/todo-api/internal/repositories"
	"github.com/your-username/go-learn/projects/todo-api/internal/services"
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
	redisClient, err := cache.Init(&cfg.Redis)
	if err != nil {
		log.Printf("警告: Redis 连接失败: %v", err)
		// Redis 连接失败不阻止启动
	}
	defer cache.Close()

	// 初始化依赖
	userRepo := repositories.NewUserRepository(db)
	todoRepo := repositories.NewTodoRepository(db)

	authService := services.NewAuthService(userRepo, &cfg.JWT)

	// 创建待办服务（带缓存或不带缓存）
	var todoService *services.TodoService
	if redisClient != nil {
		todoCache := cache.NewTodoCache(redisClient, 5*time.Minute)
		todoService = services.NewTodoServiceWithCache(todoRepo, todoCache)
	} else {
		todoService = services.NewTodoService(todoRepo)
	}

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
