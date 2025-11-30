# CRUD 操作 (CRUD Operations)

## 概述

GORM 提供简洁的 API 进行数据库增删改查操作。

## 创建 (Create)

### 创建单条记录

```go
// 创建用户
user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
result := db.Create(&user)

// 检查结果
fmt.Println(user.ID)             // 获取自动生成的 ID
fmt.Println(result.Error)        // 错误信息
fmt.Println(result.RowsAffected) // 影响的行数
```

### 创建时选择字段

```go
// 只插入指定字段
db.Select("Name", "Email").Create(&user)
// INSERT INTO users (name, email) VALUES ("Alice", "alice@example.com")

// 忽略指定字段
db.Omit("Age", "CreatedAt").Create(&user)
// INSERT INTO users (name, email, updated_at) VALUES (...)
```

### 批量创建

```go
users := []User{
    {Name: "Alice", Email: "alice@example.com"},
    {Name: "Bob", Email: "bob@example.com"},
    {Name: "Charlie", Email: "charlie@example.com"},
}

result := db.Create(&users)
fmt.Println(result.RowsAffected) // 3

// 获取所有创建的 ID
for _, user := range users {
    fmt.Println(user.ID)
}
```

### 分批创建

```go
// 每批 100 条
users := make([]User, 1000)
// ... 填充数据

db.CreateInBatches(&users, 100)
// 执行 10 次 INSERT，每次 100 条
```

### 使用 Map 创建

```go
// 使用 map
db.Model(&User{}).Create(map[string]interface{}{
    "Name":  "Alice",
    "Email": "alice@example.com",
    "Age":   25,
})

// 批量 map
db.Model(&User{}).Create([]map[string]interface{}{
    {"Name": "Alice", "Email": "alice@example.com"},
    {"Name": "Bob", "Email": "bob@example.com"},
})
```

### Upsert (创建或更新)

```go
import "gorm.io/gorm/clause"

// 冲突时更新所有字段
db.Clauses(clause.OnConflict{
    UpdateAll: true,
}).Create(&user)

// 冲突时更新指定字段
db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "email"}},
    DoUpdates: clause.AssignmentColumns([]string{"name", "age"}),
}).Create(&user)

// 冲突时什么都不做
db.Clauses(clause.OnConflict{DoNothing: true}).Create(&user)
```

## 查询 (Read)

### 查询单条记录

```go
var user User

// 按主键查询
db.First(&user, 1)                      // SELECT * FROM users WHERE id = 1 ORDER BY id LIMIT 1
db.First(&user, "id = ?", 1)            // 条件查询
db.First(&user, "10")                   // 字符串主键

// First vs Take vs Last
db.First(&user)  // 按主键升序，取第一条
db.Take(&user)   // 不排序，取一条
db.Last(&user)   // 按主键降序，取第一条

// 检查记录是否找到
result := db.First(&user, 100)
if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        // 记录不存在
    }
}

// 使用 RowsAffected
if result.RowsAffected == 0 {
    // 没有找到记录
}
```

### 查询多条记录

```go
var users []User

// 查询所有
db.Find(&users)

// 条件查询
db.Where("age > ?", 18).Find(&users)

// 按主键批量查询
db.Find(&users, []int{1, 2, 3})
// SELECT * FROM users WHERE id IN (1, 2, 3)
```

### 条件查询

```go
// 字符串条件
db.Where("name = ?", "Alice").First(&user)
db.Where("name <> ?", "Alice").Find(&users)
db.Where("name IN ?", []string{"Alice", "Bob"}).Find(&users)
db.Where("name LIKE ?", "%ali%").Find(&users)
db.Where("name = ? AND age >= ?", "Alice", 25).First(&user)
db.Where("updated_at > ?", time.Now().Add(-7*24*time.Hour)).Find(&users)
db.Where("created_at BETWEEN ? AND ?", start, end).Find(&users)

// Struct 条件（零值字段会被忽略）
db.Where(&User{Name: "Alice", Age: 25}).First(&user)
// SELECT * FROM users WHERE name = "Alice" AND age = 25

// Map 条件（不会忽略零值）
db.Where(map[string]interface{}{"Name": "Alice", "Age": 0}).Find(&users)
// SELECT * FROM users WHERE name = "Alice" AND age = 0

// 指定 Struct 查询字段
db.Where(&User{Name: "Alice"}, "Name", "Age").Find(&users)
// SELECT * FROM users WHERE name = "Alice" AND age = 0
```

