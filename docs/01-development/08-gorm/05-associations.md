# 关联关系 (Associations)

## 概述

GORM 支持四种关联关系：Belongs To、Has One、Has Many 和 Many To Many。

## Belongs To (属于)

### 定义

```go
// User 属于一个 Company
type User struct {
    ID        uint
    Name      string
    CompanyID uint    // 外键
    Company   Company // 关联
}

type Company struct {
    ID   uint
    Name string
}
```

### 自定义外键

```go
type User struct {
    ID        uint
    Name      string
    // 使用自定义外键名
    CompanyRefer uint    `gorm:"column:company_id"`
    Company      Company `gorm:"foreignKey:CompanyRefer"`
}

// 引用不同的字段
type User struct {
    ID          uint
    Name        string
    CompanyCode string
    Company     Company `gorm:"foreignKey:CompanyCode;references:Code"`
}

type Company struct {
    ID   uint
    Code string // 被引用的字段
    Name string
}
```

### 关联操作

```go
// 创建时自动关联
user := User{
    Name: "Alice",
    Company: Company{Name: "Acme Inc"},
}
db.Create(&user)

// 查询时加载关联
var user User
db.Preload("Company").First(&user, 1)

// 设置关联
db.Model(&user).Association("Company").Append(&company)

// 替换关联
db.Model(&user).Association("Company").Replace(&newCompany)

// 清除关联
db.Model(&user).Association("Company").Clear()
```

## Has One (拥有一个)

### 定义

```go
// User 拥有一个 Profile
type User struct {
    ID      uint
    Name    string
    Profile Profile // Has One
}

type Profile struct {
    ID     uint
    UserID uint // 外键在 Profile 表
    Bio    string
    Avatar string
}
```

### 自定义外键

```go
type User struct {
    ID      uint
    Name    string
    Profile Profile `gorm:"foreignKey:UserRefer"`
}

type Profile struct {
    ID        uint
    UserRefer uint   // 自定义外键名
    Bio       string
}

// 引用不同字段
type User struct {
    ID      uint
    UUID    string `gorm:"uniqueIndex"`
    Name    string
    Profile Profile `gorm:"foreignKey:UserUUID;references:UUID"`
}

type Profile struct {
    ID       uint
    UserUUID string
    Bio      string
}
```

### 关联操作

```go
// 创建用户同时创建 Profile
user := User{
    Name: "Alice",
    Profile: Profile{
        Bio:    "Software Engineer",
        Avatar: "avatar.jpg",
    },
}
db.Create(&user)

// 查询
var user User
db.Preload("Profile").First(&user, 1)

// 添加/修改关联
profile := Profile{Bio: "Updated bio"}
db.Model(&user).Association("Profile").Append(&profile)
```

## Has Many (拥有多个)

### 定义

```go
// User 拥有多个 Post
type User struct {
    ID    uint
    Name  string
    Posts []Post `gorm:"foreignKey:AuthorID"`
}

type Post struct {
    ID       uint
    Title    string
    Content  string
    AuthorID uint // 外键
}
```

### 自定义外键

```go
type User struct {
    ID    uint
    Name  string
    Posts []Post `gorm:"foreignKey:AuthorID;references:ID"`
}

// 复合外键
type User struct {
    ID    uint
    Name  string
    Posts []Post `gorm:"foreignKey:AuthorID,AuthorName;references:ID,Name"`
}
```

### 关联操作

```go
// 创建用户和文章
user := User{
    Name: "Alice",
    Posts: []Post{
        {Title: "Post 1", Content: "Content 1"},
        {Title: "Post 2", Content: "Content 2"},
    },
}
db.Create(&user)

// 查询
var user User
db.Preload("Posts").First(&user, 1)

// 添加关联
newPost := Post{Title: "Post 3", Content: "Content 3"}
db.Model(&user).Association("Posts").Append(&newPost)

// 添加多个
posts := []Post{
    {Title: "Post 4"},
    {Title: "Post 5"},
}
db.Model(&user).Association("Posts").Append(&posts)

// 删除关联（不删除记录，只清除外键）
db.Model(&user).Association("Posts").Delete(&post)

// 清除所有关联
db.Model(&user).Association("Posts").Clear()

// 替换所有关联
db.Model(&user).Association("Posts").Replace(&newPosts)

// 统计关联数量
count := db.Model(&user).Association("Posts").Count()
```

## Many To Many (多对多)

### 基本定义

