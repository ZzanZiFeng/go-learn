# 发布订阅 (Publish/Subscribe)

## 概述

发布订阅模式用于将消息广播给多个消费者，每个消费者都能收到消息的副本。

```
                              ┌──▶ Queue1 ──▶ Consumer1 (Email)
Producer ──▶ Fanout Exchange ─┼──▶ Queue2 ──▶ Consumer2 (SMS)
                              └──▶ Queue3 ──▶ Consumer3 (Push)

一条消息，多个消费者各收到一份
```

## Exchange 类型

### 1. Fanout Exchange (扇出)

广播到所有绑定的队列，忽略 routing key。

```go
// 声明 Fanout 交换机
err := ch.ExchangeDeclare(
    "logs",   // 交换机名
    "fanout", // 类型
    true,     // durable
    false,    // autoDelete
    false,    // internal
    false,    // noWait
    nil,      // args
)
```

### 2. Direct Exchange (直连)

精确匹配 routing key。

```go
err := ch.ExchangeDeclare(
    "logs_direct",
    "direct",
    true,
    false,
    false,
    false,
    nil,
)

// 绑定时指定 routing key
ch.QueueBind("error_logs", "error", "logs_direct", false, nil)
ch.QueueBind("all_logs", "info", "logs_direct", false, nil)
ch.QueueBind("all_logs", "warning", "logs_direct", false, nil)
ch.QueueBind("all_logs", "error", "logs_direct", false, nil)
```

### 3. Topic Exchange (主题)

模式匹配 routing key，支持通配符。

```go
err := ch.ExchangeDeclare(
    "logs_topic",
    "topic",
    true,
    false,
    false,
    false,
    nil,
)

// 绑定模式
// * 匹配一个单词
// # 匹配零个或多个单词
ch.QueueBind("queue1", "*.error", "logs_topic", false, nil)     // user.error, order.error
ch.QueueBind("queue2", "order.*", "logs_topic", false, nil)     // order.created, order.paid
ch.QueueBind("queue3", "#.critical", "logs_topic", false, nil)  // system.critical, app.db.critical
ch.QueueBind("queue4", "#", "logs_topic", false, nil)           // 所有消息
```

### 4. Headers Exchange (头)

基于消息头属性匹配，不使用 routing key。

```go
err := ch.ExchangeDeclare(
    "logs_headers",
    "headers",
    true,
    false,
    false,
    false,
    nil,
)

// 绑定时指定匹配条件
ch.QueueBind("queue1", "", "logs_headers", false, amqp.Table{
    "x-match": "all", // all=全部匹配, any=任一匹配
    "format":  "pdf",
    "type":    "report",
})
```

## Fanout 示例：日志广播

### 发布者

```go
package main

import (
    "context"
    "log"
    "os"
    "strings"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    // 声明 Fanout 交换机
    err := ch.ExchangeDeclare(
        "logs",   // name
        "fanout", // type
        true,     // durable
        false,    // autoDelete
        false,    // internal
        false,    // noWait
        nil,      // args
    )
    if err != nil {
        log.Fatal(err)
    }

    body := bodyFrom(os.Args)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 发布消息到交换机（routing key 被忽略）
    err = ch.PublishWithContext(ctx,
        "logs", // exchange
        "",     // routing key（Fanout 忽略）
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte(body),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Printf(" [x] Sent %s", body)
}

func bodyFrom(args []string) string {
    if len(args) < 2 {
        return "hello"
    }
    return strings.Join(args[1:], " ")
}
```

### 订阅者

