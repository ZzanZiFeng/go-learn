# 工作队列 (Work Queues)

## 概述

工作队列（Work Queue）也称为任务队列（Task Queue），用于在多个消费者之间分配耗时任务。

```
                            ┌──▶ Consumer1 (Worker)
Producer ──▶ Queue ─────────┼──▶ Consumer2 (Worker)
                            └──▶ Consumer3 (Worker)

每个消息只被一个 Worker 处理
```

## 基本工作队列

### 生产者

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

    // 声明持久化队列
    q, err := ch.QueueDeclare(
        "task_queue", // 队列名
        true,         // durable: 持久化
        false,        // autoDelete
        false,        // exclusive
        false,        // noWait
        nil,          // args
    )
    if err != nil {
        log.Fatal(err)
    }

    // 从命令行获取消息
    body := bodyFrom(os.Args)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 发送持久化消息
    err = ch.PublishWithContext(ctx,
        "",     // exchange
        q.Name, // routing key
        false,  // mandatory
        false,  // immediate
        amqp.Publishing{
            DeliveryMode: amqp.Persistent, // 持久化消息
            ContentType:  "text/plain",
            Body:         []byte(body),
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Printf(" [x] Sent %s", body)
}

func bodyFrom(args []string) string {
    if len(args) < 2 || args[1] == "" {
        return "hello"
    }
    return strings.Join(args[1:], " ")
}
```

### 消费者 (Worker)

```go
package main

import (
    "bytes"
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

    // 声明队列（需与生产者一致）
    q, err := ch.QueueDeclare(
        "task_queue",
        true,  // durable
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // 设置预取数量，公平分发
    err = ch.Qos(
        1,     // prefetchCount: 每次只获取 1 条消息
        0,     // prefetchSize
        false, // global
    )
    if err != nil {
        log.Fatal(err)
    }

    // 消费消息
    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer tag
        false,  // autoAck: 手动确认
        false,  // exclusive
        false,  // noLocal
        false,  // noWait
        nil,    // args
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Println(" [*] Waiting for messages. Press CTRL+C to exit")

    for msg := range msgs {
        log.Printf(" [x] Received %s", msg.Body)

        // 模拟耗时任务（每个 . 代表 1 秒）
        dotCount := bytes.Count(msg.Body, []byte("."))
        time.Sleep(time.Duration(dotCount) * time.Second)

        log.Println(" [x] Done")

        // 手动确认
        msg.Ack(false)
    }
}
```

### 运行示例

```bash
# 启动多个 Worker
go run worker.go &
go run worker.go &

# 发送任务
go run new_task.go "First task."
go run new_task.go "Second task.."
go run new_task.go "Third task..."
go run new_task.go "Fourth task...."
go run new_task.go "Fifth task....."

# 任务会在两个 Worker 之间分配
```

## 消息分发策略

### 1. 轮询分发 (Round-Robin)

默认策略，按顺序分发消息给每个消费者。

```
Producer: M1, M2, M3, M4, M5, M6

Consumer1: M1, M3, M5
Consumer2: M2, M4, M6
```

**问题**：不考虑消费者处理能力，可能导致部分消费者过载。

### 2. 公平分发 (Fair Dispatch)

通过 QoS 设置预取数量，确保消费者处理完当前消息后才获取新消息。

```go
// 设置 prefetchCount = 1
// 消费者同时只处理 1 条消息
ch.Qos(1, 0, false)
```

```
Producer: M1, M2, M3, M4, M5, M6

Consumer1 (快): M1, M2, M3, M4
Consumer2 (慢): M5, M6
```

### QoS 配置

```go
err := ch.Qos(
    prefetchCount, // 预取消息数量
    prefetchSize,  // 预取消息大小（字节），0 表示不限制
    global,        // true=整个连接，false=仅当前 Channel
)
```

| 参数 | 说明 | 推荐值 |
|------|------|--------|
| prefetchCount=1 | 严格公平分发 | 低吞吐高公平 |
| prefetchCount=10 | 批量预取 | 高吞吐一般公平 |
| prefetchCount=0 | 无限制 | 最高吞吐无公平 |

## 消息确认机制

### 自动确认 (Auto-Ack)

```go
msgs, _ := ch.Consume(
    q.Name,
    "",
    true,  // autoAck = true
    false,
    false,
    false,
    nil,
)

// 消息投递后立即确认，不保证处理成功
// 适用场景：日志收集等允许丢失的场景
```

### 手动确认 (Manual-Ack)

```go
msgs, _ := ch.Consume(
    q.Name,
    "",
    false, // autoAck = false
    false,
    false,
    false,
    nil,
)

for msg := range msgs {
    err := processMessage(msg)

    if err == nil {
        // 确认消息
        msg.Ack(false) // false = 只确认当前消息
    } else if isRetryable(err) {
        // 重新入队
        msg.Nack(false, true) // (multiple, requeue)
    } else {
        // 拒绝消息（不重新入队）
        msg.Reject(false) // requeue = false
    }
}
```

### 确认方法对比

| 方法 | 说明 | 参数 |
|------|------|------|
| `Ack(multiple)` | 确认消息 | multiple: 是否批量确认 |
| `Nack(multiple, requeue)` | 否定确认 | requeue: 是否重新入队 |
| `Reject(requeue)` | 拒绝消息 | requeue: 是否重新入队 |

### 批量确认

```go
var deliveryTag uint64

for msg := range msgs {
    processMessage(msg)
    deliveryTag = msg.DeliveryTag

    // 每 10 条消息批量确认
    if deliveryTag % 10 == 0 {
        ch.Ack(deliveryTag, true) // true = 确认 <= deliveryTag 的所有消息
    }
}
```

## 消息持久化

### 队列持久化

```go
q, _ := ch.QueueDeclare(
    "task_queue",
    true,  // durable = true
    false,
    false,
    false,
    nil,
)
```

### 消息持久化

```go
ch.PublishWithContext(ctx, "", q.Name, false, false,
    amqp.Publishing{
        DeliveryMode: amqp.Persistent, // 2 = 持久化
        ContentType:  "text/plain",
        Body:         []byte(body),
    },
)
```

### 持久化注意事项

1. **队列和消息都需要持久化**
2. **不是 100% 可靠**：消息可能在磁盘缓冲期间丢失
3. **性能影响**：写入磁盘会降低吞吐量
4. **更强保证**：结合发布确认使用

## 完整工作队列封装

```go
package workqueue

import (
    "context"
    "encoding/json"
    "log"
    "sync"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Task 表示一个任务
type Task struct {
    ID        string      `json:"id"`
    Type      string      `json:"type"`
    Payload   interface{} `json:"payload"`
    CreatedAt time.Time   `json:"created_at"`
    Retries   int         `json:"retries"`
}

// Handler 任务处理函数
type Handler func(task *Task) error

// WorkQueue 工作队列
type WorkQueue struct {
    conn       *amqp.Connection
    channel    *amqp.Channel
    queueName  string
    handlers   map[string]Handler
    mu         sync.RWMutex
    prefetch   int
    maxRetries int
}

// Config 配置
type Config struct {
    URL        string
    QueueName  string
    Prefetch   int
    MaxRetries int
}

// New 创建工作队列
func New(cfg Config) (*WorkQueue, error) {
    conn, err := amqp.Dial(cfg.URL)
    if err != nil {
        return nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, err
    }

    // 声明队列
    _, err = ch.QueueDeclare(
        cfg.QueueName,
        true,  // durable
        false, // autoDelete
        false, // exclusive
        false, // noWait
        amqp.Table{
            "x-dead-letter-exchange":    "",
            "x-dead-letter-routing-key": cfg.QueueName + ".dlq",
        },
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }

    // 声明死信队列
    _, err = ch.QueueDeclare(
        cfg.QueueName+".dlq",
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        ch.Close()
        conn.Close()
        return nil, err
    }

    // 设置 QoS
    prefetch := cfg.Prefetch
    if prefetch == 0 {
        prefetch = 1
    }
    ch.Qos(prefetch, 0, false)

    maxRetries := cfg.MaxRetries
    if maxRetries == 0 {
        maxRetries = 3
    }

    return &WorkQueue{
        conn:       conn,
        channel:    ch,
        queueName:  cfg.QueueName,
        handlers:   make(map[string]Handler),
        prefetch:   prefetch,
        maxRetries: maxRetries,
    }, nil
}

// RegisterHandler 注册任务处理器
func (wq *WorkQueue) RegisterHandler(taskType string, handler Handler) {
    wq.mu.Lock()
    defer wq.mu.Unlock()
    wq.handlers[taskType] = handler
}

// Enqueue 添加任务
func (wq *WorkQueue) Enqueue(ctx context.Context, task *Task) error {
    if task.ID == "" {
        task.ID = generateID()
    }
    task.CreatedAt = time.Now()

    data, err := json.Marshal(task)
    if err != nil {
        return err
    }

    return wq.channel.PublishWithContext(ctx,
        "",
        wq.queueName,
        false,
        false,
        amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            ContentType:  "application/json",
            Body:         data,
            MessageId:    task.ID,
            Timestamp:    task.CreatedAt,
            Headers: amqp.Table{
                "x-task-type": task.Type,
            },
        },
    )
}

// Start 启动消费
func (wq *WorkQueue) Start(ctx context.Context) error {
    msgs, err := wq.channel.Consume(
        wq.queueName,
        "",
        false, // autoAck
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }

    log.Printf("WorkQueue started, queue: %s, prefetch: %d", wq.queueName, wq.prefetch)

    for {
        select {
        case <-ctx.Done():
            log.Println("WorkQueue stopping...")
            return ctx.Err()
        case msg, ok := <-msgs:
            if !ok {
                return nil
            }
            wq.processMessage(ctx, msg)
        }
    }
}

func (wq *WorkQueue) processMessage(ctx context.Context, msg amqp.Delivery) {
    var task Task
    if err := json.Unmarshal(msg.Body, &task); err != nil {
        log.Printf("Invalid task format: %v", err)
        msg.Reject(false) // 不重新入队
        return
    }

    wq.mu.RLock()
    handler, ok := wq.handlers[task.Type]
    wq.mu.RUnlock()

    if !ok {
        log.Printf("No handler for task type: %s", task.Type)
        msg.Reject(false)
        return
    }

    log.Printf("Processing task: %s (type: %s, retries: %d)", task.ID, task.Type, task.Retries)

    if err := handler(&task); err != nil {
        log.Printf("Task failed: %s, error: %v", task.ID, err)

        // 检查重试次数
        if task.Retries < wq.maxRetries {
            task.Retries++
            // 重新入队
            wq.Enqueue(ctx, &task)
            msg.Ack(false)
        } else {
            // 超过重试次数，发送到死信队列
            log.Printf("Task exceeded max retries: %s", task.ID)
            msg.Reject(false) // 发送到 DLQ
        }
        return
    }

    log.Printf("Task completed: %s", task.ID)
    msg.Ack(false)
}

// Close 关闭连接
func (wq *WorkQueue) Close() {
    if wq.channel != nil {
        wq.channel.Close()
    }
    if wq.conn != nil {
        wq.conn.Close()
    }
}

func generateID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
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
    // 创建工作队列
    wq, err := workqueue.New(workqueue.Config{
        URL:        "amqp://guest:guest@localhost:5672/",
        QueueName:  "tasks",
        Prefetch:   5,
        MaxRetries: 3,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer wq.Close()

    // 注册处理器
    wq.RegisterHandler("email", handleEmail)
    wq.RegisterHandler("report", handleReport)
    wq.RegisterHandler("notification", handleNotification)

    // 添加任务
    ctx := context.Background()
    wq.Enqueue(ctx, &workqueue.Task{
        Type: "email",
        Payload: map[string]string{
            "to":      "user@example.com",
            "subject": "Welcome",
            "body":    "Hello!",
        },
    })

    wq.Enqueue(ctx, &workqueue.Task{
        Type: "report",
        Payload: map[string]interface{}{
            "report_id": 123,
            "format":    "pdf",
        },
    })

    // 启动消费者
    ctx, cancel := context.WithCancel(context.Background())

    // 优雅关闭
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        log.Println("Shutdown signal received")
        cancel()
    }()

    // 启动多个 Worker
    for i := 0; i < 3; i++ {
        go wq.Start(ctx)
    }

    <-ctx.Done()
    log.Println("Workers stopped")
}

func handleEmail(task *workqueue.Task) error {
    payload := task.Payload.(map[string]interface{})
    log.Printf("Sending email to: %s", payload["to"])
    time.Sleep(2 * time.Second) // 模拟发送
    return nil
}

func handleReport(task *workqueue.Task) error {
    payload := task.Payload.(map[string]interface{})
    log.Printf("Generating report: %v", payload["report_id"])
    time.Sleep(5 * time.Second) // 模拟生成
    return nil
}

func handleNotification(task *workqueue.Task) error {
    log.Printf("Sending notification: %v", task.Payload)
    return nil
}
```

## 与 Node.js 对比

### Node.js (Bull Queue)

```javascript
import Queue from 'bull';

// 创建队列
const emailQueue = new Queue('email', 'redis://localhost:6379');

// 添加任务
await emailQueue.add({
    to: 'user@example.com',
    subject: 'Welcome',
});

// 处理任务
emailQueue.process(async (job) => {
    console.log('Processing:', job.data);
    await sendEmail(job.data);
});

// 并发处理
emailQueue.process(5, async (job) => {
    // 最多 5 个并发
});
```

### Go (RabbitMQ)

```go
// 创建队列
wq, _ := workqueue.New(workqueue.Config{
    URL:       "amqp://guest:guest@localhost:5672/",
    QueueName: "email",
    Prefetch:  5, // 并发控制
})

// 添加任务
wq.Enqueue(ctx, &workqueue.Task{
    Type: "email",
    Payload: map[string]string{
        "to":      "user@example.com",
        "subject": "Welcome",
    },
})

// 处理任务
wq.RegisterHandler("email", func(task *workqueue.Task) error {
    fmt.Println("Processing:", task.Payload)
    return sendEmail(task.Payload)
})

// 启动多个 Worker
for i := 0; i < 5; i++ {
    go wq.Start(ctx)
}
```

## 最佳实践

### 1. 任务幂等性

```go
func handleOrder(task *workqueue.Task) error {
    orderID := task.Payload.(map[string]interface{})["order_id"].(string)

    // 检查是否已处理
    processed, _ := redis.Get(ctx, "processed:"+orderID).Bool()
    if processed {
        log.Printf("Order already processed: %s", orderID)
        return nil // 幂等返回成功
    }

    // 处理订单
    if err := processOrder(orderID); err != nil {
        return err
    }

    // 标记已处理
    redis.Set(ctx, "processed:"+orderID, true, 24*time.Hour)
    return nil
}
```

### 2. 任务超时

```go
func handleWithTimeout(task *workqueue.Task) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    done := make(chan error, 1)
    go func() {
        done <- doWork(task)
    }()

    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return fmt.Errorf("task timeout: %s", task.ID)
    }
}
```

### 3. 优先级队列

```go
// 创建不同优先级的队列
highPriority, _ := workqueue.New(workqueue.Config{
    QueueName: "tasks.high",
    Prefetch:  10,
})

