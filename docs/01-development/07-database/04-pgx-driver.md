# pgx 驱动 (High-Performance PostgreSQL Driver)

## 概述

pgx 是一个高性能的 PostgreSQL 驱动，相比 lib/pq 提供更好的性能和更多特性。

## pgx vs lib/pq

| 特性 | lib/pq | pgx |
|------|--------|-----|
| 性能 | 标准 | 更快（原生协议） |
| 维护状态 | 维护模式 | 活跃开发 |
| 连接池 | 无（依赖 database/sql） | 内置 pgxpool |
| 批量操作 | 无 | 原生 Batch |
| COPY 协议 | 有限支持 | 完整支持 |
| 类型支持 | 基础 | 扩展（UUID、JSON、数组等） |
| database/sql 兼容 | 是 | 是（通过 pgx/stdlib） |

## 安装

```bash
go get github.com/jackc/pgx/v5
```

## 两种使用方式

### 方式 1：原生 API（推荐）

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    ctx := context.Background()
    connStr := "postgres://user:password@localhost:5432/testdb"

    // 创建连接池
    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // 查询
    var name string
    err = pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", 1).Scan(&name)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Name:", name)
}
```

### 方式 2：database/sql 兼容

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/jackc/pgx/v5/stdlib"  // pgx database/sql 驱动
)

func main() {
    connStr := "postgres://user:password@localhost:5432/testdb"

    db, err := sql.Open("pgx", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    var name string
    err = db.QueryRow("SELECT name FROM users WHERE id = $1", 1).Scan(&name)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Name:", name)
}
```

## 连接池 (pgxpool)

### 基本配置

```go
package main

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

func setupPool(ctx context.Context) (*pgxpool.Pool, error) {
    connStr := "postgres://user:password@localhost:5432/testdb"

    // 解析配置
    config, err := pgxpool.ParseConfig(connStr)
    if err != nil {
        return nil, err
    }

    // 连接池配置
    config.MaxConns = 25                          // 最大连接数
    config.MinConns = 5                           // 最小连接数
    config.MaxConnLifetime = 1 * time.Hour        // 连接最大生命周期
    config.MaxConnIdleTime = 30 * time.Minute     // 空闲连接最大时间
    config.HealthCheckPeriod = 1 * time.Minute    // 健康检查周期

    // 连接配置
    config.ConnConfig.ConnectTimeout = 5 * time.Second

    // 创建连接池
    return pgxpool.NewWithConfig(ctx, config)
}
```

### 获取连接

```go
// 自动管理连接（推荐）
func queryWithPool(ctx context.Context, pool *pgxpool.Pool) error {
    rows, err := pool.Query(ctx, "SELECT id, name FROM users")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            return err
        }
        fmt.Printf("ID: %d, Name: %s\n", id, name)
    }

    return rows.Err()
}

// 手动获取连接（需要特殊操作时）
func queryWithConn(ctx context.Context, pool *pgxpool.Pool) error {
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()  // 归还连接

    // 使用连接
    _, err = conn.Exec(ctx, "SET search_path = myschema")
    if err != nil {
        return err
    }

    // 更多操作...
    return nil
}
```

## 查询操作

### Query（多行）

```go
func listUsers(ctx context.Context, pool *pgxpool.Pool) ([]User, error) {
    rows, err := pool.Query(ctx, "SELECT id, name, email FROM users WHERE status = $1", "active")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    return users, rows.Err()
}
```

### QueryRow（单行）

```go
func getUserByID(ctx context.Context, pool *pgxpool.Pool, id int) (*User, error) {
    var u User
    err := pool.QueryRow(ctx,
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name, &u.Email)

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil  // 未找到
        }
        return nil, err
    }

    return &u, nil
}
```

### Exec（INSERT/UPDATE/DELETE）

```go
func createUser(ctx context.Context, pool *pgxpool.Pool, name, email string) (int, error) {
    var id int
    err := pool.QueryRow(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        name, email,
    ).Scan(&id)

    return id, err
}

func updateUser(ctx context.Context, pool *pgxpool.Pool, id int, name string) error {
    tag, err := pool.Exec(ctx,
        "UPDATE users SET name = $1 WHERE id = $2",
        name, id,
    )
    if err != nil {
        return err
    }

    if tag.RowsAffected() == 0 {
        return ErrUserNotFound
    }

    return nil
}
```

### CollectRows（简化查询）

