# 查询构建 (Query Builder)

## 概述

GORM 提供链式 API 构建复杂查询，支持多种条件组合和高级功能。

## 链式操作

### 基本链式

```go
// 链式查询
db.Where("name = ?", "Alice").
   Where("age > ?", 18).
   Order("created_at DESC").
   Limit(10).
   Find(&users)

// 等价 SQL:
// SELECT * FROM users
// WHERE name = 'Alice' AND age > 18
// ORDER BY created_at DESC
// LIMIT 10
```

### 可复用作用域

```go
// 定义作用域
func Active(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "active")
}

func Adult(db *gorm.DB) *gorm.DB {
    return db.Where("age >= ?", 18)
}

func OrderByRecent(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC")
}

// 使用作用域
db.Scopes(Active, Adult, OrderByRecent).Find(&users)

// 参数化作用域
func AgeGreaterThan(age int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("age > ?", age)
    }
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        offset := (page - 1) * pageSize
        return db.Offset(offset).Limit(pageSize)
    }
}

db.Scopes(AgeGreaterThan(18), Paginate(1, 10)).Find(&users)
```

## 排序

```go
// 单字段排序
db.Order("age desc").Find(&users)
db.Order("age desc, name").Find(&users)

// 多个 Order
db.Order("age desc").Order("name").Find(&users)

// 使用 clause.OrderByColumn
db.Clauses(clause.OrderBy{
    Columns: []clause.OrderByColumn{
        {Column: clause.Column{Name: "age"}, Desc: true},
        {Column: clause.Column{Name: "name"}, Desc: false},
    },
}).Find(&users)

// 清除排序
db.Order("age desc").Find(&users1).Order(clause.OrderBy{}).Find(&users2)
```

## 分页

### Limit 和 Offset

```go
// 基本分页
db.Limit(10).Offset(0).Find(&users)   // 第1页
db.Limit(10).Offset(10).Find(&users)  // 第2页
db.Limit(10).Offset(20).Find(&users)  // 第3页

// 取消 Limit
db.Limit(10).Find(&users1).Limit(-1).Find(&users2)
// users1: LIMIT 10
// users2: 无 LIMIT
```

### 分页封装

```go
type Pagination struct {
    Page     int   `json:"page"`
    PageSize int   `json:"page_size"`
    Total    int64 `json:"total"`
}

func (p *Pagination) Offset() int {
    return (p.Page - 1) * p.PageSize
}

type PaginatedResult[T any] struct {
    Items      []T        `json:"items"`
    Pagination Pagination `json:"pagination"`
}

func Paginate[T any](db *gorm.DB, page, pageSize int) (*PaginatedResult[T], error) {
    var items []T
    var total int64

    // 获取总数
    if err := db.Model(new(T)).Count(&total).Error; err != nil {
        return nil, err
    }

    // 获取数据
    offset := (page - 1) * pageSize
    if err := db.Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
        return nil, err
    }

    return &PaginatedResult[T]{
        Items: items,
        Pagination: Pagination{
            Page:     page,
            PageSize: pageSize,
            Total:    total,
        },
    }, nil
}

// 使用
result, err := Paginate[User](db.Where("status = ?", "active"), 1, 10)
```

## 分组与聚合

### Group By

```go
type Result struct {
    Status string
    Count  int64
}

var results []Result
db.Model(&User{}).Select("status, count(*) as count").Group("status").Find(&results)
// SELECT status, count(*) as count FROM users GROUP BY status

// 复杂分组
type AgeStats struct {
    AgeGroup string
    Count    int64
    AvgAge   float64
}

db.Model(&User{}).
    Select("CASE WHEN age < 18 THEN 'minor' WHEN age < 60 THEN 'adult' ELSE 'senior' END as age_group, count(*) as count, avg(age) as avg_age").
    Group("age_group").
    Find(&results)
```

### Having

```go
// Group + Having
db.Model(&User{}).
    Select("status, count(*) as count").
    Group("status").
    Having("count(*) > ?", 5).
    Find(&results)
// SELECT status, count(*) as count FROM users
// GROUP BY status HAVING count(*) > 5
```

