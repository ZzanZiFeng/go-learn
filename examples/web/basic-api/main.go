// Package main demonstrates a complete REST API using Gin framework
// 完整的 Gin REST API 示例
package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =============================================================================
// Models
// =============================================================================

type User struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// =============================================================================
// Request/Response Types
// =============================================================================

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Email string `json:"email" binding:"required,email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
}

type ListUsersQuery struct {
	Page   int    `form:"page,default=1" binding:"gte=1"`
	Size   int    `form:"size,default=10" binding:"gte=1,lte=100"`
	Search string `form:"search"`
}

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type PagedResponse struct {
	Items      any   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalPages int   `json:"total_pages"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// =============================================================================
// In-Memory Store (for demo purposes)
// =============================================================================

type UserStore struct {
	mu     sync.RWMutex
	users  map[uint]*User
	nextID uint
}

func NewUserStore() *UserStore {
	store := &UserStore{
		users:  make(map[uint]*User),
		nextID: 1,
	}
	// Add sample data
	store.Create("Alice", "alice@example.com")
	store.Create("Bob", "bob@example.com")
	store.Create("Charlie", "charlie@example.com")
	return store
}

func (s *UserStore) Create(name, email string) *User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &User{
		ID:        s.nextID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.users[s.nextID] = user
	s.nextID++
	return user
}

func (s *UserStore) GetByID(id uint) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[id]
}

func (s *UserStore) List(page, size int, search string) ([]*User, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*User
	for _, u := range s.users {
		if search == "" ||
			strings.Contains(strings.ToLower(u.Name), strings.ToLower(search)) ||
			strings.Contains(strings.ToLower(u.Email), strings.ToLower(search)) {
			filtered = append(filtered, u)
		}
	}

	total := int64(len(filtered))
	start := (page - 1) * size
	end := start + size

	if start >= len(filtered) {
		return []*User{}, total
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total
}

func (s *UserStore) Update(id uint, name, email string) *User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := s.users[id]
	if user == nil {
		return nil
	}

	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}
	user.UpdatedAt = time.Now()
	return user
}

func (s *UserStore) Delete(id uint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return false
	}
	delete(s.users, id)
	return true
}

func (s *UserStore) ExistsByEmail(email string, excludeID uint) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Email == email && u.ID != excludeID {
			return true
		}
	}
	return false
}

// =============================================================================
// Middleware
// =============================================================================

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Info("request",
			"request_id", c.GetString("RequestID"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start),
			"ip", c.ClientIP(),
		)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"request_id", c.GetString("RequestID"),
					"error", r,
				)
				c.AbortWithStatusJSON(500, ErrorResponse{
					Code:    "INTERNAL_ERROR",
					Message: "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

// =============================================================================
// Response Helpers
// =============================================================================

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "created",
		Data:    data,
	})
}

func Paged(c *gin.Context, items any, total int64, page, size int) {
	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PagedResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			Size:       size,
			TotalPages: totalPages,
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    "BAD_REQUEST",
		Message: message,
	})
}

func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Code:    "NOT_FOUND",
		Message: resource + " not found",
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{
		Code:    "CONFLICT",
		Message: message,
	})
}

// =============================================================================
// Handlers
// =============================================================================

type UserHandler struct {
	store *UserStore
}

func NewUserHandler(store *UserStore) *UserHandler {
	return &UserHandler{store: store}
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query ListUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		BadRequest(c, err.Error())
		return
	}

	users, total := h.store.List(query.Page, query.Size, query.Search)
	Paged(c, users, total, query.Page, query.Size)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid user ID")
		return
	}

	user := h.store.GetByID(uint(id))
	if user == nil {
		NotFound(c, "User")
		return
	}

	Success(c, user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if h.store.ExistsByEmail(req.Email, 0) {
		Conflict(c, "Email already exists")
		return
	}

	user := h.store.Create(req.Name, req.Email)
	Created(c, user)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if req.Email != "" && h.store.ExistsByEmail(req.Email, uint(id)) {
		Conflict(c, "Email already exists")
		return
	}

	user := h.store.Update(uint(id), req.Name, req.Email)
	if user == nil {
		NotFound(c, "User")
		return
	}

	Success(c, user)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid user ID")
		return
	}

	if !h.store.Delete(uint(id)) {
		NotFound(c, "User")
		return
	}

	c.Status(http.StatusNoContent)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Setup store
	store := NewUserStore()

	// Setup handler
	userHandler := NewUserHandler(store)

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Global middleware
	r.Use(RequestIDMiddleware())
	r.Use(LoggingMiddleware(logger))
	r.Use(RecoveryMiddleware(logger))
	r.Use(CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// Start server
	logger.Info("Server starting", "port", 8080)
	if err := r.Run(":8080"); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
