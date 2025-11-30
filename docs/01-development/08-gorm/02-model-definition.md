# 模型定义 (Model Definition)

## 概述

GORM 使用 Go 结构体定义数据库表结构，通过结构体标签（tags）配置字段属性。

## 基本模型

### 最简模型

```go
type User struct {
    ID   uint
    Name string
}

// GORM 会创建:
// CREATE TABLE users (
//     id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
//     name VARCHAR(255)
// )
```

### 使用 gorm.Model

```go
import "gorm.io/gorm"

// gorm.Model 定义
type Model struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 继承 gorm.Model
type User struct {
    gorm.Model          // ID, CreatedAt, UpdatedAt, DeletedAt
    Name  string
    Email string
}

// 等价于:
type User struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Name      string
    Email     string
}
```

## 字段标签

### 常用标签

```go
type User struct {
    // 主键
    ID uint `gorm:"primaryKey"`

    // 列名
    Name string `gorm:"column:user_name"`

    // 类型
    Content string `gorm:"type:text"`

    // 大小
    Name string `gorm:"size:100"`

    // 非空
    Email string `gorm:"not null"`

    // 默认值
    Status string `gorm:"default:'active'"`

    // 唯一
    Email string `gorm:"unique"`

    // 索引
    Email string `gorm:"index"`

    // 唯一索引
    Email string `gorm:"uniqueIndex"`

    // 忽略字段
    Password string `gorm:"-"`

    // 只读
    CreatedAt time.Time `gorm:"<-:create"`

    // 只写
    UpdatedBy uint `gorm:"->:false;<-"`

    // 自动创建时间
    CreatedAt time.Time `gorm:"autoCreateTime"`

    // 自动更新时间
    UpdatedAt time.Time `gorm:"autoUpdateTime"`

    // 自动更新时间（Unix 时间戳）
    UpdatedAt int64 `gorm:"autoUpdateTime:milli"` // 毫秒
    UpdatedAt int64 `gorm:"autoUpdateTime:nano"`  // 纳秒
}
```

### 完整示例

```go
type User struct {
    ID        uint           `gorm:"primaryKey;autoIncrement"`
    UUID      string         `gorm:"type:uuid;default:gen_random_uuid()"`
    Name      string         `gorm:"column:user_name;size:100;not null;index"`
    Email     string         `gorm:"size:255;uniqueIndex"`
    Password  string         `gorm:"size:255;not null"`
    Age       int            `gorm:"check:age >= 0"`
    Status    string         `gorm:"type:varchar(20);default:'active';index"`
    Role      string         `gorm:"type:varchar(20);default:'user'"`
    Bio       string         `gorm:"type:text"`
    Avatar    string         `gorm:"size:500"`
    Balance   float64        `gorm:"type:decimal(15,2);default:0"`
    IsAdmin   bool           `gorm:"default:false"`
    LastLogin *time.Time
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

## 主键

### 自增主键

```go
type User struct {
    ID uint `gorm:"primaryKey;autoIncrement"`
}
```

### UUID 主键

```go
import "github.com/google/uuid"

type User struct {
    ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
}

// 或手动设置
func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.ID = uuid.New()
    return nil
}
```

### 复合主键

```go
type UserRole struct {
    UserID uint `gorm:"primaryKey"`
    RoleID uint `gorm:"primaryKey"`
}
```

## 索引

### 单字段索引

```go
type User struct {
    Name  string `gorm:"index"`
    Email string `gorm:"uniqueIndex"`
}
```

### 命名索引

```go
type User struct {
    Name  string `gorm:"index:idx_name"`
    Email string `gorm:"uniqueIndex:idx_email"`
}
```

### 复合索引

```go
type User struct {
    // 复合索引
    FirstName string `gorm:"index:idx_name"`
    LastName  string `gorm:"index:idx_name"`

    // 复合唯一索引
    TenantID uint   `gorm:"uniqueIndex:idx_tenant_email"`
    Email    string `gorm:"uniqueIndex:idx_tenant_email"`
}
```

### 索引选项

```go
type User struct {
    // 索引排序
    Name string `gorm:"index:idx_name,sort:desc"`

    // 索引优先级
    Name  string `gorm:"index:idx_name,priority:1"`
    Email string `gorm:"index:idx_name,priority:2"`

    // 索引类型
    Location string `gorm:"index:idx_location,type:btree"`

    // 条件索引（PostgreSQL）
    Status string `gorm:"index:idx_active,where:status='active'"`
}
```

## 约束

### 外键约束

```go
type User struct {
    ID   uint
    Name string
}

