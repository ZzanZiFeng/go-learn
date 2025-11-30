# Select

本文档介绍 Go 语言的 select 语句，用于处理多个 channel 操作。

## 目录

- [Select 基础](#select-基础)
- [Default 分支](#default-分支)
- [超时处理](#超时处理)
- [常见模式](#常见模式)
- [与 JavaScript 对比](#与-javascript-对比)

---

## Select 基础

### 什么是 Select

Select 让你可以同时等待多个 channel 操作，类似于 switch，但专门用于 channel。

```go
select {
case val := <-ch1:
    // ch1 可读时执行
case ch2 <- value:
    // ch2 可写时执行
case val, ok := <-ch3:
    // ch3 可读时执行，ok 表示是否关闭
}
```

### 基本示例

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "one"
    }()

    go func() {
        time.Sleep(200 * time.Millisecond)
        ch2 <- "two"
    }()

    // 等待两个 channel
    for i := 0; i < 2; i++ {
        select {
        case msg1 := <-ch1:
            fmt.Println("Received from ch1:", msg1)
        case msg2 := <-ch2:
            fmt.Println("Received from ch2:", msg2)
        }
    }
}
```

**输出**:
```
Received from ch1: one
Received from ch2: two
```

### 随机选择

当多个 case 同时就绪时，select 会随机选择一个执行：

```go
package main

import "fmt"

func main() {
    ch1 := make(chan int, 1)
    ch2 := make(chan int, 1)

    ch1 <- 1
    ch2 <- 2

    // 两个都就绪，随机选择
    select {
    case val := <-ch1:
        fmt.Println("ch1:", val)
    case val := <-ch2:
        fmt.Println("ch2:", val)
    }
}
```

---

## Default 分支

### 非阻塞操作

default 分支在没有其他 case 就绪时立即执行：

```go
package main

import "fmt"

func main() {
    ch := make(chan int)

    // 非阻塞接收
    select {
    case val := <-ch:
        fmt.Println("Received:", val)
    default:
        fmt.Println("No value available")
    }
}
```

**输出**: `No value available`

### 非阻塞发送

```go
package main

import "fmt"

func main() {
    ch := make(chan int, 1)
    ch <- 1 // 填满缓冲

    // 非阻塞发送
    select {
    case ch <- 2:
        fmt.Println("Sent")
    default:
        fmt.Println("Channel full")
    }
}
```

**输出**: `Channel full`

### 轮询模式

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan int, 1)

    // 模拟生产者
    go func() {
        time.Sleep(500 * time.Millisecond)
        ch <- 42
    }()

    // 轮询等待
    for {
        select {
        case val := <-ch:
            fmt.Println("Received:", val)
            return
        default:
            fmt.Println("Waiting...")
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

---

## 超时处理

### 使用 time.After

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan string)

    go func() {
        time.Sleep(2 * time.Second) // 模拟慢操作
        ch <- "result"
    }()

    select {
    case result := <-ch:
        fmt.Println("Result:", result)
    case <-time.After(1 * time.Second):
        fmt.Println("Timeout!")
    }
}
```

**输出**: `Timeout!`

### 使用 time.Ticker

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    done := make(chan bool)
    go func() {
        time.Sleep(2 * time.Second)
        done <- true
    }()

    for {
        select {
        case <-done:
            fmt.Println("Done!")
            return
        case t := <-ticker.C:
            fmt.Println("Tick at", t.Format("15:04:05"))
        }
    }
}
```

---

## 常见模式

### 1. Done Channel 模式

```go
package main

import (
    "fmt"
    "time"
)

func worker(done <-chan struct{}) {
    for {
        select {
        case <-done:
            fmt.Println("Worker: received done signal")
            return
        default:
            fmt.Println("Worker: working...")
            time.Sleep(500 * time.Millisecond)
        }
    }
}

func main() {
    done := make(chan struct{})

    go worker(done)

    time.Sleep(2 * time.Second)
    close(done) // 发送停止信号
    time.Sleep(100 * time.Millisecond)
    fmt.Println("Main: exiting")
}
```

### 2. 多路复用

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

func generator(name string) <-chan string {
    ch := make(chan string)
    go func() {
        for i := 0; ; i++ {
            ch <- fmt.Sprintf("%s: %d", name, i)
            time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
        }
    }()
    return ch
}

func fanIn(ch1, ch2 <-chan string) <-chan string {
    ch := make(chan string)
    go func() {
        for {
            select {
            case val := <-ch1:
                ch <- val
            case val := <-ch2:
                ch <- val
            }
        }
    }()
    return ch
}

func main() {
    ch := fanIn(generator("A"), generator("B"))

    for i := 0; i < 10; i++ {
        fmt.Println(<-ch)
    }
}
```

### 3. 优先级选择

```go
package main

import "fmt"

func main() {
    highPriority := make(chan string, 10)
    lowPriority := make(chan string, 10)

    // 填充队列
    lowPriority <- "low 1"
    lowPriority <- "low 2"
    highPriority <- "HIGH 1"
    lowPriority <- "low 3"

    // 优先处理高优先级
    for i := 0; i < 4; i++ {
        select {
        case msg := <-highPriority:
            fmt.Println("High priority:", msg)
        default:
            select {
            case msg := <-highPriority:
                fmt.Println("High priority:", msg)
            case msg := <-lowPriority:
                fmt.Println("Low priority:", msg)
            }
        }
    }
}
```

### 4. 心跳模式

```go
package main

import (
    "fmt"
    "time"
)

func workerWithHeartbeat(done <-chan struct{}) <-chan struct{} {
    heartbeat := make(chan struct{})

    go func() {
        defer close(heartbeat)

        ticker := time.NewTicker(500 * time.Millisecond)
        defer ticker.Stop()

        for {
            select {
            case <-done:
                return
            case <-ticker.C:
                select {
                case heartbeat <- struct{}{}:
                default:
                    // 如果没人接收心跳，跳过
                }
            }
        }
    }()

    return heartbeat
}

func main() {
    done := make(chan struct{})
    heartbeat := workerWithHeartbeat(done)

    timeout := time.After(3 * time.Second)
    for {
        select {
        case <-heartbeat:
            fmt.Println("Worker is alive")
        case <-timeout:
            close(done)
            fmt.Println("Shutting down")
            return
        }
    }
}
```

---

## 与 JavaScript 对比

### Promise.race

```javascript
// JavaScript - Promise.race
const result = await Promise.race([
    fetch(url1),
    fetch(url2),
    new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Timeout')), 5000)
    )
]);
```

```go
// Go - select with timeout
func fetchWithTimeout(url string, timeout time.Duration) (string, error) {
    ch := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        resp, err := http.Get(url)
        if err != nil {
            errCh <- err
            return
        }
        defer resp.Body.Close()
        body, _ := io.ReadAll(resp.Body)
        ch <- string(body)
    }()

    select {
    case result := <-ch:
        return result, nil
    case err := <-errCh:
        return "", err
    case <-time.After(timeout):
        return "", errors.New("timeout")
    }
}
```

### 事件监听

```javascript
// JavaScript - EventEmitter
emitter.on('data', handleData);
emitter.on('error', handleError);
emitter.on('end', handleEnd);
```

```go
// Go - select on multiple channels
func eventLoop(dataCh, errCh <-chan Event, doneCh <-chan struct{}) {
    for {
        select {
        case data := <-dataCh:
            handleData(data)
        case err := <-errCh:
            handleError(err)
        case <-doneCh:
            return
        }
    }
}
```

---

## 注意事项

### 1. 空 Select 永远阻塞

```go
select {} // 永远阻塞，常用于保持 main 运行
```

### 2. nil Channel 永不就绪

```go
var ch chan int // nil

select {
case <-ch:
    // 永远不会执行
default:
    fmt.Println("ch is nil, executing default")
}
```

### 3. 关闭的 Channel 总是就绪

```go
ch := make(chan int)
close(ch)

select {
case val, ok := <-ch:
    fmt.Println(val, ok) // 0 false（立即返回零值）
}
```

---

## 下一步

- [Sync 包](./05-sync-package.md) - 学习使用同步原语
