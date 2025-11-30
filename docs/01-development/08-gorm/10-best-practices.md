# GORM 最佳实践 (Best Practices)

## 概述

正确使用 GORM 可以提高性能、减少 bug，本节总结关键最佳实践。

## 性能优化

### 1. 避免 N+1 查询

```go
// ❌ N+1 问题
var users []User
db.Find(&users)
for _, user := range users {
    db.Model(&user).Association("Posts").Find(&user.Posts)
}

// ✅ 使用 Preload
db.Preload("Posts").Find(&users)

// ✅ 或使用 Joins（单表关联）
db.Joins("Company").Find(&users)
```

### 2. 选择需要的字段

```go
// ❌ 查询所有字段
db.Find(&users)

// ✅ 只查询需要的字段
db.Select("id", "name", "email").Find(&users)

// ✅ 使用 DTO
type UserDTO struct {
    ID    uint
    Name  string
    Email string
}
var dtos []UserDTO
db.Model(&User{}).Select("id", "name", "email").Scan(&dtos)
```

### 3. 使用批量操作

```go
// ❌ 循环插入
for _, user := range users {
    db.Create(&user)
}

// ✅ 批量插入
db.Create(&users)

// ✅ 分批插入（大量数据）
db.CreateInBatches(&users, 100)
```

### 4. 禁用默认事务

```go
// 单个操作不需要事务包装
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
    SkipDefaultTransaction: true,
})

// 需要事务时手动开启
db.Transaction(func(tx *gorm.DB) error {
    // ...
    return nil
})
```

### 5. 使用 Prepared Statement

```go
// 启用 prepared statement 缓存
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
    PrepareStmt: true,
})

// 或 Session 级别
tx := db.Session(&gorm.Session{PrepareStmt: true})
```

### 6. 合理配置连接池

```go
sqlDB, _ := db.DB()

// 根据应用负载调整
sqlDB.SetMaxIdleConns(10)           // 空闲连接数
sqlDB.SetMaxOpenConns(100)          // 最大连接数
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
sqlDB.SetConnMaxIdleTime(10*time.Minute) // 空闲连接最大时间
```

### 7. 索引优化

```go
type User struct {
    ID        uint   `gorm:"primaryKey"`
    Email     string `gorm:"uniqueIndex;size:255"`                // 唯一索引
    Status    string `gorm:"index;size:20"`                       // 普通索引
    Name      string `gorm:"index:idx_name_status"`               // 复合索引
    CreatedAt time.Time `gorm:"index:idx_created,sort:desc"`      // 降序索引
}

// 确保常用查询条件有索引
// 使用 EXPLAIN ANALYZE 验证查询计划
```

## 代码组织

### 1. 模型定义

```go
// models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    UUID      string         `gorm:"type:uuid;uniqueIndex" json:"uuid"`
    Name      string         `gorm:"size:100;not null" json:"name"`
    Email     string         `gorm:"size:255;uniqueIndex" json:"email"`
    Password  string         `gorm:"size:255" json:"-"`
    Status    UserStatus     `gorm:"size:20;default:'active'" json:"status"`
    Role      string         `gorm:"size:20;default:'user'" json:"role"`
    Profile   *Profile       `json:"profile,omitempty"`
    Posts     []Post         `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserStatus 用户状态枚举
type UserStatus string

const (
    UserStatusActive   UserStatus = "active"
    UserStatusInactive UserStatus = "inactive"
    UserStatusBanned   UserStatus = "banned"
)

// TableName 自定义表名
func (User) TableName() string {
    return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.UUID == "" {
        u.UUID = uuid.New().String()
    }
    return nil
}
```

### 2. Repository 模式

```go
// repository/user_repository.go
package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"
    "myapp/models"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// FindByID 按 ID 查找
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// FindByEmail 按邮箱查找
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// List 分页列表
func (r *UserRepository) List(ctx context.Context, opts ListOptions) ([]models.User, int64, error) {
    var users []models.User
    var total int64

    query := r.db.WithContext(ctx).Model(&models.User{})

    // 应用过滤条件
    if opts.Status != "" {
        query = query.Where("status = ?", opts.Status)
    }

    // 获取总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 获取数据
    offset := (opts.Page - 1) * opts.PageSize
    err := query.Offset(offset).Limit(opts.PageSize).
        Order("created_at DESC").
        Find(&users).Error

    return users, total, err
}

// WithTx 使用事务
func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
    return &UserRepository{db: tx}
}
```

### 3. Service 层

```go
// service/user_service.go
package service

import (
    "context"
    "errors"

    "gorm.io/gorm"
    "myapp/models"
    "myapp/repository"
)

type UserService struct {
    db       *gorm.DB
    userRepo *repository.UserRepository
}

func NewUserService(db *gorm.DB) *UserService {
    return &UserService{
        db:       db,
        userRepo: repository.NewUserRepository(db),
    }
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
    // 检查邮箱是否已存在
    existing, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }
    if existing != nil {
        return nil, errors.New("email already exists")
    }

    // 创建用户
    user := &models.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password, // 密码会在 BeforeCreate 中加密
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

