# Step 4: 完善与优化

## 目标

添加搜索、统计等高级功能，完善用户体验。

## 4.1 添加搜索功能

更新 `internal/todo/list.go`，添加搜索方法:

```go
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
```

## 4.2 更新主程序添加搜索命令

在 `cmd/todo/main.go` 添加搜索处理:

```go
func main() {
	// ... 前面的代码不变

	switch command {
	// ... 其他 case
	case "search":
		handleSearch()
	// ...
	}
}

func handleSearch() {
	if len(os.Args) < 3 {
		printer.PrintError("请提供搜索关键词")
		fmt.Println("用法: todo search <keyword>")
		os.Exit(1)
	}

	keyword := os.Args[2]
	results := application.List.Search(keyword)

	if len(results) == 0 {
		printer.PrintInfo("没有找到匹配 \"%s\" 的待办事项", keyword)
		return
	}

	fmt.Printf("\n🔍 搜索结果: \"%s\" (%d 项)\n", keyword, len(results))
	printer.PrintTodoList(results, true)
}
```

## 4.3 添加编辑功能

更新 `internal/todo/list.go`:

```go
// Update 更新待办事项描述
func (l *List) Update(id int, newDescription string) error {
	if newDescription == "" {
		return ErrEmptyDesc
	}

	todo, err := l.Get(id)
	if err != nil {
		return err
	}

	todo.Description = newDescription
	return nil
}
```

在 `internal/app/app.go` 添加:

```go
// Update 更新待办描述并保存
func (a *App) Update(id int, description string) error {
	if err := a.List.Update(id, description); err != nil {
		return err
	}
	return a.Save()
}
```

在主程序添加:

```go
case "edit":
	handleEdit()

func handleEdit() {
	if len(os.Args) < 4 {
		printer.PrintError("请提供 ID 和新描述")
		fmt.Println("用法: todo edit <id> <new description>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printer.PrintError("无效的 ID: %s", os.Args[2])
		os.Exit(1)
	}

	newDescription := os.Args[3]
	for i := 4; i < len(os.Args); i++ {
		newDescription += " " + os.Args[i]
	}

	if err := application.Update(id, newDescription); err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	printer.PrintSuccess("已更新 ID %d 的描述", id)
}
```

## 4.4 添加优先级功能（可选扩展）

更新 `internal/todo/todo.go`:

```go
// Priority 优先级
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "低"
	case PriorityMedium:
		return "中"
	case PriorityHigh:
		return "高"
	default:
		return "未知"
	}
}

func (p Priority) Icon() string {
	switch p {
	case PriorityLow:
		return "🟢"
	case PriorityMedium:
		return "🟡"
	case PriorityHigh:
		return "🔴"
	default:
		return "⚪"
	}
}

// Todo 更新后的结构体
type Todo struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Priority    Priority  `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}
```

## 4.5 添加过期时间功能（可选扩展）

```go
// Todo 带过期时间的结构体
type Todo struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt time.Time  `json:"completed_at,omitempty"`
}

// IsOverdue 检查是否已过期
func (t *Todo) IsOverdue() bool {
	if t.DueDate == nil || t.Completed {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// DaysUntilDue 返回距离截止日期的天数
func (t *Todo) DaysUntilDue() int {
	if t.DueDate == nil {
		return -1
	}
	duration := time.Until(*t.DueDate)
	return int(duration.Hours() / 24)
}
```

## 4.6 改进统计功能

更新 `internal/ui/printer.go`:

```go
// PrintDetailedStats 打印详细统计信息
func (p *Printer) PrintDetailedStats(list *todo.List) {
	total := list.Count()
	completed := list.CompletedCount()
	pending := list.PendingCount()

	fmt.Println()
	fmt.Println("📊 " + p.color("详细统计", Blue))
	printSeparator(40)

	// 基本统计
	fmt.Printf(" %-15s %d\n", "总计:", total)
	fmt.Printf(" %-15s %s\n", "已完成:",
		p.color(fmt.Sprintf("%d", completed), Green))
	fmt.Printf(" %-15s %s\n", "待完成:",
		p.color(fmt.Sprintf("%d", pending), Yellow))

	// 完成率
	if total > 0 {
		percentage := float64(completed) / float64(total) * 100
		fmt.Printf(" %-15s %.1f%%\n", "完成率:", percentage)
		printProgressBar(percentage, p.useColor)
	}

	// 今日统计
	todayCompleted := countTodayCompleted(list.All())
	todayAdded := countTodayAdded(list.All())
	fmt.Println()
	fmt.Printf(" %-15s %d\n", "今日添加:", todayAdded)
	fmt.Printf(" %-15s %d\n", "今日完成:", todayCompleted)

	printSeparator(40)
}

func countTodayCompleted(todos []*todo.Todo) int {
	count := 0
	today := time.Now().Truncate(24 * time.Hour)

	for _, t := range todos {
		if t.Completed && !t.CompletedAt.IsZero() {
			if t.CompletedAt.Truncate(24 * time.Hour).Equal(today) {
				count++
			}
		}
	}
	return count
}

func countTodayAdded(todos []*todo.Todo) int {
	count := 0
	today := time.Now().Truncate(24 * time.Hour)

	for _, t := range todos {
		if t.CreatedAt.Truncate(24 * time.Hour).Equal(today) {
			count++
		}
	}
	return count
}
```

## 4.7 添加配置功能

创建 `internal/config/config.go`:

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用配置
type Config struct {
	StoragePath string `json:"storage_path"`
	UseColor    bool   `json:"use_color"`
	ShowCompleted bool `json:"show_completed"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		StoragePath:   filepath.Join(homeDir, ".todos.json"),
		UseColor:      true,
		ShowCompleted: false,
	}
}

