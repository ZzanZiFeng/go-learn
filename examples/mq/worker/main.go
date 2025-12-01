// Package main demonstrates work queue pattern with multiple workers
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Task represents a work task
type Task struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Payload     string    `json:"payload"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	RetryCount  int       `json:"retry_count"`
	MaxRetries  int       `json:"max_retries"`
	ProcessTime int       `json:"process_time"` // Simulated processing time in ms
}

// TaskResult represents the result of processing a task
type TaskResult struct {
	TaskID    string    `json:"task_id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Duration  int64     `json:"duration_ms"`
	ProcessAt time.Time `json:"processed_at"`
	WorkerID  string    `json:"worker_id"`
}

// WorkQueue manages the work queue
type WorkQueue struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	queueName  string
	prefetch   int
	maxRetries int
}

// WorkQueueConfig configuration
type WorkQueueConfig struct {
	URL        string
	QueueName  string
	Prefetch   int
	MaxRetries int
}

// NewWorkQueue creates a new work queue
func NewWorkQueue(cfg WorkQueueConfig) (*WorkQueue, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare the task queue with DLQ support
	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": cfg.QueueName + ".dlq",
	}

	_, err = ch.QueueDeclare(
		cfg.QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		args,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Declare DLQ
	_, err = ch.QueueDeclare(
		cfg.QueueName+".dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Set QoS for fair dispatch
	prefetch := cfg.Prefetch
	if prefetch == 0 {
		prefetch = 1
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &WorkQueue{
		conn:       conn,
		channel:    ch,
		queueName:  cfg.QueueName,
		prefetch:   prefetch,
		maxRetries: maxRetries,
	}, nil
}

// Enqueue adds a task to the queue
func (wq *WorkQueue) Enqueue(ctx context.Context, task *Task) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = wq.maxRetries
	}

	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return wq.channel.PublishWithContext(ctx,
		"",
		wq.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			MessageId:    task.ID,
			Timestamp:    task.CreatedAt,
			Priority:     uint8(task.Priority),
		},
	)
}

// TaskHandler is a function that processes a task
type TaskHandler func(task *Task) error

// Worker processes tasks from the queue
type Worker struct {
	id       string
	wq       *WorkQueue
	handler  TaskHandler
	results  chan<- TaskResult
	stopCh   chan struct{}
	stoppedCh chan struct{}
}

