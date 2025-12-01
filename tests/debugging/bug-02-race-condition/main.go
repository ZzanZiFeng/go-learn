// tests/debugging/bug-02-race-condition/main.go
// Bug 场景 2: 竞态条件
// 使用 go run -race main.go 检测

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========================================
// 问题代码: 竞态条件
// ========================================
func buggyCode() {
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // 竞态条件: 多个 goroutine 同时读写
		}()
	}

	wg.Wait()
	fmt.Printf("Counter (buggy): %d\n", counter) // 结果不确定
}

/*
运行 go run -race main.go 会输出:

==================
WARNING: DATA RACE
Read at 0x00c0000b4008 by goroutine 8:
  main.buggyCode.func1()
      /path/main.go:19 +0x...

Previous write at 0x00c0000b4008 by goroutine 7:
  main.buggyCode.func1()
      /path/main.go:19 +0x...
==================
*/

// ========================================
// 修复 1: 使用 Mutex
// ========================================
func fixedWithMutex() {
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Printf("Counter (mutex): %d\n", counter) // 总是 1000
}

// ========================================
// 修复 2: 使用 atomic
// ========================================
func fixedWithAtomic() {
	var counter int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()
	fmt.Printf("Counter (atomic): %d\n", counter) // 总是 1000
}

// ========================================
// 修复 3: 使用 channel
// ========================================
func fixedWithChannel() {
	counter := 0
	ch := make(chan int, 1000)
	var wg sync.WaitGroup

	// 接收者
	go func() {
		for range ch {
			counter++
		}
	}()

	// 发送者
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch <- 1
		}()
	}

	wg.Wait()
	close(ch)
	time.Sleep(10 * time.Millisecond) // 等待接收者处理完
	fmt.Printf("Counter (channel): %d\n", counter)
}

// ========================================
// 另一个常见竞态: 闭包捕获
// ========================================
func closureBug() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("Bug: i = %d\n", i) // 问题: 所有 goroutine 可能打印相同的值
		}()
	}
	wg.Wait()
}

func closureFixed() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) { // 修复: 将 i 作为参数传递
			defer wg.Done()
			fmt.Printf("Fixed: n = %d\n", n)
		}(i)
	}
	wg.Wait()
}

// ========================================
// 调试说明
// ========================================
/*
检测竞态条件:
$ go run -race main.go
$ go test -race ./...

关键点:
1. 多个 goroutine 同时访问共享变量
2. 至少有一个是写操作
3. 没有同步机制

修复方法:
1. sync.Mutex / sync.RWMutex
2. sync/atomic 包
3. channel 通信
4. 避免共享状态
*/

func main() {
	fmt.Println("=== 竞态条件演示 ===")
	fmt.Println("\n运行 'go run -race main.go' 查看竞态检测")

	// 问题代码
	fmt.Println("\n--- Buggy Code ---")
	buggyCode()

	// 修复方案
	fmt.Println("\n--- Fixed with Mutex ---")
	fixedWithMutex()

	fmt.Println("\n--- Fixed with Atomic ---")
	fixedWithAtomic()

	fmt.Println("\n--- Fixed with Channel ---")
	fixedWithChannel()

	// 闭包问题
	fmt.Println("\n--- Closure Bug ---")
	closureBug()

	fmt.Println("\n--- Closure Fixed ---")
	closureFixed()
}
