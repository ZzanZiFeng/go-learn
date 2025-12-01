# 自定义指标 (Custom Metrics)

## 概述

自定义指标让你能够监控业务特定的数据，而不仅仅是技术指标。

```
┌──────────────────────────────────────────────────────────────────┐
│                       指标分类                                    │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 技术指标 (Infrastructure)                                    ││
│  │ • CPU、内存、磁盘使用率                                       ││
│  │ • HTTP 请求数、延迟、错误率                                   ││
│  │ • 数据库连接数、查询延迟                                      ││
│  │ • Go 运行时指标 (GC, Goroutines)                             ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 业务指标 (Business)                                          ││
│  │ • 用户注册数、活跃用户数                                      ││
│  │ • 订单数量、订单金额                                          ││
│  │ • 支付成功率、退款率                                          ││
│  │ • 库存水平、发货延迟                                          ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Counter 深入

### 基本使用

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 简单 Counter
var requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
    Namespace: "myapp",
    Subsystem: "http",
    Name:      "requests_total",
    Help:      "Total number of HTTP requests",
})

// 带标签的 Counter
var requestsByStatus = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Namespace: "myapp",
        Subsystem: "http",
        Name:      "requests_by_status_total",
        Help:      "Total number of HTTP requests by status",
    },
    []string{"method", "status"},
)

func RecordRequest(method string, status int) {
    requestsTotal.Inc()
    requestsByStatus.WithLabelValues(method, fmt.Sprintf("%d", status)).Inc()
}
```

### Counter 应用场景

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 错误计数
    ErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "myapp",
            Name:      "errors_total",
            Help:      "Total number of errors",
        },
        []string{"type", "code"},
    )

    // 邮件发送计数
    EmailsSentTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "myapp",
            Subsystem: "email",
            Name:      "sent_total",
            Help:      "Total emails sent",
        },
        []string{"template", "status"},
    )

    // 缓存操作计数
    CacheOperationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "myapp",
            Subsystem: "cache",
            Name:      "operations_total",
            Help:      "Total cache operations",
        },
        []string{"operation", "result"}, // get/set/delete, hit/miss/error
    )

    // 消息处理计数
    MessagesProcessedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "myapp",
            Subsystem: "queue",
            Name:      "messages_processed_total",
            Help:      "Total messages processed",
        },
        []string{"queue", "status"}, // success/failed/retried
    )
)

// 使用示例
func ProcessMessage(queue string, err error) {
    status := "success"
    if err != nil {
        status = "failed"
        ErrorsTotal.WithLabelValues("queue_processing", "QUEUE_ERROR").Inc()
    }
    MessagesProcessedTotal.WithLabelValues(queue, status).Inc()
}
```

### Counter 查询示例

```promql
# 每秒请求数 (QPS)
rate(myapp_http_requests_total[5m])

# 每秒错误数
rate(myapp_errors_total[5m])

# 错误率
rate(myapp_errors_total[5m]) / rate(myapp_http_requests_total[5m])

# 按状态分组的 QPS
sum by (status) (rate(myapp_http_requests_by_status_total[5m]))

# 缓存命中率
sum(rate(myapp_cache_operations_total{result="hit"}[5m])) / sum(rate(myapp_cache_operations_total{operation="get"}[5m]))
```

## Gauge 深入

### 基本使用

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 简单 Gauge
var activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
    Namespace: "myapp",
    Name:      "active_connections",
    Help:      "Number of active connections",
})

// 带标签的 Gauge
var queueLength = promauto.NewGaugeVec(
    prometheus.GaugeOpts{
        Namespace: "myapp",
        Subsystem: "queue",
        Name:      "length",
        Help:      "Current queue length",
    },
    []string{"queue_name"},
)

func UpdateQueueLength(queue string, length int) {
    queueLength.WithLabelValues(queue).Set(float64(length))
}
```

### Gauge 应用场景

