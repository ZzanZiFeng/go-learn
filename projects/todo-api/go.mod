module github.com/your-username/go-learn/projects/todo-api

go 1.21

// Todo API 项目 - RESTful API服务
// 学习目标: Gin框架、GORM、JWT认证、Redis缓存

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/spf13/viper v1.17.0
	go.uber.org/zap v1.26.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)
