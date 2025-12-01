package services

import (
	"context"
	"errors"
	"time"

	"github.com/your-username/go-learn/projects/todo-api/internal/cache"
	"github.com/your-username/go-learn/projects/todo-api/internal/models"
	"github.com/your-username/go-learn/projects/todo-api/internal/repositories"
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
func NewTodoServiceWithCache(repo *repositories.TodoRepository, todoCache *cache.TodoCache) *TodoService {
	return &TodoService{
		repo:  repo,
		cache: todoCache,
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
