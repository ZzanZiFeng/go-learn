# Web API 练习 (Exercises)

## 练习 1：基础 CRUD API

创建一个书籍管理 API。

### 要求

1. 实现以下端点：
   - `GET /books` - 获取书籍列表（支持分页、搜索）
   - `GET /books/:id` - 获取单本书籍
   - `POST /books` - 创建书籍
   - `PUT /books/:id` - 更新书籍
   - `DELETE /books/:id` - 删除书籍

2. 书籍模型：
   ```go
   type Book struct {
       ID        uint      `json:"id"`
       Title     string    `json:"title"`
       Author    string    `json:"author"`
       ISBN      string    `json:"isbn"`
       Price     float64   `json:"price"`
       Stock     int       `json:"stock"`
       CreatedAt time.Time `json:"created_at"`
   }
   ```

3. 实现请求验证
4. 使用统一响应格式
5. 添加 Swagger 文档

### 提示

```go
// 验证规则
type CreateBookRequest struct {
    Title  string  `json:"title" binding:"required,min=1,max=200"`
    Author string  `json:"author" binding:"required,min=1,max=100"`
    ISBN   string  `json:"isbn" binding:"required,len=13"`
    Price  float64 `json:"price" binding:"required,gt=0"`
    Stock  int     `json:"stock" binding:"gte=0"`
}
```

---

## 练习 2：认证中间件

实现 JWT 认证系统。

### 要求

1. 实现以下端点：
   - `POST /auth/register` - 用户注册
   - `POST /auth/login` - 用户登录（返回 JWT）
   - `POST /auth/refresh` - 刷新 token
   - `GET /auth/me` - 获取当前用户（需认证）

2. 实现认证中间件
3. 实现角色权限中间件
4. 管理员专属端点 `GET /admin/users`

### 提示

```go
// JWT Claims
type Claims struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

// 生成 token
func GenerateToken(userID uint, role string) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte("secret"))
}
```

---

## 练习 3：文件上传服务

创建一个图片上传服务。

### 要求

1. 实现以下端点：
   - `POST /images` - 上传图片
   - `GET /images/:id` - 获取图片信息
   - `GET /images/:id/file` - 下载图片
   - `DELETE /images/:id` - 删除图片

2. 验证：
   - 文件类型：只允许 jpg, png, gif
   - 文件大小：最大 5MB
   - 生成唯一文件名

3. 生成缩略图（可选）
4. 返回图片元信息

### 提示

```go
type ImageInfo struct {
    ID          string    `json:"id"`
    Filename    string    `json:"filename"`
    Size        int64     `json:"size"`
    ContentType string    `json:"content_type"`
    URL         string    `json:"url"`
    ThumbnailURL string   `json:"thumbnail_url,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}
```

---

## 练习 4：限流和日志

实现 API 限流和结构化日志。

### 要求

1. 实现基于 IP 的限流中间件：
   - 每分钟最多 60 次请求
   - 返回 429 状态码和重试时间

2. 实现结构化日志中间件：
   - 记录请求方法、路径、状态码、耗时
   - 记录请求 ID
   - 使用 JSON 格式

3. 实现请求 ID 中间件

### 提示

```go
// 限流响应
type RateLimitResponse struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    RetryAfter int    `json:"retry_after"` // 秒
}

