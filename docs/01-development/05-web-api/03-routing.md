# 路由 (Routing)

## 基础路由

### HTTP 方法路由

```go
func main() {
    r := gin.Default()

    r.GET("/get", getHandler)
    r.POST("/post", postHandler)
    r.PUT("/put", putHandler)
    r.PATCH("/patch", patchHandler)
    r.DELETE("/delete", deleteHandler)
    r.HEAD("/head", headHandler)
    r.OPTIONS("/options", optionsHandler)

    // 匹配任何方法
    r.Any("/any", anyHandler)

    // 只匹配指定方法
    r.Handle("GET", "/handle", handleHandler)

    r.Run()
}
```

## 路径参数

### 单个参数

```go
// 匹配 /users/123, /users/abc
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})

// 多个参数
// 匹配 /users/123/orders/456
r.GET("/users/:userId/orders/:orderId", func(c *gin.Context) {
    userId := c.Param("userId")
    orderId := c.Param("orderId")
    c.JSON(200, gin.H{
        "userId":  userId,
        "orderId": orderId,
    })
})
```

### 通配符参数

```go
// *filepath 匹配所有剩余路径
// /files/docs/report.pdf -> filepath = "docs/report.pdf"
r.GET("/files/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")  // "/docs/report.pdf"
    c.String(200, "File: %s", filepath)
})

// 注意：通配符必须在路径末尾
```

### 可选参数（需要多个路由）

```go
// Go/Gin 不直接支持可选参数，需要注册多个路由
r.GET("/users", listUsers)
r.GET("/users/:id", getUser)

// 或者使用通配符 + 手动解析
```

## 查询参数

```go
// /search?keyword=go&page=1&limit=10
r.GET("/search", func(c *gin.Context) {
    // 基本获取
    keyword := c.Query("keyword")

    // 带默认值
    page := c.DefaultQuery("page", "1")
    limit := c.DefaultQuery("limit", "10")

    // 检查是否存在
    sort, exists := c.GetQuery("sort")

    // 获取数组 ?tags=go&tags=web&tags=api
    tags := c.QueryArray("tags")

    // 获取 map ?filter[status]=active&filter[type]=admin
    filters := c.QueryMap("filter")

    c.JSON(200, gin.H{
        "keyword": keyword,
        "page":    page,
        "limit":   limit,
        "sort":    sort,
        "exists":  exists,
        "tags":    tags,
        "filters": filters,
    })
})
```

### 绑定查询参数到结构体

```go
type SearchParams struct {
    Keyword string   `form:"keyword"`
    Page    int      `form:"page,default=1"`
    Limit   int      `form:"limit,default=10"`
    Tags    []string `form:"tags"`
}

r.GET("/search", func(c *gin.Context) {
    var params SearchParams
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, params)
})
```

## 路由组

### 基本分组

```go
func main() {
    r := gin.Default()

    // API v1
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", getUsers)
        v1.POST("/users", createUser)
        v1.GET("/users/:id", getUser)
    }

    // API v2
    v2 := r.Group("/api/v2")
    {
        v2.GET("/users", getUsersV2)
    }

    r.Run()
}
```

### 嵌套分组

```go
func main() {
    r := gin.Default()

    api := r.Group("/api")
    {
        // /api/v1/...
        v1 := api.Group("/v1")
        {
            users := v1.Group("/users")
            {
                users.GET("", listUsers)      // GET /api/v1/users
                users.POST("", createUser)    // POST /api/v1/users
                users.GET("/:id", getUser)    // GET /api/v1/users/:id

                // 嵌套资源
                orders := users.Group("/:userId/orders")
                {
                    orders.GET("", listOrders)    // GET /api/v1/users/:userId/orders
                    orders.POST("", createOrder)  // POST /api/v1/users/:userId/orders
                }
            }
        }
    }

    r.Run()
}
```

### 组级中间件

```go
func main() {
    r := gin.Default()

    // 公开路由
    public := r.Group("/api")
    {
        public.POST("/login", login)
        public.POST("/register", register)
    }

    // 需要认证的路由
    authorized := r.Group("/api")
    authorized.Use(authMiddleware())  // 应用中间件
    {
        authorized.GET("/profile", getProfile)
        authorized.PUT("/profile", updateProfile)

        // 管理员路由
        admin := authorized.Group("/admin")
        admin.Use(adminMiddleware())  // 额外的中间件
        {
            admin.GET("/users", adminListUsers)
            admin.DELETE("/users/:id", adminDeleteUser)
        }
    }

    r.Run()
}
```

## 与 Express.js 对比

### Express.js Router

