# 可靠性 (Reliability)

## 概述

消息队列的可靠性涉及消息不丢失、不重复、顺序保证等方面。

```
┌──────────────────────────────────────────────────────────────────┐
│                       消息可靠性保障点                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Producer ──(1)──▶ Broker ──(2)──▶ Queue ──(3)──▶ Consumer      │
│                                                                  │
│  (1) 发布确认: 确保消息到达 Broker                                │
│  (2) 持久化: 消息写入磁盘                                         │
│  (3) 消费确认: 确保消息被正确处理                                  │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 发布确认 (Publisher Confirms)

### 基本确认模式

```go
package main

import (
    "context"
    "log"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
    defer conn.Close()

    ch, _ := conn.Channel()
    defer ch.Close()

    // 启用发布确认模式
    err := ch.Confirm(false) // false = 不等待
    if err != nil {
        log.Fatal("Failed to enable confirm mode:", err)
    }

    // 获取确认通道
    confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

    // 声明队列
    q, _ := ch.QueueDeclare("test-queue", true, false, false, false, nil)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 发送消息
    err = ch.PublishWithContext(ctx,
        "",
        q.Name,
        false,
        false,
        amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            ContentType:  "text/plain",
            Body:         []byte("Hello"),
        },
    )
    if err != nil {
        log.Fatal("Failed to publish:", err)
    }

    // 等待确认
    select {
    case confirm := <-confirms:
        if confirm.Ack {
            log.Printf("Message %d confirmed", confirm.DeliveryTag)
        } else {
            log.Printf("Message %d nacked", confirm.DeliveryTag)
        }
    case <-time.After(5 * time.Second):
        log.Fatal("Confirm timeout")
    }
}
```

### 批量确认

```go
// 批量发送并确认
func publishBatch(ch *amqp.Channel, messages [][]byte) error {
    ch.Confirm(false)
    confirms := ch.NotifyPublish(make(chan amqp.Confirmation, len(messages)))

    ctx := context.Background()

    // 发送所有消息
    for _, body := range messages {
        ch.PublishWithContext(ctx, "", "queue", false, false,
            amqp.Publishing{
                DeliveryMode: amqp.Persistent,
                Body:         body,
            },
        )
    }

    // 等待所有确认
    confirmed := 0
    nacked := 0
    timeout := time.After(30 * time.Second)

    for confirmed+nacked < len(messages) {
        select {
        case confirm := <-confirms:
            if confirm.Ack {
                confirmed++
            } else {
                nacked++
            }
        case <-timeout:
            return fmt.Errorf("confirm timeout: confirmed=%d, nacked=%d, total=%d",
                confirmed, nacked, len(messages))
        }
    }

    if nacked > 0 {
        return fmt.Errorf("%d messages were nacked", nacked)
    }

    return nil
}
```

### 异步确认

```go
// 异步确认处理器
type AsyncPublisher struct {
    channel      *amqp.Channel
    confirms     chan amqp.Confirmation
    pending      map[uint64][]byte // deliveryTag -> message body
    mu           sync.Mutex
    onConfirm    func(tag uint64, ack bool)
}

func NewAsyncPublisher(ch *amqp.Channel) (*AsyncPublisher, error) {
    if err := ch.Confirm(false); err != nil {
        return nil, err
    }

    ap := &AsyncPublisher{
        channel:  ch,
        confirms: ch.NotifyPublish(make(chan amqp.Confirmation, 100)),
        pending:  make(map[uint64][]byte),
    }

    // 启动确认处理
    go ap.handleConfirms()

    return ap, nil
}

func (ap *AsyncPublisher) handleConfirms() {
    for confirm := range ap.confirms {
        ap.mu.Lock()
        if confirm.Ack {
            delete(ap.pending, confirm.DeliveryTag)
        } else {
            // 重试发送
            if body, ok := ap.pending[confirm.DeliveryTag]; ok {
                go ap.Publish(body) // 重新发送
            }
        }

        if ap.onConfirm != nil {
            ap.onConfirm(confirm.DeliveryTag, confirm.Ack)
        }
        ap.mu.Unlock()
    }
}

