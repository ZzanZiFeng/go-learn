# 死信队列 (Dead Letter Queue)

## 概述

死信队列（Dead Letter Queue, DLQ）用于存储无法被正常消费的消息，便于后续分析和处理。

```
┌──────────────────────────────────────────────────────────────────┐
│                          死信队列流程                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Producer ──▶ Normal Queue ──(消息变成死信)──▶ DLX ──▶ DLQ       │
│                    │                                             │
│                    │ 死信原因:                                    │
│                    │ 1. 消息被拒绝 (Reject/Nack + requeue=false) │
│                    │ 2. 消息过期 (TTL)                            │
│                    │ 3. 队列满了 (max-length)                     │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 死信产生条件

| 条件 | 说明 |
|------|------|
| 消息被拒绝 | `Reject(false)` 或 `Nack(false, false)` |
| 消息过期 | 消息 TTL 或队列 TTL 过期 |
| 队列溢出 | 超过 `x-max-length` 或 `x-max-length-bytes` |

## 基本配置

### 1. 创建死信队列

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

    // 1. 声明死信交换机
    err := ch.ExchangeDeclare(
        "dlx",    // 死信交换机名称
        "direct", // 类型
        true,     // durable
        false,    // autoDelete
        false,    // internal
        false,    // noWait
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // 2. 声明死信队列
    _, err = ch.QueueDeclare(
        "dlq",  // 死信队列名称
        true,   // durable
        false,  // autoDelete
        false,  // exclusive
        false,  // noWait
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // 3. 绑定死信队列到死信交换机
    err = ch.QueueBind(
        "dlq",       // 队列名
        "dead",      // routing key
        "dlx",       // 交换机名
        false,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // 4. 声明正常队列，配置死信交换机
    _, err = ch.QueueDeclare(
        "normal_queue",
        true,
        false,
        false,
        false,
        amqp.Table{
            "x-dead-letter-exchange":    "dlx",  // 死信交换机
            "x-dead-letter-routing-key": "dead", // 死信路由键
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Dead letter queue setup complete")
}
```

### 2. 带更多配置的队列

```go
// 声明带完整死信配置的队列
_, err = ch.QueueDeclare(
    "orders",
    true,
    false,
    false,
    false,
    amqp.Table{
        // 死信配置
        "x-dead-letter-exchange":    "dlx",
        "x-dead-letter-routing-key": "orders.dead",

        // 消息 TTL
        "x-message-ttl": 60000, // 60 秒

        // 队列长度限制
        "x-max-length":       10000,      // 最大消息数
        "x-max-length-bytes": 104857600,  // 最大字节数 (100MB)

        // 溢出行为
        "x-overflow": "reject-publish-dlx", // 溢出时发送到 DLX
    },
)
```

## 死信队列封装

