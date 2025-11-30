module github.com/your-username/go-learn

go 1.21

// 这是Go Learn教程的根模块
// 主要用于管理共享依赖和示例代码

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.17.0
	github.com/rabbitmq/amqp091-go v1.9.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/spf13/viper v1.17.0
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.26.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)
