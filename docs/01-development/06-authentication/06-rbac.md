# RBAC (Role-Based Access Control)

## 概述

RBAC 是一种基于角色的访问控制模型，通过角色来管理用户权限。

## 核心概念

```
用户 (User)
    └── 角色 (Role)
            └── 权限 (Permission)
                    └── 资源:操作 (resource:action)

示例:
用户 Alice
    └── 角色: admin
            └── 权限: user:read, user:write, user:delete
                      post:read, post:write, post:delete
                      *:* (所有权限)
```

## 基础实现

### 数据模型

```go
package models

// User 用户
type User struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"-"`
    RoleID   uint   `json:"role_id"`
    Role     *Role  `json:"role,omitempty"`
}

// Role 角色
type Role struct {
    ID          uint         `json:"id"`
    Name        string       `json:"name"`        // admin, user, moderator
    Description string       `json:"description"`
    Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

// Permission 权限
type Permission struct {
    ID          uint   `json:"id"`
    Name        string `json:"name"`        // user:read
    Description string `json:"description"`
    Resource    string `json:"resource"`    // user
    Action      string `json:"action"`      // read, write, delete
}
```

### RBAC 服务

```go
package auth

import (
    "context"
    "sync"

    "myapp/internal/models"
    "myapp/internal/repositories"
)

type RBACService struct {
    roleRepo       *repositories.RoleRepository
    permissionRepo *repositories.PermissionRepository
    cache          *sync.Map // 简单缓存
}

func NewRBACService(roleRepo *repositories.RoleRepository, permissionRepo *repositories.PermissionRepository) *RBACService {
    return &RBACService{
        roleRepo:       roleRepo,
        permissionRepo: permissionRepo,
        cache:          &sync.Map{},
    }
}

// HasPermission 检查用户是否有权限
func (s *RBACService) HasPermission(ctx context.Context, userID uint, resource, action string) bool {
    // 获取用户角色的所有权限
    permissions, err := s.GetUserPermissions(ctx, userID)
    if err != nil {
        return false
    }

    required := resource + ":" + action

    for _, perm := range permissions {
        // 检查通配符权限
        if perm == "*:*" {
            return true
        }
        if perm == resource+":*" {
            return true
        }
        if perm == "*:"+action {
            return true
        }
        if perm == required {
            return true
        }
    }

    return false
}

// HasRole 检查用户是否有角色
func (s *RBACService) HasRole(ctx context.Context, userID uint, roleName string) bool {
    role, err := s.GetUserRole(ctx, userID)
    if err != nil {
        return false
    }
    return role.Name == roleName
}

// HasAnyRole 检查用户是否有任一角色
func (s *RBACService) HasAnyRole(ctx context.Context, userID uint, roleNames ...string) bool {
    role, err := s.GetUserRole(ctx, userID)
    if err != nil {
        return false
    }
    for _, name := range roleNames {
        if role.Name == name {
            return true
        }
    }
    return false
}

// GetUserPermissions 获取用户所有权限
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
    // 检查缓存
    cacheKey := fmt.Sprintf("perms:%d", userID)
    if cached, ok := s.cache.Load(cacheKey); ok {
        return cached.([]string), nil
    }

    // 从数据库获取
    role, err := s.GetUserRole(ctx, userID)
    if err != nil {
        return nil, err
    }

    permissions := make([]string, len(role.Permissions))
    for i, perm := range role.Permissions {
        permissions[i] = perm.Resource + ":" + perm.Action
    }

    // 存入缓存
    s.cache.Store(cacheKey, permissions)

    return permissions, nil
}

// GetUserRole 获取用户角色
func (s *RBACService) GetUserRole(ctx context.Context, userID uint) (*models.Role, error) {
    return s.roleRepo.GetByUserID(ctx, userID)
}

// InvalidateCache 清除缓存
func (s *RBACService) InvalidateCache(userID uint) {
    s.cache.Delete(fmt.Sprintf("perms:%d", userID))
}
```

## Gin 中间件

### 权限检查中间件

```go
package middleware