```go
package dlq

import (
    "context"
    "encoding/json"
    "log"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// DeadLetter 死信消息
type DeadLetter struct {
    OriginalQueue   string      `json:"original_queue"`
    OriginalMessage []byte      `json:"original_message"`
    Reason          string      `json:"reason"`
    Error           string      `json:"error"`
    RetryCount      int         `json:"retry_count"`
    FirstDeath      time.Time   `json:"first_death"`
    LastDeath       time.Time   `json:"last_death"`
    Headers         amqp.Table  `json:"headers"`
}

// DLQManager 死信队列管理器
type DLQManager struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    dlx     string // 死信交换机
}

// Config 配置
type Config struct {
    URL string
    DLX string
}

// New 创建 DLQ 管理器
func New(cfg Config) (*DLQManager, error) {
    conn, err := amqp.Dial(cfg.URL)
    if err != nil {
        return nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }

    // 声明死信交换机
    err = ch.ExchangeDeclare(cfg.DLX, "direct", true, false, false, false, nil)
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }

    return &DLQManager{
        conn:    conn,
        channel: ch,
        dlx:     cfg.DLX,
    }, nil
}

// SetupQueue 为队列配置死信
func (m *DLQManager) SetupQueue(queueName string, opts ...QueueOption) error {
    options := &queueOptions{
        ttl:       0,
        maxLength: 0,
        maxBytes:  0,
    }

    for _, opt := range opts {
        opt(options)
    }

    // 死信队列名称
    dlqName := queueName + ".dlq"

    // 声明死信队列
    _, err := m.channel.QueueDeclare(dlqName, true, false, false, false, nil)
    if err != nil {
        return err
    }

    // 绑定死信队列
    err = m.channel.QueueBind(dlqName, queueName+".dead", m.dlx, false, nil)
    if err != nil {
        return err
    }

    // 构建队列参数
    args := amqp.Table{
        "x-dead-letter-exchange":    m.dlx,
        "x-dead-letter-routing-key": queueName + ".dead",
    }

    if options.ttl > 0 {
        args["x-message-ttl"] = options.ttl
    }
    if options.maxLength > 0 {
        args["x-max-length"] = options.maxLength
    }
    if options.maxBytes > 0 {
        args["x-max-length-bytes"] = options.maxBytes
    }

    // 声明正常队列
    _, err = m.channel.QueueDeclare(queueName, true, false, false, false, args)
    return err
}

type queueOptions struct {
    ttl       int64
    maxLength int64
    maxBytes  int64
}

type QueueOption func(*queueOptions)

// WithTTL 设置消息 TTL
func WithTTL(ttlMs int64) QueueOption {
    return func(o *queueOptions) {
        o.ttl = ttlMs
    }
}

// WithMaxLength 设置最大消息数
func WithMaxLength(max int64) QueueOption {
    return func(o *queueOptions) {
        o.maxLength = max
    }
}

// WithMaxBytes 设置最大字节数
func WithMaxBytes(max int64) QueueOption {
    return func(o *queueOptions) {
        o.maxBytes = max
    }
}

// ProcessDLQ 处理死信队列
func (m *DLQManager) ProcessDLQ(ctx context.Context, queueName string, handler func(*DeadLetter) error) error {
    dlqName := queueName + ".dlq"

    msgs, err := m.channel.Consume(dlqName, "", false, false, false, false, nil)
    if err != nil {
        return err
    }

    log.Printf("Processing dead letters from: %s", dlqName)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg, ok := <-msgs:
            if !ok {
                return nil
            }

            deadLetter := parseDeadLetter(msg)
            if err := handler(deadLetter); err != nil {
                log.Printf("Failed to process dead letter: %v", err)
                msg.Nack(false, true) // 重新入队
            } else {
                msg.Ack(false)
            }
        }
    }
}

func parseDeadLetter(msg amqp.Delivery) *DeadLetter {
    dl := &DeadLetter{
        OriginalMessage: msg.Body,
        Headers:         msg.Headers,
    }

    // 解析 x-death 头
    if deaths, ok := msg.Headers["x-death"].([]interface{}); ok && len(deaths) > 0 {
        if death, ok := deaths[0].(amqp.Table); ok {
            if queue, ok := death["queue"].(string); ok {
                dl.OriginalQueue = queue
            }
            if reason, ok := death["reason"].(string); ok {
                dl.Reason = reason
            }
            if t, ok := death["time"].(time.Time); ok {
                dl.FirstDeath = t
            }
            if count, ok := death["count"].(int64); ok {
                dl.RetryCount = int(count)
            }
        }

        // 最后一次死亡时间
        if len(deaths) > 0 {
            if lastDeath, ok := deaths[len(deaths)-1].(amqp.Table); ok {
                if t, ok := lastDeath["time"].(time.Time); ok {
                    dl.LastDeath = t
                }
            }
        }
    }

    return dl
}

// Retry 重试死信消息
func (m *DLQManager) Retry(ctx context.Context, dl *DeadLetter) error {
    return m.channel.PublishWithContext(ctx,
        "",
        dl.OriginalQueue,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         dl.OriginalMessage,
            Headers: amqp.Table{
                "x-retry-count": dl.RetryCount + 1,
            },
        },
    )
}

// Close 关闭连接
func (m *DLQManager) Close() {
    if m.channel != nil {
        m.channel.Close()
    }
    if m.conn != nil {
        m.conn.Close()
    }
}
```

## 消息重试机制

### 带重试次数限制