```go
package metrics

import (
    "runtime"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 活跃用户数
    ActiveUsers = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "myapp",
        Name:      "active_users",
        Help:      "Number of currently active users",
    })

    // Goroutine 数量
    GoroutineCount = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "myapp",
        Name:      "goroutine_count",
        Help:      "Current number of goroutines",
    })

    // 内存使用
    MemoryUsage = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "myapp",
            Name:      "memory_bytes",
            Help:      "Memory usage in bytes",
        },
        []string{"type"}, // heap, stack, alloc
    )

    // 数据库连接池
    DBPoolStats = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "myapp",
            Subsystem: "db",
            Name:      "pool_connections",
            Help:      "Database connection pool stats",
        },
        []string{"state"}, // active, idle, waiting
    )

    // 库存水平
    InventoryLevel = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "myapp",
            Subsystem: "inventory",
            Name:      "level",
            Help:      "Current inventory level",
        },
        []string{"product_id", "warehouse"},
    )

    // 任务队列深度
    TaskQueueDepth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "myapp",
            Subsystem: "task",
            Name:      "queue_depth",
            Help:      "Number of tasks in queue",
        },
        []string{"priority"}, // high, normal, low
    )
)

// 定期更新运行时指标
func StartRuntimeMetricsCollector() {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            GoroutineCount.Set(float64(runtime.NumGoroutine()))

            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            MemoryUsage.WithLabelValues("heap").Set(float64(m.HeapAlloc))
            MemoryUsage.WithLabelValues("stack").Set(float64(m.StackInuse))
            MemoryUsage.WithLabelValues("alloc").Set(float64(m.Alloc))
        }
    }()
}

// 更新数据库连接池指标
func UpdateDBPoolStats(active, idle, waiting int) {
    DBPoolStats.WithLabelValues("active").Set(float64(active))
    DBPoolStats.WithLabelValues("idle").Set(float64(idle))
    DBPoolStats.WithLabelValues("waiting").Set(float64(waiting))
}
```

### Gauge 查询示例

```promql
# 当前活跃用户数
myapp_active_users

# 内存使用趋势
myapp_memory_bytes{type="heap"}

# 数据库连接池使用率
myapp_db_pool_connections{state="active"} / (myapp_db_pool_connections{state="active"} + myapp_db_pool_connections{state="idle"})

# 库存低于阈值的产品
myapp_inventory_level < 10

# 队列深度变化
delta(myapp_task_queue_depth[5m])
```

## Histogram 深入

### 基本使用

```go
package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP 请求延迟
var httpRequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Namespace: "myapp",
        Subsystem: "http",
        Name:      "request_duration_seconds",
        Help:      "HTTP request duration in seconds",
        Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    },
    []string{"method", "path"},
)

// 使用 Timer 记录延迟
func MeasureRequestDuration(method, path string, fn func()) {
    timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(method, path))
    defer timer.ObserveDuration()
    fn()
}

// 或手动记录
func RecordRequestDuration(method, path string, duration time.Duration) {
    httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}
```

### Histogram 应用场景

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 数据库查询延迟
    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Subsystem: "db",
            Name:      "query_duration_seconds",
            Help:      "Database query duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
        },
        []string{"operation", "table"},
    )

    // 外部 API 调用延迟
    ExternalAPILatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Subsystem: "external",
            Name:      "api_latency_seconds",
            Help:      "External API call latency",
            Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30},
        },
        []string{"service", "endpoint"},
    )

    // 请求/响应大小
    RequestSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Subsystem: "http",
            Name:      "request_size_bytes",
            Help:      "HTTP request size in bytes",
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to 1GB
        },
        []string{"method", "path"},
    )

    ResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Subsystem: "http",
            Name:      "response_size_bytes",
            Help:      "HTTP response size in bytes",
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "path"},
    )

    // 订单金额分布
    OrderAmount = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Subsystem: "order",
            Name:      "amount_dollars",
            Help:      "Order amount distribution",
            Buckets:   []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
        },
        []string{"category"},
    )

    // 批处理任务耗时
    BatchJobDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Subsystem: "batch",
            Name:      "job_duration_seconds",
            Help:      "Batch job duration",
            Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600},
        },
        []string{"job_name"},
    )
)

// 数据库操作包装器
func MeasureDBQuery(operation, table string, fn func() error) error {
    timer := prometheus.NewTimer(DBQueryDuration.WithLabelValues(operation, table))
    defer timer.ObserveDuration()
    return fn()
}
```

### Histogram 查询示例

```promql
# P50 延迟
histogram_quantile(0.5, rate(myapp_http_request_duration_seconds_bucket[5m]))

# P95 延迟
histogram_quantile(0.95, rate(myapp_http_request_duration_seconds_bucket[5m]))

# P99 延迟（按路径分组）
histogram_quantile(0.99, sum by (le, path) (rate(myapp_http_request_duration_seconds_bucket[5m])))

# 平均延迟
rate(myapp_http_request_duration_seconds_sum[5m]) / rate(myapp_http_request_duration_seconds_count[5m])

# 延迟分布 (每个桶的比例)
rate(myapp_http_request_duration_seconds_bucket[5m]) / ignoring(le) group_left rate(myapp_http_request_duration_seconds_count[5m])

