# ELK 日志聚合 (可选)

## 概述

ELK Stack (Elasticsearch, Logstash, Kibana) 是流行的日志聚合解决方案，用于集中收集、存储和分析日志。

```
┌──────────────────────────────────────────────────────────────────┐
│                       ELK Stack 架构                              │
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
│                 │   Filebeat /    │  (日志收集)                  │
│                 │   Fluentd       │                             │
│                 └────────┬────────┘                             │
│                          │                                       │
│                          ▼                                       │
│                 ┌─────────────────┐                             │
│                 │   Logstash      │  (日志处理/转换)             │
│                 └────────┬────────┘                             │
│                          │                                       │
│                          ▼                                       │
│                 ┌─────────────────┐                             │
│                 │ Elasticsearch   │  (存储/索引)                 │
│                 └────────┬────────┘                             │
│                          │                                       │
│                          ▼                                       │
│                 ┌─────────────────┐                             │
│                 │    Kibana       │  (可视化)                    │
│                 │    :5601        │                             │
│                 └─────────────────┘                             │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Docker Compose 部署

```yaml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.10.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch-data:/usr/share/elasticsearch/data
    networks:
      - elk

  logstash:
    image: docker.elastic.co/logstash/logstash:8.10.0
    volumes:
      - ./logstash/pipeline:/usr/share/logstash/pipeline
    ports:
      - "5044:5044"   # Beats input
      - "5000:5000"   # TCP input
      - "9600:9600"   # API
    environment:
      - "LS_JAVA_OPTS=-Xms256m -Xmx256m"
    depends_on:
      - elasticsearch
    networks:
      - elk

  kibana:
    image: docker.elastic.co/kibana/kibana:8.10.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch
    networks:
      - elk

  filebeat:
    image: docker.elastic.co/beats/filebeat:8.10.0
    user: root
    volumes:
      - ./filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    depends_on:
      - logstash
    networks:
      - elk

volumes:
  elasticsearch-data:

networks:
  elk:
    driver: bridge
```

## Logstash 配置

### pipeline/logstash.conf

```conf
input {
  # Beats 输入
  beats {
    port => 5044
  }

  # TCP 输入（用于直接发送日志）
  tcp {
    port => 5000
    codec => json
  }
}

