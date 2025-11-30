# 依赖注入 (Dependency Injection)

## 什么是依赖注入？

依赖注入（DI）是一种设计模式，将对象的依赖从内部创建改为外部传入。

### 没有 DI

```go
// ✗ 硬编码依赖
type UserService struct {
    repo *UserRepository // 直接依赖具体实现
}

func NewUserService() *UserService {
    db, _ := gorm.Open(postgres.Open("..."))  // 内部创建
    return &UserService{
        repo: &UserRepository{db: db},        // 内部创建
    }
}
```

### 使用 DI

```go
// ✓ 依赖注入
type UserService struct {
    repo UserRepository  // 依赖接口
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}  // 从外部传入
}
```

## 与 JavaScript/Node.js 对比

### Node.js DI

```javascript
// NestJS 使用装饰器
@Injectable()
class UserService {
  constructor(
    @Inject('UserRepository')
    private readonly userRepository: UserRepository,
  ) {}
}

// InversifyJS
@injectable()
class UserService {
  @inject(TYPES.UserRepository)
  private userRepository: UserRepository;
}
```

### Go 手动 DI

```go
// Go 更倾向于显式的手动注入
type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

// 在 main 中组装
func main() {
    db := initDB()
    repo := repositories.NewUserRepository(db)
    service := services.NewUserService(repo)
    handler := handlers.NewUserHandler(service)
}
```

## Go 中的三种 DI 方式

### 1. 构造函数注入（推荐）

```go
type UserHandler struct {
    service UserService
}

func NewUserHandler(service UserService) *UserHandler {
    return &UserHandler{service: service}
}

// 使用
handler := NewUserHandler(userService)
```

### 2. 方法注入

```go
type EmailSender struct{}

func (e *EmailSender) SendWithTemplate(tpl Template, to string, data interface{}) error {
    // tpl 通过方法参数注入
    return tpl.Execute(to, data)
}
```

### 3. 接口注入

```go
type Configurable interface {
    Configure(config Config)
}

type Server struct {
    config Config
}

func (s *Server) Configure(config Config) {
    s.config = config
}
```

## 完整示例：手动 DI

### 项目结构

```
myapp/
├── cmd/
│   └── api/
│       └── main.go         # 依赖组装
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   └── database.go
│   ├── handlers/
│   │   └── user.go
│   ├── services/
│   │   └── user.go
│   └── repositories/
│       └── user.go
└── go.mod
```

### 代码实现

```go
// internal/config/config.go
package config

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type DatabaseConfig struct {
    Host     string
    Port     int
    Name     string
    User     string
    Password string
}

func Load() (*Config, error) {
    // 从环境变量或配置文件加载
    return &Config{
        Server: ServerConfig{
            Port:         getEnv("SERVER_PORT", ":8080"),
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnvInt("DB_PORT", 5432),
            Name:     getEnv("DB_NAME", "myapp"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", ""),
        },
    }, nil
}
```

```go
// internal/database/database.go
package database

import (
    "fmt"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "myapp/internal/config"
)

func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
    )

    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
```

```go
// internal/repositories/interfaces.go
package repositories

import (
    "context"
    "myapp/internal/models"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByID(ctx context.Context, id uint) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type OrderRepository interface {
    Create(ctx context.Context, order *models.Order) error
    FindByUserID(ctx context.Context, userID uint) ([]*models.Order, error)
}
```

```go
// internal/services/interfaces.go
package services

import (
    "context"
    "myapp/internal/models"
)

type UserService interface {
    CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
    GetUserByID(ctx context.Context, id uint) (*models.User, error)
}

type OrderService interface {
    CreateOrder(ctx context.Context, input CreateOrderInput) (*models.Order, error)
    GetUserOrders(ctx context.Context, userID uint) ([]*models.Order, error)
}
```