# 超过 1 秒的请求比例
1 - (rate(myapp_http_request_duration_seconds_bucket{le="1"}[5m]) / rate(myapp_http_request_duration_seconds_count[5m]))
```

### 选择合适的桶

```go
package metrics

import "github.com/prometheus/client_golang/prometheus"

// API 请求 (毫秒级响应)
var apiBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// 数据库查询 (毫秒到秒级)
var dbBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5}

// 批处理任务 (秒到小时级)
var batchBuckets = []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600}

// 使用指数桶
var exponentialBuckets = prometheus.ExponentialBuckets(0.001, 2, 15) // 从 1ms 开始，每次 x2，共 15 个桶

// 使用线性桶
var linearBuckets = prometheus.LinearBuckets(0, 0.1, 20) // 从 0 开始，步长 0.1，共 20 个桶
```

## 自定义收集器

### 实现 Collector 接口

```go
package metrics

import (
    "database/sql"

    "github.com/prometheus/client_golang/prometheus"
)

// DBStatsCollector 收集数据库连接池统计
type DBStatsCollector struct {
    db *sql.DB

    maxOpenConnections *prometheus.Desc
    openConnections    *prometheus.Desc
    inUseConnections   *prometheus.Desc
    idleConnections    *prometheus.Desc
    waitCount          *prometheus.Desc
    waitDuration       *prometheus.Desc
}

func NewDBStatsCollector(db *sql.DB, dbName string) *DBStatsCollector {
    labels := prometheus.Labels{"db": dbName}

    return &DBStatsCollector{
        db: db,
        maxOpenConnections: prometheus.NewDesc(
            "myapp_db_max_open_connections",
            "Maximum number of open connections to the database",
            nil, labels,
        ),
        openConnections: prometheus.NewDesc(
            "myapp_db_open_connections",
            "Number of open connections to the database",
            nil, labels,
        ),
        inUseConnections: prometheus.NewDesc(
            "myapp_db_in_use_connections",
            "Number of connections currently in use",
            nil, labels,
        ),
        idleConnections: prometheus.NewDesc(
            "myapp_db_idle_connections",
            "Number of idle connections",
            nil, labels,
        ),
        waitCount: prometheus.NewDesc(
            "myapp_db_wait_count_total",
            "Total number of connections waited for",
            nil, labels,
        ),
        waitDuration: prometheus.NewDesc(
            "myapp_db_wait_duration_seconds_total",
            "Total time blocked waiting for a new connection",
            nil, labels,
        ),
    }
}

// Describe 发送所有指标的描述
func (c *DBStatsCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.maxOpenConnections
    ch <- c.openConnections
    ch <- c.inUseConnections
    ch <- c.idleConnections
    ch <- c.waitCount
    ch <- c.waitDuration
}

// Collect 收集指标值
func (c *DBStatsCollector) Collect(ch chan<- prometheus.Metric) {
    stats := c.db.Stats()

    ch <- prometheus.MustNewConstMetric(
        c.maxOpenConnections, prometheus.GaugeValue, float64(stats.MaxOpenConnections),
    )
    ch <- prometheus.MustNewConstMetric(
        c.openConnections, prometheus.GaugeValue, float64(stats.OpenConnections),
    )
    ch <- prometheus.MustNewConstMetric(
        c.inUseConnections, prometheus.GaugeValue, float64(stats.InUse),
    )
    ch <- prometheus.MustNewConstMetric(
        c.idleConnections, prometheus.GaugeValue, float64(stats.Idle),
    )
    ch <- prometheus.MustNewConstMetric(
        c.waitCount, prometheus.CounterValue, float64(stats.WaitCount),
    )
    ch <- prometheus.MustNewConstMetric(
        c.waitDuration, prometheus.CounterValue, stats.WaitDuration.Seconds(),
    )
}

// 注册收集器
func RegisterDBStatsCollector(db *sql.DB, dbName string) {
    prometheus.MustRegister(NewDBStatsCollector(db, dbName))
}
```

### 业务指标收集器

```go
package metrics

import (
    "context"

    "github.com/prometheus/client_golang/prometheus"
)

// OrderStats 订单统计数据
type OrderStats struct {
    PendingCount   int
    ProcessingCount int
    CompletedToday int
    TotalAmount    float64
}

// OrderStatsProvider 订单统计数据提供者接口
type OrderStatsProvider interface {
    GetOrderStats(ctx context.Context) (*OrderStats, error)
}

// OrderStatsCollector 订单统计收集器
type OrderStatsCollector struct {
    provider OrderStatsProvider

    pendingOrders    *prometheus.Desc
    processingOrders *prometheus.Desc
    completedOrders  *prometheus.Desc
    totalAmount      *prometheus.Desc
}