// NewWorker creates a new worker
func NewWorker(id string, wq *WorkQueue, handler TaskHandler, results chan<- TaskResult) *Worker {
	return &Worker{
		id:        id,
		wq:        wq,
		handler:   handler,
		results:   results,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) error {
	msgs, err := w.wq.channel.Consume(
		w.wq.queueName,
		w.id, // consumer tag
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[Worker-%s] Started, prefetch=%d", w.id, w.wq.prefetch)

	go func() {
		defer close(w.stoppedCh)

		for {
			select {
			case <-ctx.Done():
				log.Printf("[Worker-%s] Context cancelled", w.id)
				return
			case <-w.stopCh:
				log.Printf("[Worker-%s] Stop signal received", w.id)
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Printf("[Worker-%s] Channel closed", w.id)
					return
				}

				w.processMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (w *Worker) processMessage(ctx context.Context, msg amqp.Delivery) {
	start := time.Now()

	var task Task
	if err := json.Unmarshal(msg.Body, &task); err != nil {
		log.Printf("[Worker-%s] Invalid task format: %v", w.id, err)
		msg.Reject(false)
		return
	}

	log.Printf("[Worker-%s] Processing task: %s (type=%s, retry=%d/%d)",
		w.id, task.ID, task.Type, task.RetryCount, task.MaxRetries)

	err := w.handler(&task)
	duration := time.Since(start).Milliseconds()

	result := TaskResult{
		TaskID:    task.ID,
		Duration:  duration,
		ProcessAt: time.Now(),
		WorkerID:  w.id,
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()

		log.Printf("[Worker-%s] Task %s failed: %v (duration=%dms)",
			w.id, task.ID, err, duration)

		if task.RetryCount < task.MaxRetries {
			// Retry with incremented count
			task.RetryCount++
			w.wq.Enqueue(ctx, &task)
			msg.Ack(false)
		} else {
			// Max retries exceeded, send to DLQ
			log.Printf("[Worker-%s] Task %s exceeded max retries, sending to DLQ",
				w.id, task.ID)
			msg.Reject(false)
		}
	} else {
		result.Success = true
		log.Printf("[Worker-%s] Task %s completed (duration=%dms)",
			w.id, task.ID, duration)
		msg.Ack(false)
	}

	// Send result (non-blocking)
	select {
	case w.results <- result:
	default:
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	close(w.stopCh)
	<-w.stoppedCh
}

// Close closes the work queue
func (wq *WorkQueue) Close() {
	if wq.channel != nil {
		wq.channel.Close()
	}
	if wq.conn != nil {
		wq.conn.Close()
	}
}

func main() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	// Create work queue
	wq, err := NewWorkQueue(WorkQueueConfig{
		URL:        url,
		QueueName:  "tasks",
		Prefetch:   1, // Fair dispatch
		MaxRetries: 3,
	})
	if err != nil {
		log.Fatal("Failed to create work queue:", err)
	}
	defer wq.Close()

	log.Println("Work queue created")

	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Results channel
	results := make(chan TaskResult, 100)

	// Task handler
	handler := func(task *Task) error {
		// Simulate processing time
		processTime := task.ProcessTime
		if processTime == 0 {
			processTime = 100 + rand.Intn(400) // 100-500ms
		}
		time.Sleep(time.Duration(processTime) * time.Millisecond)

		// Simulate random failure (20% chance)
		if rand.Float32() < 0.2 {
			return fmt.Errorf("simulated failure")
		}

		return nil
	}

	// Start workers
	numWorkers := 3
	workers := make([]*Worker, numWorkers)

	for i := 0; i < numWorkers; i++ {
		workerID := fmt.Sprintf("%d", i+1)
		workers[i] = NewWorker(workerID, wq, handler, results)
		if err := workers[i].Start(ctx); err != nil {
			log.Fatalf("Failed to start worker %s: %v", workerID, err)
		}
	}

	// Results collector
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var total, success, failed int
		var totalDuration int64

		for {
			select {
			case <-ctx.Done():
				log.Printf("Results: total=%d, success=%d, failed=%d, avg_duration=%dms",
					total, success, failed, totalDuration/int64(max(total, 1)))
				return
			case result := <-results:
				total++
				totalDuration += result.Duration
				if result.Success {
					success++
				} else {
					failed++
				}
			}
		}
	}()

	// Producer: enqueue tasks
	wg.Add(1)
	go func() {
		defer wg.Done()

		taskTypes := []string{"email", "report", "notification", "backup"}

		for i := 1; i <= 20; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				task := &Task{
					Type:        taskTypes[rand.Intn(len(taskTypes))],
					Payload:     fmt.Sprintf("Task payload #%d", i),
					Priority:    rand.Intn(10),
					ProcessTime: 100 + rand.Intn(400),
				}

				if err := wq.Enqueue(ctx, task); err != nil {
					log.Printf("Failed to enqueue: %v", err)
				} else {
					log.Printf("Enqueued: %s (type=%s, priority=%d)",
						task.ID, task.Type, task.Priority)
				}

				time.Sleep(100 * time.Millisecond)
			}
		}

		log.Println("All tasks enqueued")
	}()

	// Wait for shutdown
	<-sigCh
	log.Println("Shutting down...")
	cancel()

	// Stop workers
	for _, w := range workers {
		w.Stop()
	}

	wg.Wait()
	log.Println("Shutdown complete")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

/*
Expected Output:

Work queue created
[Worker-1] Started, prefetch=1
[Worker-2] Started, prefetch=1
[Worker-3] Started, prefetch=1
Enqueued: task-1234567890 (type=email, priority=5)
Enqueued: task-1234567891 (type=report, priority=3)
[Worker-1] Processing task: task-1234567890 (type=email, retry=0/3)
[Worker-2] Processing task: task-1234567891 (type=report, retry=0/3)
Enqueued: task-1234567892 (type=notification, priority=7)
[Worker-3] Processing task: task-1234567892 (type=notification, retry=0/3)
[Worker-1] Task task-1234567890 completed (duration=234ms)
[Worker-2] Task task-1234567891 failed: simulated failure (duration=156ms)
...
All tasks enqueued
^C
Shutting down...
Results: total=20, success=16, failed=4, avg_duration=245ms
Shutdown complete

To Run:
1. Start RabbitMQ: docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management
2. Run: go run main.go
*/
