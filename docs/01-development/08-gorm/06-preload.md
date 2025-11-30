# 预加载 (Preload / Eager Loading)

## 概述

预加载用于解决 N+1 查询问题，一次性加载关联数据而不是逐条查询。

## N+1 问题

### 问题示例

```go
// ❌ N+1 问题
var users []User
db.Find(&users)  // 1 次查询

for _, user := range users {
    var posts []Post
    db.Where("author_id = ?", user.ID).Find(&posts)  // N 次查询
    user.Posts = posts
}
// 总计: 1 + N 次查询

// ✅ 使用 Preload
var users []User
db.Preload("Posts").Find(&users)  // 2 次查询
// SELECT * FROM users
// SELECT * FROM posts WHERE author_id IN (1, 2, 3, ...)
```

## 基本预加载

### 单个关联

```go
// 预加载 Has Many
var user User
db.Preload("Posts").First(&user, 1)

// 预加载 Belongs To
var post Post
db.Preload("Author").First(&post, 1)

// 预加载 Has One
var user User
db.Preload("Profile").First(&user, 1)

// 预加载 Many To Many
var post Post
db.Preload("Tags").First(&post, 1)
```

### 多个关联

```go
var user User
db.Preload("Posts").Preload("Profile").First(&user, 1)

// 或链式
db.Preload("Posts").
   Preload("Profile").
   Preload("Comments").
   First(&user, 1)
```

### 嵌套预加载

```go
// 预加载 User -> Posts -> Comments
var user User
db.Preload("Posts.Comments").First(&user, 1)
// 执行 3 次查询:
// SELECT * FROM users WHERE id = 1
// SELECT * FROM posts WHERE author_id = 1
// SELECT * FROM comments WHERE post_id IN (...)

// 多层嵌套
db.Preload("Posts.Comments.User").First(&user, 1)

// 混合嵌套
db.Preload("Posts").
   Preload("Posts.Comments").
   Preload("Posts.Tags").
   First(&user, 1)
```

## 条件预加载

### 基本条件

```go
// 只预加载已发布的文章
db.Preload("Posts", "status = ?", "published").Find(&users)

// 预加载并排序
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Order("posts.created_at DESC")
}).Find(&users)

// 预加载并限制数量
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Order("created_at DESC").Limit(5)
}).Find(&users)
// ⚠️ 注意: Limit 会应用到整个查询，不是每个用户
```

### 复杂条件

```go
// 多个条件
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "published").
              Where("created_at > ?", time.Now().AddDate(0, -1, 0)).
              Order("created_at DESC")
}).Find(&users)

// 带 Join 的预加载
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Joins("LEFT JOIN post_stats ON post_stats.post_id = posts.id").
              Where("post_stats.views > ?", 100)
}).Find(&users)
```

### 嵌套条件预加载

```go
// 嵌套预加载带条件
db.Preload("Posts", "status = ?", "published").
   Preload("Posts.Comments", func(db *gorm.DB) *gorm.DB {
       return db.Where("approved = ?", true).Order("created_at DESC")
   }).
   Find(&users)
```

## Joins 预加载

### 基本 Joins 预加载

```go
// 使用 Joins 替代 Preload（单次查询）
var users []User
db.Joins("Company").Find(&users)
// SELECT users.*, Company.* FROM users LEFT JOIN companies AS Company ON ...

// 带条件的 Joins 预加载
db.Joins("Company", db.Where(&Company{Status: "active"})).Find(&users)
```

### Joins vs Preload

```go
// Preload: 2 次查询，加载所有字段
db.Preload("Company").Find(&users)

// Joins: 1 次查询，但只能加载一层关联
db.Joins("Company").Find(&users)

// 何时使用 Joins:
// - 需要基于关联字段筛选主记录
// - 想减少查询次数
// - 只需要一层关联

// 何时使用 Preload:
// - 需要嵌套关联
// - 关联数据量大（Joins 可能产生大量重复数据）
// - Has Many 关系
```

### 带条件筛选

```go
// 使用 Joins 筛选主记录
db.Joins("Company").Where("Company.name = ?", "Acme").Find(&users)
// 只返回公司名为 Acme 的用户

// Inner Join vs Left Join
db.InnerJoins("Company").Find(&users)  // 只返回有公司的用户
db.Joins("Company").Find(&users)        // 返回所有用户（Left Join）
```

## 预加载所有关联

```go
// 预加载所有关联（不推荐）
db.Preload(clause.Associations).Find(&users)

// 嵌套预加载所有
db.Preload("Posts." + clause.Associations).Find(&users)
```

## 自定义预加载

### 自定义 SQL

```go
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Select("id, title, author_id").  // 只选择需要的字段
              Where("status = ?", "published").
              Order("created_at DESC")
}).Find(&users)
```

### 预加载到 Map

