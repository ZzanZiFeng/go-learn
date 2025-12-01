// tests/debugging/bug-03-deadlock/main.go
// Bug 场景 3: 死锁
// Go 运行时会检测到某些死锁并 panic

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========================================
// 问题 1: Channel 死锁
// ========================================
func channelDeadlock() {
	ch := make(chan int)
	ch <- 1 // 死锁: 无缓冲 channel，没有接收者
	fmt.Println(<-ch)
}

/*
输出:
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.channelDeadlock()
*/

// 修复: 使用 goroutine 或 buffered channel
func channelFixed1() {
	ch := make(chan int, 1) // buffered channel
	ch <- 1
	fmt.Println(<-ch)
}

func channelFixed2() {
	ch := make(chan int)
	go func() {
		ch <- 1 // 在 goroutine 中发送
	}()
	fmt.Println(<-ch)
}

// ========================================
// 问题 2: Mutex 死锁 (重复锁定)
// ========================================
func mutexDeadlock() {
	var mu sync.Mutex

	mu.Lock()
	fmt.Println("First lock acquired")

	// 再次尝试获取同一个锁 - 死锁
	mu.Lock() // 永远阻塞
	fmt.Println("Second lock acquired")

	mu.Unlock()
	mu.Unlock()
}

// 修复: 确保每个 Lock 对应一个 Unlock
func mutexFixed() {
	var mu sync.Mutex

	mu.Lock()
	fmt.Println("Lock acquired")
	mu.Unlock()

	mu.Lock()
	fmt.Println("Lock acquired again")
	mu.Unlock()
}

// ========================================
// 问题 3: 循环等待死锁
// ========================================
func circularDeadlock() {
	var mu1, mu2 sync.Mutex
	var wg sync.WaitGroup

	wg.Add(2)

	// Goroutine 1: 先锁 mu1，再锁 mu2
	go func() {
		defer wg.Done()
		mu1.Lock()
		time.Sleep(100 * time.Millisecond)
		mu2.Lock() // 等待 mu2，但 goroutine 2 持有 mu2
		mu2.Unlock()
		mu1.Unlock()
	}()

	// Goroutine 2: 先锁 mu2，再锁 mu1
	go func() {
		defer wg.Done()
		mu2.Lock()
		time.Sleep(100 * time.Millisecond)
		mu1.Lock() // 等待 mu1，但 goroutine 1 持有 mu1
		mu1.Unlock()
		mu2.Unlock()
	}()

	wg.Wait()
}

// 修复: 统一锁的获取顺序
func circularFixed() {
	var mu1, mu2 sync.Mutex
	var wg sync.WaitGroup

	wg.Add(2)

	// 两个 goroutine 都按相同顺序获取锁
	go func() {
		defer wg.Done()
		mu1.Lock()
		defer mu1.Unlock()
		time.Sleep(100 * time.Millisecond)
		mu2.Lock()
		defer mu2.Unlock()
		fmt.Println("Goroutine 1 completed")
	}()

	go func() {
		defer wg.Done()
		mu1.Lock() // 先获取 mu1
		defer mu1.Unlock()
		time.Sleep(100 * time.Millisecond)
		mu2.Lock() // 再获取 mu2
		defer mu2.Unlock()
		fmt.Println("Goroutine 2 completed")
	}()

	wg.Wait()
}

// ========================================
// 问题 4: WaitGroup 使用错误
// ========================================
func waitGroupDeadlock() {
	var wg sync.WaitGroup

	wg.Add(1)
	// 忘记调用 wg.Done()

	wg.Wait() // 永远等待
}

// 修复: 确保 Done 被调用
func waitGroupFixed() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done() // 使用 defer 确保调用
		fmt.Println("Work done")
	}()

	wg.Wait()
}

// ========================================
// 使用 context 超时避免死锁
// ========================================
func withTimeout() {
	ch := make(chan int)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case val := <-ch:
		fmt.Println("Received:", val)
	case <-ctx.Done():
		fmt.Println("Timeout - avoided deadlock")
	}
}

// ========================================
// 调试说明
// ========================================
/*
死锁类型:
1. Channel 死锁: 发送无接收者，接收无发送者
2. Mutex 死锁: 重复锁定，循环等待
3. WaitGroup 死锁: Add 和 Done 不匹配

检测方法:
1. Go 运行时会检测简单死锁并 panic
2. 使用 pprof goroutine 查看阻塞的 goroutine
3. 添加超时机制

调试命令:
$ curl http://localhost:6060/debug/pprof/goroutine?debug=1

预防措施:
1. 使用 defer 释放锁
2. 统一锁的获取顺序
3. 使用 context 添加超时
4. 避免在持有锁时调用外部函数
*/

func main() {
	fmt.Println("=== 死锁演示 ===")

	// 修复后的代码
	fmt.Println("\n--- Channel Fixed ---")
	channelFixed1()
	channelFixed2()

	fmt.Println("\n--- Mutex Fixed ---")
	mutexFixed()

	fmt.Println("\n--- Circular Fixed ---")
	circularFixed()

	fmt.Println("\n--- WaitGroup Fixed ---")
	waitGroupFixed()

	fmt.Println("\n--- With Timeout ---")
	withTimeout()

	// 取消注释以查看死锁
	// channelDeadlock()
	// mutexDeadlock()
	// circularDeadlock()
	// waitGroupDeadlock()
}
