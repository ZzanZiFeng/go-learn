// tests/benchmark/concurrent_bench_test.go
// 并发操作基准测试示例

package benchmark

import (
	"sync"
	"sync/atomic"
	"testing"
)

// 互斥锁 vs 读写锁 vs 原子操作

// 使用 Mutex
func BenchmarkMutex_Write(b *testing.B) {
	var mu sync.Mutex
	var counter int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			counter++
			mu.Unlock()
		}
	})
}

func BenchmarkMutex_Read(b *testing.B) {
	var mu sync.Mutex
	var counter int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			_ = counter
			mu.Unlock()
		}
	})
}

// 使用 RWMutex
func BenchmarkRWMutex_Write(b *testing.B) {
	var mu sync.RWMutex
	var counter int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			counter++
			mu.Unlock()
		}
	})
}

func BenchmarkRWMutex_Read(b *testing.B) {
	var mu sync.RWMutex
	var counter int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			_ = counter
			mu.RUnlock()
		}
	})
}

// 使用原子操作
func BenchmarkAtomic_Add(b *testing.B) {
	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&counter, 1)
		}
	})
}

func BenchmarkAtomic_Load(b *testing.B) {
	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = atomic.LoadInt64(&counter)
		}
	})
}

// 使用 Channel
func BenchmarkChannel_Unbuffered(b *testing.B) {
	ch := make(chan struct{})
	go func() {
		for range ch {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
	}
}

func BenchmarkChannel_Buffered(b *testing.B) {
	ch := make(chan struct{}, 100)
	go func() {
		for range ch {
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
	}
}

// sync.Pool 基准测试
type Buffer struct {
	data []byte
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{data: make([]byte, 1024)}
	},
}

func BenchmarkWithoutPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := &Buffer{data: make([]byte, 1024)}
		_ = buf
	}
}

func BenchmarkWithPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := bufferPool.Get().(*Buffer)
		bufferPool.Put(buf)
	}
}

// sync.Map vs 普通 map + Mutex
func BenchmarkMapWithMutex_Write(b *testing.B) {
	var mu sync.Mutex
	m := make(map[int]int)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mu.Lock()
			m[i%100] = i
			mu.Unlock()
			i++
		}
	})
}

func BenchmarkSyncMap_Write(b *testing.B) {
	var m sync.Map

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Store(i%100, i)
			i++
		}
	})
}

func BenchmarkMapWithMutex_Read(b *testing.B) {
	var mu sync.Mutex
	m := make(map[int]int)
	for i := 0; i < 100; i++ {
		m[i] = i
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mu.Lock()
			_ = m[i%100]
			mu.Unlock()
			i++
		}
	})
}

func BenchmarkSyncMap_Read(b *testing.B) {
	var m sync.Map
	for i := 0; i < 100; i++ {
		m.Store(i, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Load(i % 100)
			i++
		}
	})
}
