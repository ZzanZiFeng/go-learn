// Package main demonstrates delayed queue implementation using TTL + DLX
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DelayedTask represents a delayed task
type DelayedTask struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Payload     interface{} `json:"payload"`
	DelayMs     int64       `json:"delay_ms"`
	ScheduledAt time.Time   `json:"scheduled_at"`
	CreatedAt   time.Time   `json:"created_at"`
}

// DelayQueue manages delayed message queue
type DelayQueue struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	dlx        string // Dead letter exchange
	workQueue  string // Work queue name
	delayLevels []int64 // Predefined delay levels in ms
}

// DelayQueueConfig configuration
type DelayQueueConfig struct {
	URL         string
	DLX         string
	WorkQueue   string
	DelayLevels []int64 // e.g., []int64{5000, 30000, 60000} for 5s, 30s, 1min
}

// NewDelayQueue creates a new delay queue
func NewDelayQueue(cfg DelayQueueConfig) (*DelayQueue, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	dq := &DelayQueue{
		conn:        conn,
		channel:     ch,
		dlx:         cfg.DLX,
		workQueue:   cfg.WorkQueue,
		delayLevels: cfg.DelayLevels,
	}

	if err := dq.setup(); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return dq, nil
}

func (dq *DelayQueue) setup() error {
	// 1. Declare dead letter exchange
	if err := dq.channel.ExchangeDeclare(
		dq.dlx,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	// 2. Declare work queue
	if _, err := dq.channel.QueueDeclare(
		dq.workQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare work queue: %w", err)
	}

	// 3. Bind work queue to DLX
	if err := dq.channel.QueueBind(
		dq.workQueue,
		"delayed",
		dq.dlx,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind work queue: %w", err)
	}

	// 4. Create delay queues for each level
	for _, delayMs := range dq.delayLevels {
		queueName := fmt.Sprintf("delay.%dms", delayMs)

		if _, err := dq.channel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			amqp.Table{
				"x-message-ttl":             delayMs,
				"x-dead-letter-exchange":    dq.dlx,
				"x-dead-letter-routing-key": "delayed",
			},
		); err != nil {
			return fmt.Errorf("failed to declare delay queue %s: %w", queueName, err)
		}

		log.Printf("Delay queue created: %s (TTL=%dms)", queueName, delayMs)
	}

	return nil
}

// Schedule schedules a task with a delay
func (dq *DelayQueue) Schedule(ctx context.Context, task *DelayedTask) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	task.CreatedAt = time.Now()
	task.ScheduledAt = task.CreatedAt.Add(time.Duration(task.DelayMs) * time.Millisecond)

	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// Find the appropriate delay queue
	queueName := dq.findDelayQueue(task.DelayMs)

	log.Printf("Scheduling task %s to queue %s (delay=%dms, scheduled=%s)",
		task.ID, queueName, task.DelayMs, task.ScheduledAt.Format(time.RFC3339))

	return dq.channel.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			MessageId:    task.ID,
			Timestamp:    task.CreatedAt,
		},
	)
}

// ScheduleAt schedules a task at a specific time
func (dq *DelayQueue) ScheduleAt(ctx context.Context, task *DelayedTask, executeAt time.Time) error {
	delay := time.Until(executeAt)
	if delay < 0 {
		delay = 0
	}
	task.DelayMs = delay.Milliseconds()
	return dq.Schedule(ctx, task)
}

// findDelayQueue finds the best matching delay queue
func (dq *DelayQueue) findDelayQueue(delayMs int64) string {
	// Find the smallest delay level >= requested delay
	var bestMatch int64 = -1

	for _, level := range dq.delayLevels {
		if level >= delayMs {
			if bestMatch == -1 || level < bestMatch {
				bestMatch = level
			}
		}
	}

	// If no suitable level found, use the largest one
	if bestMatch == -1 {
		bestMatch = dq.delayLevels[len(dq.delayLevels)-1]
	}

	return fmt.Sprintf("delay.%dms", bestMatch)
}

// TaskHandler processes delayed tasks
type TaskHandler func(task *DelayedTask) error

// Consume starts consuming delayed tasks
func (dq *DelayQueue) Consume(ctx context.Context, handler TaskHandler) error {
	// Set QoS
	if err := dq.channel.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := dq.channel.Consume(
		dq.workQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("Consumer started, listening on queue: %s", dq.workQueue)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			var task DelayedTask
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				log.Printf("Invalid task format: %v", err)
				msg.Reject(false)
				continue
			}

			// Calculate actual delay
			actualDelay := time.Since(task.CreatedAt)
			expectedDelay := time.Duration(task.DelayMs) * time.Millisecond
			diff := actualDelay - expectedDelay

			log.Printf("Processing task %s (type=%s, expected=%v, actual=%v, diff=%v)",
				task.ID, task.Type, expectedDelay, actualDelay, diff)

			if err := handler(&task); err != nil {
				log.Printf("Task %s failed: %v", task.ID, err)
				msg.Nack(false, true)
			} else {
				log.Printf("Task %s completed", task.ID)
				msg.Ack(false)
			}
		}
	}
}

// Close closes the connection
func (dq *DelayQueue) Close() {
	if dq.channel != nil {
		dq.channel.Close()
	}
	if dq.conn != nil {
		dq.conn.Close()
	}
}

// OrderTimeoutChecker demonstrates order timeout checking
type OrderTimeoutChecker struct {
	dq *DelayQueue
}

func NewOrderTimeoutChecker(dq *DelayQueue) *OrderTimeoutChecker {
	return &OrderTimeoutChecker{dq: dq}
}

