# Jaeger 分布式追踪

## 概述

Jaeger 是 Uber 开源的分布式追踪系统，现为 CNCF 毕业项目，用于监控和故障排查微服务架构。

```
┌──────────────────────────────────────────────────────────────────┐
│                       Jaeger 架构                                 │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐            │
│  │Service A│  │Service B│  │Service C│  │Service D│            │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘            │
│       │            │            │            │                   │
│       └────────────┴────────────┴────────────┘                   │
│                          │                                       │
│                          ▼                                       │
│                 ┌─────────────────┐                             │
│                 │  Jaeger Agent   │  (可选，边车模式)            │
│                 └────────┬────────┘                             │
│                          │                                       │
│                          ▼                                       │
│                 ┌─────────────────┐                             │
│                 │Jaeger Collector │  (接收和处理追踪数据)        │
│                 └────────┬────────┘                             │
│                          │                                       │
│                          ▼                                       │
│                 ┌─────────────────┐                             │
│                 │    Storage      │  (Cassandra/ES/Memory)      │
│                 └────────┬────────┘                             │
│                          │                                       │
│                          ▼                                       │
│                 ┌─────────────────┐                             │
│                 │   Jaeger UI     │  (可视化界面)                │
│                 │  :16686         │                             │
│                 └─────────────────┘                             │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 快速开始

### Docker 部署

```bash
# All-in-One 模式（开发环境）
docker run -d --name jaeger \
  -p 6831:6831/udp \
  -p 16686:16686 \
  -p 14268:14268 \
  jaegertracing/all-in-one:1.50

# 端口说明:
# 6831 - Jaeger Agent (UDP)
# 16686 - Jaeger UI
# 14268 - Collector HTTP
```

### Docker Compose

```yaml
version: '3.8'

services:
  jaeger:
    image: jaegertracing/all-in-one:1.50
    ports:
      - "6831:6831/udp"   # Agent - Thrift compact
      - "6832:6832/udp"   # Agent - Thrift binary
      - "5778:5778"       # Agent - configs
      - "16686:16686"     # UI
      - "14268:14268"     # Collector HTTP
      - "14250:14250"     # Collector gRPC
      - "4317:4317"       # OTLP gRPC
      - "4318:4318"       # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - tracing

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4317
    depends_on:
      - jaeger
    networks:
      - tracing

networks:
  tracing:
    driver: bridge
```

## Go 集成

### 安装依赖

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```

### 初始化追踪器

```go
package tracing

import (
    "context"
    "log"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// Config 追踪配置
type Config struct {
    ServiceName    string
    ServiceVersion string
    Environment    string
    JaegerEndpoint string // e.g., "jaeger:4317"
    SampleRate     float64
}

// InitTracer 初始化追踪器
func InitTracer(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
    // 创建 gRPC 连接
    conn, err := grpc.DialContext(ctx, cfg.JaegerEndpoint,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        return nil, err
    }

    // 创建 OTLP Exporter
    exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
    if err != nil {
        return nil, err
    }

    // 创建资源
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(cfg.ServiceName),
            semconv.ServiceVersion(cfg.ServiceVersion),
            semconv.DeploymentEnvironment(cfg.Environment),
        ),
        resource.WithHost(),
        resource.WithProcess(),
    )
    if err != nil {
        return nil, err
    }

    // 选择采样器
    var sampler sdktrace.Sampler
    if cfg.SampleRate >= 1.0 {
        sampler = sdktrace.AlwaysSample()
    } else if cfg.SampleRate <= 0 {
        sampler = sdktrace.NeverSample()
    } else {
        sampler = sdktrace.ParentBased(
            sdktrace.TraceIDRatioBased(cfg.SampleRate),
        )
    }

    // 创建 TracerProvider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sampler),
    )

    // 设置全局 TracerProvider 和 Propagator
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    log.Printf("Tracing initialized: service=%s, endpoint=%s, sample_rate=%.2f",
        cfg.ServiceName, cfg.JaegerEndpoint, cfg.SampleRate)

    return tp, nil
}

// Shutdown 关闭追踪器
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
    return tp.Shutdown(ctx)
}
```

### HTTP 使用 (Jaeger HTTP Exporter)

```go
package tracing

import (
    "context"

    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracerHTTP 使用 HTTP Exporter
func InitTracerHTTP(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
    // HTTP Exporter
    exporter, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint(cfg.JaegerEndpoint),
        otlptracehttp.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    // ... 其余配置与 gRPC 相同
}
```

## 完整示例应用

### 主程序

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "myapp/tracing"
)

var tracer = otel.Tracer("myapp")

