# 数据库练习 (Exercises)

## 练习 1：基础 CRUD

实现用户仓储的基本 CRUD 操作。

### 要求

1. 使用 `database/sql` 或 `pgx`
2. 实现以下方法：
   - `Create(user *User) error`
   - `GetByID(id int) (*User, error)`
   - `GetByEmail(email string) (*User, error)`
   - `Update(user *User) error`
   - `Delete(id int) error`
   - `List(offset, limit int) ([]*User, error)`

3. 正确处理 `sql.ErrNoRows`
4. 使用参数化查询

### 提示

```go
type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type UserRepository struct {
    db *sql.DB
}
```

---

## 练习 2：连接池配置

配置和监控数据库连接池。

### 要求

1. 配置 `database/sql` 连接池：
   - MaxOpenConns: 25
   - MaxIdleConns: 5
   - ConnMaxLifetime: 5 分钟
   - ConnMaxIdleTime: 1 分钟

2. 实现连接池监控函数，每 10 秒打印：
   - 打开的连接数
   - 使用中的连接数
   - 空闲连接数
   - 等待连接的请求数

3. 使用 Prometheus 导出指标（可选）

### 提示

```go
func monitorPool(db *sql.DB) {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        stats := db.Stats()
        // 打印统计信息
    }
}
```

---

## 练习 3：事务处理

实现银行转账功能。

### 要求

1. 创建 `accounts` 表：
   ```sql
   CREATE TABLE accounts (
       id SERIAL PRIMARY KEY,
       user_id INTEGER NOT NULL,
       balance DECIMAL(15,2) NOT NULL DEFAULT 0,
       created_at TIMESTAMP DEFAULT NOW()
   );
   ```

2. 实现 `Transfer(fromID, toID int, amount float64) error`：
   - 检查源账户余额
   - 扣款和入账在同一事务中
   - 余额不足时回滚
   - 记录转账日志

3. 使用 `BeginTx` 设置隔离级别为 `Serializable`

4. 实现乐观锁版本（使用 version 字段）

### 提示

```go
type Account struct {
    ID      int
    UserID  int
    Balance float64
    Version int  // 乐观锁
}

type TransferLog struct {
    ID        int
    FromID    int
    ToID      int
    Amount    float64
    CreatedAt time.Time
}
```

---

## 练习 4：批量操作

实现高效的批量插入。

### 要求

1. 创建 `products` 表
2. 实现三种批量插入方法：
   - `BatchInsertLoop` - 循环单条插入
   - `BatchInsertBulk` - 构建多值 INSERT
   - `BatchInsertCopy` - 使用 pgx CopyFrom

3. 比较三种方法插入 10000 条记录的性能
4. 使用预处理语句优化循环插入

### 提示

```go
type Product struct {
    ID    int
    Name  string
    Price float64
    Stock int
}

// 基准测试
func BenchmarkBatchInsertLoop(b *testing.B) { ... }
func BenchmarkBatchInsertBulk(b *testing.B) { ... }
func BenchmarkBatchInsertCopy(b *testing.B) { ... }
```

---

## 练习 5：数据库迁移

使用 golang-migrate 管理 schema。

### 要求

1. 安装 golang-migrate CLI
2. 创建以下迁移：
   - `000001_create_users_table`
   - `000002_create_posts_table`
   - `000003_add_status_to_users`
   - `000004_create_comments_table`

3. 每个迁移都要有 up 和 down 文件
4. 确保迁移可以完全回滚
5. 在 Go 代码中嵌入迁移文件

### 提示

```bash
migrate create -ext sql -dir migrations -seq create_users_table
```

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS
```

---

## 练习 6：复杂查询

实现文章列表查询。

### 要求

1. 查询包含：
   - 分页 (offset, limit)
   - 按状态筛选
   - 按作者筛选
   - 按创建时间排序
   - 包含作者信息 (JOIN)
   - 包含评论数量

2. 实现动态查询构建
3. 使用 squirrel 或手动构建

### 提示

```go
type PostQuery struct {
    Status   *string
    AuthorID *int
    OrderBy  string
    OrderDir string
    Offset   int
    Limit    int
}

