# 消息队列概述 (Message Queue Introduction)

## 什么是消息队列

消息队列 (Message Queue, MQ) 是一种进程间通信或服务间通信的方式，生产者将消息发送到队列，消费者从队列获取消息进行处理。

```
┌──────────┐     ┌─────────────────────────┐     ┌──────────┐
│ Producer │────▶│     Message Queue       │────▶│ Consumer │
│  生产者   │     │  ┌───┬───┬───┬───┬───┐ │     │  消费者   │
│          │     │  │ M │ M │ M │ M │ M │ │     │          │
└──────────┘     │  └───┴───┴───┴───┴───┘ │     └──────────┘
                 │         消息            │
                 └─────────────────────────┘
```

## 为什么需要消息队列

### 1. 异步处理

```
同步处理:
用户注册 ──▶ 保存用户 ──▶ 发送邮件 ──▶ 发送短信 ──▶ 返回
             10ms        200ms       200ms
                     总计: 410ms

异步处理:
用户注册 ──▶ 保存用户 ──▶ 发送消息 ──▶ 返回
             10ms        5ms
                     总计: 15ms

后台处理:
消息队列 ──▶ 发送邮件
         ──▶ 发送短信
```

```go
// 同步处理
func Register(user *User) error {
    if err := db.Save(user); err != nil {
        return err
    }
    if err := sendEmail(user.Email); err != nil {  // 耗时 200ms
        return err
    }
    if err := sendSMS(user.Phone); err != nil {    // 耗时 200ms
        return err
    }
    return nil
}

// 异步处理
func Register(user *User) error {
    if err := db.Save(user); err != nil {
        return err
    }
    // 发送消息到队列，立即返回
    mq.Publish("user.registered", user)  // 耗时 5ms
    return nil
}

// 消费者处理
func handleUserRegistered(user *User) {
    sendEmail(user.Email)
    sendSMS(user.Phone)
}
```

### 2. 服务解耦

```
紧耦合:
┌────────────┐     ┌────────────┐
│  订单服务   │────▶│  库存服务   │  订单服务直接调用库存服务
│            │     │            │  库存服务不可用导致订单失败
│            │────▶│  积分服务   │
└────────────┘     └────────────┘

松耦合:
┌────────────┐     ┌────────────┐     ┌────────────┐
│  订单服务   │────▶│  消息队列   │────▶│  库存服务   │
│            │     │            │────▶│  积分服务   │
│            │     │            │────▶│  通知服务   │
└────────────┘     └────────────┘     └────────────┘
    发布消息           解耦               订阅消息
```

```go
// 紧耦合
func CreateOrder(order *Order) error {
    if err := orderService.Create(order); err != nil {
        return err
    }
    // 直接调用，强依赖
    if err := inventoryService.Deduct(order.ProductID, order.Quantity); err != nil {
        return err  // 库存服务不可用导致订单失败
    }
    if err := pointsService.Add(order.UserID, order.Amount/10); err != nil {
        return err
    }
    return nil
}

// 松耦合
func CreateOrder(order *Order) error {
    if err := orderService.Create(order); err != nil {
        return err
    }
    // 发送事件，不关心谁消费
    mq.Publish("order.created", order)
    return nil
}
```

### 3. 流量削峰

```
无削峰:
┌────────────┐                    ┌────────────┐
│   客户端    │  10000 请求/秒     │   服务器    │  直接处理
│            │ ══════════════════▶│            │  可能崩溃
└────────────┘                    └────────────┘

有削峰:
┌────────────┐     ┌────────────┐     ┌────────────┐
│   客户端    │────▶│  消息队列   │────▶│   服务器    │
│ 10000/秒   │     │   缓冲区    │     │  100/秒    │
└────────────┘     └────────────┘     └────────────┘
                        削峰
```

```go
// 秒杀场景
func Seckill(userID, productID uint) error {
    // 请求入队，快速返回
    msg := SeckillRequest{
        UserID:    userID,
        ProductID: productID,
        Timestamp: time.Now(),
    }
    return mq.Publish("seckill.request", msg)
}

// 后台按能力消费
func processSeckill(msg SeckillRequest) {
    // 按顺序处理，保证系统稳定
    result := doSeckill(msg)
    notifyUser(msg.UserID, result)
}
```

## 消息队列对比

| 特性 | RabbitMQ | Kafka | Redis | RocketMQ |
|------|----------|-------|-------|----------|
| 定位 | 通用消息中间件 | 分布式日志系统 | 内存数据结构 | 电商级消息队列 |
| 吞吐量 | 万级 | 百万级 | 十万级 | 十万级 |
| 延迟 | 微秒级 | 毫秒级 | 微秒级 | 毫秒级 |
| 可靠性 | 高 | 高 | 中 | 高 |
| 协议 | AMQP | 自定义 | RESP | 自定义 |
| 消息确认 | 支持 | 支持 | 不支持 | 支持 |
| 延迟消息 | 插件支持 | 不支持 | 支持 | 支持 |
| 死信队列 | 支持 | 不支持 | 不支持 | 支持 |
| 适用场景 | 业务解耦 | 日志收集、流处理 | 简单队列 | 电商、金融 |

## RabbitMQ 概念

### 基本组件

