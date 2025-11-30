# 预处理语句 (Prepared Statements)

## 概述

预处理语句是防止 SQL 注入攻击的关键技术，同时也能提高批量操作的性能。

## SQL 注入攻击

### 什么是 SQL 注入？

```go
// ❌ 危险代码 - SQL 注入漏洞
func unsafeLogin(db *sql.DB, username, password string) bool {
    query := fmt.Sprintf(
        "SELECT * FROM users WHERE username = '%s' AND password = '%s'",
        username, password,
    )

    row := db.QueryRow(query)
    // ...
    return true
}

// 攻击者输入:
// username: admin' --
// password: anything

// 生成的 SQL:
// SELECT * FROM users WHERE username = 'admin' --' AND password = 'anything'
// -- 注释掉了密码检查，直接登录为 admin

// 更严重的攻击:
// username: '; DROP TABLE users; --
// 生成: SELECT * FROM users WHERE username = ''; DROP TABLE users; --'
```

### 参数化查询防护

```go
// ✅ 安全代码 - 参数化查询
func safeLogin(db *sql.DB, username, password string) bool {
    // $1, $2 是占位符，参数值在执行时绑定
    query := "SELECT * FROM users WHERE username = $1 AND password = $2"

    row := db.QueryRow(query, username, password)
    // ...
    return true
}

// 即使输入 "admin' --"，也会被当作普通字符串处理
// 数据库驱动会正确转义特殊字符
```

## 占位符语法

不同数据库使用不同的占位符：

| 数据库 | 占位符 | 示例 |
|--------|--------|------|
| PostgreSQL | $1, $2, $3... | `WHERE id = $1` |
| MySQL | ? | `WHERE id = ?` |
| SQLite | ?, $1, :name | `WHERE id = ?` |
| SQL Server | @p1, @p2... | `WHERE id = @p1` |

```go
// PostgreSQL
db.Query("SELECT * FROM users WHERE id = $1 AND status = $2", id, status)

// MySQL
db.Query("SELECT * FROM users WHERE id = ? AND status = ?", id, status)
```

## database/sql 预处理

### 即时参数化（推荐大多数场景）

```go
// 每次调用自动处理参数化
func getUser(db *sql.DB, id int) (*User, error) {
    var u User
    err := db.QueryRow(
        "SELECT id, name, email FROM users WHERE id = $1",
        id,  // 参数绑定
    ).Scan(&u.ID, &u.Name, &u.Email)

    return &u, err
}

func createUser(db *sql.DB, name, email string) error {
    _, err := db.Exec(
        "INSERT INTO users (name, email) VALUES ($1, $2)",
        name, email,  // 参数绑定
    )
    return err
}
```

### 显式预处理（批量操作优化）

```go
// 使用 Prepare 显式创建预处理语句
func batchInsertUsers(db *sql.DB, users []User) error {
    // 准备语句（只编译一次）
    stmt, err := db.Prepare("INSERT INTO users (name, email) VALUES ($1, $2)")
    if err != nil {
        return err
    }
    defer stmt.Close()  // 必须关闭！

    // 重复使用预处理语句
    for _, u := range users {
        _, err := stmt.Exec(u.Name, u.Email)
        if err != nil {
            return err
        }
    }

    return nil
}
```

### 带 Context 的预处理

```go
func batchInsertWithContext(ctx context.Context, db *sql.DB, users []User) error {
    stmt, err := db.PrepareContext(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, u := range users {
        select {
        case <-ctx.Done():
            return ctx.Err()  // 支持取消
        default:
            if _, err := stmt.ExecContext(ctx, u.Name, u.Email); err != nil {
                return err
            }
        }
    }

    return nil
}
```

## pgx 预处理

### 自动预处理

```go
import "github.com/jackc/pgx/v5/pgxpool"

// pgx 自动处理参数化
func getUser(ctx context.Context, pool *pgxpool.Pool, id int) (*User, error) {
    var u User
    err := pool.QueryRow(ctx,
        "SELECT id, name, email FROM users WHERE id = $1",
        id,
    ).Scan(&u.ID, &u.Name, &u.Email)

    return &u, err
}
```

