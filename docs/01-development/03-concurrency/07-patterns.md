# 并发模式

本文档介绍 Go 语言中常见的并发设计模式。

## 目录

- [Worker Pool](#worker-pool)
- [Fan-out/Fan-in](#fan-outfan-in)
- [Pipeline](#pipeline)
- [Generator](#generator)
- [Pub/Sub](#pubsub)
- [Rate Limiting](#rate-limiting)

---

## Worker Pool

### 基本实现

Worker Pool 限制并发 goroutine 数量，避免资源耗尽。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(100 * time.Millisecond) // 模拟工作
        results <- job * 2
    }
}

func main() {
    const numWorkers = 3
    const numJobs = 10

    jobs := make(chan int, numJobs)
    results := make(chan int, numJobs)

    var wg sync.WaitGroup

    // 启动 workers
    for i := 1; i <= numWorkers; i++ {
        wg.Add(1)
        go worker(i, jobs, results, &wg)
    }

    // 发送任务
    for j := 1; j <= numJobs; j++ {
        jobs <- j
    }
    close(jobs)

    // 等待所有 worker 完成
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

### 带 Context 的 Worker Pool

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type Job struct {
    ID   int
    Data string
}

type Result struct {
    JobID  int
    Output string
    Err    error
}

func workerWithContext(ctx context.Context, id int, jobs <-chan Job, results chan<- Result) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Worker %d: shutting down\n", id)
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            // 处理任务
            time.Sleep(100 * time.Millisecond)
            results <- Result{
                JobID:  job.ID,
                Output: fmt.Sprintf("Processed: %s", job.Data),
            }
        }
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    jobs := make(chan Job, 100)
    results := make(chan Result, 100)

    // 启动 workers
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            workerWithContext(ctx, id, jobs, results)
        }(i)
    }

    // 发送任务
    go func() {
        for i := 0; i < 50; i++ {
            jobs <- Job{ID: i, Data: fmt.Sprintf("job-%d", i)}
        }
        close(jobs)
    }()

    // 收集结果
    go func() {
        wg.Wait()
        close(results)
    }()

    for result := range results {
        fmt.Printf("Job %d: %s\n", result.JobID, result.Output)
    }
}
```

---

## Fan-out/Fan-in

### Fan-out: 分发任务

```go
package main

import (
    "fmt"
    "sync"
)

func fanOut(input <-chan int, numWorkers int) []<-chan int {
    outputs := make([]<-chan int, numWorkers)

    for i := 0; i < numWorkers; i++ {
        outputs[i] = worker(input)
    }

    return outputs
}

func worker(input <-chan int) <-chan int {
    output := make(chan int)
    go func() {
        defer close(output)
        for n := range input {
            output <- n * n // 处理
        }
    }()
    return output
}

func main() {
    input := make(chan int)

    // 发送数据
    go func() {
        for i := 1; i <= 10; i++ {
            input <- i
        }
        close(input)
    }()

    // Fan-out 到 3 个 worker
    outputs := fanOut(input, 3)

    // 收集结果
    var wg sync.WaitGroup
    for _, ch := range outputs {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for result := range c {
                fmt.Println("Result:", result)
            }
        }(ch)
    }
    wg.Wait()
}
```

### Fan-in: 合并结果

```go
package main

import (
    "fmt"
    "sync"
)

func fanIn(inputs ...<-chan int) <-chan int {
    output := make(chan int)

    var wg sync.WaitGroup
    for _, ch := range inputs {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                output <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(output)
    }()

    return output
}

func generator(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func main() {
    // 创建多个输入
    ch1 := generator(1, 2, 3)
    ch2 := generator(4, 5, 6)
    ch3 := generator(7, 8, 9)

    // Fan-in 合并
    merged := fanIn(ch1, ch2, ch3)

    for val := range merged {
        fmt.Println(val)
    }
}
```

---

## Pipeline

### 阶段处理

Pipeline 将处理分成多个阶段，每个阶段是独立的 goroutine。

```go
package main

import "fmt"

// 阶段 1: 生成数字
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

// 阶段 2: 平方
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

// 阶段 3: 过滤
func filter(in <-chan int, predicate func(int) bool) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            if predicate(n) {
                out <- n
            }
        }
    }()
    return out
}

// 阶段 4: 格式化
func format(in <-chan int) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for n := range in {
            out <- fmt.Sprintf("Result: %d", n)
        }
    }()
    return out
}

func main() {
    // 构建 pipeline
    numbers := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
    squared := square(numbers)
    filtered := filter(squared, func(n int) bool { return n > 20 })
    formatted := format(filtered)

    // 消费结果
    for s := range formatted {
        fmt.Println(s)
    }
}
```

---

## Generator

### 惰性生成

```go
package main

import "fmt"

// 无限生成器
func infiniteCounter() <-chan int {
    ch := make(chan int)
    go func() {
        for i := 0; ; i++ {
            ch <- i
        }
    }()
    return ch
}

// 斐波那契生成器
func fibonacci() <-chan int {
    ch := make(chan int)
    go func() {
        a, b := 0, 1
        for {
            ch <- a
            a, b = b, a+b
        }
    }()
    return ch
}

// 带取消的生成器
func countWithCancel(done <-chan struct{}) <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; ; i++ {
            select {
            case <-done:
                return
            case ch <- i:
            }
        }
    }()
    return ch
}

