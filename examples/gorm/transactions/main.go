// examples/gorm/transactions/main.go
// GORM 事务处理示例
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Account 账户
type Account struct {
	ID        uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"uniqueIndex"`
	Balance   float64 `gorm:"type:decimal(15,2);default:0"`
	Version   int     `gorm:"default:0"` // 乐观锁
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TransferLog 转账记录
type TransferLog struct {
	ID            uint    `gorm:"primaryKey"`
	FromAccountID uint    `gorm:"index"`
	ToAccountID   uint    `gorm:"index"`
	Amount        float64 `gorm:"type:decimal(15,2)"`
	Status        string  `gorm:"size:20"`
	CreatedAt     time.Time
}

// Order 订单
type Order struct {
	ID        uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"index"`
	Amount    float64 `gorm:"type:decimal(15,2)"`
	Status    string  `gorm:"size:20;default:'pending'"`
	CreatedAt time.Time
}

// OrderItem 订单项
type OrderItem struct {
	ID        uint    `gorm:"primaryKey"`
	OrderID   uint    `gorm:"index"`
	ProductID uint    `gorm:"index"`
	Quantity  int
	Price     float64 `gorm:"type:decimal(10,2)"`
}

// Product 商品
type Product struct {
	ID    uint    `gorm:"primaryKey"`
	Name  string  `gorm:"size:100"`
	Price float64 `gorm:"type:decimal(10,2)"`
	Stock int     `gorm:"default:0"`
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrAccountNotFound     = errors.New("account not found")
)

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	// 迁移
	db.AutoMigrate(&Account{}, &TransferLog{}, &Order{}, &OrderItem{}, &Product{})

	// 初始化数据
	initData(db)

	fmt.Println("=== Basic Transaction ===")
	basicTransactionDemo(db)

	fmt.Println("\n=== Transaction Method ===")
	transactionMethodDemo(db)

	fmt.Println("\n=== Transfer with Transaction ===")
	transferDemo(db)

	fmt.Println("\n=== Order Creation ===")
	orderDemo(db)

	fmt.Println("\n=== Nested Transaction ===")
	nestedTransactionDemo(db)
}

func initData(db *gorm.DB) {
	// 创建账户
	accounts := []Account{
		{UserID: 1, Balance: 1000},
		{UserID: 2, Balance: 500},
		{UserID: 3, Balance: 0},
	}
	db.Create(&accounts)

	// 创建商品
	products := []Product{
		{Name: "Product A", Price: 100, Stock: 10},
		{Name: "Product B", Price: 200, Stock: 5},
	}
	db.Create(&products)

	fmt.Println("Initialized test data")
}

func basicTransactionDemo(db *gorm.DB) {
	// 手动事务
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Begin error: %v", tx.Error)
		return
	}

	// 使用 defer 确保回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 事务中的操作
	var account Account
	if err := tx.First(&account, 1).Error; err != nil {
		tx.Rollback()
		log.Printf("Query error: %v", err)
		return
	}

	if err := tx.Model(&account).Update("balance", account.Balance+100).Error; err != nil {
		tx.Rollback()
		log.Printf("Update error: %v", err)
		return
	}

	// 提交
	if err := tx.Commit().Error; err != nil {
		log.Printf("Commit error: %v", err)
		return
	}

	fmt.Println("Basic transaction committed")
	printAccounts(db)
}

func transactionMethodDemo(db *gorm.DB) {
	// 使用 Transaction 方法（推荐）
	err := db.Transaction(func(tx *gorm.DB) error {
		var account Account
		if err := tx.First(&account, 2).Error; err != nil {
			return err
		}

		if err := tx.Model(&account).Update("balance", account.Balance+50).Error; err != nil {
			return err
		}

		// 返回 nil 自动提交，返回 error 自动回滚
		return nil
	})

	if err != nil {
		log.Printf("Transaction error: %v", err)
		return
	}

	fmt.Println("Transaction method completed")
	printAccounts(db)
}

