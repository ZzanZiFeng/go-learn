# 连接池配置 (Connection Pool)

## 概述

数据库连接池是管理和复用数据库连接的机制，对于高并发应用至关重要。

## 为什么需要连接池？

```
❌ 没有连接池:
    请求1 → 建立连接 → 查询 → 关闭连接 (耗时)
    请求2 → 建立连接 → 查询 → 关闭连接 (耗时)
    请求3 → 建立连接 → 查询 → 关闭连接 (耗时)

✅ 有连接池:
    启动 → 预创建连接池 [conn1, conn2, conn3, ...]
    请求1 → 获取 conn1 → 查询 → 归还 conn1
    请求2 → 获取 conn2 → 查询 → 归还 conn2
    请求3 → 获取 conn1 → 查询 → 归还 conn1  (复用)
```

## database/sql 连接池

### 核心参数

```go
db, err := sql.Open("postgres", connStr)
if err != nil {
    return nil, err
}

// MaxOpenConns: 最大打开连接数
// - 默认: 0 (无限制)
// - 建议: 根据数据库服务器配置和应用负载设置
db.SetMaxOpenConns(25)

// MaxIdleConns: 最大空闲连接数
// - 默认: 2
// - 建议: 设置为 MaxOpenConns 的一部分 (如 25%)
db.SetMaxIdleConns(5)

// ConnMaxLifetime: 连接最大生命周期
// - 默认: 0 (无限制)
// - 建议: 5-30 分钟，小于数据库 wait_timeout
db.SetConnMaxLifetime(5 * time.Minute)

// ConnMaxIdleTime: 空闲连接最大存活时间 (Go 1.15+)
// - 默认: 0 (无限制)
// - 建议: 1-5 分钟
db.SetConnMaxIdleTime(1 * time.Minute)
```

### 参数说明

```
┌─────────────────────────────────────────────────────────┐
│                      连接池                               │
│  ┌─────────────────────────────────────────────────┐    │
│  │    活跃连接 (正在使用)                             │    │
│  │    [conn1] [conn2] [conn3] ...                  │    │
│  │         ← MaxOpenConns 限制总数 →                │    │
│  └─────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────┐    │
│  │    空闲连接 (等待复用)                             │    │
│  │    [conn4] [conn5]                              │    │
│  │    ← MaxIdleConns 限制空闲数 →                   │    │
│  │    ← ConnMaxIdleTime 过期清理 →                  │    │
│  └─────────────────────────────────────────────────┘    │
│                                                         │
│  所有连接: ← ConnMaxLifetime 过期清理 →                  │
└─────────────────────────────────────────────────────────┘
```

### 配置建议

| 场景 | MaxOpenConns | MaxIdleConns | ConnMaxLifetime |
|------|-------------|--------------|-----------------|
| 低负载（开发） | 10 | 2 | 10m |
| 中等负载 | 25 | 5 | 5m |
| 高负载 | 50-100 | 10-25 | 3m |
| 微服务（多实例） | 10-25 | 5 | 5m |

### 监控连接池状态

```go
func monitorDBStats(db *sql.DB) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stats := db.Stats()

        fmt.Printf("=== DB Pool Stats ===\n")
        fmt.Printf("MaxOpenConnections: %d\n", stats.MaxOpenConnections)
        fmt.Printf("OpenConnections: %d\n", stats.OpenConnections)
        fmt.Printf("InUse: %d\n", stats.InUse)
        fmt.Printf("Idle: %d\n", stats.Idle)
        fmt.Printf("WaitCount: %d\n", stats.WaitCount)
        fmt.Printf("WaitDuration: %v\n", stats.WaitDuration)
        fmt.Printf("MaxIdleClosed: %d\n", stats.MaxIdleClosed)
        fmt.Printf("MaxLifetimeClosed: %d\n", stats.MaxLifetimeClosed)
    }
}

// 使用
go monitorDBStats(db)
```

## pgxpool 连接池

### 基本配置

