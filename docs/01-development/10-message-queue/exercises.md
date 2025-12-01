# 消息队列练习 (Message Queue Exercises)

## 练习 1：基本消息收发

### 目标

实现一个简单的消息生产者和消费者。

### 要求

1. 创建一个持久化队列 `hello`
2. 生产者发送 10 条消息
3. 消费者接收并打印消息

### 参考代码

```go
// producer.go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatal(err)
    }
    defer ch.Close()

    // 声明队列
    q, err := ch.QueueDeclare(
        "hello",
        true,  // durable
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 发送 10 条消息
    for i := 1; i <= 10; i++ {
        body := fmt.Sprintf("Message %d", i)
        err := ch.PublishWithContext(ctx,
            "",
            q.Name,
            false,
            false,
            amqp.Publishing{
                DeliveryMode: amqp.Persistent,
                ContentType:  "text/plain",
                Body:         []byte(body),
            },
        )
        if err != nil {
            log.Printf("Failed to send: %v", err)
        } else {
            log.Printf("Sent: %s", body)
        }
    }
}

// consumer.go
package main

import (
    "log"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatal(err)
    }
    defer ch.Close()

    q, err := ch.QueueDeclare("hello", true, false, false, false, nil)
    if err != nil {
        log.Fatal(err)
    }

    msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Waiting for messages...")

    for msg := range msgs {
        log.Printf("Received: %s", msg.Body)
        msg.Ack(false)
    }
}
```

### 预期输出

```
# Producer
Sent: Message 1
Sent: Message 2
...
Sent: Message 10

# Consumer
Received: Message 1
Received: Message 2
...
Received: Message 10
```

---

## 练习 2：工作队列

### 目标

实现多个 Worker 竞争消费任务。

### 要求

1. 创建一个任务队列 `tasks`
2. 发送模拟任务（点号表示耗时：`task...` 表示 3 秒）
3. 启动 2 个 Worker 并发消费
4. 使用 QoS 实现公平分发

### 解决方案要点

```go
// Worker 设置 QoS
ch.Qos(1, 0, false) // 每次只处理 1 条

// 模拟任务处理
func processTask(body []byte) {
    dots := bytes.Count(body, []byte("."))
    time.Sleep(time.Duration(dots) * time.Second)
}
```

---

## 练习 3：发布订阅

### 目标

实现日志广播系统。

### 要求

1. 创建 Fanout 交换机 `logs`
2. 实现日志发布者
3. 实现多个日志订阅者（控制台、文件）
4. 每个订阅者收到所有日志

### 解决方案要点

```go
// 发布者
ch.ExchangeDeclare("logs", "fanout", true, false, false, false, nil)
ch.PublishWithContext(ctx, "logs", "", false, false, amqp.Publishing{
    Body: []byte(logMessage),
})

// 订阅者 - 创建临时队列
q, _ := ch.QueueDeclare("", false, false, true, false, nil)
ch.QueueBind(q.Name, "", "logs", false, nil)
```

---

## 练习 4：路由日志

### 目标

按日志级别路由到不同队列。

### 要求

1. 创建 Direct 交换机 `logs_direct`
2. 支持 `info`, `warning`, `error` 三个级别
3. 实现按级别订阅

### 解决方案要点

```go
// 发布
ch.PublishWithContext(ctx, "logs_direct", "error", false, false, amqp.Publishing{
    Body: []byte(errorMessage),
})

// 订阅 error 和 warning
ch.QueueBind(q.Name, "error", "logs_direct", false, nil)
ch.QueueBind(q.Name, "warning", "logs_direct", false, nil)
```

---

## 练习 5：延迟任务

### 目标

实现订单 30 分钟超时取消。

### 要求

1. 创建订单时发送延迟消息
2. 30 分钟后检查订单状态
3. 如果未支付则取消订单

### 解决方案要点

```go
// 延迟队列设置
ch.QueueDeclare(
    "order.delay.30m",
    true, false, false, false,
    amqp.Table{
        "x-message-ttl":             1800000, // 30 分钟
        "x-dead-letter-exchange":    "dlx.orders",
        "x-dead-letter-routing-key": "order.timeout",
    },
)

// 处理超时
func handleOrderTimeout(orderID string) error {
    order, _ := db.GetOrder(orderID)
    if order.Status == "pending" {
        return db.UpdateOrderStatus(orderID, "cancelled")
    }
    return nil
}
```

---

## 练习 6：消息重试

### 目标

实现失败消息自动重试（最多 3 次）。

### 要求

1. 消费失败时自动重试
2. 使用指数退避（1s, 4s, 16s）
3. 超过重试次数发送到死信队列

### 解决方案要点

