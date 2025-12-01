package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/your-username/go-learn/projects/todo-api/internal/models"
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
