// Package main demonstrates Viper configuration management
package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    "github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    Log      LogConfig      `mapstructure:"log"`
    Debug    bool           `mapstructure:"debug"`
}

type ServerConfig struct {
    Port         string        `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Name     string `mapstructure:"name"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
}

type RedisConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    DB   int    `mapstructure:"db"`
}

type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
}

func main() {
    fmt.Println("=== Viper Configuration Demo ===\n")

    // Demo 1: Basic Viper usage
    fmt.Println("--- Demo 1: Basic Viper Usage ---")
    basicDemo()

    // Demo 2: Loading from YAML
    fmt.Println("\n--- Demo 2: Load from YAML ---")
    yamlDemo()

    // Demo 3: Environment variables
    fmt.Println("\n--- Demo 3: Environment Variables ---")
    envDemo()

    // Demo 4: Struct binding
    fmt.Println("\n--- Demo 4: Struct Binding ---")
    structDemo()

    // Demo 5: Default values
    fmt.Println("\n--- Demo 5: Default Values ---")
    defaultsDemo()
}

func basicDemo() {
    v := viper.New()

    // Set values directly
    v.Set("app.name", "myapp")
    v.Set("app.version", "1.0.0")
    v.Set("app.debug", true)

    // Get values
    fmt.Println("App Name:", v.GetString("app.name"))
    fmt.Println("Version:", v.GetString("app.version"))
    fmt.Println("Debug:", v.GetBool("app.debug"))
}

func yamlDemo() {
    // Create a temporary config file
    configContent := `
server:
  port: ":8080"
  read_timeout: 10s
  write_timeout: 10s

database:
  host: localhost
  port: 5432
  name: myapp
  user: postgres
  password: secret

redis:
  host: localhost
  port: 6379
  db: 0

log:
  level: info
  format: json

debug: true
`

    // Write temporary config
    tmpFile := "/tmp/config_demo.yaml"
    if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
        log.Fatal(err)
    }
    defer os.Remove(tmpFile)

    v := viper.New()
    v.SetConfigFile(tmpFile)

    if err := v.ReadInConfig(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Server Port:", v.GetString("server.port"))
    fmt.Println("DB Host:", v.GetString("database.host"))
    fmt.Println("DB Port:", v.GetInt("database.port"))
    fmt.Println("Log Level:", v.GetString("log.level"))
}

func envDemo() {
    v := viper.New()

    // Set defaults
    v.SetDefault("database.host", "localhost")
    v.SetDefault("database.port", 5432)

    // Enable environment variables
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // Bind specific env vars
    v.BindEnv("database.host", "DB_HOST")
    v.BindEnv("database.port", "DB_PORT")

    // Set environment variables for demo
    os.Setenv("DB_HOST", "production-db.example.com")
    os.Setenv("DB_PORT", "5433")
    defer os.Unsetenv("DB_HOST")
    defer os.Unsetenv("DB_PORT")

    // Environment variables override defaults
    fmt.Println("DB Host (from env):", v.GetString("database.host"))
    fmt.Println("DB Port (from env):", v.GetInt("database.port"))
}

func structDemo() {
    configContent := `
server:
  port: ":9090"
  read_timeout: 15s
  write_timeout: 15s

database:
  host: db.example.com
  port: 5432
  name: production
  user: admin
  password: supersecret

redis:
  host: redis.example.com
  port: 6379
  db: 1

log:
  level: warn
  format: text

debug: false
`

    tmpFile := "/tmp/config_struct_demo.yaml"
    if err := os.WriteFile(tmpFile, []byte(configContent), 0644); err != nil {
        log.Fatal(err)
    }
    defer os.Remove(tmpFile)

    v := viper.New()
    v.SetConfigFile(tmpFile)

    if err := v.ReadInConfig(); err != nil {
        log.Fatal(err)
    }

    // Unmarshal to struct
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Config: %+v\n", cfg)
    fmt.Printf("Server Port: %s\n", cfg.Server.Port)
    fmt.Printf("Server ReadTimeout: %v\n", cfg.Server.ReadTimeout)
    fmt.Printf("Database: %s@%s:%d/%s\n",
        cfg.Database.User,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Name,
    )
}

func defaultsDemo() {
    v := viper.New()

    // Set defaults
    v.SetDefault("server.port", ":8080")
    v.SetDefault("server.read_timeout", "10s")
    v.SetDefault("server.write_timeout", "10s")
    v.SetDefault("database.host", "localhost")
    v.SetDefault("database.port", 5432)
    v.SetDefault("database.max_connections", 25)
    v.SetDefault("log.level", "info")
    v.SetDefault("log.format", "json")
    v.SetDefault("debug", false)

    // Without loading any config file, defaults are used
    fmt.Println("Server Port (default):", v.GetString("server.port"))
    fmt.Println("DB Host (default):", v.GetString("database.host"))
    fmt.Println("Max Connections (default):", v.GetInt("database.max_connections"))
    fmt.Println("Log Level (default):", v.GetString("log.level"))
    fmt.Println("Debug (default):", v.GetBool("debug"))

    // Override with Set
    v.Set("debug", true)
    fmt.Println("Debug (overridden):", v.GetBool("debug"))
}

/*
Real-world usage pattern:

package config

import (
    "github.com/spf13/viper"
)

var globalConfig *Config

func Load() (*Config, error) {
    v := viper.New()

    // Set config file
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("./configs")
    v.AddConfigPath(".")

    // Set defaults
    setDefaults(v)

    // Enable env vars
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // Bind sensitive configs to env vars
    v.BindEnv("database.password", "DB_PASSWORD")
    v.BindEnv("jwt.secret", "JWT_SECRET")

    // Read config
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }

    // Unmarshal
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    // Validate
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    globalConfig = &cfg
    return &cfg, nil
}

func Get() *Config {
    return globalConfig
}
*/
