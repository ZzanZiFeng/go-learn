# 告警配置 (Alerting)

## 概述

告警是可观测性的重要组成部分，帮助团队在问题影响用户之前发现和响应。

```
┌──────────────────────────────────────────────────────────────────┐
│                       告警流程                                    │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌────────┐ │
│  │Prometheus│────▶│ Alert    │────▶│Alertmgr │────▶│ 通知   │ │
│  │ 评估规则  │     │ 触发     │     │ 路由/抑制│     │ 渠道   │ │
│  └──────────┘     └──────────┘     └──────────┘     └────────┘ │
│                                                                  │
│  告警状态:                                                        │
│  • Inactive - 条件未满足                                          │
│  • Pending  - 条件满足，等待 for 时间                             │
│  • Firing   - 触发告警，发送通知                                   │
│  • Resolved - 问题已解决                                          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Prometheus 告警规则

### 基本结构

```yaml
# prometheus/rules/alerts.yml
groups:
  - name: example
    rules:
      - alert: AlertName
        expr: <PromQL表达式>
        for: 5m                    # 持续多长时间才触发
        labels:
          severity: critical       # 告警级别
        annotations:
          summary: "简短描述"
          description: "详细描述 {{ $value }}"
```

### 服务可用性告警

```yaml
groups:
  - name: availability
    rules:
      # 服务下线
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务 {{ $labels.job }} 下线"
          description: "实例 {{ $labels.instance }} 已下线超过 1 分钟"

      # 高错误率
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) by (job)
          /
          sum(rate(http_requests_total[5m])) by (job)
          > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.job }} 错误率过高"
          description: "错误率为 {{ printf \"%.2f\" $value }}%，超过 5% 阈值"

      # 高延迟
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, sum by (le, job) (rate(http_request_duration_seconds_bucket[5m])))
          > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.job }} P95 延迟过高"
          description: "P95 延迟为 {{ printf \"%.2f\" $value }}s，超过 1s 阈值"
```

### 资源告警

```yaml
groups:
  - name: resources
    rules:
      # 内存使用过高
      - alert: HighMemoryUsage
        expr: |
          go_memstats_alloc_bytes / go_memstats_sys_bytes * 100 > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.job }} 内存使用过高"
          description: "内存使用率为 {{ printf \"%.1f\" $value }}%"

      # Goroutine 过多
      - alert: TooManyGoroutines
        expr: go_goroutines > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.job }} Goroutine 过多"
          description: "当前 Goroutine 数量: {{ $value }}"

      # 数据库连接池饱和
      - alert: DBPoolSaturated
        expr: |
          db_pool_connections_active
          /
          (db_pool_connections_active + db_pool_connections_idle)
          > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "数据库连接池接近饱和"
          description: "连接池使用率为 {{ printf \"%.1f\" $value }}%"

      # 数据库连接等待
      - alert: DBPoolWaiting
        expr: db_pool_connections_waiting > 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "存在等待数据库连接的请求"
          description: "等待连接数: {{ $value }}"
```

### 业务告警

```yaml
groups:
  - name: business
    rules:
      # 订单处理延迟
      - alert: OrderProcessingDelayed
        expr: |
          histogram_quantile(0.95, sum by (le) (rate(order_processing_seconds_bucket[5m])))
          > 60
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "订单处理延迟过高"
          description: "P95 处理时间为 {{ printf \"%.1f\" $value }}s"

      # 支付失败率
      - alert: HighPaymentFailureRate
        expr: |
          sum(rate(payments_total{status="failed"}[5m]))
          /
          sum(rate(payments_total[5m]))
          > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "支付失败率过高"
          description: "支付失败率为 {{ printf \"%.1f\" $value }}%"

      # 库存低
      - alert: LowInventory
        expr: inventory_level < 10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "产品 {{ $labels.product_id }} 库存不足"
          description: "当前库存: {{ $value }}"

      # 消息队列积压
      - alert: QueueBacklog
        expr: queue_messages_pending > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "队列 {{ $labels.queue }} 消息积压"
          description: "待处理消息数: {{ $value }}"