// Transfer 转账（事务示例）
func (s *UserService) Transfer(ctx context.Context, fromID, toID uint, amount float64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        accountRepo := repository.NewAccountRepository(tx)

        // 获取并锁定账户
        from, err := accountRepo.FindByIDForUpdate(ctx, fromID)
        if err != nil {
            return err
        }

        to, err := accountRepo.FindByIDForUpdate(ctx, toID)
        if err != nil {
            return err
        }

        // 检查余额
        if from.Balance < amount {
            return errors.New("insufficient balance")
        }

        // 执行转账
        from.Balance -= amount
        to.Balance += amount

        if err := accountRepo.Update(ctx, from); err != nil {
            return err
        }

        return accountRepo.Update(ctx, to)
    })
}
```

## 错误处理

### 1. 检查特定错误

```go
err := db.First(&user, id).Error
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf("query failed: %w", err)
}
```

### 2. 检查约束冲突

```go
import "github.com/jackc/pgx/v5/pgconn"

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
    err := r.db.WithContext(ctx).Create(user).Error
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            switch pgErr.Code {
            case "23505": // unique_violation
                return ErrDuplicateEmail
            case "23503": // foreign_key_violation
                return ErrInvalidReference
            }
        }
        return err
    }
    return nil
}
```

### 3. 包装业务错误

```go
var (
    ErrUserNotFound   = errors.New("user not found")
    ErrDuplicateEmail = errors.New("email already exists")
    ErrInvalidInput   = errors.New("invalid input")
)

type BusinessError struct {
    Code    string
    Message string
    Err     error
}

func (e *BusinessError) Error() string {
    return e.Message
}

func (e *BusinessError) Unwrap() error {
    return e.Err
}
```

## 安全实践

### 1. 防止 SQL 注入

```go
// ✅ 使用参数化查询
db.Where("name = ?", userInput).Find(&users)

// ✅ 使用结构体
db.Where(&User{Name: userInput}).Find(&users)

// ❌ 字符串拼接
db.Where("name = '" + userInput + "'").Find(&users)
```

### 2. 验证动态字段

```go
// 白名单验证排序字段
func ValidateSortField(field string) string {
    allowed := map[string]bool{
        "name": true, "email": true, "created_at": true,
    }
    if allowed[field] {
        return field
    }
    return "id"  // 默认
}

db.Order(ValidateSortField(sortField) + " " + sortOrder).Find(&users)
```

### 3. 限制批量操作

```go
// 防止误删除所有数据
db.Where("1 = 1").Delete(&User{})  // 需要明确条件

// 或禁用全局更新
db.Session(&gorm.Session{AllowGlobalUpdate: false})
```

### 4. 敏感数据处理

```go
type User struct {
    ID       uint
    Email    string
    Password string `json:"-" gorm:"column:password"` // JSON 不输出
}

// 密码加密
func (u *User) BeforeCreate(tx *gorm.DB) error {
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
    u.Password = string(hashedPassword)
    return nil
}
```

## 测试实践

### 1. 使用 SQLite 内存数据库

```go
func setupTestDB() *gorm.DB {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    db.AutoMigrate(&User{}, &Post{})
    return db
}

func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB()
    repo := NewUserRepository(db)

    user := &User{Name: "Test", Email: "test@example.com"}
    err := repo.Create(context.Background(), user)

    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
}
```

### 2. 事务回滚测试

```go
func TestWithTransaction(t *testing.T) {
    db := setupTestDB()

    // 每个测试使用事务并回滚
    tx := db.Begin()
    defer tx.Rollback()

    repo := NewUserRepository(tx)

    // 测试代码...
}
```

### 3. Mock Repository

```go
type UserRepositoryInterface interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uint) (*User, error)
    // ...
}

// 使用接口便于 mock 测试
type mockUserRepository struct {
    users map[uint]*User
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, nil
}
```

## 监控和日志

### 1. 自定义 Logger

```go
type GormLogger struct {
    logger *zap.Logger
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
    return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    l.logger.Sugar().Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    l.logger.Sugar().Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    l.logger.Sugar().Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()

    fields := []zap.Field{
        zap.Duration("elapsed", elapsed),
        zap.Int64("rows", rows),
        zap.String("sql", sql),
    }

    if err != nil {
        fields = append(fields, zap.Error(err))
        l.logger.Error("query failed", fields...)
    } else if elapsed > time.Second {
        l.logger.Warn("slow query", fields...)
    } else {
        l.logger.Debug("query", fields...)
    }
}
```

### 2. 添加查询追踪

```go
func AddQueryTracing(db *gorm.DB) {
    db.Callback().Query().Before("gorm:query").Register("tracing:before", func(tx *gorm.DB) {
        tx.Statement.Context = context.WithValue(tx.Statement.Context, "query_start", time.Now())
    })

    db.Callback().Query().After("gorm:query").Register("tracing:after", func(tx *gorm.DB) {
        if start, ok := tx.Statement.Context.Value("query_start").(time.Time); ok {
            elapsed := time.Since(start)
            // 记录到 metrics
            queryDuration.Observe(elapsed.Seconds())
        }
    })
}
```

## 常见陷阱

### 1. 更新零值

```go
// ❌ 零值不会被更新
db.Model(&user).Updates(User{Age: 0, Status: ""})

// ✅ 使用 map
db.Model(&user).Updates(map[string]interface{}{"age": 0, "status": ""})

// ✅ 使用 Select
db.Model(&user).Select("Age", "Status").Updates(User{Age: 0, Status: ""})
```

### 2. Preload 性能

```go
// ❌ 预加载过多数据
db.Preload("Posts").Preload("Posts.Comments").Preload("Posts.Comments.User").Find(&users)

// ✅ 按需预加载，限制数量
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "title", "author_id").Limit(10)
}).Find(&users)
```

### 3. 事务中的错误处理

```go
// ❌ 忘记处理错误
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&user1)  // 可能失败
    tx.Create(&user2)
    return nil
})

// ✅ 检查每个操作
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user1).Error; err != nil {
        return err
    }
    if err := tx.Create(&user2).Error; err != nil {
        return err
    }
    return nil
})
```

**下一节**：[练习](./exercises.md) - GORM 实践练习
