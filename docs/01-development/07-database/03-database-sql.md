# database/sql 标准库

## 概述

`database/sql` 是 Go 标准库提供的数据库接口，提供统一的 API 来操作不同的数据库。

## 与 JavaScript 对比

### Node.js

```javascript
// pg (node-postgres)
const { Pool } = require('pg');
const pool = new Pool({ connectionString: DATABASE_URL });

// 查询
const result = await pool.query('SELECT * FROM users WHERE id = $1', [1]);
console.log(result.rows);

// 释放连接自动管理
```

### Go

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, _ := sql.Open("postgres", DATABASE_URL)
defer db.Close()

// 查询
rows, _ := db.Query("SELECT * FROM users WHERE id = $1", 1)
defer rows.Close()

for rows.Next() {
    // 手动扫描到变量
}
```

## 架构

```
┌─────────────────────────────────────────────┐
│                应用代码                       │
├─────────────────────────────────────────────┤
│             database/sql                     │
│  ┌─────────────────────────────────────┐    │
│  │           *sql.DB                   │    │
│  │    (连接池 + 驱动接口封装)            │    │
│  └─────────────────────────────────────┘    │
├─────────────────────────────────────────────┤
│         driver.Driver 接口                   │
│    lib/pq | pgx | mysql | sqlite3           │
├─────────────────────────────────────────────┤
│              数据库服务器                     │
└─────────────────────────────────────────────┘
```

## 安装驱动

```bash
# PostgreSQL
go get github.com/lib/pq             # 标准驱动
go get github.com/jackc/pgx/v5       # 高性能驱动

# MySQL
go get github.com/go-sql-driver/mysql

# SQLite
go get modernc.org/sqlite            # 纯 Go 实现
go get github.com/mattn/go-sqlite3   # CGO 实现
```

## 连接数据库

### 基本连接

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"  // 导入驱动（副作用导入）
)

func main() {
    // 连接字符串
    connStr := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

    // sql.Open 不会真正连接，只是验证参数
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Failed to open:", err)
    }
    defer db.Close()  // 程序结束时关闭

    // Ping 才会真正连接
    if err := db.Ping(); err != nil {
        log.Fatal("Failed to connect:", err)
    }

    fmt.Println("Connected to database!")
}
```

### 配置连接池

```go
func setupDB() (*sql.DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    // 最大打开连接数
    db.SetMaxOpenConns(25)

    // 最大空闲连接数
    db.SetMaxIdleConns(5)

    // 连接最大生命周期
    db.SetConnMaxLifetime(5 * time.Minute)

    // 空闲连接最大生命周期
    db.SetConnMaxIdleTime(1 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

### 带 Context 连接

```go
func connectWithContext() error {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return err
    }
    defer db.Close()

    // 带超时的 Ping
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return db.PingContext(ctx)
}
```

## 查询数据

### 查询多行 (Query)

```go
func listUsers(db *sql.DB) ([]User, error) {
    rows, err := db.Query("SELECT id, name, email FROM users WHERE status = $1", "active")
    if err != nil {
        return nil, err
    }
    defer rows.Close()  // 必须关闭！

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    // 检查迭代过程中的错误
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}
```

### 查询单行 (QueryRow)

```go
func getUserByID(db *sql.DB, id int) (*User, error) {
    row := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id)

    var u User
    err := row.Scan(&u.ID, &u.Name, &u.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil  // 未找到
        }
        return nil, err
    }

    return &u, nil
}
```

### 带 Context 查询

```go
func getUserWithContext(ctx context.Context, db *sql.DB, id int) (*User, error) {
    row := db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = $1", id)

    var u User
    err := row.Scan(&u.ID, &u.Name, &u.Email)
    if err != nil {
        return nil, err
    }

    return &u, nil
}

// 使用
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

user, err := getUserWithContext(ctx, db, 1)
```

## 执行语句

### Exec（INSERT, UPDATE, DELETE）

```go
func createUser(db *sql.DB, name, email string) (int64, error) {
    result, err := db.Exec(
        "INSERT INTO users (name, email) VALUES ($1, $2)",
        name, email,
    )
    if err != nil {
        return 0, err
    }

    // 获取影响的行数
    rowsAffected, _ := result.RowsAffected()
    fmt.Printf("Rows affected: %d\n", rowsAffected)

    // PostgreSQL 不支持 LastInsertId，需要用 RETURNING
    return rowsAffected, nil
}

// PostgreSQL 获取插入 ID
func createUserReturningID(db *sql.DB, name, email string) (int, error) {
    var id int
    err := db.QueryRow(
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        name, email,
    ).Scan(&id)
    return id, err
}
```

### 更新

```go
func updateUser(db *sql.DB, id int, name string) error {
    result, err := db.Exec(
        "UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2",
        name, id,
    )
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return ErrUserNotFound
    }

    return nil
}
```

### 删除

```go
func deleteUser(db *sql.DB, id int) error {
    result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return ErrUserNotFound
    }

    return nil
}

// 软删除
func softDeleteUser(db *sql.DB, id int) error {
    _, err := db.Exec(
        "UPDATE users SET deleted_at = NOW() WHERE id = $1",
        id,
    )
    return err
}
```

## 处理 NULL 值

### sql.Null 类型

```go
import "database/sql"

type User struct {
    ID        int
    Name      string
    Email     sql.NullString   // 可能为 NULL
    Phone     sql.NullString
    Age       sql.NullInt64
    Balance   sql.NullFloat64
    IsActive  sql.NullBool
    CreatedAt sql.NullTime
}