```go
package main

import (
    "log"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    // 声明交换机
    ch.ExchangeDeclare("logs", "fanout", true, false, false, false, nil)

    // 声明临时队列（自动生成名称，断开连接后自动删除）
    q, err := ch.QueueDeclare(
        "",    // 空名称，自动生成
        false, // durable
        false, // deleteWhenUnused
        true,  // exclusive（独占，断开连接自动删除）
        false, // noWait
        nil,   // args
    )
    if err != nil {
        log.Fatal(err)
    }

    // 绑定队列到交换机
    err = ch.QueueBind(
        q.Name, // queue name
        "",     // routing key（Fanout 忽略）
        "logs", // exchange
        false,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

    log.Printf(" [*] Waiting for logs. Queue: %s", q.Name)

    for msg := range msgs {
        log.Printf(" [x] %s", msg.Body)
    }
}
```

## Direct 示例：日志分级

### 发布者

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    // 声明 Direct 交换机
    ch.ExchangeDeclare("logs_direct", "direct", true, false, false, false, nil)

    severity := severityFrom(os.Args)
    body := bodyFrom(os.Args)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 发布消息，使用 severity 作为 routing key
    ch.PublishWithContext(ctx,
        "logs_direct", // exchange
        severity,      // routing key
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte(body),
        },
    )

    log.Printf(" [x] Sent [%s] %s", severity, body)
}

func severityFrom(args []string) string {
    if len(args) < 2 {
        return "info"
    }
    return args[1]
}

func bodyFrom(args []string) string {
    if len(args) < 3 {
        return "Hello"
    }
    return args[2]
}
```

### 订阅者

```go
package main

import (
    "log"
    "os"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    ch.ExchangeDeclare("logs_direct", "direct", true, false, false, false, nil)

    q, _ := ch.QueueDeclare("", false, false, true, false, nil)

    // 绑定指定的 severity
    severities := os.Args[1:]
    if len(severities) == 0 {
        log.Fatal("Usage: go run subscriber.go [info] [warning] [error]")
    }

    for _, severity := range severities {
        ch.QueueBind(q.Name, severity, "logs_direct", false, nil)
        log.Printf(" [*] Binding queue to exchange with key: %s", severity)
    }

    msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

    log.Printf(" [*] Waiting for logs")

    for msg := range msgs {
        log.Printf(" [x] [%s] %s", msg.RoutingKey, msg.Body)
    }
}
```

### 运行示例

```bash
# 订阅所有日志
go run subscriber.go info warning error

# 只订阅错误日志
go run subscriber.go error

# 发送不同级别的日志
go run publisher.go info "Info message"
go run publisher.go warning "Warning message"
go run publisher.go error "Error message"
```

## Topic 示例：复杂路由

### 发布者

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    // 声明 Topic 交换机
    ch.ExchangeDeclare("logs_topic", "topic", true, false, false, false, nil)

    routingKey := routingKeyFrom(os.Args)
    body := bodyFrom(os.Args)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    ch.PublishWithContext(ctx,
        "logs_topic",
        routingKey,
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte(body),
        },
    )

    log.Printf(" [x] Sent [%s] %s", routingKey, body)
}

func routingKeyFrom(args []string) string {
    if len(args) < 2 {
        return "anonymous.info"
    }
    return args[1]
}

func bodyFrom(args []string) string {
    if len(args) < 3 {
        return "Hello"
    }
    return args[2]
}
```

### 订阅者

```go
package main

import (
    "log"
    "os"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    ch.ExchangeDeclare("logs_topic", "topic", true, false, false, false, nil)

    q, _ := ch.QueueDeclare("", false, false, true, false, nil)

    // 绑定 routing key 模式
    patterns := os.Args[1:]
    if len(patterns) == 0 {
        log.Fatal("Usage: go run subscriber.go <pattern> [pattern]...")
    }

    for _, pattern := range patterns {
        ch.QueueBind(q.Name, pattern, "logs_topic", false, nil)
        log.Printf(" [*] Binding pattern: %s", pattern)
    }

    msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

    log.Printf(" [*] Waiting for logs")

    for msg := range msgs {
        log.Printf(" [x] [%s] %s", msg.RoutingKey, msg.Body)
    }
}
```

### 运行示例

