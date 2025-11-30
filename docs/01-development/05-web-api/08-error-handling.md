# 错误处理 (Error Handling)

## 概述

良好的 API 错误处理对用户体验和调试至关重要。本节介绍 Gin 中的错误处理最佳实践。

## 与 Express.js 对比

### Express.js

```javascript
// 错误处理中间件
app.use((err, req, res, next) => {
  console.error(err);
  res.status(err.status || 500).json({
    error: {
      code: err.code || 'INTERNAL_ERROR',
      message: err.message || 'Something went wrong'
    }
  });
});

// 抛出错误
app.get('/users/:id', async (req, res, next) => {
  try {
    const user = await findUser(req.params.id);
    if (!user) {
      const error = new Error('User not found');
      error.status = 404;
      error.code = 'NOT_FOUND';
      throw error;
    }
    res.json(user);
  } catch (err) {
    next(err);
  }
});
```

### Gin

```go
// 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            // 处理错误...
        }
    }
}

// 处理错误
r.GET("/users/:id", func(c *gin.Context) {
    user, err := findUser(c.Param("id"))
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            c.JSON(404, gin.H{"error": "User not found"})
            return
        }
        c.JSON(500, gin.H{"error": "Internal error"})
        return
    }
    c.JSON(200, user)
})
```

## 定义错误类型

### 应用错误结构

```go
// pkg/apperror/error.go
package apperror

import "net/http"

// AppError 应用错误
type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    HTTPStatus int    `json:"-"`
    Details    any    `json:"details,omitempty"`
    Err        error  `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return e.Err.Error()
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// 创建错误的工厂函数
func New(code string, message string, status int) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        HTTPStatus: status,
    }
}

func Wrap(err error, code string, message string, status int) *AppError {
    return &AppError{
        Code:       code,
        Message:    message,
        HTTPStatus: status,
        Err:        err,
    }
}

func (e *AppError) WithDetails(details any) *AppError {
    e.Details = details
    return e
}
```

### 预定义错误

```go
// pkg/apperror/errors.go
package apperror

import "net/http"

// 通用错误
var (
    ErrBadRequest = &AppError{
        Code:       "BAD_REQUEST",
        Message:    "Bad request",
        HTTPStatus: http.StatusBadRequest,
    }

    ErrUnauthorized = &AppError{
        Code:       "UNAUTHORIZED",
        Message:    "Authentication required",
        HTTPStatus: http.StatusUnauthorized,
    }

    ErrForbidden = &AppError{
        Code:       "FORBIDDEN",
        Message:    "Access denied",
        HTTPStatus: http.StatusForbidden,
    }

    ErrNotFound = &AppError{
        Code:       "NOT_FOUND",
        Message:    "Resource not found",
        HTTPStatus: http.StatusNotFound,
    }

    ErrConflict = &AppError{
        Code:       "CONFLICT",
        Message:    "Resource already exists",
        HTTPStatus: http.StatusConflict,
    }

    ErrValidation = &AppError{
        Code:       "VALIDATION_ERROR",
        Message:    "Validation failed",
        HTTPStatus: http.StatusBadRequest,
    }

    ErrInternal = &AppError{
        Code:       "INTERNAL_ERROR",
        Message:    "Internal server error",
        HTTPStatus: http.StatusInternalServerError,
    }

    ErrServiceUnavailable = &AppError{
        Code:       "SERVICE_UNAVAILABLE",
        Message:    "Service temporarily unavailable",
        HTTPStatus: http.StatusServiceUnavailable,
    }
)

// 资源特定错误
func NotFound(resource string) *AppError {
    return &AppError{
        Code:       "NOT_FOUND",
        Message:    resource + " not found",
        HTTPStatus: http.StatusNotFound,
    }
}

func AlreadyExists(resource string) *AppError {
    return &AppError{
        Code:       "ALREADY_EXISTS",
        Message:    resource + " already exists",
        HTTPStatus: http.StatusConflict,
    }
}

func InvalidInput(field, reason string) *AppError {
    return &AppError{
        Code:       "INVALID_INPUT",
        Message:    "Invalid " + field + ": " + reason,
        HTTPStatus: http.StatusBadRequest,
    }
}
```

## 错误处理中间件

### 基础错误处理

```go
// pkg/middleware/error.go
package middleware

