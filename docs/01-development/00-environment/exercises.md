# DEV-00 练习与解答

本文档包含环境搭建章节的练习题和参考解答。

## 练习 1: 验证 Go 安装

### 题目

完成以下任务并记录输出结果：

1. 运行 `go version` 查看 Go 版本
2. 运行 `go env GOROOT` 查看 Go 安装目录
3. 运行 `go env GOPATH` 查看工作目录
4. 运行 `go env GOPROXY` 查看代理配置

### 解答

```bash
# 1. Go 版本
go version
# 预期输出: go version go1.21.x darwin/arm64 (或其他平台)

# 2. GOROOT
go env GOROOT
# 预期输出: /usr/local/go 或 /opt/homebrew/opt/go/libexec

# 3. GOPATH
go env GOPATH
# 预期输出: /Users/你的用户名/go

# 4. GOPROXY
go env GOPROXY
# 建议配置为: https://goproxy.cn,direct
```

---

## 练习 2: 创建并运行 Hello World

### 题目

1. 创建一个新的 Go 项目目录
2. 初始化 Go 模块
3. 编写 Hello World 程序，输出你的名字
4. 使用 `go run` 运行程序
5. 使用 `go build` 编译程序并运行

### 解答

```bash
# 1. 创建项目目录
mkdir -p ~/go-projects/exercise1
cd ~/go-projects/exercise1

# 2. 初始化模块
go mod init exercise1
```

创建 `main.go`:

```go
package main

import "fmt"

func main() {
    name := "你的名字"
    fmt.Printf("Hello, %s!\n", name)
    fmt.Println("Welcome to Go!")
}
```

```bash
# 4. 使用 go run 运行
go run .
# 输出:
# Hello, 你的名字!
# Welcome to Go!

# 5. 编译并运行
go build -o exercise1
./exercise1
# 输出相同
```

---

## 练习 3: 使用命令行参数

### 题目

编写一个程序，接收命令行参数作为用户名，输出个性化问候语。

要求：
- 如果提供了参数，使用参数作为名字
- 如果没有提供参数，使用默认名字 "Guest"

示例：
```bash
./greet World
# 输出: Hello, World!

./greet
# 输出: Hello, Guest!
```

### 解答

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // 默认名字
    name := "Guest"

    // 检查是否有命令行参数
    // os.Args[0] 是程序名，os.Args[1] 开始是参数
    if len(os.Args) > 1 {
        name = os.Args[1]
    }

    fmt.Printf("Hello, %s!\n", name)
}
```

运行测试：
```bash
go build -o greet
./greet World
# 输出: Hello, World!

./greet
# 输出: Hello, Guest!

./greet "Go Developer"
# 输出: Hello, Go Developer!
```

---

## 练习 4: 多文件项目

### 题目

创建一个包含多个文件的项目：

1. `main.go` - 主程序入口
2. `greeting.go` - 包含问候函数
3. `farewell.go` - 包含告别函数

程序应该：
- 调用问候函数打印 "Hello, World!"
- 调用告别函数打印 "Goodbye, World!"

### 解答

项目结构：
```
exercise4/
├── go.mod
├── main.go
├── greeting.go
└── farewell.go
```

`go.mod`:
```
module exercise4

go 1.21
```

`greeting.go`:
```go
package main

import "fmt"

func greet(name string) {
    fmt.Printf("Hello, %s!\n", name)
}
```

`farewell.go`:
```go
package main

import "fmt"

func farewell(name string) {
    fmt.Printf("Goodbye, %s!\n", name)
}
```

`main.go`:
```go
package main

func main() {
    name := "World"
    greet(name)
    farewell(name)
}
```

运行：
```bash
go run .
# 输出:
# Hello, World!
# Goodbye, World!
```

---

## 练习 5: 创建子包

### 题目

将练习 4 重构为使用子包：

1. 创建 `greetings` 子包
2. 在子包中实现 `Hello` 和 `Goodbye` 函数
3. 在 `main.go` 中导入并使用这些函数

### 解答

项目结构：
```
exercise5/
├── go.mod
├── main.go
└── greetings/
    └── greetings.go
