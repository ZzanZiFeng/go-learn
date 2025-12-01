// Package main demonstrates basic RabbitMQ publish/consume operations
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

// Message represents a simple message
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// RabbitMQ client wrapper
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQ creates a new RabbitMQ client
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

// Close closes the connection
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// DeclareQueue declares a queue
func (r *RabbitMQ) DeclareQueue(name string, durable bool) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,
		durable,
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
}

// Publish publishes a message to a queue
func (r *RabbitMQ) Publish(ctx context.Context, queue string, msg *Message) error {
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg-%d", time.Now().UnixNano())
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return r.channel.PublishWithContext(ctx,
		"",    // exchange (default)
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			MessageId:    msg.ID,
			Timestamp:    msg.Timestamp,
		},
	)
}

// Consume consumes messages from a queue
func (r *RabbitMQ) Consume(ctx context.Context, queue string, handler func(*Message) error) error {
	// Set QoS
	if err := r.channel.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := r.channel.Consume(
		queue,
		"",    // consumer tag
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-msgs:
			if !ok {
				return nil
			}

			var msg Message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				log.Printf("Invalid message format: %v", err)
				delivery.Reject(false) // Don't requeue invalid messages
				continue
			}

			if err := handler(&msg); err != nil {
				log.Printf("Handler error: %v", err)
				delivery.Nack(false, true) // Requeue on error
			} else {
				delivery.Ack(false)
			}
		}
	}
}

func main() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	mq, err := NewRabbitMQ(url)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer mq.Close()

	log.Println("Connected to RabbitMQ")

	// Declare queue
	queueName := "demo-queue"
	_, err = mq.DeclareQueue(queueName, true)
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
	}
	log.Printf("Queue '%s' declared", queueName)

	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Consumer started, waiting for messages...")

		err := mq.Consume(ctx, queueName, func(msg *Message) error {
			log.Printf("Received: ID=%s, Content=%s, Time=%s",
				msg.ID, msg.Content, msg.Timestamp.Format(time.RFC3339))
			return nil
		})

		if err != nil && err != context.Canceled {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Start producer
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 1; i <= 10; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				msg := &Message{
					Content: fmt.Sprintf("Hello RabbitMQ #%d", i),
				}

				if err := mq.Publish(ctx, queueName, msg); err != nil {
					log.Printf("Failed to publish: %v", err)
				} else {
					log.Printf("Published: %s", msg.Content)
				}

				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	// Wait for shutdown signal
	<-sigCh
	log.Println("Shutting down...")
	cancel()

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Shutdown complete")
	case <-time.After(5 * time.Second):
		log.Println("Shutdown timeout")
	}
}

/*
Expected Output:

Connected to RabbitMQ
Queue 'demo-queue' declared
Consumer started, waiting for messages...
Published: Hello RabbitMQ #1
Received: ID=msg-1234567890, Content=Hello RabbitMQ #1, Time=2024-01-01T10:00:00Z
Published: Hello RabbitMQ #2
Received: ID=msg-1234567891, Content=Hello RabbitMQ #2, Time=2024-01-01T10:00:00Z
...
Published: Hello RabbitMQ #10
Received: ID=msg-1234567899, Content=Hello RabbitMQ #10, Time=2024-01-01T10:00:05Z
^C
Shutting down...
Shutdown complete

To Run:
1. Start RabbitMQ: docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management
2. Run: go run main.go
*/
