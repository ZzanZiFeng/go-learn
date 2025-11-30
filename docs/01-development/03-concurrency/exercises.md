# 并发编程练习

## 练习 1: 并行计算 (Parallel Computation)

**难度**: 基础

实现一个并行求和函数，将大数组分成多个部分并行计算。

```go
// ParallelSum 并行计算切片元素之和
// workers 指定并行工作者数量
func ParallelSum(nums []int, workers int) int {
    // TODO: 实现并行求和
}

// 测试
func main() {
    nums := make([]int, 1000000)
    for i := range nums {
        nums[i] = i + 1
    }

    sum := ParallelSum(nums, 4)
    fmt.Println("Sum:", sum)
    // 期望: 500000500000
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "fmt"
    "sync"
)

func ParallelSum(nums []int, workers int) int {
    if len(nums) == 0 {
        return 0
    }

    // 每个 worker 处理的大小
    chunkSize := (len(nums) + workers - 1) / workers

    results := make(chan int, workers)
    var wg sync.WaitGroup

    for i := 0; i < workers; i++ {
        start := i * chunkSize
        end := start + chunkSize
        if end > len(nums) {
            end = len(nums)
        }
        if start >= len(nums) {
            break
        }

        wg.Add(1)
        go func(slice []int) {
            defer wg.Done()
            sum := 0
            for _, n := range slice {
                sum += n
            }
            results <- sum
        }(nums[start:end])
    }

    // 关闭 results channel
    go func() {
        wg.Wait()
        close(results)
    }()

    // 汇总结果
    total := 0
    for partialSum := range results {
        total += partialSum
    }

    return total
}

func main() {
    nums := make([]int, 1000000)
    for i := range nums {
        nums[i] = i + 1
    }

    sum := ParallelSum(nums, 4)
    fmt.Println("Sum:", sum)
}
```

</details>

---

## 练习 2: 带超时的 Worker Pool

**难度**: 中级

实现一个带超时控制的 Worker Pool，如果任务执行超时则取消。

```go
type Task struct {
    ID       int
    Duration time.Duration // 模拟执行时间
}

type Result struct {
    TaskID  int
    Success bool
    Error   string
}

// WorkerPool 处理任务，单个任务超时时间为 timeout
func WorkerPool(tasks []Task, workers int, timeout time.Duration) []Result {
    // TODO: 实现带超时的 worker pool
}

func main() {
    tasks := []Task{
        {ID: 1, Duration: 100 * time.Millisecond},
        {ID: 2, Duration: 500 * time.Millisecond}, // 超时
        {ID: 3, Duration: 50 * time.Millisecond},
        {ID: 4, Duration: 300 * time.Millisecond}, // 超时
        {ID: 5, Duration: 80 * time.Millisecond},
    }

    results := WorkerPool(tasks, 3, 200*time.Millisecond)
    for _, r := range results {
        fmt.Printf("Task %d: success=%v, error=%s\n", r.TaskID, r.Success, r.Error)
    }
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type Task struct {
    ID       int
    Duration time.Duration
}

type Result struct {
    TaskID  int
    Success bool
    Error   string
}

func WorkerPool(tasks []Task, workers int, timeout time.Duration) []Result {
    taskCh := make(chan Task, len(tasks))
    resultCh := make(chan Result, len(tasks))

    var wg sync.WaitGroup

    // 启动 workers
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range taskCh {
                ctx, cancel := context.WithTimeout(context.Background(), timeout)

                // 执行任务
                done := make(chan struct{})
                go func() {
                    time.Sleep(task.Duration) // 模拟工作
                    close(done)
                }()

                select {
                case <-done:
                    resultCh <- Result{
                        TaskID:  task.ID,
                        Success: true,
                    }
                case <-ctx.Done():
                    resultCh <- Result{
                        TaskID:  task.ID,
                        Success: false,
                        Error:   "timeout",
                    }
                }
                cancel()
            }
        }()
    }

    // 发送任务
    for _, task := range tasks {
        taskCh <- task
    }
    close(taskCh)

    // 等待完成并关闭结果通道
    go func() {
        wg.Wait()
        close(resultCh)
    }()

    // 收集结果
    results := make([]Result, 0, len(tasks))
    for r := range resultCh {
        results = append(results, r)
    }

    return results
}

func main() {
    tasks := []Task{
        {ID: 1, Duration: 100 * time.Millisecond},
        {ID: 2, Duration: 500 * time.Millisecond},
        {ID: 3, Duration: 50 * time.Millisecond},
        {ID: 4, Duration: 300 * time.Millisecond},
        {ID: 5, Duration: 80 * time.Millisecond},
    }

    results := WorkerPool(tasks, 3, 200*time.Millisecond)
    for _, r := range results {
        fmt.Printf("Task %d: success=%v, error=%s\n", r.TaskID, r.Success, r.Error)
    }
}
```

