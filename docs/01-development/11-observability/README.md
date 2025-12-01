# 第十一章：可观测性 (Observability)

## 章节概述

可观测性是现代分布式系统的核心能力，本章学习日志、监控、追踪的实现。

## 学习目标

完成本章后，你将能够：

- 理解可观测性的三大支柱
- 使用 Zap 实现结构化日志
- 使用 Prometheus 收集和查询指标
- 使用 Grafana 创建监控仪表板
- 使用 Jaeger 实现分布式追踪
- 配置告警规则

## 与 JavaScript 对比

### Node.js 日志

```javascript
import winston from 'winston';

const logger = winston.createLogger({
    level: 'info',
    format: winston.format.json(),
    transports: [
        new winston.transports.Console(),
        new winston.transports.File({ filename: 'app.log' }),
    ],
});

logger.info('User logged in', { userId: 123 });
```

### Go 日志 (Zap)

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("User logged in",
    zap.Int("userId", 123),
)
```

### Node.js 指标

```javascript
import client from 'prom-client';

const counter = new client.Counter({
    name: 'http_requests_total',
    help: 'Total HTTP requests',
    labelNames: ['method', 'path'],
});

counter.inc({ method: 'GET', path: '/api/users' });
```

### Go 指标 (Prometheus)

```go
import "github.com/prometheus/client_golang/prometheus"

var httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests",
    },
    []string{"method", "path"},
)

httpRequestsTotal.WithLabelValues("GET", "/api/users").Inc()
```

## 章节内容

| 文档 | 主题 | 描述 |
|------|------|------|
| [01-observability-intro.md](./01-observability-intro.md) | 可观测性概述 | 三大支柱介绍 |
| [02-structured-logging.md](./02-structured-logging.md) | 结构化日志 | Zap 日志库 |
| [03-log-levels.md](./03-log-levels.md) | 日志级别 | 级别和上下文 |
| [04-prometheus.md](./04-prometheus.md) | Prometheus | 指标收集 |
| [05-custom-metrics.md](./05-custom-metrics.md) | 自定义指标 | Counter/Gauge/Histogram |
| [06-grafana.md](./06-grafana.md) | Grafana | 监控仪表板 |
| [07-alerting.md](./07-alerting.md) | 告警 | 告警规则配置 |
| [08-tracing-intro.md](./08-tracing-intro.md) | 分布式追踪 | 追踪概念 |
| [09-jaeger.md](./09-jaeger.md) | Jaeger | 追踪集成 |
| [10-elk.md](./10-elk.md) | ELK | 日志聚合 (可选) |
| [exercises.md](./exercises.md) | 练习 | 实践练习 |

## 快速开始

### 安装依赖

```bash
# Zap 日志
go get go.uber.org/zap

# Prometheus 客户端
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp

# OpenTelemetry (追踪)
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/jaeger
```

### 基础示例

```go
package main

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
)

var (
    logger *zap.Logger

    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: []float64{.001, .005, .01, .05, .1, .5, 1},
        },
        []string{"method", "path"},
    )
)

func init() {
    // 初始化日志
    logger, _ = zap.NewProduction()

    // 注册指标
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

// 可观测性中间件
func observabilityMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 处理请求
        c.Next()

        // 记录指标
        duration := time.Since(start).Seconds()
        status := c.Writer.Status()

        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            fmt.Sprintf("%d", status),
        ).Inc()

        httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)

        // 记录日志
        logger.Info("HTTP request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", status),
            zap.Duration("duration", time.Since(start)),
        )
    }
}

func main() {
    defer logger.Sync()

    r := gin.New()
    r.Use(observabilityMiddleware())

    // 业务路由
    r.GET("/api/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello"})
    })

    // 指标端点
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    logger.Info("Server starting", zap.Int("port", 8080))
    r.Run(":8080")
}
```

## 可观测性三大支柱

```
┌──────────────────────────────────────────────────────────────────┐
│                        可观测性 (Observability)                    │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐│
│  │      Logs        │  │     Metrics      │  │     Traces       ││
│  │      日志         │  │      指标        │  │      追踪        ││
│  ├──────────────────┤  ├──────────────────┤  ├──────────────────┤│
│  │ 发生了什么        │  │ 系统状态如何      │  │ 请求路径是什么   ││
│  │ What happened?   │  │ How is it doing? │  │ Where did it go? ││
│  ├──────────────────┤  ├──────────────────┤  ├──────────────────┤│
│  │ • 错误信息        │  │ • 请求数量       │  │ • 调用链路        ││
│  │ • 调试信息        │  │ • 响应时间       │  │ • 服务依赖        ││
│  │ • 审计日志        │  │ • 错误率         │  │ • 延迟分布        ││
│  │ • 业务事件        │  │ • 资源使用       │  │ • 瓶颈定位        ││
│  └──────────────────┘  └──────────────────┘  └──────────────────┘│
│                                                                  │
│  工具:                  工具:                  工具:              │
│  • Zap                 • Prometheus           • Jaeger           │
│  • Zerolog             • Grafana              • Zipkin           │
│  • ELK Stack           • InfluxDB             • OpenTelemetry    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                     可观测性基础设施                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐           │
│  │ Service │  │ Service │  │ Service │  │ Service │           │
│  │    A    │  │    B    │  │    C    │  │    D    │           │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘           │
│       │            │            │            │                  │
│       └────────────┴────────────┴────────────┘                  │
│                          │                                      │
│           ┌──────────────┼──────────────┐                      │
│           ▼              ▼              ▼                      │
│    ┌──────────┐   ┌──────────┐   ┌──────────┐                 │
│    │  Fluentd │   │Prometheus│   │  Jaeger  │                 │
│    │   日志    │   │   指标   │   │   追踪   │                 │
│    └─────┬────┘   └─────┬────┘   └─────┬────┘                 │
│          │              │              │                        │
│          ▼              ▼              ▼                        │
│    ┌──────────┐   ┌──────────┐   ┌──────────┐                 │
│    │   ELK    │   │ Grafana  │   │ Jaeger   │                 │
│    │  日志UI   │   │  指标UI  │   │  追踪UI  │                 │
│    └──────────┘   └──────────┘   └──────────┘                 │
│                          │                                      │
│                          ▼                                      │
│                   ┌──────────┐                                 │
│                   │Alertmgr │                                 │
│                   │  告警    │                                 │
│                   └──────────┘                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 结构化日志 | 使用 JSON 格式，便于搜索和分析 |
| 请求 ID | 每个请求分配唯一 ID，贯穿所有服务 |
| 上下文传递 | 在服务间传递追踪上下文 |
| 采样策略 | 高流量时使用采样减少开销 |
| 告警分级 | 区分紧急和非紧急告警 |
| 仪表板规划 | 按服务/功能组织仪表板 |

**下一节**：[可观测性概述](./01-observability-intro.md) - 深入理解可观测性概念
