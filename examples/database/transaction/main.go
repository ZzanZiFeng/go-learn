// examples/database/transaction/main.go
// 演示数据库事务处理
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrAccountNotFound     = errors.New("account not found")
	ErrConcurrentUpdate    = errors.New("concurrent update detected")
)

// Account 账户模型
type Account struct {
	ID      int
	UserID  int
	Balance float64
	Version int // 乐观锁版本号
}

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/testdb?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// 初始化表
	if err := initTables(ctx, db); err != nil {
		log.Fatal(err)
	}

	// 演示基本事务
	fmt.Println("=== 基本事务 ===")
	if err := basicTransactionDemo(ctx, db); err != nil {
		log.Printf("Basic transaction error: %v\n", err)
	}

	// 演示转账事务
	fmt.Println("\n=== 转账事务 ===")
	if err := transferDemo(ctx, db); err != nil {
		log.Printf("Transfer error: %v\n", err)
	}

	// 演示事务辅助函数
	fmt.Println("\n=== 事务辅助函数 ===")
	if err := withTxDemo(ctx, db); err != nil {
		log.Printf("WithTx error: %v\n", err)
	}

	// 演示乐观锁
	fmt.Println("\n=== 乐观锁 ===")
	if err := optimisticLockDemo(ctx, db); err != nil {
		log.Printf("Optimistic lock error: %v\n", err)
	}
}

func initTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS transfer_log;
		DROP TABLE IF EXISTS accounts;

		CREATE TABLE accounts (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			balance DECIMAL(15,2) NOT NULL DEFAULT 0,
			version INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE transfer_log (
			id SERIAL PRIMARY KEY,
			from_account_id INTEGER NOT NULL,
			to_account_id INTEGER NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);

		-- 初始化测试数据
		INSERT INTO accounts (user_id, balance) VALUES (1, 1000.00);
		INSERT INTO accounts (user_id, balance) VALUES (2, 500.00);
		INSERT INTO accounts (user_id, balance) VALUES (3, 0.00);
	`)
	return err
}

// basicTransactionDemo 演示基本事务操作
func basicTransactionDemo(ctx context.Context, db *sql.DB) error {
	// 开始事务
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 使用 defer 确保事务被处理
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新 panic
		}
	}()

	// 执行多个操作
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		100.00, 1,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		100.00, 2,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Println("Basic transaction committed successfully")
	printBalances(ctx, db)
	return nil
}

// transferDemo 演示转账事务
func transferDemo(ctx context.Context, db *sql.DB) error {
	// 成功的转账
	err := transfer(ctx, db, 1, 2, 200.00)
	if err != nil {
		return fmt.Errorf("transfer failed: %w", err)
	}
	fmt.Println("Transfer 1 -> 2: $200.00 successful")
	printBalances(ctx, db)

	// 失败的转账（余额不足）
	err = transfer(ctx, db, 2, 1, 10000.00)
	if errors.Is(err, ErrInsufficientBalance) {
		fmt.Println("Transfer 2 -> 1: $10000.00 failed (insufficient balance)")
	} else if err != nil {
		return err
	}

	printBalances(ctx, db)
	return nil
}

// transfer 执行转账事务
func transfer(ctx context.Context, db *sql.DB, fromID, toID int, amount float64) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // 最高隔离级别
	})
	if err != nil {
		return err
	}
	defer tx.Rollback() // 如果已提交，Rollback 是空操作

	// 检查并扣款
	result, err := tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
		amount, fromID,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrInsufficientBalance
	}

	// 入账
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		amount, toID,
	)
	if err != nil {
		return err
	}

	// 记录转账日志
	_, err = tx.ExecContext(ctx,
		"INSERT INTO transfer_log (from_account_id, to_account_id, amount) VALUES ($1, $2, $3)",
		fromID, toID, amount,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// WithTx 事务辅助函数
func WithTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
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

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// withTxDemo 演示使用事务辅助函数
func withTxDemo(ctx context.Context, db *sql.DB) error {
	err := WithTx(ctx, db, func(tx *sql.Tx) error {
		// 事务内的操作
		_, err := tx.ExecContext(ctx,
			"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
			50.00, 3,
		)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
			50.00, 1,
		)
		return err
	})

	if err != nil {
		return err
	}

	fmt.Println("WithTx transaction committed successfully")
	printBalances(ctx, db)
	return nil
}

// optimisticLockDemo 演示乐观锁
func optimisticLockDemo(ctx context.Context, db *sql.DB) error {
	// 获取账户
	var account Account
	err := db.QueryRowContext(ctx,
		"SELECT id, user_id, balance, version FROM accounts WHERE id = $1",
		1,
	).Scan(&account.ID, &account.UserID, &account.Balance, &account.Version)
	if err != nil {
		return err
	}

	fmt.Printf("Current account: ID=%d, Balance=%.2f, Version=%d\n",
		account.ID, account.Balance, account.Version)

	// 模拟修改
	newBalance := account.Balance + 100.00

	// 使用乐观锁更新
	result, err := db.ExecContext(ctx,
		"UPDATE accounts SET balance = $1, version = version + 1 WHERE id = $2 AND version = $3",
		newBalance, account.ID, account.Version,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrConcurrentUpdate
	}

	fmt.Printf("Updated with optimistic lock: new balance=%.2f\n", newBalance)

	// 尝试用旧版本号再次更新（会失败）
	result, err = db.ExecContext(ctx,
		"UPDATE accounts SET balance = $1, version = version + 1 WHERE id = $2 AND version = $3",
		newBalance + 50, account.ID, account.Version, // 使用旧的 version
	)
	if err != nil {
		return err
	}

	rowsAffected, _ = result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Println("Second update failed (expected - version mismatch)")
	}

	return nil
}

// printBalances 打印所有账户余额
func printBalances(ctx context.Context, db *sql.DB) {
	rows, err := db.QueryContext(ctx, "SELECT id, balance FROM accounts ORDER BY id")
	if err != nil {
		log.Printf("Error querying balances: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Account balances:")
	for rows.Next() {
		var id int
		var balance float64
		if err := rows.Scan(&id, &balance); err != nil {
			log.Printf("Error scanning: %v\n", err)
			continue
		}
		fmt.Printf("  Account %d: $%.2f\n", id, balance)
	}
}