</details>

---

## 练习 3: 并发安全的缓存

**难度**: 中级

实现一个并发安全的缓存，支持过期时间。

```go
type Cache struct {
    // TODO: 定义字段
}

type cacheItem struct {
    value      interface{}
    expiration time.Time
}

func NewCache() *Cache {
    // TODO: 实现
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    // TODO: 实现
}

func (c *Cache) Get(key string) (interface{}, bool) {
    // TODO: 实现
}

func (c *Cache) Delete(key string) {
    // TODO: 实现
}

// 测试
func main() {
    cache := NewCache()

    cache.Set("user:1", "Alice", 100*time.Millisecond)

    if val, ok := cache.Get("user:1"); ok {
        fmt.Println("Found:", val)
    }

    time.Sleep(150 * time.Millisecond)

    if _, ok := cache.Get("user:1"); !ok {
        fmt.Println("Expired!")
    }
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type cacheItem struct {
    value      interface{}
    expiration time.Time
}

type Cache struct {
    mu    sync.RWMutex
    items map[string]cacheItem
}

func NewCache() *Cache {
    c := &Cache{
        items: make(map[string]cacheItem),
    }

    // 后台清理过期项
    go c.cleanupLoop()

    return c
}

func (c *Cache) cleanupLoop() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, item := range c.items {
            if !item.expiration.IsZero() && now.After(item.expiration) {
                delete(c.items, key)
            }
        }
        c.mu.Unlock()
    }
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    var expiration time.Time
    if ttl > 0 {
        expiration = time.Now().Add(ttl)
    }

    c.items[key] = cacheItem{
        value:      value,
        expiration: expiration,
    }
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, found := c.items[key]
    if !found {
        return nil, false
    }

    // 检查是否过期
    if !item.expiration.IsZero() && time.Now().After(item.expiration) {
        return nil, false
    }

    return item.value, true
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, key)
}

func main() {
    cache := NewCache()

    cache.Set("user:1", "Alice", 100*time.Millisecond)

    if val, ok := cache.Get("user:1"); ok {
        fmt.Println("Found:", val)
    }

    time.Sleep(150 * time.Millisecond)

    if _, ok := cache.Get("user:1"); !ok {
        fmt.Println("Expired!")
    }
}
```

</details>

---

## 练习 4: 并发限流器

**难度**: 中级

实现一个令牌桶限流器 (Token Bucket Rate Limiter)。