```go
package retry

import (
    "context"
    "fmt"
    "log"

    amqp "github.com/rabbitmq/amqp091-go"
)

// RetryConfig 重试配置
type RetryConfig struct {
    MaxRetries int   // 最大重试次数
    RetryDelays []int // 重试延迟（毫秒）
}

// Consumer 带重试的消费者
type Consumer struct {
    channel     *amqp.Channel
    dlqManager  *DLQManager
    retryConfig RetryConfig
}

// ConsumeWithRetry 带重试的消费
func (c *Consumer) ConsumeWithRetry(ctx context.Context, queue string, handler func([]byte) error) error {
    msgs, err := c.channel.Consume(queue, "", false, false, false, false, nil)
    if err != nil {
        return err
    }

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg, ok := <-msgs:
            if !ok {
                return nil
            }

            // 获取重试次数
            retryCount := 0
            if count, ok := msg.Headers["x-retry-count"].(int32); ok {
                retryCount = int(count)
            }

            err := handler(msg.Body)
            if err != nil {
                log.Printf("Handler error (retry %d): %v", retryCount, err)

                if retryCount < c.retryConfig.MaxRetries {
                    // 重试
                    c.retryMessage(ctx, queue, msg, retryCount)
                    msg.Ack(false)
                } else {
                    // 超过重试次数，发送到死信队列
                    log.Printf("Max retries exceeded, sending to DLQ")
                    msg.Reject(false)
                }
            } else {
                msg.Ack(false)
            }
        }
    }
}

func (c *Consumer) retryMessage(ctx context.Context, queue string, msg amqp.Delivery, retryCount int) {
    // 获取延迟时间
    delay := 1000 // 默认 1 秒
    if retryCount < len(c.retryConfig.RetryDelays) {
        delay = c.retryConfig.RetryDelays[retryCount]
    }

    // 发送到延迟队列
    delayQueue := fmt.Sprintf("%s.retry.%dms", queue, delay)

    // 确保延迟队列存在
    c.channel.QueueDeclare(
        delayQueue,
        true,
        false,
        false,
        false,
        amqp.Table{
            "x-message-ttl":             int64(delay),
            "x-dead-letter-exchange":    "",
            "x-dead-letter-routing-key": queue,
        },
    )

    // 发送消息
    c.channel.PublishWithContext(ctx,
        "",
        delayQueue,
        false,
        false,
        amqp.Publishing{
            ContentType:  msg.ContentType,
            DeliveryMode: amqp.Persistent,
            Body:         msg.Body,
            Headers: amqp.Table{
                "x-retry-count": retryCount + 1,
            },
        },
    )
}
```

### 指数退避重试

```go
package retry

import (
    "context"
    "math"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// ExponentialBackoff 指数退避配置
type ExponentialBackoff struct {
    InitialDelay time.Duration // 初始延迟
    MaxDelay     time.Duration // 最大延迟
    Multiplier   float64       // 乘数
    MaxRetries   int           // 最大重试次数
}

// DefaultBackoff 默认配置
var DefaultBackoff = ExponentialBackoff{
    InitialDelay: time.Second,
    MaxDelay:     time.Hour,
    Multiplier:   2.0,
    MaxRetries:   5,
}

// CalculateDelay 计算延迟时间
func (b *ExponentialBackoff) CalculateDelay(retryCount int) time.Duration {
    delay := float64(b.InitialDelay) * math.Pow(b.Multiplier, float64(retryCount))
    if delay > float64(b.MaxDelay) {
        delay = float64(b.MaxDelay)
    }
    return time.Duration(delay)
}

// RetryableConsumer 可重试消费者
type RetryableConsumer struct {
    channel  *amqp.Channel
    backoff  ExponentialBackoff
    dlx      string
}

// Consume 消费并自动重试
func (c *RetryableConsumer) Consume(ctx context.Context, queue string, handler func([]byte) error) error {
    msgs, err := c.channel.Consume(queue, "", false, false, false, false, nil)
    if err != nil {
        return err
    }

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg, ok := <-msgs:
            if !ok {
                return nil
            }

            retryCount := getRetryCount(msg.Headers)

            if err := handler(msg.Body); err != nil {
                if retryCount < c.backoff.MaxRetries {
                    delay := c.backoff.CalculateDelay(retryCount)
                    c.scheduleRetry(ctx, queue, msg, retryCount, delay)
                    msg.Ack(false)
                } else {
                    msg.Reject(false) // 发送到 DLQ
                }
            } else {
                msg.Ack(false)
            }
        }
    }
}

func getRetryCount(headers amqp.Table) int {
    if headers == nil {
        return 0
    }
    if count, ok := headers["x-retry-count"].(int32); ok {
        return int(count)
    }
    return 0
}

func (c *RetryableConsumer) scheduleRetry(ctx context.Context, queue string, msg amqp.Delivery, retryCount int, delay time.Duration) {
    // 创建延迟队列
    delayMs := delay.Milliseconds()
    delayQueue := fmt.Sprintf("retry.%s.%dms", queue, delayMs)

    c.channel.QueueDeclare(
        delayQueue,
        true,
        true, // autoDelete
        false,
        false,
        amqp.Table{
            "x-message-ttl":             delayMs,
            "x-dead-letter-exchange":    "",
            "x-dead-letter-routing-key": queue,
            "x-expires":                 delayMs + 60000, // 队列过期时间
        },
    )

    c.channel.PublishWithContext(ctx,
        "",
        delayQueue,
        false,
        false,
        amqp.Publishing{
            ContentType:  msg.ContentType,
            DeliveryMode: amqp.Persistent,
            Body:         msg.Body,
            Headers: amqp.Table{
                "x-retry-count":  retryCount + 1,
                "x-first-retry": msg.Headers["x-first-retry"],
            },
        },
    )
}
```