import (
    "net/http"

    "myapp/internal/auth"

    "github.com/gin-gonic/gin"
)

// RequirePermission 权限检查中间件
func RequirePermission(rbac *auth.RBACService, resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")
        if userID == 0 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Not authenticated",
            })
            return
        }

        if !rbac.HasPermission(c.Request.Context(), userID, resource, action) {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error":    "Permission denied",
                "required": resource + ":" + action,
            })
            return
        }

        c.Next()
    }
}

// RequireRole 角色检查中间件
func RequireRole(rbac *auth.RBACService, roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")
        if userID == 0 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Not authenticated",
            })
            return
        }

        if !rbac.HasAnyRole(c.Request.Context(), userID, roles...) {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error":    "Role required",
                "required": roles,
            })
            return
        }

        c.Next()
    }
}

// RequireOwnerOrAdmin 资源所有者或管理员
func RequireOwnerOrAdmin(rbac *auth.RBACService, getOwnerID func(c *gin.Context) uint) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("userID")
        if userID == 0 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Not authenticated",
            })
            return
        }

        ownerID := getOwnerID(c)

        // 是所有者或管理员
        if userID == ownerID || rbac.HasRole(c.Request.Context(), userID, "admin") {
            c.Next()
            return
        }

        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
            "error": "Permission denied",
        })
    }
}
```

### 使用示例

```go
func SetupRoutes(r *gin.Engine, rbac *auth.RBACService) {
    // 认证中间件
    auth := r.Group("/api")
    auth.Use(AuthMiddleware())

    // 用户管理 - 需要特定权限
    users := auth.Group("/users")
    {
        users.GET("", RequirePermission(rbac, "user", "read"), listUsers)
        users.POST("", RequirePermission(rbac, "user", "create"), createUser)
        users.GET("/:id", RequirePermission(rbac, "user", "read"), getUser)
        users.PUT("/:id", RequirePermission(rbac, "user", "update"), updateUser)
        users.DELETE("/:id", RequirePermission(rbac, "user", "delete"), deleteUser)
    }

    // 管理员路由 - 需要管理员角色
    admin := auth.Group("/admin")
    admin.Use(RequireRole(rbac, "admin"))
    {
        admin.GET("/stats", getStats)
        admin.GET("/logs", getLogs)
    }

    // 文章管理 - 所有者或管理员
    posts := auth.Group("/posts")
    {
        posts.GET("", listPosts)
        posts.POST("", createPost)
        posts.GET("/:id", getPost)
        posts.PUT("/:id", RequireOwnerOrAdmin(rbac, getPostOwnerID), updatePost)
        posts.DELETE("/:id", RequireOwnerOrAdmin(rbac, getPostOwnerID), deletePost)
    }
}

func getPostOwnerID(c *gin.Context) uint {
    postID := c.Param("id")
    // 从数据库获取文章的所有者 ID
    // post, _ := postService.GetByID(postID)
    // return post.AuthorID
    return 0
}
```

## 权限管理 API

```go
package handlers

type RBACHandler struct {
    rbacService *auth.RBACService
    roleRepo    *repositories.RoleRepository
}

// ListRoles 获取所有角色
func (h *RBACHandler) ListRoles(c *gin.Context) {
    roles, err := h.roleRepo.List(c.Request.Context())
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to list roles"})
        return
    }
    c.JSON(200, roles)
}

// CreateRole 创建角色
func (h *RBACHandler) CreateRole(c *gin.Context) {
    var req struct {
        Name        string   `json:"name" binding:"required"`
        Description string   `json:"description"`
        Permissions []string `json:"permissions"` // ["user:read", "user:write"]
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    role, err := h.roleRepo.Create(c.Request.Context(), req.Name, req.Description, req.Permissions)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create role"})
        return
    }

    c.JSON(201, role)
}

