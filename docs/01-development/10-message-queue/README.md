# 第十章：消息队列 (Message Queue)

## 章节概述

消息队列是实现异步处理、服务解耦、流量削峰的关键技术，本章学习 RabbitMQ 的使用。

## 学习目标

完成本章后，你将能够：

- 理解消息队列的核心概念和应用场景
- 安装和配置 RabbitMQ
- 使用 Go 实现生产者和消费者
- 实现工作队列、发布订阅等模式
- 处理消息确认、持久化、死信队列
- 实现延迟消息和可靠消息传递

## 与 JavaScript 对比

### Node.js + RabbitMQ

```javascript
import amqp from 'amqplib';

// 连接
const connection = await amqp.connect('amqp://localhost');
const channel = await connection.createChannel();

// 声明队列
await channel.assertQueue('tasks', { durable: true });

// 发送消息
channel.sendToQueue('tasks', Buffer.from('Hello'));

// 消费消息
channel.consume('tasks', (msg) => {
    console.log('Received:', msg.content.toString());
    channel.ack(msg);
});
```

### Go + RabbitMQ

```go
import amqp "github.com/rabbitmq/amqp091-go"

// 连接
conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
ch, _ := conn.Channel()

// 声明队列
q, _ := ch.QueueDeclare("tasks", true, false, false, false, nil)

// 发送消息
ch.PublishWithContext(ctx, "", q.Name, false, false,
    amqp.Publishing{Body: []byte("Hello")})

// 消费消息
msgs, _ := ch.Consume(q.Name, "", false, false, false, false, nil)
for msg := range msgs {
    fmt.Println("Received:", string(msg.Body))
    msg.Ack(false)
}
```

## 章节内容

| 文档 | 主题 | 描述 |
|------|------|------|
| [01-mq-intro.md](./01-mq-intro.md) | 消息队列概述 | 概念和应用场景 |
| [02-rabbitmq-setup.md](./02-rabbitmq-setup.md) | RabbitMQ 安装 | 安装和管理 UI |
| [03-rabbitmq-go.md](./03-rabbitmq-go.md) | Go 客户端 | amqp091-go 使用 |
| [04-work-queues.md](./04-work-queues.md) | 工作队列 | Work Queue 模式 |
| [05-pubsub.md](./05-pubsub.md) | 发布订阅 | Pub/Sub 模式 |
| [06-delayed-queue.md](./06-delayed-queue.md) | 延迟队列 | 延迟消息实现 |
| [07-dead-letter.md](./07-dead-letter.md) | 死信队列 | 消息重试和容错 |
| [08-reliability.md](./08-reliability.md) | 可靠性 | 确认和持久化 |
| [exercises.md](./exercises.md) | 练习 | 实践练习 |

## 快速开始

### 安装 RabbitMQ

```bash
# macOS
brew install rabbitmq
brew services start rabbitmq

# Docker
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management

# 管理 UI
# http://localhost:15672
# 默认用户: guest / guest
```

### 安装 Go 客户端

```bash
go get github.com/rabbitmq/amqp091-go
```

### 基本示例

```go
package main

import (
    "context"
    "log"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    // 连接 RabbitMQ
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // 创建 Channel
    ch, err := conn.Channel()
    if err != nil {
        log.Fatal(err)
    }
    defer ch.Close()

    // 声明队列
    q, err := ch.QueueDeclare(
        "hello", // 队列名
        false,   // 持久化
        false,   // 自动删除
        false,   // 排他
        false,   // 不等待
        nil,     // 参数
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 发送消息
    body := "Hello, RabbitMQ!"
    err = ch.PublishWithContext(ctx,
        "",     // exchange
        q.Name, // routing key
        false,  // mandatory
        false,  // immediate
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte(body),
        })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Sent: %s", body)
}
```

## 消息队列核心概念

### 为什么需要消息队列

```
同步调用问题：
┌────────┐     ┌────────┐     ┌────────┐
│ 服务A  │────▶│ 服务B  │────▶│ 服务C  │
└────────┘     └────────┘     └────────┘
    │                              │
    └──────── 等待响应 ─────────────┘
           总延迟 = A + B + C

异步解耦：
┌────────┐     ┌───────────┐     ┌────────┐
│ 服务A  │────▶│ 消息队列  │────▶│ 服务B  │
└────────┘     │           │     └────────┘
               │           │────▶┌────────┐
               └───────────┘     │ 服务C  │
                                 └────────┘
    A 发送后立即返回，不等待 B、C
```

### 应用场景

| 场景 | 说明 | 示例 |
|------|------|------|
| 异步处理 | 耗时操作异步执行 | 发送邮件、生成报表 |
| 服务解耦 | 服务间松耦合 | 订单服务 → 库存服务 |
| 流量削峰 | 平滑处理突发流量 | 秒杀、促销活动 |
| 日志收集 | 集中收集处理日志 | ELK 日志系统 |
| 事件驱动 | 基于事件的架构 | 用户注册 → 发送欢迎邮件 |

### RabbitMQ 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                        RabbitMQ                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Producer ──▶ Exchange ──▶ Binding ──▶ Queue ──▶ Consumer  │
│                  │                       │                  │
│              路由规则                  消息存储              │
│                                                             │
│  Exchange 类型:                                             │
│  - Direct:  精确匹配 routing key                           │
│  - Fanout:  广播到所有绑定队列                              │
│  - Topic:   模式匹配 routing key                           │
│  - Headers: 基于消息头匹配                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 消息模式对比

| 模式 | Exchange 类型 | 说明 | 适用场景 |
|------|--------------|------|----------|
| Simple | 默认 | 一对一 | 简单任务 |
| Work Queue | 默认 | 多消费者竞争 | 任务分发 |
| Pub/Sub | Fanout | 广播 | 通知、日志 |
| Routing | Direct | 路由 | 日志分级 |
| Topic | Topic | 模式匹配 | 复杂路由 |
| RPC | 默认 | 请求响应 | 远程调用 |

## 最佳实践

| 实践 | 说明 |
|------|------|
| 消息持久化 | 重要消息开启持久化 |
| 手动确认 | 处理完成后手动 ACK |
| 死信队列 | 处理失败消息 |
| 连接复用 | 复用 Connection，多个 Channel |
| 预取限制 | 设置 QoS 防止消费者过载 |
| 幂等处理 | 消费者实现幂等性 |

**下一节**：[消息队列概述](./01-mq-intro.md) - 深入理解消息队列概念