```go
// cmd/api/main.go - 依赖组装
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "myapp/internal/config"
    "myapp/internal/database"
    "myapp/internal/handlers"
    "myapp/internal/repositories"
    "myapp/internal/services"
)

func main() {
    // 1. 加载配置
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // 2. 初始化数据库
    db, err := database.New(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect database:", err)
    }

    // 3. 创建 Repositories
    userRepo := repositories.NewUserRepository(db)
    orderRepo := repositories.NewOrderRepository(db)

    // 4. 创建 Services（注入 Repositories）
    userService := services.NewUserService(userRepo)
    orderService := services.NewOrderService(orderRepo, userRepo)

    // 5. 创建 Handlers（注入 Services）
    userHandler := handlers.NewUserHandler(userService)
    orderHandler := handlers.NewOrderHandler(orderService)

    // 6. 设置路由
    router := gin.Default()
    setupRoutes(router, userHandler, orderHandler)

    // 7. 启动服务器
    log.Fatal(router.Run(cfg.Server.Port))
}

func setupRoutes(r *gin.Engine, userH *handlers.UserHandler, orderH *handlers.OrderHandler) {
    api := r.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.POST("", userH.Create)
            users.GET("/:id", userH.GetByID)
            users.GET("/:id/orders", orderH.GetUserOrders)
        }

        orders := api.Group("/orders")
        {
            orders.POST("", orderH.Create)
        }
    }
}
```

## 使用 Wire（Google 依赖注入工具）

Wire 是 Google 开发的编译时依赖注入工具。

### 安装

```bash
go install github.com/google/wire/cmd/wire@latest
```

### 项目结构

```
myapp/
├── cmd/
│   └── api/
│       ├── main.go
│       ├── wire.go         # Wire 配置
│       └── wire_gen.go     # Wire 生成的代码
└── internal/
    └── ...
```

### Wire 配置

```go
// cmd/api/wire.go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "myapp/internal/config"
    "myapp/internal/database"
    "myapp/internal/handlers"
    "myapp/internal/repositories"
    "myapp/internal/services"
)

// InitializeApp 初始化应用
func InitializeApp(cfg *config.Config) (*App, error) {
    wire.Build(
        // Database
        database.New,

        // Repositories
        repositories.NewUserRepository,
        repositories.NewOrderRepository,

        // Services
        services.NewUserService,
        services.NewOrderService,

        // Handlers
        handlers.NewUserHandler,
        handlers.NewOrderHandler,

        // App
        NewApp,
    )
    return nil, nil
}

// Provider Sets（可选，用于组织 providers）
var DatabaseSet = wire.NewSet(database.New)

var RepositorySet = wire.NewSet(
    repositories.NewUserRepository,
    repositories.NewOrderRepository,
)

var ServiceSet = wire.NewSet(
    services.NewUserService,
    services.NewOrderService,
)

var HandlerSet = wire.NewSet(
    handlers.NewUserHandler,
    handlers.NewOrderHandler,
)

// 使用 Sets 的版本
func InitializeAppWithSets(cfg *config.Config) (*App, error) {
    wire.Build(
        DatabaseSet,
        RepositorySet,
        ServiceSet,
        HandlerSet,
        NewApp,
    )
    return nil, nil
}
```

```go
// cmd/api/app.go
package main

import (
    "github.com/gin-gonic/gin"
    "myapp/internal/config"
    "myapp/internal/handlers"
)

type App struct {
    config       *config.Config
    router       *gin.Engine
    userHandler  *handlers.UserHandler
    orderHandler *handlers.OrderHandler
}

func NewApp(
    cfg *config.Config,
    userHandler *handlers.UserHandler,
    orderHandler *handlers.OrderHandler,
) *App {
    app := &App{
        config:       cfg,
        router:       gin.Default(),
        userHandler:  userHandler,
        orderHandler: orderHandler,
    }
    app.setupRoutes()
    return app
}

func (a *App) setupRoutes() {
    api := a.router.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.POST("", a.userHandler.Create)
            users.GET("/:id", a.userHandler.GetByID)
        }
    }
}

func (a *App) Run() error {
    return a.router.Run(a.config.Server.Port)
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
        log.Fatal("Failed to load config:", err)
    }

    app, err := InitializeApp(cfg)
    if err != nil {
        log.Fatal("Failed to initialize app:", err)
    }

    log.Fatal(app.Run())
}
```