func transferDemo(db *gorm.DB) {
	// 成功的转账
	err := transfer(db, 1, 2, 200)
	if err != nil {
		log.Printf("Transfer error: %v", err)
	} else {
		fmt.Println("Transfer 200 from account 1 to 2: SUCCESS")
	}
	printAccounts(db)

	// 失败的转账（余额不足）
	err = transfer(db, 2, 1, 10000)
	if errors.Is(err, ErrInsufficientBalance) {
		fmt.Println("Transfer 10000: FAILED (insufficient balance)")
	}
	printAccounts(db)
}

func transfer(db *gorm.DB, fromID, toID uint, amount float64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var from, to Account

		// 锁定账户（FOR UPDATE）
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&from, fromID).Error; err != nil {
			return ErrAccountNotFound
		}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&to, toID).Error; err != nil {
			return ErrAccountNotFound
		}

		// 检查余额
		if from.Balance < amount {
			return ErrInsufficientBalance
		}

		// 执行转账
		if err := tx.Model(&from).Update("balance", from.Balance-amount).Error; err != nil {
			return err
		}

		if err := tx.Model(&to).Update("balance", to.Balance+amount).Error; err != nil {
			return err
		}

		// 记录日志
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

func orderDemo(db *gorm.DB) {
	ctx := context.Background()
	order, err := createOrder(ctx, db, 1, []OrderItemInput{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 1},
	})

	if err != nil {
		log.Printf("Order creation error: %v", err)
		return
	}

	fmt.Printf("Created order ID: %d, Amount: %.2f\n", order.ID, order.Amount)

	// 显示库存变化
	var products []Product
	db.Find(&products)
	fmt.Println("Product stock after order:")
	for _, p := range products {
		fmt.Printf("  %s: %d\n", p.Name, p.Stock)
	}
}

type OrderItemInput struct {
	ProductID uint
	Quantity  int
}

func createOrder(ctx context.Context, db *gorm.DB, userID uint, items []OrderItemInput) (*Order, error) {
	var order Order

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var totalAmount float64
		var orderItems []OrderItem

		// 验证并锁定商品
		for _, item := range items {
			var product Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return fmt.Errorf("product %d not found", item.ProductID)
			}

			if product.Stock < item.Quantity {
				return ErrInsufficientStock
			}

			// 扣减库存
			if err := tx.Model(&product).
				Update("stock", product.Stock-item.Quantity).Error; err != nil {
				return err
			}

			totalAmount += product.Price * float64(item.Quantity)
			orderItems = append(orderItems, OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
			})
		}

		// 创建订单
		order = Order{
			UserID: userID,
			Amount: totalAmount,
			Status: "pending",
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// 创建订单项
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func nestedTransactionDemo(db *gorm.DB) {
	err := db.Transaction(func(tx *gorm.DB) error {
		// 外层事务
		var account Account
		tx.First(&account, 3)
		tx.Model(&account).Update("balance", account.Balance+100)
		fmt.Println("Outer: Added 100 to account 3")

		// 嵌套事务（SavePoint）
		err := tx.Transaction(func(tx2 *gorm.DB) error {
			tx2.Model(&account).Update("balance", account.Balance+200)
			fmt.Println("Inner: Added 200 to account 3")

			// 模拟嵌套事务失败
			return errors.New("inner transaction failed")
		})

		if err != nil {
			fmt.Printf("Inner transaction rolled back: %v\n", err)
			// 外层事务继续
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction error: %v", err)
	}

	fmt.Println("Nested transaction completed")
	printAccounts(db)
}

func printAccounts(db *gorm.DB) {
	var accounts []Account
	db.Order("id").Find(&accounts)
	fmt.Println("Account balances:")
	for _, a := range accounts {
		fmt.Printf("  Account %d (User %d): %.2f\n", a.ID, a.UserID, a.Balance)
	}
}
