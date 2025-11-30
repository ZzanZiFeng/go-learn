# Swagger API 文档 (Swagger/OpenAPI)

## 概述

Swagger/OpenAPI 是 API 文档的标准规范。Go 中常用 `swaggo/swag` 自动生成 API 文档。

## 安装

```bash
# 安装 swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# 安装 Gin 的 Swagger 中间件
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

## 基础配置

### 项目结构

```
myapp/
├── cmd/
│   └── main.go         # 主入口
├── docs/               # swag 生成的文档
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   └── handlers/
│       └── user.go     # 带注释的 handlers
└── go.mod
```

### 主文件注释

```go
// cmd/main.go
package main

import (
    "myapp/docs"
    "myapp/internal/handlers"

    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           My API
// @version         1.0
// @description     This is a sample API server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
    r := gin.Default()

    // 程序化设置 host（可选）
    docs.SwaggerInfo.Host = "localhost:8080"

    // Swagger 路由
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API 路由
    v1 := r.Group("/api/v1")
    handlers.SetupRoutes(v1)

    r.Run(":8080")
}
```

## API 注释

### 基本结构

```go
// @Summary      简短描述
// @Description  详细描述
// @Tags         分组标签
// @Accept       json
// @Produce      json
// @Param        参数名 参数位置 类型 是否必须 "描述"
// @Success      200 {object} ResponseType
// @Failure      400 {object} ErrorResponse
// @Router       /path [method]
```

### GET 请求示例

```go
// internal/handlers/user.go

// User 用户模型
type User struct {
    ID        uint      `json:"id" example:"1"`
    Name      string    `json:"name" example:"Alice"`
    Email     string    `json:"email" example:"alice@example.com"`
    CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
    Data  []User `json:"data"`
    Total int64  `json:"total" example:"100"`
    Page  int    `json:"page" example:"1"`
    Size  int    `json:"size" example:"10"`
}

// ListUsers godoc
// @Summary      获取用户列表
// @Description  获取分页的用户列表
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        page   query    int     false  "页码"     default(1)
// @Param        size   query    int     false  "每页数量" default(10) minimum(1) maximum(100)
// @Param        search query    string  false  "搜索关键词"
// @Success      200    {object} UserListResponse
// @Failure      400    {object} ErrorResponse
// @Failure      500    {object} ErrorResponse
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
    // 实现...
}

// GetUser godoc
// @Summary      获取用户详情
// @Description  根据 ID 获取单个用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户 ID"
// @Success      200  {object}  User
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
    // 实现...
}
```

### POST 请求示例

```go
// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100" example:"Alice"`
    Email    string `json:"email" binding:"required,email" example:"alice@example.com"`
    Password string `json:"password" binding:"required,min=8" example:"password123"`
    Age      int    `json:"age" binding:"omitempty,gte=0,lte=150" example:"25"`
}

// CreateUser godoc
// @Summary      创建用户
// @Description  创建新用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body CreateUserRequest true "用户信息"
// @Success      201 {object} User
// @Failure      400 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse "邮箱已存在"
// @Failure      500 {object} ErrorResponse
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 实现...
}
```

### PUT/PATCH 请求示例

```go
// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
    Name  string `json:"name" binding:"omitempty,min=2,max=100" example:"Bob"`
    Email string `json:"email" binding:"omitempty,email" example:"bob@example.com"`
    Age   *int   `json:"age" binding:"omitempty,gte=0,lte=150" example:"30"`
}

// UpdateUser godoc
// @Summary      更新用户
// @Description  更新用户信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id      path     int               true "用户 ID"
// @Param        request body     UpdateUserRequest true "更新内容"
// @Success      200     {object} User
// @Failure      400     {object} ErrorResponse
// @Failure      404     {object} ErrorResponse
// @Failure      500     {object} ErrorResponse
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
    // 实现...
}
```

### DELETE 请求示例

```go
// DeleteUser godoc
// @Summary      删除用户
// @Description  删除指定用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户 ID"
// @Success      204  "No Content"
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
    // 实现...
}
```

## 参数类型

### 路径参数

```go
// @Param id path int true "用户 ID"
// @Param slug path string true "文章 slug"
```

### 查询参数

```go
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10) minimum(1) maximum(100)
// @Param sort query string false "排序字段" Enums(created_at, name, email)
// @Param order query string false "排序方向" Enums(asc, desc) default(desc)
// @Param tags query []string false "标签列表" collectionFormat(csv)
```

### 请求体

```go
// @Param request body CreateUserRequest true "请求体"
```

### Header 参数

```go
// @Param Authorization header string true "Bearer token"
// @Param X-Request-ID header string false "请求 ID"
```

### 表单参数

```go
// @Param file formData file true "上传文件"
// @Param name formData string true "文件名"
```

## 认证

### Bearer Token

```go
// 在 main.go 定义
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// 在需要认证的 API 上使用
// @Security Bearer

