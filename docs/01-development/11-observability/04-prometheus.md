# Prometheus 指标收集

## 概述

Prometheus 是一个开源的监控和告警系统，使用拉取（Pull）模式收集时序数据。

```
┌──────────────────────────────────────────────────────────────────┐
│                     Prometheus 架构                               │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐      │
│  │   Service   │      │   Service   │      │   Service   │      │
│  │  /metrics   │      │  /metrics   │      │  /metrics   │      │
│  └──────┬──────┘      └──────┬──────┘      └──────┬──────┘      │
│         │                    │                    │              │
│         └────────────────────┼────────────────────┘              │
│                              │ Pull                              │
│                              ▼                                   │
│                    ┌─────────────────┐                          │
│                    │   Prometheus    │                          │
│                    │    Server       │                          │
│                    │  (TSDB存储)     │                          │
│                    └────────┬────────┘                          │
│                             │                                    │
│              ┌──────────────┼──────────────┐                    │
│              ▼              ▼              ▼                    │
│       ┌──────────┐   ┌──────────┐   ┌──────────┐              │
│       │ Grafana  │   │Alertmgr │   │  PromQL  │              │
│       │  可视化   │   │  告警    │   │  查询    │              │
│       └──────────┘   └──────────┘   └──────────┘              │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 安装

### Go 客户端库

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/prometheus/client_golang/prometheus/promauto
```

### Prometheus 服务器 (Docker)

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v /path/to/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

## 基本使用

### 暴露指标端点

```go
package main

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // 暴露 /metrics 端点
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

访问 `http://localhost:8080/metrics` 查看默认指标。

### 与 Gin 集成

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    r := gin.Default()

    // 指标端点
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // 业务路由
    r.GET("/api/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "hello"})
    })

    r.Run(":8080")
}
```

## 指标类型

Prometheus 有四种核心指标类型：

### 1. Counter（计数器）

只增不减的累计值，适用于请求数、错误数等。

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 定义 Counter
var httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

func main() {
    // 增加计数
    httpRequestsTotal.WithLabelValues("GET", "/api/users", "200").Inc()
    httpRequestsTotal.WithLabelValues("POST", "/api/users", "201").Inc()
    httpRequestsTotal.WithLabelValues("GET", "/api/users", "500").Inc()

    // 增加指定值
    httpRequestsTotal.WithLabelValues("GET", "/api/users", "200").Add(10)
}
```

### 2. Gauge（仪表盘）

可增可减的瞬时值，适用于温度、内存使用、活跃连接数等。

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 定义 Gauge
var (
    activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "active_connections",
        Help: "Number of active connections",
    })

    cpuTemperature = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cpu_temperature_celsius",
            Help: "Current CPU temperature",
        },
        []string{"cpu"},
    )
)

func main() {
    // 设置值
    activeConnections.Set(100)

    // 增加
    activeConnections.Inc()
    activeConnections.Add(10)

    // 减少
    activeConnections.Dec()
    activeConnections.Sub(5)

    // 带标签
    cpuTemperature.WithLabelValues("cpu0").Set(65.5)
    cpuTemperature.WithLabelValues("cpu1").Set(63.2)
}
```

### 3. Histogram（直方图）

观测值的分布统计，适用于请求延迟、响应大小等。

```go
package main

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 定义 Histogram
var httpRequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    },
    []string{"method", "path"},
)

func main() {
    // 记录观测值
    httpRequestDuration.WithLabelValues("GET", "/api/users").Observe(0.025)
    httpRequestDuration.WithLabelValues("POST", "/api/users").Observe(0.150)

    // 使用 Timer
    timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues("GET", "/api/orders"))
    // ... 执行操作 ...
    time.Sleep(100 * time.Millisecond)
    timer.ObserveDuration()
}
```

Histogram 会自动生成三个指标：
- `http_request_duration_seconds_bucket{le="0.1"}` - 小于等于 0.1 秒的请求数
- `http_request_duration_seconds_sum` - 所有请求延迟的总和
- `http_request_duration_seconds_count` - 请求总数

### 4. Summary（摘要）

计算滑动窗口内的分位数。

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 定义 Summary
var httpRequestDurationSummary = promauto.NewSummaryVec(
    prometheus.SummaryOpts{
        Name:       "http_request_duration_summary_seconds",
        Help:       "HTTP request duration summary",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
        // 0.5 分位数误差 5%，0.9 分位数误差 1%，0.99 分位数误差 0.1%
    },
    []string{"method", "path"},
)

func main() {
    httpRequestDurationSummary.WithLabelValues("GET", "/api/users").Observe(0.025)
}
```

### Histogram vs Summary

| 特性 | Histogram | Summary |
|------|-----------|---------|
| 分位数计算 | 服务端（PromQL） | 客户端 |
| 可聚合 | 是 | 否 |
| 准确性 | 取决于桶配置 | 取决于误差配置 |
| 资源消耗 | 低 | 高 |
| 推荐场景 | 通用场景 | 单实例精确分位数 |

