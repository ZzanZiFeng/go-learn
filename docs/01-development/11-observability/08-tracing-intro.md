# 分布式追踪概念 (Distributed Tracing)

## 概述

分布式追踪用于跟踪请求在分布式系统中的完整路径，帮助理解服务间的调用关系和性能瓶颈。

```
┌──────────────────────────────────────────────────────────────────┐
│                     分布式追踪概念                                 │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  请求流程:                                                        │
│                                                                  │
│  Client ──▶ API Gateway ──▶ User Service ──▶ Database           │
│                  │                │                              │
│                  │                └──▶ Cache                     │
│                  │                                               │
│                  └──▶ Order Service ──▶ Payment Service         │
│                            │                                     │
│                            └──▶ Inventory Service               │
│                                                                  │
│  Trace (追踪):                                                   │
│  一个完整请求从开始到结束的全部记录                                  │
│                                                                  │
│  Span (跨度):                                                    │
│  追踪中的一个操作单元（如一个服务调用、数据库查询）                    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 核心概念

### Trace 和 Span

```
Trace ID: abc123
│
├── Span: API Gateway (总耗时: 200ms)
│   ├── trace_id: abc123
│   ├── span_id: span001
│   ├── parent_span_id: null
│   ├── operation: "HTTP GET /api/orders/123"
│   ├── start_time: 2024-01-01T10:00:00.000Z
│   ├── duration: 200ms
│   └── tags: {http.method: "GET", http.status: 200}
│
├── Span: User Service (耗时: 50ms)
│   ├── trace_id: abc123
│   ├── span_id: span002
│   ├── parent_span_id: span001
│   ├── operation: "getUser"
│   └── duration: 50ms
│
├── Span: Order Service (耗时: 100ms)
│   ├── trace_id: abc123
│   ├── span_id: span003
│   ├── parent_span_id: span001
│   ├── operation: "getOrder"
│   └── duration: 100ms
│   │
│   └── Span: Database Query (耗时: 30ms)
│       ├── trace_id: abc123
│       ├── span_id: span004
│       ├── parent_span_id: span003
│       ├── operation: "SELECT * FROM orders"
│       └── duration: 30ms
│
└── Span: Payment Service (耗时: 80ms)
    ├── trace_id: abc123
    ├── span_id: span005
    ├── parent_span_id: span001
    └── duration: 80ms
```

### Span 属性

```go
// Span 包含的信息
type Span struct {
    TraceID      string            // 追踪 ID（整个请求唯一）
    SpanID       string            // 当前 Span ID
    ParentSpanID string            // 父 Span ID
    OperationName string           // 操作名称
    StartTime    time.Time         // 开始时间
    Duration     time.Duration     // 持续时间
    Tags         map[string]string // 标签（索引字段）
    Logs         []Log             // 日志事件
    Status       Status            // 状态（OK, Error）
}

// 常用标签
tags := map[string]string{
    "http.method":      "GET",
    "http.url":         "/api/users/123",
    "http.status_code": "200",
    "db.type":          "postgresql",
    "db.statement":     "SELECT * FROM users",
    "error":            "true",
    "error.message":    "connection timeout",
}
```

### 上下文传播

```
┌──────────────────────────────────────────────────────────────────┐
│                     上下文传播 (Context Propagation)               │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Service A                      Service B                        │
│  ┌─────────────────┐           ┌─────────────────┐              │
│  │ Create Span     │           │ Extract Context │              │
│  │ trace_id: abc   │──HTTP────▶│ trace_id: abc   │              │
│  │ span_id: 001    │  Header   │ parent_id: 001  │              │
│  └─────────────────┘           │ span_id: 002    │              │
│                                └─────────────────┘              │
│                                                                  │
│  HTTP Headers:                                                   │
│  traceparent: 00-abc123-span001-01                              │
│  tracestate: vendor=value                                        │
│                                                                  │
│  传播格式:                                                        │
│  • W3C Trace Context (标准)                                      │
│  • B3 (Zipkin)                                                   │
│  • Jaeger                                                        │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## OpenTelemetry 基础

OpenTelemetry 是 CNCF 的可观测性标准，整合了追踪、指标、日志。

