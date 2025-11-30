// examples/gorm/query/main.go
// GORM 查询构建示例
package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User 用户
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:255;uniqueIndex"`
	Age       int       `gorm:"check:age >= 0"`
	Status    string    `gorm:"size:20;default:'active'"`
	Role      string    `gorm:"size:20;default:'user'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Order 订单
type Order struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Amount    float64   `gorm:"type:decimal(10,2)"`
	Status    string    `gorm:"size:20;default:'pending'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&User{}, &Order{})
	initTestData(db)

	fmt.Println("=== Where Conditions ===")
	whereDemo(db)

	fmt.Println("\n=== Order & Limit ===")
	orderLimitDemo(db)

	fmt.Println("\n=== Group & Having ===")
	groupHavingDemo(db)

	fmt.Println("\n=== Select Fields ===")
	selectDemo(db)

	fmt.Println("\n=== Joins ===")
	joinDemo(db)

	fmt.Println("\n=== Scopes ===")
	scopesDemo(db)

	fmt.Println("\n=== Pluck & Scan ===")
	pluckScanDemo(db)

	fmt.Println("\n=== Subquery ===")
	subqueryDemo(db)
}

func initTestData(db *gorm.DB) {
	users := []User{
		{Name: "Alice", Email: "alice@example.com", Age: 25, Status: "active", Role: "admin"},
		{Name: "Bob", Email: "bob@example.com", Age: 30, Status: "active", Role: "user"},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35, Status: "inactive", Role: "user"},
		{Name: "David", Email: "david@example.com", Age: 28, Status: "active", Role: "user"},
		{Name: "Eve", Email: "eve@example.com", Age: 22, Status: "active", Role: "moderator"},
	}
	db.Create(&users)

	orders := []Order{
		{UserID: 1, Amount: 100, Status: "completed"},
		{UserID: 1, Amount: 200, Status: "completed"},
		{UserID: 2, Amount: 150, Status: "pending"},
		{UserID: 2, Amount: 300, Status: "completed"},
		{UserID: 3, Amount: 50, Status: "cancelled"},
		{UserID: 4, Amount: 250, Status: "completed"},
	}
	db.Create(&orders)

	fmt.Printf("Created %d users and %d orders\n", len(users), len(orders))
}

func whereDemo(db *gorm.DB) {
	var users []User

	// 字符串条件
	db.Where("age > ?", 25).Find(&users)
	fmt.Printf("Age > 25: %d users\n", len(users))

	// Struct 条件（零值被忽略）
	db.Where(&User{Status: "active", Role: "user"}).Find(&users)
	fmt.Printf("Active users: %d\n", len(users))

	// Map 条件（零值不被忽略）
	db.Where(map[string]interface{}{"status": "active"}).Find(&users)
	fmt.Printf("Active (map): %d users\n", len(users))

	// IN 条件
	db.Where("role IN ?", []string{"admin", "moderator"}).Find(&users)
	fmt.Printf("Admin or Moderator: %d users\n", len(users))

	// LIKE 条件
	db.Where("email LIKE ?", "%example.com").Find(&users)
	fmt.Printf("Email like example.com: %d users\n", len(users))

	// BETWEEN
	db.Where("age BETWEEN ? AND ?", 25, 30).Find(&users)
	fmt.Printf("Age 25-30: %d users\n", len(users))

	// Not
	db.Not("status = ?", "inactive").Find(&users)
	fmt.Printf("Not inactive: %d users\n", len(users))

	// Or
	db.Where("role = ?", "admin").Or("role = ?", "moderator").Find(&users)
	fmt.Printf("Admin or Moderator (Or): %d users\n", len(users))
}

func orderLimitDemo(db *gorm.DB) {
	var users []User

	// Order
	db.Order("age desc").Find(&users)
	fmt.Printf("Ordered by age desc, first: %s (%d)\n", users[0].Name, users[0].Age)

	// Multiple Order
	db.Order("role").Order("age desc").Find(&users)
	fmt.Println("Ordered by role, then age desc")

	// Limit & Offset (分页)
	db.Order("id").Limit(2).Offset(0).Find(&users)
	fmt.Printf("Page 1 (2 per page): ")
	for _, u := range users {
		fmt.Printf("%s ", u.Name)
	}
	fmt.Println()

	db.Order("id").Limit(2).Offset(2).Find(&users)
	fmt.Printf("Page 2 (2 per page): ")
	for _, u := range users {
		fmt.Printf("%s ", u.Name)
	}
	fmt.Println()
}

func groupHavingDemo(db *gorm.DB) {
	// Group By
	type RoleCount struct {
		Role  string
		Count int64
	}
	var roleCounts []RoleCount
	db.Model(&User{}).Select("role, count(*) as count").Group("role").Scan(&roleCounts)

	fmt.Println("Users by role:")
	for _, rc := range roleCounts {
		fmt.Printf("  %s: %d\n", rc.Role, rc.Count)
	}

	// Group By with Having
	type StatusStats struct {
		Status     string
		TotalCount int64
		TotalSum   float64
	}
	var statusStats []StatusStats
	db.Model(&Order{}).
		Select("status, count(*) as total_count, sum(amount) as total_sum").
		Group("status").
		Having("count(*) > ?", 1).
		Scan(&statusStats)

	fmt.Println("Order stats (count > 1):")
	for _, s := range statusStats {
		fmt.Printf("  %s: %d orders, $%.2f total\n", s.Status, s.TotalCount, s.TotalSum)
	}
}

