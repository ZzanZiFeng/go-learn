# Gin 入门 (Gin Introduction)

## 什么是 Gin？

Gin 是一个高性能的 Go Web 框架，提供了：
- 高性能的 HTTP 路由
- 中间件支持
- JSON 验证
- 错误处理
- 渲染（JSON、XML、HTML）

## 安装

```bash
go get -u github.com/gin-gonic/gin
```

## 与 Express.js 对比

### Express.js

```javascript
const express = require('express');
const app = express();

app.use(express.json());

app.get('/users/:id', (req, res) => {
  const id = req.params.id;
  res.json({ id, name: 'Alice' });
});

app.post('/users', (req, res) => {
  const { name, email } = req.body;
  res.status(201).json({ id: 1, name, email });
});

app.listen(8080);
```

### Gin

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.GET("/users/:id", func(c *gin.Context) {
        id := c.Param("id")
        c.JSON(http.StatusOK, gin.H{"id": id, "name": "Alice"})
    })

    r.POST("/users", func(c *gin.Context) {
        var user struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }
        c.ShouldBindJSON(&user)
        c.JSON(http.StatusCreated, gin.H{"id": 1, "name": user.Name, "email": user.Email})
    })

    r.Run(":8080")
}
```

## 基础用法

### Hello World

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    // 创建默认引擎（带 Logger 和 Recovery 中间件）
    r := gin.Default()

    // 或创建无中间件引擎
    // r := gin.New()

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    r.Run() // 默认 :8080
}
```

### gin.H

`gin.H` 是 `map[string]interface{}` 的快捷方式：

```go
// 等价于
gin.H{"key": "value"}
map[string]interface{}{"key": "value"}
```

## HTTP 方法

```go
func main() {
    r := gin.Default()

    r.GET("/users", getUsers)
    r.GET("/users/:id", getUser)
    r.POST("/users", createUser)
    r.PUT("/users/:id", updateUser)
    r.PATCH("/users/:id", patchUser)
    r.DELETE("/users/:id", deleteUser)

    // 匹配所有方法
    r.Any("/any", anyHandler)

    // 处理 404
    r.NoRoute(func(c *gin.Context) {
        c.JSON(404, gin.H{"error": "Page not found"})
    })

    r.Run(":8080")
}
```

## 请求参数

### 路径参数

```go
// /users/123
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")  // "123"
    c.JSON(200, gin.H{"id": id})
})

// /users/123/orders/456
r.GET("/users/:userId/orders/:orderId", func(c *gin.Context) {
    userId := c.Param("userId")
    orderId := c.Param("orderId")
    c.JSON(200, gin.H{"userId": userId, "orderId": orderId})
})

// 通配符
// /files/path/to/file.txt -> path = "path/to/file.txt"
r.GET("/files/*path", func(c *gin.Context) {
    path := c.Param("path")  // "/path/to/file.txt"
    c.String(200, "Path: %s", path)
})
```

### 查询参数

```go
// /users?page=1&limit=10&active=true
r.GET("/users", func(c *gin.Context) {
    // 获取单个参数
    page := c.Query("page")           // "1"
    limit := c.DefaultQuery("limit", "10")  // 默认值
    active := c.Query("active")       // "true"

    // 检查参数是否存在
    name, exists := c.GetQuery("name")
    if !exists {
        // name 参数不存在
    }

    // 获取数组参数 ?ids=1&ids=2&ids=3
    ids := c.QueryArray("ids")  // ["1", "2", "3"]

    // 获取 map 参数 ?filters[status]=active&filters[type]=admin
    filters := c.QueryMap("filters")  // map[string]string

    c.JSON(200, gin.H{
        "page":   page,
        "limit":  limit,
        "active": active,
    })
})
```

### 请求体

```go
// JSON 请求体
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest

    // 绑定 JSON（失败返回错误）
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 或使用 BindJSON（失败自动返回 400）
    // if err := c.BindJSON(&req); err != nil {
    //     return  // 已自动响应错误
    // }

    c.JSON(201, gin.H{
        "id":    1,
        "name":  req.Name,
        "email": req.Email,
    })
})
```

### 表单数据

```go
// application/x-www-form-urlencoded
r.POST("/login", func(c *gin.Context) {
    username := c.PostForm("username")
    password := c.DefaultPostForm("password", "")

    c.JSON(200, gin.H{
        "username": username,
    })
})

// multipart/form-data
type UploadForm struct {
    Name  string                `form:"name"`
    Email string                `form:"email"`
    File  *multipart.FileHeader `form:"file"`
}

r.POST("/upload", func(c *gin.Context) {
    var form UploadForm
    if err := c.ShouldBind(&form); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 保存文件
    c.SaveUploadedFile(form.File, "./uploads/"+form.File.Filename)

    c.JSON(200, gin.H{"message": "uploaded"})
})
```

## 响应

### JSON 响应

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