```go
type RateLimiter struct {
    // TODO: 定义字段
}

// NewRateLimiter 创建限流器
// rate: 每秒令牌数
// burst: 桶容量（最大突发量）
func NewRateLimiter(rate float64, burst int) *RateLimiter {
    // TODO: 实现
}

// Allow 检查是否允许一个请求
func (r *RateLimiter) Allow() bool {
    // TODO: 实现
}

// Wait 等待直到获得令牌
func (r *RateLimiter) Wait(ctx context.Context) error {
    // TODO: 实现
}

func main() {
    limiter := NewRateLimiter(10, 3) // 每秒10个，最大突发3个

    for i := 0; i < 10; i++ {
        if limiter.Allow() {
            fmt.Printf("Request %d: allowed at %s\n", i, time.Now().Format("15:04:05.000"))
        } else {
            fmt.Printf("Request %d: rejected\n", i)
        }
    }
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type RateLimiter struct {
    mu       sync.Mutex
    rate     float64       // 每秒令牌数
    burst    int           // 桶容量
    tokens   float64       // 当前令牌数
    lastTime time.Time     // 上次更新时间
}

func NewRateLimiter(rate float64, burst int) *RateLimiter {
    return &RateLimiter{
        rate:     rate,
        burst:    burst,
        tokens:   float64(burst), // 初始满桶
        lastTime: time.Now(),
    }
}

func (r *RateLimiter) Allow() bool {
    r.mu.Lock()
    defer r.mu.Unlock()

    // 补充令牌
    now := time.Now()
    elapsed := now.Sub(r.lastTime).Seconds()
    r.tokens += elapsed * r.rate
    if r.tokens > float64(r.burst) {
        r.tokens = float64(r.burst)
    }
    r.lastTime = now

    // 检查是否有令牌
    if r.tokens >= 1 {
        r.tokens--
        return true
    }

    return false
}

func (r *RateLimiter) Wait(ctx context.Context) error {
    for {
        if r.Allow() {
            return nil
        }

        // 计算等待时间
        r.mu.Lock()
        waitTime := time.Duration((1 - r.tokens) / r.rate * float64(time.Second))
        r.mu.Unlock()

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(waitTime):
            // 继续尝试
        }
    }
}

func main() {
    limiter := NewRateLimiter(10, 3)

    // 测试 Allow
    fmt.Println("=== Testing Allow ===")
    for i := 0; i < 10; i++ {
        if limiter.Allow() {
            fmt.Printf("Request %d: allowed at %s\n", i, time.Now().Format("15:04:05.000"))
        } else {
            fmt.Printf("Request %d: rejected\n", i)
        }
    }

    // 测试 Wait
    fmt.Println("\n=== Testing Wait ===")
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    for i := 0; i < 5; i++ {
        if err := limiter.Wait(ctx); err != nil {
            fmt.Printf("Request %d: error - %v\n", i, err)
            break
        }
        fmt.Printf("Request %d: allowed at %s\n", i, time.Now().Format("15:04:05.000"))
    }
}
```

</details>

---

## 练习 5: Pipeline 数据处理

**难度**: 中级

实现一个可取消的 Pipeline，处理数据流。

```go
// Pipeline 阶段函数类型
type Stage func(ctx context.Context, in <-chan int) <-chan int

// BuildPipeline 构建 pipeline
func BuildPipeline(ctx context.Context, source <-chan int, stages ...Stage) <-chan int {
    // TODO: 实现
}

// 示例阶段
func Filter(predicate func(int) bool) Stage {
    // TODO: 实现 - 只保留满足条件的元素
}

func Map(transform func(int) int) Stage {
    // TODO: 实现 - 转换每个元素
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // 数据源
    source := make(chan int)
    go func() {
        defer close(source)
        for i := 1; i <= 100; i++ {
            select {
            case source <- i:
            case <-ctx.Done():
                return
            }
        }
    }()

    // 构建 pipeline: 过滤偶数 -> 平方
    result := BuildPipeline(ctx, source,
        Filter(func(n int) bool { return n%2 == 0 }),
        Map(func(n int) int { return n * n }),
    )

    for n := range result {
        fmt.Println(n)
    }
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "context"
    "fmt"
    "time"
)

type Stage func(ctx context.Context, in <-chan int) <-chan int

func BuildPipeline(ctx context.Context, source <-chan int, stages ...Stage) <-chan int {
    current := source
    for _, stage := range stages {
        current = stage(ctx, current)
    }
    return current
}

func Filter(predicate func(int) bool) Stage {
    return func(ctx context.Context, in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for n := range in {
                if predicate(n) {
                    select {
                    case out <- n:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }()
        return out
    }
}

func Map(transform func(int) int) Stage {
    return func(ctx context.Context, in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for n := range in {
                select {
                case out <- transform(n):
                case <-ctx.Done():
                    return
                }
            }
        }()
        return out
    }
}

func Reduce(initial int, reducer func(acc, n int) int) Stage {
    return func(ctx context.Context, in <-chan int) <-chan int {
        out := make(chan int, 1)
        go func() {
            defer close(out)
            acc := initial
            for n := range in {
                select {
                case <-ctx.Done():
                    return
                default:
                    acc = reducer(acc, n)
                }
            }
            select {
            case out <- acc:
            case <-ctx.Done():
            }
        }()
        return out
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // 数据源
    source := make(chan int)
    go func() {
        defer close(source)
        for i := 1; i <= 100; i++ {
            select {
            case source <- i:
            case <-ctx.Done():
                return
            }
        }
    }()

    // 构建 pipeline
    result := BuildPipeline(ctx, source,
        Filter(func(n int) bool { return n%2 == 0 }),
        Map(func(n int) int { return n * n }),
    )

    for n := range result {
        fmt.Println(n) // 4, 16, 36, 64, 100, ...
    }
}
```

