# 结构体标签

本文档介绍 Go 结构体标签的使用，特别是 JSON 和数据库相关的标签。

## 目录

- [标签基础](#标签基础)
- [JSON 标签](#json-标签)
- [数据库标签](#数据库标签)
- [验证标签](#验证标签)
- [自定义标签](#自定义标签)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 标签基础

### 什么是结构体标签？

结构体标签是附加在字段后面的元数据字符串，用于运行时反射。

```go
type User struct {
    Name string `json:"name" db:"user_name"`
    Age  int    `json:"age" db:"user_age"`
}
```

### 标签语法

```go
type Example struct {
    Field Type `key1:"value1" key2:"value2,option1,option2"`
}
```

- 使用反引号 `` ` `` 包围
- 键值对格式：`key:"value"`
- 多个标签用空格分隔
- 值可以包含逗号分隔的选项

### 读取标签

```go
package main

import (
    "fmt"
    "reflect"
)

type User struct {
    Name string `json:"name" validate:"required"`
    Age  int    `json:"age" validate:"min=0,max=150"`
}

func main() {
    t := reflect.TypeOf(User{})

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fmt.Printf("Field: %s\n", field.Name)
        fmt.Printf("  json tag: %s\n", field.Tag.Get("json"))
        fmt.Printf("  validate tag: %s\n", field.Tag.Get("validate"))
    }
}
```

---

## JSON 标签

### 基本用法

```go
package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    ID        int    `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
}

func main() {
    user := User{
        ID:        1,
        FirstName: "John",
        LastName:  "Doe",
        Email:     "john@example.com",
    }

    // 序列化
    data, _ := json.Marshal(user)
    fmt.Println(string(data))
    // {"id":1,"first_name":"John","last_name":"Doe","email":"john@example.com"}

    // 反序列化
    jsonStr := `{"id":2,"first_name":"Jane","last_name":"Smith"}`
    var user2 User
    json.Unmarshal([]byte(jsonStr), &user2)
    fmt.Printf("%+v\n", user2)
}
```

### 常用 JSON 标签选项

```go
type Example struct {
    // 重命名字段
    Name string `json:"name"`

    // 忽略字段
    Password string `json:"-"`

    // 忽略空值
    Email string `json:"email,omitempty"`

    // 字符串类型数字
    Age int `json:"age,string"`

    // 内联（展平嵌套结构）
    Address Address `json:",inline"`
}
```

### omitempty 详解

```go
type Response struct {
    Code    int     `json:"code"`
    Message string  `json:"message,omitempty"`
    Data    *Result `json:"data,omitempty"`
    Error   string  `json:"error,omitempty"`
}

func main() {
    // 成功响应
    success := Response{Code: 200, Data: &Result{}}
    data, _ := json.Marshal(success)
    fmt.Println(string(data))
    // {"code":200,"data":{}}

    // 错误响应
    failure := Response{Code: 500, Error: "Internal error"}
    data, _ = json.Marshal(failure)
    fmt.Println(string(data))
    // {"code":500,"error":"Internal error"}
}
```

**omitempty 判断规则**:
- 数值：0
- 字符串：""
- 布尔：false
- 指针/接口/切片/映射/通道：nil

### 嵌套结构体

```go
type Address struct {
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    Name    string  `json:"name"`
    Address Address `json:"address"`
}

// JSON 输出:
// {"name":"Alice","address":{"city":"Beijing","country":"China"}}
```

### 匿名字段（内嵌）

```go
type Metadata struct {
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
    Name     string `json:"name"`
    Metadata        // 匿名字段
}

// JSON 输出（字段被展平）:
// {"name":"Alice","created_at":"...","updated_at":"..."}
```

---

## 数据库标签

### GORM 标签

```go
import "gorm.io/gorm"

type User struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string `gorm:"column:user_name;size:100;not null"`
    Email     string `gorm:"uniqueIndex;size:255"`
    Age       int    `gorm:"default:18"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### 常用 GORM 标签

| 标签 | 说明 |
|-----|------|
| `primaryKey` | 主键 |
| `column:name` | 指定列名 |
| `size:n` | 字段大小 |
| `not null` | 非空约束 |
| `unique` | 唯一约束 |
| `uniqueIndex` | 唯一索引 |
| `index` | 普通索引 |
| `default:value` | 默认值 |
| `type:typename` | 指定数据库类型 |
| `foreignKey` | 外键 |
| `-` | 忽略字段 |

### sqlx 标签

```go
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

// 使用 sqlx
var user User
err := db.Get(&user, "SELECT * FROM users WHERE id = ?", 1)
```

---

## 验证标签

### go-playground/validator

```go
import "github.com/go-playground/validator/v10"

type User struct {
    Name     string `validate:"required,min=2,max=50"`
    Email    string `validate:"required,email"`
    Age      int    `validate:"required,min=0,max=150"`
    Password string `validate:"required,min=8"`
    Phone    string `validate:"omitempty,e164"`
}

func main() {
    validate := validator.New()

    user := User{
        Name:     "A",           // 太短
        Email:    "invalid",     // 无效邮箱
        Age:      200,           // 超出范围
        Password: "123",         // 太短
    }

    err := validate.Struct(user)
    if err != nil {
        for _, e := range err.(validator.ValidationErrors) {
            fmt.Printf("Field: %s, Tag: %s, Value: %v\n",
                e.Field(), e.Tag(), e.Value())
        }
    }
}
```

### 常用验证标签

| 标签 | 说明 |
|-----|------|
| `required` | 必填 |
| `email` | 邮箱格式 |
| `min=n` | 最小值/长度 |
| `max=n` | 最大值/长度 |
| `len=n` | 固定长度 |
| `oneof=a b c` | 枚举值 |
| `url` | URL 格式 |
| `uuid` | UUID 格式 |
| `omitempty` | 空值跳过验证 |

### Gin 绑定验证

```go
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

func loginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // 处理登录...
}
```

---

## 自定义标签

### 读取自定义标签

```go
type Config struct {
    Host    string `env:"APP_HOST" default:"localhost"`
    Port    int    `env:"APP_PORT" default:"8080"`
    Debug   bool   `env:"APP_DEBUG" default:"false"`
}

func loadConfig(cfg interface{}) error {
    v := reflect.ValueOf(cfg).Elem()
    t := v.Type()

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        envKey := field.Tag.Get("env")
        defaultVal := field.Tag.Get("default")

        value := os.Getenv(envKey)
        if value == "" {
            value = defaultVal
        }

        // 设置字段值...
        fmt.Printf("Field %s: env=%s, default=%s, value=%s\n",
            field.Name, envKey, defaultVal, value)
    }
    return nil
}
```

### 组合多个标签

```go
type User struct {
    ID       int    `json:"id" db:"id" gorm:"primaryKey"`
    Name     string `json:"name" db:"name" validate:"required,min=2"`
    Email    string `json:"email" db:"email" validate:"required,email" gorm:"uniqueIndex"`
    Password string `json:"-" db:"password" validate:"required,min=8"`
}
```

---

## 与 JavaScript/TypeScript 对比

### 装饰器 vs 标签

```typescript
// TypeScript - 使用装饰器
import { IsEmail, MinLength, IsNotEmpty } from 'class-validator';

class User {
    @IsNotEmpty()
    name: string;

    @IsEmail()
    email: string;

    @MinLength(8)
    password: string;
}
```

```go
// Go - 使用标签
type User struct {
    Name     string `validate:"required"`
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}
```

### JSON 序列化对比

```typescript
// TypeScript
class User {
    @JsonProperty('first_name')
    firstName: string;

    @JsonIgnore()
    password: string;
}
```

```go
// Go
type User struct {
    FirstName string `json:"first_name"`
    Password  string `json:"-"`
}
```

### 主要差异

| 特性 | TypeScript | Go |
|-----|-----------|-----|
| 语法 | 装饰器 `@` | 标签 `` ` `` |
| 位置 | 字段/方法上方 | 字段后面 |
| 运行时访问 | 需要 reflect-metadata | 内置 reflect |
| 类型检查 | 编译时 | 运行时 |

---

## 最佳实践

### 1. 始终使用 json 标签

```go
// 好 - 明确指定 JSON 键名
type User struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

// 避免 - 依赖默认行为
type User struct {
    FirstName string
    LastName  string
}
```

### 2. 敏感字段使用 `-`

```go
type User struct {
    Password    string `json:"-"`
    CreditCard  string `json:"-"`
    AccessToken string `json:"-"`
}
```

### 3. 合理使用 omitempty

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### 4. 标签对齐

```go
// 好 - 对齐易读
type User struct {
    ID        int       `json:"id"         db:"id"         gorm:"primaryKey"`
    Name      string    `json:"name"       db:"name"       validate:"required"`
    Email     string    `json:"email"      db:"email"      validate:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

---

## 下一步

- [方法](./03-methods.md) - 学习为结构体定义方法
