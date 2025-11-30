# 钩子 (Hooks / Callbacks)

## 概述

GORM 允许在 Create、Update、Delete、Query 操作前后执行自定义逻辑。

## 可用钩子

### 创建钩子

```go
// BeforeCreate 在插入前调用
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 返回 error 会阻止创建
    return nil
}

// AfterCreate 在插入后调用
func (u *User) AfterCreate(tx *gorm.DB) error {
    return nil
}
```

### 更新钩子

```go
// BeforeUpdate 在更新前调用
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    return nil
}

// AfterUpdate 在更新后调用
func (u *User) AfterUpdate(tx *gorm.DB) error {
    return nil
}
```

### 保存钩子（创建和更新都会触发）

```go
// BeforeSave 在创建或更新前调用
func (u *User) BeforeSave(tx *gorm.DB) error {
    return nil
}

// AfterSave 在创建或更新后调用
func (u *User) AfterSave(tx *gorm.DB) error {
    return nil
}
```

### 删除钩子

```go
// BeforeDelete 在删除前调用
func (u *User) BeforeDelete(tx *gorm.DB) error {
    return nil
}

// AfterDelete 在删除后调用
func (u *User) AfterDelete(tx *gorm.DB) error {
    return nil
}
```

### 查询钩子

```go
// AfterFind 在查询后调用
func (u *User) AfterFind(tx *gorm.DB) error {
    return nil
}
```

## 钩子执行顺序

### 创建

```
BeforeSave
BeforeCreate
// 执行 INSERT
AfterCreate
AfterSave
```

### 更新

```
BeforeSave
BeforeUpdate
// 执行 UPDATE
AfterUpdate
AfterSave
```

### 删除

```
BeforeDelete
// 执行 DELETE
AfterDelete
```

### 查询

```
// 执行 SELECT
AfterFind
```

## 常用场景

### 自动生成 UUID

```go
import "github.com/google/uuid"

type User struct {
    ID        uint   `gorm:"primaryKey"`
    UUID      string `gorm:"type:uuid;uniqueIndex"`
    Name      string
    CreatedAt time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.UUID = uuid.New().String()
    return nil
}
```

### 密码加密

```go
import "golang.org/x/crypto/bcrypt"

type User struct {
    ID       uint
    Email    string
    Password string `json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Password != "" {
        hashedPassword, err := bcrypt.GenerateFromPassword(
            []byte(u.Password), bcrypt.DefaultCost,
        )
        if err != nil {
            return err
        }
        u.Password = string(hashedPassword)
    }
    return nil
}

// 只在密码变更时重新加密
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    if tx.Statement.Changed("Password") {
        hashedPassword, err := bcrypt.GenerateFromPassword(
            []byte(u.Password), bcrypt.DefaultCost,
        )
        if err != nil {
            return err
        }
        u.Password = string(hashedPassword)
    }
    return nil
}
```

### 自动生成 Slug

```go
import (
    "regexp"
    "strings"
)

type Post struct {
    ID      uint
    Title   string
    Slug    string `gorm:"uniqueIndex"`
    Content string
}

func (p *Post) BeforeCreate(tx *gorm.DB) error {
    if p.Slug == "" {
        p.Slug = generateSlug(p.Title)
    }
    return nil
}

func generateSlug(title string) string {
    // 转小写
    slug := strings.ToLower(title)
    // 替换空格为连字符
    slug = strings.ReplaceAll(slug, " ", "-")
    // 移除特殊字符
    reg := regexp.MustCompile("[^a-z0-9-]")
    slug = reg.ReplaceAllString(slug, "")
    // 移除多余连字符
    reg = regexp.MustCompile("-+")
    slug = reg.ReplaceAllString(slug, "-")
    return strings.Trim(slug, "-")
}
```

### 数据验证

```go
import (
    "errors"
    "regexp"
)

type User struct {
    ID    uint
    Name  string
    Email string
    Age   int
}

