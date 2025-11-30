# 第一个 Go 程序

本文档将带你编写并理解第一个 Go 程序 —— Hello World。

## 目录

- [创建项目](#创建项目)
- [编写代码](#编写代码)
- [运行程序](#运行程序)
- [代码详解](#代码详解)
- [编译程序](#编译程序)
- [扩展练习](#扩展练习)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 创建项目

### 步骤 1: 创建项目目录

```bash
# 创建项目目录（可以在任意位置）
mkdir -p ~/go-projects/hello
cd ~/go-projects/hello
```

### 步骤 2: 初始化 Go 模块

```bash
# 初始化 Go 模块
go mod init hello

# 查看生成的 go.mod 文件
cat go.mod
```

输出：
```
module hello

go 1.21
```

---

## 编写代码

### 步骤 3: 创建 main.go

使用 VS Code 或任意编辑器创建 `main.go` 文件：

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### 项目结构

```
hello/
├── go.mod      # 模块定义文件
└── main.go     # 程序入口文件
```

---

## 运行程序

### 方法 1: go run（开发时使用）

```bash
go run main.go
# 或
go run .
```

输出：
```
Hello, World!
```

`go run` 会编译并立即运行程序，不会生成可执行文件。适合开发时快速测试。

### 方法 2: 先编译再运行

```bash
# 编译
go build -o hello

# 运行编译后的程序
./hello
```

输出：
```
Hello, World!
```

---

## 代码详解

让我们逐行分析这个简单的程序：

```go
package main
```

**package 声明**
- 每个 Go 文件必须以 `package` 声明开头
- `package main` 表示这是一个可执行程序（不是库）
- 可执行程序必须使用 `main` 包

```go
import "fmt"
```

**import 语句**
- 导入标准库中的 `fmt` 包
- `fmt` 提供格式化输入输出功能
- 类似于 JavaScript 的 `import` 或 `require`

```go
func main() {
```

**main 函数**
- 程序的入口点
- 可执行程序必须有 `main` 函数
- `func` 关键字定义函数
- Go 的大括号 `{` 必须在同一行（这是语法要求）

```go
    fmt.Println("Hello, World!")
```

**打印语句**
- `fmt.Println` 打印内容并换行
- 函数名首字母大写表示这是一个"导出"的（公开的）函数
- 字符串使用双引号 `"`

```go
}
```

**函数结束**

---

## 编译程序

### 基本编译

```bash
# 编译当前目录的程序
go build

# 指定输出文件名
go build -o myprogram

# 编译指定文件
go build main.go
```

### 交叉编译

Go 支持交叉编译，可以在一个平台编译另一个平台的程序：

```bash
# 在 Mac 上编译 Linux 程序
GOOS=linux GOARCH=amd64 go build -o hello-linux

# 在 Mac 上编译 Windows 程序
GOOS=windows GOARCH=amd64 go build -o hello.exe

# 查看所有支持的平台
go tool dist list
```

### 安装到 GOPATH/bin

```bash
# 编译并安装到 $GOPATH/bin
go install

# 之后可以直接运行（确保 $GOPATH/bin 在 PATH 中）
hello
```

---

## 扩展练习

### 练习 1: 接收用户输入

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("请输入你的名字: ")
    name, _ := reader.ReadString('\n')
    name = strings.TrimSpace(name)

    fmt.Printf("你好, %s!\n", name)
}
```

运行：
```bash
go run .
# 输入: Gopher
# 输出: 你好, Gopher!
```

### 练习 2: 使用命令行参数

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // os.Args[0] 是程序名
    // os.Args[1:] 是传入的参数
    args := os.Args

    fmt.Println("程序名:", args[0])

    if len(args) > 1 {
        fmt.Println("参数:", args[1:])
        fmt.Printf("你好, %s!\n", args[1])
    } else {
        fmt.Println("用法: hello <name>")
    }
}
```

运行：
```bash
go run . World
# 输出:
# 程序名: /tmp/go-build.../hello
# 参数: [World]
# 你好, World!
```

### 练习 3: 多文件项目

创建 `greet.go`：

```go
package main

import "fmt"

func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
```

修改 `main.go`：

```go
package main

import "fmt"

func main() {
    message := greet("Gopher")
    fmt.Println(message)
}
```

运行：
```bash
go run .
# 输出: Hello, Gopher!
```

**注意**: 同一个 `package main` 下的所有文件会被一起编译。

### 练习 4: 创建子包

项目结构：
```
hello/
├── go.mod
├── main.go
└── greeting/
    └── greeting.go
```

`greeting/greeting.go`：

```go
package greeting

import "fmt"

// Hello 返回问候语（首字母大写表示导出）
func Hello(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}

// 小写开头的函数只能在包内使用
func privateFunc() {
    fmt.Println("这是私有函数")
}
```

`main.go`：

```go
package main

import (
    "fmt"
    "hello/greeting"  // 导入子包
)

func main() {
    message := greeting.Hello("Gopher")
    fmt.Println(message)
}
```

运行：
```bash
go run .
# 输出: Hello, Gopher!
```

---

## 与 JavaScript/TypeScript 对比

### Hello World 对比

```javascript
// JavaScript
console.log("Hello, World!");
```

```typescript
// TypeScript
console.log("Hello, World!");
```

```go
// Go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### 关键差异

| 特性 | Go | JavaScript/TypeScript |
|-----|----|-----------------------|
| 入口点 | 必须有 `main` 函数 | 直接执行 |
| 包声明 | 必须有 `package` | 可选 module/export |
| 大括号位置 | 必须同一行 | 可换行 |
| 分号 | 自动插入，通常不写 | 可选 |
| 导出 | 首字母大写 | `export` 关键字 |
| 打印函数 | `fmt.Println` | `console.log` |
| 编译 | 编译型，生成二进制 | 解释型/JIT |
| 运行时 | 无需运行时 | 需要 Node.js/浏览器 |

### 模块系统对比

```javascript
// JavaScript (ES Modules)
// greeting.js
export function hello(name) {
    return `Hello, ${name}!`;
}

// main.js
import { hello } from './greeting.js';
console.log(hello('World'));
```

```go
// Go
// greeting/greeting.go
package greeting

func Hello(name string) string {
    return "Hello, " + name + "!"
}

// main.go
package main

import (
    "fmt"
    "myproject/greeting"
)

func main() {
    fmt.Println(greeting.Hello("World"))
}
```

### 可见性对比

```typescript
// TypeScript - 使用关键字控制可见性
export class User {
    public name: string;      // 公开
    private age: number;      // 私有
    protected id: string;     // 受保护
}
```

```go
// Go - 使用首字母大小写控制可见性
type User struct {
    Name string    // 公开 (首字母大写)
    age  int       // 私有 (首字母小写)
    ID   string    // 公开
}
```

---

## 常见错误

### 错误 1: 大括号换行

```go
// 错误! Go 不允许这样
func main()
{
    fmt.Println("Hello")
}
```

```go
// 正确
func main() {
    fmt.Println("Hello")
}
```

### 错误 2: 未使用的导入

```go
// 错误! 导入了但未使用
import (
    "fmt"
    "os"  // 编译错误: imported and not used
)

func main() {
    fmt.Println("Hello")
}
```

```go
// 正确: 删除未使用的导入或使用 _
import (
    "fmt"
    _ "os"  // 空白标识符，仅执行 init()
)
```

### 错误 3: 未使用的变量

```go
// 错误! 声明了但未使用
func main() {
    x := 10  // 编译错误: x declared but not used
    fmt.Println("Hello")
}
```

```go
// 正确: 使用变量或用 _ 忽略
func main() {
    x := 10
    fmt.Println("Hello", x)

    // 或者忽略
    _ = 10  // 这不会报错
}
```

---

## 验收清单

完成本节后，请确认你能够：

- [ ] 使用 `go mod init` 创建新项目
- [ ] 编写基本的 Hello World 程序
- [ ] 使用 `go run` 运行程序
- [ ] 使用 `go build` 编译程序
- [ ] 理解 `package`, `import`, `func main` 的作用
- [ ] 理解 Go 与 JavaScript/TypeScript 的基本差异

---

## 下一步

恭喜你完成了第一个 Go 程序！接下来：

- [语法基础](../01-syntax/) - 学习 Go 的基本语法

---

## 参考资源

- [Go 语言之旅](https://tour.go-zh.org/) - 官方交互式教程
- [Go by Example](https://gobyexample.com/) - 通过示例学习 Go
- [Effective Go](https://go.dev/doc/effective_go) - Go 编程风格指南
