# 延迟队列 (Delayed Queue)

## 概述

延迟队列用于在指定时间后处理消息，常用于订单超时取消、定时任务、延迟通知等场景。

```
Producer ──▶ 延迟队列 ──(等待 N 秒)──▶ 实际队列 ──▶ Consumer

消息在延迟队列中等待指定时间后，转发到实际队列
```

## 实现方式

RabbitMQ 实现延迟队列主要有两种方式：

1. **TTL + 死信队列** - 原生支持，无需插件
2. **延迟消息插件** - 需要安装 `rabbitmq_delayed_message_exchange` 插件

## 方式一：TTL + 死信队列

### 原理

```
┌──────────────────────────────────────────────────────────────────┐
│                          延迟队列实现                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Producer ──▶ delay.queue ──(TTL 过期)──▶ DLX ──▶ work.queue    │
│               (设置 TTL)      (死信交换机)      ──▶ Consumer     │
│                                                                  │
│  消息流程:                                                        │
│  1. 消息发送到 delay.queue                                        │
│  2. 消息在队列中等待 TTL 时间                                      │
│  3. 过期后通过死信交换机转发到 work.queue                          │
│  4. Consumer 从 work.queue 消费消息                               │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 固定延迟时间

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

    // 1. 声明死信交换机
    ch.ExchangeDeclare("dlx", "direct", true, false, false, false, nil)

    // 2. 声明工作队列（消费队列）
    ch.QueueDeclare("work.queue", true, false, false, false, nil)

    // 3. 绑定工作队列到死信交换机
    ch.QueueBind("work.queue", "work", "dlx", false, nil)

    // 4. 声明延迟队列（设置 TTL 和死信交换机）
    ch.QueueDeclare(
        "delay.queue.30s",
        true,
        false,
        false,
        false,
        amqp.Table{
            "x-message-ttl":             30000,  // 30 秒 TTL
            "x-dead-letter-exchange":    "dlx",  // 死信交换机
            "x-dead-letter-routing-key": "work", // 死信路由键
        },
    )

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 5. 发送延迟消息
    ch.PublishWithContext(ctx,
        "",               // 默认交换机
        "delay.queue.30s", // 延迟队列
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte("Delayed message"),
        },
    )

    log.Println("Message sent, will be delivered in 30 seconds")
}
```

### 动态延迟时间

使用消息级别的 TTL 实现动态延迟：

```go
package delayqueue

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// DelayQueue 延迟队列
type DelayQueue struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    dlx     string // 死信交换机
    workQ   string // 工作队列
}

// Config 配置
type Config struct {
    URL      string
    DLX      string // 死信交换机名称
    WorkQ    string // 工作队列名称
}

// DelayedMessage 延迟消息
type DelayedMessage struct {
    ID        string      `json:"id"`
    Type      string      `json:"type"`
    Payload   interface{} `json:"payload"`
    DelayMs   int64       `json:"delay_ms"`
    CreatedAt time.Time   `json:"created_at"`
}

// New 创建延迟队列
func New(cfg Config) (*DelayQueue, error) {
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
        conn:    conn,
        channel: ch,
        dlx:     cfg.DLX,
        workQ:   cfg.WorkQ,
    }

    if err := dq.setup(); err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }

    return dq, nil
}

func (dq *DelayQueue) setup() error {
    // 声明死信交换机
    if err := dq.channel.ExchangeDeclare(dq.dlx, "direct", true, false, false, false, nil); err != nil {
        return err
    }

    // 声明工作队列
    if _, err := dq.channel.QueueDeclare(dq.workQ, true, false, false, false, nil); err != nil {
        return err
    }

    // 绑定工作队列到死信交换机
    return dq.channel.QueueBind(dq.workQ, "delayed", dq.dlx, false, nil)
}

// getOrCreateDelayQueue 获取或创建指定延迟时间的队列
func (dq *DelayQueue) getOrCreateDelayQueue(delayMs int64) (string, error) {
    queueName := fmt.Sprintf("delay.queue.%dms", delayMs)

    _, err := dq.channel.QueueDeclare(
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
    )

    return queueName, err
}

// Publish 发送延迟消息
func (dq *DelayQueue) Publish(ctx context.Context, msg *DelayedMessage) error {
    if msg.CreatedAt.IsZero() {
        msg.CreatedAt = time.Now()
    }

    // 获取或创建延迟队列
    queueName, err := dq.getOrCreateDelayQueue(msg.DelayMs)
    if err != nil {
        return err
    }

    data, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    return dq.channel.PublishWithContext(ctx,
        "",
        queueName,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         data,
            MessageId:    msg.ID,
        },
    )
}

// PublishWithTTL 使用消息级 TTL 发送延迟消息
func (dq *DelayQueue) PublishWithTTL(ctx context.Context, msg *DelayedMessage) error {
    if msg.CreatedAt.IsZero() {
        msg.CreatedAt = time.Now()
    }

    // 使用统一的延迟队列（无队列级 TTL）
    queueName := "delay.queue.dynamic"

    // 确保队列存在
    _, err := dq.channel.QueueDeclare(
        queueName,
        true,
        false,
        false,
        false,
        amqp.Table{
            "x-dead-letter-exchange":    dq.dlx,
            "x-dead-letter-routing-key": "delayed",
        },
    )
    if err != nil {
        return err
    }

    data, _ := json.Marshal(msg)

    // 使用消息级 TTL
    return dq.channel.PublishWithContext(ctx,
        "",
        queueName,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         data,
            MessageId:    msg.ID,
            Expiration:   fmt.Sprintf("%d", msg.DelayMs), // 消息级 TTL
        },
    )
}

// Consume 消费延迟消息
func (dq *DelayQueue) Consume(ctx context.Context, handler func(*DelayedMessage) error) error {
    msgs, err := dq.channel.Consume(dq.workQ, "", false, false, false, false, nil)
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

            var delayedMsg DelayedMessage
            if err := json.Unmarshal(msg.Body, &delayedMsg); err != nil {
                msg.Reject(false)
                continue
            }

            if err := handler(&delayedMsg); err != nil {
                msg.Nack(false, true)
            } else {
                msg.Ack(false)
            }
        }
    }
}

// Close 关闭连接
func (dq *DelayQueue) Close() {
    if dq.channel != nil {
        dq.channel.Close()
    }
    if dq.conn != nil {
        dq.conn.Close()
    }
}
```