```go
type RetryableMessage struct {
    Body       []byte
    RetryCount int
}

func (c *Consumer) handleWithRetry(msg amqp.Delivery, handler Handler) {
    retryCount := getRetryCount(msg.Headers)

    if err := handler(msg.Body); err != nil {
        if retryCount < 3 {
            delay := time.Duration(1<<(retryCount*2)) * time.Second
            c.scheduleRetry(msg, retryCount, delay)
            msg.Ack(false)
        } else {
            msg.Reject(false) // 发送到 DLQ
        }
    } else {
        msg.Ack(false)
    }
}
```

---

## 综合项目：订单处理系统

### 目标

构建一个完整的订单处理系统，包含多个服务通过消息队列通信。

### 架构

```
┌──────────────────────────────────────────────────────────────────┐
│                        订单处理系统                               │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐     ┌─────────────────────────────────────────┐│
│  │ API Gateway │────▶│           RabbitMQ                       ││
│  └─────────────┘     │  ┌───────────────────────────────────┐  ││
│                      │  │ Exchange: orders (topic)           │  ││
│                      │  └───────────────────────────────────┘  ││
│                      │              │                          ││
│                      │   ┌──────────┼──────────┐              ││
│                      │   ▼          ▼          ▼              ││
│                      │ order.*  order.paid  order.shipped     ││
│                      └─────────────────────────────────────────┘│
│                                │                                │
│            ┌───────────────────┼───────────────────┐           │
│            ▼                   ▼                   ▼           │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐  │
│  │ Inventory Svc   │ │ Notification Svc│ │ Analytics Svc   │  │
│  │ (order.created) │ │ (order.*)       │ │ (order.*)       │  │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘  │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 延迟队列: order.timeout.30m ──▶ 订单超时检查                   ││
│  │ 死信队列: order.dlq ──▶ 失败订单处理                           ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 要求

1. **订单服务**
   - 创建订单时发布 `order.created` 事件
   - 支付成功发布 `order.paid` 事件
   - 发货时发布 `order.shipped` 事件

2. **库存服务**
   - 订阅 `order.created`，扣减库存
   - 订阅 `order.cancelled`，恢复库存

3. **通知服务**
   - 订阅 `order.*`，发送邮件/短信通知

4. **延迟任务**
   - 订单创建 30 分钟后检查是否支付
   - 未支付自动取消

5. **可靠性**
   - 发布确认
   - 消息持久化
   - 消费确认
   - 失败重试（3 次）
   - 死信队列

### 代码结构

```
order-system/
├── cmd/
│   ├── api/main.go           # API 服务
│   ├── inventory/main.go     # 库存服务
│   ├── notification/main.go  # 通知服务
│   └── timeout/main.go       # 超时检查服务
├── internal/
│   ├── events/
│   │   └── types.go          # 事件定义
│   ├── mq/
│   │   ├── publisher.go      # 可靠发布者
│   │   └── consumer.go       # 可靠消费者
│   └── models/
│       └── order.go          # 订单模型
├── docker-compose.yml
└── go.mod
```

### 事件定义

```go
// internal/events/types.go
package events

const (
    OrderCreated   = "order.created"
    OrderPaid      = "order.paid"
    OrderShipped   = "order.shipped"
    OrderCancelled = "order.cancelled"
    OrderTimeout   = "order.timeout"
)

type OrderEvent struct {
    EventType string    `json:"event_type"`
    OrderID   string    `json:"order_id"`
    UserID    uint      `json:"user_id"`
    Amount    float64   `json:"amount"`
    Products  []Product `json:"products"`
    Timestamp time.Time `json:"timestamp"`
}

type Product struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Quantity int    `json:"quantity"`
    Price    float64 `json:"price"`
}
```

### Docker Compose

```yaml
version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: admin123
    volumes:
      - rabbitmq-data:/var/lib/rabbitmq

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: orders
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  rabbitmq-data:
  postgres-data:
```

### 评估标准

| 要求 | 分数 |
|------|------|
| 基本消息收发 | 20 |
| 多服务订阅 | 20 |
| 延迟队列实现 | 15 |
| 死信队列配置 | 15 |
| 发布确认 | 10 |
| 消息持久化 | 10 |
| 幂等消费 | 10 |

### 提示

1. 使用 Topic 交换机实现灵活路由
2. 每个服务使用独立的队列
3. 使用 Redis 实现消息去重
4. 监控队列积压情况
5. 记录详细日志便于排查

### 扩展挑战

1. 添加 Prometheus 监控指标
2. 实现优先级队列（VIP 订单优先处理）
3. 添加消息追踪（记录消息流转路径）
4. 实现消费者动态扩缩容
5. 添加消息压缩（大消息体场景）

---

## 总结

通过这些练习，你应该掌握了：

- RabbitMQ 基本操作（队列、交换机、绑定）
- 工作队列和负载均衡
- 发布订阅模式
- 延迟队列实现
- 死信队列和重试机制
- 消息可靠性保证

**下一章**：[可观测性](../11-observability/README.md) - 学习日志、监控、追踪
