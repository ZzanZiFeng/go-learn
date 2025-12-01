// Package main 是 Todo CLI 应用的入口
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/todo-cli/internal/storage"
	"github.com/user/todo-cli/internal/todo"
)

// ANSI 颜色代码
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Gray   = "\033[90m"
)

var (
	list    *todo.List
	store   *storage.JSONStorage
)

func main() {
	// 初始化存储
	var err error
	store, err = storage.DefaultJSONStorage()
	if err != nil {
		printError("初始化存储失败: %v", err)
		os.Exit(1)
	}

	// 初始化列表
	list = todo.NewList()

	// 加载数据
	todos, err := store.Load()
	if err != nil {
		printError("加载数据失败: %v", err)
		os.Exit(1)
	}
	list.SetItems(todos)

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
	case "edit":
		handleEdit()
	case "search":
		handleSearch()
	case "clear":
		handleClear()
	case "stats":
		handleStats()
	case "help", "-h", "--help":
		printUsage()
	default:
		printError("未知命令: %s", command)
		printUsage()
		os.Exit(1)
	}
}

func handleAdd() {
	if len(os.Args) < 3 {
		printError("请提供待办事项描述")
		fmt.Println("用法: todo add <description>")
		os.Exit(1)
	}

	description := strings.Join(os.Args[2:], " ")

	t, err := list.Add(description)
	if err != nil {
		printError("%v", err)
		os.Exit(1)
	}

	if err := save(); err != nil {
		printError("保存失败: %v", err)
		os.Exit(1)
	}

	printSuccess("已添加: \"%s\" (ID: %d)", t.Description, t.ID)
}

func handleList() {
	showAll := false
	if len(os.Args) > 2 && (os.Args[2] == "--all" || os.Args[2] == "-a") {
		showAll = true
	}

	todos := list.All()

	if len(todos) == 0 {
		printInfo("待办列表为空")
		return
	}

	// 过滤
	displayTodos := todos
	if !showAll {
		displayTodos = list.Pending()
	}

	// 统计
	total := len(todos)
	completed := list.CompletedCount()

	fmt.Println()
	fmt.Printf("📋 %s待办列表 (%d 项, %d 已完成)%s\n", Blue, total, completed, Reset)
	printSeparator(50)
	fmt.Printf(" %-4s  %-6s  %s\n", "ID", "状态", "描述")
	printSeparator(50)

	for _, t := range displayTodos {
		printTodoItem(t)
	}

	printSeparator(50)

	if showAll {
		fmt.Printf(" %s总计%s: %d  |  %s已完成%s: %d  |  %s待完成%s: %d\n",
			Blue, Reset, total,
			Green, Reset, completed,
			Yellow, Reset, list.PendingCount(),
		)
	} else {
		fmt.Printf(" %s待完成%s: %d\n", Yellow, Reset, len(displayTodos))
	}
}

