# Step 2: CRUD 操作

## 目标

实现命令行参数解析，完成添加、列出、完成、删除功能。

## 2.1 命令行参数解析

更新 `cmd/todo/main.go`，使用 `os.Args` 解析命令：

```go
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/user/todo-cli/internal/todo"
)

// 全局待办列表
var list *todo.List

func main() {
	list = todo.NewList()

	// 检查是否有参数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 获取命令
	command := os.Args[1]

	// 执行命令
	switch command {
	case "add":
		handleAdd()
	case "list", "ls":
		handleList()
	case "complete", "done":
		handleComplete()
	case "delete", "rm":
		handleDelete()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usage := `
Todo CLI - 命令行待办事项管理工具

用法:
  todo <command> [arguments]

命令:
  add <description>    添加一个新的待办事项
  list, ls             列出所有待办事项
  complete, done <id>  标记待办事项为完成
  delete, rm <id>      删除待办事项
  help                 显示帮助信息

示例:
  todo add "学习 Go 语言"
  todo list
  todo complete 1
  todo delete 1
`
	fmt.Println(usage)
}

func handleAdd() {
	// 检查是否有描述
	if len(os.Args) < 3 {
		fmt.Println("错误: 请提供待办事项描述")
		fmt.Println("用法: todo add <description>")
		os.Exit(1)
	}

	// 获取描述（支持多个参数拼接）
	description := os.Args[2]
	if len(os.Args) > 3 {
		for i := 3; i < len(os.Args); i++ {
			description += " " + os.Args[i]
		}
	}

	// 添加待办
	t, err := list.Add(description)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 已添加: \"%s\" (ID: %d)\n", t.Description, t.ID)
}

func handleList() {
	todos := list.All()

	if len(todos) == 0 {
		fmt.Println("📝 待办列表为空")
		return
	}

	fmt.Printf("\n📋 待办列表 (%d 项)\n", len(todos))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf(" %-4s %-8s %s\n", "ID", "状态", "描述")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, t := range todos {
		status := "[ ]"
		if t.Completed {
			status = "[✓]"
		}
		fmt.Printf(" %-4d %-8s %s\n", t.ID, status, t.Description)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func handleComplete() {
	// 检查是否有 ID
	if len(os.Args) < 3 {
		fmt.Println("错误: 请提供待办事项 ID")
		fmt.Println("用法: todo complete <id>")
		os.Exit(1)
	}

	// 解析 ID
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("错误: 无效的 ID: %s\n", os.Args[2])
		os.Exit(1)
	}

	// 获取待办（用于显示）
	t, err := list.Get(id)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 标记完成
	err = list.Complete(id)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 已完成: \"%s\"\n", t.Description)
}

func handleDelete() {
	// 检查是否有 ID
	if len(os.Args) < 3 {
		fmt.Println("错误: 请提供待办事项 ID")
		fmt.Println("用法: todo delete <id>")
		os.Exit(1)
	}

	// 解析 ID
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("错误: 无效的 ID: %s\n", os.Args[2])
		os.Exit(1)
	}

	// 获取待办（用于显示）
	t, err := list.Get(id)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	description := t.Description

	// 删除
	err = list.Delete(id)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🗑️  已删除: \"%s\"\n", description)
}
```

## 2.2 测试命令行功能

```bash
# 编译
go build -o todo cmd/todo/main.go

# 测试各命令
./todo help
./todo add "学习 Go 语言"
./todo add "完成 Todo CLI"
./todo list
./todo complete 1
./todo list
./todo delete 2
./todo list
```

当前问题：数据不会保存，每次运行都是空列表。我们将在 Step 3 解决。

## 2.3 使用 flag 包（可选）

对于更复杂的命令行参数，可以使用 `flag` 包：

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

// 子命令
var (
	addCmd      = flag.NewFlagSet("add", flag.ExitOnError)
	listCmd     = flag.NewFlagSet("list", flag.ExitOnError)
	completeCmd = flag.NewFlagSet("complete", flag.ExitOnError)
	deleteCmd   = flag.NewFlagSet("delete", flag.ExitOnError)
)

// list 命令的选项
var listAll bool