```go
import "github.com/jackc/pgx/v5"

func listUsersSimple(ctx context.Context, pool *pgxpool.Pool) ([]User, error) {
    rows, err := pool.Query(ctx, "SELECT id, name, email FROM users")
    if err != nil {
        return nil, err
    }

    // CollectRows 自动处理迭代和关闭
    return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}

// 或使用匿名函数
func listUsersCustom(ctx context.Context, pool *pgxpool.Pool) ([]User, error) {
    rows, err := pool.Query(ctx, "SELECT id, name, email FROM users")
    if err != nil {
        return nil, err
    }

    return pgx.CollectRows(rows, func(row pgx.CollectableRow) (User, error) {
        var u User
        err := row.Scan(&u.ID, &u.Name, &u.Email)
        return u, err
    })
}
```

## 批量操作 (Batch)

### 批量查询

```go
func batchQueries(ctx context.Context, pool *pgxpool.Pool) error {
    batch := &pgx.Batch{}

    // 添加多个查询
    batch.Queue("SELECT COUNT(*) FROM users")
    batch.Queue("SELECT COUNT(*) FROM posts")
    batch.Queue("SELECT MAX(created_at) FROM users")

    // 发送批量请求
    br := pool.SendBatch(ctx, batch)
    defer br.Close()

    // 获取结果
    var userCount, postCount int
    var latestUser time.Time

    if err := br.QueryRow().Scan(&userCount); err != nil {
        return err
    }
    if err := br.QueryRow().Scan(&postCount); err != nil {
        return err
    }
    if err := br.QueryRow().Scan(&latestUser); err != nil {
        return err
    }

    fmt.Printf("Users: %d, Posts: %d, Latest: %v\n", userCount, postCount, latestUser)
    return nil
}
```

### 批量插入

```go
func batchInsert(ctx context.Context, pool *pgxpool.Pool, users []User) error {
    batch := &pgx.Batch{}

    for _, u := range users {
        batch.Queue(
            "INSERT INTO users (name, email) VALUES ($1, $2)",
            u.Name, u.Email,
        )
    }

    br := pool.SendBatch(ctx, batch)
    defer br.Close()

    // 检查每个操作的结果
    for range users {
        _, err := br.Exec()
        if err != nil {
            return err
        }
    }

    return nil
}
```

## COPY 协议（高性能批量导入）

### CopyFrom

```go
func bulkInsertUsers(ctx context.Context, pool *pgxpool.Pool, users []User) error {
    // 准备数据
    rows := make([][]interface{}, len(users))
    for i, u := range users {
        rows[i] = []interface{}{u.Name, u.Email}
    }

    // 使用 COPY 协议
    count, err := pool.CopyFrom(
        ctx,
        pgx.Identifier{"users"},           // 表名
        []string{"name", "email"},          // 列名
        pgx.CopyFromRows(rows),             // 数据源
    )

    if err != nil {
        return err
    }

    fmt.Printf("Inserted %d rows\n", count)
    return nil
}

// 使用 CopyFromSlice（更高效）
func bulkInsertWithSlice(ctx context.Context, pool *pgxpool.Pool, users []User) error {
    count, err := pool.CopyFrom(
        ctx,
        pgx.Identifier{"users"},
        []string{"name", "email"},
        pgx.CopyFromSlice(len(users), func(i int) ([]interface{}, error) {
            return []interface{}{users[i].Name, users[i].Email}, nil
        }),
    )

    if err != nil {
        return err
    }

    fmt.Printf("Inserted %d rows\n", count)
    return nil
}
```

## 事务

### 基本事务