### Not 条件

```go
db.Not("name = ?", "Alice").Find(&users)
// SELECT * FROM users WHERE NOT name = "Alice"

db.Not(map[string]interface{}{"name": []string{"Alice", "Bob"}}).Find(&users)
// SELECT * FROM users WHERE name NOT IN ("Alice", "Bob")

db.Not(User{Name: "Alice"}).First(&user)
// SELECT * FROM users WHERE name <> "Alice"
```

### Or 条件

```go
db.Where("name = ?", "Alice").Or("name = ?", "Bob").Find(&users)
// SELECT * FROM users WHERE name = "Alice" OR name = "Bob"

db.Where("name = ?", "Alice").Or(User{Name: "Bob", Age: 25}).Find(&users)
```

### 内联条件

```go
// First/Find 的第二个参数
db.First(&user, "name = ?", "Alice")
db.Find(&users, "name <> ? AND age > ?", "Alice", 18)
db.Find(&users, User{Age: 25})
db.Find(&users, map[string]interface{}{"age": 25})
```

### 选择字段

```go
// Select
db.Select("name", "age").Find(&users)
db.Select([]string{"name", "age"}).Find(&users)
// SELECT name, age FROM users

// 查询到 struct
type Result struct {
    Name string
    Age  int
}
var results []Result
db.Model(&User{}).Select("name", "age").Find(&results)

// Distinct
db.Distinct("name").Find(&users)
```

## 更新 (Update)

### 更新单个字段

```go
// 先查询再更新
db.First(&user)
db.Model(&user).Update("name", "Alice Updated")
// UPDATE users SET name = "Alice Updated", updated_at = NOW() WHERE id = 1

// 使用 Where
db.Model(&User{}).Where("id = ?", 1).Update("name", "Alice")
```

### 更新多个字段

```go
// 使用 struct（零值字段不会被更新）
db.Model(&user).Updates(User{Name: "Alice", Age: 26})
// UPDATE users SET name = "Alice", age = 26, updated_at = NOW() WHERE id = 1

// 使用 map（零值会被更新）
db.Model(&user).Updates(map[string]interface{}{
    "name": "Alice",
    "age":  0,  // 会被更新为 0
})

// Select + Updates（强制更新零值）
db.Model(&user).Select("Name", "Age").Updates(User{Name: "Alice", Age: 0})
// UPDATE users SET name = "Alice", age = 0 WHERE id = 1

// Omit
db.Model(&user).Omit("Age").Updates(User{Name: "Alice", Age: 0})
// 只更新 name，忽略 age
```

### 批量更新

```go
// 更新所有匹配记录
db.Model(&User{}).Where("status = ?", "inactive").Update("status", "deleted")

// 更新多个字段
db.Model(&User{}).Where("age < ?", 18).Updates(map[string]interface{}{
    "status": "minor",
    "role":   "restricted",
})
```

### 更新表达式

```go
import "gorm.io/gorm/clause"

// SQL 表达式
db.Model(&user).Update("age", gorm.Expr("age + ?", 1))
// UPDATE users SET age = age + 1 WHERE id = 1

db.Model(&user).Updates(map[string]interface{}{
    "age":        gorm.Expr("age + ?", 1),
    "updated_at": gorm.Expr("NOW()"),
})

// 子查询更新
db.Model(&user).Update("company_name",
    db.Model(&Company{}).Select("name").Where("companies.id = users.company_id"),
)
```

### 不触发 Hooks 和时间跟踪

```go
// UpdateColumn 不触发 Hooks，不更新时间
db.Model(&user).UpdateColumn("name", "Alice")

// UpdateColumns
db.Model(&user).UpdateColumns(map[string]interface{}{
    "name": "Alice",
    "age":  25,
})
```

### 检查更新结果

```go
result := db.Model(&user).Where("age > ?", 30).Updates(User{Name: "Senior"})

// 检查错误
if result.Error != nil {
    // 处理错误
}

// 检查更新行数
if result.RowsAffected == 0 {
    // 没有记录被更新
}
```

