package repositories

import (
	"context"
	"errors"

	"github.com/your-username/go-learn/projects/todo-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrTodoNotFound = errors.New("待办不存在")
)

// TodoRepository 待办仓储
type TodoRepository struct {
	db *gorm.DB
}

// NewTodoRepository 创建待办仓储
func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create 创建待办
func (r *TodoRepository) Create(ctx context.Context, todo *models.Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

// FindByID 根据 ID 查找待办
func (r *TodoRepository) FindByID(ctx context.Context, id uint) (*models.Todo, error) {
	var todo models.Todo
	result := r.db.WithContext(ctx).First(&todo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTodoNotFound
		}
		return nil, result.Error
	}
	return &todo, nil
}

// FindByIDAndUserID 根据 ID 和用户 ID 查找待办
func (r *TodoRepository) FindByIDAndUserID(ctx context.Context, id, userID uint) (*models.Todo, error) {
	var todo models.Todo
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&todo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTodoNotFound
		}
		return nil, result.Error
	}
	return &todo, nil
}

// ListOptions 列表查询选项
type ListOptions struct {
	Page      int
	PageSize  int
	Completed *bool
	Search    string
}

// FindByUserID 根据用户 ID 查找待办列表
func (r *TodoRepository) FindByUserID(ctx context.Context, userID uint, opts ListOptions) ([]*models.Todo, int64, error) {
	var todos []*models.Todo
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Todo{}).Where("user_id = ?", userID)

	// 筛选条件
	if opts.Completed != nil {
		query = query.Where("completed = ?", *opts.Completed)
	}

	if opts.Search != "" {
		search := "%" + opts.Search + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", search, search)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (opts.Page - 1) * opts.PageSize
	result := query.Order("created_at DESC").Offset(offset).Limit(opts.PageSize).Find(&todos)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return todos, total, nil
}

// Update 更新待办
func (r *TodoRepository) Update(ctx context.Context, todo *models.Todo) error {
	return r.db.WithContext(ctx).Save(todo).Error
}

// Delete 删除待办
func (r *TodoRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Todo{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

// DeleteByIDAndUserID 根据 ID 和用户 ID 删除待办
func (r *TodoRepository) DeleteByIDAndUserID(ctx context.Context, id, userID uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.Todo{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

// CountByUserID 统计用户的待办数量
func (r *TodoRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Todo{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountCompletedByUserID 统计用户已完成的待办数量
func (r *TodoRepository) CountCompletedByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Todo{}).
		Where("user_id = ? AND completed = ?", userID, true).
		Count(&count).Error
	return count, err
}