func init() {
	// 为 list 命令添加 --all 选项
	listCmd.BoolVar(&listAll, "all", false, "显示所有待办（包括已完成）")
	listCmd.BoolVar(&listAll, "a", false, "显示所有待办（简写）")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		if addCmd.NArg() < 1 {
			fmt.Println("错误: 请提供待办事项描述")
			os.Exit(1)
		}
		// 使用 addCmd.Args() 获取描述
		description := addCmd.Arg(0)
		fmt.Printf("添加: %s\n", description)

	case "list", "ls":
		listCmd.Parse(os.Args[2:])
		if listAll {
			fmt.Println("显示所有待办...")
		} else {
			fmt.Println("显示未完成待办...")
		}

	// ... 其他命令
	}
}
```

## 2.4 命令行输出美化

创建 `internal/ui/printer.go`：

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/user/todo-cli/internal/todo"
)

// 颜色代码
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Gray   = "\033[90m"
)

// Printer 负责格式化输出
type Printer struct {
	useColor bool
}

// NewPrinter 创建打印器
func NewPrinter(useColor bool) *Printer {
	return &Printer{useColor: useColor}
}

// color 应用颜色
func (p *Printer) color(text, colorCode string) string {
	if p.useColor {
		return colorCode + text + Reset
	}
	return text
}

// PrintSuccess 打印成功消息
func (p *Printer) PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color("✅ "+msg, Green))
}

// PrintError 打印错误消息
func (p *Printer) PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color("❌ "+msg, Red))
}

// PrintWarning 打印警告消息
func (p *Printer) PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color("⚠️  "+msg, Yellow))
}

// PrintInfo 打印信息消息
func (p *Printer) PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color("ℹ️  "+msg, Blue))
}

// PrintTodoList 打印待办列表
func (p *Printer) PrintTodoList(todos []*todo.Todo, showAll bool) {
	if len(todos) == 0 {
		p.PrintInfo("待办列表为空")
		return
	}

	// 过滤
	displayTodos := todos
	if !showAll {
		displayTodos = filterPending(todos)
	}

	// 统计
	total := len(todos)
	completed := countCompleted(todos)
	pending := total - completed

	// 标题
	fmt.Println()
	fmt.Printf("📋 %s\n", p.color(fmt.Sprintf("待办列表 (%d 项, %d 已完成)", total, completed), Blue))
	printSeparator(50)

	// 表头
	fmt.Printf(" %-4s  %-6s  %s\n", "ID", "状态", "描述")
	printSeparator(50)

	// 内容
	for _, t := range displayTodos {
		p.printTodo(t)
	}

	printSeparator(50)

	// 底部统计
	if showAll {
		fmt.Printf(" %s: %d  |  %s: %d  |  %s: %d\n",
			p.color("总计", Blue), total,
			p.color("已完成", Green), completed,
			p.color("待完成", Yellow), pending,
		)
	} else {
		fmt.Printf(" %s: %d\n", p.color("待完成", Yellow), pending)
	}
}

func (p *Printer) printTodo(t *todo.Todo) {
	var status, desc string

	if t.Completed {
		status = p.color("[✓]", Green)
		desc = p.color(t.Description, Gray) // 已完成用灰色
	} else {
		status = p.color("[ ]", Yellow)
		desc = t.Description
	}

	fmt.Printf(" %-4d  %-6s  %s\n", t.ID, status, desc)
}

// PrintStats 打印统计信息
func (p *Printer) PrintStats(list *todo.List) {
	total := list.Count()
	completed := list.CompletedCount()
	pending := list.PendingCount()

	fmt.Println()
	fmt.Println("📊 " + p.color("待办统计", Blue))
	printSeparator(30)

	percentage := float64(0)
	if total > 0 {
		percentage = float64(completed) / float64(total) * 100
	}

	fmt.Printf(" 总计:     %d\n", total)
	fmt.Printf(" 已完成:   %s (%s)\n",
		p.color(fmt.Sprintf("%d", completed), Green),
		p.color(fmt.Sprintf("%.0f%%", percentage), Green),
	)
	fmt.Printf(" 待完成:   %s (%s)\n",
		p.color(fmt.Sprintf("%d", pending), Yellow),
		p.color(fmt.Sprintf("%.0f%%", 100-percentage), Yellow),
	)

	// 进度条
	printProgressBar(percentage, p.useColor)
}

func printSeparator(width int) {
	fmt.Println(strings.Repeat("─", width))
}

func printProgressBar(percentage float64, useColor bool) {
	width := 30
	filled := int(percentage / 100 * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	if useColor {
		bar = Green + strings.Repeat("█", filled) + Reset + Gray + strings.Repeat("░", width-filled) + Reset
	}

	fmt.Printf(" 进度:     [%s] %.0f%%\n", bar, percentage)
}

func filterPending(todos []*todo.Todo) []*todo.Todo {
	result := make([]*todo.Todo, 0)
	for _, t := range todos {
		if !t.Completed {
			result = append(result, t)
		}
	}
	return result
}

func countCompleted(todos []*todo.Todo) int {
	count := 0
	for _, t := range todos {
		if t.Completed {
			count++
		}
	}
	return count
}
```

## 2.5 更新主程序使用 Printer

```go
package main

import (
	"os"
	"strconv"

	"github.com/user/todo-cli/internal/todo"
	"github.com/user/todo-cli/internal/ui"
)

var (
	list    *todo.List
	printer *ui.Printer
)

func main() {
	list = todo.NewList()
	printer = ui.NewPrinter(true) // 启用颜色

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
		os.Exit(1)
	}

	description := os.Args[2]
	for i := 3; i < len(os.Args); i++ {
		description += " " + os.Args[i]
	}

	t, err := list.Add(description)
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

	printer.PrintTodoList(list.All(), showAll)
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

	t, err := list.Get(id)
	if err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	list.Complete(id)
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

	t, err := list.Get(id)
	if err != nil {
		printer.PrintError("%v", err)
		os.Exit(1)
	}

	description := t.Description
	list.Delete(id)
	printer.PrintSuccess("已删除: \"%s\"", description)
}

func handleStats() {
	printer.PrintStats(list)
}

func printUsage() {
	// ... 同前
}
```

## 检查点

完成本步骤后，你应该：

- [x] 理解如何解析命令行参数
- [x] 实现了完整的 CRUD 命令
- [x] 掌握了基本的输出格式化

## 下一步

[Step 3: 数据持久化](./step-03-storage.md) - 实现 JSON 文件存储，让数据能够保存。
