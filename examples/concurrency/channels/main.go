// Package main demonstrates channels in Go
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== Channels Demo ===")

	// Basic channel usage
	fmt.Println("\n--- Basic Channel ---")
	ch := make(chan string)
	go func() {
		ch <- "Hello from goroutine!"
	}()

	msg := <-ch
	fmt.Println("Received:", msg)
	// Output: Received: Hello from goroutine!

	// Sending multiple values
	fmt.Println("\n--- Multiple Values ---")
	numbers := make(chan int)

	go func() {
		for i := 1; i <= 5; i++ {
			numbers <- i
		}
		close(numbers)
	}()

	for num := range numbers {
		fmt.Println("Received:", num)
	}
	// Output:
	// Received: 1
	// Received: 2
	// Received: 3
	// Received: 4
	// Received: 5

	// Buffered channel
	fmt.Println("\n--- Buffered Channel ---")
	buffered := make(chan int, 3)

	buffered <- 1
	buffered <- 2
	buffered <- 3
	// buffered <- 4 // Would block if uncommented

	fmt.Println("Buffer length:", len(buffered))
	fmt.Println("Buffer capacity:", cap(buffered))
	// Output:
	// Buffer length: 3
	// Buffer capacity: 3

	fmt.Println(<-buffered)
	fmt.Println(<-buffered)
	fmt.Println(<-buffered)
	// Output: 1, 2, 3

	// Channel direction - send only
	fmt.Println("\n--- Channel Directions ---")
	sendOnly := func(ch chan<- int) {
		ch <- 42
	}

	receiveOnly := func(ch <-chan int) int {
		return <-ch
	}

	bidirectional := make(chan int, 1)
	sendOnly(bidirectional)
	result := receiveOnly(bidirectional)
	fmt.Println("Result:", result)
	// Output: Result: 42

	// Closing channels
	fmt.Println("\n--- Closing Channels ---")
	jobs := make(chan int, 3)
	jobs <- 1
	jobs <- 2
	close(jobs)

	for {
		job, ok := <-jobs
		if !ok {
			fmt.Println("Channel closed")
			break
		}
		fmt.Println("Job:", job)
	}
	// Output:
	// Job: 1
	// Job: 2
	// Channel closed

	// Select statement
	fmt.Println("\n--- Select Statement ---")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "one"
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		ch2 <- "two"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Println("Received from ch1:", msg1)
		case msg2 := <-ch2:
			fmt.Println("Received from ch2:", msg2)
		}
	}
	// Output:
	// Received from ch1: one
	// Received from ch2: two

	// Non-blocking select with default
	fmt.Println("\n--- Non-blocking Select ---")
	nonBlocking := make(chan int)

	select {
	case val := <-nonBlocking:
		fmt.Println("Received:", val)
	default:
		fmt.Println("No value available")
	}
	// Output: No value available

	// Timeout with select
	fmt.Println("\n--- Timeout ---")
	slowCh := make(chan string)

	go func() {
		time.Sleep(2 * time.Second)
		slowCh <- "done"
	}()

	select {
	case msg := <-slowCh:
		fmt.Println("Received:", msg)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("Timeout!")
	}
	// Output: Timeout!

	// Done channel for signaling
	fmt.Println("\n--- Done Channel ---")
	done := make(chan struct{})

	go func() {
		fmt.Println("Worker: working...")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Worker: done")
		close(done)
	}()

	<-done
	fmt.Println("Main: received done signal")
	// Output:
	// Worker: working...
	// Worker: done
	// Main: received done signal

	// Fan-out: one sender, multiple receivers
	fmt.Println("\n--- Fan-out ---")
	fanOut := make(chan int)

	// Multiple receivers
	for i := 1; i <= 3; i++ {
		go func(id int) {
			for val := range fanOut {
				fmt.Printf("Worker %d received: %d\n", id, val)
			}
		}(i)
	}

	// Single sender
	for i := 1; i <= 6; i++ {
		fanOut <- i
	}
	close(fanOut)
	time.Sleep(100 * time.Millisecond)
	// Output (workers get different values):
	// Worker 1 received: 1
	// Worker 2 received: 2
	// etc.

	// Fan-in: multiple senders, one receiver
	fmt.Println("\n--- Fan-in ---")
	fanIn := make(chan string)

	// Multiple senders
	go func() { fanIn <- "A1"; fanIn <- "A2" }()
	go func() { fanIn <- "B1"; fanIn <- "B2" }()

	// Single receiver
	for i := 0; i < 4; i++ {
		fmt.Println("Received:", <-fanIn)
	}
	// Output (order may vary):
	// Received: A1
	// Received: B1
	// etc.

	// Pipeline pattern
	fmt.Println("\n--- Pipeline ---")

	// Stage 1: Generate numbers
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

	// Stage 2: Square numbers
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

	// Run pipeline
	for n := range square(gen(1, 2, 3, 4, 5)) {
		fmt.Println("Squared:", n)
	}
	// Output:
	// Squared: 1
	// Squared: 4
	// Squared: 9
	// Squared: 16
	// Squared: 25

	fmt.Println("\n=== Channels Demo Complete ===")
}
