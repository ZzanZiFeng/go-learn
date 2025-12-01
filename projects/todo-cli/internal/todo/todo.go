// Package todo 提供待办事项的核心数据结构和操作
package todo

import (
	"errors"
	"strings"
	"time"
)

// Todo 代表一个待办事项
type Todo struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// 定义错误
var (
	ErrNotFound  = errors.New("todo not found")
	ErrEmptyDesc = errors.New("description cannot be empty")
	ErrInvalidID = errors.New("invalid todo id")
)

// NewTodo 创建一个新的待办事项
func NewTodo(id int, description string) (*Todo, error) {
	if strings.TrimSpace(description) == "" {
		return nil, ErrEmptyDesc
	}

	return &Todo{
		ID:          id,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
	}, nil
}

// Complete 标记待办事项为完成
func (t *Todo) Complete() {
	t.Completed = true
	t.CompletedAt = time.Now()
}

// Uncomplete 取消完成状态
func (t *Todo) Uncomplete() {
	t.Completed = false
	t.CompletedAt = time.Time{}
}

// IsCompleted 检查是否已完成
func (t *Todo) IsCompleted() bool {
	return t.Completed
}

// String 返回待办事项的字符串表示
func (t *Todo) String() string {
	status := "[ ]"
	if t.Completed {
		status = "[✓]"
	}
	return status + " " + t.Description
}

// List 管理待办事项列表
type List struct {
	items  []*Todo
	nextID int
}

// NewList 创建一个新的待办事项列表
func NewList() *List {
	return &List{
		items:  make([]*Todo, 0),
		nextID: 1,
	}
}

// Add 添加一个新的待办事项
func (l *List) Add(description string) (*Todo, error) {
	todo, err := NewTodo(l.nextID, description)
	if err != nil {
		return nil, err
	}

	l.items = append(l.items, todo)
	l.nextID++
	return todo, nil
}

// Get 根据 ID 获取待办事项
func (l *List) Get(id int) (*Todo, error) {
	for _, item := range l.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, ErrNotFound
}

// Complete 标记指定 ID 的待办事项为完成
func (l *List) Complete(id int) error {
	todo, err := l.Get(id)
	if err != nil {
		return err
	}
	todo.Complete()
	return nil
}

// Delete 删除指定 ID 的待办事项
func (l *List) Delete(id int) error {
	for i, item := range l.items {
		if item.ID == id {
			l.items = append(l.items[:i], l.items[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// Update 更新待办事项描述
func (l *List) Update(id int, newDescription string) error {
	if strings.TrimSpace(newDescription) == "" {
		return ErrEmptyDesc
	}

	todo, err := l.Get(id)
	if err != nil {
		return err
	}

	todo.Description = newDescription
	return nil
}

// All 返回所有待办事项
func (l *List) All() []*Todo {
	return l.items
}

// Pending 返回未完成的待办事项
func (l *List) Pending() []*Todo {
	result := make([]*Todo, 0)
	for _, item := range l.items {
		if !item.Completed {
			result = append(result, item)
		}
	}
	return result
}

// Completed 返回已完成的待办事项
func (l *List) Completed() []*Todo {
	result := make([]*Todo, 0)
	for _, item := range l.items {
		if item.Completed {
			result = append(result, item)
		}
	}
	return result
}

// Search 根据关键词搜索待办事项
func (l *List) Search(keyword string) []*Todo {
	if keyword == "" {
		return l.items
	}

	keyword = strings.ToLower(keyword)
	result := make([]*Todo, 0)

	for _, item := range l.items {
		if strings.Contains(strings.ToLower(item.Description), keyword) {
			result = append(result, item)
		}
	}

	return result
}

// Count 返回待办事项总数
func (l *List) Count() int {
	return len(l.items)
}

// CompletedCount 返回已完成数量
func (l *List) CompletedCount() int {
	count := 0
	for _, item := range l.items {
		if item.Completed {
			count++
		}
	}
	return count
}

// PendingCount 返回未完成数量
func (l *List) PendingCount() int {
	return l.Count() - l.CompletedCount()
}

// SetItems 设置待办事项列表（用于从存储加载）
func (l *List) SetItems(items []*Todo) {
	l.items = items
	maxID := 0
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	l.nextID = maxID + 1
}

// Clear 清除所有已完成的待办事项
func (l *List) Clear() int {
	cleared := 0
	newItems := make([]*Todo, 0)
	for _, item := range l.items {
		if !item.Completed {
			newItems = append(newItems, item)
		} else {
			cleared++
		}
	}
	l.items = newItems
	return cleared
}
