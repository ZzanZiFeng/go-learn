# 性能测试

## 学习目标

掌握 Go 基准测试和性能分析工具。

## 1. 基准测试基础

### 1.1 编写基准测试

```go
// benchmark_test.go
package mypackage

import "testing"

func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(1, 2)
    }
}

func BenchmarkFibonacci(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Fibonacci(20)
    }
}
```

### 1.2 运行基准测试

```bash
# 运行所有基准测试
go test -bench=.

# 运行特定基准测试
go test -bench=BenchmarkAdd

# 指定运行时间
go test -bench=. -benchtime=5s

# 运行多次取平均
go test -bench=. -count=5

# 显示内存分配
go test -bench=. -benchmem

# 禁用 CPU 优化
go test -bench=. -gcflags=-N
```

### 1.3 基准测试输出解读

```
BenchmarkAdd-8          1000000000          0.256 ns/op
BenchmarkFibonacci-8       30000             45123 ns/op           0 B/op          0 allocs/op
```

| 字段 | 说明 |
|------|------|
| -8 | GOMAXPROCS 数量 |
| 1000000000 | 运行次数 |
| 0.256 ns/op | 每次操作耗时 |
| 0 B/op | 每次操作分配内存 |
| 0 allocs/op | 每次操作分配次数 |

## 2. 高级基准测试

### 2.1 子基准测试

```go
func BenchmarkJSON(b *testing.B) {
    data := []byte(`{"name":"test","value":123}`)

    b.Run("Unmarshal", func(b *testing.B) {
        var result map[string]interface{}
        for i := 0; i < b.N; i++ {
            json.Unmarshal(data, &result)
        }
    })

    b.Run("Marshal", func(b *testing.B) {
        obj := map[string]interface{}{"name": "test", "value": 123}
        for i := 0; i < b.N; i++ {
            json.Marshal(obj)
        }
    })
}
```

### 2.2 并行基准测试

```go
func BenchmarkParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            // 并行执行的代码
            processRequest()
        }
    })
}
```

### 2.3 重置计时器

```go
func BenchmarkWithSetup(b *testing.B) {
    // 准备工作（不计入基准时间）
    data := generateLargeData()

    b.ResetTimer() // 重置计时器

    for i := 0; i < b.N; i++ {
        processData(data)
    }
}

func BenchmarkWithPause(b *testing.B) {
    for i := 0; i < b.N; i++ {
        b.StopTimer()
        data := generateTestData() // 不计入时间
        b.StartTimer()

        processData(data)
    }
}
```

### 2.4 报告内存分配

```go
func BenchmarkMemory(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        _ = make([]byte, 1024)
    }
}
```

## 3. 性能比较

### 3.1 比较不同实现

```go
// 字符串拼接比较
func BenchmarkStringConcat(b *testing.B) {
    b.Run("plus", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            s := ""
            for j := 0; j < 100; j++ {
                s += "x"
            }
        }
    })

    b.Run("builder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            for j := 0; j < 100; j++ {
                builder.WriteString("x")
            }
            _ = builder.String()
        }
    })

    b.Run("buffer", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var buffer bytes.Buffer
            for j := 0; j < 100; j++ {
                buffer.WriteString("x")
            }
            _ = buffer.String()
        }
    })
}
```

### 3.2 使用 benchstat

```bash
# 安装 benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# 保存基准测试结果
go test -bench=. -count=10 > old.txt

# 修改代码后重新测试
go test -bench=. -count=10 > new.txt

# 比较结果
benchstat old.txt new.txt
```

输出示例：
```
name          old time/op  new time/op  delta
StringConcat  12.3µs ± 2%  10.1µs ± 1%  -17.89%  (p=0.000 n=10+10)
```

## 4. pprof 性能分析

### 4.1 CPU 分析

```go
import (
    "os"
    "runtime/pprof"
    "testing"
)

func BenchmarkWithCPUProfile(b *testing.B) {
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    for i := 0; i < b.N; i++ {
        doWork()
    }
}
```

