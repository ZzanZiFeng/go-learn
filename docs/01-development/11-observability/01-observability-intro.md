# 可观测性概述 (Observability Introduction)

## 什么是可观测性

可观测性（Observability）是指通过系统的外部输出来推断系统内部状态的能力。

```
┌──────────────────────────────────────────────────────────────────┐
│                     可观测性 vs 监控                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  监控 (Monitoring):                                              │
│  • 回答已知问题："CPU 使用率是多少？"                              │
│  • 基于预定义的指标                                               │
│  • 被动：等待问题发生                                             │
│                                                                  │
│  可观测性 (Observability):                                       │
│  • 回答未知问题："为什么这个请求这么慢？"                          │
│  • 基于任意维度的查询                                             │
│  • 主动：能够探索和调试                                           │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 三大支柱

### 1. 日志 (Logs)

日志是系统发生事件的离散记录。

```go
// 传统日志
log.Printf("[INFO] User %d logged in from %s", userID, ip)

// 结构化日志
logger.Info("User logged in",
    zap.Int("user_id", userID),
    zap.String("ip", ip),
    zap.String("user_agent", userAgent),
)
```

**特点**：
- 高基数：每个事件都是唯一的
- 易于调试：包含详细上下文
- 存储成本高：数据量大

**适用场景**：
- 错误调试
- 审计追踪
- 业务事件记录

### 2. 指标 (Metrics)

指标是系统状态的数值测量。

```go
// Counter: 累计值，只增不减
httpRequestsTotal.Inc()

// Gauge: 可增可减的瞬时值
activeConnections.Set(100)

// Histogram: 分布统计
requestDuration.Observe(0.25) // 250ms
```

**特点**：
- 低基数：聚合的数值
- 高效查询：时间序列数据库优化
- 存储成本低：压缩效率高

**适用场景**：
- 系统健康监控
- 趋势分析
- 容量规划
- 告警

### 3. 追踪 (Traces)

追踪是请求在分布式系统中的完整路径。

```
Trace ID: abc123
├── Span: API Gateway (10ms)
│   └── Span: Auth Service (5ms)
├── Span: Order Service (50ms)
│   ├── Span: Database Query (20ms)
│   └── Span: Cache Lookup (2ms)
└── Span: Payment Service (100ms)
    └── Span: External API (80ms)
```

**特点**：
- 因果关系：展示请求流程
- 延迟分析：识别瓶颈
- 采样：减少数据量

**适用场景**：
- 性能调优
- 故障定位
- 依赖分析

## 三者关系

```
┌──────────────────────────────────────────────────────────────────┐
│                     可观测性数据关联                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│              ┌─────────────────────────────────┐                │
│              │         Traces                  │                │
│              │    (请求级别，因果关系)          │                │
│              │                                 │                │
│              │   trace_id: abc123              │                │
│              │   span_id: def456               │                │
│              └─────────────────────────────────┘                │
│                           │                                      │
│                           │ 关联                                 │
│                           │                                      │
│  ┌───────────────────────┴───────────────────────┐              │
│  │                                               │              │
│  ▼                                               ▼              │
│  ┌─────────────────────┐      ┌─────────────────────┐          │
│  │       Logs          │      │      Metrics        │          │
│  │   (事件级别)         │      │   (聚合级别)        │          │
│  │                     │      │                     │          │
│  │ trace_id: abc123    │      │ http_requests_total │          │
│  │ message: "Error..." │      │ {status="500"}      │          │
│  └─────────────────────┘      └─────────────────────┘          │
│                                                                  │
│  通过 trace_id 关联:                                             │
│  • 从指标发现异常                                                 │
│  • 查找相关追踪                                                   │
│  • 定位具体日志                                                   │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 常用工具

### 日志工具

| 工具 | 说明 | Go 库 |
|------|------|-------|
| Zap | Uber 高性能日志 | `go.uber.org/zap` |
| Zerolog | 零分配 JSON 日志 | `github.com/rs/zerolog` |
| Logrus | 结构化日志 | `github.com/sirupsen/logrus` |
| ELK | 日志聚合平台 | Filebeat/Logstash |
| Loki | Grafana 日志聚合 | Promtail |

