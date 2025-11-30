# 请求验证 (Request Validation)

## 概述

Gin 使用 [go-playground/validator](https://github.com/go-playground/validator) 进行数据验证，通过 `binding` 标签定义验证规则。

## 与 JavaScript 对比

### Express + Joi/Zod

```javascript
// 使用 Zod
import { z } from 'zod';

const userSchema = z.object({
  name: z.string().min(2).max(100),
  email: z.string().email(),
  age: z.number().int().min(0).max(150).optional(),
});

app.post('/users', (req, res) => {
  const result = userSchema.safeParse(req.body);
  if (!result.success) {
    return res.status(400).json({ errors: result.error.errors });
  }
  // 使用 result.data
});
```

### Gin Validator

```go
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required,min=2,max=100"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"omitempty,gte=0,lte=150"`
}

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // 使用 req
})
```

## 基础验证标签

### 必填验证

```go
type Request struct {
    // 必填字段
    Name string `binding:"required"`

    // 非零值（对数字/布尔有意义）
    Count int `binding:"required"`

    // 指针可以区分 null 和缺失
    Age *int `binding:"required"`
}
```

### 字符串验证

```go
type StringValidation struct {
    // 长度验证
    Name     string `binding:"min=2,max=100"`       // 长度 2-100
    Code     string `binding:"len=6"`               // 精确长度 6
    Password string `binding:"min=8"`               // 最少 8 字符

    // 格式验证
    Email    string `binding:"email"`               // 邮箱格式
    URL      string `binding:"url"`                 // URL 格式
    UUID     string `binding:"uuid"`                // UUID 格式
    IP       string `binding:"ip"`                  // IP 地址
    IPv4     string `binding:"ipv4"`                // IPv4 地址
    IPv6     string `binding:"ipv6"`                // IPv6 地址

    // 内容验证
    Alpha    string `binding:"alpha"`               // 只含字母
    Alphanum string `binding:"alphanum"`            // 字母数字
    Numeric  string `binding:"numeric"`             // 数字字符串
    ASCII    string `binding:"ascii"`               // ASCII 字符
    Lowercase string `binding:"lowercase"`          // 小写
    Uppercase string `binding:"uppercase"`          // 大写

    // 正则验证（需要自定义）
    Phone    string `binding:"required"`            // 需自定义
}
```

### 数字验证

```go
type NumberValidation struct {
    // 范围验证
    Age      int     `binding:"gte=0,lte=150"`      // 0 <= x <= 150
    Price    float64 `binding:"gt=0"`               // x > 0
    Discount float64 `binding:"gte=0,lt=1"`         // 0 <= x < 1
    Quantity int     `binding:"min=1,max=100"`      // 1 <= x <= 100

    // 比较运算符
    // gt: >   gte: >=   lt: <   lte: <=
    // eq: ==  ne: !=
}
```

### 枚举验证

```go
type EnumValidation struct {
    // oneof: 必须是列表中的一个
    Status string `binding:"oneof=active inactive pending"`
    Role   string `binding:"oneof=admin user guest"`
    Type   int    `binding:"oneof=1 2 3"`

    // 排除某些值
    Name string `binding:"nefield=admin"`  // 不能等于 "admin"
}
```

## 条件验证

### 可选字段

```go
type OptionalFields struct {
    // omitempty: 为空时跳过验证
    Bio      string `binding:"omitempty,max=500"`
    Website  string `binding:"omitempty,url"`
    Age      int    `binding:"omitempty,gte=0,lte=150"`

    // 使用指针表示可选
    Phone    *string `binding:"omitempty,len=11"`
}
```

### 字段依赖验证

```go
type FieldDependency struct {
    // required_with: 当其他字段存在时必填
    Password        string `binding:"required_with=PasswordConfirm"`
    PasswordConfirm string `binding:"required_with=Password,eqfield=Password"`

    // required_without: 当其他字段不存在时必填
    Email string `binding:"required_without=Phone"`
    Phone string `binding:"required_without=Email"`

    // required_if: 条件必填
    PaymentMethod string `binding:"required"`
    CardNumber    string `binding:"required_if=PaymentMethod card"`

    // required_unless: 除非条件满足，否则必填
    Reason string `binding:"required_unless=Status approved"`
}
```

### 字段比较

```go
type FieldComparison struct {
    // 字段相等
    Password        string `binding:"required,min=8"`
    PasswordConfirm string `binding:"required,eqfield=Password"`

    // 字段不等
    NewPassword string `binding:"required,nefield=OldPassword"`
    OldPassword string `binding:"required"`

    // 数字比较
    StartDate time.Time `binding:"required"`
    EndDate   time.Time `binding:"required,gtfield=StartDate"`

    MinPrice float64 `binding:"required,gte=0"`
    MaxPrice float64 `binding:"required,gtefield=MinPrice"`
}
```

## 数组和嵌套验证

### 数组验证

```go
type ArrayValidation struct {
    // 数组长度
    Tags []string `binding:"min=1,max=10"`  // 1-10 个元素

    // dive: 验证数组元素
    Emails []string `binding:"required,dive,email"`
    Scores []int    `binding:"required,dive,gte=0,lte=100"`

    // 嵌套对象数组
    Items []OrderItem `binding:"required,dive"`
}