r.GET("/user", func(c *gin.Context) {
    user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}

    // JSON 响应
    c.JSON(200, user)

    // 使用 gin.H
    c.JSON(200, gin.H{"user": user})

    // 安全 JSON（防止 JSON 劫持）
    c.SecureJSON(200, user)

    // 漂亮 JSON（格式化）
    c.IndentedJSON(200, user)

    // JSONP
    c.JSONP(200, user)
})
```

### 其他响应类型

```go
// 字符串
c.String(200, "Hello %s", name)

// XML
c.XML(200, user)

// YAML
c.YAML(200, user)

// HTML
c.HTML(200, "index.html", gin.H{"title": "Hello"})

// 文件
c.File("./files/report.pdf")

// 文件下载
c.FileAttachment("./files/report.pdf", "report.pdf")

// 重定向
c.Redirect(302, "/new-location")

// 读取器
c.DataFromReader(200, contentLength, contentType, reader, extraHeaders)

// 二进制数据
c.Data(200, "application/octet-stream", []byte{...})
```

### 响应头

```go
r.GET("/headers", func(c *gin.Context) {
    // 设置响应头
    c.Header("X-Custom-Header", "value")
    c.Header("Content-Type", "application/json")

    // 设置 Cookie
    c.SetCookie("token", "abc123", 3600, "/", "localhost", false, true)

    c.JSON(200, gin.H{"message": "ok"})
})
```

## Context

Gin 的 `*gin.Context` 是核心对象，包含请求和响应的所有信息：

```go
r.GET("/context", func(c *gin.Context) {
    // 请求
    c.Request        // *http.Request
    c.Param("id")    // 路径参数
    c.Query("page")  // 查询参数
    c.PostForm("x")  // 表单数据

    // 响应
    c.Writer         // http.ResponseWriter
    c.JSON(200, obj) // JSON 响应
    c.Status(200)    // 设置状态码

    // 请求头
    c.GetHeader("Authorization")

    // 绑定
    c.ShouldBindJSON(&obj)

    // 上下文值
    c.Set("userId", 123)
    userId, _ := c.Get("userId")
    userId := c.MustGet("userId").(int)

    // 获取客户端 IP
    clientIP := c.ClientIP()

    // Content-Type
    contentType := c.ContentType()

    // 是否是 WebSocket
    isWebSocket := c.IsWebsocket()

    // 中止请求
    c.Abort()
    c.AbortWithStatus(403)
    c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
})
```

## 模式

### Debug/Release 模式

```go
// 设置模式
gin.SetMode(gin.DebugMode)    // 开发模式（默认）
gin.SetMode(gin.ReleaseMode)  // 生产模式
gin.SetMode(gin.TestMode)     // 测试模式

// 或通过环境变量
// GIN_MODE=release ./myapp

func main() {
    // 生产环境
    gin.SetMode(gin.ReleaseMode)

    r := gin.New()
    r.Use(gin.Recovery())  // 只使用 Recovery

    r.Run()
}
```

## 完整示例

```go
package main

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

var users = []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}

func main() {
    r := gin.Default()

    // 获取所有用户
    r.GET("/users", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"data": users})
    })

    // 获取单个用户
    r.GET("/users/:id", func(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
            return
        }

        for _, user := range users {
            if user.ID == id {
                c.JSON(http.StatusOK, gin.H{"data": user})
                return
            }
        }

        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    })

    // 创建用户
    r.POST("/users", func(c *gin.Context) {
        var newUser User
        if err := c.ShouldBindJSON(&newUser); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        newUser.ID = len(users) + 1
        users = append(users, newUser)

        c.JSON(http.StatusCreated, gin.H{"data": newUser})
    })

    // 更新用户
    r.PUT("/users/:id", func(c *gin.Context) {
        id, _ := strconv.Atoi(c.Param("id"))

        var updateUser User
        if err := c.ShouldBindJSON(&updateUser); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        for i, user := range users {
            if user.ID == id {
                users[i].Name = updateUser.Name
                users[i].Email = updateUser.Email
                c.JSON(http.StatusOK, gin.H{"data": users[i]})
                return
            }
        }

        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    })

    // 删除用户
    r.DELETE("/users/:id", func(c *gin.Context) {
        id, _ := strconv.Atoi(c.Param("id"))

        for i, user := range users {
            if user.ID == id {
                users = append(users[:i], users[i+1:]...)
                c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
                return
            }
        }

        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    })

    r.Run(":8080")
}
```

## 总结

| Express.js | Gin |
|------------|-----|
| `req.params.id` | `c.Param("id")` |
| `req.query.page` | `c.Query("page")` |
| `req.body` | `c.ShouldBindJSON(&obj)` |
| `res.json()` | `c.JSON()` |
| `res.status(201)` | `c.JSON(201, ...)` |
| `app.use()` | `r.Use()` |
| `express.Router()` | `r.Group()` |

**下一节**：[路由](./03-routing.md) - 深入学习 Gin 路由