// AssignRole 分配角色给用户
func (h *RBACHandler) AssignRole(c *gin.Context) {
    userID := c.Param("userId")
    var req struct {
        RoleID uint `json:"role_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := h.roleRepo.AssignToUser(c.Request.Context(), userID, req.RoleID); err != nil {
        c.JSON(500, gin.H{"error": "Failed to assign role"})
        return
    }

    // 清除缓存
    h.rbacService.InvalidateCache(userID)

    c.JSON(200, gin.H{"message": "Role assigned"})
}

// AddPermissionToRole 添加权限到角色
func (h *RBACHandler) AddPermissionToRole(c *gin.Context) {
    roleID := c.Param("roleId")
    var req struct {
        Permission string `json:"permission" binding:"required"` // "user:write"
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := h.roleRepo.AddPermission(c.Request.Context(), roleID, req.Permission); err != nil {
        c.JSON(500, gin.H{"error": "Failed to add permission"})
        return
    }

    c.JSON(200, gin.H{"message": "Permission added"})
}
```

## 初始化权限数据

```go
package seeds

func SeedRolesAndPermissions(db *gorm.DB) error {
    // 创建权限
    permissions := []models.Permission{
        {Name: "user:read", Resource: "user", Action: "read"},
        {Name: "user:create", Resource: "user", Action: "create"},
        {Name: "user:update", Resource: "user", Action: "update"},
        {Name: "user:delete", Resource: "user", Action: "delete"},
        {Name: "post:read", Resource: "post", Action: "read"},
        {Name: "post:create", Resource: "post", Action: "create"},
        {Name: "post:update", Resource: "post", Action: "update"},
        {Name: "post:delete", Resource: "post", Action: "delete"},
    }

    for _, perm := range permissions {
        db.FirstOrCreate(&perm, models.Permission{Name: perm.Name})
    }

    // 创建角色
    roles := map[string][]string{
        "admin": {
            "user:read", "user:create", "user:update", "user:delete",
            "post:read", "post:create", "post:update", "post:delete",
        },
        "moderator": {
            "user:read",
            "post:read", "post:update", "post:delete",
        },
        "user": {
            "user:read",
            "post:read", "post:create",
        },
        "guest": {
            "post:read",
        },
    }

    for roleName, permNames := range roles {
        role := models.Role{Name: roleName}
        db.FirstOrCreate(&role, models.Role{Name: roleName})

        var perms []models.Permission
        db.Where("name IN ?", permNames).Find(&perms)
        db.Model(&role).Association("Permissions").Replace(perms)
    }

    return nil
}
```

## Casbin 集成

对于复杂的权限需求，可以使用 Casbin：

```bash
go get github.com/casbin/casbin/v2
go get github.com/casbin/gorm-adapter/v3
```

```go
package auth

import (
    "github.com/casbin/casbin/v2"
    gormadapter "github.com/casbin/gorm-adapter/v3"
    "gorm.io/gorm"
)

func NewCasbinEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
    adapter, err := gormadapter.NewAdapterByDB(db)
    if err != nil {
        return nil, err
    }

    enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
    if err != nil {
        return nil, err
    }

    // 加载策略
    enforcer.LoadPolicy()

    return enforcer, nil
}

// config/rbac_model.conf
// [request_definition]
// r = sub, obj, act
//
// [policy_definition]
// p = sub, obj, act
//
// [role_definition]
// g = _, _
//
// [policy_effect]
// e = some(where (p.eft == allow))
//
// [matchers]
// m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act

// 使用
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        path := c.Request.URL.Path
        method := c.Request.Method

        allowed, _ := enforcer.Enforce(userID, path, method)
        if !allowed {
            c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
            return
        }

        c.Next()
    }
}
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 最小权限原则 | 只授予必要的权限 |
| 使用角色 | 避免直接给用户分配权限 |
| 缓存权限 | 减少数据库查询 |
| 审计日志 | 记录权限变更 |
| 定期审查 | 检查不合理的权限分配 |
| 分离关注点 | 认证和授权分开处理 |

**下一节**：[练习](./exercises.md) - 实践练习
