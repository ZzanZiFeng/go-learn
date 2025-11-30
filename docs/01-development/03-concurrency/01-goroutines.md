# Goroutines

本文档介绍 Go 语言的 goroutine，与 JavaScript 的 async/await 进行对比。

## 目录

- [什么是 Goroutine](#什么是-goroutine)
- [创建 Goroutine](#创建-goroutine)
- [Goroutine 生命周期](#goroutine-生命周期)
- [匿名函数 Goroutine](#匿名函数-goroutine)
- [常见陷阱](#常见陷阱)
- [与 JavaScript 对比](#与-javascript-对比)

---

## 什么是 Goroutine

### 基本概念

Goroutine 是 Go 的轻量级线程，由 Go 运行时管理。相比操作系统线程：

| 特性 | OS 线程 | Goroutine |
|-----|--------|-----------|
| 内存占用 | ~1MB | ~2KB |
| 创建成本 | 高 | 极低 |
| 切换成本 | 高 | 低 |
| 数量限制 | 数千个 | 数百万个 |

### 为什么使用 Goroutine

```go
package main

import (
    "fmt"
    "time"
)

func slowTask(name string) {
    fmt.Printf("%s started\n", name)
    time.Sleep(2 * time.Second) // 模拟耗时操作
    fmt.Printf("%s completed\n", name)
}

func main() {
    start := time.Now()

    // 顺序执行：6秒
    // slowTask("Task 1")
    // slowTask("Task 2")
    // slowTask("Task 3")

    // 并发执行：2秒
    go slowTask("Task 1")
    go slowTask("Task 2")
    go slowTask("Task 3")

    time.Sleep(3 * time.Second) // 等待完成

    fmt.Printf("Total time: %v\n", time.Since(start))
}
```

---

## 创建 Goroutine

### 基本语法

使用 `go` 关键字启动 goroutine：

```go
go functionName(arguments)
```

### 示例

```go
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

func countTo(n int) {
    for i := 1; i <= n; i++ {
        fmt.Printf("Count: %d\n", i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    // 启动 goroutine
    go sayHello("Alice")
    go sayHello("Bob")
    go countTo(5)

    // main 函数继续执行
    fmt.Println("Main function running...")

    // 等待 goroutine 完成
    time.Sleep(1 * time.Second)
    fmt.Println("Main function done")
}
```

**输出**（顺序可能不同）:
```
Main function running...
Hello, Alice!
Hello, Bob!
Count: 1
Count: 2
Count: 3
Count: 4
Count: 5
Main function done
```

---

## Goroutine 生命周期

### Main Goroutine

程序从 main goroutine 开始，当 main 函数返回时，所有 goroutine 都会被终止。

```go
package main

import (
    "fmt"
    "time"
)

func worker() {
    for i := 0; ; i++ {
        fmt.Println("Working...", i)
        time.Sleep(500 * time.Millisecond)
    }
}

func main() {
    go worker()

    // main 2秒后退出，worker goroutine 也会被终止
    time.Sleep(2 * time.Second)
    fmt.Println("Main exiting")
}
```

### 正确等待 Goroutine

使用 `time.Sleep` 是不可靠的，应该使用 `sync.WaitGroup`：

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done() // 完成时通知 WaitGroup

    fmt.Printf("Worker %d starting\n", id)
    time.Sleep(time.Duration(id) * 100 * time.Millisecond)
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 5; i++ {
        wg.Add(1) // 增加计数
        go worker(i, &wg)
    }

    wg.Wait() // 等待所有 worker 完成
    fmt.Println("All workers completed")
}
```

---

## 匿名函数 Goroutine

### 基本用法

```go
go func() {
    // goroutine 代码
}()
```

### 捕获变量

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup

    // 错误示例：捕获循环变量
    fmt.Println("Wrong way:")
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            fmt.Println(i) // 可能打印 3, 3, 3
        }()
    }
    wg.Wait()

    fmt.Println("\nRight way (pass as parameter):")
    // 正确示例：作为参数传递
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            fmt.Println(n) // 打印 0, 1, 2（顺序可能不同）
        }(i)
    }
    wg.Wait()
}
```

### 返回值处理

Goroutine 不能直接返回值，需要使用 channel：

```go
package main

import "fmt"

func square(n int, ch chan<- int) {
    ch <- n * n
}

func main() {
    ch := make(chan int)

    go square(5, ch)

    result := <-ch
    fmt.Println("Square:", result) // Square: 25
}
```

---

## 常见陷阱

### 1. Goroutine 泄漏

```go
// 错误：永远阻塞的 goroutine
func leak() {
    ch := make(chan int)
    go func() {
        val := <-ch // 永远等待，因为没有人发送
        fmt.Println(val)
    }()
    // ch 没有被发送数据，goroutine 泄漏
}

// 正确：使用 context 或 done channel
func noLeak() {
    ch := make(chan int)
    done := make(chan struct{})

    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-done:
            return // 收到取消信号，退出
        }
    }()

    close(done) // 取消 goroutine
}
```

### 2. 数据竞争

```go
// 错误：多个 goroutine 同时访问共享变量
var counter int

func increment() {
    counter++ // 数据竞争！
}

func main() {
    for i := 0; i < 1000; i++ {
        go increment()
    }
    // counter 的值不确定
}

// 正确：使用 sync.Mutex 或 atomic
import "sync"

var (
    counter int
    mu      sync.Mutex
)

func safeIncrement() {
    mu.Lock()
    counter++
    mu.Unlock()
}
```

### 3. Main 函数提前退出

```go
// 错误：main 退出太快
func main() {
    go fmt.Println("Hello")
    // 程序立即退出，可能看不到输出
}

// 正确：等待 goroutine 完成
func main() {
    done := make(chan struct{})
    go func() {
        fmt.Println("Hello")
        close(done)
    }()
    <-done // 等待
}
```

---

## 与 JavaScript 对比

### 异步执行模型

```javascript
// JavaScript - 事件循环（单线程）
async function fetchData() {
    console.log("Fetching...");
    const response = await fetch(url);
    console.log("Done");
    return response.json();
}

// 调用
fetchData().then(data => console.log(data));
console.log("This runs immediately");
```

```go
// Go - Goroutine（多线程）
func fetchData(url string, ch chan<- []byte) {
    fmt.Println("Fetching...")
    resp, _ := http.Get(url)
    body, _ := io.ReadAll(resp.Body)
    fmt.Println("Done")
    ch <- body
}

// 调用
ch := make(chan []byte)
go fetchData(url, ch)
fmt.Println("This runs immediately")
data := <-ch // 等待结果
```

### Promise.all vs WaitGroup

```javascript
// JavaScript - Promise.all
const results = await Promise.all([
    fetch(url1),
    fetch(url2),
    fetch(url3)
]);
```

```go
// Go - WaitGroup
var wg sync.WaitGroup
results := make([][]byte, 3)

for i, url := range []string{url1, url2, url3} {
    wg.Add(1)
    go func(idx int, u string) {
        defer wg.Done()
        resp, _ := http.Get(u)
        results[idx], _ = io.ReadAll(resp.Body)
    }(i, url)
}

wg.Wait() // 等待所有完成
```

### 错误处理

```javascript
// JavaScript
try {
    const result = await riskyOperation();
} catch (error) {
    console.error(error);
}
```

```go
// Go - 通过 channel 传递错误
type Result struct {
    Data []byte
    Err  error
}

func riskyOperation(ch chan<- Result) {
    data, err := doSomething()
    ch <- Result{Data: data, Err: err}
}

ch := make(chan Result)
go riskyOperation(ch)

result := <-ch
if result.Err != nil {
    log.Println(result.Err)
}
```

### 主要差异

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 执行模型 | 事件循环（单线程） | goroutine（多线程） |
| 阻塞 | 永远不阻塞 | 可以阻塞 |
| 启动语法 | `async`/`await` | `go` |
| 返回值 | Promise | Channel |
| 错误处理 | try/catch | Channel 传递 |
| CPU 密集任务 | 阻塞事件循环 | 可以并行 |

---

## 最佳实践

### 1. 明确 Goroutine 的所有者

```go
// 好：函数创建并管理 goroutine
func startWorker() (cancel func()) {
    done := make(chan struct{})

    go func() {
        for {
            select {
            case <-done:
                return
            default:
                // 工作
            }
        }
    }()

    return func() { close(done) }
}
```

### 2. 使用 WaitGroup 等待完成

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        process(i)
    }(item)
}
wg.Wait()
```

### 3. 传递参数而非捕获

```go
// 好
for i := 0; i < 10; i++ {
    go func(n int) {
        fmt.Println(n)
    }(i)
}
```

---

## 下一步

- [Channels](./02-channels.md) - 学习使用 channel 在 goroutine 之间通信