// GetProfile godoc
// @Summary      获取当前用户信息
// @Description  获取已登录用户的个人信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} User
// @Failure      401 {object} ErrorResponse
// @Router       /users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
    // 实现...
}
```

### API Key

```go
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// @Security ApiKeyAuth
```

### Basic Auth

```go
// @securityDefinitions.basic BasicAuth

// @Security BasicAuth
```

### OAuth2

```go
// @securityDefinitions.oauth2.implicit OAuth2Implicit
// @authorizationUrl https://example.com/oauth/authorize
// @scope.read Grants read access
// @scope.write Grants write access
```

## 响应模型

### 通用响应

```go
// Response 通用响应
type Response struct {
    Code    int         `json:"code" example:"0"`
    Message string      `json:"message" example:"success"`
    Data    interface{} `json:"data"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Code    string `json:"code" example:"VALIDATION_ERROR"`
    Message string `json:"message" example:"Validation failed"`
    Details any    `json:"details,omitempty"`
}

// ValidationError 验证错误详情
type ValidationError struct {
    Field   string `json:"field" example:"email"`
    Message string `json:"message" example:"Invalid email format"`
}
```

### 分页响应

```go
// PaginatedResponse 分页响应
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
    Total      int64 `json:"total" example:"100"`
    Page       int   `json:"page" example:"1"`
    Size       int   `json:"size" example:"10"`
    TotalPages int   `json:"total_pages" example:"10"`
}
```

## 生成文档

### 命令

```bash
# 在项目根目录执行
swag init

# 指定入口文件
swag init -g cmd/main.go

# 指定输出目录
swag init -o ./api/docs

# 解析依赖包中的注释
swag init --parseDependency --parseInternal

# 格式化注释
swag fmt
```

### 生成的文件

```
docs/
├── docs.go      # Go 代码
├── swagger.json # JSON 格式
└── swagger.yaml # YAML 格式
```

## 完整示例

```go
// cmd/main.go
package main

import (
    _ "myapp/docs"
    "myapp/internal/handlers"

    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           User Management API
// @version         1.0
// @description     A user management service API in Go using Gin framework.

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"

func main() {
    r := gin.Default()

    // Swagger
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API v1
    v1 := r.Group("/api/v1")
    {
        userHandler := handlers.NewUserHandler()
        users := v1.Group("/users")
        {
            users.GET("", userHandler.ListUsers)
            users.POST("", userHandler.CreateUser)
            users.GET("/:id", userHandler.GetUser)
            users.PUT("/:id", userHandler.UpdateUser)
            users.DELETE("/:id", userHandler.DeleteUser)
        }
    }

    r.Run(":8080")
}
```

```go
// internal/handlers/user.go
package handlers

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
    return &UserHandler{}
}

// 模型定义
type User struct {
    ID        uint      `json:"id" example:"1"`
    Name      string    `json:"name" example:"Alice"`
    Email     string    `json:"email" example:"alice@example.com"`
    Age       int       `json:"age" example:"25"`
    CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
    UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100" example:"Alice"`
    Email    string `json:"email" binding:"required,email" example:"alice@example.com"`
    Password string `json:"password" binding:"required,min=8" example:"password123"`
    Age      int    `json:"age" binding:"omitempty,gte=0,lte=150" example:"25"`
}

type UpdateUserRequest struct {
    Name  string `json:"name" binding:"omitempty,min=2,max=100" example:"Bob"`
    Email string `json:"email" binding:"omitempty,email" example:"bob@example.com"`
    Age   *int   `json:"age" binding:"omitempty,gte=0,lte=150" example:"30"`
}

type UserListResponse struct {
    Data       []User     `json:"data"`
    Pagination Pagination `json:"pagination"`
}

type Pagination struct {
    Total      int64 `json:"total" example:"100"`
    Page       int   `json:"page" example:"1"`
    Size       int   `json:"size" example:"10"`
    TotalPages int   `json:"total_pages" example:"10"`
}