```go
import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

func setupPgxPool(ctx context.Context) (*pgxpool.Pool, error) {
    connStr := "postgres://user:password@localhost:5432/testdb"

    config, err := pgxpool.ParseConfig(connStr)
    if err != nil {
        return nil, err
    }

    // 连接池大小
    config.MaxConns = 25          // 最大连接数
    config.MinConns = 5           // 最小连接数 (预热)

    // 连接生命周期
    config.MaxConnLifetime = 1 * time.Hour          // 最大生命周期
    config.MaxConnIdleTime = 30 * time.Minute       // 空闲连接最大时间

    // 健康检查
    config.HealthCheckPeriod = 1 * time.Minute      // 健康检查周期

    // 连接超时
    config.ConnConfig.ConnectTimeout = 5 * time.Second

    return pgxpool.NewWithConfig(ctx, config)
}
```

### 高级配置

```go
func setupAdvancedPool(ctx context.Context) (*pgxpool.Pool, error) {
    connStr := "postgres://user:password@localhost:5432/testdb"

    config, err := pgxpool.ParseConfig(connStr)
    if err != nil {
        return nil, err
    }

    // 基本配置
    config.MaxConns = 25
    config.MinConns = 5

    // 连接回调 - 每个新连接建立时调用
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        // 设置 session 变量
        _, err := conn.Exec(ctx, "SET timezone = 'UTC'")
        if err != nil {
            return err
        }
        _, err = conn.Exec(ctx, "SET statement_timeout = '30s'")
        return err
    }

    // 连接获取前回调
    config.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
        // 返回 false 会丢弃该连接
        // 可用于检查连接是否有效
        return conn.Ping(ctx) == nil
    }

    // 连接释放后回调
    config.AfterRelease = func(conn *pgx.Conn) bool {
        // 返回 false 会关闭该连接
        // 可用于重置连接状态
        _, err := conn.Exec(context.Background(), "RESET ALL")
        return err == nil
    }

    // 连接关闭前回调
    config.BeforeClose = func(conn *pgx.Conn) {
        // 清理资源
        log.Println("Closing connection")
    }

    return pgxpool.NewWithConfig(ctx, config)
}
```

### 监控 pgxpool

```go
func monitorPgxPool(pool *pgxpool.Pool) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stat := pool.Stat()

        fmt.Printf("=== pgxpool Stats ===\n")
        fmt.Printf("TotalConns: %d\n", stat.TotalConns())
        fmt.Printf("AcquiredConns: %d\n", stat.AcquiredConns())
        fmt.Printf("IdleConns: %d\n", stat.IdleConns())
        fmt.Printf("MaxConns: %d\n", stat.MaxConns())
        fmt.Printf("AcquireCount: %d\n", stat.AcquireCount())
        fmt.Printf("AcquireDuration: %v\n", stat.AcquireDuration())
        fmt.Printf("CanceledAcquireCount: %d\n", stat.CanceledAcquireCount())
        fmt.Printf("ConstructingConns: %d\n", stat.ConstructingConns())
        fmt.Printf("EmptyAcquireCount: %d\n", stat.EmptyAcquireCount())
    }
}
```

## 连接池调优

### 确定最大连接数

```go
// PostgreSQL 默认 max_connections = 100
// 计算公式: max_conns_per_instance = (DB max_connections - reserved) / instances

// 示例:
// - PostgreSQL max_connections = 100
// - 保留连接 = 10 (管理、监控等)
// - 应用实例数 = 3
// - 每实例最大连接 = (100 - 10) / 3 = 30

const MaxConnsPerInstance = 30
```

### 避免连接泄露

```go
// ❌ 错误: 没有关闭 rows
func leakyQuery(db *sql.DB) error {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        return err
    }
    // 忘记 defer rows.Close()
    for rows.Next() {
        // ...
    }
    return nil
}

// ✅ 正确: 始终关闭 rows
func properQuery(db *sql.DB) error {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        return err
    }
    defer rows.Close()  // 确保关闭

    for rows.Next() {
        // ...
    }
    return rows.Err()
}
```

### 超时配置

```go
func queryWithTimeout(db *sql.DB) error {
    // 使用 context 设置超时
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    rows, err := db.QueryContext(ctx, "SELECT * FROM users")
    if err != nil {
        return err
    }
    defer rows.Close()

    // ...
    return nil
}

// 全局语句超时 (PostgreSQL)
func setStatementTimeout(db *sql.DB) error {
    _, err := db.Exec("SET statement_timeout = '30s'")
    return err
}
```

