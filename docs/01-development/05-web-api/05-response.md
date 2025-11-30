# 响应处理 (Response Handling)

## 基础响应

### JSON 响应

```go
// 使用结构体
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

r.GET("/user", func(c *gin.Context) {
    user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
    c.JSON(200, user)
})

// 使用 gin.H（map 快捷方式）
r.GET("/message", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "message": "Hello, World!",
        "status":  "success",
    })
})
```

### 其他格式

```go
// 字符串
c.String(200, "Hello %s", name)

// XML
c.XML(200, user)

// YAML
c.YAML(200, user)

// ProtoBuf
c.ProtoBuf(200, protoMessage)

// HTML（需要先加载模板）
c.HTML(200, "index.html", gin.H{"title": "Hello"})

// 文件
c.File("./path/to/file.pdf")
c.FileAttachment("./path/to/file.pdf", "download.pdf")

// 重定向
c.Redirect(302, "/new-url")
c.Redirect(301, "https://google.com")

// 无内容
c.Status(204)
```

## 统一响应格式

### 定义响应结构

```go
// pkg/response/response.go
package response

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// PagedData 分页数据
type PagedData struct {
    Items      interface{} `json:"items"`
    Total      int64       `json:"total"`
    Page       int         `json:"page"`
    PageSize   int         `json:"page_size"`
    TotalPages int         `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, Response{
        Code:    0,
        Message: "created",
        Data:    data,
    })
}

// NoContent 无内容响应
func NoContent(c *gin.Context) {
    c.Status(http.StatusNoContent)
}

// Paged 分页响应
func Paged(c *gin.Context, items interface{}, total int64, page, pageSize int) {
    totalPages := int(total) / pageSize
    if int(total)%pageSize > 0 {
        totalPages++
    }

    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data: PagedData{
            Items:      items,
            Total:      total,
            Page:       page,
            PageSize:   pageSize,
            TotalPages: totalPages,
        },
    })
}
```

### 使用统一响应

```go
import "myapp/pkg/response"

r.GET("/users", func(c *gin.Context) {
    users := []User{...}
    total := int64(100)

    response.Paged(c, users, total, 1, 10)
})

r.GET("/users/:id", func(c *gin.Context) {
    user := User{ID: 1, Name: "Alice"}
    response.Success(c, user)
})

r.POST("/users", func(c *gin.Context) {
    // 创建用户...
    response.Created(c, newUser)
})

r.DELETE("/users/:id", func(c *gin.Context) {
    // 删除用户...
    response.NoContent(c)
})
```

## 错误响应

### 定义错误响应

```go
// pkg/response/error.go
package response

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type ErrorResponse struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

func Error(c *gin.Context, httpCode int, code, message string) {
    c.JSON(httpCode, ErrorResponse{
        Code:    code,
        Message: message,
    })
}

func ErrorWithDetails(c *gin.Context, httpCode int, code, message string, details interface{}) {
    c.JSON(httpCode, ErrorResponse{
        Code:    code,
        Message: message,
        Details: details,
    })
}

// 常用错误响应
func BadRequest(c *gin.Context, message string) {
    Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func ValidationError(c *gin.Context, details interface{}) {
    ErrorWithDetails(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)
}

func Unauthorized(c *gin.Context, message string) {
    if message == "" {
        message = "Unauthorized"
    }
    Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context) {
    Error(c, http.StatusForbidden, "FORBIDDEN", "Access denied")
}

func NotFound(c *gin.Context, resource string) {
    Error(c, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

func Conflict(c *gin.Context, message string) {
    Error(c, http.StatusConflict, "CONFLICT", message)
}

func InternalError(c *gin.Context) {
    Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
```

### 使用错误响应

```go
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")

    user, err := userService.GetByID(id)
    if err != nil {
        if errors.Is(err, services.ErrNotFound) {
            response.NotFound(c, "User")
            return
        }
        response.InternalError(c)
        return
    }

    response.Success(c, user)
})

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, formatValidationErrors(err))
        return
    }

    user, err := userService.Create(req)
    if err != nil {
        if errors.Is(err, services.ErrUserExists) {
            response.Conflict(c, "User already exists")
            return
        }
        response.InternalError(c)
        return
    }

    response.Created(c, user)
})
```

## 响应头

```go
r.GET("/headers", func(c *gin.Context) {
    // 设置单个头
    c.Header("X-Custom-Header", "value")

    // 设置多个头
    c.Writer.Header().Set("X-Request-ID", requestID)
    c.Writer.Header().Set("Cache-Control", "no-cache")

    // Content-Type（通常自动设置）
    c.Header("Content-Type", "application/json; charset=utf-8")

    c.JSON(200, gin.H{"message": "ok"})
})
```

## Cookie

```go
// 设置 Cookie
r.GET("/set-cookie", func(c *gin.Context) {
    c.SetCookie(
        "token",           // name
        "abc123",          // value
        3600,              // maxAge (秒)
        "/",               // path
        "localhost",       // domain
        false,             // secure (HTTPS only)
        true,              // httpOnly
    )

    c.JSON(200, gin.H{"message": "Cookie set"})
})

// 读取 Cookie
r.GET("/get-cookie", func(c *gin.Context) {
    token, err := c.Cookie("token")
    if err != nil {
        c.JSON(400, gin.H{"error": "Cookie not found"})
        return
    }
    c.JSON(200, gin.H{"token": token})
})

