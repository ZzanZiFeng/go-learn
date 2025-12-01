# Grafana 监控仪表板

## 概述

Grafana 是一个开源的可视化和分析平台，支持多种数据源，用于创建交互式监控仪表板。

```
┌──────────────────────────────────────────────────────────────────┐
│                       Grafana 架构                                │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│                    ┌─────────────────────┐                      │
│                    │      Grafana        │                      │
│                    │     Dashboard       │                      │
│                    └─────────┬───────────┘                      │
│                              │                                   │
│           ┌──────────────────┼──────────────────┐               │
│           │                  │                  │               │
│           ▼                  ▼                  ▼               │
│    ┌──────────┐       ┌──────────┐       ┌──────────┐          │
│    │Prometheus│       │   Loki   │       │  Jaeger  │          │
│    │  指标     │       │   日志   │       │   追踪   │          │
│    └──────────┘       └──────────┘       └──────────┘          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## 安装与配置

### Docker 安装

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -v grafana-storage:/var/lib/grafana \
  grafana/grafana:10.0.0
```

### Docker Compose

```yaml
version: '3.8'

services:
  grafana:
    image: grafana/grafana:10.0.0
    ports:
      - "3000:3000"
    volumes:
      - grafana-storage:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
    depends_on:
      - prometheus

  prometheus:
    image: prom/prometheus:v2.45.0
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus

volumes:
  grafana-storage:
  prometheus-data:
```

### 数据源配置 (Provisioning)

创建 `grafana/provisioning/datasources/datasources.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: false

  - name: Jaeger
    type: jaeger
    access: proxy
    url: http://jaeger:16686
    editable: false
```

## 基础面板类型

### 1. Time Series（时间序列）

最常用的面板类型，用于显示随时间变化的数据。

```yaml
# 面板配置示例
title: "HTTP 请求率"
type: timeseries
datasource: Prometheus
targets:
  - expr: rate(http_requests_total[5m])
    legendFormat: "{{method}} {{path}}"
options:
  legend:
    displayMode: table
    placement: bottom
    calcs: [mean, max, last]
```

PromQL 示例:
```promql
# QPS
rate(http_requests_total[5m])

# 按状态码分组的 QPS
sum by (status) (rate(http_requests_total[5m]))

# 错误率
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
```

### 2. Stat（统计）

显示单个统计值，如当前值、平均值等。

```yaml
title: "当前活跃连接数"
type: stat
datasource: Prometheus
targets:
  - expr: active_connections
options:
  reduceOptions:
    calcs: [lastNotNull]
  colorMode: value
  graphMode: area
  textMode: auto
fieldConfig:
  defaults:
    thresholds:
      mode: absolute
      steps:
        - color: green
          value: null
        - color: yellow
          value: 50
        - color: red
          value: 100
```

### 3. Gauge（仪表盘）

显示值在范围内的位置。

```yaml
title: "CPU 使用率"
type: gauge
datasource: Prometheus
targets:
  - expr: avg(rate(process_cpu_seconds_total[5m])) * 100
options:
  reduceOptions:
    calcs: [lastNotNull]
fieldConfig:
  defaults:
    min: 0
    max: 100
    unit: percent
    thresholds:
      mode: absolute
      steps:
        - color: green
          value: null
        - color: yellow
          value: 70
        - color: red
          value: 90
```

### 4. Bar Gauge（条形仪表）

```yaml
title: "各服务请求数"
type: bargauge
datasource: Prometheus
targets:
  - expr: sum by (service) (rate(http_requests_total[5m]))
options:
  displayMode: gradient
  orientation: horizontal
```

### 5. Table（表格）

```yaml
title: "服务状态"
type: table
datasource: Prometheus
targets:
  - expr: up
    format: table
    instant: true
transformations:
  - id: organize
    options:
      excludeByName:
        __name__: true
        Time: true
      renameByName:
        instance: "实例"
        job: "服务"
        Value: "状态"
```

### 6. Heatmap（热力图）

用于显示 Histogram 数据的分布。

```yaml
title: "请求延迟分布"
type: heatmap
datasource: Prometheus
targets:
  - expr: sum by (le) (rate(http_request_duration_seconds_bucket[5m]))
    format: heatmap
options:
  yAxis:
    unit: s
  color:
    scheme: Spectral
```

## 仪表板示例

### Go 服务监控仪表板