```go
func transferMoney(ctx context.Context, pool *pgxpool.Pool, from, to int, amount float64) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)  // 如果已提交，Rollback 是空操作

    // 扣款
    tag, err := tx.Exec(ctx,
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
        amount, from,
    )
    if err != nil {
        return err
    }
    if tag.RowsAffected() == 0 {
        return errors.New("insufficient balance")
    }

    // 入账
    _, err = tx.Exec(ctx,
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, to,
    )
    if err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

### 使用 BeginFunc（推荐）

```go
func transferWithBeginFunc(ctx context.Context, pool *pgxpool.Pool, from, to int, amount float64) error {
    return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
        // 事务成功自动提交，失败自动回滚
        tag, err := tx.Exec(ctx,
            "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
            amount, from,
        )
        if err != nil {
            return err
        }
        if tag.RowsAffected() == 0 {
            return errors.New("insufficient balance")
        }

        _, err = tx.Exec(ctx,
            "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
            amount, to,
        )
        return err
    })
}
```

### 带选项的事务

```go
func readOnlyTransaction(ctx context.Context, pool *pgxpool.Pool) error {
    tx, err := pool.BeginTx(ctx, pgx.TxOptions{
        IsoLevel:   pgx.Serializable,
        AccessMode: pgx.ReadOnly,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // 只读查询
    // ...

    return tx.Commit(ctx)
}
```

## 类型支持

### UUID

```go
import "github.com/google/uuid"

type User struct {
    ID   uuid.UUID
    Name string
}

func createUserWithUUID(ctx context.Context, pool *pgxpool.Pool) error {
    id := uuid.New()

    _, err := pool.Exec(ctx,
        "INSERT INTO users (id, name) VALUES ($1, $2)",
        id, "Alice",
    )

    return err
}

func getUser(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*User, error) {
    var u User
    err := pool.QueryRow(ctx,
        "SELECT id, name FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name)

    return &u, err
}
```

### JSON/JSONB

```go
type Settings struct {
    Theme    string `json:"theme"`
    Language string `json:"language"`
}

type User struct {
    ID       int
    Name     string
    Settings Settings  // JSONB 列
}

func saveUser(ctx context.Context, pool *pgxpool.Pool, u *User) error {
    _, err := pool.Exec(ctx,
        "INSERT INTO users (name, settings) VALUES ($1, $2)",
        u.Name, u.Settings,  // Settings 自动序列化为 JSON
    )
    return err
}

func getUser(ctx context.Context, pool *pgxpool.Pool, id int) (*User, error) {
    var u User
    err := pool.QueryRow(ctx,
        "SELECT id, name, settings FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name, &u.Settings)  // 自动反序列化

    return &u, err
}
```

### 数组

```go
func insertTags(ctx context.Context, pool *pgxpool.Pool, postID int, tags []string) error {
    _, err := pool.Exec(ctx,
        "INSERT INTO posts (id, tags) VALUES ($1, $2)",
        postID, tags,  // []string 自动转为 PostgreSQL 数组
    )
    return err
}

func getPostTags(ctx context.Context, pool *pgxpool.Pool, postID int) ([]string, error) {
    var tags []string
    err := pool.QueryRow(ctx,
        "SELECT tags FROM posts WHERE id = $1", postID,
    ).Scan(&tags)

    return tags, err
}
```

### hstore

```go
import "github.com/jackc/pgx/v5/pgtype"

func insertMeta(ctx context.Context, pool *pgxpool.Pool) error {
    meta := pgtype.Hstore{
        "key1": pgtype.Text{String: "value1", Valid: true},
        "key2": pgtype.Text{String: "value2", Valid: true},
    }

    _, err := pool.Exec(ctx,
        "INSERT INTO items (id, metadata) VALUES ($1, $2)",
        1, meta,
    )
    return err
}
```

## 预处理语句

```go
func preparedQueries(ctx context.Context, pool *pgxpool.Pool) error {
    // 获取连接以使用预处理语句
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()

    // 准备语句
    _, err = conn.Conn().Prepare(ctx, "get_user", "SELECT id, name FROM users WHERE id = $1")
    if err != nil {
        return err
    }

    // 使用预处理语句
    rows, err := conn.Query(ctx, "get_user", 1)
    if err != nil {
        return err
    }
    defer rows.Close()

    // 处理结果...
    return nil
}
```

## 连接池统计

```go
func printPoolStats(pool *pgxpool.Pool) {
    stat := pool.Stat()

    fmt.Printf("Total connections: %d\n", stat.TotalConns())
    fmt.Printf("Acquired connections: %d\n", stat.AcquiredConns())
    fmt.Printf("Idle connections: %d\n", stat.IdleConns())
    fmt.Printf("Max connections: %d\n", stat.MaxConns())
    fmt.Printf("Connections acquired count: %d\n", stat.AcquireCount())
    fmt.Printf("Connections acquired duration: %v\n", stat.AcquireDuration())
}
```

## 错误处理

```go
import "github.com/jackc/pgx/v5/pgconn"

func handlePgxError(err error) error {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505":  // unique_violation
            return ErrDuplicate
        case "23503":  // foreign_key_violation
            return ErrForeignKeyViolation
        case "23502":  // not_null_violation
            return ErrNullViolation
        default:
            return fmt.Errorf("database error: %s (code: %s)", pgErr.Message, pgErr.Code)
        }
    }

    if errors.Is(err, pgx.ErrNoRows) {
        return ErrNotFound
    }

    return err
}
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 使用 pgxpool | 生产环境使用连接池 |
| 使用 BeginFunc | 自动管理事务提交/回滚 |
| 使用 Batch | 减少网络往返 |
| 使用 CopyFrom | 大量数据导入 |
| 使用 CollectRows | 简化查询代码 |
| 设置超时 | 使用带超时的 Context |

**下一节**：[连接池](./05-connection-pool.md) - 学习连接池配置和优化
