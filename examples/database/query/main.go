// examples/database/query/main.go
// 演示数据库查询操作
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// User 用户模型
type User struct {
	ID        int
	Name      string
	Email     string
	Phone     sql.NullString // 可能为 NULL
	CreatedAt time.Time
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
	if err := initTable(ctx, db); err != nil {
		log.Fatal(err)
	}

	// 演示各种查询操作
	fmt.Println("=== 插入数据 ===")
	if err := insertDemo(ctx, db); err != nil {
		log.Printf("Insert error: %v\n", err)
	}

	fmt.Println("\n=== 查询单行 ===")
	if err := queryRowDemo(ctx, db); err != nil {
		log.Printf("QueryRow error: %v\n", err)
	}

	fmt.Println("\n=== 查询多行 ===")
	if err := queryDemo(ctx, db); err != nil {
		log.Printf("Query error: %v\n", err)
	}

	fmt.Println("\n=== 更新数据 ===")
	if err := updateDemo(ctx, db); err != nil {
		log.Printf("Update error: %v\n", err)
	}

	fmt.Println("\n=== 删除数据 ===")
	if err := deleteDemo(ctx, db); err != nil {
		log.Printf("Delete error: %v\n", err)
	}
}

func initTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		DROP TABLE IF EXISTS users;
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			phone VARCHAR(20),
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	return err
}

// insertDemo 演示插入操作
func insertDemo(ctx context.Context, db *sql.DB) error {
	// 方式 1: Exec（不需要返回值）
	result, err := db.ExecContext(ctx,
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"Alice", "alice@example.com",
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d row(s)\n", rowsAffected)

	// 方式 2: QueryRow + RETURNING（获取生成的 ID）
	var id int
	err = db.QueryRowContext(ctx,
		"INSERT INTO users (name, email, phone) VALUES ($1, $2, $3) RETURNING id",
		"Bob", "bob@example.com", "123-456-7890",
	).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("Inserted user with ID: %d\n", id)

	// 方式 3: 预处理语句（批量插入）
	stmt, err := db.PrepareContext(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	users := []struct{ name, email string }{
		{"Charlie", "charlie@example.com"},
		{"David", "david@example.com"},
		{"Eve", "eve@example.com"},
	}

	for _, u := range users {
		_, err := stmt.ExecContext(ctx, u.name, u.email)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Batch inserted %d users\n", len(users))

	return nil
}

// queryRowDemo 演示单行查询
func queryRowDemo(ctx context.Context, db *sql.DB) error {
	// 查询存在的记录
	var user User
	err := db.QueryRowContext(ctx,
		"SELECT id, name, email, phone, created_at FROM users WHERE id = $1",
		1,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found")
			return nil
		}
		return err
	}

	fmt.Printf("User: ID=%d, Name=%s, Email=%s\n", user.ID, user.Name, user.Email)

	// 处理可能为 NULL 的字段
	if user.Phone.Valid {
		fmt.Printf("Phone: %s\n", user.Phone.String)
	} else {
		fmt.Println("Phone: NULL")
	}

	// 查询不存在的记录
	err = db.QueryRowContext(ctx,
		"SELECT id FROM users WHERE id = $1", 999,
	).Scan(&user.ID)

	if err == sql.ErrNoRows {
		fmt.Println("User 999 not found (expected)")
	}

	return nil
}

// queryDemo 演示多行查询
func queryDemo(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx,
		"SELECT id, name, email, phone, created_at FROM users ORDER BY id",
	)
	if err != nil {
		return err
	}
	defer rows.Close() // 必须关闭！

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt); err != nil {
			return err
		}
		users = append(users, u)
	}

	// 检查迭代过程中的错误
	if err := rows.Err(); err != nil {
		return err
	}

	fmt.Printf("Found %d users:\n", len(users))
	for _, u := range users {
		phone := "NULL"
		if u.Phone.Valid {
			phone = u.Phone.String
		}
		fmt.Printf("  - ID=%d, Name=%s, Phone=%s\n", u.ID, u.Name, phone)
	}

	return nil
}

// updateDemo 演示更新操作
func updateDemo(ctx context.Context, db *sql.DB) error {
	result, err := db.ExecContext(ctx,
		"UPDATE users SET name = $1 WHERE id = $2",
		"Alice Updated", 1,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated %d row(s)\n", rowsAffected)

	// 更新不存在的记录
	result, err = db.ExecContext(ctx,
		"UPDATE users SET name = $1 WHERE id = $2",
		"Nobody", 999,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Updated %d row(s) for ID=999 (expected 0)\n", rowsAffected)

	return nil
}

// deleteDemo 演示删除操作
func deleteDemo(ctx context.Context, db *sql.DB) error {
	result, err := db.ExecContext(ctx,
		"DELETE FROM users WHERE id = $1",
		5,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Deleted %d row(s)\n", rowsAffected)

	// 统计剩余用户
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}
	fmt.Printf("Remaining users: %d\n", count)

	return nil
}
