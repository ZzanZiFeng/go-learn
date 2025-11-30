module github.com/your-username/go-learn/projects/fullstack-demo/backend

go 1.21

// Fullstack Demo 后端 - 全栈应用示例
// 学习目标: 完整后端服务、微服务架构、生产级部署

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/prometheus/client_golang v1.17.0
	github.com/rabbitmq/amqp091-go v1.9.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/spf13/viper v1.17.0
	go.opentelemetry.io/otel v1.21.0
	go.uber.org/zap v1.26.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)
