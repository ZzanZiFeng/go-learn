# 分层架构 (Layered Architecture)

## 概述

分层架构是最常见的后端架构模式，将应用分为多个层，每层有特定的职责。

```
┌─────────────────────────────────────────┐
│            Presentation Layer           │  处理 HTTP 请求/响应
│              (Handlers)                 │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│             Business Layer              │  业务逻辑
│              (Services)                 │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│              Data Layer                 │  数据访问
│            (Repositories)               │
└─────────────────────────────────────────┘
                    │
                    ▼
              Database/Cache
```

## 各层职责

### Handler 层（表现层）

**职责**：
- 解析 HTTP 请求（路径参数、查询参数、请求体）
- 验证输入数据
- 调用 Service 层
- 格式化 HTTP 响应
- 处理 HTTP 错误

**不应该做**：
- 业务逻辑
- 直接访问数据库
- 复杂的数据转换

```go
// internal/handlers/user.go
package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "myproject/internal/services"
)

type UserHandler struct {
    service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
    return &UserHandler{service: service}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Email string `json:"email" binding:"required,email"`
    Name  string `json:"name" binding:"required,min=2,max=100"`
}

// UserResponse 用户响应
type UserResponse struct {
    ID    uint   `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

// Create 创建用户
// @Summary 创建新用户
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request",
            "details": err.Error(),
        })
        return
    }

    user, err := h.service.CreateUser(c.Request.Context(), services.CreateUserInput{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, UserResponse{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    })
}

// GetByID 根据 ID 获取用户
func (h *UserHandler) GetByID(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    user, err := h.service.GetUserByID(c.Request.Context(), uint(id))
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, UserResponse{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    })
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

    users, total, err := h.service.ListUsers(c.Request.Context(), page, pageSize)
    if err != nil {
        handleError(c, err)
        return
    }

    response := make([]UserResponse, len(users))
    for i, u := range users {
        response[i] = UserResponse{
            ID:    u.ID,
            Email: u.Email,
            Name:  u.Name,
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "data":      response,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}
```

### Service 层（业务层）

**职责**：
- 实现业务逻辑
- 事务管理
- 调用多个 Repository
- 业务规则验证
- 调用外部服务

**不应该做**：
- HTTP 相关操作
- 直接的 SQL 查询

```go
// internal/services/user.go
package services

import (
    "context"
    "errors"

    "myproject/internal/models"
    "myproject/internal/repositories"
)

var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("user already exists")
    ErrInvalidInput = errors.New("invalid input")
)

// UserService 用户服务接口
type UserService interface {
    CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
    GetUserByID(ctx context.Context, id uint) (*models.User, error)
    ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error)
    UpdateUser(ctx context.Context, id uint, input UpdateUserInput) (*models.User, error)
    DeleteUser(ctx context.Context, id uint) error
}

// CreateUserInput 创建用户输入
type CreateUserInput struct {
    Email string
    Name  string
}

// UpdateUserInput 更新用户输入
type UpdateUserInput struct {
    Name *string
}

type userService struct {
    repo         repositories.UserRepository
    emailService EmailService // 外部服务依赖
}

func NewUserService(repo repositories.UserRepository, emailSvc EmailService) UserService {
    return &userService{
        repo:         repo,
        emailService: emailSvc,
    }
}

func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error) {
    // 业务规则：检查邮箱是否已存在
    existing, err := s.repo.FindByEmail(ctx, input.Email)
    if err == nil && existing != nil {
        return nil, ErrUserExists
    }

    // 创建用户
    user := &models.User{
        Email: input.Email,
        Name:  input.Name,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }

    // 业务逻辑：发送欢迎邮件（异步）
    go s.emailService.SendWelcomeEmail(user.Email)

    return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}

func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error) {
    // 业务规则：限制页面大小
    if pageSize > 100 {
        pageSize = 100
    }
    if pageSize < 1 {
        pageSize = 10
    }
    if page < 1 {
        page = 1
    }

    offset := (page - 1) * pageSize
    return s.repo.FindAll(ctx, offset, pageSize)
}

