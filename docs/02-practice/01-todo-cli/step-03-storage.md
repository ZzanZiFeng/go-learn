# Step 3: 数据持久化

## 目标

实现 JSON 文件存储，让待办事项数据能够持久保存。

## 3.1 创建存储接口

创建 `internal/storage/storage.go`:

```go
package storage

import "github.com/user/todo-cli/internal/todo"

// Storage 定义存储接口
type Storage interface {
	// Load 从存储加载待办列表
	Load() ([]*todo.Todo, error)
	// Save 保存待办列表到存储
	Save(todos []*todo.Todo) error
}
```

## 3.2 实现 JSON 存储

创建 `internal/storage/json.go`:

```go
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/user/todo-cli/internal/todo"
)

// JSONStorage 使用 JSON 文件存储待办事项
type JSONStorage struct {
	filepath string
}

// NewJSONStorage 创建 JSON 存储实例
func NewJSONStorage(filepath string) *JSONStorage {
	return &JSONStorage{
		filepath: filepath,
	}
}

// DefaultJSONStorage 使用默认路径创建存储
// 默认存储在用户主目录下的 .todos.json
func DefaultJSONStorage() (*JSONStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	filepath := filepath.Join(homeDir, ".todos.json")
	return NewJSONStorage(filepath), nil
}

// Load 从 JSON 文件加载待办列表
func (s *JSONStorage) Load() ([]*todo.Todo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(s.filepath); errors.Is(err, os.ErrNotExist) {
		// 文件不存在，返回空列表
		return []*todo.Todo{}, nil
	}

	// 读取文件
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, err
	}

	// 空文件
	if len(data) == 0 {
		return []*todo.Todo{}, nil
	}

	// 解析 JSON
	var todos []*todo.Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}

// Save 保存待办列表到 JSON 文件
func (s *JSONStorage) Save(todos []*todo.Todo) error {
	// 序列化为 JSON（带缩进，便于阅读）
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(s.filepath, data, 0644)
}

// GetFilePath 返回存储文件路径
func (s *JSONStorage) GetFilePath() string {
	return s.filepath
}
```

### 代码解析

**文件操作**:
```go
// 读取文件
data, err := os.ReadFile(s.filepath)

// 写入文件 (权限 0644: 所有者读写，其他人只读)
os.WriteFile(s.filepath, data, 0644)

// 检查文件是否存在
if _, err := os.Stat(s.filepath); errors.Is(err, os.ErrNotExist) {
    // 文件不存在
}
```

**JSON 序列化**:
```go
// 解析 JSON
json.Unmarshal(data, &todos)

// 序列化为 JSON（带缩进）
json.MarshalIndent(todos, "", "  ")
```

## 3.3 添加存储单元测试

创建 `internal/storage/json_test.go`:

```go
package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/todo-cli/internal/todo"
)

func TestJSONStorage(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_todos.json")
	storage := NewJSONStorage(testFile)

	// 测试保存
	todos := []*todo.Todo{
		{
			ID:          1,
			Description: "Test todo 1",
			Completed:   false,
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			Description: "Test todo 2",
			Completed:   true,
			CreatedAt:   time.Now(),
			CompletedAt: time.Now(),
		},
	}

	err := storage.Save(todos)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Save() did not create file")
	}

	// 测试加载
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != len(todos) {
		t.Errorf("Load() returned %d items, want %d", len(loaded), len(todos))
	}

	// 验证数据
	if loaded[0].Description != todos[0].Description {
		t.Errorf("Load() first todo description = %s, want %s",
			loaded[0].Description, todos[0].Description)
	}

	if loaded[1].Completed != todos[1].Completed {
		t.Errorf("Load() second todo completed = %v, want %v",
			loaded[1].Completed, todos[1].Completed)
	}
}

func TestJSONStorage_EmptyFile(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty_todos.json")
	storage := NewJSONStorage(testFile)

	// 加载不存在的文件应该返回空列表
	todos, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("Load() returned %d items, want 0", len(todos))
	}
}

func TestJSONStorage_SaveAndLoadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty_save.json")
	storage := NewJSONStorage(testFile)

	// 保存空列表
	err := storage.Save([]*todo.Todo{})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 加载空列表
	todos, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("Load() returned %d items, want 0", len(todos))
	}
}
```

运行测试:

```bash
go test ./internal/storage/... -v
```

## 3.4 创建应用程序入口

创建 `internal/app/app.go`，整合 List 和 Storage:

```go
package app

import (
	"github.com/user/todo-cli/internal/storage"
	"github.com/user/todo-cli/internal/todo"
)

// App 代表应用程序
type App struct {
	List    *todo.List
	Storage storage.Storage
}

// New 创建应用实例
func New(s storage.Storage) *App {
	return &App{
		List:    todo.NewList(),
		Storage: s,
	}
}

// Load 从存储加载数据
func (a *App) Load() error {
	todos, err := a.Storage.Load()
	if err != nil {
		return err
	}
	a.List.SetItems(todos)
	return nil
}

// Save 保存数据到存储
func (a *App) Save() error {
	return a.Storage.Save(a.List.All())
}

// Add 添加待办并保存
func (a *App) Add(description string) (*todo.Todo, error) {
	t, err := a.List.Add(description)
	if err != nil {
		return nil, err
	}

	if err := a.Save(); err != nil {
		return nil, err
	}

	return t, nil
}

// Complete 完成待办并保存
func (a *App) Complete(id int) error {
	if err := a.List.Complete(id); err != nil {
		return err
	}
	return a.Save()
}

// Delete 删除待办并保存
func (a *App) Delete(id int) error {
	if err := a.List.Delete(id); err != nil {
		return err
	}
	return a.Save()
}

// Clear 清除已完成的待办并保存
func (a *App) Clear() (int, error) {
	count := a.List.Clear()
	if err := a.Save(); err != nil {
		return 0, err
	}
	return count, nil
}
```