</details>

---

## 练习 6: 并发安全的发布订阅

**难度**: 高级

实现一个简单的发布订阅系统。

```go
type PubSub struct {
    // TODO: 定义字段
}

func NewPubSub() *PubSub {
    // TODO: 实现
}

// Subscribe 订阅主题，返回消息通道
func (ps *PubSub) Subscribe(topic string) <-chan string {
    // TODO: 实现
}

// Unsubscribe 取消订阅
func (ps *PubSub) Unsubscribe(topic string, ch <-chan string) {
    // TODO: 实现
}

// Publish 发布消息到主题
func (ps *PubSub) Publish(topic string, msg string) {
    // TODO: 实现
}

// Close 关闭所有订阅
func (ps *PubSub) Close() {
    // TODO: 实现
}

func main() {
    ps := NewPubSub()
    defer ps.Close()

    // 订阅者 1
    sub1 := ps.Subscribe("news")
    go func() {
        for msg := range sub1 {
            fmt.Println("Sub1 received:", msg)
        }
    }()

    // 订阅者 2
    sub2 := ps.Subscribe("news")
    go func() {
        for msg := range sub2 {
            fmt.Println("Sub2 received:", msg)
        }
    }()

    time.Sleep(100 * time.Millisecond)

    // 发布消息
    ps.Publish("news", "Breaking: Go is awesome!")
    ps.Publish("news", "Update: Concurrency made easy")

    time.Sleep(100 * time.Millisecond)
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type PubSub struct {
    mu     sync.RWMutex
    topics map[string]map[chan string]struct{}
    closed bool
}

func NewPubSub() *PubSub {
    return &PubSub{
        topics: make(map[string]map[chan string]struct{}),
    }
}

func (ps *PubSub) Subscribe(topic string) <-chan string {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if ps.closed {
        return nil
    }

    ch := make(chan string, 10) // 缓冲防止阻塞

    if ps.topics[topic] == nil {
        ps.topics[topic] = make(map[chan string]struct{})
    }
    ps.topics[topic][ch] = struct{}{}

    return ch
}

func (ps *PubSub) Unsubscribe(topic string, ch <-chan string) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if subs, ok := ps.topics[topic]; ok {
        // 类型转换获取可写通道
        for sub := range subs {
            if sub == ch {
                delete(subs, sub)
                close(sub)
                break
            }
        }

        if len(subs) == 0 {
            delete(ps.topics, topic)
        }
    }
}

func (ps *PubSub) Publish(topic string, msg string) {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    if ps.closed {
        return
    }

    if subs, ok := ps.topics[topic]; ok {
        for ch := range subs {
            // 非阻塞发送
            select {
            case ch <- msg:
            default:
                // 通道满，跳过
            }
        }
    }
}

func (ps *PubSub) Close() {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if ps.closed {
        return
    }

    ps.closed = true

    for _, subs := range ps.topics {
        for ch := range subs {
            close(ch)
        }
    }
    ps.topics = nil
}

func main() {
    ps := NewPubSub()
    defer ps.Close()

    var wg sync.WaitGroup

    // 订阅者 1
    sub1 := ps.Subscribe("news")
    wg.Add(1)
    go func() {
        defer wg.Done()
        for msg := range sub1 {
            fmt.Println("Sub1 received:", msg)
        }
        fmt.Println("Sub1: channel closed")
    }()

    // 订阅者 2
    sub2 := ps.Subscribe("news")
    wg.Add(1)
    go func() {
        defer wg.Done()
        for msg := range sub2 {
            fmt.Println("Sub2 received:", msg)
        }
        fmt.Println("Sub2: channel closed")
    }()

    time.Sleep(100 * time.Millisecond)

    // 发布消息
    ps.Publish("news", "Breaking: Go is awesome!")
    ps.Publish("news", "Update: Concurrency made easy")

    time.Sleep(100 * time.Millisecond)

    ps.Close()
    wg.Wait()
}
```

</details>

---

## 练习 7: 优雅关闭 (Graceful Shutdown)

**难度**: 高级

实现一个支持优雅关闭的 HTTP 风格服务器模拟。