### 聚合函数

```go
var count int64
db.Model(&User{}).Count(&count)
// SELECT count(*) FROM users

var sum int64
db.Model(&Order{}).Select("sum(amount)").Scan(&sum)

var avg float64
db.Model(&User{}).Select("avg(age)").Scan(&avg)

// 多个聚合
type Stats struct {
    Total    int64
    AvgAge   float64
    MinAge   int
    MaxAge   int
}

var stats Stats
db.Model(&User{}).Select(
    "count(*) as total, avg(age) as avg_age, min(age) as min_age, max(age) as max_age",
).Scan(&stats)
```

## Distinct

```go
db.Distinct("name", "age").Find(&users)
// SELECT DISTINCT name, age FROM users

db.Model(&User{}).Distinct().Pluck("name", &names)
// SELECT DISTINCT name FROM users
```

## Join 查询

### 基本 Join

```go
type UserWithCompany struct {
    User
    CompanyName string
}

var results []UserWithCompany
db.Model(&User{}).
    Select("users.*, companies.name as company_name").
    Joins("LEFT JOIN companies ON companies.id = users.company_id").
    Find(&results)
```

### Join 预加载

```go
// 使用 Joins 预加载（单次查询）
db.Joins("Company").Find(&users)
// SELECT users.*, Company.* FROM users LEFT JOIN companies AS Company ON ...

// 带条件的 Joins 预加载
db.Joins("Company", db.Where(&Company{Status: "active"})).Find(&users)
```

### 多表 Join

```go
db.Model(&Order{}).
    Select("orders.id, orders.amount, users.name as user_name, products.title as product_title").
    Joins("LEFT JOIN users ON users.id = orders.user_id").
    Joins("LEFT JOIN products ON products.id = orders.product_id").
    Where("orders.status = ?", "completed").
    Find(&results)
```

## 子查询

### Where 子查询

```go
// 子查询条件
db.Where("user_id IN (?)",
    db.Model(&Admin{}).Select("user_id"),
).Find(&users)
// SELECT * FROM users WHERE user_id IN (SELECT user_id FROM admins)

// 复杂子查询
db.Where("age > (?)",
    db.Model(&User{}).Select("avg(age)"),
).Find(&users)
// SELECT * FROM users WHERE age > (SELECT avg(age) FROM users)
```

### From 子查询

```go
// 子查询作为表
subQuery := db.Model(&User{}).
    Select("status, count(*) as count").
    Group("status")

db.Table("(?) as u", subQuery).
    Where("u.count > ?", 10).
    Find(&results)
```

### Select 子查询

```go
// 子查询作为列
subQuery := db.Model(&Order{}).
    Select("count(*)").
    Where("orders.user_id = users.id")

db.Model(&User{}).
    Select("users.*, (?) as order_count", subQuery).
    Find(&users)
```

## 高级条件

### 原生 SQL 条件

```go
db.Where("name = ? AND age > ?", "Alice", 18).Find(&users)

// Named 参数
db.Where("name = @name OR age = @age",
    sql.Named("name", "Alice"),
    sql.Named("age", 25),
).Find(&users)

// Map Named 参数
db.Where("name = @name OR age = @age", map[string]interface{}{
    "name": "Alice",
    "age":  25,
}).Find(&users)
```

### FirstOrInit / FirstOrCreate

```go
// FirstOrInit: 找到或初始化（不保存）
var user User
db.FirstOrInit(&user, User{Name: "Alice"})
// 如果存在，返回找到的记录
// 如果不存在，返回初始化的结构体（未保存）

// 带默认值
db.Where(User{Name: "Alice"}).Attrs(User{Age: 25}).FirstOrInit(&user)
// 如果不存在，用 Attrs 设置默认值

// Assign 总是会赋值
db.Where(User{Name: "Alice"}).Assign(User{Age: 30}).FirstOrInit(&user)
// 无论是否存在，Age 都会被设置为 30

// FirstOrCreate: 找到或创建（会保存）
db.FirstOrCreate(&user, User{Name: "Alice"})
// 如果存在，返回找到的记录
// 如果不存在，创建并返回

db.Where(User{Name: "Alice"}).
   Attrs(User{Age: 25}).
   FirstOrCreate(&user)
```

