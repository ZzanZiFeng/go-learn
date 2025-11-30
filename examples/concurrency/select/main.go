// Package main demonstrates select statement in Go
package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("=== Select Demo ===")

	// Basic select
	fmt.Println("\n--- Basic Select ---")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "first"
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "second"
	}()

	select {
	case msg := <-ch1:
		fmt.Println("Received from ch1:", msg)
	case msg := <-ch2:
		fmt.Println("Received from ch2:", msg)
	}
	// Output: Received from ch1: first

	// Select with timeout
	fmt.Println("\n--- Select with Timeout ---")
	slowCh := make(chan string)

	go func() {
		time.Sleep(500 * time.Millisecond)
		slowCh <- "slow result"
	}()

	select {
	case msg := <-slowCh:
		fmt.Println("Received:", msg)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Timeout after 100ms")
	}
	// Output: Timeout after 100ms

	// Non-blocking with default
	fmt.Println("\n--- Non-blocking Operations ---")
	ch := make(chan int, 1)

	// Non-blocking receive
	select {
	case val := <-ch:
		fmt.Println("Received:", val)
	default:
		fmt.Println("No value to receive")
	}
	// Output: No value to receive

	// Non-blocking send
	ch <- 42
	select {
	case ch <- 100:
		fmt.Println("Sent 100")
	default:
		fmt.Println("Channel full, could not send")
	}
	// Output: Channel full, could not send

	fmt.Println("Value in channel:", <-ch)
	// Output: Value in channel: 42

	// Multiple channel waiting
	fmt.Println("\n--- Multiple Channels ---")
	c1 := make(chan string)
	c2 := make(chan string)
	done := make(chan struct{})

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(50 * time.Millisecond)
			c1 <- fmt.Sprintf("c1-%d", i)
		}
	}()

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(80 * time.Millisecond)
			c2 <- fmt.Sprintf("c2-%d", i)
		}
	}()

	go func() {
		time.Sleep(300 * time.Millisecond)
		close(done)
	}()

	for {
		select {
		case msg := <-c1:
			fmt.Println("From c1:", msg)
		case msg := <-c2:
			fmt.Println("From c2:", msg)
		case <-done:
			fmt.Println("Done!")
			goto afterLoop
		}
	}
afterLoop:
	// Output (interleaved):
	// From c1: c1-0
	// From c2: c2-0
	// From c1: c1-1
	// etc.

	// Random selection when multiple ready
	fmt.Println("\n--- Random Selection ---")
	r1 := make(chan int, 10)
	r2 := make(chan int, 10)

	// Fill both channels
	for i := 0; i < 5; i++ {
		r1 <- i
		r2 <- i + 100
	}

	counts := make(map[string]int)
	for i := 0; i < 10; i++ {
		select {
		case <-r1:
			counts["r1"]++
		case <-r2:
			counts["r2"]++
		}
	}
	fmt.Printf("Selection counts: r1=%d, r2=%d\n", counts["r1"], counts["r2"])
	// Output: varies, roughly even distribution

	// Ticker for periodic events
	fmt.Println("\n--- Ticker ---")
	ticker := time.NewTicker(100 * time.Millisecond)
	tickDone := make(chan struct{})

	go func() {
		time.Sleep(350 * time.Millisecond)
		close(tickDone)
	}()

	for {
		select {
		case t := <-ticker.C:
			fmt.Println("Tick at", t.Format("15:04:05.000"))
		case <-tickDone:
			ticker.Stop()
			fmt.Println("Ticker stopped")
			goto afterTicker
		}
	}
afterTicker:
	// Output:
	// Tick at HH:MM:SS.mmm
	// Tick at HH:MM:SS.mmm
	// Tick at HH:MM:SS.mmm
	// Ticker stopped

	// Priority select pattern
	fmt.Println("\n--- Priority Pattern ---")
	high := make(chan string, 5)
	low := make(chan string, 5)

	high <- "HIGH 1"
	low <- "low 1"
	low <- "low 2"
	high <- "HIGH 2"
	low <- "low 3"

	for i := 0; i < 5; i++ {
		select {
		case msg := <-high:
			fmt.Println("Priority:", msg)
		default:
			select {
			case msg := <-high:
				fmt.Println("Priority:", msg)
			case msg := <-low:
				fmt.Println("Normal:", msg)
			}
		}
	}
	// High priority messages are handled first

	// Heartbeat pattern
	fmt.Println("\n--- Heartbeat Pattern ---")
	doWork := func(done <-chan struct{}) <-chan struct{} {
		heartbeat := make(chan struct{}, 1)
		go func() {
			defer close(heartbeat)
			for {
				select {
				case <-done:
					return
				case <-time.After(50 * time.Millisecond):
					select {
					case heartbeat <- struct{}{}:
					default:
					}
				}
			}
		}()
		return heartbeat
	}

	workDone := make(chan struct{})
	heartbeat := doWork(workDone)

	timeout := time.After(200 * time.Millisecond)
	beats := 0
heartbeatLoop:
	for {
		select {
		case <-heartbeat:
			beats++
			fmt.Printf("Heartbeat %d received\n", beats)
		case <-timeout:
			close(workDone)
			fmt.Printf("Total heartbeats: %d\n", beats)
			break heartbeatLoop
		}
	}

	// Quit channel pattern
	fmt.Println("\n--- Quit Channel ---")
	quit := make(chan struct{})
	data := make(chan int)

	// Producer
	go func() {
		for i := 0; ; i++ {
			select {
			case data <- i:
			case <-quit:
				fmt.Println("Producer: received quit signal")
				return
			}
		}
	}()

	// Consume a few values
	for i := 0; i < 5; i++ {
		fmt.Println("Received:", <-data)
	}
	close(quit)
	time.Sleep(50 * time.Millisecond)
	// Output:
	// Received: 0
	// Received: 1
	// Received: 2
	// Received: 3
	// Received: 4
	// Producer: received quit signal

	// Dynamic channel selection (nil channel trick)
	fmt.Println("\n--- Nil Channel Trick ---")
	var activeCh chan int
	enableCh := make(chan int, 1)

	enableCh <- 42
	activeCh = nil // Disabled initially

	select {
	case val := <-activeCh: // Will never execute because activeCh is nil
		fmt.Println("Active:", val)
	case val := <-enableCh:
		fmt.Println("Enable:", val)
	}
	// Output: Enable: 42

	// Simulate enabling the channel
	activeCh = make(chan int, 1)
	activeCh <- 100

	select {
	case val := <-activeCh:
		fmt.Println("Active:", val)
	default:
		fmt.Println("Nothing ready")
	}
	// Output: Active: 100

	fmt.Println("\n=== Select Demo Complete ===")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