```
┌──────────────────────────────────────────────────────────────┐
│                        RabbitMQ Broker                       │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐    ┌──────────────────────────────────────┐   │
│  │ Producer │───▶│              Exchange                │   │
│  └──────────┘    │  (Direct/Fanout/Topic/Headers)       │   │
│                  └──────────────┬───────────────────────┘   │
│                                 │ Binding (routing key)     │
│                                 ▼                           │
│                  ┌──────────────────────────────────────┐   │
│                  │              Queue                    │   │
│                  │  ┌───┬───┬───┬───┬───┬───┬───┬───┐  │   │
│                  │  │ M │ M │ M │ M │ M │ M │ M │ M │  │   │
│                  │  └───┴───┴───┴───┴───┴───┴───┴───┘  │   │
│                  └──────────────┬───────────────────────┘   │
│                                 │                           │
│  ┌──────────┐                   │                           │
│  │ Consumer │◀──────────────────┘                           │
│  └──────────┘                                               │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

| 组件 | 说明 |
|------|------|
| Producer | 消息生产者，发送消息 |
| Consumer | 消息消费者，接收处理消息 |
| Exchange | 交换机，接收消息并路由到队列 |
| Queue | 队列，存储消息 |
| Binding | 绑定，Exchange 和 Queue 的关联规则 |
| Routing Key | 路由键，决定消息路由到哪个队列 |
| Connection | 客户端与 Broker 的 TCP 连接 |
| Channel | 连接中的虚拟连接，轻量级 |
| Virtual Host | 虚拟主机，隔离不同应用 |

### Exchange 类型

#### 1. Direct Exchange (直连交换机)

```
                        ┌─────────────────┐
                        │ Direct Exchange │
                        └────────┬────────┘
                                 │
           ┌─────────────────────┼─────────────────────┐
           │                     │                     │
    key="error"           key="info"           key="warning"
           │                     │                     │
           ▼                     ▼                     ▼
    ┌──────────┐          ┌──────────┐          ┌──────────┐
    │  Queue1  │          │  Queue2  │          │  Queue3  │
    └──────────┘          └──────────┘          └──────────┘
```

精确匹配 routing key，适合日志分级、一对一场景。

#### 2. Fanout Exchange (扇出交换机)

```
                        ┌─────────────────┐
                        │ Fanout Exchange │
                        └────────┬────────┘
                                 │
           ┌─────────────────────┼─────────────────────┐
           │                     │                     │
           ▼                     ▼                     ▼
    ┌──────────┐          ┌──────────┐          ┌──────────┐
    │  Queue1  │          │  Queue2  │          │  Queue3  │
    └──────────┘          └──────────┘          └──────────┘
```

广播到所有绑定的队列，忽略 routing key，适合广播通知。

#### 3. Topic Exchange (主题交换机)

```
                        ┌─────────────────┐
                        │ Topic Exchange  │
                        └────────┬────────┘
                                 │
    ┌────────────────────────────┼────────────────────────────┐
    │                            │                            │
key="*.error"            key="order.*"              key="#.critical"
    │                            │                            │
    ▼                            ▼                            ▼
┌──────────┐              ┌──────────┐              ┌──────────┐
│  Queue1  │              │  Queue2  │              │  Queue3  │
└──────────┘              └──────────┘              └──────────┘

匹配规则:
  * : 匹配一个单词
  # : 匹配零个或多个单词

示例:
  order.created    → Queue2
  user.error       → Queue1
  payment.critical → Queue3
  order.critical   → Queue2, Queue3
```

模式匹配 routing key，适合复杂路由场景。

#### 4. Headers Exchange (头交换机)

基于消息头属性匹配，不使用 routing key，适合复杂匹配条件。

## 消息投递模式

### 1. 点对点 (Point-to-Point)

```
Producer ──▶ Queue ──▶ Consumer

一个消息只被一个消费者处理
```

### 2. 发布订阅 (Pub/Sub)

```
              ┌──▶ Queue1 ──▶ Consumer1
Producer ──▶ Exchange
              └──▶ Queue2 ──▶ Consumer2

一个消息被多个消费者处理（每个队列一份）
```

### 3. 工作队列 (Work Queue)

```
                           ┌──▶ Consumer1
Producer ──▶ Queue ────────┼──▶ Consumer2
                           └──▶ Consumer3

多个消费者竞争处理消息，每个消息只被一个消费者处理
```

## 消息属性

```go
type Publishing struct {
    // 消息内容
    ContentType     string    // 内容类型，如 "application/json"
    ContentEncoding string    // 编码
    Body            []byte    // 消息体

    // 投递选项
    DeliveryMode    uint8     // 1=非持久化, 2=持久化
    Priority        uint8     // 优先级 0-9
    Expiration      string    // 过期时间 (毫秒)

    // 标识
    CorrelationId   string    // 关联 ID (用于 RPC)
    ReplyTo         string    // 回复队列
    MessageId       string    // 消息 ID

    // 元数据
    Timestamp       time.Time // 时间戳
    Type            string    // 消息类型
    UserId          string    // 用户 ID
    AppId           string    // 应用 ID
    Headers         Table     // 自定义头
}
```

## 适用场景总结

| 场景 | 模式 | Exchange | 说明 |
|------|------|----------|------|
| 异步任务 | Work Queue | 默认 | 邮件发送、图片处理 |
| 广播通知 | Pub/Sub | Fanout | 配置更新、系统公告 |
| 日志收集 | Routing | Direct/Topic | 按级别路由日志 |
| 订单处理 | Work Queue | 默认 | 订单创建、支付处理 |
| 事件驱动 | Pub/Sub | Fanout/Topic | 用户注册后触发多个服务 |
| RPC 调用 | Request/Reply | 默认 | 同步远程调用 |

**下一节**：[RabbitMQ 安装](./02-rabbitmq-setup.md) - 学习 RabbitMQ 安装和配置