func main() {
    ctx := context.Background()

    // 初始化追踪
    tp, err := tracing.InitTracer(ctx, tracing.Config{
        ServiceName:    "order-service",
        ServiceVersion: "1.0.0",
        Environment:    "development",
        JaegerEndpoint: "localhost:4317",
        SampleRate:     1.0,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 优雅关闭
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := tracing.Shutdown(ctx, tp); err != nil {
            log.Printf("Error shutting down tracer: %v", err)
        }
    }()

    // 创建路由
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(otelgin.Middleware("order-service"))

    // 路由
    r.GET("/api/orders/:id", getOrder)
    r.POST("/api/orders", createOrder)
    r.GET("/health", healthCheck)

    // 启动服务器
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}

func getOrder(c *gin.Context) {
    ctx := c.Request.Context()
    orderID := c.Param("id")

    ctx, span := tracer.Start(ctx, "getOrder")
    defer span.End()

    span.SetAttributes(attribute.String("order.id", orderID))

    // 模拟获取订单
    order, err := fetchOrderFromDB(ctx, orderID)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 获取用户信息
    user, err := fetchUserFromService(ctx, order.UserID)
    if err != nil {
        span.AddEvent("failed_to_fetch_user",
            trace.WithAttributes(attribute.String("user.id", order.UserID)),
        )
        // 不影响主流程，只记录
    }

    span.SetStatus(codes.Ok, "")
    c.JSON(http.StatusOK, gin.H{
        "order": order,
        "user":  user,
    })
}

func createOrder(c *gin.Context) {
    ctx := c.Request.Context()

    ctx, span := tracer.Start(ctx, "createOrder")
    defer span.End()

    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "invalid request")
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    span.SetAttributes(
        attribute.String("user.id", req.UserID),
        attribute.Int("order.item_count", len(req.Items)),
    )

    // 验证用户
    if err := validateUser(ctx, req.UserID); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 检查库存
    if err := checkInventory(ctx, req.Items); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 创建订单
    order, err := saveOrder(ctx, req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    span.SetAttributes(attribute.String("order.id", order.ID))
    span.SetStatus(codes.Ok, "")
    c.JSON(http.StatusCreated, order)
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
    "go.opentelemetry.io/otel/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var tracer = otel.Tracer("database")

type TracedDB struct {
    db       *sql.DB
    dbSystem string
    dbName   string
}

func NewTracedDB(db *sql.DB, dbSystem, dbName string) *TracedDB {
    return &TracedDB{
        db:       db,
        dbSystem: dbSystem,
        dbName:   dbName,
    }
}

func (t *TracedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    ctx, span := tracer.Start(ctx, "db.query",
        trace.WithSpanKind(trace.SpanKindClient),
    )
    defer span.End()

    t.addDBAttributes(span, query)
    return t.db.QueryRowContext(ctx, query, args...)
}

func (t *TracedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    ctx, span := tracer.Start(ctx, "db.query",
        trace.WithSpanKind(trace.SpanKindClient),
    )
    defer span.End()

    t.addDBAttributes(span, query)

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

    t.addDBAttributes(span, query)

    result, err := t.db.ExecContext(ctx, query, args...)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
    return result, err
}

func (t *TracedDB) addDBAttributes(span trace.Span, query string) {
    span.SetAttributes(
        semconv.DBSystemKey.String(t.dbSystem),
        semconv.DBName(t.dbName),
        semconv.DBStatement(query),
    )
}
```

### HTTP 客户端追踪

```go
package httpclient

import (
    "context"
    "net/http"
    "time"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("httpclient")

// Client 带追踪的 HTTP 客户端
type Client struct {
    client  *http.Client
    baseURL string
}

func NewClient(baseURL string, timeout time.Duration) *Client {
    return &Client{
        client: &http.Client{
            Timeout:   timeout,
            Transport: otelhttp.NewTransport(http.DefaultTransport),
        },
        baseURL: baseURL,
    }
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
    ctx, span := tracer.Start(ctx, "http.client.get",
        trace.WithSpanKind(trace.SpanKindClient),
    )
    defer span.End()

    url := c.baseURL + path
    span.SetAttributes(
        attribute.String("http.method", "GET"),
        attribute.String("http.url", url),
    )

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    resp, err := c.client.Do(req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
    if resp.StatusCode >= 400 {
        span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
    }

    return resp, nil
}
```

### Redis 追踪

```go
package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("redis")

type TracedRedis struct {
    client *redis.Client
}

func NewTracedRedis(client *redis.Client) *TracedRedis {
    return &TracedRedis{client: client}
}

func (r *TracedRedis) Get(ctx context.Context, key string) (string, error) {
    ctx, span := tracer.Start(ctx, "redis.get",
        trace.WithSpanKind(trace.SpanKindClient),
    )
    defer span.End()

    span.SetAttributes(
        attribute.String("db.system", "redis"),
        attribute.String("db.operation", "GET"),
        attribute.String("db.redis.key", key),
    )

    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        span.AddEvent("cache_miss")
        return "", nil
    }
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return "", err
    }

    span.AddEvent("cache_hit")
    return val, nil
}

func (r *TracedRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    ctx, span := tracer.Start(ctx, "redis.set",
        trace.WithSpanKind(trace.SpanKindClient),
    )
    defer span.End()

    span.SetAttributes(
        attribute.String("db.system", "redis"),
        attribute.String("db.operation", "SET"),
        attribute.String("db.redis.key", key),
        attribute.Int64("db.redis.ttl_seconds", int64(expiration.Seconds())),
    )

    err := r.client.Set(ctx, key, value, expiration).Err()
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }

    return err
}
```

## Jaeger UI 使用

### 搜索追踪

访问 http://localhost:16686 打开 Jaeger UI：

1. **服务选择**：从下拉列表选择服务
2. **操作过滤**：选择特定操作
3. **标签过滤**：添加标签条件，如 `error=true`
4. **时间范围**：设置查询时间范围
5. **限制数量**：设置返回追踪数量

### 追踪详情

追踪详情页面显示：
- **时间线视图**：所有 Span 的时间关系
- **Span 详情**：标签、日志、持续时间
- **依赖关系**：服务间调用关系
- **关键路径**：识别性能瓶颈

### 比较追踪

选择两个追踪进行比较，找出性能差异。

## 与日志关联

### 将 Trace ID 添加到日志

```go
package logger

import (
    "context"

    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
)

// WithTraceContext 添加追踪上下文到日志
func WithTraceContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
    span := trace.SpanFromContext(ctx)
    if !span.SpanContext().IsValid() {
        return logger
    }

    return logger.With(
        zap.String("trace_id", span.SpanContext().TraceID().String()),
        zap.String("span_id", span.SpanContext().SpanID().String()),
    )
}