### 指标工具

| 工具 | 说明 | Go 库 |
|------|------|-------|
| Prometheus | 指标收集和存储 | `github.com/prometheus/client_golang` |
| Grafana | 可视化仪表板 | - |
| InfluxDB | 时序数据库 | `github.com/influxdata/influxdb-client-go` |
| StatsD | 指标聚合 | `github.com/DataDog/datadog-go` |

### 追踪工具

| 工具 | 说明 | Go 库 |
|------|------|-------|
| Jaeger | 分布式追踪 | `go.opentelemetry.io/otel` |
| Zipkin | 分布式追踪 | `github.com/openzipkin/zipkin-go` |
| OpenTelemetry | 统一可观测性 | `go.opentelemetry.io/otel` |

## OpenTelemetry

OpenTelemetry 是可观测性的统一标准，整合了日志、指标、追踪。

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/metric"
)

// 创建追踪器
tracer := otel.Tracer("my-service")

// 创建 Span
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// 添加属性
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.Int("order.amount", amount),
)

// 记录事件
span.AddEvent("cache-hit", trace.WithAttributes(
    attribute.String("key", cacheKey),
))

// 创建指标
meter := otel.Meter("my-service")
counter, _ := meter.Int64Counter("requests_total")
counter.Add(ctx, 1)
```

## 最佳实践

### 1. 日志最佳实践

```go
// 使用结构化日志
logger.Info("Order created",
    zap.String("order_id", orderID),
    zap.Int("user_id", userID),
    zap.Float64("amount", amount),
)

// 包含请求上下文
func handler(c *gin.Context) {
    logger := logger.With(
        zap.String("request_id", c.GetString("request_id")),
        zap.String("user_id", c.GetString("user_id")),
    )
    // ...
}

// 分级日志
logger.Debug("Detailed info for debugging")
logger.Info("Important business events")
logger.Warn("Potential issues")
logger.Error("Errors that need attention")
```

### 2. 指标最佳实践

```go
// 使用有意义的标签
httpRequestsTotal.WithLabelValues(
    method,    // GET, POST, etc.
    path,      // /api/users, /api/orders
    status,    // 200, 400, 500
)

// 避免高基数标签
// 不好：user_id, request_id
// 好：user_type, region

// 遵循命名规范
// <namespace>_<name>_<unit>
// http_requests_total
// http_request_duration_seconds
// db_connections_active
```

### 3. 追踪最佳实践

```go
// 创建有意义的 Span 名称
ctx, span := tracer.Start(ctx, "db.query.users")

// 记录重要属性
span.SetAttributes(
    attribute.String("db.system", "postgresql"),
    attribute.String("db.statement", query),
    attribute.Int("db.rows_affected", rows),
)

// 记录错误
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}

// 传递上下文
result, err := service.Process(ctx, data) // ctx 包含追踪信息
```

## Docker Compose 示例

```yaml
version: '3.8'

services:
  # 应用服务
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - JAEGER_AGENT_HOST=jaeger
      - PROMETHEUS_PUSHGATEWAY=prometheus:9091

  # Prometheus
  prometheus:
    image: prom/prometheus:v2.45.0
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  # Grafana
  grafana:
    image: grafana/grafana:10.0.0
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

  # Jaeger
  jaeger:
    image: jaegertracing/all-in-one:1.47
    ports:
      - "16686:16686"  # UI
      - "6831:6831/udp"  # Agent

  # Loki (日志)
  loki:
    image: grafana/loki:2.8.0
    ports:
      - "3100:3100"
```

## 总结

| 支柱 | 问题 | 工具 | 成本 |
|------|------|------|------|
| 日志 | 发生了什么？ | Zap + ELK/Loki | 高 |
| 指标 | 系统状态如何？ | Prometheus + Grafana | 低 |
| 追踪 | 请求路径是什么？ | Jaeger/OpenTelemetry | 中 |

**下一节**：[结构化日志](./02-structured-logging.md) - 学习 Zap 日志库
