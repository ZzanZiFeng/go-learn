# Data Model: 前端工程师Go后端开发学习教程

**Feature**: [spec.md](./spec.md) | **Date**: 2025-11-30

## Overview

本项目是一个教程文档项目，核心数据模型围绕教程内容组织结构和示例代码项目展开。由于这是文档型项目而非应用型项目，数据模型主要用于定义内容结构和示例项目中的数据库模型。

## Part 1: 教程内容模型 (Tutorial Content Model)

### 1.1 章节 (Chapter)

教程的基本组织单元。

```
Chapter {
  id: string              // 章节编号，如 "01", "02"
  title: string           // 章节标题
  priority: P1 | P2 | P3  // 学习优先级
  difficulty: 🟢 | 🟡 | 🔴  // 难度等级
  prerequisites: string[] // 前置章节ID列表
  learningObjectives: string[] // 学习目标列表
  content: Markdown       // 正文内容
  examples: CodeExample[] // 代码示例列表
  exercises: Exercise[]   // 练习题列表
  summary: string         // 章节总结
}
```

**章节依赖关系**:
```
01-environment-setup (P1)
    ↓
02-syntax-basics (P1)
    ↓
03-struct-interface (P2)
    ↓
04-concurrency (P2)
    ↓
05-architecture (P2) ─────────────────────┐
    ↓                                      │
06-web-api (P2) ──────────────────────────┤
    ↓                                      │
07-authentication (P2)                     │
    ↓                                      │
08-database-sql (P2)                       │
    ↓                                      │
09-gorm (P2)                               │
    ↓                                      │
10-caching (P2)                            │
    ↓                                      ↓
11-message-queue (P3) ←── 12-observability (P3)
    ↓                           ↓
13-deployment (P3) ←────────────┘
    ↓
14-debugging (P3)
```

### 1.2 代码示例 (Code Example)

可独立运行的Go程序。

```
CodeExample {
  id: string              // 示例编号
  title: string           // 示例标题
  description: string     // 说明文字
  sourceCode: string      // 完整源代码
  expectedOutput: string  // 预期输出
  goVersion: string       // Go版本要求（如 "1.21+"）
  dependencies: string[]  // 依赖包列表
  jsComparison?: string   // JS/TS对比代码（可选）
  keyPoints: string[]     // 关键点解释
}
```

### 1.3 练习题 (Exercise)

学习验收检查点。

```
Exercise {
  id: string                 // 练习编号
  title: string              // 练习标题
  difficulty: 🟢 | 🟡 | 🔴    // 难度等级
  description: string        // 题目描述
  requirements: string[]     // 具体要求列表
  acceptanceCriteria: string[] // 验收标准
  hints: string[]            // 提示（可选展开）
  referenceAnswer: string    // 参考答案
  estimatedTime: string      // 预计完成时间
}
```

### 1.4 综合项目 (Project)

整合多个知识点的完整应用。

```
Project {
  id: string              // 项目ID
  name: string            // 项目名称
  description: string     // 项目描述
  chapters: string[]      // 涵盖的章节列表
  techStack: string[]     // 技术栈
  structure: FileTree     // 项目文件结构
  setupSteps: string[]    // 环境搭建步骤
  runningSteps: string[]  // 运行步骤
  verificationChecklist: string[] // 验证清单
}
```

## Part 2: 示例项目数据模型 (Example Project Data Models)

### 2.1 用户模型 (User)

用于认证、GORM关联关系等示例。

```go
// User 用户模型
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    // 基本信息
    Username  string `gorm:"size:50;uniqueIndex;not null" json:"username"`
    Email     string `gorm:"size:100;uniqueIndex;not null" json:"email"`
    Password  string `gorm:"size:255;not null" json:"-"`

    // 个人资料
    Nickname  string `gorm:"size:50" json:"nickname"`
    Avatar    string `gorm:"size:255" json:"avatar"`
    Bio       string `gorm:"size:500" json:"bio"`

    // 状态
    Status    int    `gorm:"default:1" json:"status"` // 1=active, 0=inactive
    Role      string `gorm:"size:20;default:user" json:"role"` // admin, user

    // 关联
    Posts     []Post    `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
    Comments  []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}
