# 第七章：数据库基础 (Database Fundamentals)

## 章节概述

本章介绍 Go 语言中数据库操作的基础知识，重点是 PostgreSQL 和标准库 `database/sql`。

## 学习目标

完成本章后，你将能够：

- 使用 `database/sql` 连接和操作数据库
- 理解连接池的配置和管理
- 编写安全的 SQL 查询（防止 SQL 注入）
- 正确处理数据库事务
- 使用迁移工具管理数据库 schema

## 与 JavaScript 对比

### Node.js 数据库操作

```javascript
// node-postgres
const { Pool } = require('pg');
const pool = new Pool({
  connectionString: process.env.DATABASE_URL
});

// 查询
const result = await pool.query('SELECT * FROM users WHERE id = $1', [1]);

// 事务
const client = await pool.connect();
try {
  await client.query('BEGIN');
  await client.query('INSERT INTO users (name) VALUES ($1)', ['Alice']);
  await client.query('COMMIT');
} catch (e) {
  await client.query('ROLLBACK');
  throw e;
} finally {
  client.release();
}
```

### Go database/sql

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

// 连接
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

// 查询
row := db.QueryRow("SELECT * FROM users WHERE id = $1", 1)

// 事务
tx, err := db.Begin()
_, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "Alice")
if err != nil {
    tx.Rollback()
    return err
}
tx.Commit()
```

## 章节内容

| 文档 | 主题 | 描述 |
|------|------|------|
| [01-sql-review.md](./01-sql-review.md) | SQL 复习 | SQL 基础语法回顾 |
| [02-postgres-setup.md](./02-postgres-setup.md) | PostgreSQL 安装 | 本地和 Docker 安装 |
| [03-database-sql.md](./03-database-sql.md) | database/sql | 标准库使用 |
| [04-pgx-driver.md](./04-pgx-driver.md) | pgx 驱动 | 高性能 PostgreSQL 驱动 |
| [05-connection-pool.md](./05-connection-pool.md) | 连接池 | 连接池配置和管理 |
| [06-prepared-stmt.md](./06-prepared-stmt.md) | 预处理语句 | SQL 注入防护 |
| [07-transactions.md](./07-transactions.md) | 事务处理 | ACID 和事务管理 |
| [08-migrations.md](./08-migrations.md) | 数据库迁移 | Schema 版本控制 |
| [exercises.md](./exercises.md) | 练习 | 实践练习 |

## 快速开始

### 1. 启动 PostgreSQL

```bash
# 使用 Docker
docker run --name postgres \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -d postgres:15

# 或使用项目的 docker-compose
cd infra && docker-compose up -d postgres
```

### 2. 基本示例

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

func main() {
    // 连接数据库
    connStr := "postgres://user:password@localhost:5432/testdb?sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 验证连接
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to database!")

    // 创建表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    // 插入数据
    var userID int
    err = db.QueryRow(
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        "Alice", "alice@example.com",
    ).Scan(&userID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Inserted user with ID: %d\n", userID)

    // 查询数据
    var name, email string
    err = db.QueryRow(
        "SELECT name, email FROM users WHERE id = $1", userID,
    ).Scan(&name, &email)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User: %s <%s>\n", name, email)
}
```

## 依赖安装

```bash
# PostgreSQL 驱动
go get github.com/lib/pq        # 标准驱动
go get github.com/jackc/pgx/v5  # 高性能驱动

# 迁移工具
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## 核心概念

### database/sql 架构

```
┌──────────────────────────────────────────────────┐
│                   应用程序                         │
└───────────────────────┬──────────────────────────┘
                        │
┌───────────────────────▼──────────────────────────┐
│                 database/sql                      │
│  ┌──────────────────────────────────────────┐   │
│  │              连接池管理                     │   │
│  │  MaxOpenConns, MaxIdleConns, MaxLifetime │   │
│  └──────────────────────────────────────────┘   │
└───────────────────────┬──────────────────────────┘
                        │
┌───────────────────────▼──────────────────────────┐
│               数据库驱动 (Driver)                  │
│         lib/pq, pgx, go-sql-driver/mysql         │
└───────────────────────┬──────────────────────────┘
                        │
┌───────────────────────▼──────────────────────────┐
│              数据库服务器                          │
│         PostgreSQL, MySQL, SQLite, etc.          │
└──────────────────────────────────────────────────┘
```

### 查询类型

| 方法 | 用途 | 返回值 |
|------|------|--------|
| `db.Query()` | 返回多行 | `*sql.Rows` |
| `db.QueryRow()` | 返回单行 | `*sql.Row` |
| `db.Exec()` | 执行无返回 | `sql.Result` |
| `db.Prepare()` | 预处理语句 | `*sql.Stmt` |

## 最佳实践

| 实践 | 说明 |
|------|------|
| 使用连接池 | `database/sql` 自动管理连接池 |
| 参数化查询 | 永远使用 `$1, $2` 占位符，防止 SQL 注入 |
| 检查错误 | 每个数据库操作都要检查错误 |
| 关闭资源 | 使用 `defer` 关闭 `Rows`、`Stmt` |
| 使用事务 | 多个相关操作放在事务中 |
| 配置超时 | 设置查询和连接超时 |
| 迁移管理 | 使用迁移工具管理 schema 变更 |

## 常见陷阱

### 1. 忘记关闭 Rows

```go
// ❌ 错误 - 可能导致连接泄露
rows, _ := db.Query("SELECT * FROM users")
for rows.Next() {
    // ...
}

// ✅ 正确
rows, err := db.Query("SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close()
for rows.Next() {
    // ...
}
```

### 2. SQL 注入

```go
// ❌ 危险 - SQL 注入
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
db.Query(query)

// ✅ 安全 - 参数化查询
db.Query("SELECT * FROM users WHERE name = $1", name)
```

### 3. 在循环中使用 Prepare

```go
// ❌ 性能差 - 每次循环都准备语句
for _, user := range users {
    db.Exec("INSERT INTO users (name) VALUES ($1)", user.Name)
}

// ✅ 优化 - 预先准备语句
stmt, _ := db.Prepare("INSERT INTO users (name) VALUES ($1)")
defer stmt.Close()
for _, user := range users {
    stmt.Exec(user.Name)
}
```

**下一节**：[SQL 复习](./01-sql-review.md) - 开始学习 SQL 基础语法
