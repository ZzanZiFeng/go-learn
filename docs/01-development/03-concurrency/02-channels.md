# Channels

本文档介绍 Go 语言的 channel，用于 goroutine 之间的通信。

## 目录

- [Channel 基础](#channel-基础)
- [发送和接收](#发送和接收)
- [Channel 方向](#channel-方向)
- [关闭 Channel](#关闭-channel)
- [遍历 Channel](#遍历-channel)
- [Channel 作为同步工具](#channel-作为同步工具)
- [与 JavaScript 对比](#与-javascript-对比)

---

## Channel 基础

### 什么是 Channel

Channel 是 Go 中 goroutine 之间通信的管道。

```go
// 创建 channel
ch := make(chan int)     // 创建 int 类型的 channel
ch := make(chan string)  // 创建 string 类型的 channel
ch := make(chan bool)    // 创建 bool 类型的 channel
```

### Channel 特性

| 特性 | 说明 |
|-----|------|
| 类型安全 | channel 只能传递指定类型的值 |
| 阻塞 | 发送/接收会阻塞直到另一端准备好 |
| 线程安全 | 多个 goroutine 可以安全地使用同一个 channel |
| 有方向 | 可以限制只发送或只接收 |

### 基本示例

```go
package main

import "fmt"

func main() {
    // 创建 channel
    ch := make(chan string)

    // 在 goroutine 中发送
    go func() {
        ch <- "Hello from goroutine!"
    }()

    // 在 main 中接收
    msg := <-ch
    fmt.Println(msg) // Hello from goroutine!
}
```

---

## 发送和接收

### 发送操作

```go
ch <- value  // 发送 value 到 channel
```

### 接收操作

```go
value := <-ch   // 从 channel 接收并赋值
<-ch            // 从 channel 接收并丢弃
```

### 阻塞行为

无缓冲 channel 的发送和接收都会阻塞：

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int)

    // 发送者
    go func() {
        fmt.Println("Sending...")
        ch <- 42 // 阻塞，直到有人接收
        fmt.Println("Sent!")
    }()

    time.Sleep(time.Second) // 让发送者先运行

    // 接收者
    fmt.Println("Receiving...")
    val := <-ch // 接收
    fmt.Println("Received:", val)
}
```

**输出**:
```
Sending...
Receiving...
Sent!
Received: 42
```

### 死锁示例

```go
// 死锁：没有人接收
func main() {
    ch := make(chan int)
    ch <- 42 // 永远阻塞
    // fatal error: all goroutines are asleep - deadlock!
}

// 死锁：没有人发送
func main() {
    ch := make(chan int)
    val := <-ch // 永远阻塞
    fmt.Println(val)
    // fatal error: all goroutines are asleep - deadlock!
}
```

---

## Channel 方向

### 双向 Channel

```go
ch := make(chan int) // 可以发送和接收
```

### 只发送 Channel

```go
func sender(ch chan<- int) {
    ch <- 42     // OK
    // val := <-ch  // 编译错误：不能接收
}
```

### 只接收 Channel

```go
func receiver(ch <-chan int) {
    val := <-ch  // OK
    // ch <- 42     // 编译错误：不能发送
}
```

### 实际应用

```go
package main

import "fmt"

// 生产者：只能发送
func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)
}

// 消费者：只能接收
func consumer(ch <-chan int) {
    for val := range ch {
        fmt.Println("Received:", val)
    }
}

func main() {
    ch := make(chan int)

    go producer(ch)
    consumer(ch) // 在 main goroutine 中消费
}
```

---

## 关闭 Channel

### 关闭语法

```go
close(ch) // 关闭 channel
```

### 关闭的影响

| 操作 | 已关闭的 Channel |
|-----|------------------|
| 发送 | panic |
| 接收 | 返回零值，ok = false |
| 再次关闭 | panic |

### 检测 Channel 是否关闭

```go
package main

import "fmt"

func main() {
    ch := make(chan int, 2)
    ch <- 1
    ch <- 2
    close(ch)

    // 使用 comma-ok 模式检测
    for {
        val, ok := <-ch
        if !ok {
            fmt.Println("Channel closed")
            break
        }
        fmt.Println("Received:", val)
    }
}
```

**输出**:
```
Received: 1
Received: 2
Channel closed
```

### 谁应该关闭 Channel

**规则**：只有发送方应该关闭 channel

```go
func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // 发送方关闭
}

func consumer(ch <-chan int) {
    for val := range ch {
        fmt.Println(val)
    }
    // 不要在接收方关闭
}
```

---

## 遍历 Channel

### 使用 for range

```go
package main

import "fmt"

func main() {
    ch := make(chan int)

    go func() {
        for i := 1; i <= 5; i++ {
            ch <- i
        }
        close(ch) // 必须关闭，否则 range 会永远等待
    }()

    // range 自动检测 channel 关闭
    for val := range ch {
        fmt.Println(val)
    }
    fmt.Println("Done")
}
```

**输出**:
```
1
2
3
4
5
Done
```

### 不关闭会怎样

```go
// 死锁！
func main() {
    ch := make(chan int)

    go func() {
        ch <- 1
        ch <- 2
        // 忘记 close(ch)
    }()

    for val := range ch {
        fmt.Println(val)
    }
    // range 永远等待，导致死锁
}
```

---

## Channel 作为同步工具

### 等待完成

```go
package main

import "fmt"

func worker(done chan<- bool) {
    fmt.Println("Working...")
    // 做一些工作
    fmt.Println("Done working")
    done <- true // 通知完成
}

func main() {
    done := make(chan bool)

    go worker(done)

    <-done // 等待完成信号
    fmt.Println("Worker finished")
}
```

### 使用空结构体节省内存

```go
done := make(chan struct{})

go func() {
    // 工作
    close(done) // 发送完成信号
}()

<-done // 等待
```

### 信号量模式

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    // 限制同时运行的 goroutine 数量
    semaphore := make(chan struct{}, 3) // 最多 3 个并发

    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            semaphore <- struct{}{} // 获取信号量
            defer func() { <-semaphore }() // 释放信号量

            fmt.Printf("Worker %d running\n", id)
            // 模拟工作
        }(i)
    }

    wg.Wait()
}
```

---

## 与 JavaScript 对比

### 异步通信对比

```javascript
// JavaScript - 使用 Promise/callback
function fetchData() {
    return new Promise(resolve => {
        setTimeout(() => {
            resolve("data");
        }, 1000);
    });
}

fetchData().then(data => console.log(data));
```

```go
// Go - 使用 Channel
func fetchData(ch chan<- string) {
    time.Sleep(time.Second)
    ch <- "data"
}

ch := make(chan string)
go fetchData(ch)
data := <-ch
fmt.Println(data)
```

### 多个异步操作

```javascript
// JavaScript - Promise.all
const results = await Promise.all([
    fetch(url1),
    fetch(url2),
]);
```

```go
// Go - 多个 channel
ch1 := make(chan Result)
ch2 := make(chan Result)

go fetch(url1, ch1)
go fetch(url2, ch2)

result1 := <-ch1
result2 := <-ch2
```

### 流处理对比

```javascript
// JavaScript - async generator
async function* numberGenerator() {
    for (let i = 0; i < 5; i++) {
        yield i;
    }
}

for await (const num of numberGenerator()) {
    console.log(num);
}
```

```go
// Go - channel + range
func numberGenerator(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)
}

ch := make(chan int)
go numberGenerator(ch)
for num := range ch {
    fmt.Println(num)
}
```

### 主要差异

| 特性 | JavaScript | Go Channel |
|-----|-----------|------------|
| 模型 | Promise/async | CSP |
| 阻塞 | 非阻塞 | 默认阻塞 |
| 缓冲 | 无（Promise 立即解析） | 可选缓冲 |
| 多值 | 需要包装 | 原生支持流 |
| 取消 | AbortController | Context |

---

## 最佳实践

### 1. 明确 Channel 所有权

```go
// 好：创建者负责关闭
func producer() <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()
    return ch
}
```

### 2. 使用方向限制

```go
// 好：明确函数只发送或只接收
func send(ch chan<- int) { ... }
func recv(ch <-chan int) { ... }
```

### 3. 防止 Goroutine 泄漏

```go
// 好：使用 context 取消
func worker(ctx context.Context, ch <-chan int) {
    for {
        select {
        case <-ctx.Done():
            return
        case val := <-ch:
            process(val)
        }
    }
}
```

---

## 下一步

- [缓冲 Channel](./03-buffered-chan.md) - 学习有缓冲的 channel