func selectDemo(db *gorm.DB) {
	// Select specific fields
	type UserDTO struct {
		ID    uint
		Name  string
		Email string
	}
	var dtos []UserDTO
	db.Model(&User{}).Select("id", "name", "email").Find(&dtos)
	fmt.Printf("Selected %d users (id, name, email)\n", len(dtos))

	// Distinct
	var roles []string
	db.Model(&User{}).Distinct("role").Pluck("role", &roles)
	fmt.Printf("Distinct roles: %v\n", roles)

	// Count
	var count int64
	db.Model(&User{}).Where("status = ?", "active").Count(&count)
	fmt.Printf("Active users count: %d\n", count)
}

func joinDemo(db *gorm.DB) {
	// 使用原生 Join
	type UserOrder struct {
		UserName string
		OrderID  uint
		Amount   float64
	}
	var results []UserOrder
	db.Model(&Order{}).
		Select("users.name as user_name, orders.id as order_id, orders.amount").
		Joins("LEFT JOIN users ON users.id = orders.user_id").
		Where("orders.status = ?", "completed").
		Scan(&results)

	fmt.Println("Completed orders with user names:")
	for _, r := range results {
		fmt.Printf("  %s - Order #%d: $%.2f\n", r.UserName, r.OrderID, r.Amount)
	}

	// 用户订单统计
	type UserStats struct {
		UserID     uint
		UserName   string
		OrderCount int64
		TotalSpent float64
	}
	var stats []UserStats
	db.Model(&Order{}).
		Select("orders.user_id, users.name as user_name, count(*) as order_count, sum(orders.amount) as total_spent").
		Joins("JOIN users ON users.id = orders.user_id").
		Where("orders.status = ?", "completed").
		Group("orders.user_id, users.name").
		Scan(&stats)

	fmt.Println("User order statistics:")
	for _, s := range stats {
		fmt.Printf("  %s: %d orders, $%.2f total\n", s.UserName, s.OrderCount, s.TotalSpent)
	}
}

// Scope 函数
func Active(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", "active")
}

func RoleIs(role string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("role = ?", role)
	}
}

func AgeGreaterThan(age int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("age > ?", age)
	}
}

func Paginate(page, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func scopesDemo(db *gorm.DB) {
	var users []User

	// 使用单个 Scope
	db.Scopes(Active).Find(&users)
	fmt.Printf("Active users: %d\n", len(users))

	// 组合多个 Scopes
	db.Scopes(Active, RoleIs("user"), AgeGreaterThan(25)).Find(&users)
	fmt.Printf("Active users, role=user, age>25: %d\n", len(users))

	// 分页 Scope
	db.Scopes(Active, Paginate(1, 2)).Find(&users)
	fmt.Printf("Active users page 1 (size 2): %d\n", len(users))
}

func pluckScanDemo(db *gorm.DB) {
	// Pluck 单列
	var names []string
	db.Model(&User{}).Pluck("name", &names)
	fmt.Printf("User names: %v\n", names)

	var ages []int
	db.Model(&User{}).Where("status = ?", "active").Pluck("age", &ages)
	fmt.Printf("Active user ages: %v\n", ages)

	// Scan 到 map
	var result map[string]interface{}
	db.Model(&User{}).Select("id", "name", "email").First(&result)
	fmt.Printf("First user as map: %v\n", result)

	// Scan 到 slice of maps
	var results []map[string]interface{}
	db.Model(&User{}).Select("id", "name").Limit(3).Find(&results)
	fmt.Printf("First 3 users as maps: %d results\n", len(results))
}

func subqueryDemo(db *gorm.DB) {
	var users []User

	// 子查询作为条件
	// 查找有已完成订单的用户
	subQuery := db.Model(&Order{}).Select("user_id").Where("status = ?", "completed")
	db.Where("id IN (?)", subQuery).Find(&users)
	fmt.Printf("Users with completed orders: %d\n", len(users))
	for _, u := range users {
		fmt.Printf("  %s\n", u.Name)
	}

	// 子查询计算
	// 查找订单总额大于平均值的用户
	avgSubQuery := db.Model(&Order{}).Select("avg(amount)")

	type UserWithTotal struct {
		User
		TotalAmount float64
	}
	var usersWithTotal []UserWithTotal
	db.Model(&User{}).
		Select("users.*, (SELECT COALESCE(SUM(amount), 0) FROM orders WHERE orders.user_id = users.id) as total_amount").
		Having("(SELECT COALESCE(SUM(amount), 0) FROM orders WHERE orders.user_id = users.id) > (?)", avgSubQuery).
		Find(&usersWithTotal)

	fmt.Println("Users with above-average total orders:")
	for _, u := range usersWithTotal {
		fmt.Printf("  %s: $%.2f\n", u.Name, u.TotalAmount)
	}
}