```

### 2.2 文章模型 (Post)

用于CRUD操作、关联关系示例。

```go
// Post 文章模型
type Post struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    // 内容
    Title     string `gorm:"size:200;not null" json:"title"`
    Slug      string `gorm:"size:200;uniqueIndex;not null" json:"slug"`
    Content   string `gorm:"type:text" json:"content"`
    Summary   string `gorm:"size:500" json:"summary"`

    // 元数据
    Status    string `gorm:"size:20;default:draft" json:"status"` // draft, published
    ViewCount int    `gorm:"default:0" json:"view_count"`

    // 关联
    AuthorID  uint      `gorm:"not null" json:"author_id"`
    Author    User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
    Comments  []Comment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
    Tags      []Tag     `gorm:"many2many:post_tags" json:"tags,omitempty"`
}
```

### 2.3 评论模型 (Comment)

用于演示嵌套关联关系。

```go
// Comment 评论模型
type Comment struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    // 内容
    Content   string `gorm:"type:text;not null" json:"content"`

    // 关联
    PostID    uint `gorm:"not null" json:"post_id"`
    Post      Post `gorm:"foreignKey:PostID" json:"-"`
    UserID    uint `gorm:"not null" json:"user_id"`
    User      User `gorm:"foreignKey:UserID" json:"user,omitempty"`

    // 嵌套回复（自引用）
    ParentID  *uint     `json:"parent_id,omitempty"`
    Parent    *Comment  `gorm:"foreignKey:ParentID" json:"-"`
    Replies   []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}
```

### 2.4 标签模型 (Tag)

用于演示多对多关系。

```go
// Tag 标签模型
type Tag struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`

    // 内容
    Name      string `gorm:"size:50;uniqueIndex;not null" json:"name"`
    Slug      string `gorm:"size:50;uniqueIndex;not null" json:"slug"`

    // 关联
    Posts     []Post `gorm:"many2many:post_tags" json:"posts,omitempty"`
}
```

### 2.5 Todo模型 (Todo)

用于基础CRUD教程的简单模型。

```go
// Todo 待办事项模型（简化版，用于入门教程）
type Todo struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`

    Title     string `gorm:"size:200;not null" json:"title"`
    Completed bool   `gorm:"default:false" json:"completed"`
    DueDate   *time.Time `json:"due_date,omitempty"`
}
```

### 2.6 会话模型 (Session)

用于Session认证示例。

```go
// Session 会话模型
type Session struct {
    ID        string    `gorm:"primaryKey;size:64" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    User      User      `gorm:"foreignKey:UserID" json:"-"`
    Data      string    `gorm:"type:text" json:"-"` // JSON encoded session data
    ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 2.7 OAuth账户模型 (OAuthAccount)

用于OAuth2第三方登录示例。

```go
// OAuthAccount OAuth关联账户
type OAuthAccount struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    CreatedAt  time.Time `json:"created_at"`

    UserID     uint   `gorm:"not null;index" json:"user_id"`
    User       User   `gorm:"foreignKey:UserID" json:"-"`
    Provider   string `gorm:"size:20;not null" json:"provider"` // github, google
    ProviderID string `gorm:"size:100;not null" json:"provider_id"`

    // 联合唯一索引
    // gorm:"uniqueIndex:idx_provider_id"
}
```

## Part 3: 缓存数据结构 (Cache Data Structures)

### 3.1 Redis缓存键设计

```
# 用户缓存
user:{id}                    -> JSON(User)          TTL: 1h
user:email:{email}           -> user_id             TTL: 1h
user:session:{session_id}    -> JSON(Session)       TTL: 24h

# 文章缓存
post:{id}                    -> JSON(Post)          TTL: 10m
post:list:page:{page}        -> JSON([]Post)        TTL: 5m
post:hot                     -> ZSET(id, score)     TTL: 1h

# 标签缓存
tag:all                      -> JSON([]Tag)         TTL: 1h
tag:{id}:posts               -> JSON([]Post)        TTL: 5m

# 计数器
post:{id}:views              -> INT                 持久化
user:{id}:post_count         -> INT                 TTL: 1h

# 分布式锁
lock:post:update:{id}        -> owner_id            TTL: 30s
```

### 3.2 本地缓存结构

```go
// CacheItem 本地缓存项
type CacheItem struct {
    Data      interface{}
    ExpiredAt time.Time
}

