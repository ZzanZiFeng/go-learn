# 配置管理 (Configuration Management)

## 概述

Go 应用通常需要管理多种配置：数据库连接、服务端口、第三方 API 密钥等。Viper 是最流行的 Go 配置库。

## 与 JavaScript/Node.js 对比

| 工具 | Node.js | Go |
|------|---------|-----|
| 环境变量 | dotenv | os.Getenv / Viper |
| 配置文件 | config, rc | Viper |
| 类型安全 | ✗（运行时） | ✓（结构体绑定） |

### Node.js 配置

```javascript
// .env
DB_HOST=localhost
DB_PORT=5432

// app.js
require('dotenv').config();
const dbHost = process.env.DB_HOST;  // string，无类型检查
const dbPort = parseInt(process.env.DB_PORT);  // 需要手动转换
```

### Go 配置

```go
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`  // 自动类型转换
}
```

## Viper 基础

### 安装

```bash
go get github.com/spf13/viper
```

### 基本用法

```go
package main

import (
    "fmt"
    "log"

    "github.com/spf13/viper"
)

func main() {
    // 设置配置文件名（不含扩展名）
    viper.SetConfigName("config")

    // 设置配置文件类型
    viper.SetConfigType("yaml")

    // 添加配置文件搜索路径
    viper.AddConfigPath(".")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath("$HOME/.myapp")

    // 读取配置文件
    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Error reading config: %v", err)
    }

    // 读取配置值
    fmt.Println("Server Port:", viper.GetString("server.port"))
    fmt.Println("DB Host:", viper.GetString("database.host"))
    fmt.Println("Debug Mode:", viper.GetBool("debug"))
}
```

### 配置文件示例

```yaml
# configs/config.yaml
server:
  port: ":8080"
  read_timeout: 10s
  write_timeout: 10s
  mode: debug  # debug, release, test

database:
  host: localhost
  port: 5432
  name: myapp
  user: postgres
  password: secret
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-secret-key
  expiration: 24h

log:
  level: info      # debug, info, warn, error
  format: json     # json, text
  output: stdout   # stdout, file

debug: true
```

## 结构体绑定

### 定义配置结构体

```go
// internal/config/config.go
package config

import "time"

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Log      LogConfig      `mapstructure:"log"`
    Debug    bool           `mapstructure:"debug"`
}