import (
    "log/slog"
    "net/http"

    "myapp/pkg/apperror"

    "github.com/gin-gonic/gin"
)

type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 检查是否有错误
        if len(c.Errors) == 0 {
            return
        }

        err := c.Errors.Last().Err

        // 处理应用错误
        var appErr *apperror.AppError
        if errors.As(err, &appErr) {
            // 记录服务器错误
            if appErr.HTTPStatus >= 500 {
                logger.Error("server error",
                    "code", appErr.Code,
                    "message", appErr.Message,
                    "error", appErr.Err,
                    "path", c.Request.URL.Path,
                )
            }

            c.JSON(appErr.HTTPStatus, ErrorResponse{
                Code:    appErr.Code,
                Message: appErr.Message,
                Details: appErr.Details,
            })
            return
        }

        // 未知错误
        logger.Error("unexpected error",
            "error", err,
            "path", c.Request.URL.Path,
        )

        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Code:    "INTERNAL_ERROR",
            Message: "An unexpected error occurred",
        })
    }
}
```

### 使用错误中间件

```go
func main() {
    logger := slog.Default()

    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.ErrorHandler(logger))

    r.GET("/users/:id", func(c *gin.Context) {
        user, err := userService.GetByID(c.Param("id"))
        if err != nil {
            c.Error(err)  // 添加错误到 context
            return
        }
        c.JSON(200, user)
    })

    r.Run(":8080")
}
```

## 服务层错误

### 定义服务错误

```go
// internal/services/errors.go
package services

import (
    "myapp/pkg/apperror"
)

var (
    ErrUserNotFound = apperror.NotFound("User")
    ErrUserExists   = apperror.AlreadyExists("User")
    ErrInvalidEmail = apperror.InvalidInput("email", "invalid format")
)

// 或更详细的错误
func UserNotFoundError(id string) *apperror.AppError {
    return apperror.New(
        "USER_NOT_FOUND",
        "User with ID "+id+" not found",
        404,
    )
}
```

### 服务层使用

```go
// internal/services/user.go
package services

import (
    "context"
    "errors"

    "myapp/internal/models"
    "myapp/internal/repositories"
    "myapp/pkg/apperror"
)

type UserService struct {
    repo *repositories.UserRepository
}

func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, repositories.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, apperror.Wrap(err, "DATABASE_ERROR", "Failed to fetch user", 500)
    }
    return user, nil
}

func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*models.User, error) {
    // 检查邮箱是否已存在
    exists, err := s.repo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, apperror.Wrap(err, "DATABASE_ERROR", "Failed to check email", 500)
    }
    if exists {
        return nil, ErrUserExists
    }

    user := &models.User{
        Name:  req.Name,
        Email: req.Email,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, apperror.Wrap(err, "DATABASE_ERROR", "Failed to create user", 500)
    }

    return user, nil
}
```

## Handler 错误处理

### 模式一：直接响应

```go
func (h *UserHandler) GetByID(c *gin.Context) {
    user, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
    if err != nil {
        var appErr *apperror.AppError
        if errors.As(err, &appErr) {
            c.JSON(appErr.HTTPStatus, gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
            })
            return
        }
        c.JSON(500, gin.H{"code": "INTERNAL_ERROR", "message": "Internal error"})
        return
    }
    c.JSON(200, gin.H{"data": user})
}
```

### 模式二：使用 c.Error()

```go
func (h *UserHandler) GetByID(c *gin.Context) {
    user, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
    if err != nil {
        c.Error(err)
        return
    }
    c.JSON(200, gin.H{"data": user})
}
```

### 模式三：辅助函数

```go
// pkg/response/response.go
package response

import (
    "errors"
    "net/http"

    "myapp/pkg/apperror"

    "github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data any) {
    c.JSON(http.StatusOK, gin.H{"data": data})
}

func Created(c *gin.Context, data any) {
    c.JSON(http.StatusCreated, gin.H{"data": data})
}

