# 缓冲 Channel

本文档介绍 Go 语言的缓冲 channel。

## 目录

- [缓冲 Channel 基础](#缓冲-channel-基础)
- [缓冲 vs 无缓冲](#缓冲-vs-无缓冲)
- [缓冲 Channel 的使用场景](#缓冲-channel-的使用场景)
- [容量和长度](#容量和长度)
- [常见模式](#常见模式)

---

## 缓冲 Channel 基础

### 创建缓冲 Channel

```go
ch := make(chan int, 10) // 创建容量为 10 的缓冲 channel
```

### 基本特性

| 特性 | 无缓冲 | 有缓冲 |
|-----|-------|-------|
| 创建 | `make(chan T)` | `make(chan T, n)` |
| 发送阻塞 | 等待接收者 | 缓冲满时阻塞 |
| 接收阻塞 | 等待发送者 | 缓冲空时阻塞 |
| 同步性 | 强同步 | 弱同步 |

### 基本示例

```go
package main

import "fmt"

func main() {
    // 创建容量为 3 的缓冲 channel
    ch := make(chan int, 3)

    // 可以连续发送 3 个值，不会阻塞
    ch <- 1
    ch <- 2
    ch <- 3

    fmt.Println("Sent 3 values")

    // 接收值
    fmt.Println(<-ch) // 1
    fmt.Println(<-ch) // 2
    fmt.Println(<-ch) // 3
}
```

---

## 缓冲 vs 无缓冲

### 无缓冲 Channel（同步）

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int) // 无缓冲

    go func() {
        fmt.Println("Sending...")
        ch <- 1 // 阻塞，直到有人接收
        fmt.Println("Sent!")
    }()

    time.Sleep(time.Second)
    fmt.Println("Receiving...")
    <-ch
    time.Sleep(100 * time.Millisecond)
}
```

**输出**:
```
Sending...
Receiving...
Sent!
```

### 缓冲 Channel（异步）

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int, 1) // 容量为 1

    go func() {
        fmt.Println("Sending...")
        ch <- 1 // 不阻塞，因为缓冲未满
        fmt.Println("Sent!")
    }()

    time.Sleep(time.Second)
    fmt.Println("Receiving...")
    <-ch
}
```

**输出**:
```
Sending...
Sent!
Receiving...
```

### 选择指南

| 场景 | 推荐 |
|-----|-----|
| 需要严格同步 | 无缓冲 |
| 生产者/消费者解耦 | 有缓冲 |
| 限制并发数量 | 有缓冲（信号量） |
| 突发流量处理 | 有缓冲 |

---

## 缓冲 Channel 的使用场景

### 1. 异步日志

```go
package main

import (
    "fmt"
    "time"
)

type Logger struct {
    logs chan string
}

func NewLogger(bufferSize int) *Logger {
    l := &Logger{
        logs: make(chan string, bufferSize),
    }
    go l.writer()
    return l
}

func (l *Logger) Log(msg string) {
    select {
    case l.logs <- msg:
        // 成功发送
    default:
        // 缓冲满了，丢弃或处理
        fmt.Println("Log buffer full, dropping:", msg)
    }
}

func (l *Logger) writer() {
    for msg := range l.logs {
        // 模拟写入磁盘
        fmt.Println("[LOG]", msg)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    logger := NewLogger(10)

    for i := 0; i < 15; i++ {
        logger.Log(fmt.Sprintf("Message %d", i))
    }

    time.Sleep(2 * time.Second)
}
```

### 2. 工作队列

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    jobs := make(chan int, 100)    // 任务队列
    results := make(chan int, 100) // 结果队列

    // 启动 3 个 worker
    var wg sync.WaitGroup
    for w := 1; w <= 3; w++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for job := range jobs {
                fmt.Printf("Worker %d processing job %d\n", id, job)
                time.Sleep(100 * time.Millisecond)
                results <- job * 2
            }
        }(w)
    }

    // 发送任务
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)

    // 等待 worker 完成
    go func() {
        wg.Wait()
        close(results)
    }()

    // 收集结果
    for result := range results {
        fmt.Println("Result:", result)
    }
}
```

### 3. 速率限制

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // 令牌桶：每 200ms 一个令牌，最多存 3 个
    limiter := make(chan time.Time, 3)

    // 预填充
    for i := 0; i < 3; i++ {
        limiter <- time.Now()
    }

    // 定期添加令牌
    go func() {
        for t := range time.Tick(200 * time.Millisecond) {
            select {
            case limiter <- t:
            default: // 桶满了，丢弃令牌
            }
        }
    }()

    // 模拟请求
    requests := make(chan int, 5)
    for i := 1; i <= 5; i++ {
        requests <- i
    }
    close(requests)

    for req := range requests {
        <-limiter // 获取令牌
        fmt.Printf("Request %d at %s\n", req, time.Now().Format("15:04:05.000"))
    }
}
```