func (ap *AsyncPublisher) Publish(body []byte) error {
    ap.mu.Lock()
    nextTag := ap.channel.GetNextPublishSeqNo()
    ap.pending[nextTag] = body
    ap.mu.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return ap.channel.PublishWithContext(ctx,
        "",
        "queue",
        false,
        false,
        amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            Body:         body,
        },
    )
}

func (ap *AsyncPublisher) OnConfirm(fn func(tag uint64, ack bool)) {
    ap.onConfirm = fn
}
```

## 消息持久化

### 队列持久化

```go
// 持久化队列
q, err := ch.QueueDeclare(
    "durable-queue",
    true,  // durable = true
    false,
    false,
    false,
    nil,
)
```

### 消息持久化

```go
// 持久化消息
ch.PublishWithContext(ctx, "", "queue", false, false,
    amqp.Publishing{
        DeliveryMode: amqp.Persistent, // 2 = 持久化
        ContentType:  "application/json",
        Body:         data,
    },
)
```

### 交换机持久化

```go
// 持久化交换机
ch.ExchangeDeclare(
    "durable-exchange",
    "direct",
    true,  // durable = true
    false,
    false,
    false,
    nil,
)
```

### 持久化注意事项

```
┌──────────────────────────────────────────────────────────────────┐
│                       持久化要点                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. 三者都需要持久化:                                              │
│     - 交换机 durable=true                                         │
│     - 队列 durable=true                                           │
│     - 消息 DeliveryMode=Persistent                               │
│                                                                  │
│  2. 性能影响:                                                      │
│     - 持久化消息需要写入磁盘                                        │
│     - 吞吐量会降低                                                 │
│     - 建议配合 SSD 使用                                            │
│                                                                  │
│  3. 不是 100% 可靠:                                                │
│     - 消息可能在内存缓冲期间丢失                                    │
│     - 需要配合发布确认使用                                          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 消费确认

### 手动确认模式

```go
// 手动确认
msgs, _ := ch.Consume(
    "queue",
    "",
    false, // autoAck = false (手动确认)
    false,
    false,
    false,
    nil,
)

for msg := range msgs {
    err := processMessage(msg)

    if err == nil {
        // 成功处理，确认消息
        msg.Ack(false)
    } else if isRetryable(err) {
        // 可重试错误，重新入队
        msg.Nack(false, true)
    } else {
        // 不可重试错误，拒绝（发送到 DLQ）
        msg.Reject(false)
    }
}
```

### 批量确认

```go
// 批量确认（每 10 条确认一次）
var count int
for msg := range msgs {
    processMessage(msg)
    count++

    if count % 10 == 0 {
        // 确认 <= deliveryTag 的所有消息
        msg.Ack(true) // multiple = true
    }
}
```

### 预取设置 (QoS)

```go
// 设置预取数量
ch.Qos(
    10,    // prefetchCount: 每次最多预取 10 条
    0,     // prefetchSize: 不限制字节数
    false, // global: false=仅当前 Channel
)
```

## 可靠消息发送封装