```bash
# 订阅所有日志
go run subscriber.go "#"

# 订阅所有 error
go run subscriber.go "*.error"

# 订阅 order 相关
go run subscriber.go "order.*"

# 发送消息
go run publisher.go "user.info" "User logged in"
go run publisher.go "user.error" "User login failed"
go run publisher.go "order.created" "New order 123"
go run publisher.go "order.error" "Order failed"
```

## 完整发布订阅封装

```go
package pubsub

import (
    "context"
    "encoding/json"
    "log"
    "sync"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Event 事件
type Event struct {
    Type      string      `json:"type"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

// Handler 事件处理函数
type Handler func(event *Event) error

// PubSub 发布订阅
type PubSub struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    exchange string
    kind     string // fanout, direct, topic
}

// Config 配置
type Config struct {
    URL      string
    Exchange string
    Kind     string // fanout, direct, topic
}

// New 创建 PubSub
func New(cfg Config) (*PubSub, error) {
    conn, err := amqp.Dial(cfg.URL)
    if err != nil {
        return nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }

    // 声明交换机
    err = ch.ExchangeDeclare(
        cfg.Exchange,
        cfg.Kind,
        true,  // durable
        false, // autoDelete
        false, // internal
        false, // noWait
        nil,
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }

    return &PubSub{
        conn:     conn,
        channel:  ch,
        exchange: cfg.Exchange,
        kind:     cfg.Kind,
    }, nil
}

// Publish 发布事件
func (ps *PubSub) Publish(ctx context.Context, routingKey string, event *Event) error {
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now()
    }

    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    return ps.channel.PublishWithContext(ctx,
        ps.exchange,
        routingKey,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         data,
            Timestamp:    event.Timestamp,
            Headers: amqp.Table{
                "x-event-type": event.Type,
            },
        },
    )
}

// Subscribe 订阅事件
func (ps *PubSub) Subscribe(ctx context.Context, name string, patterns []string, handler Handler) error {
    // 声明队列
    q, err := ps.channel.QueueDeclare(
        name,  // 队列名（空字符串自动生成）
        true,  // durable
        false, // autoDelete
        false, // exclusive
        false, // noWait
        nil,
    )
    if err != nil {
        return err
    }

    // 绑定模式
    for _, pattern := range patterns {
        err = ps.channel.QueueBind(q.Name, pattern, ps.exchange, false, nil)
        if err != nil {
            return err
        }
        log.Printf("Subscribed to: %s", pattern)
    }

    // 消费消息
    msgs, err := ps.channel.Consume(q.Name, "", false, false, false, false, nil)
    if err != nil {
        return err
    }

    log.Printf("Subscriber started: %s", name)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg, ok := <-msgs:
            if !ok {
                return nil
            }

            var event Event
            if err := json.Unmarshal(msg.Body, &event); err != nil {
                log.Printf("Invalid event: %v", err)
                msg.Reject(false)
                continue
            }

            if err := handler(&event); err != nil {
                log.Printf("Handler error: %v", err)
                msg.Nack(false, true) // 重新入队
            } else {
                msg.Ack(false)
            }
        }
    }
}

// Close 关闭连接
func (ps *PubSub) Close() {
    if ps.channel != nil {
        ps.channel.Close()
    }
    if ps.conn != nil {
        ps.conn.Close()
    }
}
```

## 事件驱动架构示例

### 事件定义

```go
package events

const (
    // 用户事件
    UserRegistered = "user.registered"
    UserUpdated    = "user.updated"
    UserDeleted    = "user.deleted"

    // 订单事件
    OrderCreated   = "order.created"
    OrderPaid      = "order.paid"
    OrderShipped   = "order.shipped"
    OrderCompleted = "order.completed"
    OrderCancelled = "order.cancelled"

    // 支付事件
    PaymentSuccess = "payment.success"
    PaymentFailed  = "payment.failed"
)

// UserRegisteredPayload 用户注册事件载荷
type UserRegisteredPayload struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

