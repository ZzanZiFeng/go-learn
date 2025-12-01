# 调试技巧

## 学习目标

掌握 Go 程序的调试方法，包括编译错误、运行时错误和性能问题。

## 1. 编译错误调试

### 1.1 常见编译错误

```go
// 错误 1: 未使用的变量
func example() {
    x := 10  // 错误: x declared but not used
}

// 解决: 使用或删除变量
func example() {
    x := 10
    _ = x  // 显式忽略
}
```

```go
// 错误 2: 类型不匹配
func add(a, b int) int {
    return a + b
}

func main() {
    result := add(1.5, 2.5)  // 错误: cannot use 1.5 (type float64) as type int
}

// 解决: 使用正确的类型
func main() {
    result := add(1, 2)
}
```

```go
// 错误 3: 导入未使用
import "fmt"  // 错误: imported and not used

// 解决: 使用或删除导入
import _ "fmt"  // 仅导入副作用
```

### 1.2 使用 go vet

```bash
# 静态分析
go vet ./...

# 检测常见问题
# - 格式化字符串错误
# - 无法到达的代码
# - 可疑的赋值
```

```go
// go vet 可以检测的问题

// 格式化字符串错误
fmt.Printf("%d", "string")  // 警告: wrong type

// 无用的赋值
x = x  // 警告: self-assignment

// 锁复制
var mu sync.Mutex
mu2 := mu  // 警告: copies lock value
```

## 2. 运行时错误调试

### 2.1 Nil 指针错误

```go
// 问题代码
type User struct {
    Name string
}

func getUser() *User {
    return nil
}

func main() {
    user := getUser()
    fmt.Println(user.Name)  // panic: nil pointer dereference
}

// 解决: 检查 nil
func main() {
    user := getUser()
    if user == nil {
        fmt.Println("user not found")
        return
    }
    fmt.Println(user.Name)
}
```

### 2.2 数组越界

```go
// 问题代码
func main() {
    arr := []int{1, 2, 3}
    fmt.Println(arr[5])  // panic: index out of range
}

// 解决: 检查边界
func main() {
    arr := []int{1, 2, 3}
    idx := 5
    if idx >= 0 && idx < len(arr) {
        fmt.Println(arr[idx])
    } else {
        fmt.Println("index out of range")
    }
}
```

### 2.3 使用 recover

```go
func safeOperation() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recovered: %v", r)
            // 打印堆栈
            debug.PrintStack()
        }
    }()

    // 可能 panic 的代码
    riskyOperation()
    return nil
}
```

## 3. 竞态条件检测

### 3.1 使用 -race 标志

```bash
# 运行时检测竞态
go run -race main.go
go test -race ./...
```

### 3.2 竞态条件示例

```go
// 问题代码
var counter int

func increment() {
    counter++  // 竞态条件
}

func main() {
    for i := 0; i < 1000; i++ {
        go increment()
    }
    time.Sleep(time.Second)
    fmt.Println(counter)  // 结果不确定
}
```

```
==================
WARNING: DATA RACE
Write at 0x... by goroutine ...:
  main.increment()
      /path/main.go:8 +0x...
==================
```

```go
// 解决 1: 使用 Mutex
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// 解决 2: 使用 atomic
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}
```

## 4. 使用 Delve 调试器

### 4.1 安装 Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 4.2 基本调试命令

```bash
# 调试程序
dlv debug ./cmd/api

# 调试测试
dlv test ./...

# 附加到进程
dlv attach <pid>
```

### 4.3 常用 Delve 命令

```
(dlv) break main.main          # 设置断点
(dlv) break myfile.go:25       # 在文件行设置断点
(dlv) breakpoints              # 列出断点
(dlv) clear 1                  # 清除断点

(dlv) continue                 # 继续执行
(dlv) next                     # 单步执行（不进入函数）
(dlv) step                     # 单步执行（进入函数）
(dlv) stepout                  # 跳出当前函数

(dlv) print variableName       # 打印变量
(dlv) locals                   # 打印本地变量
(dlv) args                     # 打印函数参数
(dlv) goroutines               # 列出 goroutines
(dlv) goroutine 5              # 切换到 goroutine 5
(dlv) stack                    # 打印堆栈

(dlv) quit                     # 退出
```

### 4.4 调试示例

```go
// main.go
package main

func add(a, b int) int {
    result := a + b
    return result
}

func main() {
    x := 10
    y := 20
    sum := add(x, y)
    println(sum)
}
```

