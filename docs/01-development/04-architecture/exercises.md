# 后端架构练习

## 练习 1: 项目结构设计

**难度**: 基础

为一个博客系统设计项目结构，包含以下功能：
- 用户管理
- 文章管理
- 评论系统
- 标签系统

画出目录结构并解释每个目录的用途。

<details>
<summary>参考答案</summary>

```
blog-api/
├── cmd/
│   └── api/
│       └── main.go           # 应用入口，依赖组装
│
├── internal/
│   ├── config/               # 配置加载和结构
│   │   └── config.go
│   │
│   ├── models/               # 数据模型
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── comment.go
│   │   └── tag.go
│   │
│   ├── handlers/             # HTTP 处理器
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── comment.go
│   │   └── tag.go
│   │
│   ├── services/             # 业务逻辑
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── comment.go
│   │   └── tag.go
│   │
│   ├── repositories/         # 数据访问
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── comment.go
│   │   └── tag.go
│   │
│   └── middleware/           # 中间件
│       ├── auth.go
│       ├── logging.go
│       └── recovery.go
│
├── pkg/                      # 可复用的公共库
│   ├── validator/
│   └── pagination/
│
├── api/                      # API 定义
│   └── openapi/
│       └── spec.yaml
│
├── configs/                  # 配置文件
│   └── config.yaml
│
├── migrations/               # 数据库迁移
│   └── ...
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

</details>

---

## 练习 2: 实现分层架构

**难度**: 中级

实现一个简单的产品管理系统，包含三层架构：

```go
// 要求:
// 1. 定义 Product 模型
// 2. 实现 ProductRepository 接口和实现
// 3. 实现 ProductService 接口和实现
// 4. 实现 ProductHandler

type Product struct {
    ID          uint
    Name        string
    Description string
    Price       float64
    Stock       int
}

// 实现以下功能:
// - 创建产品
// - 获取产品
// - 更新库存
// - 列出产品（分页）
```

<details>
<summary>参考答案</summary>

```go
// models/product.go
package models