// Load 加载配置
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig(), nil
	}

	configPath := filepath.Join(homeDir, ".todo-config.json")

	// 配置文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil
	}

	return &cfg, nil
}

// Save 保存配置
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".todo-config.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
```

## 4.8 最终的帮助信息

```go
func printUsage() {
	usage := `
╔═══════════════════════════════════════════════════════════════╗
║                    Todo CLI v1.0.0                             ║
║              命令行待办事项管理工具                             ║
╚═══════════════════════════════════════════════════════════════╝

用法:
  todo <command> [arguments]

命令:
  add <description>      添加一个新的待办事项
  list, ls [--all]       列出待办事项（默认只显示未完成）
  complete, done <id>    标记待办事项为完成
  delete, rm <id>        删除待办事项
  edit <id> <desc>       编辑待办事项描述
  search <keyword>       搜索待办事项
  clear                  清除所有已完成的待办事项
  stats                  显示统计信息
  help                   显示帮助信息

选项:
  --all, -a              显示所有待办事项（包括已完成）

示例:
  todo add "学习 Go 语言"
  todo add 学习 Go 语言        # 不带引号也可以
  todo list
  todo list --all
  todo complete 1
  todo edit 1 "新的描述"
  todo search Go
  todo delete 1
  todo clear
  todo stats

数据存储:
  待办事项保存在 ~/.todos.json 文件中

更多信息:
  https://github.com/user/todo-cli
`
	fmt.Println(usage)
}
```

## 4.9 完整测试

```bash
# 编译最终版本
go build -o todo cmd/todo/main.go

# 完整功能测试
./todo help

./todo add "学习 Go 语言基础"
./todo add "完成 Todo CLI 项目"
./todo add "学习 Go 并发编程"
./todo add "开始 Todo API 项目"

./todo list
./todo list --all

./todo complete 1
./todo complete 2

./todo search Go
./todo search API

./todo edit 3 "深入学习 Go 并发编程"

./todo stats

./todo delete 4

./todo clear

./todo list --all
```

## 4.10 项目总结

### 学到的知识点

1. **Go 基础语法**
   - 变量声明和类型
   - 结构体定义和方法
   - 切片操作

2. **文件操作**
   - 读写文件
   - JSON 序列化

3. **命令行开发**
   - 参数解析
   - 格式化输出

4. **代码组织**
   - 包的划分
   - 接口设计

### 项目结构回顾

```
todo-cli/
├── cmd/
│   └── todo/
│       └── main.go           # 入口
├── internal/
│   ├── app/
│   │   └── app.go            # 应用层
│   ├── config/
│   │   └── config.go         # 配置
│   ├── storage/
│   │   ├── storage.go        # 接口
│   │   ├── json.go           # JSON 实现
│   │   └── json_test.go      # 测试
│   ├── todo/
│   │   ├── todo.go           # 模型
│   │   ├── list.go           # 列表管理
│   │   └── todo_test.go      # 测试
│   └── ui/
│       └── printer.go        # 输出
├── go.mod
└── README.md
```

## 检查点

完成本步骤后，你应该：

- [x] 实现了搜索功能
- [x] 实现了编辑功能
- [x] 改进了统计显示
- [x] 完成了一个功能完整的 CLI 应用

## 扩展挑战

如果想继续提升，可以尝试：

1. **添加标签功能** - 给待办事项打标签
2. **添加优先级** - 高/中/低优先级
3. **添加截止日期** - 设置 due date
4. **导出功能** - 导出为 Markdown 或 CSV
5. **云同步** - 使用 GitHub Gist 同步

## 下一步

恭喜完成 Todo CLI 项目！现在可以继续：

**[项目 2: Todo API](../02-todo-api/README.md)** - 将 CLI 升级为 RESTful API 服务！