func (u *User) BeforeSave(tx *gorm.DB) error {
    // 验证邮箱格式
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(u.Email) {
        return errors.New("invalid email format")
    }

    // 验证年龄范围
    if u.Age < 0 || u.Age > 150 {
        return errors.New("invalid age")
    }

    // 验证名称长度
    if len(u.Name) < 2 || len(u.Name) > 100 {
        return errors.New("name must be between 2 and 100 characters")
    }

    return nil
}
```

### 自动更新时间戳

```go
type Article struct {
    ID          uint
    Title       string
    Content     string
    PublishedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (a *Article) BeforeUpdate(tx *gorm.DB) error {
    // 自定义更新逻辑
    if tx.Statement.Changed("Content") {
        // 内容变更时重置发布时间
        a.PublishedAt = nil
    }
    return nil
}
```

### 级联操作

```go
type User struct {
    ID       uint
    Name     string
    Posts    []Post
    Comments []Comment
}

// 删除用户前清理关联数据
func (u *User) BeforeDelete(tx *gorm.DB) error {
    // 软删除用户的所有文章
    if err := tx.Where("author_id = ?", u.ID).Delete(&Post{}).Error; err != nil {
        return err
    }

    // 软删除用户的所有评论
    if err := tx.Where("user_id = ?", u.ID).Delete(&Comment{}).Error; err != nil {
        return err
    }

    return nil
}
```

### 审计日志

```go
type AuditLog struct {
    ID        uint
    TableName string
    RecordID  uint
    Action    string
    OldValue  string
    NewValue  string
    UserID    uint
    CreatedAt time.Time
}

type Product struct {
    ID    uint
    Name  string
    Price float64
}

func (p *Product) AfterCreate(tx *gorm.DB) error {
    return createAuditLog(tx, "products", p.ID, "CREATE", "", toJSON(p))
}

func (p *Product) AfterUpdate(tx *gorm.DB) error {
    // 获取变更前的值（需要额外查询或传入）
    return createAuditLog(tx, "products", p.ID, "UPDATE", "", toJSON(p))
}

func (p *Product) AfterDelete(tx *gorm.DB) error {
    return createAuditLog(tx, "products", p.ID, "DELETE", toJSON(p), "")
}

func createAuditLog(tx *gorm.DB, table string, recordID uint, action, oldVal, newVal string) error {
    log := AuditLog{
        TableName: table,
        RecordID:  recordID,
        Action:    action,
        OldValue:  oldVal,
        NewValue:  newVal,
        // UserID 需要从 context 获取
        CreatedAt: time.Now(),
    }
    return tx.Create(&log).Error
}

func toJSON(v interface{}) string {
    data, _ := json.Marshal(v)
    return string(data)
}
```

### 缓存失效

```go
type Product struct {
    ID    uint
    Name  string
    Price float64
}

var cache *redis.Client

func (p *Product) AfterSave(tx *gorm.DB) error {
    // 清除缓存
    return cache.Del(context.Background(), fmt.Sprintf("product:%d", p.ID)).Err()
}

func (p *Product) AfterDelete(tx *gorm.DB) error {
    return cache.Del(context.Background(), fmt.Sprintf("product:%d", p.ID)).Err()
}

func (p *Product) AfterFind(tx *gorm.DB) error {
    // 查询后可以设置缓存
    data, _ := json.Marshal(p)
    cache.Set(context.Background(), fmt.Sprintf("product:%d", p.ID), data, time.Hour)
    return nil
}
```

## 高级用法

### 条件执行

```go
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // 只在特定字段变更时执行
    if tx.Statement.Changed("Email") {
        u.EmailVerified = false
        u.EmailVerifiedAt = nil
    }

    if tx.Statement.Changed("Password") {
        // 重新加密密码
    }

    return nil
}
```

### 访问数据库

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 检查邮箱是否已存在
    var count int64
    tx.Model(&User{}).Where("email = ?", u.Email).Count(&count)
    if count > 0 {
        return errors.New("email already exists")
    }
    return nil
}
```

### 使用 Context

```go
type contextKey string

const userIDKey contextKey = "userID"

func (p *Post) BeforeCreate(tx *gorm.DB) error {
    // 从 context 获取当前用户
    if userID, ok := tx.Statement.Context.Value(userIDKey).(uint); ok {
        p.AuthorID = userID
    }
    return nil
}

// 使用
ctx := context.WithValue(context.Background(), userIDKey, uint(1))
db.WithContext(ctx).Create(&post)
```

### 跳过钩子

```go
// 跳过所有钩子
db.Session(&gorm.Session{SkipHooks: true}).Create(&user)

// 或使用 UpdateColumn/UpdateColumns（不触发 BeforeUpdate/AfterUpdate）
db.Model(&user).UpdateColumn("name", "new name")
db.Model(&user).UpdateColumns(map[string]interface{}{"name": "new name"})
```

### 注册全局钩子

```go
// 使用 Callback
func SetupCallbacks(db *gorm.DB) {
    // 注册创建前回调
    db.Callback().Create().Before("gorm:create").Register("myPlugin:beforeCreate", func(tx *gorm.DB) {
        // 全局创建前逻辑
        if tx.Statement.Schema != nil {
            // 可以访问 Schema 信息
        }
    })

    // 注册更新后回调
    db.Callback().Update().After("gorm:update").Register("myPlugin:afterUpdate", func(tx *gorm.DB) {
        // 全局更新后逻辑
    })
}
```

## 错误处理

### 阻止操作

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Age < 18 {
        // 返回错误会阻止创建
        return errors.New("user must be at least 18 years old")
    }
    return nil
}

// 使用
err := db.Create(&user).Error
if err != nil {
    // 处理错误
}
```

### 事务回滚

```go
func (o *Order) AfterCreate(tx *gorm.DB) error {
    // 如果库存不足，返回错误会回滚整个事务
    result := tx.Model(&Product{}).
        Where("id = ? AND stock >= ?", o.ProductID, o.Quantity).
        Update("stock", gorm.Expr("stock - ?", o.Quantity))

    if result.RowsAffected == 0 {
        return errors.New("insufficient stock")
    }
    return nil
}
```

## 最佳实践

### 1. 保持钩子简洁

```go
// ❌ 钩子中做太多事情
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 验证
    // 加密
    // 生成UUID
    // 发送邮件
    // 记录日志
    // ...
    return nil
}

// ✅ 拆分职责
func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.generateUUID()
    if err := u.hashPassword(); err != nil {
        return err
    }
    return u.validate()
}

func (u *User) AfterCreate(tx *gorm.DB) error {
    // 异步处理非关键操作
    go sendWelcomeEmail(u.Email)
    return nil
}
```

### 2. 避免无限循环

```go
// ❌ 可能导致无限循环
func (u *User) AfterUpdate(tx *gorm.DB) error {
    tx.Model(u).Update("updated_count", u.UpdatedCount + 1)  // 会再次触发 AfterUpdate
    return nil
}

// ✅ 使用 UpdateColumn 跳过钩子
func (u *User) AfterUpdate(tx *gorm.DB) error {
    tx.Model(u).UpdateColumn("updated_count", u.UpdatedCount + 1)
    return nil
}
```

### 3. 正确处理错误

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if err := u.validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

**下一节**：[事务](./08-transactions.md) - 学习 GORM 事务处理