```

### 缓存告警

```yaml
groups:
  - name: cache
    rules:
      # 缓存命中率低
      - alert: LowCacheHitRate
        expr: |
          sum(rate(cache_operations_total{result="hit"}[5m]))
          /
          sum(rate(cache_operations_total{operation="get"}[5m]))
          < 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "缓存命中率过低"
          description: "缓存命中率为 {{ printf \"%.1f\" $value }}%"

      # Redis 连接失败
      - alert: RedisConnectionFailed
        expr: redis_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis 连接失败"
          description: "无法连接到 Redis 实例 {{ $labels.instance }}"
```

## Alertmanager 配置

### 基本配置

```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alertmanager@example.com'
  smtp_auth_username: 'alertmanager@example.com'
  smtp_auth_password: 'password'

# 路由配置
route:
  receiver: 'default'
  group_by: ['alertname', 'job']
  group_wait: 30s       # 等待收集同组告警
  group_interval: 5m    # 同组告警发送间隔
  repeat_interval: 4h   # 重复发送间隔

  routes:
    # Critical 告警立即发送
    - match:
        severity: critical
      receiver: 'critical-alerts'
      group_wait: 10s
      repeat_interval: 1h

    # Warning 告警
    - match:
        severity: warning
      receiver: 'warning-alerts'

    # 特定服务告警
    - match_re:
        job: 'payment.*'
      receiver: 'payment-team'

# 接收者配置
receivers:
  - name: 'default'
    email_configs:
      - to: 'team@example.com'

  - name: 'critical-alerts'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/xxx'
        channel: '#critical-alerts'
        title: '🚨 Critical Alert'
        text: "{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}"
    pagerduty_configs:
      - service_key: 'your-service-key'

  - name: 'warning-alerts'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/xxx'
        channel: '#alerts'
        title: '⚠️ Warning Alert'

  - name: 'payment-team'
    email_configs:
      - to: 'payment-team@example.com'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/xxx'
        channel: '#payment-alerts'

# 告警抑制
inhibit_rules:
  # 当服务 down 时，抑制其他相关告警
  - source_match:
      alertname: 'ServiceDown'
    target_match_re:
      alertname: '.*'
    equal: ['job', 'instance']
```

### Slack 通知模板

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/xxx'
        channel: '#alerts'
        send_resolved: true
        title: '{{ if eq .Status "firing" }}🔥{{ else }}✅{{ end }} {{ .CommonAnnotations.summary }}'
        text: |
          {{ range .Alerts }}
          *Alert:* {{ .Annotations.summary }}
          *Description:* {{ .Annotations.description }}
          *Severity:* {{ .Labels.severity }}
          *Job:* {{ .Labels.job }}
          *Instance:* {{ .Labels.instance }}
          {{ end }}
        actions:
          - type: button
            text: 'View in Grafana'
            url: 'http://grafana:3000/d/alerts'
          - type: button
            text: 'Silence'
            url: '{{ template "__silenceUrl" . }}'
```

### 钉钉通知

```yaml
receivers:
  - name: 'dingtalk'
    webhook_configs:
      - url: 'http://dingtalk-webhook:8060/dingtalk/webhook1/send'
        send_resolved: true
```

钉钉 Webhook 服务 (Go 实现):

```go
package main

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"

    "github.com/gin-gonic/gin"
)

type AlertmanagerWebhook struct {
    Status   string  `json:"status"`
    Alerts   []Alert `json:"alerts"`
}

type Alert struct {
    Status      string            `json:"status"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
}

type DingTalkMessage struct {
    Msgtype  string          `json:"msgtype"`
    Markdown MarkdownContent `json:"markdown"`
}

type MarkdownContent struct {
    Title string `json:"title"`
    Text  string `json:"text"`
}

func main() {
    r := gin.Default()

    r.POST("/dingtalk/:token/send", func(c *gin.Context) {
        token := c.Param("token")

        var webhook AlertmanagerWebhook
        if err := c.ShouldBindJSON(&webhook); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // 构建钉钉消息
        message := buildDingTalkMessage(webhook)

        // 发送到钉钉
        dingTalkURL := "https://oapi.dingtalk.com/robot/send?access_token=" + token
        sendToDingTalk(dingTalkURL, message)

        c.JSON(200, gin.H{"status": "ok"})
    })

    r.Run(":8060")
}