```json
{
  "title": "Go Service Dashboard",
  "uid": "go-service",
  "panels": [
    {
      "title": "QPS",
      "type": "timeseries",
      "gridPos": {"x": 0, "y": 0, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "sum(rate(http_requests_total[5m]))",
          "legendFormat": "Total QPS"
        }
      ]
    },
    {
      "title": "错误率",
      "type": "timeseries",
      "gridPos": {"x": 12, "y": 0, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m])) * 100",
          "legendFormat": "Error Rate %"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 1},
              {"color": "red", "value": 5}
            ]
          }
        }
      }
    },
    {
      "title": "P95 延迟",
      "type": "timeseries",
      "gridPos": {"x": 0, "y": 8, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))",
          "legendFormat": "P95 Latency"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "s"
        }
      }
    },
    {
      "title": "Goroutines",
      "type": "timeseries",
      "gridPos": {"x": 12, "y": 8, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "go_goroutines",
          "legendFormat": "{{instance}}"
        }
      ]
    }
  ]
}
```

### RED 方法仪表板

RED (Rate, Errors, Duration) 是微服务监控的黄金指标。

```json
{
  "title": "RED Dashboard",
  "panels": [
    {
      "title": "Rate - 请求率",
      "type": "stat",
      "gridPos": {"x": 0, "y": 0, "w": 8, "h": 4},
      "targets": [
        {
          "expr": "sum(rate(http_requests_total[5m]))",
          "legendFormat": "QPS"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "reqps"
        }
      }
    },
    {
      "title": "Errors - 错误率",
      "type": "stat",
      "gridPos": {"x": 8, "y": 0, "w": 8, "h": 4},
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m])) * 100",
          "legendFormat": "Error Rate"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "thresholds": {
            "steps": [
              {"color": "green", "value": null},
              {"color": "red", "value": 1}
            ]
          }
        }
      }
    },
    {
      "title": "Duration - P99 延迟",
      "type": "stat",
      "gridPos": {"x": 16, "y": 0, "w": 8, "h": 4},
      "targets": [
        {
          "expr": "histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))",
          "legendFormat": "P99"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "s"
        }
      }
    },
    {
      "title": "请求率趋势",
      "type": "timeseries",
      "gridPos": {"x": 0, "y": 4, "w": 24, "h": 8},
      "targets": [
        {
          "expr": "sum by (path) (rate(http_requests_total[5m]))",
          "legendFormat": "{{path}}"
        }
      ]
    },
    {
      "title": "延迟分位数",
      "type": "timeseries",
      "gridPos": {"x": 0, "y": 12, "w": 24, "h": 8},
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))",
          "legendFormat": "P50"
        },
        {
          "expr": "histogram_quantile(0.90, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))",
          "legendFormat": "P90"
        },
        {
          "expr": "histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))",
          "legendFormat": "P99"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "s"
        }
      }
    }
  ]
}
```

### USE 方法仪表板

USE (Utilization, Saturation, Errors) 用于基础设施监控。

```json
{
  "title": "USE Dashboard - Database",
  "panels": [
    {
      "title": "Utilization - 连接池使用率",
      "type": "gauge",
      "targets": [
        {
          "expr": "db_pool_connections_active / (db_pool_connections_active + db_pool_connections_idle) * 100"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "min": 0,
          "max": 100
        }
      }
    },
    {
      "title": "Saturation - 等待连接数",
      "type": "timeseries",
      "targets": [
        {
          "expr": "db_pool_connections_waiting",
          "legendFormat": "Waiting"
        }
      ]
    },
    {
      "title": "Errors - 数据库错误率",
      "type": "timeseries",
      "targets": [
        {
          "expr": "rate(db_errors_total[5m])",
          "legendFormat": "{{error_type}}"
        }
      ]
    }
  ]
}
```

## 变量和模板

### 定义变量

```yaml
# 仪表板变量配置
templating:
  list:
    - name: instance
      type: query
      datasource: Prometheus
      query: label_values(up, instance)
      refresh: 2  # 仪表板加载时刷新
      sort: 1     # 按字母排序

    - name: job
      type: query
      datasource: Prometheus
      query: label_values(up, job)
      multi: true  # 允许多选
      includeAll: true

    - name: interval
      type: interval
      options: 1m,5m,10m,30m,1h
      auto: true
      auto_min: 10s
```

### 使用变量