### 使用示例

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    dq, err := delayqueue.New(delayqueue.Config{
        URL:   "amqp://guest:guest@localhost:5672/",
        DLX:   "dlx.delayed",
        WorkQ: "work.queue",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer dq.Close()

    ctx, cancel := context.WithCancel(context.Background())

    // 优雅关闭
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    // 启动消费者
    go func() {
        dq.Consume(ctx, func(msg *delayqueue.DelayedMessage) error {
            delay := time.Since(msg.CreatedAt)
            log.Printf("Received: %s, actual delay: %v, expected: %dms",
                msg.Type, delay, msg.DelayMs)
            return nil
        })
    }()

    // 发送延迟消息
    messages := []struct {
        Type    string
        DelayMs int64
    }{
        {"order.timeout", 30000},   // 30 秒
        {"reminder", 60000},        // 1 分钟
        {"notification", 5000},     // 5 秒
    }

    for _, m := range messages {
        dq.Publish(ctx, &delayqueue.DelayedMessage{
            ID:      fmt.Sprintf("msg-%d", time.Now().UnixNano()),
            Type:    m.Type,
            DelayMs: m.DelayMs,
            Payload: map[string]string{"message": "test"},
        })
        log.Printf("Sent: %s, delay: %dms", m.Type, m.DelayMs)
    }

    <-ctx.Done()
}
```

## 方式二：延迟消息插件

### 安装插件

```bash
# 启用插件
rabbitmq-plugins enable rabbitmq_delayed_message_exchange

# Docker 方式
docker exec rabbitmq rabbitmq-plugins enable rabbitmq_delayed_message_exchange
docker restart rabbitmq
```

### 使用插件

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

    // 声明延迟交换机
    err := ch.ExchangeDeclare(
        "delayed.exchange",
        "x-delayed-message", // 延迟消息类型
        true,
        false,
        false,
        false,
        amqp.Table{
            "x-delayed-type": "direct", // 实际路由类型
        },
    )
    if err != nil {
        log.Fatal("Plugin not installed:", err)
    }

    // 声明队列
    q, _ := ch.QueueDeclare("delayed.queue", true, false, false, false, nil)

    // 绑定队列
    ch.QueueBind(q.Name, "delayed", "delayed.exchange", false, nil)

    ctx := context.Background()

    // 发送延迟消息
    ch.PublishWithContext(ctx,
        "delayed.exchange",
        "delayed",
        false,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte("Delayed message via plugin"),
            Headers: amqp.Table{
                "x-delay": 30000, // 延迟 30 秒（毫秒）
            },
        },
    )

    log.Println("Message sent, will be delivered in 30 seconds")
}
```

### 插件方式封装

```go
package delayqueue

import (
    "context"
    "encoding/json"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// PluginDelayQueue 使用插件的延迟队列
type PluginDelayQueue struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    exchange string
    queue    string
}

// NewPluginDelayQueue 创建延迟队列（插件版）
func NewPluginDelayQueue(url, exchange, queue string) (*PluginDelayQueue, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }

    // 声明延迟交换机
    err = ch.ExchangeDeclare(
        exchange,
        "x-delayed-message",
        true,
        false,
        false,
        false,
        amqp.Table{
            "x-delayed-type": "direct",
        },
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }

    // 声明队列
    _, err = ch.QueueDeclare(queue, true, false, false, false, nil)
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }

    // 绑定队列
    ch.QueueBind(queue, "delayed", exchange, false, nil)

    return &PluginDelayQueue{
        conn:     conn,
        channel:  ch,
        exchange: exchange,
        queue:    queue,
    }, nil
}

// Publish 发送延迟消息
func (dq *PluginDelayQueue) Publish(ctx context.Context, msg *DelayedMessage) error {
    if msg.CreatedAt.IsZero() {
        msg.CreatedAt = time.Now()
    }

    data, _ := json.Marshal(msg)

    return dq.channel.PublishWithContext(ctx,
        dq.exchange,
        "delayed",
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         data,
            MessageId:    msg.ID,
            Headers: amqp.Table{
                "x-delay": msg.DelayMs, // 延迟时间
            },
        },
    )
}

// Consume 消费延迟消息
func (dq *PluginDelayQueue) Consume(ctx context.Context, handler func(*DelayedMessage) error) error {
    msgs, err := dq.channel.Consume(dq.queue, "", false, false, false, false, nil)
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

            var delayedMsg DelayedMessage
            if err := json.Unmarshal(msg.Body, &delayedMsg); err != nil {
                msg.Reject(false)
                continue
            }

            if err := handler(&delayedMsg); err != nil {
                msg.Nack(false, true)
            } else {
                msg.Ack(false)
            }
        }
    }
}

// Close 关闭连接
func (dq *PluginDelayQueue) Close() {
    if dq.channel != nil {
        dq.channel.Close()
    }
    if dq.conn != nil {
        dq.conn.Close()
    }
}
```

## 两种方式对比

| 特性 | TTL + 死信队列 | 延迟消息插件 |
|------|---------------|-------------|
| 安装要求 | 无需额外安装 | 需要安装插件 |
| 延迟精度 | 队列级或消息级 TTL | 消息级精确延迟 |
| 队列管理 | 可能需要多个队列 | 单个交换机 |
| 消息顺序 | 先入先出（可能阻塞） | 按延迟时间排序 |
| 性能 | 较高 | 中等 |
| 最大延迟 | 无限制 | 依赖插件版本 |

### TTL + 死信队列的问题

消息级 TTL 的队列中，消息按先入先出顺序检查过期：

```
队列: [M1(TTL=60s), M2(TTL=5s), M3(TTL=10s)]

问题: M2 和 M3 虽然 TTL 更短，但需要等 M1 过期后才能检查
解决: 使用不同 TTL 的多个队列，或使用插件
```

## 实际应用场景

### 1. 订单超时取消

```go
// 创建订单时发送延迟消息
func CreateOrder(ctx context.Context, order *Order) error {
    // 保存订单
    if err := db.Create(order).Error; err != nil {
        return err
    }

    // 发送 30 分钟后的超时检查消息
    return delayQueue.Publish(ctx, &delayqueue.DelayedMessage{
        ID:      order.ID,
        Type:    "order.timeout.check",
        DelayMs: 30 * 60 * 1000, // 30 分钟
        Payload: map[string]string{
            "order_id": order.ID,
        },
    })
}

// 处理超时检查
func handleOrderTimeout(msg *delayqueue.DelayedMessage) error {
    orderID := msg.Payload.(map[string]interface{})["order_id"].(string)

    var order Order
    if err := db.First(&order, "id = ?", orderID).Error; err != nil {
        return err
    }

    // 检查订单状态
    if order.Status == "pending" {
        // 取消订单
        order.Status = "cancelled"
        order.CancelReason = "超时未支付"
        return db.Save(&order).Error
    }

    return nil // 订单已支付或已取消，无需处理
}
```

### 2. 延迟通知

```go
// 用户注册后 24 小时发送引导邮件
func OnUserRegistered(ctx context.Context, user *User) error {
    // 立即发送欢迎邮件
    sendWelcomeEmail(user.Email)

    // 24 小时后发送引导邮件
    return delayQueue.Publish(ctx, &delayqueue.DelayedMessage{
        ID:      fmt.Sprintf("guide-%d", user.ID),
        Type:    "email.guide",
        DelayMs: 24 * 60 * 60 * 1000, // 24 小时
        Payload: map[string]interface{}{
            "user_id": user.ID,
            "email":   user.Email,
        },
    })
}

// 处理延迟邮件
func handleGuideEmail(msg *delayqueue.DelayedMessage) error {
    payload := msg.Payload.(map[string]interface{})
    email := payload["email"].(string)

    return sendGuideEmail(email)
}
```

### 3. 定时任务

```go
// 定时任务调度器
type Scheduler struct {
    dq *delayqueue.DelayQueue
}

// Schedule 调度定时任务
func (s *Scheduler) Schedule(taskType string, payload interface{}, executeAt time.Time) error {
    delay := time.Until(executeAt)
    if delay < 0 {
        delay = 0
    }

    return s.dq.Publish(context.Background(), &delayqueue.DelayedMessage{
        ID:      fmt.Sprintf("task-%d", time.Now().UnixNano()),
        Type:    taskType,
        DelayMs: delay.Milliseconds(),
        Payload: payload,
    })
}

// ScheduleAfter 延迟执行任务
func (s *Scheduler) ScheduleAfter(taskType string, payload interface{}, delay time.Duration) error {
    return s.dq.Publish(context.Background(), &delayqueue.DelayedMessage{
        ID:      fmt.Sprintf("task-%d", time.Now().UnixNano()),
        Type:    taskType,
        DelayMs: delay.Milliseconds(),
        Payload: payload,
    })
}

// 使用示例
func main() {
    scheduler := &Scheduler{dq: delayQueue}

    // 5 分钟后执行
    scheduler.ScheduleAfter("report.generate", map[string]interface{}{
        "report_id": 123,
    }, 5*time.Minute)

    // 指定时间执行
    executeAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
    scheduler.Schedule("newsletter.send", map[string]interface{}{
        "campaign_id": 456,
    }, executeAt)
}
```

### 4. 重试机制

```go
// 带重试的任务处理
func handleWithRetry(msg *delayqueue.DelayedMessage) error {
    payload := msg.Payload.(map[string]interface{})
    retryCount := int(payload["retry_count"].(float64))

    err := processTask(msg)
    if err != nil {
        if retryCount < 3 {
            // 指数退避重试：1s, 4s, 16s
            delay := time.Duration(1<<(retryCount*2)) * time.Second

            payload["retry_count"] = retryCount + 1
            msg.Payload = payload
            msg.DelayMs = delay.Milliseconds()

            return delayQueue.Publish(context.Background(), msg)
        }

        // 超过重试次数，发送到失败队列
        return sendToFailedQueue(msg, err)
    }

    return nil
}
```

## 与 Node.js 对比

### Node.js (Bull Queue)

```javascript
import Queue from 'bull';

const queue = new Queue('delayed-tasks', 'redis://localhost:6379');

// 延迟任务
await queue.add(
    { orderId: '123' },
    { delay: 30 * 60 * 1000 } // 30 分钟
);

// 处理任务
queue.process(async (job) => {
    console.log('Processing:', job.data);
});
```

### Go (RabbitMQ)

```go
// 延迟任务
delayQueue.Publish(ctx, &delayqueue.DelayedMessage{
    Type:    "order.timeout",
    DelayMs: 30 * 60 * 1000, // 30 分钟
    Payload: map[string]string{"order_id": "123"},
})

// 处理任务
delayQueue.Consume(ctx, func(msg *delayqueue.DelayedMessage) error {
    log.Printf("Processing: %v", msg.Payload)
    return nil
})
```

## 最佳实践

### 1. 选择合适的实现方式

- **固定延迟时间**：使用队列级 TTL + 死信队列
- **动态延迟时间**：使用插件或消息级 TTL + 多队列
- **高精度要求**：使用延迟消息插件

### 2. 消息幂等性

延迟消息可能被重复投递，消费者必须实现幂等处理。

### 3. 监控延迟队列

```go
var (
    delayedMessagesGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "delayed_queue_messages",
            Help: "Number of messages in delay queue",
        },
        []string{"queue"},
    )
)
```

### 4. 设置合理的最大延迟

避免设置过长的延迟时间，可能导致消息积压。

**下一节**：[死信队列](./07-dead-letter.md) - 学习死信队列和消息重试