**推荐使用 Histogram**，因为它可以在服务端聚合多个实例的数据。

## HTTP 中间件

### Gin 指标中间件

```go
package middleware

import (
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "myapp",
            Name:      "http_requests_total",
            Help:      "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Name:      "http_request_duration_seconds",
            Help:      "HTTP request duration in seconds",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "path"},
    )

    httpRequestSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Name:      "http_request_size_bytes",
            Help:      "HTTP request size in bytes",
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "path"},
    )

    httpResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "myapp",
            Name:      "http_response_size_bytes",
            Help:      "HTTP response size in bytes",
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "path"},
    )
)

// PrometheusMiddleware Gin Prometheus 中间件
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.FullPath() // 使用路由模式而不是实际路径
        if path == "" {
            path = "unknown"
        }

        // 记录请求大小
        reqSize := float64(c.Request.ContentLength)
        if reqSize > 0 {
            httpRequestSize.WithLabelValues(c.Request.Method, path).Observe(reqSize)
        }

        // 处理请求
        c.Next()

        // 记录响应
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        respSize := float64(c.Writer.Size())

        httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
        httpResponseSize.WithLabelValues(c.Request.Method, path).Observe(respSize)
    }
}
```

### 使用中间件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "myapp/middleware"
)

func main() {
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.PrometheusMiddleware())

    // 指标端点（不经过中间件）
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // 业务路由
    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"users": []string{"alice", "bob"}})
    })

    r.GET("/api/users/:id", func(c *gin.Context) {
        c.JSON(200, gin.H{"id": c.Param("id")})
    })

    r.Run(":8080")
}
```

## Prometheus 配置

### prometheus.yml

```yaml
global:
  scrape_interval: 15s      # 默认抓取间隔
  evaluation_interval: 15s  # 规则评估间隔

# 告警配置
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

# 规则文件
rule_files:
  - "rules/*.yml"

# 抓取配置
scrape_configs:
  # Prometheus 自身
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # Go 应用
  - job_name: 'myapp'
    static_configs:
      - targets: ['app:8080']
    metrics_path: /metrics
    scrape_interval: 10s

  # 多实例应用
  - job_name: 'api-servers'
    static_configs:
      - targets:
        - 'api1:8080'
        - 'api2:8080'
        - 'api3:8080'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        regex: '(.+):\d+'
        replacement: '${1}'

  # 服务发现 (Docker)
  - job_name: 'docker'
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
    relabel_configs:
      - source_labels: [__meta_docker_container_label_prometheus_job]
        regex: (.+)
        target_label: job
```

## PromQL 查询

### 基本查询

```promql
# 查看指标
http_requests_total

# 按标签过滤
http_requests_total{method="GET"}
http_requests_total{method="GET", status="200"}

# 正则匹配
http_requests_total{path=~"/api/.*"}
http_requests_total{status!="200"}
```

### 聚合函数

```promql
# 总和
sum(http_requests_total)
sum by (method) (http_requests_total)

# 平均值
avg(http_request_duration_seconds_sum / http_request_duration_seconds_count)

# 最大/最小
max(active_connections)
min(active_connections)

# 计数
count(up)
```

### 速率计算

```promql
# 每秒请求数 (QPS)
rate(http_requests_total[5m])

# 按方法分组的 QPS
sum by (method) (rate(http_requests_total[5m]))

# 错误率
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
```

### 分位数计算

```promql
# P50 延迟
histogram_quantile(0.5, rate(http_request_duration_seconds_bucket[5m]))

# P95 延迟
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))

# P99 延迟（按路径分组）
histogram_quantile(0.99, sum by (le, path) (rate(http_request_duration_seconds_bucket[5m])))
```

### 时间范围

```promql
# 5 分钟内的数据
http_requests_total[5m]

# 1 小时前的瞬时值
http_requests_total offset 1h

# 1 小时前 5 分钟范围的数据
http_requests_total[5m] offset 1h
```

## 注册自定义指标

### 手动注册

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    customCounter = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "custom_counter_total",
        Help: "A custom counter",
    })

    customGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "custom_gauge",
            Help: "A custom gauge",
        },
        []string{"label1", "label2"},
    )
)

func init() {
    // 手动注册
    prometheus.MustRegister(customCounter)
    prometheus.MustRegister(customGauge)
}
```

