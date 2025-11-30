# 原生 SQL (Raw SQL)

## 概述

有时 GORM 的链式 API 无法满足复杂查询需求，可以使用原生 SQL。

## Raw 查询

### 基本查询

```go
// 查询到结构体
type Result struct {
    ID   uint
    Name string
    Age  int
}

var result Result
db.Raw("SELECT id, name, age FROM users WHERE id = ?", 1).Scan(&result)

// 查询到切片
var results []Result
db.Raw("SELECT id, name, age FROM users WHERE age > ?", 18).Scan(&results)
```

### 使用 Named 参数

```go
// SQL Named 参数
db.Raw("SELECT * FROM users WHERE name = @name OR age = @age",
    sql.Named("name", "Alice"),
    sql.Named("age", 25),
).Scan(&users)

// Map Named 参数
db.Raw("SELECT * FROM users WHERE name = @name OR age = @age",
    map[string]interface{}{"name": "Alice", "age": 25},
).Scan(&users)
```

### 复杂查询

```go
// CTE (Common Table Expression)
var results []UserStats
db.Raw(`
    WITH monthly_stats AS (
        SELECT
            user_id,
            DATE_TRUNC('month', created_at) as month,
            COUNT(*) as count,
            SUM(amount) as total
        FROM orders
        WHERE created_at > ?
        GROUP BY user_id, DATE_TRUNC('month', created_at)
    )
    SELECT
        u.id,
        u.name,
        ms.month,
        ms.count,
        ms.total
    FROM users u
    JOIN monthly_stats ms ON ms.user_id = u.id
    ORDER BY u.id, ms.month
`, time.Now().AddDate(-1, 0, 0)).Scan(&results)

// 窗口函数
db.Raw(`
    SELECT
        id,
        name,
        age,
        RANK() OVER (ORDER BY age DESC) as age_rank,
        AVG(age) OVER () as avg_age
    FROM users
`).Scan(&results)
```

### DryRun 模式

```go
// 只生成 SQL，不执行
stmt := db.Session(&gorm.Session{DryRun: true}).
    Raw("SELECT * FROM users WHERE id = ?", 1).
    Statement

fmt.Println(stmt.SQL.String())  // SELECT * FROM users WHERE id = ?
fmt.Println(stmt.Vars)          // [1]
```

## Exec 执行

### 基本执行

```go
// 执行 INSERT
db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")

// 执行 UPDATE
result := db.Exec("UPDATE users SET name = ? WHERE id = ?", "Alice Updated", 1)
fmt.Println(result.RowsAffected)

// 执行 DELETE
db.Exec("DELETE FROM users WHERE status = ?", "inactive")
```

### 批量操作

```go
// 批量插入
db.Exec(`
    INSERT INTO users (name, email, age) VALUES
    (?, ?, ?),
    (?, ?, ?),
    (?, ?, ?)
`, "Alice", "alice@example.com", 25,
   "Bob", "bob@example.com", 30,
   "Charlie", "charlie@example.com", 35)

// PostgreSQL COPY
// 需要使用底层连接
sqlDB, _ := db.DB()
tx, _ := sqlDB.Begin()
stmt, _ := tx.Prepare(pq.CopyIn("users", "name", "email", "age"))

for _, user := range users {
    stmt.Exec(user.Name, user.Email, user.Age)
}

stmt.Exec()
stmt.Close()
tx.Commit()
```

## 行级操作

### Row

```go
// 查询单行
row := db.Raw("SELECT name, age FROM users WHERE id = ?", 1).Row()

var name string
var age int
row.Scan(&name, &age)
```

### Rows

```go
// 查询多行
rows, err := db.Raw("SELECT name, age FROM users WHERE age > ?", 18).Rows()
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var name string
    var age int
    rows.Scan(&name, &age)
    fmt.Printf("%s: %d\n", name, age)
}
```

### ScanRows

```go
// 使用 ScanRows 扫描到结构体
rows, _ := db.Raw("SELECT * FROM users WHERE age > ?", 18).Rows()
defer rows.Close()

for rows.Next() {
    var user User
    db.ScanRows(rows, &user)
    fmt.Printf("%+v\n", user)
}
```

## SQL Builder

### 构建 Where