// OrderCreatedPayload 订单创建事件载荷
type OrderCreatedPayload struct {
    OrderID   string  `json:"order_id"`
    UserID    uint    `json:"user_id"`
    Amount    float64 `json:"amount"`
    ProductID string  `json:"product_id"`
}
```

### 事件发布者（订单服务）

```go
package main

import (
    "context"
    "log"
    "time"

    "myapp/events"
    "myapp/pubsub"
)

func main() {
    // 创建发布者
    ps, err := pubsub.New(pubsub.Config{
        URL:      "amqp://guest:guest@localhost:5672/",
        Exchange: "events",
        Kind:     "topic",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer ps.Close()

    ctx := context.Background()

    // 发布订单创建事件
    ps.Publish(ctx, events.OrderCreated, &pubsub.Event{
        Type: events.OrderCreated,
        Payload: events.OrderCreatedPayload{
            OrderID:   "order-123",
            UserID:    1,
            Amount:    99.99,
            ProductID: "prod-456",
        },
    })

    log.Println("Order created event published")

    // 模拟支付完成
    time.Sleep(time.Second)

    ps.Publish(ctx, events.OrderPaid, &pubsub.Event{
        Type: events.OrderPaid,
        Payload: map[string]interface{}{
            "order_id": "order-123",
            "user_id":  1,
        },
    })

    log.Println("Order paid event published")
}
```

### 事件订阅者（通知服务）

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"

    "myapp/events"
    "myapp/pubsub"
)

func main() {
    ps, err := pubsub.New(pubsub.Config{
        URL:      "amqp://guest:guest@localhost:5672/",
        Exchange: "events",
        Kind:     "topic",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer ps.Close()

    ctx, cancel := context.WithCancel(context.Background())

    // 优雅关闭
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    // 订阅用户和订单事件
    go ps.Subscribe(ctx, "notification-service", []string{
        "user.*",   // 所有用户事件
        "order.*",  // 所有订单事件
    }, func(event *pubsub.Event) error {
        return handleNotification(event)
    })

    <-ctx.Done()
    log.Println("Notification service stopped")
}

func handleNotification(event *pubsub.Event) error {
    log.Printf("Received event: %s", event.Type)

    switch event.Type {
    case events.UserRegistered:
        var payload events.UserRegisteredPayload
        data, _ := json.Marshal(event.Payload)
        json.Unmarshal(data, &payload)
        return sendWelcomeEmail(payload.Email)

    case events.OrderCreated:
        var payload events.OrderCreatedPayload
        data, _ := json.Marshal(event.Payload)
        json.Unmarshal(data, &payload)
        return sendOrderConfirmation(payload.OrderID, payload.UserID)

    case events.OrderPaid:
        return sendPaymentConfirmation(event.Payload)

    case events.OrderShipped:
        return sendShippingNotification(event.Payload)
    }

    return nil
}

func sendWelcomeEmail(email string) error {
    log.Printf("Sending welcome email to: %s", email)
    return nil
}

func sendOrderConfirmation(orderID string, userID uint) error {
    log.Printf("Sending order confirmation: %s to user: %d", orderID, userID)
    return nil
}

func sendPaymentConfirmation(payload interface{}) error {
    log.Printf("Sending payment confirmation: %v", payload)
    return nil
}

func sendShippingNotification(payload interface{}) error {
    log.Printf("Sending shipping notification: %v", payload)
    return nil
}
```

### 事件订阅者（库存服务）

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"

    "myapp/events"
    "myapp/pubsub"
)

func main() {
    ps, err := pubsub.New(pubsub.Config{
        URL:      "amqp://guest:guest@localhost:5672/",
        Exchange: "events",
        Kind:     "topic",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer ps.Close()

    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    // 只订阅订单事件
    go ps.Subscribe(ctx, "inventory-service", []string{
        "order.created",
        "order.cancelled",
    }, func(event *pubsub.Event) error {
        return handleInventory(event)
    })

    <-ctx.Done()
    log.Println("Inventory service stopped")
}

func handleInventory(event *pubsub.Event) error {
    log.Printf("Received event: %s", event.Type)

    switch event.Type {
    case events.OrderCreated:
        var payload events.OrderCreatedPayload
        data, _ := json.Marshal(event.Payload)
        json.Unmarshal(data, &payload)
        return deductInventory(payload.ProductID, 1)

    case events.OrderCancelled:
        // 恢复库存
        return restoreInventory(event.Payload)
    }

    return nil
}

func deductInventory(productID string, quantity int) error {
    log.Printf("Deducting inventory: %s x %d", productID, quantity)
    return nil
}

func restoreInventory(payload interface{}) error {
    log.Printf("Restoring inventory: %v", payload)
    return nil
}
```

## 与 Node.js 对比

### Node.js (EventEmitter + Redis Pub/Sub)

```javascript
import Redis from 'ioredis';

const pub = new Redis();
const sub = new Redis();

// 发布
await pub.publish('user:registered', JSON.stringify({
    userId: 1,
    email: 'user@example.com',
}));

// 订阅
sub.subscribe('user:*');
sub.on('message', (channel, message) => {
    console.log(`Received: ${channel}`, JSON.parse(message));
});
```

### Go (RabbitMQ)

```go
// 发布
ps.Publish(ctx, "user.registered", &pubsub.Event{
    Type: "user.registered",
    Payload: map[string]interface{}{
        "user_id": 1,
        "email":   "user@example.com",
    },
})

// 订阅
ps.Subscribe(ctx, "notification-service", []string{"user.*"}, func(event *pubsub.Event) error {
    log.Printf("Received: %s, %v", event.Type, event.Payload)
    return nil
})
```

## 最佳实践

### 1. 事件命名规范

```go
// 推荐：领域.动作
"user.registered"
"order.created"
"payment.completed"

// 不推荐
"userRegistered"
"createOrder"
"PAYMENT_DONE"
```

### 2. 事件版本控制

```go
type Event struct {
    Version   string      `json:"version"`
    Type      string      `json:"type"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

// 发布时指定版本
ps.Publish(ctx, "user.registered", &pubsub.Event{
    Version: "v1",
    Type:    "user.registered",
    Payload: payload,
})

// 消费时检查版本
func handleEvent(event *pubsub.Event) error {
    switch event.Version {
    case "v1":
        return handleV1(event)
    case "v2":
        return handleV2(event)
    default:
        return fmt.Errorf("unsupported version: %s", event.Version)
    }
}
```

### 3. 幂等消费

```go
func handleEvent(event *pubsub.Event) error {
    eventID := generateEventID(event)

    // 检查是否已处理
    exists, _ := redis.SetNX(ctx, "event:"+eventID, "1", 24*time.Hour).Result()
    if !exists {
        log.Printf("Event already processed: %s", eventID)
        return nil
    }

    // 处理事件
    return processEvent(event)
}

func generateEventID(event *pubsub.Event) string {
    data, _ := json.Marshal(event)
    hash := md5.Sum(data)
    return hex.EncodeToString(hash[:])
}
```

### 4. 错误处理与重试

```go
func (ps *PubSub) SubscribeWithRetry(ctx context.Context, name string, patterns []string, handler Handler, maxRetries int) error {
    return ps.Subscribe(ctx, name, patterns, func(event *pubsub.Event) error {
        var lastErr error
        for i := 0; i <= maxRetries; i++ {
            if err := handler(event); err != nil {
                lastErr = err
                log.Printf("Retry %d/%d for event %s: %v", i, maxRetries, event.Type, err)
                time.Sleep(time.Duration(i*i) * time.Second) // 指数退避
                continue
            }
            return nil
        }
        return lastErr
    })
}
```

**下一节**：[延迟队列](./06-delayed-queue.md) - 学习延迟消息实现