### 显式预处理

```go
func batchInsertPgx(ctx context.Context, pool *pgxpool.Pool, users []User) error {
    // 获取连接
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()

    // 准备语句
    _, err = conn.Conn().Prepare(ctx, "insert_user",
        "INSERT INTO users (name, email) VALUES ($1, $2)")
    if err != nil {
        return err
    }

    // 使用预处理语句
    for _, u := range users {
        _, err := conn.Exec(ctx, "insert_user", u.Name, u.Email)
        if err != nil {
            return err
        }
    }

    return nil
}
```

## 安全最佳实践

### 1. 永远使用参数化查询

```go
// ✅ 正确
db.Query("SELECT * FROM users WHERE name = $1", name)

// ❌ 错误
db.Query(fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name))
```

### 2. 动态列名/表名处理

```go
// 列名和表名不能参数化
// ❌ 这不会工作
db.Query("SELECT * FROM $1 WHERE $2 = $3", table, column, value)

// ✅ 使用白名单验证
var allowedColumns = map[string]bool{
    "name":       true,
    "email":      true,
    "created_at": true,
}

func safeQuery(db *sql.DB, column string, value interface{}) (*sql.Rows, error) {
    // 验证列名
    if !allowedColumns[column] {
        return nil, errors.New("invalid column name")
    }

    // 列名通过字符串拼接（已验证安全）
    query := fmt.Sprintf("SELECT * FROM users WHERE %s = $1", column)
    return db.Query(query, value)
}
```

### 3. 动态 ORDER BY 处理

```go
var allowedSortColumns = map[string]bool{
    "id":         true,
    "name":       true,
    "created_at": true,
}

var allowedSortOrders = map[string]bool{
    "ASC":  true,
    "DESC": true,
}

func safeOrderBy(db *sql.DB, column, order string) (*sql.Rows, error) {
    // 验证
    if !allowedSortColumns[column] {
        column = "id"  // 默认值
    }
    order = strings.ToUpper(order)
    if !allowedSortOrders[order] {
        order = "ASC"
    }

    query := fmt.Sprintf("SELECT * FROM users ORDER BY %s %s", column, order)
    return db.Query(query)
}
```

### 4. IN 子句处理

```go
// ❌ 错误 - 不能直接传数组
db.Query("SELECT * FROM users WHERE id IN ($1)", ids)

// ✅ 正确 - 动态构建占位符
func getUsersByIDs(db *sql.DB, ids []int) ([]*User, error) {
    if len(ids) == 0 {
        return nil, nil
    }

    // 构建占位符 $1, $2, $3...
    placeholders := make([]string, len(ids))
    args := make([]interface{}, len(ids))
    for i, id := range ids {
        placeholders[i] = fmt.Sprintf("$%d", i+1)
        args[i] = id
    }

    query := fmt.Sprintf(
        "SELECT id, name, email FROM users WHERE id IN (%s)",
        strings.Join(placeholders, ", "),
    )

    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, &u)
    }

    return users, rows.Err()
}

// 使用 pgx 的 ANY
func getUsersByIDsPgx(ctx context.Context, pool *pgxpool.Pool, ids []int) ([]*User, error) {
    rows, err := pool.Query(ctx,
        "SELECT id, name, email FROM users WHERE id = ANY($1)",
        ids,  // pgx 原生支持数组
    )
    // ...
}
```

### 5. LIKE 子句处理

```go
// ✅ 正确处理 LIKE 中的特殊字符
func searchUsers(db *sql.DB, pattern string) ([]*User, error) {
    // 转义特殊字符
    pattern = strings.ReplaceAll(pattern, "%", "\\%")
    pattern = strings.ReplaceAll(pattern, "_", "\\_")

    // 添加通配符
    pattern = "%" + pattern + "%"

    rows, err := db.Query(
        "SELECT id, name FROM users WHERE name LIKE $1 ESCAPE '\\'",
        pattern,
    )
    // ...
}
```

## 预处理语句性能

### 什么时候使用显式 Prepare