```go
package reliable

import (
    "context"
    "encoding/json"
    "errors"
    "log"
    "sync"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher 可靠发布者
type Publisher struct {
    conn        *amqp.Connection
    channel     *amqp.Channel
    confirms    chan amqp.Confirmation
    returns     chan amqp.Return
    pending     map[uint64]*pendingMessage
    mu          sync.RWMutex
    notifyClose chan *amqp.Error
    url         string
    exchange    string
}

type pendingMessage struct {
    body       []byte
    routingKey string
    timestamp  time.Time
    retries    int
}

// Config 配置
type Config struct {
    URL      string
    Exchange string
}

// NewPublisher 创建可靠发布者
func NewPublisher(cfg Config) (*Publisher, error) {
    p := &Publisher{
        url:      cfg.URL,
        exchange: cfg.Exchange,
        pending:  make(map[uint64]*pendingMessage),
    }

    if err := p.connect(); err != nil {
        return nil, err
    }

    // 启动重连监控
    go p.handleReconnect()

    // 启动确认处理
    go p.handleConfirms()

    // 启动返回处理
    go p.handleReturns()

    return p, nil
}

func (p *Publisher) connect() error {
    conn, err := amqp.Dial(p.url)
    if err != nil {
        return err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return err
    }

    // 启用确认模式
    if err := ch.Confirm(false); err != nil {
        ch.Close()
        conn.Close()
        return err
    }

    p.conn = conn
    p.channel = ch
    p.confirms = ch.NotifyPublish(make(chan amqp.Confirmation, 100))
    p.returns = ch.NotifyReturn(make(chan amqp.Return, 100))
    p.notifyClose = conn.NotifyClose(make(chan *amqp.Error, 1))

    return nil
}

func (p *Publisher) handleReconnect() {
    for {
        err := <-p.notifyClose
        if err == nil {
            return // 正常关闭
        }

        log.Printf("Connection closed: %v, reconnecting...", err)

        for {
            time.Sleep(5 * time.Second)

            if err := p.connect(); err != nil {
                log.Printf("Reconnect failed: %v", err)
                continue
            }

            log.Println("Reconnected!")

            // 重发未确认的消息
            p.retryPending()
            break
        }
    }
}

func (p *Publisher) handleConfirms() {
    for confirm := range p.confirms {
        p.mu.Lock()
        if confirm.Ack {
            delete(p.pending, confirm.DeliveryTag)
        } else {
            // Nack: 稍后重试
            if pm, ok := p.pending[confirm.DeliveryTag]; ok {
                pm.retries++
                if pm.retries < 3 {
                    go p.retryMessage(pm)
                } else {
                    log.Printf("Message permanently failed after %d retries", pm.retries)
                    // 可以发送到失败队列
                }
                delete(p.pending, confirm.DeliveryTag)
            }
        }
        p.mu.Unlock()
    }
}

func (p *Publisher) handleReturns() {
    for ret := range p.returns {
        log.Printf("Message returned: %s, reason: %s", ret.RoutingKey, ret.ReplyText)
        // 可以重试或发送到失败队列
    }
}

func (p *Publisher) retryPending() {
    p.mu.RLock()
    pending := make([]*pendingMessage, 0, len(p.pending))
    for _, pm := range p.pending {
        pending = append(pending, pm)
    }
    p.mu.RUnlock()

    for _, pm := range pending {
        p.retryMessage(pm)
    }
}

func (p *Publisher) retryMessage(pm *pendingMessage) {
    time.Sleep(time.Duration(pm.retries) * time.Second) // 指数退避

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    p.PublishWithContext(ctx, pm.routingKey, pm.body)
}

// PublishWithContext 发布消息
func (p *Publisher) PublishWithContext(ctx context.Context, routingKey string, body []byte) error {
    p.mu.Lock()
    nextTag := p.channel.GetNextPublishSeqNo()
    p.pending[nextTag] = &pendingMessage{
        body:       body,
        routingKey: routingKey,
        timestamp:  time.Now(),
    }
    p.mu.Unlock()

    return p.channel.PublishWithContext(ctx,
        p.exchange,
        routingKey,
        true,  // mandatory: 如果无法路由则返回
        false, // immediate
        amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            ContentType:  "application/json",
            Body:         body,
            Timestamp:    time.Now(),
            MessageId:    fmt.Sprintf("%d-%d", time.Now().UnixNano(), nextTag),
        },
    )
}

// Publish 发布 JSON 消息
func (p *Publisher) Publish(ctx context.Context, routingKey string, data interface{}) error {
    body, err := json.Marshal(data)
    if err != nil {
        return err
    }
    return p.PublishWithContext(ctx, routingKey, body)
}

// Close 关闭连接
func (p *Publisher) Close() {
    if p.channel != nil {
        p.channel.Close()
    }
    if p.conn != nil {
        p.conn.Close()
    }
}
```

## 可靠消费封装