```go
type User struct {
    ID    uint
    Name  string
    Posts map[string]Post `gorm:"-"` // 忽略自动关联
}

var users []User
var posts []Post

db.Find(&users)
db.Where("author_id IN ?", getUserIDs(users)).Find(&posts)

// 手动映射
for i := range users {
    users[i].Posts = make(map[string]Post)
    for _, post := range posts {
        if post.AuthorID == users[i].ID {
            users[i].Posts[post.Title] = post
        }
    }
}
```

## 性能优化

### 选择性预加载

```go
// ❌ 预加载不需要的数据
db.Preload("Posts").Preload("Comments").Preload("Profile").Find(&users)

// ✅ 只预加载需要的
db.Preload("Profile").Find(&users)
```

### 预加载字段选择

```go
// 只预加载需要的字段
db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "title", "author_id")  // 不加载 content
}).Find(&users)
```

### 分批预加载

```go
// 对于大量数据，分批处理
var users []User
db.FindInBatches(&users, 100, func(tx *gorm.DB, batch int) error {
    // 为当前批次预加载
    var ids []uint
    for _, u := range users {
        ids = append(ids, u.ID)
    }

    var posts []Post
    tx.Where("author_id IN ?", ids).Find(&posts)

    // 映射到用户
    postMap := make(map[uint][]Post)
    for _, p := range posts {
        postMap[p.AuthorID] = append(postMap[p.AuthorID], p)
    }
    for i := range users {
        users[i].Posts = postMap[users[i].ID]
    }

    return nil
})
```

### 延迟加载

```go
// 手动延迟加载
type User struct {
    ID    uint
    Name  string
    posts []Post
}

func (u *User) GetPosts(db *gorm.DB) []Post {
    if u.posts == nil {
        db.Where("author_id = ?", u.ID).Find(&u.posts)
    }
    return u.posts
}
```

## 实际应用示例

### API 响应优化

```go
type PostService struct {
    db *gorm.DB
}

// ListPosts 列表页（简要信息）
func (s *PostService) ListPosts(page, pageSize int) ([]Post, error) {
    var posts []Post
    offset := (page - 1) * pageSize

    err := s.db.
        Select("id", "title", "created_at", "author_id").
        Preload("Author", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "name")
        }).
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&posts).Error

    return posts, err
}

// GetPost 详情页（完整信息）
func (s *PostService) GetPost(id uint) (*Post, error) {
    var post Post

    err := s.db.
        Preload("Author").
        Preload("Author.Profile").
        Preload("Tags").
        Preload("Comments", func(db *gorm.DB) *gorm.DB {
            return db.Where("parent_id IS NULL").
                      Order("created_at DESC").
                      Limit(20)
        }).
        Preload("Comments.User", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "name", "avatar")
        }).
        First(&post, id).Error

    if err != nil {
        return nil, err
    }

    return &post, nil
}

// GetPostsWithStats 带统计信息的列表
func (s *PostService) GetPostsWithStats(page, pageSize int) ([]PostWithStats, error) {
    type PostWithStats struct {
        Post
        CommentCount int64 `gorm:"column:comment_count"`
        ViewCount    int64 `gorm:"column:view_count"`
    }

    var posts []PostWithStats
    offset := (page - 1) * pageSize

    err := s.db.Model(&Post{}).
        Select(`posts.*,
                (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comment_count,
                COALESCE(post_stats.views, 0) as view_count`).
        Joins("LEFT JOIN post_stats ON post_stats.post_id = posts.id").
        Preload("Author", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "name")
        }).
        Preload("Tags").
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&posts).Error

    return posts, err
}
```

### 条件预加载封装

```go
type PreloadOptions struct {
    IncludeAuthor   bool
    IncludeTags     bool
    IncludeComments bool
    CommentLimit    int
}

func (s *PostService) GetPostsWithOptions(opts PreloadOptions) ([]Post, error) {
    query := s.db.Model(&Post{})

    if opts.IncludeAuthor {
        query = query.Preload("Author", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "name", "avatar")
        })
    }

    if opts.IncludeTags {
        query = query.Preload("Tags")
    }

    if opts.IncludeComments {
        query = query.Preload("Comments", func(db *gorm.DB) *gorm.DB {
            q := db.Order("created_at DESC")
            if opts.CommentLimit > 0 {
                q = q.Limit(opts.CommentLimit)
            }
            return q
        }).Preload("Comments.User")
    }

    var posts []Post
    err := query.Find(&posts).Error
    return posts, err
}

// 使用
posts, _ := service.GetPostsWithOptions(PreloadOptions{
    IncludeAuthor:   true,
    IncludeTags:     true,
    IncludeComments: true,
    CommentLimit:    5,
})
```

**下一节**：[钩子](./07-hooks.md) - 学习生命周期回调
