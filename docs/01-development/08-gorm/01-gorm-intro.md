# GORM 入门 (GORM Introduction)

## 概述

GORM 是一个功能强大的 Go ORM 库，支持多种数据库。

## 安装

```bash
# GORM 核心
go get -u gorm.io/gorm

# 数据库驱动
go get -u gorm.io/driver/postgres  # PostgreSQL
go get -u gorm.io/driver/mysql     # MySQL
go get -u gorm.io/driver/sqlite    # SQLite
go get -u gorm.io/driver/sqlserver # SQL Server
```

## 连接数据库

### PostgreSQL

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func connectPostgres() (*gorm.DB, error) {
    dsn := "host=localhost user=user password=password dbname=testdb port=5432 sslmode=disable TimeZone=Asia/Shanghai"
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// 使用连接字符串
func connectPostgresURL() (*gorm.DB, error) {
    dsn := "postgres://user:password@localhost:5432/testdb?sslmode=disable"
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
```

### MySQL

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func connectMySQL() (*gorm.DB, error) {
    dsn := "user:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
    return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
```

### SQLite

```go
import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func connectSQLite() (*gorm.DB, error) {
    return gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
}

// 内存数据库（测试用）
func connectSQLiteMemory() (*gorm.DB, error) {
    return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
}
```

## GORM 配置

### 基本配置

```go
import (
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func connectWithConfig() (*gorm.DB, error) {
    return gorm.Open(postgres.Open(dsn), &gorm.Config{
        // 跳过默认事务
        SkipDefaultTransaction: true,

        // 命名策略
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "t_",    // 表名前缀
            SingularTable: true,    // 使用单数表名
        },

        // 时区
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },

        // 禁用外键约束
        DisableForeignKeyConstraintWhenMigrating: true,

        // 日志级别
        Logger: logger.Default.LogMode(logger.Info),
    })
}
```

### 日志配置

```go
import (
    "log"
    "os"
    "time"

    "gorm.io/gorm/logger"
)

// 自定义日志
func setupLogger() logger.Interface {
    return logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,   // 慢查询阈值
            LogLevel:                  logger.Info,   // 日志级别
            IgnoreRecordNotFoundError: true,          // 忽略 ErrRecordNotFound
            Colorful:                  true,          // 彩色输出
        },
    )
}

// 使用
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: setupLogger(),
})

// 运行时更改日志级别
db.Logger = db.Logger.LogMode(logger.Silent)
```

### 连接池配置

```go
func configurePool(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }

    // 最大空闲连接数
    sqlDB.SetMaxIdleConns(10)

    // 最大打开连接数
    sqlDB.SetMaxOpenConns(100)

    // 连接最大生命周期
    sqlDB.SetConnMaxLifetime(time.Hour)

    return nil
}
```

## 模型定义

### 基本模型

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"uniqueIndex;size:255"`
    Age       int
    Birthday  *time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 使用 gorm.Model

```go
// gorm.Model 定义
type Model struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 继承 gorm.Model
type User struct {
    gorm.Model  // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
    Name  string
    Email string
}
```

### 自定义表名

```go
// 方法 1: 实现 Tabler 接口
func (User) TableName() string {
    return "app_users"
}

// 方法 2: 使用配置
db.Table("app_users").Create(&user)

// 方法 3: 全局表名前缀
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    NamingStrategy: schema.NamingStrategy{
        TablePrefix: "app_",
    },
})
```

## 自动迁移

### 基本迁移

```go
// 自动创建/更新表结构
db.AutoMigrate(&User{})

// 迁移多个模型
db.AutoMigrate(&User{}, &Post{}, &Comment{})
```

### 迁移选项

```go
// 检查表是否存在
if db.Migrator().HasTable(&User{}) {
    // 表存在
}

// 创建表
db.Migrator().CreateTable(&User{})

// 删除表
db.Migrator().DropTable(&User{})

