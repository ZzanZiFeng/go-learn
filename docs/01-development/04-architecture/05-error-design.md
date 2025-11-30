# 错误设计 (Error Design)

## 概述

Go 的错误处理与 JavaScript 的 try/catch 完全不同。Go 使用显式返回值来传递错误。

## 与 JavaScript 对比

### JavaScript 异常处理

```javascript
// 抛出异常
function divide(a, b) {
  if (b === 0) {
    throw new Error('Division by zero');
  }
  return a / b;
}

// 捕获异常
try {
  const result = divide(10, 0);
} catch (error) {
  console.error(error.message);
}

// 自定义错误类
class ValidationError extends Error {
  constructor(message, field) {
    super(message);
    this.name = 'ValidationError';
    this.field = field;
  }
}
```

### Go 错误处理

```go
// 返回错误
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// 检查错误
result, err := divide(10, 0)
if err != nil {
    log.Println("Error:", err)
    return
}
```

## 错误类型设计

### 1. 哨兵错误 (Sentinel Errors)

预定义的错误变量，用于比较：

```go
// 定义
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrAlreadyExists = errors.New("already exists")
)

// 使用
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrNotFound
    }
    return &user, err
}

// 检查
user, err := repo.FindByID(ctx, 123)
if errors.Is(err, ErrNotFound) {
    // 处理未找到
}
```

### 2. 自定义错误类型

携带更多上下文信息：

```go
// 定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// 使用
func validateUser(u *User) error {
    if u.Email == "" {
        return &ValidationError{Field: "email", Message: "is required"}
    }
    if len(u.Name) < 2 {
        return &ValidationError{Field: "name", Message: "must be at least 2 characters"}
    }
    return nil
}

// 类型断言检查
if err := validateUser(user); err != nil {
    var valErr *ValidationError
    if errors.As(err, &valErr) {
        fmt.Printf("Validation failed for field %s: %s\n", valErr.Field, valErr.Message)
    }
}
```

### 3. 错误包装 (Error Wrapping)

保留原始错误并添加上下文：

```go
import "fmt"

// 包装错误
func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
    user := &User{Email: input.Email, Name: input.Name}

    if err := s.repo.Create(ctx, user); err != nil {
        // 使用 %w 包装原始错误
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}

// 解包检查
err := service.CreateUser(ctx, input)
if errors.Is(err, ErrAlreadyExists) {
    // 即使被包装，仍然可以匹配原始错误
}

// 获取原始错误
var dbErr *DatabaseError
if errors.As(err, &dbErr) {
    // 获取包装链中的特定错误类型
}
```

## 业务错误设计

### 定义领域错误

```go
// internal/services/errors.go
package services

import "errors"

// 业务错误
var (
    ErrUserNotFound     = errors.New("user not found")
    ErrUserExists       = errors.New("user already exists")
    ErrInvalidPassword  = errors.New("invalid password")
    ErrTokenExpired     = errors.New("token expired")
    ErrPermissionDenied = errors.New("permission denied")
)

// 带详情的错误
type AppError struct {
    Code    string // 错误码
    Message string // 用户友好消息
    Err     error  // 原始错误
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

// 创建函数
func NewAppError(code, message string, err error) *AppError {
    return &AppError{Code: code, Message: message, Err: err}
}

// 预定义应用错误
var (
    ErrUserNotFoundApp = &AppError{
        Code:    "USER_NOT_FOUND",
        Message: "The requested user was not found",
    }
    ErrValidation = &AppError{
        Code:    "VALIDATION_ERROR",
        Message: "Request validation failed",
    }
)
```

### Service 层使用

```go
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}
```

## HTTP 错误响应

### 错误响应结构

```go
// internal/handlers/response.go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

// 常用错误响应
func BadRequest(c *gin.Context, message string, details any) {
    c.JSON(http.StatusBadRequest, ErrorResponse{
        Code:    "BAD_REQUEST",
        Message: message,
        Details: details,
    })
}

func NotFound(c *gin.Context, resource string) {
    c.JSON(http.StatusNotFound, ErrorResponse{
        Code:    "NOT_FOUND",
        Message: fmt.Sprintf("%s not found", resource),
    })
}

func InternalError(c *gin.Context) {
    c.JSON(http.StatusInternalServerError, ErrorResponse{
        Code:    "INTERNAL_ERROR",
        Message: "An internal error occurred",
    })
}

func Unauthorized(c *gin.Context, message string) {
    c.JSON(http.StatusUnauthorized, ErrorResponse{
        Code:    "UNAUTHORIZED",
        Message: message,
    })
}

func Forbidden(c *gin.Context) {
    c.JSON(http.StatusForbidden, ErrorResponse{
        Code:    "FORBIDDEN",
        Message: "You don't have permission to access this resource",
    })
}

func Conflict(c *gin.Context, message string) {
    c.JSON(http.StatusConflict, ErrorResponse{
        Code:    "CONFLICT",
        Message: message,
    })
}
```