// 日志条目
type LogEntry struct {
    RequestID  string        `json:"request_id"`
    Method     string        `json:"method"`
    Path       string        `json:"path"`
    Status     int           `json:"status"`
    Latency    time.Duration `json:"latency"`
    ClientIP   string        `json:"client_ip"`
    UserAgent  string        `json:"user_agent"`
    Error      string        `json:"error,omitempty"`
}
```

---

## 练习 5：RESTful API 设计

设计一个博客系统 API。

### 要求

1. 资源：
   - 用户 (Users)
   - 文章 (Posts)
   - 评论 (Comments)
   - 标签 (Tags)

2. 实现关系：
   - 用户有多篇文章
   - 文章有多条评论
   - 文章有多个标签

3. 端点设计：
   ```
   GET    /users/:id/posts          # 用户的文章
   GET    /posts/:id/comments       # 文章的评论
   POST   /posts/:id/comments       # 创建评论
   GET    /posts?tag=go             # 按标签筛选
   ```

4. 实现过滤、排序、分页
5. 添加完整的错误处理

### 提示

```go
// 查询参数
type PostQuery struct {
    Page    int    `form:"page,default=1"`
    Size    int    `form:"size,default=10"`
    Sort    string `form:"sort,default=created_at"`
    Order   string `form:"order,default=desc" binding:"oneof=asc desc"`
    Tag     string `form:"tag"`
    Author  uint   `form:"author"`
    Search  string `form:"search"`
}
```

---

## 练习 6：错误处理系统

实现完整的错误处理系统。

### 要求

1. 定义应用错误类型
2. 实现验证错误格式化
3. 实现错误处理中间件
4. 实现 panic 恢复
5. 区分开发/生产环境错误信息

### 提示

```go
// 应用错误
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Details    any
    Internal   error
}

// 错误响应
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
    Stack   string `json:"stack,omitempty"` // 仅开发环境
}
```

---

## 综合练习：Todo API

创建一个完整的 Todo 应用 API。

### 功能要求

1. **用户管理**
   - 注册、登录、登出
   - JWT 认证

2. **Todo 管理**
   - CRUD 操作
   - 状态切换（pending/completed）
   - 按状态、日期筛选
   - 支持优先级

3. **分类管理**
   - 创建、编辑、删除分类
   - Todo 关联分类

### 技术要求

1. 使用分层架构（Handler/Service/Repository）
2. 统一响应格式
3. 完整的请求验证
4. 错误处理中间件
5. 日志中间件
6. Swagger 文档

### 数据模型

```go
type User struct {
    ID        uint      `json:"id"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

type Category struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    Name      string    `json:"name"`
    Color     string    `json:"color"`
    CreatedAt time.Time `json:"created_at"`
}

type Todo struct {
    ID          uint       `json:"id"`
    UserID      uint       `json:"user_id"`
    CategoryID  *uint      `json:"category_id"`
    Title       string     `json:"title"`
    Description string     `json:"description"`
    Priority    string     `json:"priority"` // low, medium, high
    Status      string     `json:"status"`   // pending, completed
    DueDate     *time.Time `json:"due_date"`
    CompletedAt *time.Time `json:"completed_at"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

### API 端点

```
# 认证
POST   /auth/register
POST   /auth/login
POST   /auth/refresh
GET    /auth/me

# 分类
GET    /categories
POST   /categories
GET    /categories/:id
PUT    /categories/:id
DELETE /categories/:id

# Todo
GET    /todos
POST   /todos
GET    /todos/:id
PUT    /todos/:id
DELETE /todos/:id
PATCH  /todos/:id/status
GET    /todos/stats

# 筛选示例
GET    /todos?status=pending&priority=high&category_id=1&due_before=2024-12-31
```

### 项目结构

```
todo-api/
├── cmd/
│   └── main.go
├── internal/
│   ├── models/
│   │   ├── user.go
│   │   ├── category.go
│   │   └── todo.go
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── category.go
│   │   └── todo.go
│   ├── services/
│   │   ├── auth.go
│   │   ├── category.go
│   │   └── todo.go
│   ├── repositories/
│   │   ├── user.go
│   │   ├── category.go
│   │   └── todo.go
│   └── middleware/
│       ├── auth.go
│       ├── logger.go
│       └── error.go
├── pkg/
│   ├── response/
│   │   └── response.go
│   ├── apperror/
│   │   └── error.go
│   └── validator/
│       └── validator.go
├── docs/
│   └── swagger/
└── go.mod
```

---

## 提交检查清单

- [ ] 代码能够编译运行
- [ ] 所有端点按预期工作
- [ ] 请求验证正确实现
- [ ] 错误处理完整
- [ ] 响应格式统一
- [ ] 添加了必要的中间件
- [ ] Swagger 文档完整
- [ ] 代码结构清晰