## 3.5 更新主程序

更新 `cmd/todo/main.go`:

```go
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/user/todo-cli/internal/app"
	"github.com/user/todo-cli/internal/storage"
	"github.com/user/todo-cli/internal/ui"
)

var (
	application *app.App
	printer     *ui.Printer
)

func main() {
	// 初始化存储
	store, err := storage.DefaultJSONStorage()
	if err != nil {
		fmt.Printf("初始化存储失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化应用
	application = app.New(store)
	printer = ui.NewPrinter(true)

	// 加载数据
	if err := application.Load(); err != nil {
		printer.PrintError("加载数据失败: %v", err)
		os.Exit(1)
	}

	// 检查参数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "add":
		handleAdd()
	case "list", "ls":
		handleList()
	case "complete", "done":
		handleComplete()
	case "delete", "rm":
		handleDelete()
	case "clear":
		handleClear()
	case "stats":
		handleStats()
	case "help":
		printUsage()
	default:
		printer.PrintError("未知命令: %s", command)
		printUsage()
		os.Exit(1)
	}
}

func handleAdd() {
	if len(os.Args) < 3 {
		printer.PrintError("请提供待办事项描述")
		fmt.Println("用法: todo add <description>")
		os.Exit(1)
	}

	description := os.Args[2]
	for i := 3; i < len(os.Args); i++ {
		description += " " + os.Args[i]
	}

	t, err := application.Add(description)
	if err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	printer.PrintSuccess("已添加: \"%s\" (ID: %d)", t.Description, t.ID)
}

func handleList() {
	showAll := false
	if len(os.Args) > 2 && (os.Args[2] == "--all" || os.Args[2] == "-a") {
		showAll = true
	}

	printer.PrintTodoList(application.List.All(), showAll)
}

func handleComplete() {
	if len(os.Args) < 3 {
		printer.PrintError("请提供待办事项 ID")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printer.PrintError("无效的 ID: %s", os.Args[2])
		os.Exit(1)
	}

	t, err := application.List.Get(id)
	if err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	if err := application.Complete(id); err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	printer.PrintSuccess("已完成: \"%s\"", t.Description)
}

func handleDelete() {
	if len(os.Args) < 3 {
		printer.PrintError("请提供待办事项 ID")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printer.PrintError("无效的 ID: %s", os.Args[2])
		os.Exit(1)
	}

	t, err := application.List.Get(id)
	if err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	description := t.Description

	if err := application.Delete(id); err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	printer.PrintSuccess("已删除: \"%s\"", description)
}

func handleClear() {
	count, err := application.Clear()
	if err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	if count == 0 {
		printer.PrintInfo("没有已完成的待办事项")
	} else {
		printer.PrintSuccess("已清除 %d 个已完成的待办事项", count)
	}
}

func handleStats() {
	printer.PrintStats(application.List)
}

func printUsage() {
	usage := `
Todo CLI - 命令行待办事项管理工具

用法:
  todo <command> [arguments]

命令:
  add <description>    添加一个新的待办事项
  list, ls [--all]     列出待办事项（默认只显示未完成）
  complete, done <id>  标记待办事项为完成
  delete, rm <id>      删除待办事项
  clear                清除所有已完成的待办事项
  stats                显示统计信息
  help                 显示帮助信息

选项:
  --all, -a            显示所有待办事项（包括已完成）

示例:
  todo add "学习 Go 语言"
  todo list
  todo list --all
  todo complete 1
  todo delete 1
  todo clear
  todo stats
`
	fmt.Println(usage)
}
```

## 3.6 测试数据持久化

```bash
# 编译
go build -o todo cmd/todo/main.go

# 添加一些待办
./todo add "学习 Go 语言"
./todo add "完成 Todo CLI"
./todo add "开始 Todo API"

# 查看列表
./todo list

# 完成一个
./todo complete 1

# 再次查看（数据应该被保存）
./todo list --all

# 查看存储文件
cat ~/.todos.json

# 重启终端后再运行，数据应该还在
./todo list
```

## 3.7 JSON 文件格式

存储的 JSON 文件格式:

```json
[
  {
    "id": 1,
    "description": "学习 Go 语言",
    "completed": true,
    "created_at": "2024-01-01T10:00:00Z",
    "completed_at": "2024-01-01T12:00:00Z"
  },
  {
    "id": 2,
    "description": "完成 Todo CLI",
    "completed": false,
    "created_at": "2024-01-01T10:05:00Z"
  }
]
```

## 检查点

完成本步骤后，你应该：

- [x] 理解 Go 的文件读写操作
- [x] 掌握 JSON 序列化和反序列化
- [x] 理解接口的使用（Storage 接口）
- [x] 数据能够持久化保存

## 下一步

[Step 4: 完善与优化](./step-04-polish.md) - 添加搜索功能、统计功能，完善用户体验。