filter {
  # 解析 JSON 日志
  if [message] =~ /^\{/ {
    json {
      source => "message"
      target => "parsed"
    }

    # 提取字段
    mutate {
      rename => { "[parsed][level]" => "level" }
      rename => { "[parsed][msg]" => "log_message" }
      rename => { "[parsed][ts]" => "timestamp" }
      rename => { "[parsed][caller]" => "caller" }
      rename => { "[parsed][trace_id]" => "trace_id" }
      rename => { "[parsed][span_id]" => "span_id" }
    }
  }

  # 添加时间戳
  date {
    match => [ "timestamp", "ISO8601", "UNIX" ]
    target => "@timestamp"
  }

  # 添加服务信息
  if [docker][container][name] {
    mutate {
      add_field => { "service" => "%{[docker][container][name]}" }
    }
  }

  # 日志级别规范化
  mutate {
    uppercase => [ "level" ]
  }

  # GeoIP（如果有 IP 字段）
  if [client_ip] {
    geoip {
      source => "client_ip"
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "logs-%{+YYYY.MM.dd}"
  }

  # 调试输出
  # stdout { codec => rubydebug }
}
```

## Filebeat 配置

### filebeat/filebeat.yml

```yaml
filebeat.inputs:
  # Docker 容器日志
  - type: container
    paths:
      - '/var/lib/docker/containers/*/*.log'
    processors:
      - add_docker_metadata:
          host: "unix:///var/run/docker.sock"

  # 应用日志文件
  - type: log
    paths:
      - /var/log/app/*.log
    json.keys_under_root: true
    json.add_error_key: true

processors:
  - add_host_metadata: ~
  - add_cloud_metadata: ~

output.logstash:
  hosts: ["logstash:5044"]

# 或直接发送到 Elasticsearch
# output.elasticsearch:
#   hosts: ["elasticsearch:9200"]
#   index: "logs-%{+yyyy.MM.dd}"
```

## Go 应用集成

### 直接发送到 Logstash

```go
package logger

import (
    "encoding/json"
    "net"
    "sync"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// LogstashWriter 将日志写入 Logstash
type LogstashWriter struct {
    addr    string
    conn    net.Conn
    mu      sync.Mutex
    timeout time.Duration
}

func NewLogstashWriter(addr string) *LogstashWriter {
    return &LogstashWriter{
        addr:    addr,
        timeout: 5 * time.Second,
    }
}

func (w *LogstashWriter) connect() error {
    conn, err := net.DialTimeout("tcp", w.addr, w.timeout)
    if err != nil {
        return err
    }
    w.conn = conn
    return nil
}

func (w *LogstashWriter) Write(p []byte) (n int, err error) {
    w.mu.Lock()
    defer w.mu.Unlock()

    if w.conn == nil {
        if err := w.connect(); err != nil {
            return 0, err
        }
    }

    // 添加换行符
    data := append(p, '\n')
    n, err = w.conn.Write(data)
    if err != nil {
        w.conn = nil // 重连
        return 0, err
    }
    return n, nil
}

func (w *LogstashWriter) Sync() error {
    return nil
}

// InitLoggerWithLogstash 创建带 Logstash 输出的 Logger
func InitLoggerWithLogstash(logstashAddr string) *zap.Logger {
    // JSON 编码器
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.TimeKey = "ts"
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    // 多输出
    logstashWriter := NewLogstashWriter(logstashAddr)

    core := zapcore.NewTee(
        // 控制台输出
        zapcore.NewCore(
            zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
            zapcore.AddSync(os.Stdout),
            zap.DebugLevel,
        ),
        // Logstash 输出
        zapcore.NewCore(
            zapcore.NewJSONEncoder(encoderConfig),
            zapcore.AddSync(logstashWriter),
            zap.InfoLevel,
        ),
    )

    return zap.New(core, zap.AddCaller())
}
```

### 使用 Zap Hook

```go
package logger

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// ElasticsearchHook 直接发送到 Elasticsearch
type ElasticsearchHook struct {
    client   *http.Client
    endpoint string
    index    string
}

func NewElasticsearchHook(endpoint, index string) *ElasticsearchHook {
    return &ElasticsearchHook{
        client:   &http.Client{Timeout: 5 * time.Second},
        endpoint: endpoint,
        index:    index,
    }
}

func (h *ElasticsearchHook) Write(p []byte) (n int, err error) {
    // 构建索引 URL
    url := h.endpoint + "/" + h.index + "-" + time.Now().Format("2006.01.02") + "/_doc"

    resp, err := h.client.Post(url, "application/json", bytes.NewReader(p))
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    return len(p), nil
}

func (h *ElasticsearchHook) Sync() error {
    return nil
}
```

### 文件日志 + Filebeat

推荐方式：应用写入文件，Filebeat 收集。

```go
package logger

import (
    "os"
    "path/filepath"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志配置
type Config struct {
    Level      string
    LogDir     string
    MaxSize    int  // MB
    MaxBackups int
    MaxAge     int  // days
    Compress   bool
}

// InitLogger 初始化文件日志
func InitLogger(cfg Config) *zap.Logger {
    // 解析日志级别
    var level zapcore.Level
    level.UnmarshalText([]byte(cfg.Level))

    // 日志轮转
    logFile := filepath.Join(cfg.LogDir, "app.log")
    writer := &lumberjack.Logger{
        Filename:   logFile,
        MaxSize:    cfg.MaxSize,
        MaxBackups: cfg.MaxBackups,
        MaxAge:     cfg.MaxAge,
        Compress:   cfg.Compress,
    }

    // JSON 编码器（便于 ELK 解析）
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "ts",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }

    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        zapcore.AddSync(writer),
        level,
    )

    return zap.New(core, zap.AddCaller())
}
```

## 日志格式规范

### 结构化日志字段

```go
// 推荐的日志字段
type LogEntry struct {
    Timestamp   string `json:"ts"`         // ISO8601 时间戳
    Level       string `json:"level"`      // 日志级别
    Message     string `json:"msg"`        // 日志消息
    Caller      string `json:"caller"`     // 调用位置

    // 追踪字段
    TraceID     string `json:"trace_id,omitempty"`
    SpanID      string `json:"span_id,omitempty"`

    // 请求字段
    RequestID   string `json:"request_id,omitempty"`
    Method      string `json:"method,omitempty"`
    Path        string `json:"path,omitempty"`
    StatusCode  int    `json:"status,omitempty"`
    Latency     string `json:"latency,omitempty"`
    ClientIP    string `json:"client_ip,omitempty"`

    // 用户字段
    UserID      string `json:"user_id,omitempty"`

    // 错误字段
    Error       string `json:"error,omitempty"`
    ErrorCode   string `json:"error_code,omitempty"`
    StackTrace  string `json:"stacktrace,omitempty"`

    // 服务字段
    Service     string `json:"service"`
    Version     string `json:"version"`
    Environment string `json:"env"`
}
```

### 示例日志输出

```json
{
  "ts": "2024-01-01T10:00:00.000Z",
  "level": "info",
  "msg": "Request completed",
  "caller": "handler/user.go:42",
  "trace_id": "abc123def456",
  "span_id": "789ghi",
  "request_id": "req-001",
  "method": "GET",
  "path": "/api/users/123",
  "status": 200,
  "latency": "25ms",
  "client_ip": "192.168.1.100",
  "user_id": "user-456",
  "service": "user-service",
  "version": "1.0.0",
  "env": "production"
}
```

## Kibana 使用

### 创建索引模式

1. 访问 Kibana: http://localhost:5601
2. 进入 Stack Management → Index Patterns
3. 创建索引模式: `logs-*`
4. 选择时间字段: `@timestamp`

### 常用查询

```
# KQL (Kibana Query Language)

# 按服务过滤
service: "order-service"

# 按日志级别
level: "error"

# 按时间范围
@timestamp >= "2024-01-01" and @timestamp < "2024-01-02"

# 组合查询
service: "order-service" and level: "error" and status >= 500

# 通配符
msg: *timeout*

# 追踪查询
trace_id: "abc123def456"

# 用户相关日志
user_id: "user-123"
```

### 创建可视化

#### 错误率趋势图

```json
{
  "visualization": {
    "type": "line",
    "params": {
      "index": "logs-*",
      "query": "level: error",
      "time_field": "@timestamp",
      "aggregation": "count"
    }
  }
}
```

#### 服务日志分布

```json
{
  "visualization": {
    "type": "pie",
    "params": {
      "index": "logs-*",
      "aggregation": "terms",
      "field": "service"
    }
  }
}
```

## Loki 替代方案

Grafana Loki 是 ELK 的轻量级替代方案。

### Docker Compose

```yaml
version: '3.8'

services:
  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - loki

  promtail:
    image: grafana/promtail:2.9.0
    volumes:
      - /var/log:/var/log
      - ./promtail-config.yml:/etc/promtail/config.yml
    command: -config.file=/etc/promtail/config.yml
    networks:
      - loki

  grafana:
    image: grafana/grafana:10.0.0
    ports:
      - "3000:3000"
    environment:
      - GF_PATHS_PROVISIONING=/etc/grafana/provisioning
    volumes:
      - ./grafana-datasources.yml:/etc/grafana/provisioning/datasources/datasources.yml
    networks:
      - loki

networks:
  loki:
    driver: bridge
```

### Go 集成 Loki

```go
package logger

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// LokiWriter 发送日志到 Loki
type LokiWriter struct {
    client   *http.Client
    endpoint string
    labels   map[string]string
    batch    []lokiEntry
    maxBatch int
}

type lokiEntry struct {
    Ts   string `json:"ts"`
    Line string `json:"line"`
}

type lokiPushRequest struct {
    Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
    Stream map[string]string `json:"stream"`
    Values [][]string        `json:"values"`
}

func NewLokiWriter(endpoint string, labels map[string]string) *LokiWriter {
    return &LokiWriter{
        client:   &http.Client{Timeout: 5 * time.Second},
        endpoint: endpoint + "/loki/api/v1/push",
        labels:   labels,
        maxBatch: 100,
    }
}

func (w *LokiWriter) Write(p []byte) (n int, err error) {
    ts := time.Now().UnixNano()

    values := [][]string{
        {fmt.Sprintf("%d", ts), string(p)},
    }

    req := lokiPushRequest{
        Streams: []lokiStream{
            {
                Stream: w.labels,
                Values: values,
            },
        },
    }

    body, _ := json.Marshal(req)
    resp, err := w.client.Post(w.endpoint, "application/json", bytes.NewReader(body))
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    return len(p), nil
}

func (w *LokiWriter) Sync() error {
    return nil
}
```

## 最佳实践

### 1. 日志格式

```go
// ✅ 好的日志：结构化、有上下文
logger.Info("Order created",
    zap.String("order_id", orderID),
    zap.String("user_id", userID),
    zap.Float64("amount", amount),
    zap.String("trace_id", traceID),
)

// ❌ 不好的日志：无结构、难以查询
logger.Info(fmt.Sprintf("Order %s created by user %s for $%.2f", orderID, userID, amount))
```

### 2. 日志级别

```go
// DEBUG: 详细调试信息
logger.Debug("SQL query executed", zap.String("query", query))

// INFO: 正常业务事件
logger.Info("Order created", zap.String("order_id", orderID))

// WARN: 潜在问题
logger.Warn("Slow query detected", zap.Duration("duration", duration))

// ERROR: 需要关注的错误
logger.Error("Failed to process payment", zap.Error(err))
```

### 3. 索引策略

```
# 按时间分割索引
logs-2024.01.01
logs-2024.01.02

# 按服务分割
logs-order-service-2024.01
logs-user-service-2024.01

# 按环境分割
logs-prod-2024.01
logs-staging-2024.01
```

### 4. 保留策略

```yaml
# Elasticsearch ILM 策略
{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_size": "50GB",
            "max_age": "1d"
          }
        }
      },
      "warm": {
        "min_age": "7d",
        "actions": {
          "shrink": {
            "number_of_shards": 1
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
```

**下一节**：[练习](./exercises.md) - 可观测性实践练习