type OrderItem struct {
    ProductID uint `binding:"required"`
    Quantity  int  `binding:"required,gte=1"`
}
```

### 嵌套结构验证

```go
type NestedValidation struct {
    Name    string  `binding:"required"`
    Address Address `binding:"required"`  // 会验证嵌套结构
}

type Address struct {
    Street  string `binding:"required"`
    City    string `binding:"required"`
    Country string `binding:"required,len=2"`  // ISO 国家代码
    ZipCode string `binding:"required,numeric,len=6"`
}
```

### Map 验证

```go
type MapValidation struct {
    // Map 验证
    Metadata map[string]string `binding:"dive,keys,min=1,max=50,endkeys,required,max=200"`
    // keys: 开始验证 key
    // endkeys: 结束验证 key，开始验证 value

    Scores map[string]int `binding:"dive,keys,alpha,endkeys,gte=0,lte=100"`
}
```

## 自定义验证器

### 定义自定义验证器

```go
package main

import (
    "regexp"

    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
)

// 手机号验证器
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func validatePhone(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    return phoneRegex.MatchString(phone)
}

// 用户名验证器（字母开头，允许字母数字下划线）
var usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{2,19}$`)

func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    return usernameRegex.MatchString(username)
}

// 密码强度验证器
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    if len(password) < 8 {
        return false
    }

    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)

    return hasUpper && hasLower && hasNumber && hasSpecial
}

func main() {
    r := gin.Default()

    // 注册自定义验证器
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("phone", validatePhone)
        v.RegisterValidation("username", validateUsername)
        v.RegisterValidation("strongpwd", validateStrongPassword)
    }

    r.POST("/register", registerHandler)
    r.Run(":8080")
}

// 使用自定义验证器
type RegisterRequest struct {
    Username string `json:"username" binding:"required,username"`
    Password string `json:"password" binding:"required,strongpwd"`
    Phone    string `json:"phone" binding:"required,phone"`
    Email    string `json:"email" binding:"required,email"`
}

func registerHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"message": "registered"})
}
```

### 带参数的自定义验证器

```go
// 验证字符串是否在数据库中存在
func validateExists(fl validator.FieldLevel) bool {
    // 获取参数（表名）
    param := fl.Param()  // 例如: "users"

    value := fl.Field().String()

    // 检查数据库（伪代码）
    // exists := db.Table(param).Where("id = ?", value).Exists()
    // return exists

    return true
}

// 注册
v.RegisterValidation("exists", validateExists)

// 使用
type Request struct {
    UserID string `binding:"required,exists=users"`
    RoleID string `binding:"required,exists=roles"`
}
```

### 结构体级别验证

```go
// 结构体验证器
func validateDateRange(sl validator.StructLevel) {
    req := sl.Current().Interface().(DateRangeRequest)

    if !req.EndDate.After(req.StartDate) {
        sl.ReportError(req.EndDate, "EndDate", "end_date", "daterange", "")
    }
}

type DateRangeRequest struct {
    StartDate time.Time `json:"start_date" binding:"required"`
    EndDate   time.Time `json:"end_date" binding:"required"`
}

func main() {
    r := gin.Default()

    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterStructValidation(validateDateRange, DateRangeRequest{})
    }

    r.POST("/report", func(c *gin.Context) {
        var req DateRangeRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"data": req})
    })

    r.Run(":8080")
}
```

## 错误处理

### 解析验证错误

