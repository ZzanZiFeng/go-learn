module github.com/your-username/go-learn/projects/auth-service

go 1.21

// Auth Service 项目 - 认证授权服务
// 学习目标: JWT认证、Session管理、OAuth2、RBAC

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/spf13/viper v1.17.0
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.16.0
	golang.org/x/oauth2 v0.15.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)
