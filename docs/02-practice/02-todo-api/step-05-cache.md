# Step 5: Redis 缓存

## 目标

添加 Redis 缓存层，提升 API 性能。

## 5.1 缓存服务

创建 `internal/cache/todo_cache.go`:

```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/user/todo-api/internal/models"
)

// TodoCache 待办缓存
type TodoCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewTodoCache 创建待办缓存
func NewTodoCache(client *redis.Client, ttl time.Duration) *TodoCache {
	return &TodoCache{
		client: client,
		ttl:    ttl,
	}
}

// 缓存键前缀
const (
	todoKeyPrefix     = "todo:"
	todoListKeyPrefix = "todo_list:"
)

// todoKey 生成待办缓存键
func todoKey(id uint) string {
	return fmt.Sprintf("%s%d", todoKeyPrefix, id)
}

// todoListKey 生成待办列表缓存键
func todoListKey(userID uint, page, pageSize int, completed *bool) string {
	completedStr := "all"
	if completed != nil {
		if *completed {
			completedStr = "completed"
		} else {
			completedStr = "pending"
		}
	}
	return fmt.Sprintf("%s%d:%s:%d:%d", todoListKeyPrefix, userID, completedStr, page, pageSize)
}

// Get 获取单个待办缓存
func (c *TodoCache) Get(ctx context.Context, id uint) (*models.Todo, error) {
	if c.client == nil {
		return nil, nil // Redis 未连接，跳过缓存
	}

	data, err := c.client.Get(ctx, todoKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}

	var todo models.Todo
	if err := json.Unmarshal(data, &todo); err != nil {
		return nil, err
	}

	return &todo, nil
}

// Set 设置单个待办缓存
func (c *TodoCache) Set(ctx context.Context, todo *models.Todo) error {
	if c.client == nil {
		return nil
	}

	data, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, todoKey(todo.ID), data, c.ttl).Err()
}

// Delete 删除单个待办缓存
func (c *TodoCache) Delete(ctx context.Context, id uint) error {
	if c.client == nil {
		return nil
	}

	return c.client.Del(ctx, todoKey(id)).Err()
}

// GetList 获取待办列表缓存
func (c *TodoCache) GetList(ctx context.Context, userID uint, page, pageSize int, completed *bool) ([]*models.Todo, int64, error) {
	if c.client == nil {
		return nil, 0, nil
	}

	key := todoListKey(userID, page, pageSize, completed)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	var result struct {
		Todos []*models.Todo `json:"todos"`
		Total int64          `json:"total"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, 0, err
	}

	return result.Todos, result.Total, nil
}

// SetList 设置待办列表缓存
func (c *TodoCache) SetList(ctx context.Context, userID uint, page, pageSize int, completed *bool, todos []*models.Todo, total int64) error {
	if c.client == nil {
		return nil
	}

	key := todoListKey(userID, page, pageSize, completed)

	result := struct {
		Todos []*models.Todo `json:"todos"`
		Total int64          `json:"total"`
	}{
		Todos: todos,
		Total: total,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// 列表缓存使用较短的 TTL
	return c.client.Set(ctx, key, data, c.ttl/2).Err()
}

// InvalidateUserCache 使用户的所有缓存失效
func (c *TodoCache) InvalidateUserCache(ctx context.Context, userID uint) error {
	if c.client == nil {
		return nil
	}

	// 删除所有匹配的列表缓存
	pattern := fmt.Sprintf("%s%d:*", todoListKeyPrefix, userID)
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}

	return iter.Err()
}
```

## 5.2 带缓存的待办服务

更新 `internal/services/todo.go`:

```go
package services

import (
	"context"
	"errors"
	"time"

	"github.com/user/todo-api/internal/cache"
	"github.com/user/todo-api/internal/models"
	"github.com/user/todo-api/internal/repositories"
)

var ErrNotFound = errors.New("not found")

// ListOptions 列表查询选项
type ListOptions struct {
	Page      int
	PageSize  int
	Completed *bool
	Search    string
}

// TodoService 待办服务
type TodoService struct {
	repo  *repositories.TodoRepository
	cache *cache.TodoCache
}

// NewTodoService 创建待办服务
func NewTodoService(repo *repositories.TodoRepository) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

// NewTodoServiceWithCache 创建带缓存的待办服务
func NewTodoServiceWithCache(repo *repositories.TodoRepository, cache *cache.TodoCache) *TodoService {
	return &TodoService{
		repo:  repo,
		cache: cache,
	}
}

// Create 创建待办
func (s *TodoService) Create(ctx context.Context, todo *models.Todo) error {
	if err := s.repo.Create(ctx, todo); err != nil {
		return err
	}

	// 使用户的列表缓存失效
	if s.cache != nil {
		s.cache.InvalidateUserCache(ctx, todo.UserID)
	}

	return nil
}