---

## 容量和长度

### 检查容量和长度

```go
package main

import "fmt"

func main() {
    ch := make(chan int, 5)

    ch <- 1
    ch <- 2
    ch <- 3

    fmt.Println("Capacity:", cap(ch)) // 5
    fmt.Println("Length:", len(ch))   // 3

    <-ch

    fmt.Println("After receive:")
    fmt.Println("Capacity:", cap(ch)) // 5
    fmt.Println("Length:", len(ch))   // 2
}
```

### 检查缓冲状态

```go
func isFull(ch chan int) bool {
    return len(ch) == cap(ch)
}

func isEmpty(ch chan int) bool {
    return len(ch) == 0
}
```

---

## 常见模式

### 1. 非阻塞发送

```go
ch := make(chan int, 1)

select {
case ch <- value:
    fmt.Println("Sent")
default:
    fmt.Println("Channel full, not sent")
}
```

### 2. 非阻塞接收

```go
select {
case val := <-ch:
    fmt.Println("Received:", val)
default:
    fmt.Println("No value available")
}
```

### 3. 超时发送

```go
select {
case ch <- value:
    fmt.Println("Sent")
case <-time.After(time.Second):
    fmt.Println("Send timeout")
}
```

### 4. 信号量（限制并发）

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    // 最多 2 个并发
    sem := make(chan struct{}, 2)
    var wg sync.WaitGroup

    for i := 1; i <= 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            sem <- struct{}{}        // 获取
            defer func() { <-sem }() // 释放

            fmt.Printf("Worker %d started at %s\n", id, time.Now().Format("15:04:05"))
            time.Sleep(time.Second)
            fmt.Printf("Worker %d finished\n", id)
        }(i)
    }

    wg.Wait()
}
```

---

## 最佳实践

### 1. 选择合适的缓冲大小

```go
// 太小：频繁阻塞
ch := make(chan int, 1)

// 太大：浪费内存，隐藏问题
ch := make(chan int, 1000000)

// 适当：根据预期负载选择
ch := make(chan int, 100) // 考虑生产者/消费者速度
```

### 2. 考虑背压

```go
// 好：有背压机制
func producer(ch chan<- int) {
    for i := 0; ; i++ {
        ch <- i // 缓冲满时会阻塞，形成背压
    }
}

// 危险：无限缓冲（可能 OOM）
func producer(ch chan<- int) {
    for i := 0; ; i++ {
        select {
        case ch <- i:
        default:
            // 丢弃或存储到磁盘
        }
    }
}
```

### 3. 文档化缓冲大小的选择

```go
// jobQueue 缓冲大小为 100，基于：
// - 平均处理时间: 100ms
// - 期望吞吐量: 1000 jobs/sec
// - 允许的最大延迟: 10s
jobQueue := make(chan Job, 100)
```

---

## 下一步

- [Select](./04-select.md) - 学习使用 select 进行多路复用