type ErrorResponse struct {
    Code    string `json:"code" example:"VALIDATION_ERROR"`
    Message string `json:"message" example:"Validation failed"`
}

// ListUsers godoc
// @Summary      获取用户列表
// @Description  获取分页的用户列表，支持搜索
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        page   query    int     false  "页码"     default(1)  minimum(1)
// @Param        size   query    int     false  "每页数量" default(10) minimum(1) maximum(100)
// @Param        search query    string  false  "搜索关键词（姓名或邮箱）"
// @Success      200    {object} UserListResponse
// @Failure      400    {object} ErrorResponse
// @Failure      500    {object} ErrorResponse
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

    users := []User{
        {ID: 1, Name: "Alice", Email: "alice@example.com", Age: 25},
        {ID: 2, Name: "Bob", Email: "bob@example.com", Age: 30},
    }

    c.JSON(http.StatusOK, UserListResponse{
        Data: users,
        Pagination: Pagination{
            Total:      100,
            Page:       page,
            Size:       size,
            TotalPages: 10,
        },
    })
}

// GetUser godoc
// @Summary      获取用户详情
// @Description  根据 ID 获取单个用户的详细信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户 ID" minimum(1)
// @Success      200  {object}  User
// @Failure      400  {object}  ErrorResponse "无效的 ID"
// @Failure      404  {object}  ErrorResponse "用户不存在"
// @Failure      500  {object}  ErrorResponse
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code: "INVALID_ID", Message: "Invalid user ID",
        })
        return
    }

    user := User{ID: uint(id), Name: "Alice", Email: "alice@example.com"}
    c.JSON(http.StatusOK, user)
}

// CreateUser godoc
// @Summary      创建用户
// @Description  创建新用户账号
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body CreateUserRequest true "用户信息"
// @Success      201 {object} User
// @Failure      400 {object} ErrorResponse "请求参数错误"
// @Failure      409 {object} ErrorResponse "邮箱已存在"
// @Failure      500 {object} ErrorResponse
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code: "VALIDATION_ERROR", Message: err.Error(),
        })
        return
    }

    user := User{
        ID:        1,
        Name:      req.Name,
        Email:     req.Email,
        Age:       req.Age,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    c.JSON(http.StatusCreated, user)
}

// UpdateUser godoc
// @Summary      更新用户
// @Description  更新指定用户的信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id      path     int               true "用户 ID" minimum(1)
// @Param        request body     UpdateUserRequest true "更新内容"
// @Success      200     {object} User
// @Failure      400     {object} ErrorResponse
// @Failure      401     {object} ErrorResponse "未授权"
// @Failure      404     {object} ErrorResponse "用户不存在"
// @Failure      500     {object} ErrorResponse
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code: "VALIDATION_ERROR", Message: err.Error(),
        })
        return
    }

    user := User{ID: 1, Name: req.Name, Email: req.Email}
    c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary      删除用户
// @Description  删除指定用户账号
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "用户 ID" minimum(1)
// @Success      204  "删除成功"
// @Failure      401  {object}  ErrorResponse "未授权"
// @Failure      404  {object}  ErrorResponse "用户不存在"
// @Failure      500  {object}  ErrorResponse
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
    c.Status(http.StatusNoContent)
}
```

## Swagger UI 配置

```go
import (
    ginSwagger "github.com/swaggo/gin-swagger"
    swaggerFiles "github.com/swaggo/files"
)

// 默认配置
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 自定义配置
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
    ginSwagger.URL("/swagger/doc.json"),          // 文档 URL
    ginSwagger.DefaultModelsExpandDepth(-1),      // 隐藏模型
    ginSwagger.PersistAuthorization(true),        // 保持认证状态
    ginSwagger.DocExpansion("list"),              // 展开方式: list, full, none
))
```

## 注释速查

| 注释 | 描述 |
|------|------|
| `@Summary` | 简短描述 |
| `@Description` | 详细描述 |
| `@Tags` | 分组标签 |
| `@Accept` | 接受的内容类型 |
| `@Produce` | 返回的内容类型 |
| `@Param` | 参数定义 |
| `@Success` | 成功响应 |
| `@Failure` | 失败响应 |
| `@Router` | 路由路径和方法 |
| `@Security` | 认证方式 |
| `@Deprecated` | 标记废弃 |

**下一节**：[练习](./exercises.md) - 实践练习
