# Web API 开发 (Web API Development)

## 学习目标

完成本章后，你将能够：
- 使用 net/http 构建基础 HTTP 服务
- 使用 Gin 框架开发 RESTful API
- 实现路由、中间件、请求验证
- 处理 JSON 请求和响应
- 生成 Swagger API 文档

## 与 JavaScript/Node.js 对比

| 功能 | Express.js | Gin (Go) |
|------|-----------|----------|
| 路由 | `app.get('/users/:id')` | `r.GET("/users/:id")` |
| 中间件 | `app.use(middleware)` | `r.Use(middleware)` |
| 请求体解析 | `req.body` (需要中间件) | `c.ShouldBindJSON(&obj)` |
| 响应 | `res.json({})` | `c.JSON(200, obj)` |
| 路由组 | `express.Router()` | `r.Group("/api")` |
| 验证 | Joi, express-validator | binding tags |

## 框架选择

### Go 常用 Web 框架

| 框架 | 特点 | Star | 推荐场景 |
|------|------|------|---------|
| **Gin** | 高性能，API 友好 | 70k+ | REST API，微服务 |
| Echo | 简洁，高性能 | 25k+ | REST API |
| Fiber | Express 风格，极快 | 28k+ | 熟悉 Express 的开发者 |
| Chi | 轻量，兼容 net/http | 15k+ | 标准库爱好者 |
| net/http | 标准库，无依赖 | - | 简单服务，学习 |

**本教程使用 Gin** - 最流行，文档丰富，适合生产环境。

## 章节内容

1. [HTTP 基础](./01-http-basics.md) - net/http 包基础
2. [Gin 入门](./02-gin-intro.md) - Gin 框架基础
3. [路由](./03-routing.md) - 路径参数、查询参数、路由组
4. [请求绑定](./04-request-binding.md) - JSON、Form、URI 绑定
5. [响应处理](./05-response.md) - JSON 响应和错误响应
6. [请求验证](./06-validation.md) - binding tags 验证
7. [中间件](./07-middleware.md) - 日志、恢复、CORS
8. [错误处理](./08-error-handling.md) - API 错误处理模式
9. [文件上传](./09-file-upload.md) - 文件上传处理
10. [Swagger 文档](./10-swagger.md) - OpenAPI 文档生成
11. [练习](./exercises.md) - 实践练习

## 示例代码

- [basic-api](../../../examples/web/basic-api/) - 完整的 REST API 示例

## 快速开始

### 安装 Gin

```bash
go get -u github.com/gin-gonic/gin
```

### Hello World

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    r.Run(":8080")
}
```

### RESTful API 示例

```go
package main

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var users = []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}

func main() {
    r := gin.Default()

    // 获取用户列表
    r.GET("/users", func(c *gin.Context) {
        c.JSON(http.StatusOK, users)
    })

    // 获取单个用户
    r.GET("/users/:id", func(c *gin.Context) {
        id, _ := strconv.Atoi(c.Param("id"))
        for _, user := range users {
            if user.ID == id {
                c.JSON(http.StatusOK, user)
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
        c.JSON(http.StatusCreated, newUser)
    })

    r.Run(":8080")
}
```

## API 设计最佳实践

### RESTful 设计原则

```
GET    /users          # 获取用户列表
GET    /users/:id      # 获取单个用户
POST   /users          # 创建用户
PUT    /users/:id      # 更新用户（完整替换）
PATCH  /users/:id      # 部分更新用户
DELETE /users/:id      # 删除用户

# 嵌套资源
GET    /users/:id/orders     # 获取用户的订单
POST   /users/:id/orders     # 为用户创建订单

# 查询和过滤
GET    /users?page=1&limit=10
GET    /users?status=active&sort=created_at
```

### 响应格式

```json
// 成功响应
{
  "data": { "id": 1, "name": "Alice" },
  "meta": { "page": 1, "total": 100 }
}

// 错误响应
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request",
    "details": [
      { "field": "email", "message": "must be a valid email" }
    ]
  }
}
```

### HTTP 状态码

| 状态码 | 含义 | 使用场景 |
|--------|------|---------|
| 200 | OK | GET 成功，PUT/PATCH 更新成功 |
| 201 | Created | POST 创建成功 |
| 204 | No Content | DELETE 成功 |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 未认证 |
| 403 | Forbidden | 无权限 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源冲突 |
| 422 | Unprocessable | 验证失败 |
| 500 | Internal Error | 服务器错误 |

## 项目结构

```
my-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/          # HTTP 处理器
│   │   └── user.go
│   ├── middleware/        # 中间件
│   │   ├── auth.go
│   │   └── logger.go
│   ├── routes/            # 路由定义
│   │   └── routes.go
│   ├── services/          # 业务逻辑
│   └── repositories/      # 数据访问
├── pkg/
│   └── response/          # 响应工具
├── configs/
├── api/
│   └── openapi/
│       └── spec.yaml
└── go.mod
```

## 实践项目

本章的知识将在以下项目中应用：
- [Todo API](../../02-practice/todo-api/) - 完整的 RESTful API 实践

## 下一步

完成本章后，继续学习 [认证授权](../06-authentication/)，为你的 API 添加安全保护。
