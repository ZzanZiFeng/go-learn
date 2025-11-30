// Package main demonstrates goroutines in Go
package main

import (
	"fmt"
	"sync"
	"time"
)

// Basic goroutine example
func sayHello(name string) {
	fmt.Printf("Hello, %s!\n", name)
}

// Worker with WaitGroup
func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Worker %d starting\n", id)
	time.Sleep(time.Duration(id*100) * time.Millisecond)
	fmt.Printf("Worker %d done\n", id)
}

// Worker that returns result via channel
func squareWorker(n int, results chan<- int) {
	time.Sleep(100 * time.Millisecond)
	results <- n * n
}

func main() {
	fmt.Println("=== Goroutines Demo ===")

	// Basic goroutine
	fmt.Println("\n--- Basic Goroutine ---")
	go sayHello("Alice")
	go sayHello("Bob")
	go sayHello("Charlie")

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Main: basic goroutines completed")
	// Output (order may vary):
	// Hello, Alice!
	// Hello, Bob!
	// Hello, Charlie!
	// Main: basic goroutines completed

	// Anonymous goroutine
	fmt.Println("\n--- Anonymous Goroutine ---")
	go func() {
		fmt.Println("Anonymous goroutine running!")
	}()
	time.Sleep(50 * time.Millisecond)
	// Output: Anonymous goroutine running!

	// Anonymous goroutine with parameter
	fmt.Println("\n--- Anonymous Goroutine with Parameter ---")
	for i := 0; i < 3; i++ {
		go func(n int) {
			fmt.Printf("Goroutine %d\n", n)
		}(i) // Pass i as parameter to avoid closure issue
	}
	time.Sleep(100 * time.Millisecond)
	// Output (order may vary):
	// Goroutine 0
	// Goroutine 1
	// Goroutine 2

	// WaitGroup example
	fmt.Println("\n--- WaitGroup Example ---")
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}

	wg.Wait()
	fmt.Println("All workers completed")
	// Output:
	// Worker 1 starting
	// Worker 2 starting
	// Worker 3 starting
	// Worker 1 done
	// Worker 2 done
	// Worker 3 done
	// All workers completed

	// Goroutine with channel for results
	fmt.Println("\n--- Goroutine with Channel ---")
	results := make(chan int, 5)

	for i := 1; i <= 5; i++ {
		go squareWorker(i, results)
	}

	// Collect results
	for i := 0; i < 5; i++ {
		result := <-results
		fmt.Printf("Result: %d\n", result)
	}
	// Output (order may vary):
	// Result: 1
	// Result: 4
	// Result: 9
	// Result: 16
	// Result: 25

	// Demonstrating concurrent vs sequential
	fmt.Println("\n--- Concurrent vs Sequential ---")

	// Sequential
	start := time.Now()
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("Sequential: %v\n", time.Since(start))
	// Output: Sequential: ~300ms

	// Concurrent
	start = time.Now()
	var wg2 sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			time.Sleep(100 * time.Millisecond)
		}()
	}
	wg2.Wait()
	fmt.Printf("Concurrent: %v\n", time.Since(start))
	// Output: Concurrent: ~100ms

	// Done channel pattern
	fmt.Println("\n--- Done Channel Pattern ---")
	done := make(chan struct{})

	go func() {
		fmt.Println("Worker: starting work")
		time.Sleep(200 * time.Millisecond)
		fmt.Println("Worker: work completed")
		close(done)
	}()

	fmt.Println("Main: waiting for worker")
	<-done
	fmt.Println("Main: worker finished")
	// Output:
	// Main: waiting for worker
	// Worker: starting work
	// Worker: work completed
	// Main: worker finished

	// Multiple goroutines with shared counter (wrong way)
	fmt.Println("\n--- Race Condition Demo (Don't do this!) ---")
	var counter int
	var wg3 sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			counter++ // Race condition!
		}()
	}
	wg3.Wait()
	fmt.Printf("Counter (may be wrong): %d\n", counter)
	// Output: Counter (may be wrong): varies each run

	// Fixed with mutex
	fmt.Println("\n--- Fixed with Mutex ---")
	var safeCounter int
	var mu sync.Mutex
	var wg4 sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg4.Add(1)
		go func() {
			defer wg4.Done()
			mu.Lock()
			safeCounter++
			mu.Unlock()
		}()
	}
	wg4.Wait()
	fmt.Printf("Safe counter: %d\n", safeCounter)
	// Output: Safe counter: 100

	fmt.Println("\n=== Goroutines Demo Complete ===")
}