type ServerConfig struct {
    Port         string        `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
    Mode         string        `mapstructure:"mode"`
}

type DatabaseConfig struct {
    Host            string        `mapstructure:"host"`
    Port            int           `mapstructure:"port"`
    Name            string        `mapstructure:"name"`
    User            string        `mapstructure:"user"`
    Password        string        `mapstructure:"password"`
    MaxOpenConns    int           `mapstructure:"max_open_conns"`
    MaxIdleConns    int           `mapstructure:"max_idle_conns"`
    ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
    Secret     string        `mapstructure:"secret"`
    Expiration time.Duration `mapstructure:"expiration"`
}

type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
    Output string `mapstructure:"output"`
}
```

### 加载配置

```go
// internal/config/loader.go
package config

import (
    "fmt"
    "strings"

    "github.com/spf13/viper"
)

func Load() (*Config, error) {
    return LoadWithPath("configs")
}

func LoadWithPath(configPath string) (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(configPath)
    viper.AddConfigPath(".")

    // 设置默认值
    setDefaults()

    // 支持环境变量
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 读取配置文件
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    // 绑定到结构体
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    // 验证配置
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    return &cfg, nil
}

func setDefaults() {
    // Server
    viper.SetDefault("server.port", ":8080")
    viper.SetDefault("server.read_timeout", "10s")
    viper.SetDefault("server.write_timeout", "10s")
    viper.SetDefault("server.mode", "debug")

    // Database
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", 5432)
    viper.SetDefault("database.max_open_conns", 25)
    viper.SetDefault("database.max_idle_conns", 5)
    viper.SetDefault("database.conn_max_lifetime", "5m")

    // Redis
    viper.SetDefault("redis.host", "localhost")
    viper.SetDefault("redis.port", 6379)
    viper.SetDefault("redis.db", 0)

    // JWT
    viper.SetDefault("jwt.expiration", "24h")

    // Log
    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.format", "json")
    viper.SetDefault("log.output", "stdout")
}

// Validate 验证配置
func (c *Config) Validate() error {
    if c.Database.Host == "" {
        return fmt.Errorf("database host is required")
    }
    if c.JWT.Secret == "" {
        return fmt.Errorf("JWT secret is required")
    }
    return nil
}
```

## 环境变量覆盖

Viper 支持使用环境变量覆盖配置文件中的值：

```go
// 启用自动环境变量绑定
viper.AutomaticEnv()

// 将 . 替换为 _（因为环境变量不能包含点）
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
```

```bash
# 环境变量覆盖配置
export DATABASE_HOST=production-db.example.com
export DATABASE_PORT=5432
export SERVER_PORT=:9090
export JWT_SECRET=production-secret

./myapp
```

### 环境变量前缀

```go
// 添加前缀避免冲突
viper.SetEnvPrefix("MYAPP")
viper.AutomaticEnv()

// 现在需要使用 MYAPP_ 前缀
// MYAPP_DATABASE_HOST, MYAPP_SERVER_PORT 等
```

## 多环境配置

### 方式 1：多配置文件

```
configs/
├── config.yaml           # 默认/共享配置
├── config.development.yaml
├── config.staging.yaml
└── config.production.yaml
```

```go
func LoadWithEnv(env string) (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("configs")

    // 读取基础配置
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    // 读取环境特定配置（覆盖基础配置）
    viper.SetConfigName("config." + env)
    if err := viper.MergeInConfig(); err != nil {
        // 环境配置可选
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }

    var cfg Config
    return &cfg, viper.Unmarshal(&cfg)
}

// 使用
// APP_ENV=production ./myapp
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}
cfg, err := LoadWithEnv(env)
```

### 方式 2：环境变量覆盖

```yaml
# config.yaml - 开发默认值
database:
  host: localhost
  port: 5432

# 生产环境通过环境变量覆盖
# DATABASE_HOST=prod-db.example.com
# DATABASE_PORT=5432
```

## 敏感信息处理

### 不要在配置文件中硬编码密码

```yaml
# ✗ 错误
database:
  password: my-secret-password

# ✓ 正确 - 使用环境变量
database:
  password: ${DB_PASSWORD}
```

### 使用环境变量或密钥管理服务

```go
func Load() (*Config, error) {
    // ...

    // 从环境变量获取敏感信息
    viper.BindEnv("database.password", "DB_PASSWORD")
    viper.BindEnv("jwt.secret", "JWT_SECRET")
    viper.BindEnv("redis.password", "REDIS_PASSWORD")

    // 或者从密钥管理服务（如 AWS Secrets Manager）
    // password, _ := secretsManager.GetSecret("db-password")
    // viper.Set("database.password", password)

    // ...
}
```

## 配置热重载

```go
func LoadWithWatch(onChange func(*Config)) (*Config, error) {
    // ... 初始加载 ...

    // 监听配置文件变化
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        fmt.Println("Config file changed:", e.Name)

        var newCfg Config
        if err := viper.Unmarshal(&newCfg); err != nil {
            fmt.Println("Error reloading config:", err)
            return
        }

        onChange(&newCfg)
    })

    var cfg Config
    return &cfg, viper.Unmarshal(&cfg)
}

// 使用
cfg, _ := LoadWithWatch(func(newCfg *Config) {
    // 更新日志级别等可热更新的配置
    logger.SetLevel(newCfg.Log.Level)
})
```

## 完整示例

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Log      LogConfig      `mapstructure:"log"`
}

// ... 结构体定义 ...

var globalConfig *Config

// Load 加载配置
func Load() (*Config, error) {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }

    v := viper.New()

    // 配置文件设置
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("./configs")
    v.AddConfigPath(".")

    // 设置默认值
    setDefaults(v)

    // 读取基础配置
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("read config error: %w", err)
    }

    // 读取环境特定配置
    v.SetConfigName("config." + env)
    _ = v.MergeInConfig() // 忽略不存在的错误

    // 环境变量支持
    v.SetEnvPrefix("APP")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // 绑定敏感配置到环境变量
    v.BindEnv("database.password", "DB_PASSWORD")
    v.BindEnv("jwt.secret", "JWT_SECRET")
    v.BindEnv("redis.password", "REDIS_PASSWORD")

    // 解析到结构体
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config error: %w", err)
    }

    // 验证
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    globalConfig = &cfg
    return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
    return globalConfig
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        c.Host, c.Port, c.User, c.Password, c.Name,
    )
}

// RedisAddr 返回 Redis 地址
func (c *RedisConfig) Addr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
```

```go
// cmd/api/main.go
package main

import (
    "log"

    "myapp/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    log.Printf("Starting server on %s in %s mode", cfg.Server.Port, cfg.Server.Mode)
    log.Printf("Database: %s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

    // ... 启动应用 ...
}
```

## 最佳实践

1. **使用结构体绑定** - 类型安全，IDE 支持
2. **设置合理的默认值** - 减少必需配置
3. **支持环境变量覆盖** - 便于容器化部署
4. **分离敏感信息** - 不要将密码提交到代码库
5. **验证配置** - 在启动时验证必需配置
6. **使用配置前缀** - 避免环境变量冲突
7. **考虑热重载** - 对于需要动态更新的配置
