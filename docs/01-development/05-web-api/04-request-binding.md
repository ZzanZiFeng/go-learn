# 请求绑定 (Request Binding)

## 概述

Gin 提供了强大的请求绑定功能，可以自动将请求数据绑定到 Go 结构体。

## 与 JavaScript 对比

### Express.js

```javascript
// 需要中间件解析请求体
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.post('/users', (req, res) => {
  const { name, email } = req.body;  // 手动解构
  // 没有类型检查
});
```

### Gin

```go
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 自动验证失败
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // req.Name, req.Email 已填充且已验证
})
```

## 绑定方法

### Must vs Should

```go
// Should* 方法 - 失败返回错误，不自动响应
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}

// Bind* 方法 - 失败自动响应 400 错误
if err := c.BindJSON(&req); err != nil {
    return  // 已自动响应
}
```

推荐使用 `Should*` 方法，更灵活的错误处理。

## JSON 绑定

### 基本用法

```go
type User struct {
    Name    string `json:"name"`
    Email   string `json:"email"`
    Age     int    `json:"age"`
    Active  bool   `json:"active"`
}

r.POST("/users", func(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, user)
})
```

### 嵌套结构

```go
type Address struct {
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    Name    string  `json:"name"`
    Address Address `json:"address"`
}

// 请求体:
// {
//   "name": "Alice",
//   "address": {
//     "city": "Beijing",
//     "country": "China"
//   }
// }
```

### 数组和切片

```go
type CreateOrderRequest struct {
    UserID uint        `json:"user_id"`
    Items  []OrderItem `json:"items"`
}

type OrderItem struct {
    ProductID uint `json:"product_id"`
    Quantity  int  `json:"quantity"`
}

// 请求体:
// {
//   "user_id": 1,
//   "items": [
//     {"product_id": 100, "quantity": 2},
//     {"product_id": 101, "quantity": 1}
//   ]
// }
```

## Query 绑定

```go
type SearchParams struct {
    Keyword string `form:"keyword"`
    Page    int    `form:"page,default=1"`
    Limit   int    `form:"limit,default=10"`
    Sort    string `form:"sort,default=created_at"`
    Order   string `form:"order,default=desc"`
}

// GET /search?keyword=go&page=2&limit=20
r.GET("/search", func(c *gin.Context) {
    var params SearchParams
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "keyword": params.Keyword,
        "page":    params.Page,
        "limit":   params.Limit,
    })
})
```

## URI 绑定

```go
type UserURI struct {
    ID uint `uri:"id" binding:"required"`
}

// GET /users/123
r.GET("/users/:id", func(c *gin.Context) {
    var uri UserURI
    if err := c.ShouldBindUri(&uri); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"id": uri.ID})
})
```

### 组合绑定

```go
type GetUserParams struct {
    ID uint `uri:"id" binding:"required"`
}

type GetUserQuery struct {
    Fields string `form:"fields"`  // ?fields=name,email
}

r.GET("/users/:id", func(c *gin.Context) {
    var params GetUserParams
    var query GetUserQuery

    if err := c.ShouldBindUri(&params); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "id":     params.ID,
        "fields": query.Fields,
    })
})
```

## Form 绑定

### URL Encoded Form

```go
type LoginForm struct {
    Username string `form:"username" binding:"required"`
    Password string `form:"password" binding:"required"`
}

// POST with Content-Type: application/x-www-form-urlencoded
r.POST("/login", func(c *gin.Context) {
    var form LoginForm
    if err := c.ShouldBind(&form); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"username": form.Username})
})
```

### Multipart Form

```go
type UploadForm struct {
    Name    string                `form:"name" binding:"required"`
    File    *multipart.FileHeader `form:"file" binding:"required"`
}

r.POST("/upload", func(c *gin.Context) {
    var form UploadForm
    if err := c.ShouldBind(&form); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 保存文件
    dst := "./uploads/" + form.File.Filename
    c.SaveUploadedFile(form.File, dst)

    c.JSON(200, gin.H{
        "name":     form.Name,
        "filename": form.File.Filename,
        "size":     form.File.Size,
    })
})
```

## Header 绑定

```go
type AuthHeader struct {
    Authorization string `header:"Authorization" binding:"required"`
    ContentType   string `header:"Content-Type"`
    UserAgent     string `header:"User-Agent"`
}

r.GET("/protected", func(c *gin.Context) {
    var headers AuthHeader
    if err := c.ShouldBindHeader(&headers); err != nil {
        c.JSON(401, gin.H{"error": "Authorization required"})
        return
    }

    c.JSON(200, gin.H{"auth": headers.Authorization})
})
```

