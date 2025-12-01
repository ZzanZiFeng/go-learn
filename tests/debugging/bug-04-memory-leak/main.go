// tests/debugging/bug-04-memory-leak/main.go
// Bug 场景 4: 内存泄漏
// 使用 pprof 分析内存使用

package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

// ========================================
// 问题 1: Goroutine 泄漏
// ========================================
func goroutineLeakBug() {
	for i := 0; i < 100; i++ {
		ch := make(chan int)
		go func() {
			// 这个 goroutine 永远阻塞，因为没有人发送数据
			<-ch
			fmt.Println("Never reached")
		}()
		// ch 被丢弃，但 goroutine 仍在运行
	}
}

// 修复: 使用 context 取消
func goroutineLeakFixed() {
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan int)

		go func() {
			select {
			case <-ch:
				fmt.Println("Received")
			case <-ctx.Done():
				return // 正常退出
			}
		}()

		// 模拟使用后取消
		time.Sleep(10 * time.Millisecond)
		cancel()
	}
}

// ========================================
// 问题 2: 无限增长的切片
// ========================================
type Cache struct {
	data []string
}

func (c *Cache) Add(item string) {
	c.data = append(c.data, item)
	// 问题: 只增不减，内存持续增长
}

// 修复: 添加容量限制和清理
type BoundedCache struct {
	data     []string
	maxSize  int
}

func NewBoundedCache(maxSize int) *BoundedCache {
	return &BoundedCache{
		data:    make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (c *BoundedCache) Add(item string) {
	if len(c.data) >= c.maxSize {
		// 移除最老的元素
		c.data = c.data[1:]
	}
	c.data = append(c.data, item)
}

func (c *BoundedCache) Clear() {
	c.data = c.data[:0]
}

// ========================================
// 问题 3: 未关闭的资源
// ========================================
func resourceLeakBug() {
	for i := 0; i < 100; i++ {
		resp, err := http.Get("http://example.com")
		if err != nil {
			continue
		}
		// 问题: 没有关闭 resp.Body
		_ = resp
	}
}

// 修复: 始终关闭资源
func resourceLeakFixed() {
	for i := 0; i < 100; i++ {
		resp, err := http.Get("http://example.com")
		if err != nil {
			continue
		}
		resp.Body.Close() // 关闭
	}
}

// 更好的做法: 使用 defer
func resourceLeakBetter() {
	resp, err := http.Get("http://example.com")
	if err != nil {
		return
	}
	defer resp.Body.Close() // 确保关闭
	// 使用 resp...
}

// ========================================
// 问题 4: 闭包持有引用
// ========================================
func closureLeakBug() {
	var funcs []func()

	for i := 0; i < 1000; i++ {
		largeData := make([]byte, 1024*1024) // 1MB
		funcs = append(funcs, func() {
			// 闭包持有 largeData 引用，即使不使用
			_ = len(largeData)
		})
	}
	// funcs 保存了所有闭包，它们都持有各自的 largeData
}

// 修复: 只捕获需要的值
func closureLeakFixed() {
	var funcs []func()

	for i := 0; i < 1000; i++ {
		largeData := make([]byte, 1024*1024)
		dataLen := len(largeData) // 只保存需要的值
		largeData = nil           // 允许 GC 回收

		funcs = append(funcs, func() {
			_ = dataLen
		})
	}
}

// ========================================
// 内存分析工具
// ========================================
func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB\n", m.Alloc/1024/1024)
	fmt.Printf("TotalAlloc = %v MiB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("Sys = %v MiB\n", m.Sys/1024/1024)
	fmt.Printf("NumGC = %v\n", m.NumGC)
	fmt.Printf("NumGoroutine = %v\n", runtime.NumGoroutine())
}

// ========================================
// 调试说明
// ========================================
/*
内存泄漏类型:
1. Goroutine 泄漏: goroutine 无法退出
2. 无限增长的数据结构
3. 未关闭的资源 (文件、连接)
4. 闭包持有大对象引用

检测方法:

1. 运行时内存统计:
   runtime.ReadMemStats(&m)

2. pprof HTTP 端点:
   go func() {
       http.ListenAndServe(":6060", nil)
   }()

   访问:
   http://localhost:6060/debug/pprof/heap
   http://localhost:6060/debug/pprof/goroutine

3. 命令行分析:
   go tool pprof http://localhost:6060/debug/pprof/heap
   (pprof) top10
   (pprof) list functionName
   (pprof) web

4. 查看 goroutine 数量:
   runtime.NumGoroutine()

常见修复方法:
1. 使用 context 控制 goroutine 生命周期
2. 为数据结构设置容量上限
3. 使用 defer 确保资源关闭
4. 避免闭包捕获大对象
*/

func main() {
	// 启动 pprof 服务
	go func() {
		fmt.Println("pprof available at http://localhost:6060/debug/pprof/")
		http.ListenAndServe(":6060", nil)
	}()

	fmt.Println("=== 内存泄漏演示 ===")

	fmt.Println("\n--- 初始状态 ---")
	printMemStats()

	fmt.Println("\n--- 运行修复后的代码 ---")
	goroutineLeakFixed()
	runtime.GC() // 强制 GC

	fmt.Println("\n--- GC 后状态 ---")
	printMemStats()

	// 保持程序运行以便使用 pprof
	fmt.Println("\n程序运行中，可使用 pprof 分析...")
	fmt.Println("尝试: go tool pprof http://localhost:6060/debug/pprof/heap")

	select {}
}