func NoContent(c *gin.Context) {
    c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, err error) {
    var appErr *apperror.AppError
    if errors.As(err, &appErr) {
        c.JSON(appErr.HTTPStatus, gin.H{
            "code":    appErr.Code,
            "message": appErr.Message,
            "details": appErr.Details,
        })
        return
    }

    c.JSON(http.StatusInternalServerError, gin.H{
        "code":    "INTERNAL_ERROR",
        "message": "Internal server error",
    })
}

// 使用
func (h *UserHandler) GetByID(c *gin.Context) {
    user, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, user)
}
```

## 验证错误处理

```go
import (
    "github.com/go-playground/validator/v10"
)

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func HandleValidationError(err error) *apperror.AppError {
    var validationErrors validator.ValidationErrors
    if errors.As(err, &validationErrors) {
        details := make([]ValidationError, 0, len(validationErrors))
        for _, e := range validationErrors {
            details = append(details, ValidationError{
                Field:   toSnakeCase(e.Field()),
                Message: getValidationMessage(e),
            })
        }
        return apperror.ErrValidation.WithDetails(details)
    }
    return apperror.ErrBadRequest
}

func getValidationMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return "This field is required"
    case "email":
        return "Invalid email format"
    case "min":
        return fmt.Sprintf("Minimum length is %s", e.Param())
    case "max":
        return fmt.Sprintf("Maximum length is %s", e.Param())
    default:
        return fmt.Sprintf("Failed %s validation", e.Tag())
    }
}

// 在 handler 中使用
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, HandleValidationError(err))
        return
    }
    // ...
}
```

## Panic 恢复

### 自定义 Recovery

```go
func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                // 获取堆栈信息
                stack := debug.Stack()

                logger.Error("panic recovered",
                    "error", r,
                    "stack", string(stack),
                    "path", c.Request.URL.Path,
                    "method", c.Request.Method,
                )

                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "code":    "INTERNAL_ERROR",
                    "message": "Internal server error",
                })
            }
        }()
        c.Next()
    }
}
```

## 错误日志

### 结构化错误日志

```go
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) == 0 {
            return
        }

        err := c.Errors.Last().Err
        requestID := c.GetString("RequestID")

        var appErr *apperror.AppError
        if errors.As(err, &appErr) {
            logLevel := slog.LevelInfo
            if appErr.HTTPStatus >= 500 {
                logLevel = slog.LevelError
            } else if appErr.HTTPStatus >= 400 {
                logLevel = slog.LevelWarn
            }

            logger.Log(c.Request.Context(), logLevel, "request error",
                "request_id", requestID,
                "code", appErr.Code,
                "message", appErr.Message,
                "status", appErr.HTTPStatus,
                "path", c.Request.URL.Path,
                "method", c.Request.Method,
                "client_ip", c.ClientIP(),
                "underlying_error", appErr.Err,
            )

            c.JSON(appErr.HTTPStatus, gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
                "details": appErr.Details,
            })
            return
        }

        // 未处理的错误
        logger.Error("unhandled error",
            "request_id", requestID,
            "error", err,
            "path", c.Request.URL.Path,
        )

        c.JSON(500, gin.H{
            "code":    "INTERNAL_ERROR",
            "message": "Internal server error",
        })
    }
}
```

## 完整示例

```go
package main

