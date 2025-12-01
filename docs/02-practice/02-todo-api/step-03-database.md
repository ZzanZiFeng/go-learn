# Step 3: 数据库集成

## 目标

实现 Repository 层，处理数据库操作。

## 3.1 用户 Repository

创建 `internal/repositories/user.go`:

```go
package repositories

import (
	"context"
	"errors"

	"github.com/user/todo-api/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
)

// UserRepository 用户仓储
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		// 检查唯一约束冲突
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return result.Error
	}
	return nil
}

// FindByID 根据 ID 查找用户
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// ExistsByEmail 检查邮箱是否已存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// ExistsByUsername 检查用户名是否已存在
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
```

## 3.2 待办 Repository

创建 `internal/repositories/todo.go`:

```go
package repositories

import (
	"context"
	"errors"

	"github.com/user/todo-api/internal/models"
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
```

## 3.3 数据库迁移

创建 `migrations/001_init.sql`:

```sql
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 待办表
CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 待办表索引
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
CREATE INDEX IF NOT EXISTS idx_todos_deleted_at ON todos(deleted_at);
CREATE INDEX IF NOT EXISTS idx_todos_user_completed ON todos(user_id, completed);

-- 更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_todos_updated_at
    BEFORE UPDATE ON todos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

## 3.4 使用迁移

```bash
# 使用 golang-migrate
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/todo_api?sslmode=disable" up

# 或者使用 GORM AutoMigrate（已在 main.go 中配置）
# db.AutoMigrate(&models.User{}, &models.Todo{})
```

## 检查点

完成本步骤后：

- [x] 实现了 UserRepository
- [x] 实现了 TodoRepository
- [x] 创建了数据库迁移脚本
- [x] 理解了 Repository 模式

## 下一步

[Step 4: JWT 认证](./step-04-auth.md) - 实现用户认证。