## 删除 (Delete)

### 删除单条记录

```go
// 删除已查询的记录
db.Delete(&user)
// DELETE FROM users WHERE id = 1

// 按主键删除
db.Delete(&User{}, 1)
db.Delete(&User{}, "10")  // 字符串主键
db.Delete(&User{}, []int{1, 2, 3})  // 批量删除

// 条件删除
db.Where("name = ?", "Alice").Delete(&User{})
```

### 软删除

```go
// 模型包含 gorm.DeletedAt 字段时自动启用软删除
type User struct {
    ID        uint
    Name      string
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 软删除
db.Delete(&user)
// UPDATE users SET deleted_at = NOW() WHERE id = 1

// 查询不包含软删除的记录
db.Find(&users)
// SELECT * FROM users WHERE deleted_at IS NULL

// 查询包含软删除的记录
db.Unscoped().Find(&users)
// SELECT * FROM users

// 查找软删除的记录
db.Unscoped().Where("deleted_at IS NOT NULL").Find(&users)
```

### 永久删除

```go
// 永久删除（跳过软删除）
db.Unscoped().Delete(&user)
// DELETE FROM users WHERE id = 1
```

### 批量删除

```go
// 删除多条记录
db.Where("age < ?", 18).Delete(&User{})
db.Delete(&User{}, "name LIKE ?", "%test%")

// ⚠️ 防止误删除所有记录
// 以下操作会报错
db.Delete(&User{})
// Error: WHERE clause is required for batch delete

// 如果确实要删除所有
db.Where("1 = 1").Delete(&User{})
// 或
db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{})
```

## 返回数据

```go
// Create + Returning
var user User
db.Clauses(clause.Returning{}).Create(&user)
// INSERT INTO users (...) RETURNING *

// Update + Returning
db.Model(&user).Clauses(clause.Returning{}).Update("age", 26)

// 指定返回字段
db.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "name"}}}).Create(&user)

// Delete + Returning
var deletedUsers []User
db.Clauses(clause.Returning{}).Where("status = ?", "inactive").Delete(&deletedUsers)
```

## 实际应用示例

### 用户服务

```go
type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
    return &UserService{db: db}
}

// Create 创建用户
func (s *UserService) Create(user *User) error {
    return s.db.Create(user).Error
}

// GetByID 按 ID 查询
func (s *UserService) GetByID(id uint) (*User, error) {
    var user User
    err := s.db.First(&user, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

// GetByEmail 按邮箱查询
func (s *UserService) GetByEmail(email string) (*User, error) {
    var user User
    err := s.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

// List 分页查询
func (s *UserService) List(page, pageSize int) ([]User, int64, error) {
    var users []User
    var total int64

    // 获取总数
    if err := s.db.Model(&User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询
    offset := (page - 1) * pageSize
    if err := s.db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, total, nil
}

// Update 更新用户
func (s *UserService) Update(id uint, updates map[string]interface{}) error {
    result := s.db.Model(&User{}).Where("id = ?", id).Updates(updates)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }
    return nil
}

// Delete 删除用户（软删除）
func (s *UserService) Delete(id uint) error {
    result := s.db.Delete(&User{}, id)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }
    return nil
}

// Restore 恢复软删除的用户
func (s *UserService) Restore(id uint) error {
    return s.db.Unscoped().Model(&User{}).Where("id = ?", id).Update("deleted_at", nil).Error
}
```

### 使用示例

```go
func main() {
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    userService := NewUserService(db)

    // 创建
    user := &User{Name: "Alice", Email: "alice@example.com"}
    if err := userService.Create(user); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created user ID: %d\n", user.ID)

    // 查询
    found, err := userService.GetByID(user.ID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found: %+v\n", found)

    // 更新
    if err := userService.Update(user.ID, map[string]interface{}{
        "name": "Alice Updated",
    }); err != nil {
        log.Fatal(err)
    }

    // 分页
    users, total, _ := userService.List(1, 10)
    fmt.Printf("Total: %d, Page 1: %d users\n", total, len(users))

    // 删除
    if err := userService.Delete(user.ID); err != nil {
        log.Fatal(err)
    }
}
```

**下一节**：[查询构建](./04-query-builder.md) - 学习高级查询方法