```javascript
// routes/users.js
const router = express.Router();

router.get('/', getUsers);
router.post('/', createUser);
router.get('/:id', getUser);
router.put('/:id', updateUser);

module.exports = router;

// app.js
const usersRouter = require('./routes/users');
app.use('/api/users', usersRouter);
```

### Gin 路由组

```go
// routes/users.go
func SetupUserRoutes(rg *gin.RouterGroup) {
    users := rg.Group("/users")
    {
        users.GET("", getUsers)
        users.POST("", createUser)
        users.GET("/:id", getUser)
        users.PUT("/:id", updateUser)
    }
}

// main.go
func main() {
    r := gin.Default()

    api := r.Group("/api")
    routes.SetupUserRoutes(api)

    r.Run()
}
```

## 路由优先级

Gin 路由匹配顺序：
1. 静态路由 (`/users/new`)
2. 参数路由 (`/users/:id`)
3. 通配符路由 (`/users/*path`)

```go
// 这些路由可以共存
r.GET("/users/new", newUserForm)      // 精确匹配 /users/new
r.GET("/users/:id", getUser)          // 匹配 /users/123
r.GET("/users/:id/edit", editUser)    // 匹配 /users/123/edit
```

## 静态文件服务

```go
func main() {
    r := gin.Default()

    // 单个静态文件
    r.StaticFile("/favicon.ico", "./resources/favicon.ico")

    // 静态文件目录
    r.Static("/static", "./static")

    // 静态文件目录（使用 http.FileSystem）
    r.StaticFS("/assets", http.Dir("assets"))

    r.Run()
}
```

## 路由信息

```go
func main() {
    r := gin.Default()

    r.GET("/users", getUsers)
    r.POST("/users", createUser)

    // 获取所有路由
    routes := r.Routes()
    for _, route := range routes {
        fmt.Printf("%s %s\n", route.Method, route.Path)
    }
    // 输出:
    // GET /users
    // POST /users

    r.Run()
}
```

## 完整示例：模块化路由

```go
// routes/routes.go
package routes

import "github.com/gin-gonic/gin"

func Setup(r *gin.Engine) {
    api := r.Group("/api/v1")
    {
        SetupUserRoutes(api)
        SetupOrderRoutes(api)
        SetupProductRoutes(api)
    }
}

// routes/user.go
package routes

import (
    "github.com/gin-gonic/gin"
    "myapp/internal/handlers"
)

func SetupUserRoutes(rg *gin.RouterGroup) {
    h := handlers.NewUserHandler()

    users := rg.Group("/users")
    {
        users.GET("", h.List)
        users.POST("", h.Create)
        users.GET("/:id", h.Get)
        users.PUT("/:id", h.Update)
        users.DELETE("/:id", h.Delete)

        // 嵌套资源
        users.GET("/:id/orders", h.GetOrders)
    }
}

// routes/order.go
package routes

func SetupOrderRoutes(rg *gin.RouterGroup) {
    orders := rg.Group("/orders")
    {
        orders.GET("", listOrders)
        orders.POST("", createOrder)
        orders.GET("/:id", getOrder)
    }
}

// main.go
package main

import (
    "github.com/gin-gonic/gin"
    "myapp/routes"
)

func main() {
    r := gin.Default()
    routes.Setup(r)
    r.Run()
}
```

## RESTful API 设计

```go
func SetupRESTRoutes(r *gin.Engine) {
    api := r.Group("/api/v1")
    {
        // 用户资源
        users := api.Group("/users")
        {
            users.GET("", listUsers)           // 获取列表
            users.POST("", createUser)         // 创建
            users.GET("/:id", getUser)         // 获取单个
            users.PUT("/:id", updateUser)      // 完整更新
            users.PATCH("/:id", patchUser)     // 部分更新
            users.DELETE("/:id", deleteUser)   // 删除

            // 子资源
            users.GET("/:id/posts", getUserPosts)
            users.POST("/:id/posts", createUserPost)
        }

        // 文章资源
        posts := api.Group("/posts")
        {
            posts.GET("", listPosts)
            posts.POST("", createPost)
            posts.GET("/:id", getPost)
            posts.PUT("/:id", updatePost)
            posts.DELETE("/:id", deletePost)

            // 评论子资源
            posts.GET("/:id/comments", getPostComments)
            posts.POST("/:id/comments", createPostComment)
        }
    }
}
```

## 总结

| 功能 | Express | Gin |
|------|---------|-----|
| 路径参数 | `/:id` | `/:id` |
| 通配符 | `/*` | `/*path` |
| 路由组 | `express.Router()` | `r.Group()` |
| 组级中间件 | `router.use()` | `group.Use()` |
| 静态文件 | `express.static()` | `r.Static()` |

**下一节**：[请求绑定](./04-request-binding.md) - 学习请求数据绑定
