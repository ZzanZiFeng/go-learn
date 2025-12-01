# Step 1: 项目初始化

## 目标

创建项目基础结构，定义 Todo 数据模型。

## 1.1 创建项目目录

```bash
# 在 projects 目录下创建
cd projects/todo-cli

# 确认目录结构
ls -la
# 应该已经存在基础结构：
# ├── cmd/
# ├── internal/
# └── go.mod
```

## 1.2 定义 Todo 模型

创建 `internal/todo/todo.go`:

```go
// Package todo 提供待办事项的核心数据结构和操作
package todo

import (
	"errors"
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
	ErrNotFound     = errors.New("todo not found")
	ErrEmptyDesc    = errors.New("description cannot be empty")
	ErrInvalidID    = errors.New("invalid todo id")
)

// NewTodo 创建一个新的待办事项
func NewTodo(id int, description string) (*Todo, error) {
	if description == "" {
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
```

### 代码解析

**结构体定义**:
```go
type Todo struct {
    ID          int       `json:"id"`           // JSON 序列化时的字段名
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    CompletedAt time.Time `json:"completed_at,omitempty"`  // omitempty: 零值时省略
}
```

**自定义错误**:
```go
var (
    ErrNotFound  = errors.New("todo not found")   // 包级变量
    ErrEmptyDesc = errors.New("description cannot be empty")
)
```

**方法定义**:
```go
// 值接收者 - 不修改原对象
func (t *Todo) IsCompleted() bool {
    return t.Completed
}

// 指针接收者 - 修改原对象
func (t *Todo) Complete() {
    t.Completed = true
    t.CompletedAt = time.Now()
}
```

## 1.3 创建 TodoList 管理器

在 `internal/todo/` 目录下继续添加 `list.go`:

```go
package todo

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
			// 删除切片元素
			l.items = append(l.items[:i], l.items[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
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
	// 更新 nextID
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
```

### 切片操作解析

**添加元素**:
```go
l.items = append(l.items, todo)
```

**删除元素**:
```go
// 删除索引 i 的元素
l.items = append(l.items[:i], l.items[i+1:]...)

// 图解:
// items = [A, B, C, D, E], 删除 C (index=2)
// items[:2] = [A, B]
// items[3:] = [D, E]
// append([A, B], [D, E]...) = [A, B, D, E]
```

**过滤元素**:
```go
// 创建新切片保存符合条件的元素
result := make([]*Todo, 0)
for _, item := range l.items {
    if !item.Completed {
        result = append(result, item)
    }
}
```

## 1.4 创建简单的主程序

创建 `cmd/todo/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/user/todo-cli/internal/todo"
)

func main() {
	// 创建待办列表
	list := todo.NewList()

	// 简单测试
	todo1, _ := list.Add("学习 Go 语言")
	fmt.Printf("添加: %s\n", todo1)

	todo2, _ := list.Add("完成 Todo CLI")
	fmt.Printf("添加: %s\n", todo2)

	// 列出所有
	fmt.Println("\n所有待办:")
	for _, t := range list.All() {
		fmt.Printf("  %d. %s\n", t.ID, t)
	}

	// 完成第一个
	list.Complete(1)
	fmt.Println("\n完成 ID=1 后:")
	for _, t := range list.All() {
		fmt.Printf("  %d. %s\n", t.ID, t)
	}

	// 统计
	fmt.Printf("\n统计: 总数=%d, 已完成=%d, 待完成=%d\n",
		list.Count(), list.CompletedCount(), list.PendingCount())
}
```

## 1.5 运行测试

```bash
# 在项目根目录运行
cd projects/todo-cli
go run cmd/todo/main.go

# 预期输出:
# 添加: [ ] 学习 Go 语言
# 添加: [ ] 完成 Todo CLI
#
# 所有待办:
#   1. [ ] 学习 Go 语言
#   2. [ ] 完成 Todo CLI
#
# 完成 ID=1 后:
#   1. [✓] 学习 Go 语言
#   2. [ ] 完成 Todo CLI
#
# 统计: 总数=2, 已完成=1, 待完成=1
```

## 1.6 添加单元测试

创建 `internal/todo/todo_test.go`:

```go
package todo

import (
	"testing"
)

func TestNewTodo(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		description string
		wantErr     error
	}{
		{
			name:        "valid todo",
			id:          1,
			description: "Test todo",
			wantErr:     nil,
		},
		{
			name:        "empty description",
			id:          1,
			description: "",
			wantErr:     ErrEmptyDesc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := NewTodo(tt.id, tt.description)

			if err != tt.wantErr {
				t.Errorf("NewTodo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if todo.ID != tt.id {
					t.Errorf("NewTodo() ID = %v, want %v", todo.ID, tt.id)
				}
				if todo.Description != tt.description {
					t.Errorf("NewTodo() Description = %v, want %v", todo.Description, tt.description)
				}
				if todo.Completed {
					t.Error("NewTodo() Completed should be false")
				}
			}
		})
	}
}

func TestTodoComplete(t *testing.T) {
	todo, _ := NewTodo(1, "Test")

	if todo.IsCompleted() {
		t.Error("new todo should not be completed")
	}

	todo.Complete()

	if !todo.IsCompleted() {
		t.Error("todo should be completed after Complete()")
	}

	if todo.CompletedAt.IsZero() {
		t.Error("CompletedAt should be set")
	}
}

func TestListAdd(t *testing.T) {
	list := NewList()

	todo, err := list.Add("First todo")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if todo.ID != 1 {
		t.Errorf("First todo ID = %v, want 1", todo.ID)
	}

	if list.Count() != 1 {
		t.Errorf("Count() = %v, want 1", list.Count())
	}

	// Add second
	todo2, _ := list.Add("Second todo")
	if todo2.ID != 2 {
		t.Errorf("Second todo ID = %v, want 2", todo2.ID)
	}
}

func TestListComplete(t *testing.T) {
	list := NewList()
	list.Add("Test todo")

	err := list.Complete(1)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	todo, _ := list.Get(1)
	if !todo.IsCompleted() {
		t.Error("todo should be completed")
	}
}

func TestListDelete(t *testing.T) {
	list := NewList()
	list.Add("First")
	list.Add("Second")

	err := list.Delete(1)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if list.Count() != 1 {
		t.Errorf("Count() = %v, want 1", list.Count())
	}

	_, err = list.Get(1)
	if err != ErrNotFound {
		t.Error("deleted todo should not be found")
	}
}

func TestListNotFound(t *testing.T) {
	list := NewList()

	_, err := list.Get(999)
	if err != ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}

	err = list.Complete(999)
	if err != ErrNotFound {
		t.Errorf("Complete() error = %v, want ErrNotFound", err)
	}

	err = list.Delete(999)
	if err != ErrNotFound {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}
```

运行测试:

```bash
go test ./internal/todo/... -v

# 预期输出:
# === RUN   TestNewTodo
# === RUN   TestNewTodo/valid_todo
# === RUN   TestNewTodo/empty_description
# --- PASS: TestNewTodo (0.00s)
#     --- PASS: TestNewTodo/valid_todo (0.00s)
#     --- PASS: TestNewTodo/empty_description (0.00s)
# ...
# PASS
```

## 检查点

完成本步骤后，你应该：

- [x] 理解 Go 项目的目录结构
- [x] 能够定义结构体和方法
- [x] 理解切片的基本操作
- [x] 能够编写简单的单元测试

## 下一步

[Step 2: CRUD 操作](./step-02-crud.md) - 实现命令行参数解析和完整的 CRUD 功能。