func NewOrderStatsCollector(provider OrderStatsProvider) *OrderStatsCollector {
    return &OrderStatsCollector{
        provider: provider,
        pendingOrders: prometheus.NewDesc(
            "myapp_orders_pending",
            "Number of pending orders",
            nil, nil,
        ),
        processingOrders: prometheus.NewDesc(
            "myapp_orders_processing",
            "Number of orders being processed",
            nil, nil,
        ),
        completedOrders: prometheus.NewDesc(
            "myapp_orders_completed_today",
            "Number of orders completed today",
            nil, nil,
        ),
        totalAmount: prometheus.NewDesc(
            "myapp_orders_total_amount_dollars",
            "Total amount of orders today",
            nil, nil,
        ),
    }
}

func (c *OrderStatsCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.pendingOrders
    ch <- c.processingOrders
    ch <- c.completedOrders
    ch <- c.totalAmount
}

func (c *OrderStatsCollector) Collect(ch chan<- prometheus.Metric) {
    stats, err := c.provider.GetOrderStats(context.Background())
    if err != nil {
        // 记录错误，但不中断收集
        return
    }

    ch <- prometheus.MustNewConstMetric(c.pendingOrders, prometheus.GaugeValue, float64(stats.PendingCount))
    ch <- prometheus.MustNewConstMetric(c.processingOrders, prometheus.GaugeValue, float64(stats.ProcessingCount))
    ch <- prometheus.MustNewConstMetric(c.completedOrders, prometheus.GaugeValue, float64(stats.CompletedToday))
    ch <- prometheus.MustNewConstMetric(c.totalAmount, prometheus.GaugeValue, stats.TotalAmount)
}
```

## 完整示例：电商服务指标

```go
package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "ecommerce"

var (
    // ==================== HTTP 指标 ====================

    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "http",
            Name:      "requests_total",
            Help:      "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: namespace,
            Subsystem: "http",
            Name:      "request_duration_seconds",
            Help:      "HTTP request duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
        []string{"method", "path"},
    )

    // ==================== 用户指标 ====================

    UserRegistrations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "user",
            Name:      "registrations_total",
            Help:      "Total user registrations",
        },
        []string{"source"},
    )

    UserLogins = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "user",
            Name:      "logins_total",
            Help:      "Total user logins",
        },
        []string{"method", "status"},
    )

    ActiveSessions = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: namespace,
        Subsystem: "user",
        Name:      "active_sessions",
        Help:      "Number of active user sessions",
    })

    // ==================== 订单指标 ====================

    OrdersCreated = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "order",
            Name:      "created_total",
            Help:      "Total orders created",
        },
        []string{"category"},
    )

    OrderAmount = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: namespace,
            Subsystem: "order",
            Name:      "amount_dollars",
            Help:      "Order amount distribution",
            Buckets:   []float64{10, 50, 100, 250, 500, 1000, 2500, 5000},
        },
        []string{"category"},
    )

    OrderStatus = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: namespace,
            Subsystem: "order",
            Name:      "status_count",
            Help:      "Orders by status",
        },
        []string{"status"},
    )

    OrderProcessingTime = promauto.NewHistogram(prometheus.HistogramOpts{
        Namespace: namespace,
        Subsystem: "order",
        Name:      "processing_seconds",
        Help:      "Order processing time",
        Buckets:   []float64{1, 5, 10, 30, 60, 120, 300},
    })

    // ==================== 支付指标 ====================

    PaymentsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "payment",
            Name:      "total",
            Help:      "Total payments",
        },
        []string{"method", "status"},
    )

    PaymentAmount = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: namespace,
            Subsystem: "payment",
            Name:      "amount_dollars",
            Help:      "Payment amount distribution",
            Buckets:   []float64{10, 50, 100, 250, 500, 1000, 2500, 5000},
        },
        []string{"method"},
    )

    PaymentLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: namespace,
            Subsystem: "payment",
            Name:      "latency_seconds",
            Help:      "Payment processing latency",
            Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method"},
    )

    // ==================== 库存指标 ====================

    InventoryLevel = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: namespace,
            Subsystem: "inventory",
            Name:      "level",
            Help:      "Current inventory level",
        },
        []string{"product_id", "warehouse"},
    )

    InventoryReservations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "inventory",
            Name:      "reservations_total",
            Help:      "Total inventory reservations",
        },
        []string{"status"}, // success, failed, timeout
    )

    // ==================== 缓存指标 ====================

    CacheOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "cache",
            Name:      "operations_total",
            Help:      "Total cache operations",
        },
        []string{"operation", "result"},
    )

    CacheLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: namespace,
            Subsystem: "cache",
            Name:      "latency_seconds",
            Help:      "Cache operation latency",
            Buckets:   []float64{.0001, .0005, .001, .005, .01, .025, .05},
        },
        []string{"operation"},
    )

    // ==================== 数据库指标 ====================

    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: namespace,
            Subsystem: "db",
            Name:      "query_duration_seconds",
            Help:      "Database query duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"operation", "table"},
    )

    DBErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: namespace,
            Subsystem: "db",
            Name:      "errors_total",
            Help:      "Database errors",
        },
        []string{"operation", "error_type"},
    )
)