## 死信处理策略

### 1. 人工处理

```go
// 死信监控和告警
func monitorDLQ(ctx context.Context, dlqManager *DLQManager, queues []string) {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            for _, queue := range queues {
                dlqName := queue + ".dlq"
                count := getQueueMessageCount(dlqName)

                if count > 0 {
                    log.Printf("DLQ %s has %d messages", dlqName, count)

                    // 发送告警
                    if count > 100 {
                        sendAlert(fmt.Sprintf("DLQ %s has %d messages", dlqName, count))
                    }
                }
            }
        }
    }
}
```

### 2. 自动重试

```go
// 定期重试死信
func autoRetryDLQ(ctx context.Context, dlqManager *DLQManager, queue string) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            dlqManager.ProcessDLQ(ctx, queue, func(dl *DeadLetter) error {
                // 检查重试次数
                if dl.RetryCount < 10 {
                    return dlqManager.Retry(ctx, dl)
                }

                // 超过重试次数，归档
                return archiveDeadLetter(dl)
            })
        }
    }
}
```

### 3. 分类处理

```go
// 按错误类型分类处理
func classifyAndHandle(dl *DeadLetter) error {
    switch dl.Reason {
    case "rejected":
        // 业务逻辑错误，记录日志
        return logAndArchive(dl)

    case "expired":
        // 消息过期，可能需要通知
        return notifyExpired(dl)

    case "maxlen":
        // 队列溢出，可能需要扩容告警
        sendCapacityAlert(dl.OriginalQueue)
        return retryLater(dl)

    default:
        return archiveDeadLetter(dl)
    }
}
```

## 完整示例