### 安装

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/trace
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/otel/exporters/jaeger
go get go.opentelemetry.io/otel/propagation
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
```

### 初始化 Tracer

```go
package tracing

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitTracer 初始化追踪器
func InitTracer(serviceName, jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
    // 创建 Jaeger Exporter
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
    )
    if err != nil {
        return nil, err
    }

    // 创建资源
    res, err := resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(serviceName),
            semconv.ServiceVersion("1.0.0"),
        ),
    )
    if err != nil {
        return nil, err
    }

    // 创建 TracerProvider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()), // 生产环境可调整采样率
    )

    // 设置全局 TracerProvider
    otel.SetTracerProvider(tp)

    // 设置全局传播器
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp, nil
}

// Shutdown 关闭追踪器
func Shutdown(tp *sdktrace.TracerProvider) error {
    return tp.Shutdown(context.Background())
}
```

### 创建 Span

```go
package main

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("myapp")

// 基本 Span 创建
func ProcessOrder(ctx context.Context, orderID string) error {
    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()

    // 添加属性
    span.SetAttributes(
        attribute.String("order.id", orderID),
        attribute.String("order.status", "processing"),
    )

    // 调用子函数（会创建子 Span）
    if err := validateOrder(ctx, orderID); err != nil {
        // 记录错误
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    if err := chargePayment(ctx, orderID); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "Order processed successfully")
    return nil
}

func validateOrder(ctx context.Context, orderID string) error {
    ctx, span := tracer.Start(ctx, "ValidateOrder")
    defer span.End()

    span.SetAttributes(attribute.String("order.id", orderID))

    // 添加事件
    span.AddEvent("validation_started")

    // 验证逻辑...

    span.AddEvent("validation_completed")
    return nil
}

func chargePayment(ctx context.Context, orderID string) error {
    ctx, span := tracer.Start(ctx, "ChargePayment",
        trace.WithSpanKind(trace.SpanKindClient), // 表示这是一个客户端调用
    )
    defer span.End()

    span.SetAttributes(
        attribute.String("order.id", orderID),
        attribute.String("payment.provider", "stripe"),
    )

    // 支付逻辑...

    return nil
}
```

### Gin 集成

```go
package main

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func main() {
    // 初始化追踪器
    tp, _ := tracing.InitTracer("order-service", "http://jaeger:14268/api/traces")
    defer tracing.Shutdown(tp)

    r := gin.New()
    r.Use(gin.Recovery())

    // 添加 OpenTelemetry 中间件
    r.Use(otelgin.Middleware("order-service"))

    r.GET("/api/orders/:id", GetOrder)
    r.POST("/api/orders", CreateOrder)

    r.Run(":8080")
}

func GetOrder(c *gin.Context) {
    ctx := c.Request.Context()
    orderID := c.Param("id")

    tracer := otel.Tracer("order-handler")
    ctx, span := tracer.Start(ctx, "GetOrder")
    defer span.End()

    span.SetAttributes(attribute.String("order.id", orderID))

    // 获取订单
    order, err := orderService.GetOrder(ctx, orderID)
    if err != nil {
        span.RecordError(err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, order)
}
```

### 数据库追踪

```go
package database

import (
    "context"
    "database/sql"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var tracer = otel.Tracer("database")

// TracedDB 带追踪的数据库包装
type TracedDB struct {
    db *sql.DB
}

func (t *TracedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    ctx, span := tracer.Start(ctx, "db.query",
        trace.WithSpanKind(trace.SpanKindClient),
    )
    defer span.End()

    span.SetAttributes(
        semconv.DBSystemPostgreSQL,
        semconv.DBStatement(query),
    )

    rows, err := t.db.QueryContext(ctx, query, args...)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }

    return rows, err
}

func (t *TracedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    ctx, span := tracer.Start(ctx, "db.exec",
        trace.WithSpanKind(trace.SpanKindClient),
    )
    defer span.End()

    span.SetAttributes(
        semconv.DBSystemPostgreSQL,
        semconv.DBStatement(query),
        semconv.DBOperation(extractOperation(query)),
    )

    result, err := t.db.ExecContext(ctx, query, args...)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }

    return result, err
}

