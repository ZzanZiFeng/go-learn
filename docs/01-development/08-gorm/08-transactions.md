# 事务 (Transactions)

## 概述

GORM 提供灵活的事务支持，确保数据操作的原子性和一致性。

## 自动事务

GORM 默认在事务中执行单个创建、更新、删除操作。

```go
// 这些操作自动在事务中执行
db.Create(&user)
db.Updates(&user)
db.Delete(&user)

// 禁用自动事务（提升性能）
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    SkipDefaultTransaction: true,
})
```

## 手动事务

### 基本用法

```go
// 开始事务
tx := db.Begin()

// 检查事务是否开始成功
if tx.Error != nil {
    return tx.Error
}

// 在事务中执行操作
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&account).Error; err != nil {
    tx.Rollback()
    return err
}

// 提交事务
return tx.Commit().Error
```

### 使用 defer

```go
func CreateUserWithAccount(db *gorm.DB, user *User, account *Account) error {
    tx := db.Begin()
    if tx.Error != nil {
        return tx.Error
    }

    // 使用 defer 确保事务被处理
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)  // 重新 panic
        }
    }()

    if err := tx.Create(user).Error; err != nil {
        tx.Rollback()
        return err
    }

    account.UserID = user.ID
    if err := tx.Create(account).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```

## Transaction 方法

### 推荐用法

```go
// 使用 Transaction 方法自动管理事务
err := db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    if err := tx.Create(&user).Error; err != nil {
        return err  // 返回 error 会自动回滚
    }

    if err := tx.Create(&account).Error; err != nil {
        return err
    }

    // 返回 nil 会自动提交
    return nil
})

if err != nil {
    // 处理错误
}
```

### 带返回值

```go
func CreateOrder(db *gorm.DB, order *Order) (*Order, error) {
    err := db.Transaction(func(tx *gorm.DB) error {
        // 创建订单
        if err := tx.Create(order).Error; err != nil {
            return err
        }

        // 扣减库存
        result := tx.Model(&Product{}).
            Where("id = ? AND stock >= ?", order.ProductID, order.Quantity).
            Update("stock", gorm.Expr("stock - ?", order.Quantity))

        if result.RowsAffected == 0 {
            return errors.New("insufficient stock")
        }

        // 创建支付记录
        payment := Payment{OrderID: order.ID, Amount: order.Amount}
        if err := tx.Create(&payment).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    return order, nil
}
```

## 嵌套事务

### 使用 SavePoint

```go
err := db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&user1)

    // 嵌套事务
    err := tx.Transaction(func(tx2 *gorm.DB) error {
        tx2.Create(&user2)
        return errors.New("rollback user2")  // 只回滚 user2
    })

    // user1 仍然会被创建
    return nil
})
```

### 手动 SavePoint

```go
tx := db.Begin()

tx.Create(&user1)

// 创建保存点
tx.SavePoint("sp1")

tx.Create(&user2)

// 回滚到保存点（只回滚 user2）
tx.RollbackTo("sp1")

tx.Create(&user3)

tx.Commit()
// 结果: user1 和 user3 被创建，user2 被回滚
```

## 事务选项

### 隔离级别

```go
// 使用特定隔离级别
tx := db.Begin(&sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
})

// 或使用 Session
db.Session(&gorm.Session{
    PrepareStmt: true,
}).Begin(&sql.TxOptions{
    Isolation: sql.LevelRepeatableRead,
})
```

### 隔离级别说明

```go
import "database/sql"

// 隔离级别（从低到高）
sql.LevelDefault          // 数据库默认
sql.LevelReadUncommitted  // 读未提交（脏读）
sql.LevelReadCommitted    // 读已提交（不可重复读）
sql.LevelRepeatableRead   // 可重复读（幻读）
sql.LevelSerializable     // 串行化（最高隔离）

// 选择指南:
// - ReadCommitted: 大多数场景足够
// - RepeatableRead: 需要一致性快照
// - Serializable: 金融等关键场景
```

### 只读事务

```go
tx := db.Begin(&sql.TxOptions{
    ReadOnly: true,  // 只读事务（可能有性能优化）
})

var users []User
tx.Find(&users)

tx.Commit()
```

## 实际应用示例

### 转账功能

```go
type TransferService struct {
    db *gorm.DB
}

func (s *TransferService) Transfer(fromID, toID uint, amount float64) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 锁定账户（悲观锁）
        var fromAccount, toAccount Account

        // FOR UPDATE 锁定
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&fromAccount, fromID).Error; err != nil {
            return fmt.Errorf("from account not found: %w", err)
        }

        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&toAccount, toID).Error; err != nil {
            return fmt.Errorf("to account not found: %w", err)
        }

        // 检查余额
        if fromAccount.Balance < amount {
            return errors.New("insufficient balance")
        }

        // 执行转账
        if err := tx.Model(&fromAccount).Update("balance", fromAccount.Balance - amount).Error; err != nil {
            return err
        }

        if err := tx.Model(&toAccount).Update("balance", toAccount.Balance + amount).Error; err != nil {
            return err
        }

        // 记录转账日志
        log := TransferLog{
            FromAccountID: fromID,
            ToAccountID:   toID,
            Amount:        amount,
            Status:        "completed",
        }
        if err := tx.Create(&log).Error; err != nil {
            return err
        }

        return nil
    })
}
```

