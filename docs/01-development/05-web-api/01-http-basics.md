# HTTP 基础 (HTTP Basics)

## net/http 包

Go 标准库提供了强大的 `net/http` 包，无需任何第三方依赖即可构建 HTTP 服务。

## 与 Node.js 对比

### Node.js 原生 HTTP

```javascript
const http = require('http');

const server = http.createServer((req, res) => {
  if (req.url === '/hello' && req.method === 'GET') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ message: 'Hello!' }));
  }
});

server.listen(8080);
```

### Go net/http

```go
package main

import (
    "encoding/json"
    "net/http"
)

func main() {
    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"message": "Hello!"})
    })

    http.ListenAndServe(":8080", nil)
}
```

## 基础服务器

### 最简单的服务器

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    // 注册处理函数
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })

    // 启动服务器
    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", nil)
}
```

### 使用自定义 ServeMux

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/", homeHandler)
    mux.HandleFunc("/about", aboutHandler)
    mux.HandleFunc("/api/users", usersHandler)

    fmt.Println("Server starting on :8080")
    http.ListenAndServe(":8080", mux)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    fmt.Fprintf(w, "Welcome to Home!")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "About Page")
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        fmt.Fprintf(w, "Get users")
    case http.MethodPost:
        fmt.Fprintf(w, "Create user")
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
```

## Go 1.22+ 新路由

Go 1.22 引入了增强的路由模式：

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    mux := http.NewServeMux()

    // 方法匹配
    mux.HandleFunc("GET /users", listUsers)
    mux.HandleFunc("POST /users", createUser)
    mux.HandleFunc("GET /users/{id}", getUser)
    mux.HandleFunc("PUT /users/{id}", updateUser)
    mux.HandleFunc("DELETE /users/{id}", deleteUser)

    http.ListenAndServe(":8080", mux)
}

func listUsers(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "List all users")
}

func createUser(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Create user")
}

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // 获取路径参数
    fmt.Fprintf(w, "Get user: %s", id)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Fprintf(w, "Update user: %s", id)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    fmt.Fprintf(w, "Delete user: %s", id)
}
```

## 请求处理

### 获取请求信息

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // 请求方法
    method := r.Method  // GET, POST, etc.

    // URL 路径
    path := r.URL.Path  // /users/123

    // 查询参数
    query := r.URL.Query()
    page := query.Get("page")     // ?page=1
    limit := query.Get("limit")   // ?limit=10

    // 请求头
    contentType := r.Header.Get("Content-Type")
    auth := r.Header.Get("Authorization")

    // 远程地址
    remoteAddr := r.RemoteAddr

    fmt.Printf("Method: %s, Path: %s, Page: %s\n", method, path, page)
}
```

### 读取请求体

```go
import (
    "encoding/json"
    "io"
    "net/http"
)

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    // 读取原始请求体
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // 或者直接解码 JSON
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // 处理请求...
}
```

### 处理表单数据

```go
func formHandler(w http.ResponseWriter, r *http.Request) {
    // 解析表单（application/x-www-form-urlencoded）
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }

    name := r.FormValue("name")
    email := r.FormValue("email")

    fmt.Fprintf(w, "Name: %s, Email: %s", name, email)
}

func multipartHandler(w http.ResponseWriter, r *http.Request) {
    // 解析 multipart/form-data (用于文件上传)
    if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
        http.Error(w, "Error parsing multipart", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("upload")
    if err != nil {
        http.Error(w, "Error getting file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    fmt.Fprintf(w, "Uploaded: %s (%d bytes)", header.Filename, header.Size)
}
```

## 响应处理

### 设置响应

```go
func responseHandler(w http.ResponseWriter, r *http.Request) {
    // 设置响应头
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Custom-Header", "value")

    // 设置状态码（必须在 Write 之前调用）
    w.WriteHeader(http.StatusOK)

    // 写入响应体
    w.Write([]byte(`{"message": "success"}`))
}
```

### JSON 响应

```go
type Response struct {
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Message string      `json:"message,omitempty"`
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func handler(w http.ResponseWriter, r *http.Request) {
    user := User{ID: 1, Name: "Alice"}
    jsonResponse(w, http.StatusOK, Response{Data: user})
}
```

## 中间件

### 基本中间件模式

```go
// 中间件函数签名
type Middleware func(http.Handler) http.Handler

// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // 调用下一个处理器
        next.ServeHTTP(w, r)

        // 记录请求
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// 恢复中间件
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic: %v", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// 使用中间件
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", homeHandler)

    // 包装中间件
    handler := loggingMiddleware(recoveryMiddleware(mux))

    http.ListenAndServe(":8080", handler)
}
```

### 中间件链

```go
func chain(handler http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", homeHandler)

    handler := chain(mux,
        loggingMiddleware,
        recoveryMiddleware,
        corsMiddleware,
    )

    http.ListenAndServe(":8080", handler)
}
```

## 服务器配置

### 自定义服务器

```go
package main

import (
    "log"
    "net/http"
    "time"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", homeHandler)

    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    log.Println("Server starting on :8080")
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

### 优雅关闭

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", homeHandler)

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // 启动服务器
    go func() {
        log.Println("Server starting on :8080")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    // 优雅关闭，等待最多 30 秒
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exited")
}
```

## HTTP 客户端

### 基本请求

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

func main() {
    // 创建带超时的客户端
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    // GET 请求
    resp, err := client.Get("https://api.example.com/users")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

### POST 请求

```go
func postJSON(url string, data interface{}) (*http.Response, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer token")

    client := &http.Client{Timeout: 10 * time.Second}
    return client.Do(req)
}
```

## 总结

| 场景 | 推荐 |
|------|------|
| 学习/简单服务 | net/http |
| 生产 API | Gin/Echo |
| 需要路由参数 | Go 1.22+ net/http 或框架 |
| 微服务内部 | net/http |

**下一节**：[Gin 入门](./02-gin-intro.md) - 学习更强大的 Web 框架