### 使用 promauto

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// promauto 自动注册到默认注册器
var (
    autoCounter = promauto.NewCounter(prometheus.CounterOpts{
        Name: "auto_counter_total",
        Help: "An auto-registered counter",
    })
)
```

### 自定义注册器

```go
package main

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // 创建自定义注册器
    reg := prometheus.NewRegistry()

    // 注册默认的 Go 运行时指标
    reg.MustRegister(prometheus.NewGoCollector())
    reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

    // 注册自定义指标
    counter := prometheus.NewCounter(prometheus.CounterOpts{
        Name: "my_counter_total",
        Help: "My counter",
    })
    reg.MustRegister(counter)

    // 使用自定义注册器的 Handler
    http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
    http.ListenAndServe(":8080", nil)
}
```

## 业务指标示例

### 用户服务指标

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 用户注册数
    UserRegistrations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "user_service",
            Name:      "registrations_total",
            Help:      "Total number of user registrations",
        },
        []string{"source"}, // web, mobile, api
    )

    // 用户登录数
    UserLogins = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "user_service",
            Name:      "logins_total",
            Help:      "Total number of user logins",
        },
        []string{"method", "status"}, // password/oauth, success/failure
    )

    // 活跃用户数
    ActiveUsers = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "user_service",
        Name:      "active_users",
        Help:      "Number of currently active users",
    })

    // 数据库查询延迟
    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "user_service",
            Name:      "db_query_duration_seconds",
            Help:      "Database query duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"operation"}, // select, insert, update, delete
    )
)

// 使用示例
func RegisterUser(source string) error {
    // ... 注册逻辑 ...
    UserRegistrations.WithLabelValues(source).Inc()
    return nil
}

func Login(method string, success bool) {
    status := "success"
    if !success {
        status = "failure"
    }
    UserLogins.WithLabelValues(method, status).Inc()
}
```

### 订单服务指标

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 订单数量
    OrdersCreated = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "order_service",
            Name:      "orders_created_total",
            Help:      "Total number of orders created",
        },
        []string{"product_type"},
    )

    // 订单金额
    OrderAmount = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "order_service",
            Name:      "order_amount_dollars",
            Help:      "Order amount in dollars",
            Buckets:   []float64{10, 50, 100, 500, 1000, 5000, 10000},
        },
        []string{"product_type"},
    )

    // 支付状态
    PaymentStatus = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "order_service",
            Name:      "payments_total",
            Help:      "Total number of payments by status",
        },
        []string{"status"}, // success, failed, pending
    )

    // 库存水平
    InventoryLevel = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "order_service",
            Name:      "inventory_level",
            Help:      "Current inventory level",
        },
        []string{"product_id"},
    )
)
```

## 与 Node.js 对比

### Node.js (prom-client)

```javascript
import client from 'prom-client';

// 收集默认指标
client.collectDefaultMetrics();

// Counter
const httpRequestsTotal = new client.Counter({
    name: 'http_requests_total',
    help: 'Total HTTP requests',
    labelNames: ['method', 'path', 'status'],
});

httpRequestsTotal.inc({ method: 'GET', path: '/api/users', status: 200 });

// Histogram
const httpDuration = new client.Histogram({
    name: 'http_request_duration_seconds',
    help: 'HTTP request duration',
    labelNames: ['method', 'path'],
    buckets: [0.1, 0.5, 1, 2, 5],
});

httpDuration.observe({ method: 'GET', path: '/api/users' }, 0.25);

// 暴露指标
app.get('/metrics', async (req, res) => {
    res.set('Content-Type', client.register.contentType);
    res.end(await client.register.metrics());
});
```

### Go (prometheus/client_golang)

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Counter
var httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests",
    },
    []string{"method", "path", "status"},
)

httpRequestsTotal.WithLabelValues("GET", "/api/users", "200").Inc()

// Histogram
var httpDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration",
        Buckets: []float64{0.1, 0.5, 1, 2, 5},
    },
    []string{"method", "path"},
)

httpDuration.WithLabelValues("GET", "/api/users").Observe(0.25)

// 暴露指标
http.Handle("/metrics", promhttp.Handler())
```

## Docker Compose 示例

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    networks:
      - monitoring

  prometheus:
    image: prom/prometheus:v2.45.0
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.enable-lifecycle'
    networks:
      - monitoring

volumes:
  prometheus_data:

networks:
  monitoring:
    driver: bridge
```

## 最佳实践

### 1. 命名规范

```go
// 格式: <namespace>_<subsystem>_<name>_<unit>

// 好的命名
var (
    httpRequestsTotal          // http_requests_total
    httpRequestDurationSeconds // http_request_duration_seconds
    dbConnectionsActive        // db_connections_active
    cacheHitsTotal            // cache_hits_total
)

// 不好的命名
var (
    requests            // 太模糊
    httpRequestDuration // 缺少单位
    http_request_count  // 使用下划线而非驼峰（Go 代码中）
)
```

### 2. 标签设计

```go
// 好的标签（低基数）
labels := []string{"method", "status_code", "endpoint"}

// 不好的标签（高基数，会导致指标爆炸）
labels := []string{"user_id", "request_id", "timestamp"}

// 使用路由模式而非实际路径
path := c.FullPath() // /api/users/:id 而非 /api/users/123
```

### 3. 桶配置

```go
// 根据实际延迟分布选择桶
// API 请求（毫秒级）
Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// 批处理任务（秒级）
Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600}

// 指数桶
Buckets: prometheus.ExponentialBuckets(0.001, 2, 15) // 从 1ms 开始，每次 x2，共 15 个桶
```

**下一节**：[自定义指标](./05-custom-metrics.md) - 学习设计和实现自定义指标