func buildDingTalkMessage(webhook AlertmanagerWebhook) DingTalkMessage {
    var text string
    status := "🔥 告警触发"
    if webhook.Status == "resolved" {
        status = "✅ 告警恢复"
    }

    for _, alert := range webhook.Alerts {
        text += "### " + alert.Annotations["summary"] + "\n\n"
        text += "**状态**: " + status + "\n\n"
        text += "**描述**: " + alert.Annotations["description"] + "\n\n"
        text += "**级别**: " + alert.Labels["severity"] + "\n\n"
        text += "---\n\n"
    }

    return DingTalkMessage{
        Msgtype: "markdown",
        Markdown: MarkdownContent{
            Title: status,
            Text:  text,
        },
    }
}

func sendToDingTalk(url string, message DingTalkMessage) error {
    body, _ := json.Marshal(message)
    resp, err := http.Post(url, "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    io.Copy(io.Discard, resp.Body)
    return nil
}
```

## Go 应用内告警

### 自定义告警检查

```go
package alerts

import (
    "context"
    "log"
    "time"
)

// AlertChecker 告警检查器
type AlertChecker struct {
    notifier Notifier
    interval time.Duration
}

// Notifier 通知接口
type Notifier interface {
    Send(ctx context.Context, alert Alert) error
}

// Alert 告警
type Alert struct {
    Name        string
    Severity    string
    Summary     string
    Description string
    Labels      map[string]string
}

// HealthCheck 健康检查函数类型
type HealthCheck func(ctx context.Context) error

func NewAlertChecker(notifier Notifier, interval time.Duration) *AlertChecker {
    return &AlertChecker{
        notifier: notifier,
        interval: interval,
    }
}

// RegisterCheck 注册健康检查
func (ac *AlertChecker) RegisterCheck(name string, check HealthCheck) {
    go func() {
        ticker := time.NewTicker(ac.interval)
        defer ticker.Stop()

        for range ticker.C {
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            if err := check(ctx); err != nil {
                ac.notifier.Send(ctx, Alert{
                    Name:        name,
                    Severity:    "warning",
                    Summary:     name + " 检查失败",
                    Description: err.Error(),
                })
            }
            cancel()
        }
    }()
}
```

### 业务阈值告警

```go
package alerts

import (
    "context"
    "fmt"
    "time"
)

// ThresholdAlert 阈值告警
type ThresholdAlert struct {
    Name        string
    Metric      func() float64
    Threshold   float64
    Comparison  string // "gt", "lt", "eq"
    Duration    time.Duration
    Severity    string
    notifier    Notifier
}

func (ta *ThresholdAlert) Start(ctx context.Context) {
    var violationStart time.Time
    var inViolation bool

    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            value := ta.Metric()
            isViolating := ta.checkThreshold(value)

            if isViolating && !inViolation {
                // 开始违反阈值
                violationStart = time.Now()
                inViolation = true
            } else if !isViolating && inViolation {
                // 恢复正常
                inViolation = false
            }

            // 检查是否超过持续时间
            if inViolation && time.Since(violationStart) >= ta.Duration {
                ta.notifier.Send(ctx, Alert{
                    Name:     ta.Name,
                    Severity: ta.Severity,
                    Summary:  fmt.Sprintf("%s 超过阈值", ta.Name),
                    Description: fmt.Sprintf(
                        "当前值: %.2f, 阈值: %.2f, 持续时间: %v",
                        value, ta.Threshold, time.Since(violationStart),
                    ),
                })
                // 重置，避免重复发送
                violationStart = time.Now()
            }
        }
    }
}

func (ta *ThresholdAlert) checkThreshold(value float64) bool {
    switch ta.Comparison {
    case "gt":
        return value > ta.Threshold
    case "lt":
        return value < ta.Threshold
    case "eq":
        return value == ta.Threshold
    default:
        return false
    }
}
```

### 使用示例

```go
package main

import (
    "context"
    "myapp/alerts"
    "time"
)