### 连接验证

```go
// database/sql: 使用 Ping
func validateConnection(db *sql.DB) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return db.PingContext(ctx)
}

// 健康检查端点
func healthCheck(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
        defer cancel()

        if err := db.PingContext(ctx); err != nil {
            http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
            return
        }

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }
}
```

## 连接池问题排查

### 常见问题

```
问题: "too many connections"
原因: MaxOpenConns 设置过大或连接泄露
解决:
  1. 降低 MaxOpenConns
  2. 检查是否有未关闭的 rows
  3. 检查长时间运行的查询

问题: "connection refused"
原因: 数据库不可达或连接数耗尽
解决:
  1. 检查数据库服务器状态
  2. 检查网络连接
  3. 增加数据库 max_connections

问题: 响应时间变长
原因: 连接池耗尽，请求在排队
解决:
  1. 查看 WaitCount 和 WaitDuration
  2. 增加 MaxOpenConns
  3. 优化慢查询
```

### 调试工具

```go
// Prometheus 指标导出
import "github.com/prometheus/client_golang/prometheus"

type DBPoolCollector struct {
    db *sql.DB

    maxOpen     *prometheus.Desc
    openConns   *prometheus.Desc
    inUse       *prometheus.Desc
    idle        *prometheus.Desc
    waitCount   *prometheus.Desc
    waitDuration *prometheus.Desc
}

func NewDBPoolCollector(db *sql.DB) *DBPoolCollector {
    return &DBPoolCollector{
        db: db,
        maxOpen: prometheus.NewDesc("db_max_open_connections",
            "Maximum number of open connections", nil, nil),
        openConns: prometheus.NewDesc("db_open_connections",
            "Number of open connections", nil, nil),
        inUse: prometheus.NewDesc("db_in_use_connections",
            "Number of connections currently in use", nil, nil),
        idle: prometheus.NewDesc("db_idle_connections",
            "Number of idle connections", nil, nil),
        waitCount: prometheus.NewDesc("db_wait_count_total",
            "Total number of connections waited for", nil, nil),
        waitDuration: prometheus.NewDesc("db_wait_duration_seconds_total",
            "Total time waited for connections", nil, nil),
    }
}

func (c *DBPoolCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.maxOpen
    ch <- c.openConns
    ch <- c.inUse
    ch <- c.idle
    ch <- c.waitCount
    ch <- c.waitDuration
}

func (c *DBPoolCollector) Collect(ch chan<- prometheus.Metric) {
    stats := c.db.Stats()

    ch <- prometheus.MustNewConstMetric(c.maxOpen, prometheus.GaugeValue,
        float64(stats.MaxOpenConnections))
    ch <- prometheus.MustNewConstMetric(c.openConns, prometheus.GaugeValue,
        float64(stats.OpenConnections))
    ch <- prometheus.MustNewConstMetric(c.inUse, prometheus.GaugeValue,
        float64(stats.InUse))
    ch <- prometheus.MustNewConstMetric(c.idle, prometheus.GaugeValue,
        float64(stats.Idle))
    ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.CounterValue,
        float64(stats.WaitCount))
    ch <- prometheus.MustNewConstMetric(c.waitDuration, prometheus.CounterValue,
        stats.WaitDuration.Seconds())
}

// 注册
func init() {
    prometheus.MustRegister(NewDBPoolCollector(db))
}
```

## 最佳实践总结

| 实践 | 说明 |
|------|------|
| 设置 MaxOpenConns | 避免耗尽数据库连接 |
| 设置 MaxIdleConns | 保持适量空闲连接 |
| 设置 ConnMaxLifetime | 定期刷新连接 |
| 监控连接池状态 | 及时发现问题 |
| 使用 Context 超时 | 避免长时间等待 |
| 始终关闭 Rows | 防止连接泄露 |
| 预热连接池 | 设置 MinConns (pgx) |

**下一节**：[预处理语句](./06-prepared-stmt.md) - 学习 SQL 注入防护