// 删除 Cookie
r.GET("/delete-cookie", func(c *gin.Context) {
    c.SetCookie("token", "", -1, "/", "localhost", false, true)
    c.JSON(200, gin.H{"message": "Cookie deleted"})
})
```

## 流式响应

```go
// Server-Sent Events (SSE)
r.GET("/events", func(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for i := 0; i < 10; i++ {
        select {
        case <-ticker.C:
            c.SSEvent("message", gin.H{
                "time": time.Now().Format(time.RFC3339),
                "id":   i,
            })
            c.Writer.Flush()
        case <-c.Request.Context().Done():
            return
        }
    }
})

// 流式写入
r.GET("/stream", func(c *gin.Context) {
    c.Stream(func(w io.Writer) bool {
        w.Write([]byte("chunk of data\n"))
        return false // false 表示结束
    })
})
```

## 文件下载

```go
// 直接发送文件
r.GET("/download/file", func(c *gin.Context) {
    c.File("./files/report.pdf")
})

// 带下载名的文件
r.GET("/download/attachment", func(c *gin.Context) {
    c.FileAttachment("./files/report.pdf", "monthly-report.pdf")
})

// 从读取器发送
r.GET("/download/reader", func(c *gin.Context) {
    data := []byte("Hello, World!")
    reader := bytes.NewReader(data)

    c.DataFromReader(
        http.StatusOK,
        int64(len(data)),
        "application/octet-stream",
        reader,
        map[string]string{
            "Content-Disposition": `attachment; filename="hello.txt"`,
        },
    )
})

// 动态生成文件
r.GET("/download/csv", func(c *gin.Context) {
    c.Header("Content-Type", "text/csv")
    c.Header("Content-Disposition", "attachment; filename=users.csv")

    c.Writer.WriteString("ID,Name,Email\n")
    c.Writer.WriteString("1,Alice,alice@example.com\n")
    c.Writer.WriteString("2,Bob,bob@example.com\n")
})
```

## 条件响应

```go
// ETag 支持
r.GET("/resource", func(c *gin.Context) {
    data := gin.H{"id": 1, "name": "Alice"}
    etag := calculateETag(data)

    c.Header("ETag", etag)

    // 检查 If-None-Match
    if c.GetHeader("If-None-Match") == etag {
        c.Status(http.StatusNotModified)
        return
    }

    c.JSON(200, data)
})

// Last-Modified 支持
r.GET("/resource", func(c *gin.Context) {
    lastModified := time.Now().UTC().Format(http.TimeFormat)

    c.Header("Last-Modified", lastModified)

    ifModifiedSince := c.GetHeader("If-Modified-Since")
    if ifModifiedSince == lastModified {
        c.Status(http.StatusNotModified)
        return
    }

    c.JSON(200, data)
})
```

## 完整示例

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// 响应结构
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Meta struct {
    Page       int   `json:"page,omitempty"`
    PageSize   int   `json:"page_size,omitempty"`
    Total      int64 `json:"total,omitempty"`
    TotalPages int   `json:"total_pages,omitempty"`
}

// 响应函数
func respondSuccess(c *gin.Context, status int, data interface{}) {
    c.JSON(status, APIResponse{
        Success: true,
        Data:    data,
    })
}

func respondError(c *gin.Context, status int, code, message string, details interface{}) {
    c.JSON(status, APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
            Details: details,
        },
    })
}

func respondPaged(c *gin.Context, data interface{}, total int64, page, pageSize int) {
    totalPages := int(total) / pageSize
    if int(total)%pageSize > 0 {
        totalPages++
    }

    c.JSON(http.StatusOK, APIResponse{
        Success: true,
        Data:    data,
        Meta: &Meta{
            Page:       page,
            PageSize:   pageSize,
            Total:      total,
            TotalPages: totalPages,
        },
    })
}

func main() {
    r := gin.Default()

    // 成功响应
    r.GET("/users", func(c *gin.Context) {
        users := []gin.H{
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
        }
        respondPaged(c, users, 100, 1, 10)
    })

    r.GET("/users/:id", func(c *gin.Context) {
        user := gin.H{"id": 1, "name": "Alice"}
        respondSuccess(c, http.StatusOK, user)
    })

    r.POST("/users", func(c *gin.Context) {
        user := gin.H{"id": 3, "name": "Charlie"}
        respondSuccess(c, http.StatusCreated, user)
    })

    // 错误响应
    r.GET("/error/not-found", func(c *gin.Context) {
        respondError(c, http.StatusNotFound, "NOT_FOUND", "User not found", nil)
    })

    r.GET("/error/validation", func(c *gin.Context) {
        respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", []gin.H{
            {"field": "email", "message": "invalid email format"},
        })
    })

    r.Run(":8080")
}
```

## 总结

| 方法 | 用途 |
|------|------|
| `c.JSON()` | JSON 响应 |
| `c.String()` | 文本响应 |
| `c.File()` | 文件响应 |
| `c.Redirect()` | 重定向 |
| `c.Status()` | 仅状态码 |
| `c.Header()` | 设置响应头 |
| `c.SetCookie()` | 设置 Cookie |

**下一节**：[请求验证](./06-validation.md) - 学习数据验证