```go
import (
    "github.com/go-playground/validator/v10"
)

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func formatValidationErrors(err error) []ValidationError {
    var errors []ValidationError

    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            errors = append(errors, ValidationError{
                Field:   toSnakeCase(e.Field()),
                Message: getErrorMessage(e),
            })
        }
    }

    return errors
}

func getErrorMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return "This field is required"
    case "email":
        return "Invalid email format"
    case "min":
        return fmt.Sprintf("Minimum length is %s", e.Param())
    case "max":
        return fmt.Sprintf("Maximum length is %s", e.Param())
    case "gte":
        return fmt.Sprintf("Must be greater than or equal to %s", e.Param())
    case "lte":
        return fmt.Sprintf("Must be less than or equal to %s", e.Param())
    case "oneof":
        return fmt.Sprintf("Must be one of: %s", e.Param())
    case "eqfield":
        return fmt.Sprintf("Must be equal to %s", e.Param())
    case "phone":
        return "Invalid phone number"
    case "username":
        return "Username must start with a letter, 3-20 characters"
    case "strongpwd":
        return "Password must contain uppercase, lowercase, number and special character"
    default:
        return fmt.Sprintf("Failed on '%s' validation", e.Tag())
    }
}

func toSnakeCase(str string) string {
    var result []rune
    for i, r := range str {
        if i > 0 && r >= 'A' && r <= 'Z' {
            result = append(result, '_')
        }
        result = append(result, r)
    }
    return strings.ToLower(string(result))
}
```

### 使用格式化错误

```go
r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        errors := formatValidationErrors(err)
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    "VALIDATION_ERROR",
            "message": "Validation failed",
            "errors":  errors,
        })
        return
    }
    // ...
})
```

响应示例：

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "errors": [
    {"field": "email", "message": "Invalid email format"},
    {"field": "password", "message": "Minimum length is 8"}
  ]
}
```

## 国际化错误消息

```go
import (
    "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
    uni      *ut.UniversalTranslator
    validate *validator.Validate
    trans    ut.Translator
)

func initValidator() {
    zh := zh.New()
    uni = ut.New(zh, zh)

    trans, _ = uni.GetTranslator("zh")

    validate = validator.New()
    zh_translations.RegisterDefaultTranslations(validate, trans)

    // 注册自定义翻译
    validate.RegisterTranslation("phone", trans, func(ut ut.Translator) error {
        return ut.Add("phone", "{0}必须是有效的手机号码", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("phone", fe.Field())
        return t
    })
}

func translateError(err error) []string {
    var messages []string
    if errs, ok := err.(validator.ValidationErrors); ok {
        for _, e := range errs {
            messages = append(messages, e.Translate(trans))
        }
    }
    return messages
}
```

## 完整示例

```go
package main

import (
    "fmt"
    "net/http"
    "regexp"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
)

// 验证错误响应
type ValidationErrorResponse struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Errors  []ValidationError `json:"errors"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// 请求结构体
type CreateUserRequest struct {
    Username string   `json:"username" binding:"required,username"`
    Email    string   `json:"email" binding:"required,email"`
    Password string   `json:"password" binding:"required,strongpwd"`
    Phone    string   `json:"phone" binding:"omitempty,phone"`
    Age      int      `json:"age" binding:"omitempty,gte=0,lte=150"`
    Role     string   `json:"role" binding:"required,oneof=admin user guest"`
    Tags     []string `json:"tags" binding:"omitempty,max=5,dive,min=1,max=20"`
    Profile  *Profile `json:"profile" binding:"omitempty"`
}

type Profile struct {
    Bio     string `json:"bio" binding:"omitempty,max=500"`
    Website string `json:"website" binding:"omitempty,url"`
}

type UpdatePasswordRequest struct {
    OldPassword     string `json:"old_password" binding:"required"`
    NewPassword     string `json:"new_password" binding:"required,strongpwd,nefield=OldPassword"`
    ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type DateRangeRequest struct {
    StartDate time.Time `json:"start_date" binding:"required" time_format:"2006-01-02"`
    EndDate   time.Time `json:"end_date" binding:"required" time_format:"2006-01-02"`
}

// 自定义验证器
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
var usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{2,19}$`)

func validatePhone(fl validator.FieldLevel) bool {
    return phoneRegex.MatchString(fl.Field().String())
}

func validateUsername(fl validator.FieldLevel) bool {
    return usernameRegex.MatchString(fl.Field().String())
}

