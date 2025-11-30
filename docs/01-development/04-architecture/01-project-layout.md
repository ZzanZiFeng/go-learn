# 项目布局 (Project Layout)

## Go 标准项目布局

Go 社区有一个广泛接受的项目布局标准：[golang-standards/project-layout](https://github.com/golang-standards/project-layout)

### 完整结构

```
myproject/
├── cmd/                    # 应用入口点
│   ├── api/
│   │   └── main.go
│   └── worker/
│       └── main.go
│
├── internal/               # 私有应用代码
│   ├── handlers/           # HTTP 处理器
│   ├── services/           # 业务逻辑
│   ├── repositories/       # 数据访问
│   ├── models/             # 数据模型
│   ├── middleware/         # 中间件
│   ├── config/             # 配置结构
│   └── pkg/                # 内部共享包
│
├── pkg/                    # 可被外部导入的公共库
│   ├── logger/
│   ├── validator/
│   └── httpclient/
│
├── api/                    # API 定义文件
│   ├── openapi/
│   │   └── spec.yaml
│   └── proto/              # gRPC protobuf 文件
│
├── configs/                # 配置文件
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
│
├── scripts/                # 脚本
│   ├── build.sh
│   └── migrate.sh
│
├── deployments/            # 部署配置
│   ├── docker/
│   │   └── Dockerfile
│   └── kubernetes/
│
├── docs/                   # 文档
│
├── tests/                  # 额外的测试
│   ├── e2e/
│   └── integration/
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 目录详解

### `/cmd` - 应用入口

每个应用一个子目录，每个子目录包含一个 `main.go`：

```go
// cmd/api/main.go
package main

import (
    "log"

    "myproject/internal/config"
    "myproject/internal/handlers"
    "myproject/internal/repositories"
    "myproject/internal/services"
)

func main() {
    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // 初始化依赖
    db := initDB(cfg.Database)
    repo := repositories.NewUserRepository(db)
    svc := services.NewUserService(repo)
    handler := handlers.NewUserHandler(svc)

    // 启动服务器
    router := setupRouter(handler)
    log.Fatal(router.Run(cfg.Server.Port))
}
```

```go
// cmd/worker/main.go
package main

import (
    "log"

    "myproject/internal/config"
    "myproject/internal/worker"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    w := worker.New(cfg)
    log.Fatal(w.Run())
}
```

### `/internal` - 私有代码

Go 编译器强制 `internal` 目录下的包只能被同一模块内的代码导入：

```
myproject/
├── internal/
│   └── config/         # 只能被 myproject 内的代码导入
│       └── config.go
└── cmd/
    └── api/
        └── main.go     # ✓ 可以导入 internal/config
```

```go
// 外部项目
import "myproject/internal/config"  // ✗ 编译错误！
```

#### 典型的 internal 结构

```go
// internal/models/user.go
package models

import "time"

type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Email     string    `json:"email" gorm:"uniqueIndex"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

```go
// internal/repositories/user.go
package repositories

import (
    "context"

    "gorm.io/gorm"
    "myproject/internal/models"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByID(ctx context.Context, id uint) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### `/pkg` - 公共库

可以被任何项目导入的代码：

```go
// pkg/validator/validator.go
package validator

import (
    "regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
    return emailRegex.MatchString(email)
}

func IsNotEmpty(s string) bool {
    return len(s) > 0
}
```

```go
// pkg/httpclient/client.go
package httpclient

import (
    "net/http"
    "time"
)

func New(timeout time.Duration) *http.Client {
    return &http.Client{
        Timeout: timeout,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
}
```

### `/configs` - 配置文件

```yaml
# configs/config.yaml
server:
  port: ":8080"
  read_timeout: 10s
  write_timeout: 10s

database:
  host: localhost
  port: 5432
  name: myapp
  user: postgres
  password: ${DB_PASSWORD}  # 环境变量占位符

redis:
  host: localhost
  port: 6379
  db: 0

log:
  level: info
  format: json
```

## 与 Node.js 对比

### Node.js (Express/NestJS)

```
express-app/
├── src/
│   ├── controllers/
│   ├── services/
│   ├── models/
│   ├── middleware/
│   └── app.js
├── config/
├── tests/
└── package.json
```

### Go 等价结构

```
go-app/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/      # = controllers
│   ├── services/
│   ├── models/
│   ├── repositories/  # Node.js 通常没有这层
│   └── middleware/
├── configs/
├── tests/
└── go.mod
```

## 简化布局（小型项目）

对于小型项目，可以简化：

```
small-app/
├── main.go            # 入口
├── handlers.go        # 所有 handler
├── services.go        # 所有 service
├── models.go          # 所有 model
├── config.go          # 配置
├── config.yaml
└── go.mod
```

## 模块和包的组织原则

### 1. 按功能/领域组织（推荐）

```
internal/
├── user/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── model.go
├── order/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── model.go
└── payment/
    └── ...
```

### 2. 按层组织

```
internal/
├── handlers/
│   ├── user.go
│   └── order.go
├── services/
│   ├── user.go
│   └── order.go
├── repositories/
│   ├── user.go
│   └── order.go
└── models/
    ├── user.go
    └── order.go
```

### 选择建议

| 项目特点 | 推荐组织方式 |
|---------|-------------|
| 小型、功能简单 | 按层组织 |
| 中大型、多领域 | 按功能组织 |
| 微服务 | 按功能组织 |

## 避免的反模式

### 1. 过度嵌套

```
# ✗ 不好
internal/core/domain/entities/user/user.go

# ✓ 好
internal/models/user.go
```

### 2. 循环依赖

```
# ✗ 循环依赖
handlers -> services -> handlers

# ✓ 单向依赖
handlers -> services -> repositories
```

### 3. 把所有东西放在 `pkg`

```
# ✗ pkg 不是垃圾桶
pkg/
├── user/           # 应该在 internal
├── config/         # 应该在 internal
└── handlers/       # 应该在 internal

# ✓ pkg 只放真正通用的代码
pkg/
├── logger/         # 通用日志
└── validator/      # 通用验证
```

## Makefile 示例

```makefile
.PHONY: build run test lint clean

# 变量
BINARY_NAME=myapp
BUILD_DIR=bin

# 构建
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api

# 运行
run:
	go run ./cmd/api

# 测试
test:
	go test -v ./...

# 带覆盖率测试
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# 代码检查
lint:
	golangci-lint run ./...

# 清理
clean:
	rm -rf $(BUILD_DIR)

# 生成 swagger 文档
swagger:
	swag init -g cmd/api/main.go -o api/openapi

# 数据库迁移
migrate-up:
	migrate -path migrations -database "postgres://..." up

migrate-down:
	migrate -path migrations -database "postgres://..." down
```

## 总结

1. **`cmd/`** - 保持简洁，只做初始化和组装
2. **`internal/`** - 放所有业务代码
3. **`pkg/`** - 只放真正可复用的通用代码
4. **按功能组织** - 中大型项目推荐
5. **避免循环依赖** - 单向依赖流
6. **保持简单** - 不要过度工程化
