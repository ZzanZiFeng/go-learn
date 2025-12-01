# 项目 1: Todo CLI

## 项目概述

构建一个命令行待办事项应用，这是你的第一个 Go 实战项目。通过这个项目，你将掌握 Go 的基础语法应用、文件操作和 CLI 开发。

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Todo CLI 架构                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│    用户输入                  业务逻辑                    数据存储          │
│  ┌─────────┐             ┌─────────────┐             ┌─────────┐        │
│  │ 命令行  │────────────▶│  Todo 服务  │────────────▶│ JSON 文件│        │
│  │ 参数    │             │             │             │ 存储    │        │
│  └─────────┘             └─────────────┘             └─────────┘        │
│                                                                         │
│    命令示例:                                                            │
│    todo add "学习 Go"                                                   │
│    todo list                                                            │
│    todo complete 1                                                      │
│    todo delete 1                                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## 功能需求

### 核心功能

1. **添加待办** - `todo add "任务描述"`
2. **列出待办** - `todo list` / `todo list --all`
3. **完成待办** - `todo complete <id>`
4. **删除待办** - `todo delete <id>`

### 附加功能

5. **搜索待办** - `todo search <关键词>`
6. **清空已完成** - `todo clear`
7. **统计信息** - `todo stats`

## 技术要点

| 技术点 | 说明 | 对应章节 |
|--------|------|----------|
| 基础语法 | 变量、类型、控制流 | DEV-01 |
| 结构体 | Todo 数据模型 | DEV-02 |
| 切片操作 | 存储和管理 Todo 列表 | DEV-01 |
| 文件 I/O | JSON 读写 | DEV-01 |
| 错误处理 | 优雅的错误提示 | DEV-01 |
| 命令行参数 | os.Args / flag 包 | DEV-01 |

## 项目结构

```
projects/todo-cli/
├── cmd/
│   └── todo/
│       └── main.go           # 入口，命令解析
├── internal/
│   ├── todo/
│   │   ├── todo.go           # Todo 结构体和操作
│   │   └── todo_test.go      # 单元测试
│   └── storage/
│       ├── json.go           # JSON 文件存储
│       └── json_test.go      # 存储测试
├── .todos.json               # 数据文件（运行时生成）
├── go.mod
├── go.sum
└── README.md
```

## 学习步骤

### Step 1: 项目初始化

[▶ step-01-init.md](./step-01-init.md)

- 创建项目目录结构
- 初始化 Go 模块
- 设计 Todo 数据模型

### Step 2: CRUD 操作

[▶ step-02-crud.md](./step-02-crud.md)

- 实现添加功能
- 实现列出功能
- 实现完成功能
- 实现删除功能

### Step 3: 数据持久化

[▶ step-03-storage.md](./step-03-storage.md)

- JSON 文件读写
- 自动保存和加载
- 错误处理

### Step 4: 完善与优化

[▶ step-04-polish.md](./step-04-polish.md)

- 命令行美化
- 搜索和过滤
- 统计功能

## 预期产出

完成后，你将拥有一个功能完整的命令行 Todo 应用：

```bash
$ todo add "学习 Go 语言基础"
✅ Added: "学习 Go 语言基础" (ID: 1)

$ todo add "完成 Todo CLI 项目"
✅ Added: "完成 Todo CLI 项目" (ID: 2)

$ todo add "开始 Todo API 项目"
✅ Added: "开始 Todo API 项目" (ID: 3)

$ todo list
Todo List (3 items)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 ID  Status  Description
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 1   [ ]     学习 Go 语言基础
 2   [ ]     完成 Todo CLI 项目
 3   [ ]     开始 Todo API 项目
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

$ todo complete 1
✅ Completed: "学习 Go 语言基础"

$ todo list
Todo List (3 items, 1 completed)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 ID  Status  Description
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 1   [✓]     学习 Go 语言基础
 2   [ ]     完成 Todo CLI 项目
 3   [ ]     开始 Todo API 项目
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

$ todo stats
📊 Todo Statistics
━━━━━━━━━━━━━━━━━━━━
Total:     3
Completed: 1 (33%)
Pending:   2 (67%)
```

## JS/TS 对比

如果你熟悉 Node.js CLI 开发，这里是主要差异：

| 方面 | Go | Node.js |
|------|-----|---------|
| 入口 | `main.go` | `index.js` + `bin` |
| 参数解析 | `os.Args` / `flag` | `commander` / `yargs` |
| 文件操作 | `os` / `io/ioutil` | `fs` 模块 |
| JSON | `encoding/json` | 内置 JSON |
| 编译 | `go build` 生成二进制 | 需要 Node.js 运行 |

```javascript
// Node.js 版本示例
const fs = require('fs');
const commander = require('commander');

// Go 版本将在教程中实现
```

## 常见问题

### Q: 数据存储在哪里？

数据默认存储在用户主目录下的 `.todos.json` 文件中。

### Q: 如何处理并发写入？

这个入门项目不考虑并发。在后续的 Todo API 项目中会学习如何处理。

### Q: 可以添加过期时间吗？

可以作为扩展功能自行实现。提示：在 Todo 结构体中添加 `DueDate` 字段。

## 下一步

准备好了吗？让我们开始：

**[Step 1: 项目初始化](./step-01-init.md)**