### 统一错误处理

```go
// internal/handlers/error_handler.go
package handlers

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "myapp/internal/services"
)

// HandleError 统一错误处理
func HandleError(c *gin.Context, err error) {
    // 检查业务错误
    switch {
    case errors.Is(err, services.ErrUserNotFound):
        NotFound(c, "User")
    case errors.Is(err, services.ErrUserExists):
        Conflict(c, "User already exists")
    case errors.Is(err, services.ErrInvalidPassword):
        Unauthorized(c, "Invalid credentials")
    case errors.Is(err, services.ErrPermissionDenied):
        Forbidden(c)
    default:
        // 检查自定义错误类型
        var appErr *services.AppError
        if errors.As(err, &appErr) {
            handleAppError(c, appErr)
            return
        }

        var valErr *ValidationError
        if errors.As(err, &valErr) {
            BadRequest(c, "Validation failed", map[string]string{
                valErr.Field: valErr.Message,
            })
            return
        }

        // 未知错误
        // TODO: 记录日志
        InternalError(c)
    }
}

func handleAppError(c *gin.Context, err *services.AppError) {
    status := http.StatusInternalServerError

    switch err.Code {
    case "NOT_FOUND":
        status = http.StatusNotFound
    case "VALIDATION_ERROR":
        status = http.StatusBadRequest
    case "UNAUTHORIZED":
        status = http.StatusUnauthorized
    case "FORBIDDEN":
        status = http.StatusForbidden
    case "CONFLICT":
        status = http.StatusConflict
    }

    c.JSON(status, ErrorResponse{
        Code:    err.Code,
        Message: err.Message,
    })
}
```

### Handler 中使用

```go
func (h *UserHandler) GetByID(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid user ID", nil)
        return
    }

    user, err := h.service.GetUserByID(c.Request.Context(), uint(id))
    if err != nil {
        HandleError(c, err)
        return
    }

    c.JSON(http.StatusOK, user)
}
```

## 错误恢复 (Panic Recovery)

### Gin 中间件

```go
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                // 记录堆栈
                stack := debug.Stack()
                log.Printf("Panic recovered: %v\n%s", r, stack)

                // 返回 500
                c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
                    Code:    "INTERNAL_ERROR",
                    Message: "An unexpected error occurred",
                })
            }
        }()

        c.Next()
    }
}
```

### 何时使用 Panic

```go
// ✓ 适合 panic
func MustLoadConfig() *Config {
    cfg, err := LoadConfig()
    if err != nil {
        panic("failed to load config: " + err.Error())
    }
    return cfg
}

// ✗ 不适合 panic - 应该返回错误
func GetUser(id int) *User {
    user, err := db.FindUser(id)
    if err != nil {
        panic(err)  // 不要这样做！
    }
    return user
}
```

## 错误日志最佳实践

```go
import (
    "go.uber.org/zap"
)

var logger *zap.Logger

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
    user := &User{Email: input.Email}

    if err := s.repo.Create(ctx, user); err != nil {
        // 记录错误（带上下文）
        logger.Error("failed to create user",
            zap.String("email", input.Email),
            zap.Error(err),
        )

        // 返回包装的错误（不暴露内部细节给用户）
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    logger.Info("user created",
        zap.Uint("user_id", user.ID),
        zap.String("email", user.Email),
    )

    return user, nil
}
```

## 错误设计原则

1. **错误应该被处理或传播**
   ```go
   // ✗ 忽略错误
   result, _ := doSomething()

   // ✓ 处理或传播
   result, err := doSomething()
   if err != nil {
       return nil, fmt.Errorf("doSomething failed: %w", err)
   }
   ```

2. **在边界层处理错误转换**
   - Handler 层：将业务错误转为 HTTP 响应
   - Service 层：将数据层错误转为业务错误

3. **包装错误时添加上下文**
   ```go
   return fmt.Errorf("failed to get user %d: %w", userID, err)
   ```

4. **使用 errors.Is 和 errors.As**
   ```go
   // 检查错误类型
   if errors.Is(err, ErrNotFound) {}

   // 提取错误详情
   var appErr *AppError
   if errors.As(err, &appErr) {}
   ```

5. **不要在错误消息中包含敏感信息**
   ```go
   // ✗ 不安全
   return fmt.Errorf("invalid password for user %s", password)

   // ✓ 安全
   return ErrInvalidPassword
   ```

## 总结

| 场景 | 推荐方式 |
|------|---------|
| 简单错误 | 哨兵错误 (`var ErrNotFound = errors.New(...)`) |
| 需要额外信息 | 自定义错误类型 |
| 添加上下文 | 错误包装 (`fmt.Errorf("...: %w", err)`) |
| HTTP 响应 | 统一错误处理中间件 |
| 程序初始化失败 | panic |
| 业务逻辑错误 | 返回 error |