```go
package reliable

import (
    "context"
    "log"
    "sync"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer 可靠消费者
type Consumer struct {
    conn        *amqp.Connection
    channel     *amqp.Channel
    url         string
    queue       string
    prefetch    int
    notifyClose chan *amqp.Error
    mu          sync.Mutex
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
    URL      string
    Queue    string
    Prefetch int
}

// NewConsumer 创建可靠消费者
func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
    c := &Consumer{
        url:      cfg.URL,
        queue:    cfg.Queue,
        prefetch: cfg.Prefetch,
    }

    if err := c.connect(); err != nil {
        return nil, err
    }

    return c, nil
}

func (c *Consumer) connect() error {
    conn, err := amqp.Dial(c.url)
    if err != nil {
        return err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return err
    }

    // 设置 QoS
    prefetch := c.prefetch
    if prefetch == 0 {
        prefetch = 10
    }
    ch.Qos(prefetch, 0, false)

    c.conn = conn
    c.channel = ch
    c.notifyClose = conn.NotifyClose(make(chan *amqp.Error, 1))

    return nil
}

// Handler 消息处理函数
type Handler func(body []byte) error

// Consume 开始消费
func (c *Consumer) Consume(ctx context.Context, handler Handler) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := c.consumeOnce(ctx, handler); err != nil {
                log.Printf("Consume error: %v, reconnecting...", err)
                time.Sleep(5 * time.Second)
                c.reconnect()
            }
        }
    }
}

func (c *Consumer) consumeOnce(ctx context.Context, handler Handler) error {
    msgs, err := c.channel.Consume(
        c.queue,
        "",
        false, // autoAck
        false, // exclusive
        false, // noLocal
        false, // noWait
        nil,
    )
    if err != nil {
        return err
    }

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()

        case err := <-c.notifyClose:
            return err

        case msg, ok := <-msgs:
            if !ok {
                return errors.New("channel closed")
            }

            if err := handler(msg.Body); err != nil {
                log.Printf("Handler error: %v", err)
                msg.Nack(false, true) // 重新入队
            } else {
                msg.Ack(false)
            }
        }
    }
}

func (c *Consumer) reconnect() {
    c.mu.Lock()
    defer c.mu.Unlock()

    for {
        if err := c.connect(); err != nil {
            log.Printf("Reconnect failed: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }
        log.Println("Consumer reconnected!")
        break
    }
}

// Close 关闭
func (c *Consumer) Close() {
    if c.channel != nil {
        c.channel.Close()
    }
    if c.conn != nil {
        c.conn.Close()
    }
}
```

## 幂等性保证

### 消息去重

```go
// 使用 Redis 实现消息去重
type IdempotentConsumer struct {
    consumer *Consumer
    redis    *redis.Client
    ttl      time.Duration
}

func (c *IdempotentConsumer) Consume(ctx context.Context, handler Handler) error {
    return c.consumer.Consume(ctx, func(body []byte) error {
        // 生成消息 ID
        messageID := generateMessageID(body)

        // 检查是否已处理
        key := "processed:" + messageID
        ok, err := c.redis.SetNX(ctx, key, "1", c.ttl).Result()
        if err != nil {
            return err
        }

        if !ok {
            // 消息已处理，跳过
            log.Printf("Message %s already processed, skipping", messageID)
            return nil
        }

        // 处理消息
        if err := handler(body); err != nil {
            // 处理失败，删除标记
            c.redis.Del(ctx, key)
            return err
        }

        return nil
    })
}

func generateMessageID(body []byte) string {
    hash := sha256.Sum256(body)
    return hex.EncodeToString(hash[:])
}
```

### 业务幂等

```go
// 订单处理幂等性
func processOrder(ctx context.Context, order *Order) error {
    // 使用订单 ID 作为幂等键
    lockKey := "order:lock:" + order.ID

    // 获取分布式锁
    locked, err := redis.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
    if err != nil {
        return err
    }
    if !locked {
        return errors.New("order is being processed")
    }
    defer redis.Del(ctx, lockKey)

    // 检查订单状态
    existingOrder, err := db.GetOrder(order.ID)
    if err == nil && existingOrder.Status != "pending" {
        // 订单已处理
        return nil
    }

    // 处理订单（幂等操作）
    return db.Transaction(func(tx *gorm.DB) error {
        // 再次检查状态（乐观锁）
        result := tx.Model(&Order{}).
            Where("id = ? AND status = ?", order.ID, "pending").
            Update("status", "processing")

        if result.RowsAffected == 0 {
            return nil // 已被其他进程处理
        }

        // 执行业务逻辑
        if err := doBusinessLogic(tx, order); err != nil {
            return err
        }

        // 更新最终状态
        return tx.Model(&Order{}).
            Where("id = ?", order.ID).
            Update("status", "completed").Error
    })
}
```