// 重命名表
db.Migrator().RenameTable(&User{}, &UserInfo{})

// 添加列
db.Migrator().AddColumn(&User{}, "Phone")

// 删除列
db.Migrator().DropColumn(&User{}, "Phone")

// 修改列
db.Migrator().AlterColumn(&User{}, "Name")

// 检查列是否存在
db.Migrator().HasColumn(&User{}, "Name")

// 创建索引
db.Migrator().CreateIndex(&User{}, "Name")
db.Migrator().CreateIndex(&User{}, "idx_name")

// 删除索引
db.Migrator().DropIndex(&User{}, "Name")
db.Migrator().DropIndex(&User{}, "idx_name")

// 检查索引是否存在
db.Migrator().HasIndex(&User{}, "Name")
```

### 迁移注意事项

```go
// ⚠️ AutoMigrate 只会:
// - 创建表
// - 添加缺失的列
// - 添加缺失的索引

// ⚠️ AutoMigrate 不会:
// - 删除列
// - 删除索引
// - 修改列类型（部分情况）

// 生产环境建议使用专门的迁移工具
// 如 golang-migrate 或 goose
```

## 快速 CRUD

### 创建

```go
// 创建单条记录
user := User{Name: "Alice", Email: "alice@example.com"}
result := db.Create(&user)

fmt.Println(user.ID)             // 获取插入的 ID
fmt.Println(result.Error)        // 错误
fmt.Println(result.RowsAffected) // 影响的行数

// 批量创建
users := []User{
    {Name: "Alice", Email: "alice@example.com"},
    {Name: "Bob", Email: "bob@example.com"},
}
db.Create(&users)
```

### 查询

```go
// 查询单条
var user User
db.First(&user, 1)                    // 按主键
db.First(&user, "id = ?", 1)          // 条件查询
db.Where("name = ?", "Alice").First(&user)

// 查询多条
var users []User
db.Find(&users)                       // 所有
db.Where("age > ?", 18).Find(&users)  // 条件

// 检查是否存在
var count int64
db.Model(&User{}).Where("email = ?", email).Count(&count)
exists := count > 0
```

### 更新

```go
// 更新单个字段
db.Model(&user).Update("Name", "Alice Updated")

// 更新多个字段
db.Model(&user).Updates(User{Name: "Alice", Age: 26})
db.Model(&user).Updates(map[string]interface{}{"Name": "Alice", "Age": 26})

// 更新所有匹配记录
db.Model(&User{}).Where("status = ?", "inactive").Update("Status", "deleted")
```

### 删除

```go
// 删除单条
db.Delete(&user)
db.Delete(&User{}, 1)

// 条件删除
db.Where("name = ?", "Bob").Delete(&User{})

// 永久删除（跳过软删除）
db.Unscoped().Delete(&user)
```

## 完整示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type User struct {
    gorm.Model
    Name     string `gorm:"size:100;not null"`
    Email    string `gorm:"uniqueIndex;size:255"`
    Age      int
    Birthday *time.Time
}

func main() {
    dsn := "host=localhost user=user password=password dbname=testdb port=5432 sslmode=disable"

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        log.Fatal(err)
    }

    // 配置连接池
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    // 自动迁移
    db.AutoMigrate(&User{})

    // 创建
    user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
    if err := db.Create(&user).Error; err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created: %+v\n", user)

    // 查询
    var found User
    if err := db.First(&found, user.ID).Error; err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found: %+v\n", found)

    // 更新
    db.Model(&found).Updates(User{Name: "Alice Updated", Age: 26})

    // 查询所有
    var users []User
    db.Find(&users)
    fmt.Printf("All users: %d\n", len(users))

    // 软删除
    db.Delete(&found)

    // 查询包括已删除
    db.Unscoped().Find(&users)
    fmt.Printf("All users (including deleted): %d\n", len(users))
}
```

**下一节**：[模型定义](./02-model-definition.md) - 学习 GORM 模型和标签