func main() {
    notifier := &SlackNotifier{webhookURL: "..."}

    // 阈值告警
    errorRateAlert := &alerts.ThresholdAlert{
        Name:       "HighErrorRate",
        Metric:     getErrorRate,
        Threshold:  0.05,
        Comparison: "gt",
        Duration:   5 * time.Minute,
        Severity:   "critical",
        notifier:   notifier,
    }

    ctx := context.Background()
    go errorRateAlert.Start(ctx)

    // 健康检查告警
    checker := alerts.NewAlertChecker(notifier, 30*time.Second)
    checker.RegisterCheck("DatabaseConnection", checkDatabase)
    checker.RegisterCheck("RedisConnection", checkRedis)

    // 保持运行
    select {}
}

func getErrorRate() float64 {
    // 从 Prometheus 或内部计数器获取
    return 0.03
}

func checkDatabase(ctx context.Context) error {
    // 检查数据库连接
    return nil
}

func checkRedis(ctx context.Context) error {
    // 检查 Redis 连接
    return nil
}
```

## 告警最佳实践

### 1. 告警分级

```yaml
# 告警级别定义
severity_levels:
  critical:
    description: "影响用户，需要立即处理"
    response_time: "5分钟"
    notification: "电话 + Slack + PagerDuty"
    examples:
      - 服务完全不可用
      - 错误率 > 10%
      - 支付系统故障

  warning:
    description: "潜在问题，需要关注"
    response_time: "30分钟"
    notification: "Slack + 邮件"
    examples:
      - 错误率 > 5%
      - 延迟增加
      - 资源使用率高

  info:
    description: "信息通知，无需立即处理"
    response_time: "工作时间内"
    notification: "邮件"
    examples:
      - 部署完成
      - 配置变更
```

### 2. 避免告警疲劳

```yaml
# 不好的告警规则（太敏感）
- alert: HighCPU
  expr: cpu_usage > 50
  for: 1m

# 好的告警规则（合理阈值和持续时间）
- alert: HighCPU
  expr: cpu_usage > 80
  for: 10m
  labels:
    severity: warning

# 更高的阈值触发更严重的级别
- alert: CriticalCPU
  expr: cpu_usage > 95
  for: 5m
  labels:
    severity: critical
```

### 3. 可操作的告警

```yaml
# 不好的告警（信息不足）
annotations:
  summary: "错误发生"

# 好的告警（信息完整，可操作）
annotations:
  summary: "{{ $labels.job }} 错误率过高 ({{ printf \"%.1f\" $value }}%)"
  description: |
    服务 {{ $labels.job }} 在实例 {{ $labels.instance }} 上的错误率为 {{ printf "%.1f" $value }}%。

    排查步骤:
    1. 查看日志: kubectl logs -l app={{ $labels.job }}
    2. 检查依赖服务状态
    3. 查看最近变更: https://git.example.com/deploys

    Grafana: http://grafana:3000/d/{{ $labels.job }}
    Runbook: https://wiki.example.com/runbooks/{{ $labels.alertname }}
```

### 4. 告警测试

```go
package alerts_test

import (
    "testing"
    "time"
)

func TestHighErrorRateAlert(t *testing.T) {
    // 模拟高错误率
    mockMetrics := &MockMetrics{
        errorRate: 0.1, // 10%
    }

    alert := &ThresholdAlert{
        Name:       "HighErrorRate",
        Metric:     mockMetrics.GetErrorRate,
        Threshold:  0.05,
        Comparison: "gt",
        Duration:   1 * time.Second, // 测试用短时间
    }

    // 验证告警触发
    triggered := make(chan bool, 1)
    mockNotifier := &MockNotifier{
        onSend: func(alert Alert) {
            triggered <- true
        },
    }
    alert.notifier = mockNotifier

    go alert.Start(context.Background())

    select {
    case <-triggered:
        // 告警正常触发
    case <-time.After(5 * time.Second):
        t.Error("告警未触发")
    }
}
```

## Docker Compose 完整示例

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
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/rules:/etc/prometheus/rules
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--web.enable-lifecycle'
    networks:
      - monitoring

  alertmanager:
    image: prom/alertmanager:v0.25.0
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:10.0.0
    ports:
      - "3000:3000"
    volumes:
      - ./grafana/provisioning:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    networks:
      - monitoring

networks:
  monitoring:
    driver: bridge
```

**下一节**：[分布式追踪](./08-tracing-intro.md) - 学习追踪概念
