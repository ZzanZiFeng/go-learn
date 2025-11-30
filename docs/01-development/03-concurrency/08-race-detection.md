# 竞态检测

本文档介绍如何检测和避免 Go 程序中的竞态条件。

## 目录

- [什么是竞态条件](#什么是竞态条件)
- [Go 竞态检测器](#go-竞态检测器)
- [常见竞态问题](#常见竞态问题)
- [避免竞态的方法](#避免竞态的方法)
- [最佳实践](#最佳实践)

---

## 什么是竞态条件

### 定义

竞态条件（Race Condition）发生在两个或多个 goroutine 同时访问共享数据，且至少一个是写操作时。

```go
package main

import (
    "fmt"
    "sync"
)

// 竞态条件示例
func main() {
    var counter int
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // 竞态！多个 goroutine 同时读写
        }()
    }

    wg.Wait()
    fmt.Println("Counter:", counter) // 结果不确定
}
```

### 为什么危险

- **不确定性**：每次运行结果可能不同
- **难以调试**：问题难以重现
- **数据损坏**：可能导致数据不一致
- **安全风险**：可能被利用进行攻击

---

## Go 竞态检测器

### 使用方法

```bash
# 运行时检测
go run -race main.go

# 测试时检测
go test -race ./...

# 编译时检测
go build -race -o myapp main.go
```

### 检测输出示例

```go
// race.go
package main

import "time"

func main() {
    var data int

    go func() {
        data = 42 // 写
    }()

    if data == 0 { // 读
        time.Sleep(time.Millisecond)
    }
}
```

```bash
$ go run -race race.go
==================
WARNING: DATA RACE
Write at 0x00c0000140a8 by goroutine 7:
  main.main.func1()
      /path/race.go:9 +0x38

Previous read at 0x00c0000140a8 by main goroutine:
  main.main()
      /path/race.go:12 +0x88

Goroutine 7 (running) created at:
  main.main()
      /path/race.go:8 +0x7a
==================
Found 1 data race(s)
```

### 注意事项

| 特性 | 说明 |
|-----|------|
| 性能开销 | 10x 慢，2-10x 内存 |
| 用途 | 开发和测试，不用于生产 |
| 检测范围 | 只检测运行时实际发生的竞态 |

---

## 常见竞态问题

### 1. 并发读写 Map

```go
// 错误：并发读写内置 map
var m = make(map[string]int)

go func() { m["key"] = 1 }()  // 写
go func() { _ = m["key"] }()  // 读
// panic: concurrent map read and map write

// 正确：使用 sync.Map
var m sync.Map

go func() { m.Store("key", 1) }()
go func() { m.Load("key") }()
```

### 2. 循环变量捕获

```go
// 错误：goroutine 捕获循环变量
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i) // 竞态：可能打印 3, 3, 3
    }()
}

// 正确：传递参数
for i := 0; i < 3; i++ {
    go func(n int) {
        fmt.Println(n) // 0, 1, 2（顺序不确定）
    }(i)
}
```

### 3. 检查后执行（Check-then-act）

```go
// 错误：检查和操作不是原子的
if counter == 0 {
    counter = 1 // 另一个 goroutine 可能已经修改了
}

// 正确：使用锁或原子操作
mu.Lock()
if counter == 0 {
    counter = 1
}
mu.Unlock()

// 或使用 sync/atomic
atomic.CompareAndSwapInt64(&counter, 0, 1)
```

### 4. 延迟初始化

```go
// 错误：不安全的单例
var instance *Database

func GetDB() *Database {
    if instance == nil {
        instance = &Database{} // 竞态
    }
    return instance
}

// 正确：使用 sync.Once
var (
    instance *Database
    once     sync.Once
)

func GetDB() *Database {
    once.Do(func() {
        instance = &Database{}
    })
    return instance
}
```

### 5. 切片追加

```go
// 错误：并发追加切片
var slice []int

go func() { slice = append(slice, 1) }()
go func() { slice = append(slice, 2) }()
// 竞态：可能丢失数据或 panic

// 正确：使用锁保护
var (
    slice []int
    mu    sync.Mutex
)

go func() {
    mu.Lock()
    slice = append(slice, 1)
    mu.Unlock()
}()
```

---

## 避免竞态的方法

### 1. 使用 Mutex

```go
type SafeCounter struct {
    mu    sync.Mutex
    value int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}
```

### 2. 使用 atomic 包

```go
import "sync/atomic"

var counter int64

// 原子加
atomic.AddInt64(&counter, 1)

// 原子读
val := atomic.LoadInt64(&counter)

// 原子写
atomic.StoreInt64(&counter, 100)

// 比较并交换
atomic.CompareAndSwapInt64(&counter, 100, 200)
```

### 3. 使用 Channel

```go
// 通过通信共享内存，而不是共享内存来通信
type Counter struct {
    ch chan func()
}

func NewCounter() *Counter {
    c := &Counter{ch: make(chan func())}
    go c.run()
    return c
}

func (c *Counter) run() {
    var value int
    for f := range c.ch {
        f()
    }
}

func (c *Counter) Increment() {
    done := make(chan struct{})
    c.ch <- func() {
        value++
        close(done)
    }
    <-done
}
```

### 4. 使用 sync.Map

```go
var m sync.Map

// 存储
m.Store("key", "value")

// 读取
if val, ok := m.Load("key"); ok {
    fmt.Println(val)
}

// 删除
m.Delete("key")

// 遍历
m.Range(func(key, value interface{}) bool {
    fmt.Println(key, value)
    return true
})
```

### 5. 不可变数据

```go
// 使用不可变数据结构
type Config struct {
    Host string
    Port int
}

var config atomic.Value // 存储 *Config

func UpdateConfig(c *Config) {
    config.Store(c) // 原子替换整个配置
}

func GetConfig() *Config {
    return config.Load().(*Config)
}
```

---

## 最佳实践

### 1. 限制共享状态

```go
// 不好：全局共享状态
var globalCounter int

func Increment() {
    globalCounter++ // 需要同步
}

// 好：将状态封装在 goroutine 中
func CounterService() (increment func(), value func() int) {
    var counter int
    ch := make(chan func())

    go func() {
        for f := range ch {
            f()
        }
    }()

    increment = func() {
        done := make(chan struct{})
        ch <- func() {
            counter++
            close(done)
        }
        <-done
    }

    value = func() int {
        result := make(chan int)
        ch <- func() {
            result <- counter
        }
        return <-result
    }

    return
}
```

### 2. 明确所有权

```go
// 好：明确谁拥有数据
func process(data []int) []int {
    result := make([]int, len(data))
    // 处理 data，返回新切片
    return result
}

// 不好：不清楚谁负责修改
func process(data []int) {
    // 直接修改 data？
}
```

### 3. 使用 defer 释放锁

```go
func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock() // 确保释放
    c.value++
}
```

### 4. 减小锁的粒度

```go
// 不好：大锁
func (s *Service) Process(item Item) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.validate(item)
    s.transform(item)
    s.save(item)
}

// 好：只锁必要的部分
func (s *Service) Process(item Item) {
    s.validate(item) // 无需锁

    s.mu.Lock()
    s.save(item) // 只锁写操作
    s.mu.Unlock()
}
```

### 5. 在 CI 中运行竞态检测

```yaml
# .github/workflows/test.yml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Test with race detector
        run: go test -race ./...
```

---

## 竞态检测工具比较

| 工具 | 类型 | 优点 | 缺点 |
|-----|------|------|------|
| `-race` flag | 动态分析 | 零误报 | 只检测执行路径 |
| go vet | 静态分析 | 快速 | 误报可能 |
| staticcheck | 静态分析 | 更多检查 | 需要安装 |

---

## 示例：修复竞态

### 修复前

```go
type Cache struct {
    data map[string]string
}

func (c *Cache) Get(key string) string {
    return c.data[key]
}

func (c *Cache) Set(key, value string) {
    c.data[key] = value
}
```

### 修复后

```go
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func NewCache() *Cache {
    return &Cache{
        data: make(map[string]string),
    }
}

func (c *Cache) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}
```

---

## 下一步

恭喜你完成了并发编程章节！接下来：

- [后端架构](../04-architecture/) - 学习 Go 项目架构设计