```promql
# 在查询中使用变量
rate(http_requests_total{instance="$instance", job=~"$job"}[$interval])

# 多选变量使用正则匹配
rate(http_requests_total{job=~"$job"}[5m])
```

## 告警集成

### 在 Grafana 中配置告警

```yaml
# 告警规则配置
alert:
  name: "High Error Rate"
  conditions:
    - evaluator:
        type: gt
        params: [0.05]  # 5%
      operator:
        type: and
      query:
        params: [A, 5m, now]
      reducer:
        type: avg
  frequency: 1m
  for: 5m
  noDataState: no_data
  execErrState: alerting
  notifications:
    - uid: slack-channel
```

### 通知渠道配置

```yaml
# grafana/provisioning/notifiers/notifiers.yml
notifiers:
  - name: Slack
    type: slack
    uid: slack-channel
    settings:
      url: https://hooks.slack.com/services/xxx
      recipient: "#alerts"

  - name: Email
    type: email
    uid: email-team
    settings:
      addresses: team@example.com
```

## Go 应用仪表板 Provisioning

### 仪表板文件结构

```
grafana/
├── provisioning/
│   ├── datasources/
│   │   └── datasources.yml
│   ├── dashboards/
│   │   └── dashboards.yml
│   └── notifiers/
│       └── notifiers.yml
└── dashboards/
    ├── go-service.json
    ├── redis.json
    └── postgresql.json
```

### dashboards.yml

```yaml
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /var/lib/grafana/dashboards
```

### 完整 Go 服务仪表板

```json
{
  "title": "Go Application Dashboard",
  "uid": "go-app",
  "tags": ["go", "application"],
  "timezone": "browser",
  "refresh": "30s",
  "templating": {
    "list": [
      {
        "name": "instance",
        "type": "query",
        "datasource": "Prometheus",
        "query": "label_values(go_info, instance)",
        "refresh": 2
      }
    ]
  },
  "panels": [
    {
      "title": "服务状态",
      "type": "stat",
      "gridPos": {"x": 0, "y": 0, "w": 4, "h": 4},
      "targets": [
        {
          "expr": "up{instance=\"$instance\"}",
          "instant": true
        }
      ],
      "fieldConfig": {
        "defaults": {
          "mappings": [
            {"type": "value", "options": {"0": {"text": "DOWN", "color": "red"}}},
            {"type": "value", "options": {"1": {"text": "UP", "color": "green"}}}
          ]
        }
      }
    },
    {
      "title": "QPS",
      "type": "stat",
      "gridPos": {"x": 4, "y": 0, "w": 4, "h": 4},
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{instance=\"$instance\"}[5m]))"
        }
      ],
      "fieldConfig": {
        "defaults": {"unit": "reqps"}
      }
    },
    {
      "title": "错误率",
      "type": "stat",
      "gridPos": {"x": 8, "y": 0, "w": 4, "h": 4},
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{instance=\"$instance\",status=~\"5..\"}[5m])) / sum(rate(http_requests_total{instance=\"$instance\"}[5m])) * 100"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "thresholds": {
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 1},
              {"color": "red", "value": 5}
            ]
          }
        }
      }
    },
    {
      "title": "P95 延迟",
      "type": "stat",
      "gridPos": {"x": 12, "y": 0, "w": 4, "h": 4},
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{instance=\"$instance\"}[5m])))"
        }
      ],
      "fieldConfig": {
        "defaults": {"unit": "s"}
      }
    },
    {
      "title": "Goroutines",
      "type": "stat",
      "gridPos": {"x": 16, "y": 0, "w": 4, "h": 4},
      "targets": [
        {
          "expr": "go_goroutines{instance=\"$instance\"}"
        }
      ]
    },
    {
      "title": "内存使用",
      "type": "stat",
      "gridPos": {"x": 20, "y": 0, "w": 4, "h": 4},
      "targets": [
        {
          "expr": "go_memstats_alloc_bytes{instance=\"$instance\"}"
        }
      ],
      "fieldConfig": {
        "defaults": {"unit": "bytes"}
      }
    },
    {
      "title": "请求率",
      "type": "timeseries",
      "gridPos": {"x": 0, "y": 4, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "sum by (path) (rate(http_requests_total{instance=\"$instance\"}[5m]))",
          "legendFormat": "{{path}}"
        }
      ],
      "fieldConfig": {
        "defaults": {"unit": "reqps"}
      }
    },
    {
      "title": "响应状态码",
      "type": "timeseries",
      "gridPos": {"x": 12, "y": 4, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "sum by (status) (rate(http_requests_total{instance=\"$instance\"}[5m]))",
          "legendFormat": "{{status}}"
        }
      ]
    },
    {
      "title": "延迟分位数",
      "type": "timeseries",
      "gridPos": {"x": 0, "y": 12, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket{instance=\"$instance\"}[5m])))",
          "legendFormat": "P50"
        },
        {
          "expr": "histogram_quantile(0.90, sum by (le) (rate(http_request_duration_seconds_bucket{instance=\"$instance\"}[5m])))",
          "legendFormat": "P90"
        },
        {
          "expr": "histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket{instance=\"$instance\"}[5m])))",
          "legendFormat": "P99"
        }
      ],
      "fieldConfig": {
        "defaults": {"unit": "s"}
      }
    },
    {
      "title": "GC 暂停时间",
      "type": "timeseries",
      "gridPos": {"x": 12, "y": 12, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "rate(go_gc_duration_seconds_sum{instance=\"$instance\"}[5m])",
          "legendFormat": "GC Duration"
        }
      ],
      "fieldConfig": {
        "defaults": {"unit": "s"}
      }
    },
    {
      "title": "数据库查询延迟",
      "type": "timeseries",
      "gridPos": {"x": 0, "y": 20, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum by (le, operation) (rate(db_query_duration_seconds_bucket{instance=\"$instance\"}[5m])))",
          "legendFormat": "{{operation}} P95"
        }
      ],
      "fieldConfig": {
        "defaults": {"unit": "s"}
      }
    },
    {
      "title": "缓存命中率",
      "type": "timeseries",
      "gridPos": {"x": 12, "y": 20, "w": 12, "h": 8},
      "targets": [
        {
          "expr": "sum(rate(cache_operations_total{instance=\"$instance\",result=\"hit\"}[5m])) / sum(rate(cache_operations_total{instance=\"$instance\",operation=\"get\"}[5m])) * 100",
          "legendFormat": "Hit Rate"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "min": 0,
          "max": 100
        }
      }
    }
  ]
}
```