```go
package main

import (
    "context"
    "encoding/json"
    "errors"
    "log"
    "math/rand"
    "os"
    "os/signal"
    "syscall"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Order 订单
type Order struct {
    ID     string  `json:"id"`
    Amount float64 `json:"amount"`
}

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    // 设置死信队列
    setupDLQ(ch)

    ctx, cancel := context.WithCancel(context.Background())

    // 优雅关闭
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    // 启动消费者
    go consumeOrders(ctx, ch)

    // 启动死信处理
    go processDLQ(ctx, ch)

    // 发送测试订单
    go func() {
        for i := 0; i < 10; i++ {
            publishOrder(ch, &Order{
                ID:     fmt.Sprintf("order-%d", i),
                Amount: float64(rand.Intn(1000)),
            })
            time.Sleep(time.Second)
        }
    }()

    <-ctx.Done()
    log.Println("Shutting down...")
}

func setupDLQ(ch *amqp.Channel) {
    // 死信交换机
    ch.ExchangeDeclare("dlx.orders", "direct", true, false, false, false, nil)

    // 死信队列
    ch.QueueDeclare("orders.dlq", true, false, false, false, nil)
    ch.QueueBind("orders.dlq", "orders.dead", "dlx.orders", false, nil)

    // 正常队列（带死信配置）
    ch.QueueDeclare(
        "orders",
        true,
        false,
        false,
        false,
        amqp.Table{
            "x-dead-letter-exchange":    "dlx.orders",
            "x-dead-letter-routing-key": "orders.dead",
        },
    )

    // 设置 QoS
    ch.Qos(1, 0, false)
}

func publishOrder(ch *amqp.Channel, order *Order) {
    data, _ := json.Marshal(order)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    ch.PublishWithContext(ctx,
        "",
        "orders",
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         data,
            MessageId:    order.ID,
        },
    )
    log.Printf("Published order: %s", order.ID)
}

func consumeOrders(ctx context.Context, ch *amqp.Channel) {
    msgs, _ := ch.Consume("orders", "", false, false, false, false, nil)

    for {
        select {
        case <-ctx.Done():
            return
        case msg, ok := <-msgs:
            if !ok {
                return
            }

            var order Order
            json.Unmarshal(msg.Body, &order)

            err := processOrder(&order)
            retryCount := getRetryCount(msg.Headers)

            if err != nil {
                log.Printf("Order %s failed (retry %d): %v", order.ID, retryCount, err)

                if retryCount < 3 {
                    // 重新发布带重试计数
                    requeue(ch, msg, retryCount)
                    msg.Ack(false)
                } else {
                    // 发送到死信队列
                    msg.Reject(false)
                }
            } else {
                log.Printf("Order %s processed successfully", order.ID)
                msg.Ack(false)
            }
        }
    }
}

func processOrder(order *Order) error {
    // 模拟处理失败
    if rand.Float64() < 0.5 {
        return errors.New("random processing error")
    }

    // 模拟大额订单失败
    if order.Amount > 500 {
        return errors.New("high value order needs manual review")
    }

    return nil
}

func getRetryCount(headers amqp.Table) int {
    if headers == nil {
        return 0
    }
    if count, ok := headers["x-retry-count"].(int32); ok {
        return int(count)
    }
    return 0
}

func requeue(ch *amqp.Channel, msg amqp.Delivery, retryCount int) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    ch.PublishWithContext(ctx,
        "",
        "orders",
        false,
        false,
        amqp.Publishing{
            ContentType:  msg.ContentType,
            DeliveryMode: amqp.Persistent,
            Body:         msg.Body,
            MessageId:    msg.MessageId,
            Headers: amqp.Table{
                "x-retry-count": retryCount + 1,
            },
        },
    )
}

func processDLQ(ctx context.Context, ch *amqp.Channel) {
    msgs, _ := ch.Consume("orders.dlq", "", false, false, false, false, nil)

    for {
        select {
        case <-ctx.Done():
            return
        case msg, ok := <-msgs:
            if !ok {
                return
            }

            var order Order
            json.Unmarshal(msg.Body, &order)

            log.Printf("Dead letter received: order %s, reason: %v",
                order.ID, msg.Headers["x-death"])

            // 处理死信（记录、告警、人工处理等）
            handleDeadLetter(msg, &order)

            msg.Ack(false)
        }
    }
}

func handleDeadLetter(msg amqp.Delivery, order *Order) {
    // 记录到数据库
    log.Printf("Archiving dead letter: %s", order.ID)

    // 发送告警
    if order.Amount > 500 {
        log.Printf("Alert: High value order %s in DLQ", order.ID)
    }
}
```

## 与 Node.js 对比

### Node.js (Bull Queue)

```javascript
import Queue from 'bull';

const queue = new Queue('orders', {
    defaultJobOptions: {
        attempts: 3,
        backoff: {
            type: 'exponential',
            delay: 1000,
        },
    },
});

// 失败处理
queue.on('failed', (job, err) => {
    console.log(`Job ${job.id} failed:`, err);
    if (job.attemptsMade >= job.opts.attempts) {
        // 移到失败队列
        failedQueue.add(job.data);
    }
});
```

### Go (RabbitMQ DLQ)

```go
// 配置死信队列
ch.QueueDeclare("orders", true, false, false, false, amqp.Table{
    "x-dead-letter-exchange":    "dlx",
    "x-dead-letter-routing-key": "orders.dead",
})

// 处理死信
dlqManager.ProcessDLQ(ctx, "orders", func(dl *DeadLetter) error {
    log.Printf("Dead letter: %s, reason: %s", dl.OriginalQueue, dl.Reason)
    return nil
})
```

## 最佳实践

### 1. 为每个队列配置 DLQ

```go
func setupQueueWithDLQ(ch *amqp.Channel, name string) {
    dlqName := name + ".dlq"
    dlxName := "dlx." + name

    ch.ExchangeDeclare(dlxName, "direct", true, false, false, false, nil)
    ch.QueueDeclare(dlqName, true, false, false, false, nil)
    ch.QueueBind(dlqName, name+".dead", dlxName, false, nil)

    ch.QueueDeclare(name, true, false, false, false, amqp.Table{
        "x-dead-letter-exchange":    dlxName,
        "x-dead-letter-routing-key": name + ".dead",
    })
}
```

### 2. 监控死信队列

```go
// Prometheus 指标
var dlqMessages = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "rabbitmq_dlq_messages",
        Help: "Number of messages in DLQ",
    },
    []string{"queue"},
)
```

### 3. 定期清理

避免死信队列无限增长，设置 TTL 或定期清理。

**下一节**：[可靠性](./08-reliability.md) - 学习消息确认和持久化
