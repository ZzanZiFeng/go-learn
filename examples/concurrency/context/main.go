// Package main demonstrates context in Go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== Context Demo ===")

	// Background context
	fmt.Println("\n--- Background Context ---")
	ctx := context.Background()
	fmt.Printf("Background context: %v\n", ctx)
	// Output: Background context: context.Background

	// WithCancel
	fmt.Println("\n--- WithCancel ---")
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Worker: cancelled, reason:", ctx.Err())
				return
			default:
				fmt.Println("Worker: working...")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}(ctx)

	time.Sleep(350 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
	// Output:
	// Worker: working...
	// Worker: working...
	// Worker: working...
	// Worker: cancelled, reason: context canceled

	// WithTimeout
	fmt.Println("\n--- WithTimeout ---")
	ctx, cancel = context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	slowOperation := func(ctx context.Context) error {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := slowOperation(ctx)
	if err != nil {
		fmt.Println("Operation failed:", err)
	}
	// Output: Operation failed: context deadline exceeded

	// WithDeadline
	fmt.Println("\n--- WithDeadline ---")
	deadline := time.Now().Add(100 * time.Millisecond)
	ctx, cancel = context.WithDeadline(context.Background(), deadline)
	defer cancel()

	if d, ok := ctx.Deadline(); ok {
		fmt.Println("Deadline:", d.Format("15:04:05.000"))
	}

	<-ctx.Done()
	fmt.Println("Context done:", ctx.Err())
	// Output: Context done: context deadline exceeded

	// WithValue
	fmt.Println("\n--- WithValue ---")
	type contextKey string
	const (
		userIDKey    contextKey = "userID"
		requestIDKey contextKey = "requestID"
	)

	ctx = context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-123")
	ctx = context.WithValue(ctx, requestIDKey, "req-456")

	processRequest := func(ctx context.Context) {
		userID := ctx.Value(userIDKey)
		requestID := ctx.Value(requestIDKey)
		fmt.Printf("Processing: userID=%v, requestID=%v\n", userID, requestID)
	}

	processRequest(ctx)
	// Output: Processing: userID=user-123, requestID=req-456

	// Cascading cancellation
	fmt.Println("\n--- Cascading Cancellation ---")
	parentCtx, parentCancel := context.WithCancel(context.Background())

	childCtx, _ := context.WithCancel(parentCtx)
	grandchildCtx, _ := context.WithCancel(childCtx)

	go func() {
		<-grandchildCtx.Done()
		fmt.Println("Grandchild: cancelled")
	}()

	go func() {
		<-childCtx.Done()
		fmt.Println("Child: cancelled")
	}()

	time.Sleep(50 * time.Millisecond)
	parentCancel() // Cancels all descendants
	time.Sleep(50 * time.Millisecond)
	// Output:
	// Child: cancelled
	// Grandchild: cancelled

	// Practical example: HTTP-like request with timeout
	fmt.Println("\n--- Practical: Request with Timeout ---")

	fetchData := func(ctx context.Context, url string) (string, error) {
		resultCh := make(chan string, 1)
		errCh := make(chan error, 1)

		go func() {
			// Simulate fetch
			time.Sleep(150 * time.Millisecond)
			resultCh <- fmt.Sprintf("Data from %s", url)
		}()

		select {
		case result := <-resultCh:
			return result, nil
		case err := <-errCh:
			return "", err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Fast request
	ctx, cancel = context.WithTimeout(context.Background(), 200*time.Millisecond)
	result, err := fetchData(ctx, "http://fast.example.com")
	cancel()
	if err != nil {
		fmt.Println("Fast request failed:", err)
	} else {
		fmt.Println("Fast request result:", result)
	}
	// Output: Fast request result: Data from http://fast.example.com

	// Slow request (will timeout)
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	result, err = fetchData(ctx, "http://slow.example.com")
	cancel()
	if err != nil {
		fmt.Println("Slow request failed:", err)
	}
	// Output: Slow request failed: context deadline exceeded

	// Checking context errors
	fmt.Println("\n--- Checking Context Errors ---")
	ctx, cancel = context.WithCancel(context.Background())
	cancel()

	if errors.Is(ctx.Err(), context.Canceled) {
		fmt.Println("Context was cancelled")
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(time.Millisecond)
	cancel()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		fmt.Println("Context deadline exceeded")
	}
	// Output:
	// Context was cancelled
	// Context deadline exceeded

	// Multiple workers with shared cancellation
	fmt.Println("\n--- Multiple Workers with Cancellation ---")
	ctx, cancel = context.WithCancel(context.Background())

	for i := 1; i <= 3; i++ {
		go func(id int, ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("Worker %d: stopped\n", id)
					return
				default:
					fmt.Printf("Worker %d: tick\n", id)
					time.Sleep(50 * time.Millisecond)
				}
			}
		}(i, ctx)
	}

	time.Sleep(130 * time.Millisecond)
	fmt.Println("Cancelling all workers...")
	cancel()
	time.Sleep(50 * time.Millisecond)
	// Output: All workers stop after cancel

	// Context best practices
	fmt.Println("\n--- Best Practices ---")
	fmt.Println(`
1. Always call cancel() - use defer
2. Pass context as first parameter
3. Don't store context in structs
4. Use context.Value sparingly
5. Check ctx.Done() in long operations
`)

	// Example of proper context usage in a function
	fmt.Println("\n--- Proper Function Signature ---")
	type Result struct {
		Data string
		Err  error
	}

	doWork := func(ctx context.Context, input string) Result {
		// Check if already cancelled
		if ctx.Err() != nil {
			return Result{Err: ctx.Err()}
		}

		// Do work with cancellation check
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				return Result{Err: ctx.Err()}
			case <-time.After(20 * time.Millisecond):
				// Continue processing
			}
		}

		return Result{Data: "Processed: " + input}
	}

	ctx, cancel = context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result2 := doWork(ctx, "test input")
	if result2.Err != nil {
		fmt.Println("Work failed:", result2.Err)
	} else {
		fmt.Println("Work result:", result2.Data)
	}
	// Output: Work result: Processed: test input

	fmt.Println("\n=== Context Demo Complete ===")
}