```go
// Post 和 Tag 多对多
type Post struct {
    ID    uint
    Title string
    Tags  []Tag `gorm:"many2many:post_tags;"`
}

type Tag struct {
    ID    uint
    Name  string
    Posts []Post `gorm:"many2many:post_tags;"`
}

// 自动创建中间表 post_tags:
// - post_id
// - tag_id
```

### 自定义中间表

```go
type Post struct {
    ID    uint
    Title string
    Tags  []Tag `gorm:"many2many:post_tags;foreignKey:ID;joinForeignKey:PostID;references:ID;joinReferences:TagID"`
}

// 手动定义中间表（需要额外字段时）
type PostTag struct {
    PostID    uint `gorm:"primaryKey"`
    TagID     uint `gorm:"primaryKey"`
    CreatedAt time.Time
    CreatedBy uint
}

// 设置表名
func (PostTag) TableName() string {
    return "post_tags"
}

// 迁移
db.AutoMigrate(&Post{}, &Tag{}, &PostTag{})
```

### 自引用多对多

```go
// 用户关注关系
type User struct {
    ID        uint
    Name      string
    Following []User `gorm:"many2many:user_follows;joinForeignKey:UserID;joinReferences:FollowingID"`
    Followers []User `gorm:"many2many:user_follows;joinForeignKey:FollowingID;joinReferences:UserID"`
}
```

### 关联操作

```go
// 创建带标签的文章
post := Post{
    Title: "GORM Tutorial",
    Tags: []Tag{
        {Name: "go"},
        {Name: "database"},
    },
}
db.Create(&post)

// 查询
var post Post
db.Preload("Tags").First(&post, 1)

// 添加标签
newTag := Tag{Name: "tutorial"}
db.Model(&post).Association("Tags").Append(&newTag)

// 添加已存在的标签
var existingTag Tag
db.First(&existingTag, "name = ?", "go")
db.Model(&post).Association("Tags").Append(&existingTag)

// 删除关联（只删除中间表记录）
db.Model(&post).Association("Tags").Delete(&tag)

// 替换所有标签
db.Model(&post).Association("Tags").Replace(&newTags)

// 清除所有关联
db.Model(&post).Association("Tags").Clear()

// 统计
count := db.Model(&post).Association("Tags").Count()
```

## 多态关联

### 定义

```go
// 评论可以属于 Post 或 Video
type Comment struct {
    ID              uint
    Content         string
    CommentableID   uint
    CommentableType string
}

type Post struct {
    ID       uint
    Title    string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}

type Video struct {
    ID       uint
    Title    string
    Comments []Comment `gorm:"polymorphic:Commentable;"`
}
```

### 使用

```go
// 创建 Post 评论
post := Post{
    Title: "Go Tutorial",
    Comments: []Comment{
        {Content: "Great post!"},
    },
}
db.Create(&post)
// comment.CommentableType = "posts"
// comment.CommentableID = post.ID

// 创建 Video 评论
video := Video{
    Title: "Go Video",
    Comments: []Comment{
        {Content: "Nice video!"},
    },
}
db.Create(&video)

// 查询
var post Post
db.Preload("Comments").First(&post, 1)
```

### 自定义多态类型

```go
type Comment struct {
    ID         uint
    Content    string
    OwnerID    uint
    OwnerType  string
}

type Post struct {
    ID       uint
    Title    string
    Comments []Comment `gorm:"polymorphic:Owner;polymorphicValue:blog_post"`
}
// comment.OwnerType = "blog_post" 而不是 "posts"
```

## 关联约束

### 外键约束

```go
type User struct {
    ID   uint
    Name string
}

type Post struct {
    ID       uint
    Title    string
    AuthorID uint
    Author   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// 生成的约束:
// FOREIGN KEY (author_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL
```

### 约束选项

```go
// CASCADE: 级联操作
// SET NULL: 设置为 NULL
// SET DEFAULT: 设置为默认值
// RESTRICT: 阻止操作
// NO ACTION: 不做任何操作

type Post struct {
    AuthorID uint
    Author   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
```

## 关联模式

### 跳过自动创建/更新

```go
// 创建时跳过关联
db.Omit("Company").Create(&user)
db.Omit("Company.*").Create(&user)

// 跳过所有关联
db.Omit(clause.Associations).Create(&user)

// 更新时跳过关联
db.Session(&gorm.Session{FullSaveAssociations: false}).Updates(&user)
```

### Select/Omit 关联字段

```go
// 只创建关联的特定字段
db.Select("Name", "Company.Name").Create(&user)

// 跳过创建关联的特定字段
db.Omit("Company.UpdatedAt").Create(&user)
```

