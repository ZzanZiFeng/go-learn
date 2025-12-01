# Step 2: 路由与处理器

## 目标

实现完整的 CRUD 处理器和请求验证。

## 2.1 数据模型

创建 `internal/models/user.go`:

```go
package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password  string         `json:"-" gorm:"not null"` // json:"-" 不输出密码
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
```

创建 `internal/models/todo.go`:

```go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Todo 待办事项模型
type Todo struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	UserID      uint           `json:"user_id" gorm:"index;not null"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"size:1000"`
	Completed   bool           `json:"completed" gorm:"default:false"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Todo) TableName() string {
	return "todos"
}
```

## 2.2 请求/响应 DTO

创建 `internal/handlers/dto.go`:

```go
package handlers

// ==================== 认证 DTO ====================

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token string      `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ==================== 待办 DTO ====================

// CreateTodoRequest 创建待办请求
type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=1000"`
}

// UpdateTodoRequest 更新待办请求
type UpdateTodoRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Completed   *bool   `json:"completed"`
}

// TodoResponse 待办响应
type TodoResponse struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Completed   bool    `json:"completed"`
	CompletedAt *string `json:"completed_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// TodoListResponse 待办列表响应
type TodoListResponse struct {
	Todos    []TodoResponse `json:"todos"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ListQuery 列表查询参数
type ListQuery struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=10"`
	Completed *bool  `form:"completed"`
	Search    string `form:"search"`
}

// ==================== 通用响应 ====================

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	Message string `json:"message"`
}
```

## 2.3 待办处理器

创建 `internal/handlers/todo.go`:

```go
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/todo-api/internal/models"
	"github.com/user/todo-api/internal/services"
)

// TodoHandler 待办处理器
type TodoHandler struct {
	service *services.TodoService
}

// NewTodoHandler 创建待办处理器
func NewTodoHandler(service *services.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

// Create 创建待办
// @Summary 创建待办
// @Tags todos
// @Accept json
// @Produce json
// @Param request body CreateTodoRequest true "创建请求"
// @Success 201 {object} TodoResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/todos [post]
func (h *TodoHandler) Create(c *gin.Context) {
	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求参数错误",
			Details: err.Error(),
		})
		return
	}

	// 从上下文获取用户 ID
	userID := c.GetUint("user_id")

	todo := &models.Todo{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.service.Create(c.Request.Context(), todo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "创建失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, toTodoResponse(todo))
}

// Get 获取单个待办
func (h *TodoHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的 ID"})
		return
	}

	userID := c.GetUint("user_id")

	todo, err := h.service.GetByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "待办不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toTodoResponse(todo))
}

// List 获取待办列表
func (h *TodoHandler) List(c *gin.Context) {
	var query ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "查询参数错误"})
		return
	}

	// 限制 page_size
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	userID := c.GetUint("user_id")

	todos, total, err := h.service.List(c.Request.Context(), userID, services.ListOptions{
		Page:      query.Page,
		PageSize:  query.PageSize,
		Completed: query.Completed,
		Search:    query.Search,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	response := TodoListResponse{
		Todos:    make([]TodoResponse, len(todos)),
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	for i, todo := range todos {
		response.Todos[i] = toTodoResponse(todo)
	}

	c.JSON(http.StatusOK, response)
}

// Update 更新待办
func (h *TodoHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的 ID"})
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求参数错误",
			Details: err.Error(),
		})
		return
	}

	userID := c.GetUint("user_id")

	todo, err := h.service.GetByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "待办不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// 更新字段
	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
		if *req.Completed && todo.CompletedAt == nil {
			now := time.Now()
			todo.CompletedAt = &now
		} else if !*req.Completed {
			todo.CompletedAt = nil
		}
	}

	if err := h.service.Update(c.Request.Context(), todo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toTodoResponse(todo))
}

// Delete 删除待办
func (h *TodoHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的 ID"})
		return
	}

	userID := c.GetUint("user_id")

	if err := h.service.Delete(c.Request.Context(), uint(id), userID); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "待办不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "删除成功"})
}

// Complete 完成待办
func (h *TodoHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的 ID"})
		return
	}

	userID := c.GetUint("user_id")

	todo, err := h.service.Complete(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "待办不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toTodoResponse(todo))
}

// toTodoResponse 转换为响应格式
func toTodoResponse(todo *models.Todo) TodoResponse {
	resp := TodoResponse{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339),
	}

	if todo.CompletedAt != nil {
		formatted := todo.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &formatted
	}

	return resp
}
```

## 2.4 测试 API

```bash
# 启动服务
go run cmd/api/main.go

# 注册用户
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

# 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.token')

# 创建待办
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"学习 Go","description":"完成 Todo API 项目"}'

# 获取待办列表
curl http://localhost:8080/api/todos \
  -H "Authorization: Bearer $TOKEN"

# 完成待办
curl -X PATCH http://localhost:8080/api/todos/1/complete \
  -H "Authorization: Bearer $TOKEN"
```

## 检查点

完成本步骤后：

- [x] 定义了数据模型
- [x] 实现了请求/响应 DTO
- [x] 实现了 CRUD 处理器
- [x] 添加了请求验证

## 下一步

[Step 3: 数据库集成](./step-03-database.md) - 实现 Repository 层和数据库操作。