```go
type Server struct {
    // TODO: 定义字段
}

func NewServer() *Server {
    // TODO: 实现
}

// Start 启动服务器
func (s *Server) Start() error {
    // TODO: 实现
}

// HandleRequest 处理请求（模拟）
func (s *Server) HandleRequest(id int) {
    // TODO: 实现
}

// Shutdown 优雅关闭，等待进行中的请求完成
func (s *Server) Shutdown(ctx context.Context) error {
    // TODO: 实现
}

func main() {
    server := NewServer()

    // 启动服务器
    go server.Start()

    // 模拟请求
    for i := 0; i < 5; i++ {
        go server.HandleRequest(i)
        time.Sleep(50 * time.Millisecond)
    }

    time.Sleep(100 * time.Millisecond)

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        fmt.Println("Shutdown error:", err)
    } else {
        fmt.Println("Server shutdown gracefully")
    }
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

type Server struct {
    mu          sync.Mutex
    running     int32         // 原子操作
    activeReqs  sync.WaitGroup
    shutdownCh  chan struct{}
    doneCh      chan struct{}
}

func NewServer() *Server {
    return &Server{
        shutdownCh: make(chan struct{}),
        doneCh:     make(chan struct{}),
    }
}

func (s *Server) Start() error {
    atomic.StoreInt32(&s.running, 1)
    fmt.Println("Server started")

    <-s.shutdownCh
    fmt.Println("Server received shutdown signal")

    // 等待所有请求完成
    s.activeReqs.Wait()
    close(s.doneCh)

    return nil
}

func (s *Server) HandleRequest(id int) {
    // 检查是否正在关闭
    select {
    case <-s.shutdownCh:
        fmt.Printf("Request %d: rejected (server shutting down)\n", id)
        return
    default:
    }

    if atomic.LoadInt32(&s.running) != 1 {
        fmt.Printf("Request %d: rejected (server not running)\n", id)
        return
    }

    s.activeReqs.Add(1)
    defer s.activeReqs.Done()

    fmt.Printf("Request %d: started\n", id)

    // 模拟处理时间
    processingTime := time.Duration(100+id*50) * time.Millisecond
    time.Sleep(processingTime)

    fmt.Printf("Request %d: completed\n", id)
}

func (s *Server) Shutdown(ctx context.Context) error {
    atomic.StoreInt32(&s.running, 0)
    close(s.shutdownCh)

    select {
    case <-s.doneCh:
        fmt.Println("All requests completed")
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func main() {
    server := NewServer()

    // 启动服务器
    go server.Start()
    time.Sleep(50 * time.Millisecond) // 等待启动

    // 模拟请求
    for i := 0; i < 5; i++ {
        go server.HandleRequest(i)
        time.Sleep(50 * time.Millisecond)
    }

    time.Sleep(100 * time.Millisecond)

    // 优雅关闭
    fmt.Println("\nInitiating shutdown...")
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        fmt.Println("Shutdown error:", err)
    } else {
        fmt.Println("Server shutdown gracefully")
    }
}
```

</details>

---

## 练习 8: 并发爬虫 (Concurrent Crawler)

**难度**: 高级

实现一个简单的并发网页爬虫模拟器。

```go
type Crawler struct {
    // TODO: 定义字段
}

type CrawlResult struct {
    URL     string
    Links   []string
    Error   error
}

// NewCrawler 创建爬虫
// maxConcurrency: 最大并发数
// maxDepth: 最大爬取深度
func NewCrawler(maxConcurrency, maxDepth int) *Crawler {
    // TODO: 实现
}

// Crawl 从起始 URL 开始爬取
func (c *Crawler) Crawl(ctx context.Context, startURL string) []CrawlResult {
    // TODO: 实现
}

// 模拟获取页面链接
func mockFetch(url string) ([]string, error) {
    // 模拟数据
    links := map[string][]string{
        "http://example.com":       {"http://example.com/a", "http://example.com/b"},
        "http://example.com/a":     {"http://example.com/a/1", "http://example.com/a/2"},
        "http://example.com/b":     {"http://example.com/b/1"},
        "http://example.com/a/1":   {},
        "http://example.com/a/2":   {},
        "http://example.com/b/1":   {},
    }

    time.Sleep(50 * time.Millisecond) // 模拟网络延迟

    if l, ok := links[url]; ok {
        return l, nil
    }
    return nil, fmt.Errorf("not found: %s", url)
}
```

<details>
<summary>参考答案</summary>

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type CrawlResult struct {
    URL   string
    Links []string
    Depth int
    Error error
}

