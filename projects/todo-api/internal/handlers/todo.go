package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-username/go-learn/projects/todo-api/internal/models"
	"github.com/your-username/go-learn/projects/todo-api/internal/services"
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
			Error:   "请求参数错误",
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
			Error:   "创建失败",
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
			Error:   "请求参数错误",
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
