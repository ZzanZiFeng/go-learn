// Package main demonstrates sync package in Go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== Sync Package Demo ===")

	// WaitGroup
	fmt.Println("\n--- WaitGroup ---")
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Worker %d starting\n", id)
			time.Sleep(time.Duration(id*50) * time.Millisecond)
			fmt.Printf("Worker %d done\n", id)
		}(i)
	}

	wg.Wait()
	fmt.Println("All workers completed")
	// Output: All workers done in order by completion time

	// Mutex
	fmt.Println("\n--- Mutex ---")
	var (
		counter int
		mu      sync.Mutex
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("Counter:", counter)
	// Output: Counter: 100

	// RWMutex
	fmt.Println("\n--- RWMutex ---")
	type SafeMap struct {
		mu   sync.RWMutex
		data map[string]int
	}

	safeMap := &SafeMap{data: make(map[string]int)}

	// Writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			safeMap.mu.Lock()
			safeMap.data[fmt.Sprintf("key%d", i)] = i
			safeMap.mu.Unlock()
		}(i)
	}

	// Readers (can read concurrently)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			safeMap.mu.RLock()
			_ = len(safeMap.data)
			safeMap.mu.RUnlock()
		}()
	}

	wg.Wait()
	fmt.Println("Map size:", len(safeMap.data))
	// Output: Map size: 5

	// Once
	fmt.Println("\n--- Once ---")
	var once sync.Once
	var initialized bool

	initialize := func() {
		fmt.Println("Initializing...")
		initialized = true
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			once.Do(initialize)
			fmt.Printf("Goroutine %d: initialized = %v\n", id, initialized)
		}(i)
	}

	wg.Wait()
	// Output: "Initializing..." only once

	// Cond
	fmt.Println("\n--- Cond ---")
	var (
		ready    bool
		condLock sync.Mutex
		cond     = sync.NewCond(&condLock)
	)

	// Waiting goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		condLock.Lock()
		for !ready {
			cond.Wait()
		}
		condLock.Unlock()
		fmt.Println("Waiter: condition is ready!")
	}()

	// Signal goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		condLock.Lock()
		ready = true
		cond.Signal()
		condLock.Unlock()
		fmt.Println("Signaler: signaled condition")
	}()

	wg.Wait()
	// Output:
	// Signaler: signaled condition
	// Waiter: condition is ready!

	// Broadcast
	fmt.Println("\n--- Cond Broadcast ---")
	ready = false

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			condLock.Lock()
			for !ready {
				cond.Wait()
			}
			condLock.Unlock()
			fmt.Printf("Waiter %d: received broadcast\n", id)
		}(i)
	}

	time.Sleep(50 * time.Millisecond)
	condLock.Lock()
	ready = true
	cond.Broadcast() // Wake all waiters
	condLock.Unlock()

	wg.Wait()
	// Output: All 3 waiters receive the broadcast

	// sync.Map
	fmt.Println("\n--- sync.Map ---")
	var m sync.Map

	// Store values
	m.Store("key1", "value1")
	m.Store("key2", 42)
	m.Store("key3", true)

	// Load values
	if val, ok := m.Load("key1"); ok {
		fmt.Println("key1:", val)
	}
	// Output: key1: value1

	// LoadOrStore
	actual, loaded := m.LoadOrStore("key4", "new value")
	fmt.Printf("key4: %v, existed: %v\n", actual, loaded)
	// Output: key4: new value, existed: false

	// Range
	fmt.Println("All entries:")
	m.Range(func(key, value interface{}) bool {
		fmt.Printf("  %v: %v\n", key, value)
		return true
	})

	// Delete
	m.Delete("key1")

	// sync.Pool
	fmt.Println("\n--- sync.Pool ---")
	pool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("Creating new object")
			return make([]byte, 1024)
		},
	}

	// First Get creates new object
	buf1 := pool.Get().([]byte)
	fmt.Printf("Got buffer of size %d\n", len(buf1))

	// Put back
	pool.Put(buf1)

	// Second Get reuses object
	buf2 := pool.Get().([]byte)
	fmt.Printf("Got buffer (reused): %p == %p: %v\n", &buf1[0], &buf2[0], &buf1[0] == &buf2[0])
	// Output: addresses may be same (reused)

	// Atomic operations
	fmt.Println("\n--- Atomic Operations ---")
	var atomicCounter int64

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&atomicCounter, 1)
		}()
	}

	wg.Wait()
	fmt.Println("Atomic counter:", atomic.LoadInt64(&atomicCounter))
	// Output: Atomic counter: 100

	// Compare and Swap
	var value int64 = 10
	swapped := atomic.CompareAndSwapInt64(&value, 10, 20)
	fmt.Printf("CAS: swapped=%v, value=%d\n", swapped, value)
	// Output: CAS: swapped=true, value=20

	swapped = atomic.CompareAndSwapInt64(&value, 10, 30) // Won't swap
	fmt.Printf("CAS: swapped=%v, value=%d\n", swapped, value)
	// Output: CAS: swapped=false, value=20

	// Atomic.Value for storing arbitrary values
	fmt.Println("\n--- atomic.Value ---")
	var config atomic.Value

	type Config struct {
		Host string
		Port int
	}

	config.Store(Config{Host: "localhost", Port: 8080})

	// Multiple goroutines can safely read
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg := config.Load().(Config)
			fmt.Printf("Goroutine %d: config = %+v\n", id, cfg)
		}(i)
	}

	wg.Wait()

	// Update config atomically
	config.Store(Config{Host: "0.0.0.0", Port: 9090})
	fmt.Println("New config:", config.Load().(Config))

	fmt.Println("\n=== Sync Package Demo Complete ===")
}