type PostWithAuthor struct {
    Post
    Author       User
    CommentCount int
}

func (r *PostRepository) List(ctx context.Context, q *PostQuery) ([]*PostWithAuthor, error)
```

---

## 综合练习：博客 API 数据层

构建完整的博客系统数据访问层。

### 功能要求

1. **用户管理**
   - 注册（邮箱唯一）
   - 通过邮箱/ID 查询
   - 更新个人信息
   - 软删除

2. **文章管理**
   - CRUD 操作
   - 分页列表
   - 按状态/作者筛选
   - 发布/取消发布

3. **评论管理**
   - 添加评论
   - 嵌套评论
   - 删除评论（级联）

4. **标签系统**
   - 多对多关系
   - 按标签查询文章

### 技术要求

1. 使用分层架构（Repository 模式）
2. 使用 pgx 连接池
3. 正确处理事务
4. 使用迁移管理 schema
5. 包含单元测试

### 数据模型

```go
type User struct {
    ID        int
    Name      string
    Email     string
    Password  string
    Status    string  // active, inactive, deleted
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}

type Post struct {
    ID          int
    Title       string
    Content     string
    AuthorID    int
    Status      string  // draft, published, archived
    PublishedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Comment struct {
    ID        int
    PostID    int
    UserID    int
    ParentID  *int  // 嵌套评论
    Content   string
    CreatedAt time.Time
}

type Tag struct {
    ID   int
    Name string
}

type PostTag struct {
    PostID int
    TagID  int
}
```

### 数据库 Schema

```sql
-- migrations/000001_create_users.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- migrations/000002_create_posts.up.sql
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    author_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'draft',
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- migrations/000003_create_comments.up.sql
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    parent_id INTEGER REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- migrations/000004_create_tags.up.sql
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE post_tags (
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);
```

### 仓储接口

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    SoftDelete(ctx context.Context, id int) error
}

type PostRepository interface {
    Create(ctx context.Context, post *Post) error
    GetByID(ctx context.Context, id int) (*Post, error)
    List(ctx context.Context, filter *PostFilter) ([]*Post, int, error)
    Update(ctx context.Context, post *Post) error
    Delete(ctx context.Context, id int) error
    Publish(ctx context.Context, id int) error
    Unpublish(ctx context.Context, id int) error
    AddTags(ctx context.Context, postID int, tagIDs []int) error
    RemoveTags(ctx context.Context, postID int, tagIDs []int) error
}

type CommentRepository interface {
    Create(ctx context.Context, comment *Comment) error
    GetByPostID(ctx context.Context, postID int) ([]*Comment, error)
    Delete(ctx context.Context, id int) error
}

type TagRepository interface {
    Create(ctx context.Context, tag *Tag) error
    GetByName(ctx context.Context, name string) (*Tag, error)
    GetByPostID(ctx context.Context, postID int) ([]*Tag, error)
    List(ctx context.Context) ([]*Tag, error)
}
```

### 项目结构

```
blog-api/
├── cmd/
│   └── main.go
├── internal/
│   ├── models/
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── comment.go
│   │   └── tag.go
│   ├── repositories/
│   │   ├── user.go
│   │   ├── post.go
│   │   ├── comment.go
│   │   └── tag.go
│   └── database/
│       ├── postgres.go
│       └── migrations.go
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   └── ...
├── tests/
│   └── repository_test.go
└── go.mod
```

---

## 提交检查清单

### 基础功能
- [ ] CRUD 操作正常工作
- [ ] 参数化查询（无 SQL 注入）
- [ ] 正确处理 NULL 值
- [ ] 正确关闭 Rows

### 连接池
- [ ] 配置了连接池参数
- [ ] 实现了监控

### 事务
- [ ] 转账功能正确
- [ ] 支持回滚
- [ ] 处理了并发问题

### 迁移
- [ ] 创建了迁移文件
- [ ] 迁移可以回滚
- [ ] 在代码中集成迁移

### 代码质量
- [ ] 使用 Repository 模式
- [ ] 有错误处理
- [ ] 有基本测试
- [ ] 代码结构清晰
