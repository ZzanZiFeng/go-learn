// Package main demonstrates common concurrency patterns in Go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Job represents a unit of work
type Job struct {
	ID   int
	Data string
}

// Result represents the output of processing a Job
type Result struct {
	JobID  int
	Output string
}

func main() {
	fmt.Println("=== Concurrency Patterns Demo ===")

	// Worker Pool Pattern
	fmt.Println("\n--- Worker Pool ---")
	workerPoolDemo()

	// Fan-out/Fan-in Pattern
	fmt.Println("\n--- Fan-out/Fan-in ---")
	fanOutFanInDemo()

	// Pipeline Pattern
	fmt.Println("\n--- Pipeline ---")
	pipelineDemo()

	// Generator Pattern
	fmt.Println("\n--- Generator ---")
	generatorDemo()

	// Semaphore Pattern
	fmt.Println("\n--- Semaphore ---")
	semaphoreDemo()

	// Rate Limiter Pattern
	fmt.Println("\n--- Rate Limiter ---")
	rateLimiterDemo()

	fmt.Println("\n=== Patterns Demo Complete ===")
}

// Worker Pool Pattern
func workerPoolDemo() {
	const numWorkers = 3
	const numJobs = 9

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				// Process job
				time.Sleep(50 * time.Millisecond)
				results <- Result{
					JobID:  job.ID,
					Output: fmt.Sprintf("Worker %d processed: %s", id, job.Data),
				}
			}
		}(w)
	}

	// Send jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- Job{ID: j, Data: fmt.Sprintf("job-%d", j)}
	}
	close(jobs)

	// Wait and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		fmt.Printf("Job %d: %s\n", result.JobID, result.Output)
	}
}

// Fan-out/Fan-in Pattern
func fanOutFanInDemo() {
	// Generator
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for _, n := range nums {
				out <- n
			}
		}()
		return out
	}

	// Fan-out: distribute work
	square := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				out <- n * n
			}
		}()
		return out
	}

	// Fan-in: merge results
	merge := func(channels ...<-chan int) <-chan int {
		out := make(chan int)
		var wg sync.WaitGroup

		for _, ch := range channels {
			wg.Add(1)
			go func(c <-chan int) {
				defer wg.Done()
				for n := range c {
					out <- n
				}
			}(ch)
		}

		go func() {
			wg.Wait()
			close(out)
		}()

		return out
	}

	// Create pipeline
	in := gen(1, 2, 3, 4, 5)

	// Fan-out to 2 workers
	sq1 := square(in)
	sq2 := square(in)

	// Fan-in results
	for n := range merge(sq1, sq2) {
		fmt.Println("Squared:", n)
	}
}

// Pipeline Pattern
func pipelineDemo() {
	// Stage 1: Generate
	generate := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for _, n := range nums {
				out <- n
			}
		}()
		return out
	}

	// Stage 2: Double
	double := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				out <- n * 2
			}
		}()
		return out
	}

	// Stage 3: Add ten
	addTen := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				out <- n + 10
			}
		}()
		return out
	}

	// Stage 4: Format
	format := func(in <-chan int) <-chan string {
		out := make(chan string)
		go func() {
			defer close(out)
			for n := range in {
				out <- fmt.Sprintf("Result: %d", n)
			}
		}()
		return out
	}

	// Build pipeline
	nums := generate(1, 2, 3, 4, 5)
	doubled := double(nums)
	added := addTen(doubled)
	formatted := format(added)

	// Consume
	for s := range formatted {
		fmt.Println(s)
	}
	// Output: 1*2+10=12, 2*2+10=14, etc.
}

// Generator Pattern
func generatorDemo() {
	// Finite generator
	countTo := func(n int) <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			for i := 1; i <= n; i++ {
				ch <- i
			}
		}()
		return ch
	}

	fmt.Print("Count to 5: ")
	for n := range countTo(5) {
		fmt.Printf("%d ", n)
	}
	fmt.Println()

	// Infinite generator with cancellation
	fibonacci := func(ctx context.Context) <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			a, b := 0, 1
			for {
				select {
				case <-ctx.Done():
					return
				case ch <- a:
					a, b = b, a+b
				}
			}
		}()
		return ch
	}

	ctx, cancel := context.WithCancel(context.Background())
	fib := fibonacci(ctx)

	fmt.Print("First 10 Fibonacci: ")
	for i := 0; i < 10; i++ {
		fmt.Printf("%d ", <-fib)
	}
	fmt.Println()
	cancel() // Stop generator
}

// Semaphore Pattern
func semaphoreDemo() {
	// Limit concurrent operations to 2
	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem <- struct{}{} // Acquire
			defer func() { <-sem }() // Release

			fmt.Printf("Worker %d: starting at %s\n", id, time.Now().Format("15:04:05.000"))
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("Worker %d: done\n", id)
		}(i)
	}

	wg.Wait()
	// Only 2 workers run at a time
}

// Rate Limiter Pattern
func rateLimiterDemo() {
	// Token bucket rate limiter
	type RateLimiter struct {
		tokens chan struct{}
	}

	newLimiter := func(rate int, interval time.Duration) *RateLimiter {
		rl := &RateLimiter{
			tokens: make(chan struct{}, rate),
		}

		// Pre-fill tokens
		for i := 0; i < rate; i++ {
			rl.tokens <- struct{}{}
		}

		// Refill tokens
		go func() {
			ticker := time.NewTicker(interval / time.Duration(rate))
			defer ticker.Stop()
			for range ticker.C {
				select {
				case rl.tokens <- struct{}{}:
				default:
				}
			}
		}()

		return rl
	}

	// Create limiter: 5 requests per second
	limiter := newLimiter(5, time.Second)

	// Make 10 requests
	for i := 1; i <= 10; i++ {
		<-limiter.tokens // Wait for token
		fmt.Printf("Request %d at %s\n", i, time.Now().Format("15:04:05.000"))
	}
}

// Or-Done Channel Pattern
func orDone(done <-chan struct{}, c <-chan int) <-chan int {
	valStream := make(chan int)
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case valStream <- v:
				case <-done:
					return
				}
			}
		}
	}()
	return valStream
}

// Bridge Channel Pattern - flattens channels of channels
func bridge(done <-chan struct{}, chanStream <-chan <-chan int) <-chan int {
	valStream := make(chan int)
	go func() {
		defer close(valStream)
		for {
			var stream <-chan int
			select {
			case maybeStream, ok := <-chanStream:
				if !ok {
					return
				}
				stream = maybeStream
			case <-done:
				return
			}
			for val := range orDone(done, stream) {
				select {
				case valStream <- val:
				case <-done:
					return
				}
			}
		}
	}()
	return valStream
}