```go
// 使用 clause.Expr
db.Where(clause.Expr{SQL: "name LIKE ?", Vars: []interface{}{"%alice%"}}).Find(&users)

// 组合表达式
db.Clauses(clause.Where{
    Exprs: []clause.Expression{
        clause.Expr{SQL: "age > ?", Vars: []interface{}{18}},
        clause.Expr{SQL: "status = ?", Vars: []interface{}{"active"}},
    },
}).Find(&users)
```

### 自定义 Clause

```go
// 自定义 SQL 片段
type MyHint struct {
    Hint string
}

func (h MyHint) Name() string {
    return ""
}

func (h MyHint) Build(builder clause.Builder) {
    builder.WriteString(h.Hint)
    builder.WriteByte(' ')
}

func (h MyHint) MergeClause(*clause.Clause) {}

// 使用
db.Clauses(MyHint{Hint: "/*+ INDEX(users idx_name) */"}).Find(&users)
// SELECT /*+ INDEX(users idx_name) */ * FROM users
```

### ToSQL

```go
// 生成 SQL 不执行
sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
    return tx.Model(&User{}).Where("id = ?", 1).First(&User{})
})
fmt.Println(sql)  // SELECT * FROM users WHERE id = 1 ORDER BY id LIMIT 1
```

## 与 Model 结合

### Model + Raw

```go
// 使用 Model 指定表，Raw 指定查询
var count int64
db.Model(&User{}).Raw("SELECT COUNT(*) FROM users WHERE age > ?", 18).Scan(&count)

// 子查询
subQuery := db.Model(&Order{}).Select("user_id, SUM(amount) as total").
    Group("user_id").
    Having("SUM(amount) > ?", 1000)

db.Raw("SELECT u.*, t.total FROM users u JOIN (?) t ON t.user_id = u.id", subQuery).
    Scan(&results)
```

### 混合使用

```go
// 链式 + Raw 条件
db.Where("status = ?", "active").
   Where(clause.Expr{SQL: "created_at > NOW() - INTERVAL '7 days'"}).
   Find(&users)

// Raw 表达式作为值
db.Model(&User{}).Update("last_login", gorm.Expr("NOW()"))

// 原生 SQL 更新
db.Model(&Product{}).Where("id = ?", 1).
   Update("price", gorm.Expr("price * ?", 1.1))
```

## 数据库特定功能

### PostgreSQL

```go
// RETURNING
var user User
db.Raw("INSERT INTO users (name, email) VALUES (?, ?) RETURNING *",
    "Alice", "alice@example.com").Scan(&user)

// JSON 操作
db.Raw("SELECT * FROM users WHERE settings->>'theme' = ?", "dark").Scan(&users)
db.Raw("SELECT * FROM users WHERE settings @> ?", `{"notifications": true}`).Scan(&users)

// 数组操作
db.Raw("SELECT * FROM users WHERE ? = ANY(tags)", "go").Scan(&users)

// 全文搜索
db.Raw(`
    SELECT *, ts_rank(to_tsvector('english', content), plainto_tsquery('english', ?)) as rank
    FROM posts
    WHERE to_tsvector('english', content) @@ plainto_tsquery('english', ?)
    ORDER BY rank DESC
`, searchTerm, searchTerm).Scan(&posts)

// UPSERT
db.Exec(`
    INSERT INTO users (email, name, updated_at)
    VALUES (?, ?, NOW())
    ON CONFLICT (email) DO UPDATE SET
        name = EXCLUDED.name,
        updated_at = NOW()
`, email, name)
```

### MySQL

```go
// REPLACE INTO
db.Exec("REPLACE INTO users (id, name, email) VALUES (?, ?, ?)", id, name, email)

// INSERT ... ON DUPLICATE KEY UPDATE
db.Exec(`
    INSERT INTO users (email, name)
    VALUES (?, ?)
    ON DUPLICATE KEY UPDATE name = VALUES(name)
`, email, name)

// JSON
db.Raw("SELECT * FROM users WHERE JSON_EXTRACT(settings, '$.theme') = ?", "dark").Scan(&users)

// 全文搜索
db.Raw("SELECT * FROM posts WHERE MATCH(title, content) AGAINST(? IN BOOLEAN MODE)", searchTerm).
    Scan(&posts)
```

## 安全注意事项

### 参数化查询

```go
// ✅ 使用占位符（安全）
db.Raw("SELECT * FROM users WHERE name = ?", userInput).Scan(&users)
db.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)

// ❌ 字符串拼接（SQL 注入风险）
db.Raw("SELECT * FROM users WHERE name = '" + userInput + "'").Scan(&users)
```

