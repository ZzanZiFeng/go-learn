# 配置管理

## 学习目标

掌握 Go 应用的配置管理策略，包括环境变量、配置文件和密钥管理。

## 1. 环境变量

### 1.1 读取环境变量

```go
package main

import (
    "fmt"
    "os"
    "strconv"
)

func main() {
    // 基本读取
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost" // 默认值
    }

    // 使用 LookupEnv 区分空值和未设置
    port, exists := os.LookupEnv("PORT")
    if !exists {
        port = "8080"
    }

    // 转换类型
    debugStr := os.Getenv("DEBUG")
    debug, _ := strconv.ParseBool(debugStr)

    fmt.Printf("DB_HOST: %s\n", dbHost)
    fmt.Printf("PORT: %s\n", port)
    fmt.Printf("DEBUG: %v\n", debug)
}
```

### 1.2 封装配置读取

```go
// internal/config/env.go
package config

import (
    "os"
    "strconv"
    "time"
)

// GetEnv 获取环境变量，支持默认值
func GetEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// GetEnvInt 获取整数环境变量
func GetEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

// GetEnvBool 获取布尔环境变量
func GetEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolVal, err := strconv.ParseBool(value); err == nil {
            return boolVal
        }
    }
    return defaultValue
}

// GetEnvDuration 获取时间间隔环境变量
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

### 1.3 使用示例

```go
// internal/config/config.go
package config

import "time"

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Host         string
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
}

type JWTConfig struct {
    Secret     string
    ExpireTime time.Duration
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() *Config {
    return &Config{
        Server: ServerConfig{
            Host:         GetEnv("SERVER_HOST", "0.0.0.0"),
            Port:         GetEnvInt("SERVER_PORT", 8080),
            ReadTimeout:  GetEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
            WriteTimeout: GetEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
        },
        Database: DatabaseConfig{
            Host:     GetEnv("DB_HOST", "localhost"),
            Port:     GetEnvInt("DB_PORT", 5432),
            User:     GetEnv("DB_USER", "postgres"),
            Password: GetEnv("DB_PASSWORD", ""),
            DBName:   GetEnv("DB_NAME", "myapp"),
            SSLMode:  GetEnv("DB_SSLMODE", "disable"),
        },
        Redis: RedisConfig{
            Host:     GetEnv("REDIS_HOST", "localhost"),
            Port:     GetEnvInt("REDIS_PORT", 6379),
            Password: GetEnv("REDIS_PASSWORD", ""),
            DB:       GetEnvInt("REDIS_DB", 0),
        },
        JWT: JWTConfig{
            Secret:     GetEnv("JWT_SECRET", "change-me-in-production"),
            ExpireTime: GetEnvDuration("JWT_EXPIRE", 24*time.Hour),
        },
    }
}
```

## 2. 配置文件

### 2.1 使用 Viper

```bash
go get github.com/spf13/viper
```

```go
// internal/config/viper.go
package config

import (
    "fmt"
    "strings"

    "github.com/spf13/viper"
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
    v := viper.New()

    // 设置配置文件
    v.SetConfigFile(configPath)

    // 或者按名称查找
    // v.SetConfigName("config")
    // v.SetConfigType("yaml")
    // v.AddConfigPath(".")
    // v.AddConfigPath("./configs")

    // 读取环境变量
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 读取配置文件
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }

    var config Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }

    return &config, nil
}
```

### 2.2 YAML 配置文件

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "10s"
  write_timeout: "10s"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "${DB_PASSWORD}"  # 可以引用环境变量
  dbname: "myapp"
  sslmode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "${JWT_SECRET}"
  expire_time: "24h"

log:
  level: "info"
  format: "json"
```

### 2.3 多环境配置

```yaml
# configs/config.development.yaml
server:
  port: 8080

database:
  host: "localhost"
  sslmode: "disable"

log:
  level: "debug"
  format: "console"
```

```yaml
# configs/config.production.yaml
server:
  port: 80

database:
  host: "${DB_HOST}"
  sslmode: "require"

log:
  level: "info"
  format: "json"
```

```go
// 根据环境加载配置
func LoadConfigByEnv() (*Config, error) {
    env := GetEnv("APP_ENV", "development")
    configPath := fmt.Sprintf("configs/config.%s.yaml", env)
    return LoadConfig(configPath)
}
```

## 3. 配置验证

### 3.1 使用 validator

```bash
go get github.com/go-playground/validator/v10
```