// GetByID 根据 ID 获取待办
func (s *TodoService) GetByID(ctx context.Context, id, userID uint) (*models.Todo, error) {
	// 先查缓存
	if s.cache != nil {
		if todo, err := s.cache.Get(ctx, id); err == nil && todo != nil {
			// 验证用户权限
			if todo.UserID == userID {
				return todo, nil
			}
		}
	}

	// 查数据库
	todo, err := s.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrTodoNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// 写入缓存
	if s.cache != nil {
		s.cache.Set(ctx, todo)
	}

	return todo, nil
}

// List 获取待办列表
func (s *TodoService) List(ctx context.Context, userID uint, opts ListOptions) ([]*models.Todo, int64, error) {
	// 如果有搜索条件，直接查数据库（不缓存搜索结果）
	if opts.Search != "" {
		return s.repo.FindByUserID(ctx, userID, repositories.ListOptions{
			Page:      opts.Page,
			PageSize:  opts.PageSize,
			Completed: opts.Completed,
			Search:    opts.Search,
		})
	}

	// 先查缓存
	if s.cache != nil {
		todos, total, err := s.cache.GetList(ctx, userID, opts.Page, opts.PageSize, opts.Completed)
		if err == nil && todos != nil {
			return todos, total, nil
		}
	}

	// 查数据库
	todos, total, err := s.repo.FindByUserID(ctx, userID, repositories.ListOptions{
		Page:      opts.Page,
		PageSize:  opts.PageSize,
		Completed: opts.Completed,
		Search:    opts.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存
	if s.cache != nil {
		s.cache.SetList(ctx, userID, opts.Page, opts.PageSize, opts.Completed, todos, total)
	}

	return todos, total, nil
}

// Update 更新待办
func (s *TodoService) Update(ctx context.Context, todo *models.Todo) error {
	if err := s.repo.Update(ctx, todo); err != nil {
		return err
	}

	// 更新缓存
	if s.cache != nil {
		s.cache.Set(ctx, todo)
		s.cache.InvalidateUserCache(ctx, todo.UserID)
	}

	return nil
}

// Delete 删除待办
func (s *TodoService) Delete(ctx context.Context, id, userID uint) error {
	// 先获取待办（验证权限）
	todo, err := s.repo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrTodoNotFound) {
			return ErrNotFound
		}
		return err
	}

	// 删除
	if err := s.repo.DeleteByIDAndUserID(ctx, id, userID); err != nil {
		return err
	}

	// 删除缓存
	if s.cache != nil {
		s.cache.Delete(ctx, id)
		s.cache.InvalidateUserCache(ctx, todo.UserID)
	}

	return nil
}

// Complete 完成待办
func (s *TodoService) Complete(ctx context.Context, id, userID uint) (*models.Todo, error) {
	todo, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	todo.Completed = true
	todo.CompletedAt = &now

	if err := s.Update(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}
```

## 5.3 更新主程序

更新 `cmd/api/main.go` 中的依赖初始化：

```go
// 初始化缓存
var todoCache *cache.TodoCache
if cache.Client != nil {
	todoCache = cache.NewTodoCache(cache.Client, 5*time.Minute)
}

// 初始化服务（带缓存）
var todoService *services.TodoService
if todoCache != nil {
	todoService = services.NewTodoServiceWithCache(todoRepo, todoCache)
} else {
	todoService = services.NewTodoService(todoRepo)
}
```

## 5.4 缓存策略

### Cache-Aside 模式

```
┌────────┐     ┌───────┐     ┌──────────┐
│ Client │────▶│ Cache │────▶│ Database │
└────────┘     └───────┘     └──────────┘
     │              │
     │    命中      │    未命中
     ◀──────────────┘         │
                              │
     ◀────────────────────────┘
           查询后写入缓存
```

### 缓存失效策略

1. **时间失效**: TTL 自动过期
2. **主动失效**: 数据变更时删除相关缓存
3. **列表缓存**: 任何变更都使列表缓存失效

## 5.5 测试缓存效果

```bash
# 第一次请求（缓存未命中，查数据库）
time curl http://localhost:8080/api/todos -H "Authorization: Bearer $TOKEN"

# 第二次请求（缓存命中，更快）
time curl http://localhost:8080/api/todos -H "Authorization: Bearer $TOKEN"

# 查看 Redis 中的缓存
redis-cli keys "todo*"
redis-cli get "todo_list:1:all:1:10"
```

## 项目总结

完成 Todo API 项目后，你已经掌握：

| 技能 | 说明 |
|------|------|
| Gin 框架 | 路由、中间件、请求处理 |
| GORM | 模型定义、CRUD、关联 |
| JWT 认证 | 令牌生成、验证、中间件 |
| Redis 缓存 | Cache-Aside 模式、缓存失效 |
| 分层架构 | Handler → Service → Repository |

## 下一步

恭喜完成 Todo API 项目！现在可以继续：

**[项目 3: Auth Service](../03-auth-service/README.md)** - 构建完整的认证授权服务！
