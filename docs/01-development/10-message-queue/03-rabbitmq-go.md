# Go RabbitMQ 客户端 (Go RabbitMQ Client)

## 概述

`amqp091-go` 是 RabbitMQ 官方维护的 Go AMQP 0.9.1 客户端库。

## 安装

```bash
go get github.com/rabbitmq/amqp091-go
```

## 连接

### 基本连接

```go
package main

import (
    "log"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    // 连接 RabbitMQ
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer conn.Close()

    // 创建 Channel
    ch, err := conn.Channel()
    if err != nil {
        log.Fatal("Failed to open channel:", err)
    }
    defer ch.Close()

    log.Println("Connected to RabbitMQ!")
}
```

### 连接 URI 格式

```
amqp://user:password@host:port/vhost

示例:
amqp://guest:guest@localhost:5672/           # 默认 vhost "/"
amqp://admin:admin123@localhost:5672/myapp   # 自定义 vhost
amqp://user:pass@rabbitmq.example.com:5672/  # 远程连接
```

### 使用配置对象

```go
config := amqp.Config{
    Vhost:      "/",
    Heartbeat:  10 * time.Second,
    Locale:     "en_US",
    Properties: amqp.Table{
        "connection_name": "my-app",
    },
}

conn, err := amqp.DialConfig("amqp://guest:guest@localhost:5672/", config)
```

### TLS 连接

```go
import "crypto/tls"

tlsConfig := &tls.Config{
    InsecureSkipVerify: true, // 生产环境应配置正确的证书
}

conn, err := amqp.DialTLS("amqps://guest:guest@localhost:5671/", tlsConfig)
```

## 声明资源

### 声明队列

```go
// 基本队列
q, err := ch.QueueDeclare(
    "my-queue", // 队列名
    true,       // durable: 持久化
    false,      // autoDelete: 自动删除
    false,      // exclusive: 排他
    false,      // noWait: 不等待确认
    nil,        // args: 额外参数
)

// 带参数的队列
q, err := ch.QueueDeclare(
    "ttl-queue",
    true,
    false,
    false,
    false,
    amqp.Table{
        "x-message-ttl":             60000,           // 消息 TTL (毫秒)
        "x-expires":                 3600000,          // 队列过期时间
        "x-max-length":              10000,           // 最大消息数
        "x-max-length-bytes":        1073741824,      // 最大字节数 (1GB)
        "x-dead-letter-exchange":    "dlx",           // 死信交换机
        "x-dead-letter-routing-key": "dlx-key",       // 死信路由键
        "x-max-priority":            10,              // 优先级队列
    },
)
```

### 声明交换机

```go
// Direct 交换机
err := ch.ExchangeDeclare(
    "my-direct", // 名称
    "direct",    // 类型: direct, fanout, topic, headers
    true,        // durable
    false,       // autoDelete
    false,       // internal
    false,       // noWait
    nil,         // args
)

// Fanout 交换机
ch.ExchangeDeclare("my-fanout", "fanout", true, false, false, false, nil)

// Topic 交换机
ch.ExchangeDeclare("my-topic", "topic", true, false, false, false, nil)
```

### 绑定

```go
// 队列绑定到交换机
err := ch.QueueBind(
    "my-queue",   // 队列名
    "my-key",     // routing key
    "my-direct",  // 交换机名
    false,        // noWait
    nil,          // args
)

// 多个绑定
ch.QueueBind("logs-error", "error", "logs", false, nil)
ch.QueueBind("logs-all", "info", "logs", false, nil)
ch.QueueBind("logs-all", "warning", "logs", false, nil)
ch.QueueBind("logs-all", "error", "logs", false, nil)

// Topic 绑定
ch.QueueBind("queue1", "*.error", "my-topic", false, nil)
ch.QueueBind("queue2", "order.*", "my-topic", false, nil)
ch.QueueBind("queue3", "#", "my-topic", false, nil)
```

## 发送消息

### 基本发送

```go
import "context"

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := ch.PublishWithContext(ctx,
    "",           // exchange (空字符串使用默认交换机)
    "my-queue",   // routing key (队列名)
    false,        // mandatory
    false,        // immediate
    amqp.Publishing{
        ContentType: "text/plain",
        Body:        []byte("Hello, RabbitMQ!"),
    },
)
```

### 完整消息属性

```go
msg := amqp.Publishing{
    // 内容
    ContentType:     "application/json",
    ContentEncoding: "utf-8",
    Body:            jsonData,

    // 投递模式
    DeliveryMode:    amqp.Persistent, // 持久化 (1=非持久化, 2=持久化)
    Priority:        5,               // 优先级 0-9

    // 标识
    MessageId:       uuid.New().String(),
    CorrelationId:   correlationID,   // RPC 关联 ID
    ReplyTo:         "reply-queue",   // RPC 回复队列

    // 过期
    Expiration:      "60000",         // 消息 TTL (毫秒)

    // 时间戳
    Timestamp:       time.Now(),

    // 自定义头
    Headers: amqp.Table{
        "x-retry-count": 0,
        "x-source":      "order-service",
    },
}

ch.PublishWithContext(ctx, "", "my-queue", false, false, msg)
```

### 发送 JSON

```go
type Order struct {
    ID        string    `json:"id"`
    UserID    uint      `json:"user_id"`
    Amount    float64   `json:"amount"`
    CreatedAt time.Time `json:"created_at"`
}

func publishOrder(ch *amqp.Channel, order Order) error {
    data, err := json.Marshal(order)
    if err != nil {
        return err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return ch.PublishWithContext(ctx,
        "",
        "orders",
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         data,
            MessageId:    order.ID,
            Timestamp:    time.Now(),
        },
    )
}
```