func main() {
    // 取前 10 个斐波那契数
    fib := fibonacci()
    for i := 0; i < 10; i++ {
        fmt.Println(<-fib)
    }

    // 带取消的计数器
    done := make(chan struct{})
    counter := countWithCancel(done)

    for i := 0; i < 5; i++ {
        fmt.Println(<-counter)
    }
    close(done) // 停止生成器
}
```

---

## Pub/Sub

### 发布/订阅模式

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type Broker struct {
    mu          sync.RWMutex
    subscribers map[string][]chan string
}

func NewBroker() *Broker {
    return &Broker{
        subscribers: make(map[string][]chan string),
    }
}

func (b *Broker) Subscribe(topic string) <-chan string {
    b.mu.Lock()
    defer b.mu.Unlock()

    ch := make(chan string, 10)
    b.subscribers[topic] = append(b.subscribers[topic], ch)
    return ch
}

func (b *Broker) Publish(topic, msg string) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    for _, ch := range b.subscribers[topic] {
        select {
        case ch <- msg:
        default:
            // 订阅者缓冲满了，跳过
        }
    }
}

func (b *Broker) Close(topic string) {
    b.mu.Lock()
    defer b.mu.Unlock()

    for _, ch := range b.subscribers[topic] {
        close(ch)
    }
    delete(b.subscribers, topic)
}

func main() {
    broker := NewBroker()

    // 订阅者 1
    sub1 := broker.Subscribe("news")
    go func() {
        for msg := range sub1 {
            fmt.Println("Sub1 received:", msg)
        }
    }()

    // 订阅者 2
    sub2 := broker.Subscribe("news")
    go func() {
        for msg := range sub2 {
            fmt.Println("Sub2 received:", msg)
        }
    }()

    // 发布消息
    time.Sleep(100 * time.Millisecond)
    broker.Publish("news", "Breaking: Go is awesome!")
    broker.Publish("news", "Update: More awesome features!")

    time.Sleep(100 * time.Millisecond)
    broker.Close("news")
    time.Sleep(100 * time.Millisecond)
}
```

---

## Rate Limiting

### 令牌桶

```go
package main

import (
    "fmt"
    "time"
)

type RateLimiter struct {
    tokens chan struct{}
}

func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, rate),
    }

    // 预填充
    for i := 0; i < rate; i++ {
        rl.tokens <- struct{}{}
    }

    // 定期补充令牌
    go func() {
        ticker := time.NewTicker(interval / time.Duration(rate))
        defer ticker.Stop()
        for range ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default:
                // 桶满了
            }
        }
    }()

    return rl
}

func (rl *RateLimiter) Allow() bool {
    select {
    case <-rl.tokens:
        return true
    default:
        return false
    }
}

func (rl *RateLimiter) Wait() {
    <-rl.tokens
}

func main() {
    // 每秒 5 个请求
    limiter := NewRateLimiter(5, time.Second)

    for i := 0; i < 10; i++ {
        if limiter.Allow() {
            fmt.Printf("Request %d: allowed at %s\n", i, time.Now().Format("15:04:05.000"))
        } else {
            fmt.Printf("Request %d: rate limited\n", i)
        }
    }

    fmt.Println("\n--- With Wait ---")
    for i := 0; i < 5; i++ {
        limiter.Wait()
        fmt.Printf("Request %d: processed at %s\n", i, time.Now().Format("15:04:05.000"))
    }
}
```

### 滑动窗口

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type SlidingWindow struct {
    mu       sync.Mutex
    window   time.Duration
    limit    int
    requests []time.Time
}

func NewSlidingWindow(window time.Duration, limit int) *SlidingWindow {
    return &SlidingWindow{
        window:   window,
        limit:    limit,
        requests: make([]time.Time, 0),
    }
}

func (sw *SlidingWindow) Allow() bool {
    sw.mu.Lock()
    defer sw.mu.Unlock()

    now := time.Now()
    windowStart := now.Add(-sw.window)

    // 清理过期请求
    valid := make([]time.Time, 0)
    for _, t := range sw.requests {
        if t.After(windowStart) {
            valid = append(valid, t)
        }
    }
    sw.requests = valid

    // 检查限制
    if len(sw.requests) >= sw.limit {
        return false
    }

    sw.requests = append(sw.requests, now)
    return true
}

func main() {
    // 1 秒内最多 3 个请求
    sw := NewSlidingWindow(time.Second, 3)

    for i := 0; i < 10; i++ {
        if sw.Allow() {
            fmt.Printf("Request %d: allowed\n", i)
        } else {
            fmt.Printf("Request %d: rate limited\n", i)
        }
        time.Sleep(200 * time.Millisecond)
    }
}
```

---

## 总结

| 模式 | 用途 | 何时使用 |
|-----|------|---------|
| Worker Pool | 限制并发数 | 资源有限，任务数量大 |
| Fan-out/Fan-in | 并行处理 | CPU 密集型任务 |
| Pipeline | 流式处理 | 数据转换链 |
| Generator | 惰性生成 | 无限序列，按需生成 |
| Pub/Sub | 事件广播 | 一对多通信 |
| Rate Limiting | 流量控制 | API 限流，防止过载 |

---

## 下一步

- [竞态检测](./08-race-detection.md) - 学习检测和避免竞态条件