// 使用示例
func ProcessRequest(ctx context.Context) {
    logger := WithTraceContext(ctx, baseLogger)

    logger.Info("Processing request",
        zap.String("user_id", "123"),
    )
    // 输出: {"level":"info","msg":"Processing request","trace_id":"abc...","span_id":"def...","user_id":"123"}
}
```

## 与 Prometheus 指标关联

### Exemplar 支持

```go
package metrics

import (
    "context"

    "github.com/prometheus/client_golang/prometheus"
    "go.opentelemetry.io/otel/trace"
)

// 创建支持 Exemplar 的 Histogram
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration",
        Buckets: prometheus.DefBuckets,
    },
    []string{"method", "path"},
)

// 记录带 Exemplar 的指标
func ObserveWithExemplar(ctx context.Context, method, path string, duration float64) {
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        requestDuration.WithLabelValues(method, path).(prometheus.ExemplarObserver).ObserveWithExemplar(
            duration,
            prometheus.Labels{"trace_id": span.SpanContext().TraceID().String()},
        )
    } else {
        requestDuration.WithLabelValues(method, path).Observe(duration)
    }
}
```

## 最佳实践

### 1. Span 命名

```go
// ✅ 好的命名：清晰、有层次
"OrderService.CreateOrder"
"db.query.orders"
"http.client.payment-service"
"cache.get.user"

// ❌ 不好的命名：模糊、无信息
"process"
"query"
"call"
```

### 2. 属性使用

```go
// 使用语义化属性
import semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

span.SetAttributes(
    semconv.HTTPMethod("GET"),
    semconv.HTTPRoute("/api/users/:id"),
    semconv.HTTPStatusCode(200),
    semconv.DBStatement("SELECT * FROM users"),
    semconv.DBSystemPostgreSQL,
)
```

### 3. 错误处理

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    span.SetAttributes(attribute.String("error.type", getErrorType(err)))
}
```

### 4. 采样策略

```go
// 开发环境：全部采样
sampler := sdktrace.AlwaysSample()

// 生产环境：按比例采样
sampler := sdktrace.ParentBased(
    sdktrace.TraceIDRatioBased(0.1), // 10% 采样
)

// 高优先级请求全部采样
sampler := sdktrace.ParentBased(
    sdktrace.TraceIDRatioBased(0.1),
    sdktrace.WithLocalParentSampled(sdktrace.AlwaysSample()),
)
```

**下一节**：[ELK](./10-elk.md) - 学习 ELK 日志聚合