### 发送到交换机

```go
// 发送到 Direct 交换机
ch.PublishWithContext(ctx, "my-direct", "error", false, false, amqp.Publishing{
    Body: []byte("Error message"),
})

// 发送到 Fanout 交换机 (routing key 被忽略)
ch.PublishWithContext(ctx, "my-fanout", "", false, false, amqp.Publishing{
    Body: []byte("Broadcast message"),
})

// 发送到 Topic 交换机
ch.PublishWithContext(ctx, "my-topic", "order.created", false, false, amqp.Publishing{
    Body: []byte("Order created"),
})
```

## 消费消息

### 基本消费

```go
// 消费消息
msgs, err := ch.Consume(
    "my-queue", // 队列名
    "",         // consumer tag
    false,      // autoAck (设为 false 手动确认)
    false,      // exclusive
    false,      // noLocal
    false,      // noWait
    nil,        // args
)
if err != nil {
    log.Fatal(err)
}

// 处理消息
for msg := range msgs {
    log.Printf("Received: %s", msg.Body)

    // 处理消息...

    // 确认消息
    msg.Ack(false) // false = 只确认当前消息
}
```

### 消息确认

```go
for msg := range msgs {
    err := processMessage(msg)

    if err == nil {
        // 成功，确认消息
        msg.Ack(false)
    } else if isRetryable(err) {
        // 可重试，重新入队
        msg.Nack(false, true) // (multiple, requeue)
    } else {
        // 不可重试，拒绝消息
        msg.Reject(false) // requeue=false，发送到死信队列
    }
}
```

### 消费 JSON

```go
for msg := range msgs {
    var order Order
    if err := json.Unmarshal(msg.Body, &order); err != nil {
        log.Printf("Invalid message: %v", err)
        msg.Reject(false)
        continue
    }

    if err := processOrder(order); err != nil {
        log.Printf("Process error: %v", err)
        msg.Nack(false, true)
        continue
    }

    msg.Ack(false)
}
```

### 设置 QoS (预取)

```go
// 设置预取数量，限制未确认消息数
err := ch.Qos(
    10,    // prefetchCount: 每次预取消息数
    0,     // prefetchSize: 预取大小（字节），0 表示不限制
    false, // global: false=仅当前 channel
)
```

## 封装示例

### RabbitMQ 客户端封装

```go
package mq

import (
    "context"
    "encoding/json"
    "log"
    "sync"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    url     string
    mu      sync.Mutex
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
    mq := &RabbitMQ{url: url}
    if err := mq.connect(); err != nil {
        return nil, err
    }
    return mq, nil
}

func (mq *RabbitMQ) connect() error {
    conn, err := amqp.Dial(mq.url)
    if err != nil {
        return err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return err
    }

    mq.conn = conn
    mq.channel = ch
    return nil
}

func (mq *RabbitMQ) Close() {
    if mq.channel != nil {
        mq.channel.Close()
    }
    if mq.conn != nil {
        mq.conn.Close()
    }
}

// DeclareQueue 声明队列
func (mq *RabbitMQ) DeclareQueue(name string, durable bool) (amqp.Queue, error) {
    return mq.channel.QueueDeclare(name, durable, false, false, false, nil)
}

// Publish 发送消息
func (mq *RabbitMQ) Publish(ctx context.Context, queue string, body interface{}) error {
    data, err := json.Marshal(body)
    if err != nil {
        return err
    }

    return mq.channel.PublishWithContext(ctx,
        "",
        queue,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            Body:         data,
            Timestamp:    time.Now(),
        },
    )
}

// Consume 消费消息
func (mq *RabbitMQ) Consume(queue string, handler func([]byte) error) error {
    msgs, err := mq.channel.Consume(queue, "", false, false, false, false, nil)
    if err != nil {
        return err
    }

    for msg := range msgs {
        if err := handler(msg.Body); err != nil {
            log.Printf("Handler error: %v", err)
            msg.Nack(false, true)
        } else {
            msg.Ack(false)
        }
    }

    return nil
}
```

### 使用示例

```go
func main() {
    // 创建客户端
    mq, err := NewRabbitMQ("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer mq.Close()

    // 声明队列
    mq.DeclareQueue("orders", true)

    // 发送消息
    ctx := context.Background()
    order := map[string]interface{}{
        "id":     "order-123",
        "amount": 99.99,
    }
    mq.Publish(ctx, "orders", order)

    // 消费消息
    mq.Consume("orders", func(body []byte) error {
        var order map[string]interface{}
        json.Unmarshal(body, &order)
        log.Printf("Received order: %v", order)
        return nil
    })
}
```

## 错误处理

### 连接断开处理

```go
func (mq *RabbitMQ) handleReconnect() {
    for {
        select {
        case err := <-mq.conn.NotifyClose(make(chan *amqp.Error)):
            if err != nil {
                log.Printf("Connection closed: %v", err)
                for {
                    time.Sleep(5 * time.Second)
                    if err := mq.connect(); err != nil {
                        log.Printf("Reconnect failed: %v", err)
                        continue
                    }
                    log.Println("Reconnected!")
                    break
                }
            }
        }
    }
}
```

### 确认模式

```go
// 启用发布确认模式
err := ch.Confirm(false)
if err != nil {
    log.Fatal(err)
}

// 获取确认通道
confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

// 发送消息
ch.PublishWithContext(ctx, "", "queue", false, false, amqp.Publishing{
    Body: []byte("message"),
})

// 等待确认
if confirmed := <-confirms; confirmed.Ack {
    log.Println("Message confirmed")
} else {
    log.Println("Message not confirmed")
}
```

**下一节**：[工作队列](./04-work-queues.md) - 学习工作队列模式
