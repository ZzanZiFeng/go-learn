// examples/database/connect/main.go
// 演示 database/sql 和 pgx 连接数据库
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/testdb?sslmode=disable"
	}

	// 方式 1: database/sql
	fmt.Println("=== database/sql 连接 ===")
	if err := databaseSQLDemo(connStr); err != nil {
		log.Printf("database/sql error: %v\n", err)
	}

	// 方式 2: pgx
	fmt.Println("\n=== pgx 连接 ===")
	if err := pgxDemo(connStr); err != nil {
		log.Printf("pgx error: %v\n", err)
	}
}

// databaseSQLDemo 演示 database/sql 连接
func databaseSQLDemo(connStr string) error {
	// sql.Open 只是验证参数，不会真正连接
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer db.Close()

	// 配置连接池
	db.SetMaxOpenConns(25)              // 最大打开连接数
	db.SetMaxIdleConns(5)               // 最大空闲连接数
	db.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期
	db.SetConnMaxIdleTime(1 * time.Minute)  // 空闲连接最大时间

	// Ping 才会真正连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}

	fmt.Println("Connected with database/sql!")

	// 查看连接池状态
	stats := db.Stats()
	fmt.Printf("MaxOpenConnections: %d\n", stats.MaxOpenConnections)
	fmt.Printf("OpenConnections: %d\n", stats.OpenConnections)
	fmt.Printf("Idle: %d\n", stats.Idle)

	// 查询 PostgreSQL 版本
	var version string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		return fmt.Errorf("failed to query version: %w", err)
	}
	fmt.Printf("PostgreSQL Version: %s\n", version)

	return nil
}

// pgxDemo 演示 pgx 连接池
func pgxDemo(connStr string) error {
	ctx := context.Background()

	// 解析配置
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// 配置连接池
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}
	defer pool.Close()

	fmt.Println("Connected with pgx!")

	// 查看连接池状态
	stat := pool.Stat()
	fmt.Printf("TotalConns: %d\n", stat.TotalConns())
	fmt.Printf("IdleConns: %d\n", stat.IdleConns())
	fmt.Printf("MaxConns: %d\n", stat.MaxConns())

	// 查询 PostgreSQL 版本
	var version string
	if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		return fmt.Errorf("failed to query version: %w", err)
	}
	fmt.Printf("PostgreSQL Version: %s\n", version)

	return nil
}
