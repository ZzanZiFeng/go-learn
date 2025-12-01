# 可观测性练习

## 练习 1：结构化日志

### 目标
实现一个带请求上下文的结构化日志系统。

### 要求
1. 使用 Zap 创建全局日志器
2. 实现请求 ID 中间件
3. 将请求上下文（request_id, user_id, path）添加到所有日志
4. 实现敏感字段脱敏

### 预期输出
```json
{
  "ts": "2024-01-01T10:00:00.000Z",
  "level": "info",
  "msg": "User login",
  "caller": "handler/auth.go:42",
  "request_id": "req-abc123",
  "user_id": "user-456",
  "path": "/api/login",
  "password": "pa****rd"
}
```

### 提示
```go
// 中间件将 logger 添加到上下文
func LoggerMiddleware(baseLogger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        logger := baseLogger.With(
            zap.String("request_id", requestID),
            zap.String("path", c.Request.URL.Path),
        )
        c.Set("logger", logger)
        c.Next()
    }
}
```

---

## 练习 2：Prometheus 指标

### 目标
为 HTTP API 添加 Prometheus 指标。

### 要求
1. 实现 HTTP 请求计数器（按 method, path, status 分组）
2. 实现 HTTP 请求延迟直方图
3. 实现活跃连接数 Gauge
4. 实现业务指标（如订单数量、订单金额）

### 预期指标
```promql
# 查询 QPS
rate(myapp_http_requests_total[5m])

# 查询 P95 延迟
histogram_quantile(0.95, rate(myapp_http_request_duration_seconds_bucket[5m]))

# 查询错误率
sum(rate(myapp_http_requests_total{status=~"5.."}[5m])) / sum(rate(myapp_http_requests_total[5m]))
```

### 提示
```go
var httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Namespace: "myapp",
        Name:      "http_requests_total",
        Help:      "Total HTTP requests",
    },
    []string{"method", "path", "status"},
)
```

---

## 练习 3：分布式追踪

### 目标
使用 OpenTelemetry 实现跨服务追踪。

### 要求
1. 初始化 Jaeger Exporter
2. 创建 HTTP 服务追踪中间件
3. 实现数据库操作追踪
4. 实现 HTTP 客户端追踪
5. 在服务间传播追踪上下文

### 预期追踪
```
Trace ID: abc123
├── Span: HTTP GET /api/orders/123 (200ms)
│   ├── Span: db.query (50ms)
│   └── Span: http.client.user-service (100ms)
│       └── Span: HTTP GET /api/users/456 (80ms)
│           └── Span: db.query (30ms)
```

### 提示
```go
// 初始化 tracer
tracer := otel.Tracer("order-service")

// 创建 span
ctx, span := tracer.Start(ctx, "getOrder")
defer span.End()

// 添加属性
span.SetAttributes(attribute.String("order.id", orderID))
```

---

## 练习 4：告警规则

### 目标
配置 Prometheus 告警规则。

### 要求
编写以下告警规则：
1. 服务下线告警
2. 错误率超过 5% 告警
3. P95 延迟超过 1 秒告警
4. 内存使用超过 80% 告警
5. 数据库连接池饱和告警

### 预期规则
```yaml
groups:
  - name: myapp
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(myapp_http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(myapp_http_requests_total[5m]))
          > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "错误率过高: {{ $value | humanizePercentage }}"
```

---

## 练习 5：Grafana 仪表板

### 目标
创建 Go 服务监控仪表板。

### 要求
创建以下面板：
1. 服务健康状态（Stat）
2. QPS 趋势（Time Series）
3. 错误率趋势（Time Series）
4. P50/P90/P99 延迟（Time Series）
5. Goroutine 数量（Time Series）
6. 内存使用（Gauge）
7. 请求分布（Pie Chart）

### 预期仪表板结构
```
┌─────────┬─────────┬─────────┬─────────┐
│ Health  │   QPS   │ Error % │  P95    │
├─────────┴─────────┴─────────┴─────────┤
│           QPS by Path (Graph)          │
├────────────────────────────────────────┤
│         Latency Percentiles (Graph)    │
├───────────────────┬────────────────────┤
│   Goroutines      │   Memory Usage     │
└───────────────────┴────────────────────┘
```

---

## 练习 6：日志与追踪关联

### 目标
将日志和追踪 ID 关联起来。

### 要求
1. 在日志中添加 trace_id 和 span_id
2. 实现从 Grafana 跳转到 Jaeger 的链接
3. 在 Jaeger 中能够查看相关日志

### 预期日志
```json
{
  "ts": "2024-01-01T10:00:00.000Z",
  "level": "info",
  "msg": "Processing order",
  "trace_id": "abc123def456ghi789",
  "span_id": "jkl012",
  "order_id": "order-001"
}
```

### 提示
```go
func WithTraceContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        return logger.With(
            zap.String("trace_id", span.SpanContext().TraceID().String()),
            zap.String("span_id", span.SpanContext().SpanID().String()),
        )
    }
    return logger
}
```

---

## 综合项目：可观测的订单服务

### 目标
构建一个完整的可观测订单服务。