func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    if len(password) < 8 {
        return false
    }
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)
    return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateDateRange(sl validator.StructLevel) {
    req := sl.Current().Interface().(DateRangeRequest)
    if !req.EndDate.After(req.StartDate) {
        sl.ReportError(req.EndDate, "EndDate", "end_date", "daterange", "")
    }
}

// 错误格式化
func formatValidationErrors(err error) []ValidationError {
    var errors []ValidationError
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            errors = append(errors, ValidationError{
                Field:   toSnakeCase(e.Field()),
                Message: getErrorMessage(e),
            })
        }
    }
    return errors
}

func getErrorMessage(e validator.FieldError) string {
    messages := map[string]string{
        "required":  "This field is required",
        "email":     "Invalid email format",
        "url":       "Invalid URL format",
        "phone":     "Invalid phone number format",
        "username":  "Username must start with letter, 3-20 chars, alphanumeric and underscore only",
        "strongpwd": "Password must have uppercase, lowercase, number and special character",
        "eqfield":   "Must match " + e.Param(),
        "nefield":   "Must be different from " + e.Param(),
        "daterange": "End date must be after start date",
    }

    if msg, ok := messages[e.Tag()]; ok {
        return msg
    }

    switch e.Tag() {
    case "min":
        return fmt.Sprintf("Minimum is %s", e.Param())
    case "max":
        return fmt.Sprintf("Maximum is %s", e.Param())
    case "gte":
        return fmt.Sprintf("Must be >= %s", e.Param())
    case "lte":
        return fmt.Sprintf("Must be <= %s", e.Param())
    case "oneof":
        return fmt.Sprintf("Must be one of: %s", e.Param())
    default:
        return fmt.Sprintf("Failed '%s' validation", e.Tag())
    }
}

func toSnakeCase(str string) string {
    var result []rune
    for i, r := range str {
        if i > 0 && r >= 'A' && r <= 'Z' {
            result = append(result, '_')
        }
        result = append(result, r)
    }
    return strings.ToLower(string(result))
}

// 响应辅助函数
func validationError(c *gin.Context, err error) {
    c.JSON(http.StatusBadRequest, ValidationErrorResponse{
        Code:    "VALIDATION_ERROR",
        Message: "Validation failed",
        Errors:  formatValidationErrors(err),
    })
}

func main() {
    r := gin.Default()

    // 注册验证器
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("phone", validatePhone)
        v.RegisterValidation("username", validateUsername)
        v.RegisterValidation("strongpwd", validateStrongPassword)
        v.RegisterStructValidation(validateDateRange, DateRangeRequest{})
    }

    // 创建用户
    r.POST("/users", func(c *gin.Context) {
        var req CreateUserRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            validationError(c, err)
            return
        }
        c.JSON(http.StatusCreated, gin.H{"data": req})
    })

    // 更新密码
    r.PUT("/users/:id/password", func(c *gin.Context) {
        var req UpdatePasswordRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            validationError(c, err)
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
    })

    // 日期范围查询
    r.GET("/reports", func(c *gin.Context) {
        var req DateRangeRequest
        if err := c.ShouldBindQuery(&req); err != nil {
            validationError(c, err)
            return
        }
        c.JSON(http.StatusOK, gin.H{"data": req})
    })

    r.Run(":8080")
}
```

## 常用验证标签速查

| 标签 | 描述 | 示例 |
|------|------|------|
| `required` | 必填 | `binding:"required"` |
| `omitempty` | 为空时跳过 | `binding:"omitempty,email"` |
| `min` | 最小值/长度 | `binding:"min=3"` |
| `max` | 最大值/长度 | `binding:"max=100"` |
| `len` | 精确长度 | `binding:"len=6"` |
| `eq` | 等于 | `binding:"eq=10"` |
| `ne` | 不等于 | `binding:"ne=0"` |
| `gt/gte` | 大于/大于等于 | `binding:"gte=0"` |
| `lt/lte` | 小于/小于等于 | `binding:"lte=100"` |
| `oneof` | 枚举值 | `binding:"oneof=a b c"` |
| `email` | 邮箱格式 | `binding:"email"` |
| `url` | URL 格式 | `binding:"url"` |
| `uuid` | UUID 格式 | `binding:"uuid"` |
| `eqfield` | 等于其他字段 | `binding:"eqfield=Password"` |
| `nefield` | 不等于其他字段 | `binding:"nefield=OldPassword"` |
| `dive` | 验证数组元素 | `binding:"dive,email"` |

**下一节**：[中间件](./07-middleware.md) - 学习 Gin 中间件
