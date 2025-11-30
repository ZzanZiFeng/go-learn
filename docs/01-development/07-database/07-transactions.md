# 事务处理 (Transaction)

## 概述

事务是一组数据库操作的逻辑单元，要么全部成功，要么全部回滚。

## ACID 特性

```
A - Atomicity (原子性)
    事务中的所有操作要么全部完成，要么全部不完成

C - Consistency (一致性)
    事务前后数据库状态保持一致

I - Isolation (隔离性)
    并发事务之间相互隔离

D - Durability (持久性)
    事务完成后，修改永久保存
```

## 基本事务

### database/sql

```go
func transferMoney(db *sql.DB, fromID, toID int, amount float64) error {
    // 开始事务
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    // 使用 defer 确保事务最终被处理
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

### pgx

```go
func transferMoneyPgx(ctx context.Context, pool *pgxpool.Pool, from, to int, amount float64) error {
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

## 事务辅助函数

### 通用模式

```go
// WithTx 执行事务内的函数
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

### 带 Context 的版本

```go
func WithTxContext(ctx context.Context, db *sql.DB, fn func(context.Context, *sql.Tx) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    if err := fn(ctx, tx); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

// 使用
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := WithTxContext(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
    // 事务内的操作
    return nil
})
```

### pgx BeginFunc（推荐）

```go
import "github.com/jackc/pgx/v5"

func transferWithBeginFunc(ctx context.Context, pool *pgxpool.Pool, from, to int, amount float64) error {
    return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
        // 事务成功自动提交，错误或 panic 自动回滚
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

## 隔离级别

### PostgreSQL 隔离级别

| 级别 | 脏读 | 不可重复读 | 幻读 |
|------|------|-----------|------|
| Read Uncommitted | 不会* | 可能 | 可能 |
| Read Committed (默认) | 不会 | 可能 | 可能 |
| Repeatable Read | 不会 | 不会 | 不会* |
| Serializable | 不会 | 不会 | 不会 |

*PostgreSQL 的 Read Uncommitted 实际等同于 Read Committed
*PostgreSQL 的 Repeatable Read 也防止幻读

### 设置隔离级别

```go
// database/sql
func serializedTransaction(db *sql.DB) error {
    tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
        Isolation: sql.LevelSerializable,
        ReadOnly:  false,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // ... 事务操作

    return tx.Commit()
}

// pgx
func serializedTransactionPgx(ctx context.Context, pool *pgxpool.Pool) error {
    tx, err := pool.BeginTx(ctx, pgx.TxOptions{
        IsoLevel:       pgx.Serializable,
        AccessMode:     pgx.ReadWrite,
        DeferrableMode: pgx.NotDeferrable,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // ... 事务操作

    return tx.Commit(ctx)
}
```

### 只读事务

```go
func readOnlyTransaction(db *sql.DB) error {
    tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
        ReadOnly: true,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 只能执行 SELECT
    rows, err := tx.Query("SELECT * FROM users")
    if err != nil {
        return err
    }
    defer rows.Close()

    // ...

    return tx.Commit()
}
```

## 并发问题与解决

### 1. 丢失更新

```go
// 问题场景：两个事务同时读取并更新同一行
/*
事务 A: SELECT balance FROM accounts WHERE id = 1  -- 返回 100
事务 B: SELECT balance FROM accounts WHERE id = 1  -- 返回 100
事务 A: UPDATE accounts SET balance = 100 + 50 = 150
事务 B: UPDATE accounts SET balance = 100 + 30 = 130  -- 覆盖了事务 A 的更新
*/

// 解决方案 1: 使用原子操作
tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", 50, 1)

// 解决方案 2: 悲观锁 (SELECT FOR UPDATE)
func updateWithLock(tx *sql.Tx, id int, amount float64) error {
    var balance float64
    err := tx.QueryRow(
        "SELECT balance FROM accounts WHERE id = $1 FOR UPDATE",
        id,
    ).Scan(&balance)
    if err != nil {
        return err
    }

    _, err = tx.Exec(
        "UPDATE accounts SET balance = $1 WHERE id = $2",
        balance + amount, id,
    )
    return err
}
```

### 2. 乐观锁

```go
// 使用版本号实现乐观锁
func updateWithOptimisticLock(db *sql.DB, user *User) error {
    result, err := db.Exec(
        "UPDATE users SET name = $1, version = version + 1 WHERE id = $2 AND version = $3",
        user.Name, user.ID, user.Version,
    )
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return ErrConcurrentUpdate  // 数据已被其他事务修改
    }

    user.Version++
    return nil
}

// 重试机制
func updateWithRetry(db *sql.DB, id int, updateFn func(*User) error) error {
    maxRetries := 3

    for i := 0; i < maxRetries; i++ {
        // 读取最新数据
        user, err := getUserByID(db, id)
        if err != nil {
            return err
        }

        // 应用更新
        if err := updateFn(user); err != nil {
            return err
        }

        // 尝试保存
        err = updateWithOptimisticLock(db, user)
        if err == nil {
            return nil  // 成功
        }

        if errors.Is(err, ErrConcurrentUpdate) {
            continue  // 重试
        }

        return err  // 其他错误
    }

    return errors.New("max retries exceeded")
}
```

### 3. 死锁

```go
// 问题场景：两个事务互相等待
/*
事务 A: UPDATE accounts SET ... WHERE id = 1  -- 锁定 id=1
事务 B: UPDATE accounts SET ... WHERE id = 2  -- 锁定 id=2
事务 A: UPDATE accounts SET ... WHERE id = 2  -- 等待 id=2
事务 B: UPDATE accounts SET ... WHERE id = 1  -- 等待 id=1 → 死锁
*/

// 解决方案 1: 按固定顺序获取锁
func transferSafe(tx *sql.Tx, from, to int, amount float64) error {
    // 确保始终按 ID 顺序锁定
    firstID, secondID := from, to
    if from > to {
        firstID, secondID = to, from
    }

    // 按顺序锁定
    _, err := tx.Exec("SELECT 1 FROM accounts WHERE id = $1 FOR UPDATE", firstID)
    if err != nil {
        return err
    }
    _, err = tx.Exec("SELECT 1 FROM accounts WHERE id = $1 FOR UPDATE", secondID)
    if err != nil {
        return err
    }

    // 执行转账
    // ...
    return nil
}

// 解决方案 2: 设置锁超时
func setLockTimeout(tx *sql.Tx) error {
    _, err := tx.Exec("SET lock_timeout = '5s'")
    return err
}

// 解决方案 3: 使用 NOWAIT
func tryLock(tx *sql.Tx, id int) error {
    _, err := tx.Exec(
        "SELECT 1 FROM accounts WHERE id = $1 FOR UPDATE NOWAIT",
        id,
    )
    if err != nil {
        // 检查是否是锁等待错误
        return ErrLockNotAvailable
    }
    return nil
}
```

## 保存点 (Savepoint)

```go
func transactionWithSavepoint(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 第一个操作
    _, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "Alice")
    if err != nil {
        return err
    }

    // 创建保存点
    _, err = tx.Exec("SAVEPOINT sp1")
    if err != nil {
        return err
    }

    // 第二个操作（可能失败）
    _, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "Bob")
    if err != nil {
        // 回滚到保存点（不是整个事务）
        tx.Exec("ROLLBACK TO SAVEPOINT sp1")
        // 继续其他操作
    }

    // 第三个操作
    _, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "Charlie")
    if err != nil {
        return err
    }

    return tx.Commit()  // Alice 和 Charlie 被提交
}
```

## 嵌套事务模式

```go
// 使用保存点模拟嵌套事务
type TxManager struct {
    db    *sql.DB
    tx    *sql.Tx
    level int
}

func NewTxManager(db *sql.DB) *TxManager {
    return &TxManager{db: db}
}

func (m *TxManager) Begin() error {
    if m.tx == nil {
        tx, err := m.db.Begin()
        if err != nil {
            return err
        }
        m.tx = tx
        m.level = 1
    } else {
        // 嵌套事务使用保存点
        _, err := m.tx.Exec(fmt.Sprintf("SAVEPOINT sp%d", m.level))
        if err != nil {
            return err
        }
        m.level++
    }
    return nil
}

func (m *TxManager) Commit() error {
    m.level--
    if m.level == 0 {
        err := m.tx.Commit()
        m.tx = nil
        return err
    }
    // 嵌套事务释放保存点
    _, err := m.tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT sp%d", m.level))
    return err
}

func (m *TxManager) Rollback() error {
    m.level--
    if m.level == 0 {
        err := m.tx.Rollback()
        m.tx = nil
        return err
    }
    // 嵌套事务回滚到保存点
    _, err := m.tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT sp%d", m.level))
    return err
}
```

## 仓储模式中的事务

```go
// 事务接口
type Transaction interface {
    Commit() error
    Rollback() error
}

// 仓储接口
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int) (*User, error)
    Update(ctx context.Context, user *User) error
}

// 带事务的仓储
type TxUserRepository struct {
    tx *sql.Tx
}

func (r *TxUserRepository) Create(ctx context.Context, user *User) error {
    return r.tx.QueryRowContext(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email,
    ).Scan(&user.ID)
}

// 工作单元
type UnitOfWork struct {
    db    *sql.DB
    tx    *sql.Tx
    Users *TxUserRepository
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
    return &UnitOfWork{db: db}
}

func (uow *UnitOfWork) Begin() error {
    tx, err := uow.db.Begin()
    if err != nil {
        return err
    }
    uow.tx = tx
    uow.Users = &TxUserRepository{tx: tx}
    return nil
}

func (uow *UnitOfWork) Commit() error {
    return uow.tx.Commit()
}

func (uow *UnitOfWork) Rollback() error {
    return uow.tx.Rollback()
}

// 使用
func createUserWithProfile(db *sql.DB, user *User, profile *Profile) error {
    uow := NewUnitOfWork(db)
    if err := uow.Begin(); err != nil {
        return err
    }
    defer uow.Rollback()

    if err := uow.Users.Create(context.Background(), user); err != nil {
        return err
    }

    // ... 创建 profile

    return uow.Commit()
}
```

## 最佳实践

| 实践 | 说明 |
|------|------|
| 保持事务短小 | 减少锁持有时间 |
| 使用 defer Rollback | 确保事务被处理 |
| 选择合适的隔离级别 | 平衡一致性和性能 |
| 处理并发冲突 | 使用乐观锁或悲观锁 |
| 避免长事务 | 超过几秒的事务需要重新设计 |
| 按固定顺序获取锁 | 防止死锁 |
| 设置超时 | 使用 context 或 lock_timeout |

**下一节**：[数据库迁移](./08-migrations.md) - 学习 Schema 版本控制