type Product struct {
    ID          uint    `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
    Stock       int     `json:"stock"`
}

// repositories/product.go
package repositories

import (
    "context"
    "errors"
    "sync"
)

var ErrNotFound = errors.New("not found")

type ProductRepository interface {
    Create(ctx context.Context, product *Product) error
    FindByID(ctx context.Context, id uint) (*Product, error)
    FindAll(ctx context.Context, offset, limit int) ([]*Product, int, error)
    Update(ctx context.Context, product *Product) error
}

type memoryProductRepo struct {
    mu       sync.RWMutex
    products map[uint]*Product
    nextID   uint
}

func NewProductRepository() ProductRepository {
    return &memoryProductRepo{
        products: make(map[uint]*Product),
        nextID:   1,
    }
}

func (r *memoryProductRepo) Create(ctx context.Context, p *Product) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    p.ID = r.nextID
    r.nextID++
    r.products[p.ID] = p
    return nil
}

func (r *memoryProductRepo) FindByID(ctx context.Context, id uint) (*Product, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    if p, ok := r.products[id]; ok {
        return p, nil
    }
    return nil, ErrNotFound
}

func (r *memoryProductRepo) FindAll(ctx context.Context, offset, limit int) ([]*Product, int, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    all := make([]*Product, 0, len(r.products))
    for _, p := range r.products {
        all = append(all, p)
    }

    total := len(all)
    if offset >= total {
        return []*Product{}, total, nil
    }

    end := offset + limit
    if end > total {
        end = total
    }

    return all[offset:end], total, nil
}

func (r *memoryProductRepo) Update(ctx context.Context, p *Product) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.products[p.ID]; !ok {
        return ErrNotFound
    }
    r.products[p.ID] = p
    return nil
}

// services/product.go
package services

import (
    "context"
    "errors"
)

var (
    ErrProductNotFound = errors.New("product not found")
    ErrInsufficientStock = errors.New("insufficient stock")
)

type ProductService interface {
    CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error)
    GetProduct(ctx context.Context, id uint) (*Product, error)
    UpdateStock(ctx context.Context, id uint, quantity int) error
    ListProducts(ctx context.Context, page, pageSize int) ([]*Product, int, error)
}

type CreateProductInput struct {
    Name        string
    Description string
    Price       float64
    Stock       int
}

type productService struct {
    repo ProductRepository
}

func NewProductService(repo ProductRepository) ProductService {
    return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error) {
    product := &Product{
        Name:        input.Name,
        Description: input.Description,
        Price:       input.Price,
        Stock:       input.Stock,
    }
    if err := s.repo.Create(ctx, product); err != nil {
        return nil, err
    }
    return product, nil
}

func (s *productService) GetProduct(ctx context.Context, id uint) (*Product, error) {
    product, err := s.repo.FindByID(ctx, id)
    if errors.Is(err, repositories.ErrNotFound) {
        return nil, ErrProductNotFound
    }
    return product, err
}

func (s *productService) UpdateStock(ctx context.Context, id uint, quantity int) error {
    product, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return ErrProductNotFound
    }

    newStock := product.Stock + quantity
    if newStock < 0 {
        return ErrInsufficientStock
    }

    product.Stock = newStock
    return s.repo.Update(ctx, product)
}

func (s *productService) ListProducts(ctx context.Context, page, pageSize int) ([]*Product, int, error) {
    if pageSize > 100 {
        pageSize = 100
    }
    offset := (page - 1) * pageSize
    return s.repo.FindAll(ctx, offset, pageSize)
}

// handlers/product.go
package handlers

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"
)

type ProductHandler struct {
    service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
    return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request")
        return
    }

    product, err := h.service.CreateProduct(r.Context(), CreateProductInput{
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        Stock:       req.Stock,
    })
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    product, err := h.service.GetProduct(r.Context(), uint(id))
    if errors.Is(err, services.ErrProductNotFound) {
        respondError(w, http.StatusNotFound, "Product not found")
        return
    }
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    var req struct {
        Quantity int `json:"quantity"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request")
        return
    }

    if err := h.service.UpdateStock(r.Context(), uint(id), req.Quantity); err != nil {
        if errors.Is(err, services.ErrInsufficientStock) {
            respondError(w, http.StatusBadRequest, "Insufficient stock")
            return
        }
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

</details>

---

## 练习 3: 配置管理

**难度**: 中级

使用 Viper 实现一个配置加载器：

1. 支持从 YAML 文件加载
2. 支持环境变量覆盖
3. 支持多环境（development, staging, production）
4. 实现配置验证

```go
// 配置结构
type Config struct {
    App      AppConfig
    Server   ServerConfig
    Database DatabaseConfig
}

type AppConfig struct {
    Name    string
    Version string
    Env     string // development, staging, production
}

// 实现 LoadConfig(env string) (*Config, error)
```

<details>
<summary>参考答案</summary>

```go
package config

import (
    "fmt"
    "strings"
    "time"

    "github.com/spf13/viper"
)

type Config struct {
    App      AppConfig      `mapstructure:"app"`
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
}

type AppConfig struct {
    Name    string `mapstructure:"name"`
    Version string `mapstructure:"version"`
    Env     string `mapstructure:"env"`
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

func LoadConfig(env string) (*Config, error) {
    v := viper.New()

    // 设置配置文件
    v.SetConfigType("yaml")
    v.AddConfigPath("./configs")
    v.AddConfigPath(".")

    // 设置默认值
    setDefaults(v)

    // 读取基础配置
    v.SetConfigName("config")
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("read base config: %w", err)
    }

    // 读取环境特定配置
    v.SetConfigName("config." + env)
    if err := v.MergeInConfig(); err != nil {
        // 环境配置可选
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("read %s config: %w", env, err)
        }
    }

    // 环境变量支持
    v.SetEnvPrefix("APP")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // 绑定敏感配置
    v.BindEnv("database.password", "DB_PASSWORD")

    // 设置环境
    v.Set("app.env", env)

    // 解析配置
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    // 验证配置
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }

    return &cfg, nil
}

func setDefaults(v *viper.Viper) {
    v.SetDefault("app.name", "myapp")
    v.SetDefault("app.version", "1.0.0")
    v.SetDefault("server.port", ":8080")
    v.SetDefault("server.read_timeout", "10s")
    v.SetDefault("server.write_timeout", "10s")
    v.SetDefault("database.host", "localhost")
    v.SetDefault("database.port", 5432)
}

func (c *Config) Validate() error {
    if c.App.Name == "" {
        return fmt.Errorf("app.name is required")
    }
    if c.Database.Host == "" {
        return fmt.Errorf("database.host is required")
    }
    if c.Database.Name == "" {
        return fmt.Errorf("database.name is required")
    }

    validEnvs := map[string]bool{
        "development": true,
        "staging":     true,
        "production":  true,
    }
    if !validEnvs[c.App.Env] {
        return fmt.Errorf("invalid app.env: %s", c.App.Env)
    }

    return nil
}
```

</details>

---

## 练习 4: 依赖注入

**难度**: 中级

重构以下代码，使用依赖注入提高可测试性：

```go
// 原始代码（紧耦合）
type OrderService struct {}

func (s *OrderService) CreateOrder(userID uint, items []Item) (*Order, error) {
    // 直接创建依赖
    db, _ := gorm.Open(postgres.Open("..."))
    userClient := &http.Client{}

    // 验证用户
    resp, _ := userClient.Get(fmt.Sprintf("http://user-service/users/%d", userID))
    // ...

    // 创建订单
    order := &Order{UserID: userID}
    db.Create(order)

    return order, nil
}
```

重构为：
1. 定义接口
2. 通过构造函数注入依赖
3. 编写单元测试

<details>
<summary>参考答案</summary>

```go
// 重构后的代码

// interfaces.go
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id uint) (*Order, error)
}

type UserClient interface {
    GetUser(ctx context.Context, id uint) (*User, error)
    ValidateUser(ctx context.Context, id uint) error
}

type InventoryService interface {
    CheckStock(ctx context.Context, items []Item) error
    ReserveStock(ctx context.Context, items []Item) error
}

// services/order.go
type OrderService struct {
    repo      OrderRepository
    userCli   UserClient
    inventory InventoryService
}

func NewOrderService(
    repo OrderRepository,
    userCli UserClient,
    inventory InventoryService,
) *OrderService {
    return &OrderService{
        repo:      repo,
        userCli:   userCli,
        inventory: inventory,
    }
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uint, items []Item) (*Order, error) {
    // 验证用户
    if err := s.userCli.ValidateUser(ctx, userID); err != nil {
        return nil, fmt.Errorf("validate user: %w", err)
    }

    // 检查库存
    if err := s.inventory.CheckStock(ctx, items); err != nil {
        return nil, fmt.Errorf("check stock: %w", err)
    }

    // 预留库存
    if err := s.inventory.ReserveStock(ctx, items); err != nil {
        return nil, fmt.Errorf("reserve stock: %w", err)
    }

    // 创建订单
    order := &Order{
        UserID: userID,
        Items:  items,
        Status: OrderStatusPending,
    }

    if err := s.repo.Create(ctx, order); err != nil {
        return nil, fmt.Errorf("create order: %w", err)
    }

    return order, nil
}

// services/order_test.go
type mockOrderRepo struct {
    orders map[uint]*Order
}

func (m *mockOrderRepo) Create(ctx context.Context, order *Order) error {
    order.ID = uint(len(m.orders) + 1)
    m.orders[order.ID] = order
    return nil
}

type mockUserClient struct {
    validUsers map[uint]bool
}

func (m *mockUserClient) ValidateUser(ctx context.Context, id uint) error {
    if !m.validUsers[id] {
        return errors.New("user not found")
    }
    return nil
}

type mockInventory struct {
    stock map[uint]int
}

func (m *mockInventory) CheckStock(ctx context.Context, items []Item) error {
    for _, item := range items {
        if m.stock[item.ProductID] < item.Quantity {
            return errors.New("insufficient stock")
        }
    }
    return nil
}

func (m *mockInventory) ReserveStock(ctx context.Context, items []Item) error {
    return nil
}

func TestOrderService_CreateOrder(t *testing.T) {
    // 创建 mock
    repo := &mockOrderRepo{orders: make(map[uint]*Order)}
    userCli := &mockUserClient{validUsers: map[uint]bool{1: true}}
    inventory := &mockInventory{stock: map[uint]int{100: 10}}

    // 注入依赖
    service := NewOrderService(repo, userCli, inventory)

    // 测试
    order, err := service.CreateOrder(context.Background(), 1, []Item{
        {ProductID: 100, Quantity: 2},
    })

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if order.UserID != 1 {
        t.Errorf("expected userID 1, got %d", order.UserID)
    }
}

func TestOrderService_CreateOrder_UserNotFound(t *testing.T) {
    repo := &mockOrderRepo{orders: make(map[uint]*Order)}
    userCli := &mockUserClient{validUsers: map[uint]bool{}}
    inventory := &mockInventory{}

    service := NewOrderService(repo, userCli, inventory)

    _, err := service.CreateOrder(context.Background(), 999, []Item{})

    if err == nil {
        t.Error("expected error for invalid user")
    }
}
```

</details>

---

## 练习 5: 错误处理设计

**难度**: 高级

设计一个完整的错误处理系统：

1. 定义业务错误类型
2. 实现错误包装
3. 统一 HTTP 错误响应
4. 错误日志记录

```go
// 要求:
// 1. 定义 AppError 类型，包含 Code, Message, Err 字段
// 2. 实现 error 接口和 Unwrap 方法
// 3. 实现 HTTP 错误处理中间件
// 4. 实现错误到 HTTP 状态码的映射
```

<details>
<summary>参考答案</summary>

```go
// errors/errors.go
package errors

import (
    "fmt"
    "net/http"
)

type ErrorCode string

const (
    ErrCodeNotFound      ErrorCode = "NOT_FOUND"
    ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"
    ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
    ErrCodeForbidden     ErrorCode = "FORBIDDEN"
    ErrCodeConflict      ErrorCode = "CONFLICT"
    ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
    ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
)

type AppError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Details any       `json:"details,omitempty"`
    Err     error     `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructor functions
func NewNotFoundError(resource string) *AppError {
    return &AppError{
        Code:    ErrCodeNotFound,
        Message: fmt.Sprintf("%s not found", resource),
    }
}

func NewValidationError(details any) *AppError {
    return &AppError{
        Code:    ErrCodeValidation,
        Message: "Validation failed",
        Details: details,
    }
}

func NewUnauthorizedError(message string) *AppError {
    return &AppError{
        Code:    ErrCodeUnauthorized,
        Message: message,
    }
}

func Wrap(err error, code ErrorCode, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

// HTTP status mapping
func (e *AppError) HTTPStatus() int {
    switch e.Code {
    case ErrCodeNotFound:
        return http.StatusNotFound
    case ErrCodeBadRequest, ErrCodeValidation:
        return http.StatusBadRequest
    case ErrCodeUnauthorized:
        return http.StatusUnauthorized
    case ErrCodeForbidden:
        return http.StatusForbidden
    case ErrCodeConflict:
        return http.StatusConflict
    default:
        return http.StatusInternalServerError
    }
}

// middleware/error.go
package middleware

import (
    "encoding/json"
    "errors"
    "log/slog"
    "net/http"

    apperrors "myapp/internal/errors"
)

type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

func ErrorHandler(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Use custom response writer to catch errors
            rw := &responseWriter{ResponseWriter: w}
            next.ServeHTTP(rw, r)

            if rw.err != nil {
                handleError(w, r, rw.err, logger)
            }
        })
    }
}