### 生成代码

```bash
cd cmd/api
wire
```

这会生成 `wire_gen.go`：

```go
// Code generated by Wire. DO NOT EDIT.
//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
    "myapp/internal/config"
    "myapp/internal/database"
    "myapp/internal/handlers"
    "myapp/internal/repositories"
    "myapp/internal/services"
)

func InitializeApp(cfg *config.Config) (*App, error) {
    db, err := database.New(cfg.Database)
    if err != nil {
        return nil, err
    }
    userRepository := repositories.NewUserRepository(db)
    orderRepository := repositories.NewOrderRepository(db)
    userService := services.NewUserService(userRepository)
    orderService := services.NewOrderService(orderRepository, userRepository)
    userHandler := handlers.NewUserHandler(userService)
    orderHandler := handlers.NewOrderHandler(orderService)
    app := NewApp(cfg, userHandler, orderHandler)
    return app, nil
}
```

## Wire 高级特性

### 接口绑定

```go
// wire.go
var RepositorySet = wire.NewSet(
    repositories.NewUserRepository,
    // 将 *userRepository 绑定到 UserRepository 接口
    wire.Bind(new(repositories.UserRepository), new(*repositories.userRepository)),
)
```

### 结构体字段注入

```go
type App struct {
    Config  *config.Config
    DB      *gorm.DB
    Handler *handlers.UserHandler
}

// wire.go
func InitializeApp() (*App, error) {
    wire.Build(
        config.Load,
        database.New,
        handlers.NewUserHandler,
        wire.Struct(new(App), "*"), // 注入所有字段
    )
    return nil, nil
}
```

### 值注入

```go
func InitializeApp() (*App, error) {
    wire.Build(
        wire.Value(config.Config{
            Server: config.ServerConfig{Port: ":8080"},
        }),
        // ...
    )
    return nil, nil
}
```

## 测试中的 DI

依赖注入使测试变得简单：

```go
// services/user_test.go
package services

import (
    "context"
    "testing"

    "myapp/internal/models"
)

// Mock Repository
type mockUserRepository struct {
    users map[uint]*models.User
}

func newMockUserRepository() *mockUserRepository {
    return &mockUserRepository{
        users: make(map[uint]*models.User),
    }
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
    user.ID = uint(len(m.users) + 1)
    m.users[user.ID] = user
    return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, ErrUserNotFound
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    for _, user := range m.users {
        if user.Email == email {
            return user, nil
        }
    }
    return nil, nil
}

// 测试
func TestUserService_CreateUser(t *testing.T) {
    // 使用 mock 替代真实依赖
    mockRepo := newMockUserRepository()
    mockEmail := &mockEmailService{}
    service := NewUserService(mockRepo, mockEmail)

    // 测试
    user, err := service.CreateUser(context.Background(), CreateUserInput{
        Email: "test@example.com",
        Name:  "Test",
    })

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Email != "test@example.com" {
        t.Errorf("expected email test@example.com, got %s", user.Email)
    }
}
```

## 手动 DI vs Wire

| 特性 | 手动 DI | Wire |
|------|--------|------|
| 学习曲线 | 低 | 中 |
| 代码量 | 多（大型项目） | 少 |
| 类型安全 | 是 | 是（编译时） |
| 调试 | 简单 | 生成代码可读 |
| 适用场景 | 小中型项目 | 大型项目 |

## 总结

1. **优先使用构造函数注入** - 最清晰、最常用
2. **依赖接口而非具体实现** - 提高可测试性
3. **在 main 中组装依赖** - 集中管理
4. **小项目手动 DI** - 简单直接
5. **大项目考虑 Wire** - 减少样板代码