func getUser(db *sql.DB, id int) (*User, error) {
    var u User
    err := db.QueryRow(
        "SELECT id, name, email, phone, age FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Age)

    if err != nil {
        return nil, err
    }

    // 检查是否有效
    if u.Email.Valid {
        fmt.Println("Email:", u.Email.String)
    } else {
        fmt.Println("Email is NULL")
    }

    return &u, nil
}
```

### 使用指针类型

```go
type User struct {
    ID    int
    Name  string
    Email *string  // nil 表示 NULL
    Phone *string
}

func getUser(db *sql.DB, id int) (*User, error) {
    var u User
    err := db.QueryRow(
        "SELECT id, name, email, phone FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.Phone)

    if err != nil {
        return nil, err
    }

    if u.Email != nil {
        fmt.Println("Email:", *u.Email)
    }

    return &u, nil
}
```

## 预处理语句

### 基本使用

```go
func batchInsertUsers(db *sql.DB, users []User) error {
    // 预编译 SQL
    stmt, err := db.Prepare("INSERT INTO users (name, email) VALUES ($1, $2)")
    if err != nil {
        return err
    }
    defer stmt.Close()  // 必须关闭！

    for _, u := range users {
        _, err := stmt.Exec(u.Name, u.Email)
        if err != nil {
            return err
        }
    }

    return nil
}
```

### 带 Context

```go
func batchInsertWithContext(ctx context.Context, db *sql.DB, users []User) error {
    stmt, err := db.PrepareContext(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, u := range users {
        if _, err := stmt.ExecContext(ctx, u.Name, u.Email); err != nil {
            return err
        }
    }

    return nil
}
```

## 事务

### 基本事务

```go
func transferMoney(db *sql.DB, fromID, toID int, amount float64) error {
    // 开始事务
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    // 确保事务最终被处理
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    // 扣款
    result, err := tx.Exec(
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
        amount, fromID,
    )
    if err != nil {
        tx.Rollback()
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        tx.Rollback()
        return errors.New("insufficient balance or account not found")
    }

    // 入账
    _, err = tx.Exec(
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, toID,
    )
    if err != nil {
        tx.Rollback()
        return err
    }

    // 提交事务
    return tx.Commit()
}
```

### 事务辅助函数

```go
// WithTx 事务辅助函数
func WithTx(db *sql.DB, fn func(*sql.Tx) error) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

// 使用
err := WithTx(db, func(tx *sql.Tx) error {
    if _, err := tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", 100, 1); err != nil {
        return err
    }
    if _, err := tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", 100, 2); err != nil {
        return err
    }
    return nil
})
```

### 带 Context 和选项的事务

```go
func transferWithContext(ctx context.Context, db *sql.DB, fromID, toID int, amount float64) error {
    // 带选项的事务
    tx, err := db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,  // 隔离级别
        ReadOnly:  false,
    })
    if err != nil {
        return err
    }

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    // ... 事务操作

    return tx.Commit()
}
```

## 完整示例：用户仓储

```go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "time"
)

var (
    ErrUserNotFound = errors.New("user not found")
)

type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (name, email, created_at, updated_at)
        VALUES ($1, $2, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).
        Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`

    var u User
    err := r.db.QueryRowContext(ctx, query, id).
        Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, err
    }

    return &u, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    query := `SELECT id, name, email, created_at, updated_at FROM users WHERE email = $1`

    var u User
    err := r.db.QueryRowContext(ctx, query, email).
        Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, err
    }

    return &u, nil
}

// List 列出用户（带分页）
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*User, error) {
    query := `
        SELECT id, name, email, created_at, updated_at
        FROM users
        ORDER BY id
        LIMIT $1 OFFSET $2
    `

    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, err
        }
        users = append(users, &u)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *User) error {
    query := `
        UPDATE users
        SET name = $1, email = $2, updated_at = NOW()
        WHERE id = $3
        RETURNING updated_at
    `

    err := r.db.QueryRowContext(ctx, query, user.Name, user.Email, user.ID).
        Scan(&user.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return ErrUserNotFound
        }
        return err
    }

    return nil
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id int) error {
    query := `DELETE FROM users WHERE id = $1`

    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return ErrUserNotFound
    }

    return nil
}

// Count 统计用户数量
func (r *UserRepository) Count(ctx context.Context) (int, error) {
    var count int
    err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
    return count, err
}
```

## 错误处理

```go
import (
    "database/sql"
    "errors"

    "github.com/lib/pq"
)

func createUser(db *sql.DB, name, email string) error {
    _, err := db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
    if err != nil {
        // PostgreSQL 特定错误
        var pqErr *pq.Error
        if errors.As(err, &pqErr) {
            switch pqErr.Code {
            case "23505":  // unique_violation
                return ErrEmailExists
            case "23503":  // foreign_key_violation
                return ErrForeignKeyViolation
            case "23502":  // not_null_violation
                return ErrNullViolation
            }
        }
        return err
    }
    return nil
}
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 始终关闭 Rows | 使用 `defer rows.Close()` |
| 检查 rows.Err() | 迭代后检查错误 |
| 使用 Context | 支持超时和取消 |
| 配置连接池 | 根据负载调整参数 |
| 预处理批量操作 | 复用 Prepare 语句 |
| 参数化查询 | 防止 SQL 注入 |
| 处理 sql.ErrNoRows | 区分"未找到"和"错误" |

**下一节**：[pgx 驱动](./04-pgx-driver.md) - 学习高性能 PostgreSQL 驱动
