# Sync 包

本文档介绍 Go 语言 sync 包中的同步原语。

## 目录

- [WaitGroup](#waitgroup)
- [Mutex](#mutex)
- [RWMutex](#rwmutex)
- [Once](#once)
- [Cond](#cond)
- [Map](#map)
- [Pool](#pool)
- [与 JavaScript 对比](#与-javascript-对比)

---

## WaitGroup

### 基本用法

WaitGroup 用于等待一组 goroutine 完成。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done() // 完成时减少计数

    fmt.Printf("Worker %d starting\n", id)
    time.Sleep(time.Second)
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 3; i++ {
        wg.Add(1) // 增加计数
        go worker(i, &wg)
    }

    wg.Wait() // 等待所有 worker 完成
    fmt.Println("All workers done")
}
```

### WaitGroup 方法

| 方法 | 说明 |
|-----|------|
| `Add(delta int)` | 增加计数（可以为负数） |
| `Done()` | 减少计数（等价于 `Add(-1)`） |
| `Wait()` | 阻塞直到计数为 0 |

### 常见错误

```go
// 错误：在 goroutine 内 Add
func main() {
    var wg sync.WaitGroup

    for i := 0; i < 3; i++ {
        go func() {
            wg.Add(1) // 错误位置！可能在 Wait 之后才执行
            defer wg.Done()
            // 工作
        }()
    }
    wg.Wait()
}

// 正确：在启动 goroutine 前 Add
func main() {
    var wg sync.WaitGroup

    for i := 0; i < 3; i++ {
        wg.Add(1) // 正确位置
        go func() {
            defer wg.Done()
            // 工作
        }()
    }
    wg.Wait()
}
```

---

## Mutex

### 基本用法

Mutex 提供互斥锁，保护共享数据。

```go
package main

import (
    "fmt"
    "sync"
)

type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

func main() {
    counter := &SafeCounter{}
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Increment()
        }()
    }

    wg.Wait()
    fmt.Println("Count:", counter.Value()) // 1000
}
```

### Mutex 方法

| 方法 | 说明 |
|-----|------|
| `Lock()` | 获取锁（阻塞直到获取） |
| `Unlock()` | 释放锁 |
| `TryLock()` | 尝试获取锁（Go 1.18+，非阻塞） |

### 死锁示例

```go
// 死锁：重复加锁
func (c *SafeCounter) Bad() {
    c.mu.Lock()
    c.mu.Lock() // 死锁！同一个 goroutine 不能重复加锁
    c.mu.Unlock()
    c.mu.Unlock()
}

// 死锁：忘记解锁
func (c *SafeCounter) AlsoBad() {
    c.mu.Lock()
    if someCondition {
        return // 忘记 Unlock！
    }
    c.mu.Unlock()
}

// 正确：使用 defer
func (c *SafeCounter) Good() {
    c.mu.Lock()
    defer c.mu.Unlock()

    if someCondition {
        return // defer 会确保 Unlock
    }
}
```

---

## RWMutex

### 读写锁

RWMutex 允许多个读者或单个写者。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type SafeMap struct {
    mu   sync.RWMutex
    data map[string]int
}

func NewSafeMap() *SafeMap {
    return &SafeMap{
        data: make(map[string]int),
    }
}

// 写操作使用写锁
func (m *SafeMap) Set(key string, value int) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.data[key] = value
}

// 读操作使用读锁
func (m *SafeMap) Get(key string) (int, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    val, ok := m.data[key]
    return val, ok
}

func main() {
    m := NewSafeMap()
    var wg sync.WaitGroup

    // 写者
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            m.Set(fmt.Sprintf("key%d", i), i)
        }(i)
    }

    // 读者
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            m.Get(fmt.Sprintf("key%d", i%10))
        }(i)
    }

    wg.Wait()
    fmt.Println("Done")
}
```

### RWMutex 方法

| 方法 | 说明 |
|-----|------|
| `Lock()` | 获取写锁 |
| `Unlock()` | 释放写锁 |
| `RLock()` | 获取读锁 |
| `RUnlock()` | 释放读锁 |

### 使用场景

| 场景 | 推荐 |
|-----|------|
| 写多读少 | Mutex |
| 读多写少 | RWMutex |
| 高频读写 | sync.Map 或分片锁 |

---

## Once

### 只执行一次

sync.Once 确保函数只执行一次，常用于初始化。

```go
package main

import (
    "fmt"
    "sync"
)

var (
    instance *Database
    once     sync.Once
)

type Database struct {
    connection string
}

func GetDatabase() *Database {
    once.Do(func() {
        fmt.Println("Initializing database...")
        instance = &Database{connection: "localhost:5432"}
    })
    return instance
}

func main() {
    var wg sync.WaitGroup

    // 多个 goroutine 同时调用
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            db := GetDatabase()
            fmt.Printf("Goroutine %d got db: %p\n", i, db)
        }(i)
    }

    wg.Wait()
}
```

**输出**:
```
Initializing database...
Goroutine 0 got db: 0xc0000101e0
Goroutine 1 got db: 0xc0000101e0
Goroutine 2 got db: 0xc0000101e0
Goroutine 3 got db: 0xc0000101e0
Goroutine 4 got db: 0xc0000101e0
```

---

## Cond

### 条件变量

Cond 用于等待或通知条件变化。

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type Queue struct {
    items []int
    cond  *sync.Cond
}

func NewQueue() *Queue {
    return &Queue{
        cond: sync.NewCond(&sync.Mutex{}),
    }
}

func (q *Queue) Push(item int) {
    q.cond.L.Lock()
    defer q.cond.L.Unlock()

    q.items = append(q.items, item)
    q.cond.Signal() // 通知一个等待者
}

func (q *Queue) Pop() int {
    q.cond.L.Lock()
    defer q.cond.L.Unlock()

    for len(q.items) == 0 {
        q.cond.Wait() // 等待通知
    }

    item := q.items[0]
    q.items = q.items[1:]
    return item
}

func main() {
    q := NewQueue()

    // 消费者
    go func() {
        for i := 0; i < 5; i++ {
            item := q.Pop()
            fmt.Println("Consumed:", item)
        }
    }()

    // 生产者
    for i := 0; i < 5; i++ {
        time.Sleep(500 * time.Millisecond)
        q.Push(i)
        fmt.Println("Produced:", i)
    }

    time.Sleep(time.Second)
}
```

### Cond 方法

| 方法 | 说明 |
|-----|------|
| `Wait()` | 等待通知（自动释放/获取锁） |
| `Signal()` | 唤醒一个等待者 |
| `Broadcast()` | 唤醒所有等待者 |

---

## Map

### sync.Map

sync.Map 是并发安全的 map，适合读多写少的场景。

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var m sync.Map

    // 存储
    m.Store("key1", "value1")
    m.Store("key2", 42)

    // 读取
    if val, ok := m.Load("key1"); ok {
        fmt.Println("key1:", val)
    }

    // 读取或存储
    actual, loaded := m.LoadOrStore("key3", "new value")
    fmt.Printf("key3: %v, existed: %v\n", actual, loaded)

    // 删除
    m.Delete("key1")

    // 遍历
    m.Range(func(key, value interface{}) bool {
        fmt.Printf("%v: %v\n", key, value)
        return true // 返回 false 停止遍历
    })
}
```

### sync.Map vs map + Mutex

| 特性 | sync.Map | map + Mutex |
|-----|----------|-------------|
| 读性能 | 高 | 中等 |
| 写性能 | 中等 | 高 |
| 内存 | 较高 | 较低 |
| 类型安全 | 无（interface{}） | 有 |

---

## Pool

### 对象池

sync.Pool 用于复用临时对象，减少 GC 压力。

```go
package main

import (
    "bytes"
    "fmt"
    "sync"
)

var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processData(data string) string {
    // 从池中获取
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()           // 重置
        bufferPool.Put(buf)   // 放回池中
    }()

    buf.WriteString("Processed: ")
    buf.WriteString(data)
    return buf.String()
}

func main() {
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            result := processData(fmt.Sprintf("data%d", i))
            fmt.Println(result)
        }(i)
    }

    wg.Wait()
}
```

---

## 与 JavaScript 对比

### 锁机制

```javascript
// JavaScript - 没有真正的锁（单线程）
// 但可以用 Promise 模拟
class AsyncLock {
    constructor() {
        this._locked = false;
        this._waiting = [];
    }

    async acquire() {
        while (this._locked) {
            await new Promise(resolve => this._waiting.push(resolve));
        }
        this._locked = true;
    }

    release() {
        this._locked = false;
        const next = this._waiting.shift();
        if (next) next();
    }
}
```

```go
// Go - sync.Mutex
var mu sync.Mutex

mu.Lock()
// 临界区
mu.Unlock()
```

### 等待多个操作

```javascript
// JavaScript - Promise.all
await Promise.all([
    doTask1(),
    doTask2(),
    doTask3()
]);
```

```go
// Go - sync.WaitGroup
var wg sync.WaitGroup

wg.Add(3)
go func() { defer wg.Done(); doTask1() }()
go func() { defer wg.Done(); doTask2() }()
go func() { defer wg.Done(); doTask3() }()

wg.Wait()
```

### 单例模式

```javascript
// JavaScript - 模块单例
let instance;

function getInstance() {
    if (!instance) {
        instance = createInstance();
    }
    return instance;
}
```

```go
// Go - sync.Once
var (
    instance *Instance
    once     sync.Once
)

func GetInstance() *Instance {
    once.Do(func() {
        instance = createInstance()
    })
    return instance
}
```

---

## 最佳实践

### 1. 始终使用 defer 释放锁

```go
mu.Lock()
defer mu.Unlock()
// 即使 panic 也会释放锁
```

### 2. 缩小锁的范围

```go
// 不好：整个函数加锁
func process() {
    mu.Lock()
    defer mu.Unlock()

    data := prepare()      // 不需要锁
    result := compute(data) // 可能需要锁
    log(result)            // 不需要锁
}

// 好：只锁需要的部分
func process() {
    data := prepare()

    mu.Lock()
    result := compute(data)
    mu.Unlock()

    log(result)
}
```

### 3. 避免锁嵌套

```go
// 危险：可能死锁
func (a *A) Method1() {
    a.mu.Lock()
    defer a.mu.Unlock()
    b.Method2() // 如果 Method2 也锁 a.mu，死锁
}

// 安全：分离锁
func (a *A) Method1() {
    a.mu.Lock()
    data := a.data
    a.mu.Unlock()

    b.Method2(data) // 不持有锁
}
```

---

## 下一步

- [Context](./06-context.md) - 学习使用 context 进行取消和超时控制