### Find To Map

```go
var result map[string]interface{}
db.Model(&User{}).First(&result, "id = ?", 1)

var results []map[string]interface{}
db.Model(&User{}).Find(&results)
```

### Pluck

```go
// 查询单列
var names []string
db.Model(&User{}).Pluck("name", &names)

var ages []int
db.Model(&User{}).Pluck("age", &ages)

// Distinct
db.Model(&User{}).Distinct().Pluck("name", &names)
```

### Scan

```go
// 扫描到自定义结构
type UserDTO struct {
    ID    uint
    Name  string
    Email string
}

var dto UserDTO
db.Model(&User{}).Select("id", "name", "email").First(&dto)

// 扫描到 map
var result map[string]interface{}
db.Model(&User{}).Select("id", "name").Take(&result)
```

## 锁

### 悲观锁

```go
// FOR UPDATE
db.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&users)
// SELECT * FROM users FOR UPDATE

// FOR SHARE
db.Clauses(clause.Locking{Strength: "SHARE"}).Find(&users)

// NOWAIT
db.Clauses(clause.Locking{
    Strength: "UPDATE",
    Options:  "NOWAIT",
}).Find(&users)
// SELECT * FROM users FOR UPDATE NOWAIT

// SKIP LOCKED
db.Clauses(clause.Locking{
    Strength: "UPDATE",
    Options:  "SKIP LOCKED",
}).Find(&users)
```

## 复杂查询示例

### 条件构建器

```go
type UserQuery struct {
    Name      string
    Email     string
    MinAge    int
    MaxAge    int
    Status    string
    SortBy    string
    SortOrder string
    Page      int
    PageSize  int
}

func (q *UserQuery) Build(db *gorm.DB) *gorm.DB {
    if q.Name != "" {
        db = db.Where("name LIKE ?", "%"+q.Name+"%")
    }
    if q.Email != "" {
        db = db.Where("email LIKE ?", "%"+q.Email+"%")
    }
    if q.MinAge > 0 {
        db = db.Where("age >= ?", q.MinAge)
    }
    if q.MaxAge > 0 {
        db = db.Where("age <= ?", q.MaxAge)
    }
    if q.Status != "" {
        db = db.Where("status = ?", q.Status)
    }
    if q.SortBy != "" {
        order := q.SortBy
        if q.SortOrder == "desc" {
            order += " DESC"
        }
        db = db.Order(order)
    }
    if q.Page > 0 && q.PageSize > 0 {
        offset := (q.Page - 1) * q.PageSize
        db = db.Offset(offset).Limit(q.PageSize)
    }
    return db
}

// 使用
func (s *UserService) Search(query *UserQuery) ([]User, error) {
    var users []User
    err := query.Build(s.db.Model(&User{})).Find(&users).Error
    return users, err
}
```

### 动态排序

```go
func OrderByField(field string, desc bool) func(*gorm.DB) *gorm.DB {
    allowedFields := map[string]bool{
        "name": true, "email": true, "age": true, "created_at": true,
    }

    return func(db *gorm.DB) *gorm.DB {
        if !allowedFields[field] {
            field = "id"  // 默认排序
        }
        order := field
        if desc {
            order += " DESC"
        }
        return db.Order(order)
    }
}

db.Scopes(OrderByField("created_at", true)).Find(&users)
```

### 全文搜索 (PostgreSQL)

```go
// 简单搜索
db.Where("to_tsvector('english', content) @@ plainto_tsquery('english', ?)", searchTerm).Find(&posts)

// 带排名
type SearchResult struct {
    Post
    Rank float64
}

db.Model(&Post{}).
    Select("posts.*, ts_rank(to_tsvector('english', content), plainto_tsquery('english', ?)) as rank", searchTerm).
    Where("to_tsvector('english', content) @@ plainto_tsquery('english', ?)", searchTerm).
    Order("rank DESC").
    Find(&results)
```

**下一节**：[关联关系](./05-associations.md) - 学习模型关联