```bash
$ dlv debug main.go

(dlv) break main.add
Breakpoint 1 set at 0x... for main.add() ./main.go:4

(dlv) continue
> main.add() ./main.go:4 (hits goroutine(1):1 total:1)

(dlv) args
a = 10
b = 20

(dlv) next
> main.add() ./main.go:5

(dlv) print result
30

(dlv) continue
30
Process exiting with status: 0
```

### 4.5 VS Code 集成

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/api",
            "env": {
                "DB_HOST": "localhost"
            }
        },
        {
            "name": "Debug Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}",
            "args": ["-test.run", "TestMyFunction"]
        }
    ]
}
```

## 5. 日志调试

### 5.1 结构化日志

```go
import "go.uber.org/zap"

func main() {
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()

    sugar := logger.Sugar()

    // 调试信息
    sugar.Debugw("processing request",
        "method", "GET",
        "path", "/users",
        "user_id", 123,
    )

    // 错误信息
    sugar.Errorw("database error",
        "error", err,
        "query", "SELECT * FROM users",
    )
}
```

### 5.2 条件日志

```go
import (
    "log"
    "os"
)

var debug = os.Getenv("DEBUG") == "true"

func debugLog(format string, args ...interface{}) {
    if debug {
        log.Printf("[DEBUG] "+format, args...)
    }
}

func main() {
    debugLog("Starting application with config: %+v", config)
}
```

## 6. 常见问题场景

### 6.1 死锁调试

```go
// 问题代码
func main() {
    ch := make(chan int)
    ch <- 1  // 死锁：没有接收者
    fmt.Println(<-ch)
}
```

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
        /path/main.go:6 +0x...
```

```go
// 解决: 使用 goroutine 或 buffered channel
func main() {
    ch := make(chan int, 1)  // buffered
    ch <- 1
    fmt.Println(<-ch)
}

// 或
func main() {
    ch := make(chan int)
    go func() {
        ch <- 1
    }()
    fmt.Println(<-ch)
}
```

### 6.2 内存泄漏调试

```go
// 问题: goroutine 泄漏
func process(ch chan int) {
    for {
        select {
        case v := <-ch:
            fmt.Println(v)
        }
        // 没有退出条件
    }
}

func main() {
    for i := 0; i < 1000; i++ {
        ch := make(chan int)
        go process(ch)
        // ch 被废弃但 goroutine 继续运行
    }
}
```

```go
// 解决: 使用 context 取消
func process(ctx context.Context, ch chan int) {
    for {
        select {
        case <-ctx.Done():
            return  // 退出
        case v := <-ch:
            fmt.Println(v)
        }
    }
}

func main() {
    for i := 0; i < 1000; i++ {
        ctx, cancel := context.WithCancel(context.Background())
        ch := make(chan int)
        go process(ctx, ch)

        // 完成后取消
        cancel()
    }
}
```

### 6.3 使用 pprof 诊断内存

```bash
# 查看 goroutine 数量
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 查看堆内存
curl http://localhost:6060/debug/pprof/heap?debug=1

# 交互式分析
go tool pprof http://localhost:6060/debug/pprof/heap
(pprof) top10
(pprof) list functionName
```

## 7. 调试清单

| 问题类型 | 调试工具 | 命令/方法 |
|----------|----------|-----------|
| 编译错误 | go vet | `go vet ./...` |
| 类型错误 | staticcheck | `staticcheck ./...` |
| 竞态条件 | race detector | `go run -race` |
| 运行时错误 | delve | `dlv debug` |
| 性能问题 | pprof | `go tool pprof` |
| 内存泄漏 | pprof heap | `/debug/pprof/heap` |
| goroutine 泄漏 | pprof goroutine | `/debug/pprof/goroutine` |

## 练习

1. 使用 Delve 调试一个带 bug 的程序
2. 使用 race detector 检测并修复竞态条件
3. 使用 pprof 分析内存使用
4. 实现带取消功能的 goroutine

## 总结

本轨道完成了 Go 测试的全面学习：

| 章节 | 内容 |
|------|------|
| 单元测试 | go test、testify、mock |
| 集成测试 | testcontainers、数据库测试 |
| E2E测试 | 端到端流程测试 |
| 性能测试 | benchmark、pprof |
| 调试技巧 | delve、race detector |

## 下一步学习

- 回顾 **[开发轨道](../../01-development/)** 深入技术细节
- 实践 **[实战轨道](../../02-practice/)** 完成项目
- 学习 **[部署轨道](../../03-deployment/)** 上线应用