```

`go.mod`:
```
module exercise5

go 1.21
```

`greetings/greetings.go`:
```go
// Package greetings provides functions for greeting and farewell messages.
package greetings

import "fmt"

// Hello prints a greeting message.
// 首字母大写表示这个函数是导出的（公开的）
func Hello(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

// Goodbye prints a farewell message.
func Goodbye(name string) {
    fmt.Printf("Goodbye, %s!\n", name)
}

// getMessage is a private helper function.
// 首字母小写表示这个函数是私有的（只能在包内使用）
func getMessage(template, name string) string {
    return fmt.Sprintf(template, name)
}
```

`main.go`:
```go
package main

import "exercise5/greetings"

func main() {
    name := "Gopher"

    greetings.Hello(name)
    greetings.Goodbye(name)

    // 下面这行会编译错误，因为 getMessage 是私有函数
    // greetings.getMessage("Hi, %s!", name)
}
```

运行：
```bash
go run .
# 输出:
# Hello, Gopher!
# Goodbye, Gopher!
```

---

## 练习 6: 使用外部包

### 题目

安装并使用 `github.com/fatih/color` 包，输出彩色的 Hello World。

要求：
- 用青色打印 "Hello"
- 用绿色打印你的名字
- 用黄色打印 "!"

### 解答

```bash
# 创建项目
mkdir -p ~/go-projects/exercise6
cd ~/go-projects/exercise6
go mod init exercise6

# 安装依赖
go get github.com/fatih/color
```

`main.go`:
```go
package main

import "github.com/fatih/color"

func main() {
    cyan := color.New(color.FgCyan)
    green := color.New(color.FgGreen)
    yellow := color.New(color.FgYellow)

    cyan.Print("Hello, ")
    green.Print("Gopher")
    yellow.Println("!")

    // 或者使用简便方法
    color.Cyan("This is cyan text")
    color.Green("This is green text")
    color.Yellow("This is yellow text")
}
```

运行：
```bash
go run .
# 输出彩色文本
```

---

## 练习 7: VS Code 功能验证

### 题目

在 VS Code 中验证以下功能是否正常工作：

1. 代码自动补全
2. 跳转到定义（F12）
3. 保存时自动格式化
4. 保存时自动添加/删除导入
5. 错误提示

### 解答

创建以下测试文件 `test.go`：

```go
package main

import "fmt"

func main() {
    // 测试 1: 输入 fmt. 应该显示自动补全列表
    fmt.Println("test")

    // 测试 2: 在 Println 上按 F12 应该跳转到定义

    // 测试 3: 故意写乱格式，保存后应该自动格式化
    x:=1;y:=2;z:=x+y
    fmt.Println(z)

    // 测试 4: 添加 time.Now()，保存后应该自动添加 import "time"
    // 删除下面这行后保存，import "time" 应该自动删除
    // _ = time.Now()

    // 测试 5: 下面这行应该显示红色错误波浪线
    // undefinedFunction()
}
```

验证清单：
- [ ] 输入 `fmt.` 后显示补全列表（包含 Println, Printf 等）
- [ ] 在 `Println` 上按 F12 跳转到 fmt 包源码
- [ ] 保存后代码自动格式化
- [ ] 添加 `time.Now()` 保存后自动添加 `import "time"`
- [ ] 错误代码显示红色波浪线

---

## 自测清单

完成所有练习后，确认你能够：

- [ ] 正确安装和配置 Go 环境
- [ ] 使用 `go mod init` 创建新项目
- [ ] 编写和运行基本的 Go 程序
- [ ] 使用 `go build` 编译程序
- [ ] 创建多文件项目
- [ ] 创建和使用子包
- [ ] 安装和使用外部依赖包
- [ ] 在 VS Code 中进行 Go 开发

---

## 下一步

完成这些练习后，你已经掌握了 Go 开发环境的基础使用。继续学习：

- [语法基础](../01-syntax/) - 学习 Go 的基本语法