## 实际应用示例

### 博客系统模型

```go
// 用户
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Name      string         `gorm:"size:100;not null" json:"name"`
    Email     string         `gorm:"size:255;uniqueIndex" json:"email"`
    Profile   *Profile       `json:"profile,omitempty"`
    Posts     []Post         `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
    Comments  []Comment      `gorm:"foreignKey:UserID" json:"comments,omitempty"`
    Followers []User         `gorm:"many2many:user_follows;joinForeignKey:FollowingID;joinReferences:UserID" json:"followers,omitempty"`
    Following []User         `gorm:"many2many:user_follows;joinForeignKey:UserID;joinReferences:FollowingID" json:"following,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 用户资料
type Profile struct {
    ID        uint   `gorm:"primaryKey" json:"id"`
    UserID    uint   `gorm:"uniqueIndex" json:"user_id"`
    Bio       string `gorm:"type:text" json:"bio"`
    Avatar    string `gorm:"size:500" json:"avatar"`
    Location  string `gorm:"size:100" json:"location"`
    Website   string `gorm:"size:255" json:"website"`
}

// 文章
type Post struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Title     string         `gorm:"size:255;not null" json:"title"`
    Content   string         `gorm:"type:text" json:"content"`
    Status    string         `gorm:"size:20;default:'draft'" json:"status"`
    AuthorID  uint           `gorm:"index" json:"author_id"`
    Author    *User          `json:"author,omitempty"`
    Tags      []Tag          `gorm:"many2many:post_tags;" json:"tags,omitempty"`
    Comments  []Comment      `gorm:"foreignKey:PostID" json:"comments,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 标签
type Tag struct {
    ID    uint   `gorm:"primaryKey" json:"id"`
    Name  string `gorm:"size:50;uniqueIndex" json:"name"`
    Slug  string `gorm:"size:50;uniqueIndex" json:"slug"`
    Posts []Post `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}

// 评论
type Comment struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Content   string         `gorm:"type:text;not null" json:"content"`
    PostID    uint           `gorm:"index" json:"post_id"`
    Post      *Post          `json:"post,omitempty"`
    UserID    uint           `gorm:"index" json:"user_id"`
    User      *User          `json:"user,omitempty"`
    ParentID  *uint          `gorm:"index" json:"parent_id"`
    Parent    *Comment       `json:"parent,omitempty"`
    Replies   []Comment      `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 博客服务

```go
type BlogService struct {
    db *gorm.DB
}

// CreatePost 创建文章
func (s *BlogService) CreatePost(authorID uint, title, content string, tagNames []string) (*Post, error) {
    // 处理标签
    var tags []Tag
    for _, name := range tagNames {
        var tag Tag
        // 查找或创建标签
        s.db.FirstOrCreate(&tag, Tag{Name: name, Slug: strings.ToLower(name)})
        tags = append(tags, tag)
    }

    post := &Post{
        Title:    title,
        Content:  content,
        AuthorID: authorID,
        Status:   "published",
        Tags:     tags,
    }

    if err := s.db.Create(post).Error; err != nil {
        return nil, err
    }

    return post, nil
}

// GetPostWithDetails 获取文章详情
func (s *BlogService) GetPostWithDetails(id uint) (*Post, error) {
    var post Post
    err := s.db.
        Preload("Author").
        Preload("Tags").
        Preload("Comments", func(db *gorm.DB) *gorm.DB {
            return db.Where("parent_id IS NULL").Order("created_at DESC")
        }).
        Preload("Comments.User").
        Preload("Comments.Replies").
        Preload("Comments.Replies.User").
        First(&post, id).Error

    if err != nil {
        return nil, err
    }
    return &post, nil
}

// FollowUser 关注用户
func (s *BlogService) FollowUser(userID, followingID uint) error {
    var user, following User
    if err := s.db.First(&user, userID).Error; err != nil {
        return err
    }
    if err := s.db.First(&following, followingID).Error; err != nil {
        return err
    }
    return s.db.Model(&user).Association("Following").Append(&following)
}

// UnfollowUser 取消关注
func (s *BlogService) UnfollowUser(userID, followingID uint) error {
    var user, following User
    if err := s.db.First(&user, userID).Error; err != nil {
        return err
    }
    if err := s.db.First(&following, followingID).Error; err != nil {
        return err
    }
    return s.db.Model(&user).Association("Following").Delete(&following)
}
```

**下一节**：[预加载](./06-preload.md) - 学习 Eager Loading