### 系统架构
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│ Order API   │────▶│   Postgres  │
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                          ├──────────▶ Redis (缓存)
                          │
                          └──────────▶ User Service (HTTP)
```

### 要求

#### 1. 日志
- 使用 Zap 结构化日志
- 所有日志包含 request_id 和 trace_id
- 实现日志级别动态调整
- 敏感字段脱敏

#### 2. 指标
```go
// HTTP 指标
http_requests_total{method, path, status}
http_request_duration_seconds{method, path}

// 业务指标
orders_created_total{status}
orders_amount_dollars{status}
order_processing_duration_seconds

// 数据库指标
db_query_duration_seconds{operation, table}
db_connections_active

// 缓存指标
cache_operations_total{operation, result}
cache_latency_seconds{operation}
```

#### 3. 追踪
- 所有 HTTP 请求创建 Span
- 数据库操作创建子 Span
- HTTP 客户端调用创建子 Span
- 缓存操作创建子 Span

#### 4. 告警
```yaml
# 必须实现的告警
- alert: OrderServiceDown
- alert: HighOrderErrorRate
- alert: SlowOrderProcessing
- alert: DatabaseConnectionPoolExhausted
- alert: HighCacheMissRate
```

#### 5. 仪表板
- 服务概览仪表板
- 订单业务仪表板
- 数据库性能仪表板

### 文件结构
```
order-service/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   └── order.go
│   ├── services/
│   │   └── order.go
│   ├── repositories/
│   │   └── order.go
│   ├── middleware/
│   │   ├── logger.go
│   │   ├── metrics.go
│   │   └── tracing.go
│   └── observability/
│       ├── logger.go
│       ├── metrics.go
│       └── tracing.go
├── docker-compose.yml
├── prometheus.yml
├── alertmanager.yml
└── grafana/
    └── dashboards/
        └── order-service.json
```

### 验收标准

1. **日志**
   - [ ] 所有请求有 request_id
   - [ ] 所有日志有 trace_id
   - [ ] 错误日志包含堆栈
   - [ ] 敏感字段已脱敏

2. **指标**
   - [ ] /metrics 端点可访问
   - [ ] QPS 指标正确
   - [ ] 延迟分位数正确
   - [ ] 业务指标正确

3. **追踪**
   - [ ] Jaeger 可看到完整追踪
   - [ ] 数据库操作有独立 Span
   - [ ] 外部调用有独立 Span

4. **告警**
   - [ ] 告警规则已配置
   - [ ] 告警可正常触发
   - [ ] 通知渠道已配置

5. **仪表板**
   - [ ] Grafana 仪表板已创建
   - [ ] 图表数据正确
   - [ ] 可以从指标跳转到追踪

---

## 参考答案

### 练习 1 参考

```go
// internal/observability/logger.go
package observability

import (
    "context"
    "strings"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey contextKey = "logger"

var sensitiveFields = []string{"password", "token", "secret", "card"}

func NewLogger(level string, dev bool) (*zap.Logger, error) {
    var config zap.Config
    if dev {
        config = zap.NewDevelopmentConfig()
    } else {
        config = zap.NewProductionConfig()
    }

    var lvl zapcore.Level
    lvl.UnmarshalText([]byte(level))
    config.Level = zap.NewAtomicLevelAt(lvl)

    return config.Build()
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
    if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
        return logger
    }
    return zap.NewNop()
}

func SensitiveField(key, value string) zap.Field {
    for _, sensitive := range sensitiveFields {
        if strings.Contains(strings.ToLower(key), sensitive) {
            return zap.String(key, maskValue(value))
        }
    }
    return zap.String(key, value)
}

func maskValue(value string) string {
    if len(value) <= 4 {
        return "****"
    }
    return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
```

### 练习 2 参考

```go
// internal/observability/metrics.go
package observability

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "myapp",
            Name:      "http_requests_total",
            Help:      "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Name:      "http_request_duration_seconds",
            Help:      "HTTP request duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
        []string{"method", "path"},
    )

    ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "myapp",
        Name:      "active_connections",
        Help:      "Number of active connections",
    })

    OrdersCreatedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "myapp",
            Name:      "orders_created_total",
            Help:      "Total orders created",
        },
        []string{"status"},
    )

    OrderAmount = promauto.NewHistogram(prometheus.HistogramOpts{
        Namespace: "myapp",
        Name:      "order_amount_dollars",
        Help:      "Order amount distribution",
        Buckets:   []float64{10, 50, 100, 250, 500, 1000, 2500, 5000},
    })
)
```

### 练习 3 参考

```go
// internal/observability/tracing.go
package observability

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(ctx context.Context, serviceName, jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(jaegerEndpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    res, _ := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
        ),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp, nil
}
```

---

## 下一步

完成本章练习后，你可以：

1. 继续学习 **轨道二：实战教程**，将可观测性应用到实际项目
2. 学习 **轨道三：部署教程**，了解生产环境监控配置
3. 学习 **轨道四：测试教程**，了解可观测性相关的测试方法

**返回**：[可观测性概述](./README.md)
