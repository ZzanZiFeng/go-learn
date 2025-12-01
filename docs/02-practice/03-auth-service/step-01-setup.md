# Step 1: 项目初始化

## 目标

搭建 Auth Service 项目基础结构。

## 1.1 项目结构

```bash
cd projects/auth-service
ls -la

# 目录结构
# ├── cmd/server/
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

## 1.2 配置文件

创建 `configs/config.yaml`:

```yaml
server:
  port: 8081
  mode: debug

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: auth_service
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 1

jwt:
  access_secret: your-access-secret-key
  refresh_secret: your-refresh-secret-key
  access_expire_minutes: 15
  refresh_expire_days: 7

oauth:
  github:
    client_id: ""
    client_secret: ""
    redirect_url: "http://localhost:8081/api/oauth/callback/github"
  google:
    client_id: ""
    client_secret: ""
    redirect_url: "http://localhost:8081/api/oauth/callback/google"
```

## 1.3 配置管理

创建 `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	OAuth    OAuthConfig    `mapstructure:"oauth"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type JWTConfig struct {
	AccessSecret        string `mapstructure:"access_secret"`
	RefreshSecret       string `mapstructure:"refresh_secret"`
	AccessExpireMinutes int    `mapstructure:"access_expire_minutes"`
	RefreshExpireDays   int    `mapstructure:"refresh_expire_days"`
}

func (c *JWTConfig) AccessExpireDuration() time.Duration {
	return time.Duration(c.AccessExpireMinutes) * time.Minute
}

func (c *JWTConfig) RefreshExpireDuration() time.Duration {
	return time.Duration(c.RefreshExpireDays) * 24 * time.Hour
}

type OAuthConfig struct {
	GitHub OAuthProviderConfig `mapstructure:"github"`
	Google OAuthProviderConfig `mapstructure:"google"`
}

type OAuthProviderConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
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

func LoadDefault() (*Config, error) {
	return Load("configs/config.yaml")
}
```

## 1.4 数据库模型

创建 `internal/models/user.go`:

```go
package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	Username      string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email         string         `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password      string         `json:"-" gorm:"size:255"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	AvatarURL     string         `json:"avatar_url" gorm:"size:500"`
	Provider      string         `json:"provider" gorm:"size:20;default:'local'"` // local, github, google
	ProviderID    string         `json:"-" gorm:"size:100"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Roles []Role `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

func (User) TableName() string {
	return "users"
}
```

创建 `internal/models/role.go`:

```go
package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Description string         `json:"description" gorm:"size:200"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

func (Role) TableName() string {
	return "roles"
}
```

创建 `internal/models/permission.go`:

```go
package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Description string         `json:"description" gorm:"size:200"`
	Resource    string         `json:"resource" gorm:"size:50;not null"` // users, todos, roles
	Action      string         `json:"action" gorm:"size:20;not null"`   // create, read, update, delete
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Permission) TableName() string {
	return "permissions"
}
```

创建 `internal/models/token.go`:

```go
package models

import (
	"time"
)

type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Token     string    `json:"-" gorm:"uniqueIndex;size:500;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"-" gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
```

## 1.5 数据库迁移

创建 `migrations/001_init.sql`:

```sql
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    avatar_url VARCHAR(500),
    provider VARCHAR(20) DEFAULT 'local',
    provider_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(200),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description VARCHAR(200),
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 刷新令牌表
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);

-- 初始数据：角色
INSERT INTO roles (name, description) VALUES
    ('admin', '系统管理员'),
    ('user', '普通用户')
ON CONFLICT (name) DO NOTHING;

-- 初始数据：权限
INSERT INTO permissions (name, description, resource, action) VALUES
    ('users:read', '查看用户', 'users', 'read'),
    ('users:write', '编辑用户', 'users', 'write'),
    ('users:delete', '删除用户', 'users', 'delete'),
    ('roles:read', '查看角色', 'roles', 'read'),
    ('roles:write', '编辑角色', 'roles', 'write'),
    ('roles:delete', '删除角色', 'roles', 'delete')
ON CONFLICT (name) DO NOTHING;

-- 为 admin 角色分配所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;
```

## 检查点

完成本步骤后：

- [x] 项目结构已搭建
- [x] 配置管理已实现
- [x] 数据模型已定义
- [x] 数据库迁移已创建

## 下一步

[Step 2: 用户认证](./step-02-auth.md) - 实现注册和登录。