## 与 Node.js 监控对比

### Node.js 常用指标

```javascript
// Express + prom-client
import express from 'express';
import client from 'prom-client';

// Event Loop Lag
const eventLoopLag = new client.Histogram({
    name: 'nodejs_eventloop_lag_seconds',
    help: 'Event loop lag',
    buckets: [0.001, 0.01, 0.1, 1],
});

// Active Handles
const activeHandles = new client.Gauge({
    name: 'nodejs_active_handles',
    help: 'Number of active handles',
});

// Heap Used
const heapUsed = new client.Gauge({
    name: 'nodejs_heap_used_bytes',
    help: 'Heap used in bytes',
});
```

### Go 常用指标

```go
// Go 运行时指标 (自动收集)
// go_goroutines          - Goroutine 数量
// go_memstats_alloc_bytes - 内存分配
// go_gc_duration_seconds  - GC 暂停时间

// 自定义业务指标
var httpRequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration",
        Buckets: prometheus.DefBuckets,
    },
    []string{"method", "path"},
)
```

## 最佳实践

### 1. 仪表板组织

```
仪表板/
├── 概览/
│   ├── 服务健康
│   └── 业务指标
├── 服务/
│   ├── API Gateway
│   ├── User Service
│   └── Order Service
├── 基础设施/
│   ├── PostgreSQL
│   ├── Redis
│   └── RabbitMQ
└── 告警/
    └── Active Alerts
```

### 2. 命名规范

```yaml
# 仪表板命名
title: "[环境] 服务名 - 功能"
# 例如: "[Prod] User Service - Overview"

# UID 命名
uid: "env-service-function"
# 例如: "prod-user-overview"

# 面板命名
# 使用清晰、描述性的标题
# 好: "HTTP 请求延迟 (P95)"
# 差: "Latency"
```

### 3. 颜色规范

```yaml
# 状态颜色
green: 正常
yellow: 警告
orange: 严重警告
red: 错误/严重

# 阈值设置示例
thresholds:
  steps:
    - color: green
      value: null    # 默认
    - color: yellow
      value: 70      # 警告阈值
    - color: red
      value: 90      # 严重阈值
```

**下一节**：[告警](./07-alerting.md) - 学习配置告警规则