func (c *OrderTimeoutChecker) ScheduleTimeoutCheck(ctx context.Context, orderID string, timeout time.Duration) error {
	return c.dq.Schedule(ctx, &DelayedTask{
		ID:      "timeout-" + orderID,
		Type:    "order.timeout.check",
		DelayMs: timeout.Milliseconds(),
		Payload: map[string]string{
			"order_id": orderID,
		},
	})
}

func main() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	// Create delay queue with predefined delay levels
	dq, err := NewDelayQueue(DelayQueueConfig{
		URL:       url,
		DLX:       "dlx.delayed",
		WorkQueue: "work.delayed",
		DelayLevels: []int64{
			5000,   // 5 seconds
			10000,  // 10 seconds
			30000,  // 30 seconds
			60000,  // 1 minute
			300000, // 5 minutes
		},
	})
	if err != nil {
		log.Fatal("Failed to create delay queue:", err)
	}
	defer dq.Close()

	log.Println("Delay queue system initialized")

	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start consumer
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := dq.Consume(ctx, func(task *DelayedTask) error {
			switch task.Type {
			case "order.timeout.check":
				payload := task.Payload.(map[string]interface{})
				orderID := payload["order_id"].(string)
				log.Printf("Checking order timeout: %s", orderID)
				// In real app: check if order is paid, cancel if not
				return nil

			case "reminder":
				log.Printf("Sending reminder: %v", task.Payload)
				return nil

			case "scheduled.task":
				log.Printf("Executing scheduled task: %v", task.Payload)
				return nil

			default:
				log.Printf("Unknown task type: %s", task.Type)
				return nil
			}
		})

		if err != nil && err != context.Canceled {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Demo: Schedule various tasks
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Wait for consumer to start
		time.Sleep(time.Second)

		log.Println("Scheduling demo tasks...")

		// 1. Order timeout check (5 seconds)
		orderChecker := NewOrderTimeoutChecker(dq)
		orderChecker.ScheduleTimeoutCheck(ctx, "order-001", 5*time.Second)
		log.Println("Scheduled: Order timeout check in 5s")

		// 2. Reminder (10 seconds)
		dq.Schedule(ctx, &DelayedTask{
			Type:    "reminder",
			DelayMs: 10000,
			Payload: map[string]string{
				"user_id": "user-123",
				"message": "Don't forget to complete your profile!",
			},
		})
		log.Println("Scheduled: Reminder in 10s")

		// 3. Scheduled task (15 seconds)
		dq.Schedule(ctx, &DelayedTask{
			Type:    "scheduled.task",
			DelayMs: 15000,
			Payload: map[string]string{
				"action": "generate_report",
			},
		})
		log.Println("Scheduled: Generate report in 15s")

		// 4. Schedule at specific time
		executeAt := time.Now().Add(20 * time.Second)
		dq.ScheduleAt(ctx, &DelayedTask{
			Type: "scheduled.task",
			Payload: map[string]string{
				"action": "send_newsletter",
			},
		}, executeAt)
		log.Printf("Scheduled: Newsletter at %s", executeAt.Format(time.RFC3339))

		log.Println("All demo tasks scheduled. Waiting for execution...")
	}()

	// Wait for shutdown
	<-sigCh
	log.Println("Shutting down...")
	cancel()
	wg.Wait()
	log.Println("Shutdown complete")
}

/*
Expected Output:

Delay queue created: delay.5000ms (TTL=5000ms)
Delay queue created: delay.10000ms (TTL=10000ms)
Delay queue created: delay.30000ms (TTL=30000ms)
Delay queue created: delay.60000ms (TTL=60000ms)
Delay queue created: delay.300000ms (TTL=300000ms)
Delay queue system initialized
Consumer started, listening on queue: work.delayed
Scheduling demo tasks...
Scheduling task timeout-order-001 to queue delay.5000ms (delay=5000ms, scheduled=2024-01-01T10:00:05Z)
Scheduled: Order timeout check in 5s
Scheduling task task-1234567890 to queue delay.10000ms (delay=10000ms, scheduled=2024-01-01T10:00:10Z)
Scheduled: Reminder in 10s
Scheduling task task-1234567891 to queue delay.30000ms (delay=15000ms, scheduled=2024-01-01T10:00:15Z)
Scheduled: Generate report in 15s
Scheduling task task-1234567892 to queue delay.30000ms (delay=20000ms, scheduled=2024-01-01T10:00:20Z)
Scheduled: Newsletter at 2024-01-01T10:00:20Z
All demo tasks scheduled. Waiting for execution...

(after ~5 seconds)
Processing task timeout-order-001 (type=order.timeout.check, expected=5s, actual=5.012s, diff=12ms)
Checking order timeout: order-001
Task timeout-order-001 completed

(after ~10 seconds)
Processing task task-1234567890 (type=reminder, expected=10s, actual=10.008s, diff=8ms)
Sending reminder: map[message:Don't forget to complete your profile! user_id:user-123]
Task task-1234567890 completed

(after ~15 seconds)
Processing task task-1234567891 (type=scheduled.task, expected=15s, actual=30.005s, diff=15.005s)
Executing scheduled task: map[action:generate_report]
Task task-1234567891 completed

(after ~20 seconds)
Processing task task-1234567892 (type=scheduled.task, expected=20s, actual=30.003s, diff=10.003s)
Executing scheduled task: map[action:send_newsletter]
Task task-1234567892 completed

^C
Shutting down...
Shutdown complete

Note: Tasks may be slightly delayed due to TTL level rounding (e.g., 15s -> 30s queue)

To Run:
1. Start RabbitMQ: docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management
2. Run: go run main.go
*/