func handleError(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        // Log with context
        logger.Error("request error",
            "code", appErr.Code,
            "message", appErr.Message,
            "path", r.URL.Path,
            "method", r.Method,
            "error", appErr.Err,
        )

        respondError(w, appErr)
        return
    }

    // Unknown error - log and return generic error
    logger.Error("unexpected error",
        "path", r.URL.Path,
        "method", r.Method,
        "error", err,
    )

    respondError(w, &apperrors.AppError{
        Code:    apperrors.ErrCodeInternal,
        Message: "An unexpected error occurred",
    })
}

func respondError(w http.ResponseWriter, appErr *apperrors.AppError) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(appErr.HTTPStatus())
    json.NewEncoder(w).Encode(ErrorResponse{
        Code:    string(appErr.Code),
        Message: appErr.Message,
        Details: appErr.Details,
    })
}

// Usage in service
func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.NewNotFoundError("user")
        }
        return nil, apperrors.Wrap(err, apperrors.ErrCodeInternal, "failed to get user")
    }
    return user, nil
}
```

</details>

---

## 挑战练习: 设计电商订单系统架构

**难度**: 专家级

设计一个电商订单系统的架构，包含：

1. **服务拆分**
   - 用户服务
   - 商品服务
   - 库存服务
   - 订单服务
   - 支付服务

2. **设计要点**
   - 服务间通信方式
   - 数据一致性策略
   - 库存扣减方案
   - 订单状态机

3. **画出架构图并说明**

提示：
- 考虑事务边界
- 考虑服务降级
- 考虑幂等性
- 考虑分布式追踪

---

## 完成标准

- [ ] 完成练习 1-2（基础和中级）
- [ ] 完成练习 3-4（中级）
- [ ] 完成练习 5（高级）
- [ ] 代码通过 `go vet` 和 `golangci-lint` 检查
- [ ] 编写了单元测试
- [ ] 理解分层架构的优势

## 下一步

完成这些练习后，你可以：
1. 继续学习 [Web API 开发](../05-web-api/)
2. 开始 [Todo API 实战项目](../../02-practice/todo-api/)