lowPriority, _ := workqueue.New(workqueue.Config{
    QueueName: "tasks.low",
    Prefetch:  2,
})

// 高优先级任务
highPriority.Enqueue(ctx, &workqueue.Task{
    Type:    "urgent",
    Payload: data,
})

// 低优先级任务
lowPriority.Enqueue(ctx, &workqueue.Task{
    Type:    "batch",
    Payload: data,
})
```

### 4. 监控指标

```go
var (
    tasksProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "workqueue_tasks_processed_total",
            Help: "Total processed tasks",
        },
        []string{"type", "status"},
    )

    taskDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "workqueue_task_duration_seconds",
            Help:    "Task processing duration",
            Buckets: []float64{.1, .5, 1, 2, 5, 10},
        },
        []string{"type"},
    )
)

func handleWithMetrics(handler Handler) Handler {
    return func(task *workqueue.Task) error {
        start := time.Now()
        err := handler(task)

        duration := time.Since(start).Seconds()
        taskDuration.WithLabelValues(task.Type).Observe(duration)

        status := "success"
        if err != nil {
            status = "error"
        }
        tasksProcessed.WithLabelValues(task.Type, status).Inc()

        return err
    }
}
```

**下一节**：[发布订阅](./05-pubsub.md) - 学习 Pub/Sub 模式