```
使用 Prepare 的场景:
✅ 同一语句执行多次（批量操作）
✅ 高频执行的查询
✅ 需要更好的执行计划缓存

不需要 Prepare 的场景:
- 一次性查询
- 低频操作
- 简单的 CRUD
```

### 性能对比

```go
// 测试代码
func benchmarkNoPrepare(b *testing.B, db *sql.DB) {
    for i := 0; i < b.N; i++ {
        db.Exec("INSERT INTO test (value) VALUES ($1)", i)
    }
}

func benchmarkWithPrepare(b *testing.B, db *sql.DB) {
    stmt, _ := db.Prepare("INSERT INTO test (value) VALUES ($1)")
    defer stmt.Close()

    for i := 0; i < b.N; i++ {
        stmt.Exec(i)
    }
}

// 结果示例:
// BenchmarkNoPrepare-8      10000    150000 ns/op
// BenchmarkWithPrepare-8    20000     80000 ns/op
```

### 批量操作优化

```go
// 1. 使用预处理语句
func batchInsertPrepared(db *sql.DB, items []Item) error {
    stmt, err := db.Prepare("INSERT INTO items (name, value) VALUES ($1, $2)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, item := range items {
        if _, err := stmt.Exec(item.Name, item.Value); err != nil {
            return err
        }
    }
    return nil
}

// 2. 使用批量 INSERT（更快）
func batchInsertBulk(db *sql.DB, items []Item) error {
    if len(items) == 0 {
        return nil
    }

    // 构建 VALUES 子句
    valueStrings := make([]string, 0, len(items))
    valueArgs := make([]interface{}, 0, len(items)*2)

    for i, item := range items {
        valueStrings = append(valueStrings,
            fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
        valueArgs = append(valueArgs, item.Name, item.Value)
    }

    query := fmt.Sprintf(
        "INSERT INTO items (name, value) VALUES %s",
        strings.Join(valueStrings, ", "),
    )

    _, err := db.Exec(query, valueArgs...)
    return err
}

// 3. 使用 pgx CopyFrom（最快）
func batchInsertCopy(ctx context.Context, pool *pgxpool.Pool, items []Item) error {
    _, err := pool.CopyFrom(
        ctx,
        pgx.Identifier{"items"},
        []string{"name", "value"},
        pgx.CopyFromSlice(len(items), func(i int) ([]interface{}, error) {
            return []interface{}{items[i].Name, items[i].Value}, nil
        }),
    )
    return err
}
```

## 常见错误

### 1. 预处理语句泄露

```go
// ❌ 错误 - 没有关闭语句
func leakyPrepare(db *sql.DB) {
    stmt, _ := db.Prepare("SELECT * FROM users WHERE id = $1")
    // 忘记 stmt.Close()
    stmt.Query(1)
}

// ✅ 正确
func properPrepare(db *sql.DB) {
    stmt, err := db.Prepare("SELECT * FROM users WHERE id = $1")
    if err != nil {
        return
    }
    defer stmt.Close()

    // ...
}
```

### 2. 在循环中创建预处理语句

```go
// ❌ 错误 - 性能差
for _, id := range ids {
    stmt, _ := db.Prepare("SELECT * FROM users WHERE id = $1")
    stmt.Query(id)
    stmt.Close()
}

// ✅ 正确 - 在循环外创建
stmt, _ := db.Prepare("SELECT * FROM users WHERE id = $1")
defer stmt.Close()

for _, id := range ids {
    stmt.Query(id)
}
```

### 3. 事务中的预处理语句

```go
// ✅ 正确 - 在事务中创建预处理语句
func transactionWithPrepare(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 在事务中创建预处理语句
    stmt, err := tx.Prepare("INSERT INTO users (name) VALUES ($1)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, name := range []string{"Alice", "Bob"} {
        if _, err := stmt.Exec(name); err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

## 安全检查清单

- [ ] 所有 SQL 查询使用参数化
- [ ] 动态表名/列名使用白名单验证
- [ ] LIKE 查询转义特殊字符
- [ ] IN 子句正确处理数组
- [ ] 预处理语句正确关闭
- [ ] 输入数据进行长度限制

**下一节**：[事务处理](./07-transactions.md) - 学习数据库事务管理