// 推荐使用场景
// - 配置信息
// - 热点数据（如热门标签）
// - 短期高频访问数据
```

## Part 4: 消息队列数据结构 (Message Queue Structures)

### 4.1 任务消息格式

```go
// TaskMessage 异步任务消息
type TaskMessage struct {
    ID        string                 `json:"id"`         // UUID
    Type      string                 `json:"type"`       // email, notification, etc.
    Payload   map[string]interface{} `json:"payload"`
    Priority  int                    `json:"priority"`   // 0=low, 1=normal, 2=high
    CreatedAt time.Time              `json:"created_at"`
    RetryCount int                   `json:"retry_count"`
    MaxRetry   int                   `json:"max_retry"`
}
```

### 4.2 邮件任务示例

```go
// EmailTask 邮件发送任务
type EmailTask struct {
    To       string `json:"to"`
    Subject  string `json:"subject"`
    Template string `json:"template"`
    Data     map[string]interface{} `json:"data"`
}
```

### 4.3 队列定义

```
# RabbitMQ Exchange和Queue设计

Exchange: tasks.direct (direct类型)
├── Queue: tasks.email      <- routing_key: email
├── Queue: tasks.notify     <- routing_key: notification
└── Queue: tasks.export     <- routing_key: export

Exchange: tasks.delayed (x-delayed-message插件)
└── Queue: tasks.scheduled  <- 延迟任务

Exchange: tasks.dlx (死信交换机)
└── Queue: tasks.failed     <- 失败任务
```

## Part 5: API响应数据结构 (API Response Structures)

### 5.1 统一响应格式

```go
// Response 统一API响应
type Response struct {
    Code    int         `json:"code"`    // 业务状态码
    Message string      `json:"message"` // 提示信息
    Data    interface{} `json:"data"`    // 响应数据
}

// PageResponse 分页响应
type PageResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    PageMeta    `json:"meta"`
}

// PageMeta 分页元数据
type PageMeta struct {
    Page      int   `json:"page"`
    PageSize  int   `json:"page_size"`
    Total     int64 `json:"total"`
    TotalPage int   `json:"total_page"`
}
```

### 5.2 错误响应格式

```go
// ErrorResponse 错误响应
type ErrorResponse struct {
    Code    int               `json:"code"`
    Message string            `json:"message"`
    Errors  []ValidationError `json:"errors,omitempty"`
}

// ValidationError 字段验证错误
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

## Part 6: ER图

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    User     │     │    Post     │     │    Tag      │
├─────────────┤     ├─────────────┤     ├─────────────┤
│ id          │──┐  │ id          │──┐  │ id          │
│ username    │  │  │ title       │  │  │ name        │
│ email       │  │  │ slug        │  │  │ slug        │
│ password    │  │  │ content     │  │  └──────┬──────┘
│ nickname    │  │  │ status      │  │         │
│ role        │  │  │ author_id   │◄─┘         │ M:N
└──────┬──────┘  │  └──────┬──────┘            │
       │         │         │                   │
       │ 1:N     │         │ 1:N      ┌────────┴────────┐
       │         │         │          │   post_tags     │
       ▼         │         ▼          ├─────────────────┤
┌─────────────┐  │  ┌─────────────┐   │ post_id         │
│   Comment   │  │  │   Comment   │   │ tag_id          │
├─────────────┤  │  ├─────────────┤   └─────────────────┘
│ id          │  │  │ (same)      │
│ content     │  │  └─────────────┘
│ post_id     │◄─┤
│ user_id     │◄─┘
│ parent_id   │◄──┐ (self-ref)
└─────────────┘   │
                  │

┌─────────────┐     ┌─────────────────┐
│   Session   │     │  OAuthAccount   │
├─────────────┤     ├─────────────────┤
│ id          │     │ id              │
│ user_id     │◄────│ user_id         │
│ data        │     │ provider        │
│ expires_at  │     │ provider_id     │
└─────────────┘     └─────────────────┘
```

## 模型使用场景映射

| 章节 | 使用模型 | 用途 |
|------|----------|------|
| 02-syntax-basics | - | 基本类型演示 |
| 03-struct-interface | User (简化) | 结构体定义 |
| 06-web-api | Todo | RESTful CRUD |
| 07-authentication | User, Session, OAuthAccount | 认证系统 |
| 08-database-sql | User, Post | 原生SQL操作 |
| 09-gorm | User, Post, Comment, Tag | ORM完整示例 |
| 10-caching | Post (缓存) | Redis缓存 |
| 11-message-queue | TaskMessage, EmailTask | 异步任务 |