### 动态表名/列名

```go
// 验证表名和列名
var allowedTables = map[string]bool{"users": true, "posts": true, "comments": true}
var allowedColumns = map[string]bool{"name": true, "email": true, "created_at": true}

func SafeQuery(db *gorm.DB, table, column, value string) error {
    if !allowedTables[table] {
        return errors.New("invalid table")
    }
    if !allowedColumns[column] {
        return errors.New("invalid column")
    }

    // 表名和列名是安全的，可以直接使用
    query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, column)
    return db.Raw(query, value).Scan(&results).Error
}
```

## 实际应用示例

### 报表查询

```go
type ReportService struct {
    db *gorm.DB
}

type SalesReport struct {
    Date       time.Time
    TotalSales float64
    OrderCount int64
    AvgOrder   float64
}

func (s *ReportService) GetDailySalesReport(start, end time.Time) ([]SalesReport, error) {
    var reports []SalesReport

    err := s.db.Raw(`
        SELECT
            DATE(created_at) as date,
            SUM(amount) as total_sales,
            COUNT(*) as order_count,
            AVG(amount) as avg_order
        FROM orders
        WHERE status = 'completed'
          AND created_at BETWEEN ? AND ?
        GROUP BY DATE(created_at)
        ORDER BY date
    `, start, end).Scan(&reports).Error

    return reports, err
}

type TopProduct struct {
    ProductID uint
    Name      string
    Quantity  int64
    Revenue   float64
}

func (s *ReportService) GetTopProducts(limit int, start, end time.Time) ([]TopProduct, error) {
    var products []TopProduct

    err := s.db.Raw(`
        SELECT
            p.id as product_id,
            p.name,
            SUM(oi.quantity) as quantity,
            SUM(oi.quantity * oi.price) as revenue
        FROM order_items oi
        JOIN products p ON p.id = oi.product_id
        JOIN orders o ON o.id = oi.order_id
        WHERE o.status = 'completed'
          AND o.created_at BETWEEN ? AND ?
        GROUP BY p.id, p.name
        ORDER BY revenue DESC
        LIMIT ?
    `, start, end, limit).Scan(&products).Error

    return products, err
}
```

### 复杂搜索

```go
type SearchService struct {
    db *gorm.DB
}

type SearchParams struct {
    Query     string
    Category  string
    MinPrice  float64
    MaxPrice  float64
    SortBy    string
    SortOrder string
    Page      int
    PageSize  int
}

func (s *SearchService) SearchProducts(params SearchParams) ([]Product, int64, error) {
    var products []Product
    var total int64

    // 构建基础查询
    baseQuery := `
        FROM products p
        LEFT JOIN categories c ON c.id = p.category_id
        WHERE 1=1
    `
    args := []interface{}{}

    // 动态条件
    if params.Query != "" {
        baseQuery += " AND (p.name ILIKE ? OR p.description ILIKE ?)"
        searchTerm := "%" + params.Query + "%"
        args = append(args, searchTerm, searchTerm)
    }

    if params.Category != "" {
        baseQuery += " AND c.slug = ?"
        args = append(args, params.Category)
    }

    if params.MinPrice > 0 {
        baseQuery += " AND p.price >= ?"
        args = append(args, params.MinPrice)
    }

    if params.MaxPrice > 0 {
        baseQuery += " AND p.price <= ?"
        args = append(args, params.MaxPrice)
    }

    // 获取总数
    countSQL := "SELECT COUNT(*) " + baseQuery
    s.db.Raw(countSQL, args...).Scan(&total)

    // 排序
    orderBy := "p.created_at DESC"
    allowedSorts := map[string]string{
        "name":       "p.name",
        "price":      "p.price",
        "created_at": "p.created_at",
    }
    if col, ok := allowedSorts[params.SortBy]; ok {
        order := "ASC"
        if params.SortOrder == "desc" {
            order = "DESC"
        }
        orderBy = col + " " + order
    }

    // 分页
    offset := (params.Page - 1) * params.PageSize
    selectSQL := fmt.Sprintf("SELECT p.* %s ORDER BY %s LIMIT ? OFFSET ?",
        baseQuery, orderBy)

    args = append(args, params.PageSize, offset)
    s.db.Raw(selectSQL, args...).Scan(&products)

    return products, total, nil
}
```

**下一节**：[最佳实践](./10-best-practices.md) - 学习 GORM 最佳实践