## 自动绑定 (ShouldBind)

`ShouldBind` 根据 Content-Type 自动选择绑定方式：

```go
type CreateRequest struct {
    Name  string `json:"name" form:"name"`
    Email string `json:"email" form:"email"`
}

r.POST("/create", func(c *gin.Context) {
    var req CreateRequest
    // 自动根据 Content-Type 选择绑定方式:
    // - application/json -> JSON
    // - application/x-www-form-urlencoded -> Form
    // - multipart/form-data -> Form
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, req)
})
```

## 自定义绑定

### 绑定 XML

```go
type XMLRequest struct {
    Name string `xml:"name"`
}

r.POST("/xml", func(c *gin.Context) {
    var req XMLRequest
    if err := c.ShouldBindXML(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, req)
})
```

### 绑定 YAML

```go
type YAMLRequest struct {
    Name string `yaml:"name"`
}

r.POST("/yaml", func(c *gin.Context) {
    var req YAMLRequest
    if err := c.ShouldBindYAML(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, req)
})
```

## 标签汇总

```go
type Example struct {
    // JSON 绑定
    Field1 string `json:"field1"`

    // Form 绑定（query 和 form 数据）
    Field2 string `form:"field2"`

    // URI 绑定
    Field3 string `uri:"field3"`

    // Header 绑定
    Field4 string `header:"X-Field4"`

    // XML 绑定
    Field5 string `xml:"field5"`

    // 绑定验证
    Field6 string `binding:"required"`

    // 多标签组合
    Field7 string `json:"field7" form:"field7" binding:"required,min=3"`
}
```

## 时间处理

```go
type Event struct {
    Name      string    `json:"name"`
    StartTime time.Time `json:"start_time" time_format:"2006-01-02 15:04:05"`
    EndTime   time.Time `json:"end_time" time_format:"2006-01-02"`
}

// 请求体:
// {
//   "name": "Meeting",
//   "start_time": "2024-01-15 14:30:00",
//   "end_time": "2024-01-15"
// }
```

## 完整示例

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// 请求结构体
type CreateUserRequest struct {
    Name     string   `json:"name" binding:"required,min=2,max=100"`
    Email    string   `json:"email" binding:"required,email"`
    Age      int      `json:"age" binding:"omitempty,gte=0,lte=150"`
    Tags     []string `json:"tags" binding:"omitempty,dive,min=1"`
    Address  Address  `json:"address" binding:"omitempty"`
}

type Address struct {
    City    string `json:"city" binding:"required"`
    Country string `json:"country" binding:"required"`
}

type GetUserParams struct {
    ID uint `uri:"id" binding:"required,gt=0"`
}

type ListUsersQuery struct {
    Page    int    `form:"page,default=1" binding:"gte=1"`
    Limit   int    `form:"limit,default=10" binding:"gte=1,lte=100"`
    Sort    string `form:"sort,default=created_at"`
    Order   string `form:"order,default=desc" binding:"oneof=asc desc"`
    Keyword string `form:"keyword"`
}

func main() {
    r := gin.Default()

    // 创建用户 - JSON Body
    r.POST("/users", func(c *gin.Context) {
        var req CreateUserRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Validation failed",
                "details": err.Error(),
            })
            return
        }
        c.JSON(http.StatusCreated, gin.H{"data": req})
    })

    // 获取用户 - URI + Query
    r.GET("/users/:id", func(c *gin.Context) {
        var params GetUserParams
        if err := c.ShouldBindUri(&params); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"id": params.ID})
    })

    // 列出用户 - Query
    r.GET("/users", func(c *gin.Context) {
        var query ListUsersQuery
        if err := c.ShouldBindQuery(&query); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "page":    query.Page,
            "limit":   query.Limit,
            "sort":    query.Sort,
            "order":   query.Order,
            "keyword": query.Keyword,
        })
    })

    r.Run(":8080")
}
```

## 总结

| 方法 | 数据来源 | 标签 |
|------|----------|------|
| `ShouldBindJSON` | Request Body (JSON) | `json` |
| `ShouldBindQuery` | URL Query | `form` |
| `ShouldBindUri` | URL Path | `uri` |
| `ShouldBind` | Body/Query (auto) | `json`/`form` |
| `ShouldBindHeader` | Request Headers | `header` |
| `ShouldBindXML` | Request Body (XML) | `xml` |

**下一节**：[响应处理](./05-response.md) - 学习响应格式化
