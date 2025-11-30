# Context

本文档介绍 Go 语言的 context 包，用于取消、超时和值传递。

## 目录

- [Context 基础](#context-基础)
- [取消 Context](#取消-context)
- [超时 Context](#超时-context)
- [Context 传值](#context-传值)
- [最佳实践](#最佳实践)
- [与 JavaScript 对比](#与-javascript-对比)

---

## Context 基础

### 什么是 Context

Context 是 Go 中用于在 goroutine 之间传递截止时间、取消信号和请求范围值的标准方式。

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```

### 为什么需要 Context

```go
// 问题：如何停止已启动的 goroutine？
func worker() {
    for {
        // 做一些工作
        time.Sleep(time.Second)
        fmt.Println("Working...")
    }
}

// 使用 context 的解决方案
func workerWithContext(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("Worker stopped:", ctx.Err())
            return
        default:
            time.Sleep(time.Second)
            fmt.Println("Working...")
        }
    }
}
```

### Context 类型

| 函数 | 说明 |
|-----|------|
| `context.Background()` | 根 context，永不取消 |
| `context.TODO()` | 占位符，不确定用什么时使用 |
| `context.WithCancel(parent)` | 可取消的 context |
| `context.WithTimeout(parent, timeout)` | 超时自动取消 |
| `context.WithDeadline(parent, deadline)` | 到期自动取消 |
| `context.WithValue(parent, key, val)` | 携带值的 context |

---

## 取消 Context

### WithCancel

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Worker %d stopped: %v\n", id, ctx.Err())
            return
        default:
            fmt.Printf("Worker %d working...\n", id)
            time.Sleep(500 * time.Millisecond)
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())

    // 启动多个 worker
    for i := 1; i <= 3; i++ {
        go worker(ctx, i)
    }

    // 运行 2 秒后取消
    time.Sleep(2 * time.Second)
    fmt.Println("Cancelling...")
    cancel() // 取消所有 worker

    time.Sleep(time.Second) // 等待 worker 退出
    fmt.Println("Done")
}
```

### 级联取消

取消父 context 会自动取消所有子 context：

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // 父 context
    parentCtx, parentCancel := context.WithCancel(context.Background())

    // 子 context
    childCtx, _ := context.WithCancel(parentCtx)

    go func() {
        <-childCtx.Done()
        fmt.Println("Child cancelled:", childCtx.Err())
    }()

    // 取消父 context
    time.Sleep(time.Second)
    parentCancel()

    time.Sleep(100 * time.Millisecond)
}
```

---

## 超时 Context

### WithTimeout

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func slowOperation(ctx context.Context) error {
    select {
    case <-time.After(5 * time.Second):
        fmt.Println("Operation completed")
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func main() {
    // 2 秒超时
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel() // 始终调用 cancel 释放资源

    err := slowOperation(ctx)
    if err != nil {
        fmt.Println("Error:", err) // context deadline exceeded
    }
}
```

### WithDeadline

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func main() {
    // 设置截止时间
    deadline := time.Now().Add(2 * time.Second)
    ctx, cancel := context.WithDeadline(context.Background(), deadline)
    defer cancel()

    // 检查截止时间
    if d, ok := ctx.Deadline(); ok {
        fmt.Println("Deadline:", d.Format("15:04:05"))
    }

    select {
    case <-time.After(5 * time.Second):
        fmt.Println("Completed")
    case <-ctx.Done():
        fmt.Println("Cancelled:", ctx.Err())
    }
}
```

### HTTP 请求超时

```go
package main

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"
)

func fetchURL(ctx context.Context, url string) (string, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return "", err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    return string(body), err
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    body, err := fetchURL(ctx, "https://example.com")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Length:", len(body))
}
```

---

## Context 传值

### WithValue

```go
package main

import (
    "context"
    "fmt"
)

// 定义 key 类型（避免冲突）
type contextKey string

const (
    userIDKey   contextKey = "userID"
    requestIDKey contextKey = "requestID"
)

func processRequest(ctx context.Context) {
    // 获取值
    if userID, ok := ctx.Value(userIDKey).(string); ok {
        fmt.Println("User ID:", userID)
    }
    if requestID, ok := ctx.Value(requestIDKey).(string); ok {
        fmt.Println("Request ID:", requestID)
    }
}

func main() {
    ctx := context.Background()

    // 添加值
    ctx = context.WithValue(ctx, userIDKey, "user-123")
    ctx = context.WithValue(ctx, requestIDKey, "req-456")

    processRequest(ctx)
}
```

### 注意事项

```go
// 不好：使用内置类型作为 key
ctx = context.WithValue(ctx, "userID", "123") // 可能冲突

// 好：使用自定义类型
type contextKey string
const userIDKey contextKey = "userID"
ctx = context.WithValue(ctx, userIDKey, "123")
```

---

## 最佳实践

### 1. 始终传递 Context

```go
// 好：context 作为第一个参数
func ProcessOrder(ctx context.Context, orderID string) error {
    // ...
}

// 不好：不接受 context
func ProcessOrder(orderID string) error {
    // 无法取消或设置超时
}
```

### 2. 始终调用 cancel

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel() // 即使操作成功完成也要调用
```

### 3. 检查 Context 错误

```go
func doWork(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // 返回取消原因
        default:
            // 做工作
        }
    }
}
```

### 4. 不要存储 Context

```go
// 不好：将 context 存储在结构体中
type Service struct {
    ctx context.Context // 不要这样做
}

// 好：每个方法接收 context
type Service struct{}

func (s *Service) Process(ctx context.Context) error {
    // ...
}
```

### 5. Context 值只用于请求范围数据

```go
// 好：请求 ID、用户 ID 等请求范围的数据
ctx = context.WithValue(ctx, requestIDKey, "req-123")

// 不好：传递可选参数
ctx = context.WithValue(ctx, "timeout", 30) // 应该用 WithTimeout
ctx = context.WithValue(ctx, "retries", 3)  // 应该是函数参数
```

---

## 与 JavaScript 对比

### AbortController

```javascript
// JavaScript - AbortController
const controller = new AbortController();

fetch(url, { signal: controller.signal })
    .then(response => response.json())
    .catch(err => {
        if (err.name === 'AbortError') {
            console.log('Request cancelled');
        }
    });

// 取消请求
controller.abort();
```

```go
// Go - context.WithCancel
ctx, cancel := context.WithCancel(context.Background())

go func() {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            fmt.Println("Request cancelled")
        }
        return
    }
    // 处理响应
}()

// 取消请求
cancel()
```

### Promise 超时

```javascript
// JavaScript - Promise.race 超时
function withTimeout(promise, ms) {
    const timeout = new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Timeout')), ms)
    );
    return Promise.race([promise, timeout]);
}

const result = await withTimeout(fetch(url), 5000);
```

```go
// Go - context.WithTimeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := http.DefaultClient.Do(req)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Timeout")
    }
}
```

### 请求范围数据

```javascript
// JavaScript - AsyncLocalStorage (Node.js)
const { AsyncLocalStorage } = require('async_hooks');
const storage = new AsyncLocalStorage();

storage.run({ requestId: 'req-123' }, () => {
    handleRequest();
});

function handleRequest() {
    const { requestId } = storage.getStore();
    console.log('Request ID:', requestId);
}
```

```go
// Go - context.WithValue
ctx := context.WithValue(context.Background(), requestIDKey, "req-123")
handleRequest(ctx)

func handleRequest(ctx context.Context) {
    requestID := ctx.Value(requestIDKey).(string)
    fmt.Println("Request ID:", requestID)
}
```

---

## 常见模式

### 1. 优雅关闭

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())

    // 启动服务
    go runServer(ctx)

    // 监听信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    <-sigCh
    fmt.Println("Shutting down...")
    cancel()

    time.Sleep(time.Second) // 等待清理
}
```

### 2. 超时重试

```go
func fetchWithRetry(ctx context.Context, url string, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        // 检查是否已取消
        if ctx.Err() != nil {
            return ctx.Err()
        }

        // 每次请求单独超时
        reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
        err := fetch(reqCtx, url)
        cancel()

        if err == nil {
            return nil
        }

        fmt.Printf("Attempt %d failed: %v\n", i+1, err)
    }
    return errors.New("max retries exceeded")
}
```

### 3. 并行任务取消

```go
func parallelFetch(ctx context.Context, urls []string) ([]string, error) {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel() // 任何一个失败就取消其他

    results := make([]string, len(urls))
    errs := make(chan error, len(urls))

    for i, url := range urls {
        go func(i int, url string) {
            body, err := fetchURL(ctx, url)
            if err != nil {
                errs <- err
                cancel() // 取消其他请求
                return
            }
            results[i] = body
            errs <- nil
        }(i, url)
    }

    for range urls {
        if err := <-errs; err != nil {
            return nil, err
        }
    }
    return results, nil
}
```

---

## 下一步

- [并发模式](./07-patterns.md) - 学习常见的并发模式