type Post struct {
    ID       uint
    Title    string
    AuthorID uint `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    Author   User
}
```

### 检查约束

```go
type User struct {
    Age     int    `gorm:"check:age >= 0 AND age <= 150"`
    Balance float64 `gorm:"check:balance >= 0"`
}
```

## 字段类型

### 基本类型映射

```go
type Example struct {
    // 整数
    Int8    int8    // SMALLINT
    Int16   int16   // SMALLINT
    Int32   int32   // INTEGER
    Int64   int64   // BIGINT
    Uint    uint    // BIGINT UNSIGNED

    // 浮点
    Float32 float32 // REAL
    Float64 float64 // DOUBLE PRECISION

    // 字符串
    String string  // VARCHAR(255)

    // 布尔
    Bool bool      // BOOLEAN

    // 时间
    Time time.Time // TIMESTAMP

    // 字节
    Bytes []byte   // BYTEA
}
```

### 自定义类型

```go
// JSON 类型
type User struct {
    Settings JSON `gorm:"type:jsonb"`
}

type JSON map[string]interface{}

// 实现 Scanner 和 Valuer 接口
func (j *JSON) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("failed to unmarshal JSON")
    }
    return json.Unmarshal(bytes, j)
}

func (j JSON) Value() (driver.Value, error) {
    return json.Marshal(j)
}

// 字符串数组
type User struct {
    Tags StringArray `gorm:"type:text[]"`
}

type StringArray []string

func (a *StringArray) Scan(value interface{}) error {
    // PostgreSQL 数组解析
    return pq.Array(a).Scan(value)
}

func (a StringArray) Value() (driver.Value, error) {
    return pq.Array(a).Value()
}
```

### 枚举类型

```go
type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
    StatusBanned   Status = "banned"
)

type User struct {
    ID     uint
    Status Status `gorm:"type:varchar(20);default:'active'"`
}

// PostgreSQL 原生枚举
type User struct {
    Status Status `gorm:"type:user_status;default:'active'"`
}

// 需要先创建枚举类型:
// CREATE TYPE user_status AS ENUM ('active', 'inactive', 'banned');
```

## 嵌入结构体

### 普通嵌入

```go
type BaseModel struct {
    ID        uint `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type User struct {
    BaseModel
    Name  string
    Email string
}
// 生成: users 表包含 id, created_at, updated_at, name, email
```

### 嵌入带前缀

```go
type Address struct {
    Street string
    City   string
    State  string
}

type User struct {
    ID      uint
    Name    string
    Address Address `gorm:"embedded;embeddedPrefix:address_"`
}
// 生成: address_street, address_city, address_state
```

## 字段权限

```go
type User struct {
    // 可读可写（默认）
    Name string

    // 只创建时可写
    CreatedBy string `gorm:"<-:create"`

    // 只更新时可写
    UpdatedBy string `gorm:"<-:update"`

    // 可写（创建和更新）
    Content string `gorm:"<-"`

    // 只读
    ViewCount int `gorm:"->"`

    // 完全忽略
    Password string `gorm:"-"`

    // 忽略读写，但可迁移
    TempField string `gorm:"-:all"`

    // 忽略迁移
    CachedData string `gorm:"-:migration"`
}
```

## 表名约定

### 默认约定

```go
type User struct{}      // 表名: users
type UserInfo struct{}  // 表名: user_infos
```

### 自定义表名

```go
// 方式 1: 实现 Tabler 接口
func (User) TableName() string {
    return "app_users"
}

// 方式 2: 动态表名
func (u User) TableName() string {
    if u.TenantID > 0 {
        return fmt.Sprintf("tenant_%d_users", u.TenantID)
    }
    return "users"
}

// 方式 3: 查询时指定
db.Table("custom_users").Create(&user)
```

### 表名策略

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    NamingStrategy: schema.NamingStrategy{
        TablePrefix:   "t_",     // 表名前缀
        SingularTable: true,     // 使用单数表名 (user 而非 users)
        NoLowerCase:   false,    // 是否保持大小写
    },
})
```

## 完整模型示例

```go
package models

import (
    "time"

    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    UUID      string         `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex" json:"uuid"`
    Name      string         `gorm:"size:100;not null;index" json:"name"`
    Email     string         `gorm:"size:255;uniqueIndex" json:"email"`
    Password  string         `gorm:"size:255;not null" json:"-"`
    Avatar    string         `gorm:"size:500" json:"avatar"`
    Bio       string         `gorm:"type:text" json:"bio"`
    Role      string         `gorm:"size:20;default:'user';index" json:"role"`
    Status    string         `gorm:"size:20;default:'active';index" json:"status"`
    IsAdmin   bool           `gorm:"default:false" json:"is_admin"`
    LastLogin *time.Time     `json:"last_login"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    // 关联
    Posts    []Post    `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
    Comments []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}

func (User) TableName() string {
    return "users"
}
```

**下一节**：[CRUD 操作](./03-crud.md) - 学习基本增删改查
