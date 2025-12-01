# Step 4: RBAC 权限

## 目标

实现基于角色的访问控制（RBAC）。

## 4.1 角色仓储

创建 `internal/repositories/role.go`:

```go
package repositories

import (
	"context"
	"errors"

	"github.com/your-username/go-learn/projects/auth-service/internal/models"
	"gorm.io/gorm"
)

var ErrRoleNotFound = errors.New("角色不存在")

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	result := r.db.WithContext(ctx).Preload("Permissions").First(&role, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	result := r.db.WithContext(ctx).Preload("Permissions").Where("name = ?", name).First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

func (r *RoleRepository) FindAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	result := r.db.WithContext(ctx).Preload("Permissions").Find(&roles)
	return roles, result.Error
}

func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *RoleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

func (r *RoleRepository) AssignPermission(ctx context.Context, roleID, permissionID uint) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		roleID, permissionID,
	).Error
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleID, permissionID uint) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?",
		roleID, permissionID,
	).Error
}

func (r *RoleRepository) FindAllPermissions(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	result := r.db.WithContext(ctx).Find(&permissions)
	return permissions, result.Error
}
```

## 4.2 RBAC 服务

创建 `internal/services/rbac.go`:

```go
package services

import (
	"context"

	"github.com/your-username/go-learn/projects/auth-service/internal/models"
	"github.com/your-username/go-learn/projects/auth-service/internal/repositories"
)

type RBACService struct {
	userRepo *repositories.UserRepository
	roleRepo *repositories.RoleRepository
}

func NewRBACService(userRepo *repositories.UserRepository, roleRepo *repositories.RoleRepository) *RBACService {
	return &RBACService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// HasRole 检查用户是否拥有指定角色
func (s *RBACService) HasRole(ctx context.Context, userID uint, roleName string) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range user.Roles {
		if role.Name == roleName {
			return true, nil
		}
	}
	return false, nil
}

// HasPermission 检查用户是否拥有指定权限
func (s *RBACService) HasPermission(ctx context.Context, userID uint, permissionName string) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Name == permissionName {
				return true, nil
			}
		}
	}
	return false, nil
}

// HasAnyRole 检查用户是否拥有任一角色
func (s *RBACService) HasAnyRole(ctx context.Context, userID uint, roleNames ...string) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	roleSet := make(map[string]bool)
	for _, name := range roleNames {
		roleSet[name] = true
	}

	for _, role := range user.Roles {
		if roleSet[role.Name] {
			return true, nil
		}
	}
	return false, nil
}

// HasAnyPermission 检查用户是否拥有任一权限
func (s *RBACService) HasAnyPermission(ctx context.Context, userID uint, permissionNames ...string) (bool, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	permSet := make(map[string]bool)
	for _, name := range permissionNames {
		permSet[name] = true
	}

	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if permSet[perm.Name] {
				return true, nil
			}
		}
	}
	return false, nil
}

// GetUserRoles 获取用户的所有角色
func (s *RBACService) GetUserRoles(ctx context.Context, userID uint) ([]models.Role, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.Roles, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]bool)
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permSet[perm.Name] = true
		}
	}

	permissions := make([]string, 0, len(permSet))
	for perm := range permSet {
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// AssignRoleToUser 为用户分配角色
func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID uint) error {
	return s.userRepo.AssignRole(ctx, userID, roleID)
}

// RemoveRoleFromUser 移除用户角色
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID uint) error {
	return s.userRepo.RemoveRole(ctx, userID, roleID)
}

// CreateRole 创建角色
func (s *RBACService) CreateRole(ctx context.Context, role *models.Role) error {
	return s.roleRepo.Create(ctx, role)
}

// GetAllRoles 获取所有角色
func (s *RBACService) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.FindAll(ctx)
}

// UpdateRole 更新角色
func (s *RBACService) UpdateRole(ctx context.Context, role *models.Role) error {
	return s.roleRepo.Update(ctx, role)
}

// DeleteRole 删除角色
func (s *RBACService) DeleteRole(ctx context.Context, roleID uint) error {
	return s.roleRepo.Delete(ctx, roleID)
}
```

## 4.3 RBAC 中间件

创建 `internal/middleware/rbac.go`:

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-username/go-learn/projects/auth-service/internal/services"
)

// RequireRole 要求用户拥有指定角色
func RequireRole(rbacService *services.RBACService, roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		hasRole, err := rbacService.HasRole(c.Request.Context(), userID, roleName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "权限检查失败",
			})
			return
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
				"code":  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}

// RequirePermission 要求用户拥有指定权限
func RequirePermission(rbacService *services.RBACService, permissionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		hasPermission, err := rbacService.HasPermission(c.Request.Context(), userID, permissionName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "权限检查失败",
			})
			return
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
				"code":  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}

// RequireAnyRole 要求用户拥有任一角色
func RequireAnyRole(rbacService *services.RBACService, roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		hasRole, err := rbacService.HasAnyRole(c.Request.Context(), userID, roleNames...)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "权限检查失败",
			})
			return
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
				"code":  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 要求用户拥有任一权限
func RequireAnyPermission(rbacService *services.RBACService, permissionNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		hasPermission, err := rbacService.HasAnyPermission(c.Request.Context(), userID, permissionNames...)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "权限检查失败",
			})
			return
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
				"code":  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}
```

## 4.4 角色处理器

创建 `internal/handlers/role.go`:

```go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-username/go-learn/projects/auth-service/internal/models"
	"github.com/your-username/go-learn/projects/auth-service/internal/services"
)

type RoleHandler struct {
	rbacService *services.RBACService
}

func NewRoleHandler(rbacService *services.RBACService) *RoleHandler {
	return &RoleHandler{rbacService: rbacService}
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=200"`
}

type RoleResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
}

type PermissionResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.rbacService.GetAllRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = toRoleResponse(role)
	}

	c.JSON(http.StatusOK, response)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.rbacService.CreateRole(c.Request.Context(), role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toRoleResponse(*role))
}

func (h *RoleHandler) AssignToUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	var req struct {
		RoleID uint `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.rbacService.AssignRoleToUser(c.Request.Context(), uint(userID), req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色分配成功"})
}

func toRoleResponse(role models.Role) RoleResponse {
	permissions := make([]PermissionResponse, len(role.Permissions))
	for i, perm := range role.Permissions {
		permissions[i] = PermissionResponse{
			ID:       perm.ID,
			Name:     perm.Name,
			Resource: perm.Resource,
			Action:   perm.Action,
		}
	}

	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
	}
}
```

## 4.5 路由配置

```go
// 在 main.go 中配置路由

// 角色管理（需要 admin 权限）
roles := api.Group("/roles")
roles.Use(middleware.Auth(tokenService))
roles.Use(middleware.RequireRole(rbacService, "admin"))
{
    roles.GET("", roleHandler.List)
    roles.POST("", roleHandler.Create)
    roles.PUT("/:id", roleHandler.Update)
    roles.DELETE("/:id", roleHandler.Delete)
}

// 用户角色分配（需要 users:write 权限）
api.POST("/users/:user_id/roles",
    middleware.Auth(tokenService),
    middleware.RequirePermission(rbacService, "users:write"),
    roleHandler.AssignToUser,
)
```

## 检查点

完成本步骤后：

- [x] 实现了角色仓储
- [x] 实现了 RBAC 服务
- [x] 实现了 RBAC 中间件
- [x] 实现了角色管理处理器

## 下一步

[Step 5: OAuth2 集成](./step-05-oauth.md) - 实现第三方登录。