func extractOperation(query string) string {
    // 简单提取 SQL 操作类型
    words := strings.Fields(query)
    if len(words) > 0 {
        return strings.ToUpper(words[0])
    }
    return "UNKNOWN"
}
```

### HTTP 客户端追踪

```go
package httpclient

import (
    "context"
    "net/http"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewTracedClient 创建带追踪的 HTTP 客户端
func NewTracedClient() *http.Client {
    return &http.Client{
        Transport: otelhttp.NewTransport(http.DefaultTransport),
    }
}

// 使用示例
func CallExternalService(ctx context.Context, url string) (*http.Response, error) {
    client := NewTracedClient()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    return client.Do(req)
}
```

### 服务间上下文传播

```go
// 发送方：注入上下文到 HTTP 头
func CallService(ctx context.Context, url string) (*http.Response, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

    // 注入追踪上下文到 HTTP 头
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    return http.DefaultClient.Do(req)
}

// 接收方：从 HTTP 头提取上下文
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // 从 HTTP 头提取追踪上下文
    ctx := otel.GetTextMapPropagator().Extract(
        r.Context(),
        propagation.HeaderCarrier(r.Header),
    )

    // 使用提取的上下文创建新 Span
    ctx, span := tracer.Start(ctx, "HandleRequest")
    defer span.End()

    // 处理请求...
}
```

## 采样策略

```go
package tracing

import (
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// 不同的采样策略

// 1. 全部采样（开发环境）
sampler := sdktrace.AlwaysSample()

// 2. 不采样
sampler := sdktrace.NeverSample()

// 3. 比例采样（生产环境，采样 10%）
sampler := sdktrace.TraceIDRatioBased(0.1)

// 4. 父级决定采样
sampler := sdktrace.ParentBased(
    sdktrace.TraceIDRatioBased(0.1),
)

// 创建带采样器的 TracerProvider
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sampler),
    // ...
)
```

## 与 Node.js 对比

### Node.js (OpenTelemetry)

```javascript
const { NodeTracerProvider } = require('@opentelemetry/sdk-trace-node');
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { trace } = require('@opentelemetry/api');

// 初始化
const provider = new NodeTracerProvider();
provider.addSpanProcessor(
    new SimpleSpanProcessor(new JaegerExporter())
);
provider.register();

const tracer = trace.getTracer('my-service');

// 创建 Span
async function processOrder(orderId) {
    const span = tracer.startSpan('processOrder');
    span.setAttribute('order.id', orderId);

    try {
        await validateOrder(orderId);
        span.setStatus({ code: SpanStatusCode.OK });
    } catch (error) {
        span.recordException(error);
        span.setStatus({ code: SpanStatusCode.ERROR });
    } finally {
        span.end();
    }
}
```

### Go (OpenTelemetry)

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

tracer := otel.Tracer("my-service")

func processOrder(ctx context.Context, orderID string) error {
    ctx, span := tracer.Start(ctx, "processOrder")
    defer span.End()

    span.SetAttributes(attribute.String("order.id", orderID))

    if err := validateOrder(ctx, orderID); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "")
    return nil
}
```

## 最佳实践

### 1. 有意义的 Span 名称

```go
// ✅ 好的 Span 名称
tracer.Start(ctx, "OrderService.CreateOrder")
tracer.Start(ctx, "db.query.users")
tracer.Start(ctx, "http.client.payment-api")

// ❌ 不好的 Span 名称
tracer.Start(ctx, "process")
tracer.Start(ctx, "query")
tracer.Start(ctx, "call")
```

### 2. 适当的属性

```go
// 添加有用的属性
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.Int("order.item_count", len(items)),
    attribute.Float64("order.total", total),
)

// 避免高基数属性
// ❌ 不好：每个请求都不同的值
span.SetAttributes(attribute.String("request.body", body))
```

### 3. 错误处理

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}

// 业务错误也应该记录
if user == nil {
    span.AddEvent("user_not_found",
        trace.WithAttributes(attribute.String("user.id", userID)),
    )
    span.SetStatus(codes.Error, "user not found")
}
```

**下一节**：[Jaeger](./09-jaeger.md) - 学习 Jaeger 追踪系统集成