func handleComplete() {
	if len(os.Args) < 3 {
		printError("请提供待办事项 ID")
		fmt.Println("用法: todo complete <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printError("无效的 ID: %s", os.Args[2])
		os.Exit(1)
	}

	t, err := list.Get(id)
	if err != nil {
		printError("%v", err)
		os.Exit(1)
	}

	if err := list.Complete(id); err != nil {
		printError("%v", err)
		os.Exit(1)
	}

	if err := save(); err != nil {
		printError("保存失败: %v", err)
		os.Exit(1)
	}

	printSuccess("已完成: \"%s\"", t.Description)
}

func handleDelete() {
	if len(os.Args) < 3 {
		printError("请提供待办事项 ID")
		fmt.Println("用法: todo delete <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printError("无效的 ID: %s", os.Args[2])
		os.Exit(1)
	}

	t, err := list.Get(id)
	if err != nil {
		printError("%v", err)
		os.Exit(1)
	}

	description := t.Description

	if err := list.Delete(id); err != nil {
		printError("%v", err)
		os.Exit(1)
	}

	if err := save(); err != nil {
		printError("保存失败: %v", err)
		os.Exit(1)
	}

	printSuccess("已删除: \"%s\"", description)
}

func handleEdit() {
	if len(os.Args) < 4 {
		printError("请提供 ID 和新描述")
		fmt.Println("用法: todo edit <id> <new description>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		printError("无效的 ID: %s", os.Args[2])
		os.Exit(1)
	}

	newDescription := strings.Join(os.Args[3:], " ")

	if err := list.Update(id, newDescription); err != nil {
		printError("%v", err)
		os.Exit(1)
	}

	if err := save(); err != nil {
		printError("保存失败: %v", err)
		os.Exit(1)
	}

	printSuccess("已更新 ID %d 的描述为: \"%s\"", id, newDescription)
}

func handleSearch() {
	if len(os.Args) < 3 {
		printError("请提供搜索关键词")
		fmt.Println("用法: todo search <keyword>")
		os.Exit(1)
	}

	keyword := os.Args[2]
	results := list.Search(keyword)

	if len(results) == 0 {
		printInfo("没有找到匹配 \"%s\" 的待办事项", keyword)
		return
	}

	fmt.Println()
	fmt.Printf("🔍 搜索结果: \"%s\" (%d 项)\n", keyword, len(results))
	printSeparator(50)

	for _, t := range results {
		printTodoItem(t)
	}

	printSeparator(50)
}

func handleClear() {
	count := list.Clear()

	if count == 0 {
		printInfo("没有已完成的待办事项")
		return
	}

	if err := save(); err != nil {
		printError("保存失败: %v", err)
		os.Exit(1)
	}

	printSuccess("已清除 %d 个已完成的待办事项", count)
}

func handleStats() {
	total := list.Count()
	completed := list.CompletedCount()
	pending := list.PendingCount()

	fmt.Println()
	fmt.Printf("📊 %s待办统计%s\n", Blue, Reset)
	printSeparator(30)

	percentage := float64(0)
	if total > 0 {
		percentage = float64(completed) / float64(total) * 100
	}

	fmt.Printf(" 总计:     %d\n", total)
	fmt.Printf(" 已完成:   %s%d%s (%.0f%%)\n", Green, completed, Reset, percentage)
	fmt.Printf(" 待完成:   %s%d%s (%.0f%%)\n", Yellow, pending, Reset, 100-percentage)

	// 进度条
	printProgressBar(percentage)

	// 今日统计
	todayCompleted := countTodayCompleted()
	todayAdded := countTodayAdded()

	fmt.Println()
	fmt.Printf(" 今日添加: %d\n", todayAdded)
	fmt.Printf(" 今日完成: %d\n", todayCompleted)

	printSeparator(30)
}

func save() error {
	return store.Save(list.All())
}

func printTodoItem(t *todo.Todo) {
	var status, desc string

	if t.Completed {
		status = Green + "[✓]" + Reset
		desc = Gray + t.Description + Reset
	} else {
		status = Yellow + "[ ]" + Reset
		desc = t.Description
	}

	fmt.Printf(" %-4d  %s  %s\n", t.ID, status, desc)
}

func printSeparator(width int) {
	fmt.Println(strings.Repeat("─", width))
}

func printProgressBar(percentage float64) {
	width := 30
	filled := int(percentage / 100 * float64(width))

	bar := Green + strings.Repeat("█", filled) + Reset + Gray + strings.Repeat("░", width-filled) + Reset

	fmt.Printf(" 进度:     [%s] %.0f%%\n", bar, percentage)
}

func countTodayCompleted() int {
	count := 0
	today := time.Now().Truncate(24 * time.Hour)

	for _, t := range list.All() {
		if t.Completed && !t.CompletedAt.IsZero() {
			if t.CompletedAt.Truncate(24 * time.Hour).Equal(today) {
				count++
			}
		}
	}
	return count
}

func countTodayAdded() int {
	count := 0
	today := time.Now().Truncate(24 * time.Hour)

	for _, t := range list.All() {
		if t.CreatedAt.Truncate(24 * time.Hour).Equal(today) {
			count++
		}
	}
	return count
}

func printSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Green + "✅ " + msg + Reset)
}

func printError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Red + "❌ " + msg + Reset)
}

func printInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Blue + "ℹ️  " + msg + Reset)
}

func printUsage() {
	usage := `
╔═══════════════════════════════════════════════════════════════╗
║                    Todo CLI v1.0.0                            ║
║              命令行待办事项管理工具                            ║
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
`
	fmt.Println(usage)
}