// ==================== 辅助函数 ====================

// RecordHTTPRequest 记录 HTTP 请求
func RecordHTTPRequest(method, path, status string, duration time.Duration) {
    HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
    HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordPayment 记录支付
func RecordPayment(method, status string, amount float64, latency time.Duration) {
    PaymentsTotal.WithLabelValues(method, status).Inc()
    PaymentAmount.WithLabelValues(method).Observe(amount)
    PaymentLatency.WithLabelValues(method).Observe(latency.Seconds())
}

// RecordCacheOperation 记录缓存操作
func RecordCacheOperation(operation, result string, latency time.Duration) {
    CacheOperations.WithLabelValues(operation, result).Inc()
    CacheLatency.WithLabelValues(operation).Observe(latency.Seconds())
}

// RecordDBQuery 记录数据库查询
func RecordDBQuery(operation, table string, duration time.Duration, err error) {
    DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
    if err != nil {
        DBErrors.WithLabelValues(operation, classifyDBError(err)).Inc()
    }
}

func classifyDBError(err error) string {
    // 根据错误类型分类
    return "unknown"
}
```

## 与 Node.js 对比

### Node.js (prom-client)

```javascript
import client from 'prom-client';

// Counter
const ordersTotal = new client.Counter({
    name: 'orders_total',
    help: 'Total orders',
    labelNames: ['status'],
});

ordersTotal.labels('completed').inc();

// Gauge
const activeUsers = new client.Gauge({
    name: 'active_users',
    help: 'Active users',
});

activeUsers.set(100);
activeUsers.inc();
activeUsers.dec();

// Histogram
const orderAmount = new client.Histogram({
    name: 'order_amount_dollars',
    help: 'Order amount',
    buckets: [10, 50, 100, 500, 1000],
    labelNames: ['category'],
});

orderAmount.labels('electronics').observe(299.99);
```

### Go (prometheus/client_golang)

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Counter
var ordersTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "orders_total",
        Help: "Total orders",
    },
    []string{"status"},
)

ordersTotal.WithLabelValues("completed").Inc()

// Gauge
var activeUsers = promauto.NewGauge(prometheus.GaugeOpts{
    Name: "active_users",
    Help: "Active users",
})

activeUsers.Set(100)
activeUsers.Inc()
activeUsers.Dec()

// Histogram
var orderAmount = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "order_amount_dollars",
        Help:    "Order amount",
        Buckets: []float64{10, 50, 100, 500, 1000},
    },
    []string{"category"},
)

orderAmount.WithLabelValues("electronics").Observe(299.99)
```

## 最佳实践

### 1. 命名规范

```go
// 格式: <namespace>_<subsystem>_<name>_<unit>
// 例如: myapp_http_requests_total
//       myapp_db_query_duration_seconds

// Counter 以 _total 结尾
requestsTotal
errorsTotal

// Histogram/Summary 包含单位
requestDurationSeconds
responseSizeBytes
```

### 2. 标签设计

```go
// ✅ 好的标签（有限、有意义的值）
labels := []string{"method", "status", "endpoint"}

// ❌ 不好的标签（无限可能的值）
labels := []string{"user_id", "request_id", "timestamp"}

// ✅ 使用路由模式
path := c.FullPath() // "/api/users/:id"

// ❌ 使用实际路径
path := c.Request.URL.Path // "/api/users/12345"
```

### 3. 避免指标爆炸

```go
// 限制标签基数
const maxLabelCardinality = 100

// 对高基数值进行分组
func bucketUserAgent(ua string) string {
    if strings.Contains(ua, "Chrome") {
        return "chrome"
    }
    if strings.Contains(ua, "Firefox") {
        return "firefox"
    }
    return "other"
}
```

**下一节**：[Grafana](./06-grafana.md) - 学习创建监控仪表板