func (s *userService) UpdateUser(ctx context.Context, id uint, input UpdateUserInput) (*models.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrUserNotFound
    }

    // 部分更新
    if input.Name != nil {
        user.Name = *input.Name
    }

    if err := s.repo.Update(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
    _, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return ErrUserNotFound
    }

    return s.repo.Delete(ctx, id)
}
```

### Repository 层（数据层）

**职责**：
- 封装数据访问逻辑
- 执行 CRUD 操作
- 数据映射

**不应该做**：
- 业务逻辑
- HTTP 相关操作

```go
// internal/repositories/user.go
package repositories

import (
    "context"

    "gorm.io/gorm"
    "myproject/internal/models"
)

// UserRepository 用户仓储接口
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByID(ctx context.Context, id uint) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
    FindAll(ctx context.Context, offset, limit int) ([]*models.User, int64, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uint) error
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

func (r *userRepository) FindAll(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
    var users []*models.User
    var total int64

    // 获取总数
    if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 获取分页数据
    if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}
```

## 与 Node.js 对比

### Express 三层架构

```javascript
// routes/userRoutes.js (相当于 Go 的 handler)
router.post('/users', userController.create);

// controllers/userController.js
exports.create = async (req, res) => {
  try {
    const user = await userService.createUser(req.body);
    res.status(201).json(user);
  } catch (error) {
    res.status(400).json({ error: error.message });
  }
};

// services/userService.js
exports.createUser = async (data) => {
  const existing = await User.findOne({ email: data.email });
  if (existing) throw new Error('User exists');
  return User.create(data);
};
```

### Go 分层架构

```go
// 路由设置
router.POST("/users", handler.Create)

// handlers/user.go
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    user, err := h.service.CreateUser(c.Request.Context(), input)
    // ...
}

// services/user.go
func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error) {
    existing, _ := s.repo.FindByEmail(ctx, input.Email)
    if existing != nil {
        return nil, ErrUserExists
    }
    // ...
}
```

## 接口与实现分离

使用接口定义各层的契约，实现松耦合：

```go
// 定义接口
type UserService interface {
    CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
}

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}

// 实现接口
type userService struct {
    repo UserRepository  // 依赖接口，不是具体实现
}

type userRepository struct {
    db *gorm.DB
}
```

### 好处

1. **可测试性** - 可以轻松 mock 依赖
2. **可替换性** - 可以替换实现而不影响使用方
3. **解耦** - 各层之间通过接口通信

## 事务处理

在 Service 层处理跨多个 Repository 的事务：

```go
// services/order.go
func (s *orderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*models.Order, error) {
    // 使用事务
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 创建订单
    order := &models.Order{UserID: input.UserID}
    if err := tx.Create(order).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // 创建订单项
    for _, item := range input.Items {
        orderItem := &models.OrderItem{
            OrderID:   order.ID,
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
        }
        if err := tx.Create(orderItem).Error; err != nil {
            tx.Rollback()
            return nil, err
        }

        // 更新库存
        if err := tx.Model(&models.Product{}).
            Where("id = ?", item.ProductID).
            Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
            tx.Rollback()
            return nil, err
        }
    }

    return order, tx.Commit().Error
}
```

## 错误处理统一

```go
// internal/handlers/errors.go
package handlers

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "myproject/internal/services"
)

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details string `json:"details,omitempty"`
}

func handleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, services.ErrUserNotFound):
        c.JSON(http.StatusNotFound, ErrorResponse{
            Error: "User not found",
            Code:  "USER_NOT_FOUND",
        })
    case errors.Is(err, services.ErrUserExists):
        c.JSON(http.StatusConflict, ErrorResponse{
            Error: "User already exists",
            Code:  "USER_EXISTS",
        })
    case errors.Is(err, services.ErrInvalidInput):
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "Invalid input",
            Code:  "INVALID_INPUT",
        })
    default:
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Internal server error",
            Code:  "INTERNAL_ERROR",
        })
    }
}
```

## 总结

| 层 | 职责 | 依赖 |
|---|------|------|
| Handler | HTTP 处理 | Service |
| Service | 业务逻辑 | Repository, 外部服务 |
| Repository | 数据访问 | Database |

**关键原则**：
1. 单向依赖：Handler → Service → Repository
2. 接口定义契约
3. 各层职责单一
4. 通过依赖注入组装
