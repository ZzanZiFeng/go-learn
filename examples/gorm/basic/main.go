// examples/gorm/basic/main.go
// GORM 基本使用示例
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"size:100;not null"`
	Email     string         `gorm:"size:255;uniqueIndex"`
	Age       int            `gorm:"check:age >= 0"`
	Status    string         `gorm:"size:20;default:'active'"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func main() {
	// 连接数据库
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}

	// 自动迁移
	db.AutoMigrate(&User{})

	// 演示 CRUD 操作
	fmt.Println("=== Create ===")
	createDemo(db)

	fmt.Println("\n=== Read ===")
	readDemo(db)

	fmt.Println("\n=== Update ===")
	updateDemo(db)

	fmt.Println("\n=== Delete ===")
	deleteDemo(db)

	fmt.Println("\n=== Soft Delete ===")
	softDeleteDemo(db)
}

func connectDB() (*gorm.DB, error) {
	// 优先使用 PostgreSQL
	connStr := os.Getenv("DATABASE_URL")
	if connStr != "" {
		return gorm.Open(postgres.Open(connStr), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	// 否则使用 SQLite 内存数据库
	fmt.Println("Using SQLite in-memory database")
	return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}

func createDemo(db *gorm.DB) {
	// 创建单条记录
	user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
	result := db.Create(&user)

	fmt.Printf("Created user ID: %d\n", user.ID)
	fmt.Printf("Rows affected: %d\n", result.RowsAffected)

	// 批量创建
	users := []User{
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}
	db.Create(&users)

	fmt.Printf("Batch created %d users\n", len(users))
}

func readDemo(db *gorm.DB) {
	// 按主键查询
	var user User
	db.First(&user, 1)
	fmt.Printf("First by ID: %s (%s)\n", user.Name, user.Email)

	// 条件查询
	db.Where("age > ?", 25).First(&user)
	fmt.Printf("First where age > 25: %s\n", user.Name)

	// 查询多条
	var users []User
	db.Find(&users)
	fmt.Printf("Total users: %d\n", len(users))

	// 检查记录是否存在
	result := db.First(&user, 999)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("User 999 not found")
	}
}

func updateDemo(db *gorm.DB) {
	// 查询后更新
	var user User
	db.First(&user, 1)

	// 更新单个字段
	db.Model(&user).Update("Age", 26)
	fmt.Printf("Updated age to: %d\n", user.Age)

	// 更新多个字段
	db.Model(&user).Updates(User{Name: "Alice Updated", Age: 27})
	fmt.Printf("Updated: %s, %d\n", user.Name, user.Age)

	// 使用 map 更新（可以更新零值）
	db.Model(&user).Updates(map[string]interface{}{
		"status": "premium",
	})

	// 条件更新
	db.Model(&User{}).Where("age < ?", 30).Update("status", "young")
}

func deleteDemo(db *gorm.DB) {
	// 创建测试用户
	testUser := User{Name: "ToDelete", Email: "delete@example.com", Age: 20}
	db.Create(&testUser)
	fmt.Printf("Created test user ID: %d\n", testUser.ID)

	// 删除（软删除，因为有 DeletedAt 字段）
	db.Delete(&testUser)
	fmt.Println("Deleted test user")

	// 验证软删除
	var count int64
	db.Model(&User{}).Count(&count)
	fmt.Printf("Users after soft delete: %d\n", count)

	// 包含软删除的查询
	db.Unscoped().Model(&User{}).Count(&count)
	fmt.Printf("Users including soft deleted: %d\n", count)
}

func softDeleteDemo(db *gorm.DB) {
	// 创建用户
	user := User{Name: "SoftDelete", Email: "soft@example.com", Age: 25}
	db.Create(&user)

	// 软删除
	db.Delete(&user)

	// 查询不包含
	var found User
	result := db.First(&found, user.ID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("Soft deleted user not found in normal query")
	}

	// Unscoped 可以查询到
	db.Unscoped().First(&found, user.ID)
	fmt.Printf("Found with Unscoped: %s (DeletedAt: %v)\n", found.Name, found.DeletedAt)

	// 永久删除
	db.Unscoped().Delete(&found)
	fmt.Println("Permanently deleted")
}