## 顺序消息

### 单队列单消费者

```go
// 最简单的顺序保证：单队列单消费者
ch.Qos(1, 0, false) // prefetch = 1

msgs, _ := ch.Consume("ordered-queue", "", false, false, false, false, nil)

for msg := range msgs {
    processMessage(msg) // 串行处理
    msg.Ack(false)
}
```

### 分区顺序

```go
// 使用路由键实现分区顺序
func publishOrderedMessage(ch *amqp.Channel, userID string, body []byte) error {
    // 同一用户的消息发送到同一路由
    routingKey := fmt.Sprintf("user.%s", userID)

    return ch.PublishWithContext(ctx,
        "ordered-exchange",
        routingKey,
        false,
        false,
        amqp.Publishing{
            Body: body,
        },
    )
}

// 使用一致性哈希交换机
ch.ExchangeDeclare(
    "ordered-exchange",
    "x-consistent-hash", // 需要安装插件
    true,
    false,
    false,
    false,
    nil,
)
```

## 监控指标

```go
var (
    // 发布指标
    publishedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mq_messages_published_total",
            Help: "Total published messages",
        },
        []string{"exchange", "routing_key", "status"},
    )

    // 确认指标
    confirmedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mq_messages_confirmed_total",
            Help: "Total confirmed messages",
        },
        []string{"status"}, // ack, nack
    )

    // 消费指标
    consumedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mq_messages_consumed_total",
            Help: "Total consumed messages",
        },
        []string{"queue", "status"}, // success, error
    )

    // 处理延迟
    processLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "mq_message_process_duration_seconds",
            Help:    "Message processing duration",
            Buckets: []float64{.001, .005, .01, .05, .1, .5, 1, 5},
        },
        []string{"queue"},
    )
)
```

## 最佳实践总结

| 实践 | 说明 |
|------|------|
| 启用发布确认 | 确保消息到达 Broker |
| 消息持久化 | 队列、交换机、消息都要持久化 |
| 手动确认 | 处理完成后再确认 |
| 设置合理的 QoS | 根据处理能力设置预取数量 |
| 实现幂等性 | 消费者必须支持重复消费 |
| 配置死信队列 | 处理失败消息 |
| 监控告警 | 监控队列积压、确认率等指标 |
| 自动重连 | 处理网络断开等异常 |

## 与 Node.js 对比

### Node.js (amqplib)

```javascript
// 发布确认
const ch = await conn.createConfirmChannel();

ch.publish('exchange', 'key', Buffer.from('msg'), {}, (err, ok) => {
    if (err) console.log('Message nacked');
    else console.log('Message confirmed');
});

// 消费确认
ch.consume('queue', async (msg) => {
    try {
        await processMessage(msg);
        ch.ack(msg);
    } catch (err) {
        ch.nack(msg, false, true);
    }
}, { noAck: false });
```

### Go

```go
// 发布确认
ch.Confirm(false)
confirms := ch.NotifyPublish(make(chan amqp.Confirmation))

ch.PublishWithContext(ctx, "exchange", "key", false, false, amqp.Publishing{
    Body: []byte("msg"),
})

if confirm := <-confirms; !confirm.Ack {
    log.Println("Message nacked")
}

// 消费确认
msgs, _ := ch.Consume("queue", "", false, false, false, false, nil)

for msg := range msgs {
    if err := processMessage(msg); err != nil {
        msg.Nack(false, true)
    } else {
        msg.Ack(false)
    }
}
```

**下一节**：[练习](./exercises.md) - 消息队列实践练习