type Crawler struct {
    maxConcurrency int
    maxDepth       int
    visited        map[string]bool
    visitedMu      sync.Mutex
    sem            chan struct{}
}

func NewCrawler(maxConcurrency, maxDepth int) *Crawler {
    return &Crawler{
        maxConcurrency: maxConcurrency,
        maxDepth:       maxDepth,
        visited:        make(map[string]bool),
        sem:            make(chan struct{}, maxConcurrency),
    }
}

func (c *Crawler) markVisited(url string) bool {
    c.visitedMu.Lock()
    defer c.visitedMu.Unlock()

    if c.visited[url] {
        return false
    }
    c.visited[url] = true
    return true
}

func (c *Crawler) Crawl(ctx context.Context, startURL string) []CrawlResult {
    results := make(chan CrawlResult, 100)
    var wg sync.WaitGroup

    var crawl func(url string, depth int)
    crawl = func(url string, depth int) {
        defer wg.Done()

        // 检查深度
        if depth > c.maxDepth {
            return
        }

        // 检查是否已访问
        if !c.markVisited(url) {
            return
        }

        // 获取信号量
        select {
        case c.sem <- struct{}{}:
            defer func() { <-c.sem }()
        case <-ctx.Done():
            results <- CrawlResult{URL: url, Error: ctx.Err()}
            return
        }

        // 爬取
        links, err := mockFetch(url)

        results <- CrawlResult{
            URL:   url,
            Links: links,
            Depth: depth,
            Error: err,
        }

        if err != nil {
            return
        }

        // 递归爬取链接
        for _, link := range links {
            wg.Add(1)
            go crawl(link, depth+1)
        }
    }

    // 开始爬取
    wg.Add(1)
    go crawl(startURL, 0)

    // 等待完成并关闭结果通道
    go func() {
        wg.Wait()
        close(results)
    }()

    // 收集结果
    var allResults []CrawlResult
    for r := range results {
        allResults = append(allResults, r)
    }

    return allResults
}

func mockFetch(url string) ([]string, error) {
    links := map[string][]string{
        "http://example.com":     {"http://example.com/a", "http://example.com/b"},
        "http://example.com/a":   {"http://example.com/a/1", "http://example.com/a/2"},
        "http://example.com/b":   {"http://example.com/b/1"},
        "http://example.com/a/1": {},
        "http://example.com/a/2": {},
        "http://example.com/b/1": {},
    }

    time.Sleep(50 * time.Millisecond)

    if l, ok := links[url]; ok {
        return l, nil
    }
    return nil, fmt.Errorf("not found: %s", url)
}

func main() {
    crawler := NewCrawler(3, 2)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    results := crawler.Crawl(ctx, "http://example.com")

    fmt.Println("Crawl Results:")
    for _, r := range results {
        if r.Error != nil {
            fmt.Printf("  [ERROR] %s: %v\n", r.URL, r.Error)
        } else {
            fmt.Printf("  [Depth %d] %s -> %v\n", r.Depth, r.URL, r.Links)
        }
    }
}
```

</details>

---

## 挑战练习: 分布式任务调度器

**难度**: 专家级

实现一个简单的分布式任务调度器，支持：
- 任务优先级
- 任务超时
- 失败重试
- 优雅关闭

```go
type Priority int

const (
    Low Priority = iota
    Normal
    High
    Critical
)

type Task struct {
    ID       string
    Priority Priority
    Payload  interface{}
    Timeout  time.Duration
    MaxRetry int
}

type TaskResult struct {
    TaskID  string
    Success bool
    Result  interface{}
    Error   error
    Retries int
}

type Scheduler struct {
    // 自行设计
}

// 自行实现接口
```

提示：
- 使用优先级队列管理任务
- 使用 context 控制超时
- 使用 WaitGroup 和 channel 协调 workers
- 实现优雅关闭

---

## 完成标准

- [ ] 完成练习 1-4（基础和中级）
- [ ] 完成练习 5-6（中级）
- [ ] 完成练习 7-8（高级）
- [ ] 代码通过 `go vet` 检查
- [ ] 代码通过 `go test -race` 测试
- [ ] 理解每个并发模式的使用场景

## 下一步

完成这些练习后，你可以：
1. 阅读 [04-error-handling](../04-error-handling/) 学习错误处理
2. 继续实战项目 [Todo API](../../02-practice/todo-api/)