import (
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "runtime/debug"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

// AppError 应用错误
type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    HTTPStatus int    `json:"-"`
    Details    any    `json:"details,omitempty"`
    Err        error  `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// 预定义错误
var (
    ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "Resource not found", HTTPStatus: 404}
    ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", HTTPStatus: 401}
    ErrForbidden    = &AppError{Code: "FORBIDDEN", Message: "Forbidden", HTTPStatus: 403}
    ErrValidation   = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed", HTTPStatus: 400}
    ErrInternal     = &AppError{Code: "INTERNAL_ERROR", Message: "Internal error", HTTPStatus: 500}
)

func NotFoundError(resource string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: resource + " not found", HTTPStatus: 404}
}

// 验证错误结构
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// Recovery 中间件
func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                logger.Error("panic", "error", r, "stack", string(debug.Stack()))
                c.AbortWithStatusJSON(500, gin.H{"code": "INTERNAL_ERROR", "message": "Internal error"})
            }
        }()
        c.Next()
    }
}

// 错误处理中间件
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) == 0 {
            return
        }

        err := c.Errors.Last().Err

        var appErr *AppError
        if errors.As(err, &appErr) {
            if appErr.HTTPStatus >= 500 {
                logger.Error("server error", "code", appErr.Code, "error", appErr.Err)
            }
            c.JSON(appErr.HTTPStatus, gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
                "details": appErr.Details,
            })
            return
        }

        logger.Error("unexpected error", "error", err)
        c.JSON(500, gin.H{"code": "INTERNAL_ERROR", "message": "Internal error"})
    }
}

// 验证错误处理
func handleValidationError(err error) *AppError {
    var validationErrors validator.ValidationErrors
    if errors.As(err, &validationErrors) {
        details := make([]ValidationError, 0)
        for _, e := range validationErrors {
            details = append(details, ValidationError{
                Field:   toSnakeCase(e.Field()),
                Message: getValidationMessage(e),
            })
        }
        return &AppError{
            Code:       ErrValidation.Code,
            Message:    ErrValidation.Message,
            HTTPStatus: ErrValidation.HTTPStatus,
            Details:    details,
        }
    }
    return ErrValidation
}

func getValidationMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return "This field is required"
    case "email":
        return "Invalid email format"
    case "min":
        return fmt.Sprintf("Minimum is %s", e.Param())
    case "max":
        return fmt.Sprintf("Maximum is %s", e.Param())
    default:
        return "Invalid value"
    }
}

func toSnakeCase(s string) string {
    var result []rune
    for i, r := range s {
        if i > 0 && r >= 'A' && r <= 'Z' {
            result = append(result, '_')
        }
        result = append(result, r)
    }
    return strings.ToLower(string(result))
}

// 模拟用户数据
var users = map[string]User{
    "1": {ID: "1", Name: "Alice", Email: "alice@example.com"},
}

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type CreateUserRequest struct {
    Name  string `json:"name" binding:"required,min=2"`
    Email string `json:"email" binding:"required,email"`
}

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    r := gin.New()
    r.Use(RecoveryMiddleware(logger))
    r.Use(ErrorHandler(logger))

    // 获取用户
    r.GET("/users/:id", func(c *gin.Context) {
        id := c.Param("id")
        user, ok := users[id]
        if !ok {
            c.Error(NotFoundError("User"))
            return
        }
        c.JSON(200, gin.H{"data": user})
    })

    // 创建用户
    r.POST("/users", func(c *gin.Context) {
        var req CreateUserRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.Error(handleValidationError(err))
            return
        }

        // 检查邮箱是否存在
        for _, u := range users {
            if u.Email == req.Email {
                c.Error(&AppError{
                    Code:       "EMAIL_EXISTS",
                    Message:    "Email already exists",
                    HTTPStatus: 409,
                })
                return
            }
        }

        user := User{ID: fmt.Sprintf("%d", len(users)+1), Name: req.Name, Email: req.Email}
        users[user.ID] = user
        c.JSON(201, gin.H{"data": user})
    })

    // 测试 panic
    r.GET("/panic", func(c *gin.Context) {
        panic("test panic")
    })

    r.Run(":8080")
}
```

## 错误响应示例

```json
// 404 Not Found
{
  "code": "NOT_FOUND",
  "message": "User not found"
}

// 400 Validation Error
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": [
    {"field": "email", "message": "Invalid email format"},
    {"field": "name", "message": "Minimum is 2"}
  ]
}

// 401 Unauthorized
{
  "code": "UNAUTHORIZED",
  "message": "Authentication required"
}

// 500 Internal Error
{
  "code": "INTERNAL_ERROR",
  "message": "Internal server error"
}
```

## 总结

| 错误类型 | HTTP 状态码 | 使用场景 |
|---------|------------|---------|
| BAD_REQUEST | 400 | 请求格式错误 |
| VALIDATION_ERROR | 400 | 验证失败 |
| UNAUTHORIZED | 401 | 未认证 |
| FORBIDDEN | 403 | 无权限 |
| NOT_FOUND | 404 | 资源不存在 |
| CONFLICT | 409 | 资源冲突 |
| INTERNAL_ERROR | 500 | 服务器错误 |

**下一节**：[文件上传](./09-file-upload.md) - 学习文件处理