```bash
# 生成 CPU profile
go test -bench=. -cpuprofile=cpu.prof

# 分析 profile
go tool pprof cpu.prof

# 常用命令
(pprof) top10        # 查看 top 10 消耗
(pprof) list doWork  # 查看函数详情
(pprof) web          # 生成火焰图（需要 graphviz）
```

### 4.2 内存分析

```bash
# 生成内存 profile
go test -bench=. -memprofile=mem.prof

# 分析
go tool pprof mem.prof

(pprof) top
(pprof) list functionName
```

### 4.3 HTTP pprof

```go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // 应用代码...
}
```

```bash
# 访问 pprof
# CPU profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# 内存 profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# goroutine
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# 使用 pprof 工具
go tool pprof -http=:8080 cpu.prof
```

## 5. 常见性能优化

### 5.1 减少内存分配

```go
// 不好：每次循环都分配
func processItems(items []string) []string {
    var result []string
    for _, item := range items {
        result = append(result, strings.ToUpper(item))
    }
    return result
}

// 好：预分配容量
func processItemsOptimized(items []string) []string {
    result := make([]string, 0, len(items))
    for _, item := range items {
        result = append(result, strings.ToUpper(item))
    }
    return result
}
```

### 5.2 使用 sync.Pool

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processWithPool() string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.WriteString("processed data")
    return buf.String()
}
```

### 5.3 避免不必要的转换

```go
// 不好
func containsSubstring(data []byte, substr string) bool {
    return strings.Contains(string(data), substr) // 分配新字符串
}

// 好
func containsSubstringOptimized(data []byte, substr string) bool {
    return bytes.Contains(data, []byte(substr))
}
```

## 6. 性能测试示例

### 6.1 JSON 处理对比

```go
// tests/benchmark/json_bench_test.go
package benchmark

import (
    "encoding/json"
    "testing"

    jsoniter "github.com/json-iterator/go"
)

type User struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

var testUser = User{
    ID:       1,
    Username: "testuser",
    Email:    "test@example.com",
}

func BenchmarkJSON_StdLib_Marshal(b *testing.B) {
    for i := 0; i < b.N; i++ {
        json.Marshal(testUser)
    }
}

func BenchmarkJSON_JsonIter_Marshal(b *testing.B) {
    var jsonFast = jsoniter.ConfigFastest
    for i := 0; i < b.N; i++ {
        jsonFast.Marshal(testUser)
    }
}

func BenchmarkJSON_StdLib_Unmarshal(b *testing.B) {
    data, _ := json.Marshal(testUser)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        var u User
        json.Unmarshal(data, &u)
    }
}

func BenchmarkJSON_JsonIter_Unmarshal(b *testing.B) {
    var jsonFast = jsoniter.ConfigFastest
    data, _ := json.Marshal(testUser)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        var u User
        jsonFast.Unmarshal(data, &u)
    }
}
```

### 6.2 并发处理对比

```go
// tests/benchmark/concurrent_bench_test.go
package benchmark

import (
    "sync"
    "testing"
)

func BenchmarkMutex(b *testing.B) {
    var mu sync.Mutex
    var count int

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            mu.Lock()
            count++
            mu.Unlock()
        }
    })
}

func BenchmarkRWMutex_Read(b *testing.B) {
    var mu sync.RWMutex
    var count int

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            mu.RLock()
            _ = count
            mu.RUnlock()
        }
    })
}

func BenchmarkAtomic(b *testing.B) {
    var count int64

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            atomic.AddInt64(&count, 1)
        }
    })
}

func BenchmarkChannel(b *testing.B) {
    ch := make(chan struct{}, 1)
    var count int

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            ch <- struct{}{}
            count++
            <-ch
        }
    })
}
```

## 练习

1. 为字符串拼接编写基准测试，比较不同方法
2. 使用 pprof 分析程序瓶颈
3. 使用 sync.Pool 优化内存分配
4. 使用 benchstat 比较优化前后的性能

## 下一步

[调试技巧](../05-debugging/) - 学习 Go 程序调试技术。