```go
// internal/config/validate.go
package config

import (
    "github.com/go-playground/validator/v10"
)

type Config struct {
    Server   ServerConfig   `validate:"required"`
    Database DatabaseConfig `validate:"required"`
}

type ServerConfig struct {
    Host string `validate:"required,ip|hostname"`
    Port int    `validate:"required,min=1,max=65535"`
}

type DatabaseConfig struct {
    Host     string `validate:"required"`
    Port     int    `validate:"required,min=1,max=65535"`
    User     string `validate:"required"`
    Password string `validate:"required"`
    DBName   string `validate:"required"`
}

// Validate 验证配置
func (c *Config) Validate() error {
    validate := validator.New()
    return validate.Struct(c)
}
```

### 3.2 自定义验证

```go
func (c *Config) ValidateCustom() error {
    if c.Server.Port == 0 {
        return fmt.Errorf("server port is required")
    }

    if c.Database.Password == "" && c.Database.SSLMode == "disable" {
        return fmt.Errorf("database password required when SSL is disabled")
    }

    if c.JWT.Secret == "change-me-in-production" {
        fmt.Println("WARNING: Using default JWT secret!")
    }

    return nil
}
```

## 4. 密钥管理

### 4.1 .env 文件（开发环境）

```bash
# .env - 不要提交到版本控制
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-secret-password
DB_NAME=myapp

JWT_SECRET=your-jwt-secret-key

REDIS_HOST=localhost
REDIS_PORT=6379
```

```go
// 使用 godotenv 加载 .env
import "github.com/joho/godotenv"

func init() {
    // 只在开发环境加载 .env
    if os.Getenv("APP_ENV") != "production" {
        godotenv.Load()
    }
}
```

### 4.2 Docker Secrets

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    image: myapp:latest
    secrets:
      - db_password
      - jwt_secret
    environment:
      - DB_PASSWORD_FILE=/run/secrets/db_password
      - JWT_SECRET_FILE=/run/secrets/jwt_secret

secrets:
  db_password:
    file: ./secrets/db_password.txt
  jwt_secret:
    file: ./secrets/jwt_secret.txt
```

```go
// 读取 Docker Secret
func GetSecretFromFile(envKey string) string {
    // 先检查 _FILE 后缀的环境变量
    fileEnv := envKey + "_FILE"
    if filePath := os.Getenv(fileEnv); filePath != "" {
        if data, err := os.ReadFile(filePath); err == nil {
            return strings.TrimSpace(string(data))
        }
    }
    // 否则直接读取环境变量
    return os.Getenv(envKey)
}
```

### 4.3 Kubernetes Secrets

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  db-password: cGFzc3dvcmQxMjM=  # base64 encoded
  jwt-secret: and0LXNlY3JldC1rZXk=
```

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: app
          env:
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: db-password
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: jwt-secret
```

## 5. 配置最佳实践

### 5.1 12-Factor App 配置原则

| 原则 | 说明 | 实践 |
|------|------|------|
| 环境分离 | 配置与代码分离 | 使用环境变量 |
| 无状态 | 不依赖本地状态 | 配置外部化 |
| 可移植 | 配置决定环境 | 同一镜像多环境 |

### 5.2 配置优先级

```
1. 命令行参数 (最高)
2. 环境变量
3. 配置文件
4. 默认值 (最低)
```

```go
// 实现配置优先级
func LoadConfigWithPriority() *Config {
    config := &Config{}

    // 1. 加载默认值
    setDefaults(config)

    // 2. 从配置文件加载
    if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
        loadFromFile(config, configFile)
    }

    // 3. 从环境变量覆盖
    loadFromEnv(config)

    // 4. 从命令行参数覆盖
    loadFromFlags(config)

    return config
}
```

### 5.3 敏感信息处理

```go
// 日志中隐藏敏感信息
func (c *Config) String() string {
    return fmt.Sprintf(
        "Server: %s:%d, DB: %s@%s:%d/%s, JWT: [HIDDEN]",
        c.Server.Host, c.Server.Port,
        c.Database.User, c.Database.Host, c.Database.Port, c.Database.DBName,
    )
}

// 永远不要在日志中打印
// log.Printf("Config: %+v", config)  // 危险！
log.Printf("Config: %s", config.String())  // 安全
```

## 练习

1. 创建一个支持多环境的配置系统
2. 实现配置热重载
3. 添加配置验证
4. 集成 Docker Secrets

## 下一步

[CI/CD 流水线](../04-cicd/) - 自动化构建和部署。