### 订单创建

```go
type OrderService struct {
    db *gorm.DB
}

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var order Order

    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. 检查商品
        var product Product
        if err := tx.First(&product, req.ProductID).Error; err != nil {
            return fmt.Errorf("product not found: %w", err)
        }

        // 2. 检查并扣减库存（乐观锁）
        result := tx.Model(&Product{}).
            Where("id = ? AND stock >= ? AND version = ?",
                req.ProductID, req.Quantity, product.Version).
            Updates(map[string]interface{}{
                "stock":   gorm.Expr("stock - ?", req.Quantity),
                "version": gorm.Expr("version + 1"),
            })

        if result.RowsAffected == 0 {
            return errors.New("stock insufficient or concurrent update")
        }

        // 3. 创建订单
        order = Order{
            UserID:    req.UserID,
            ProductID: req.ProductID,
            Quantity:  req.Quantity,
            Amount:    product.Price * float64(req.Quantity),
            Status:    "pending",
        }
        if err := tx.Create(&order).Error; err != nil {
            return err
        }

        // 4. 创建订单项
        orderItem := OrderItem{
            OrderID:   order.ID,
            ProductID: req.ProductID,
            Quantity:  req.Quantity,
            Price:     product.Price,
        }
        if err := tx.Create(&orderItem).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    return &order, nil
}
```

### 批量操作

```go
func (s *ProductService) BatchUpdate(updates []ProductUpdate) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        for _, update := range updates {
            if err := tx.Model(&Product{}).
                Where("id = ?", update.ID).
                Updates(update.Data).Error; err != nil {
                return fmt.Errorf("failed to update product %d: %w", update.ID, err)
            }
        }
        return nil
    })
}
```

## 事务辅助函数

### 通用事务封装

```go
// WithTransaction 通用事务封装
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
    return db.Transaction(fn)
}

// WithRetry 带重试的事务
func WithRetry(db *gorm.DB, maxRetries int, fn func(tx *gorm.DB) error) error {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        err := db.Transaction(fn)
        if err == nil {
            return nil
        }

        // 检查是否可重试（如死锁、序列化失败）
        if isRetryable(err) {
            lastErr = err
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
            continue
        }

        return err
    }

    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryable(err error) bool {
    // PostgreSQL: deadlock_detected (40P01), serialization_failure (40001)
    errStr := err.Error()
    return strings.Contains(errStr, "40P01") ||
           strings.Contains(errStr, "40001") ||
           strings.Contains(errStr, "deadlock")
}
```

### 上下文感知事务

```go
func WithTxContext(ctx context.Context, db *gorm.DB, fn func(ctx context.Context, tx *gorm.DB) error) error {
    return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        return fn(ctx, tx)
    })
}

// 使用
err := WithTxContext(ctx, db, func(ctx context.Context, tx *gorm.DB) error {
    // 事务操作
    return nil
})
```

## 分布式事务

### Saga 模式

```go
type SagaStep struct {
    Execute    func(tx *gorm.DB) error
    Compensate func(tx *gorm.DB) error
}

func ExecuteSaga(db *gorm.DB, steps []SagaStep) error {
    var executedSteps []SagaStep

    for _, step := range steps {
        if err := step.Execute(db); err != nil {
            // 执行补偿操作
            for i := len(executedSteps) - 1; i >= 0; i-- {
                if compensateErr := executedSteps[i].Compensate(db); compensateErr != nil {
                    // 记录补偿失败，需要人工处理
                    log.Printf("compensation failed: %v", compensateErr)
                }
            }
            return err
        }
        executedSteps = append(executedSteps, step)
    }

    return nil
}

// 使用示例
steps := []SagaStep{
    {
        Execute:    func(tx *gorm.DB) error { return tx.Create(&order).Error },
        Compensate: func(tx *gorm.DB) error { return tx.Delete(&order).Error },
    },
    {
        Execute:    func(tx *gorm.DB) error { return deductInventory(tx, order) },
        Compensate: func(tx *gorm.DB) error { return restoreInventory(tx, order) },
    },
    {
        Execute:    func(tx *gorm.DB) error { return createPayment(tx, order) },
        Compensate: func(tx *gorm.DB) error { return cancelPayment(tx, order) },
    },
}

err := ExecuteSaga(db, steps)
```

## 最佳实践

### 1. 保持事务简短

```go
// ❌ 事务中包含耗时操作
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&order)
    callExternalAPI()  // 可能很慢
    tx.Create(&payment)
    return nil
})

// ✅ 先准备数据，再执行事务
data := callExternalAPI()
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&order)
    payment.Data = data
    tx.Create(&payment)
    return nil
})
```

### 2. 正确处理错误

```go
// ✅ 总是检查并返回错误
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err  // 会自动回滚
    }
    return nil
})
```

### 3. 避免嵌套事务滥用

```go
// 只在真正需要部分回滚时使用嵌套事务
```

**下一节**：[原生 SQL](./09-raw-sql.md) - 学习 Raw SQL 和 SQL Builder
