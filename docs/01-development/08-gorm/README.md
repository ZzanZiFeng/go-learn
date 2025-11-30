# 第八章：GORM 框架 (GORM ORM)

## 章节概述

GORM 是 Go 语言最流行的 ORM 框架，提供强大的数据库操作功能。

## 学习目标

完成本章后，你将能够：

- 使用 GORM 定义模型和关联关系
- 执行 CRUD 操作和复杂查询
- 正确处理关联关系（1:1, 1:N, M:N）
- 使用钩子和事务
- 了解 GORM 最佳实践

## 与 JavaScript 对比

### TypeORM (TypeScript)

```typescript
import { Entity, PrimaryGeneratedColumn, Column, ManyToOne } from "typeorm"

@Entity()
export class User {
    @PrimaryGeneratedColumn()
    id: number

    @Column()
    name: string

    @OneToMany(() => Post, post => post.author)
    posts: Post[]
}

// 查询
const user = await userRepository.findOne({ where: { id: 1 } })
const users = await userRepository.find({ where: { status: 'active' } })
```

### GORM (Go)

```go
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string
    Posts []Post `gorm:"foreignKey:AuthorID"`
}

// 查询
var user User
db.First(&user, 1)

var users []User
db.Where("status = ?", "active").Find(&users)
```

## 章节内容

| 文档 | 主题 | 描述 |
|------|------|------|
| [01-gorm-intro.md](./01-gorm-intro.md) | GORM 入门 | 安装和基本配置 |
| [02-model-definition.md](./02-model-definition.md) | 模型定义 | 结构体标签和约定 |
| [03-crud.md](./03-crud.md) | CRUD 操作 | 增删改查 |
| [04-query-builder.md](./04-query-builder.md) | 查询构建 | Where、Order、Limit |
| [05-associations.md](./05-associations.md) | 关联关系 | 1:1, 1:N, M:N |
| [06-preload.md](./06-preload.md) | 预加载 | Eager Loading |
| [07-hooks.md](./07-hooks.md) | 钩子 | 生命周期回调 |
| [08-transactions.md](./08-transactions.md) | 事务 | GORM 事务处理 |
| [09-raw-sql.md](./09-raw-sql.md) | 原生 SQL | Raw SQL 和 SQL Builder |
| [10-best-practices.md](./10-best-practices.md) | 最佳实践 | 性能和安全建议 |
| [exercises.md](./exercises.md) | 练习 | 实践练习 |

## 快速开始

### 安装

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```

### 基本示例

```go
package main

import (
    "fmt"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string `gorm:"size:100;not null"`
    Email     string `gorm:"uniqueIndex;size:255"`
    Age       int
    CreatedAt time.Time
    UpdatedAt time.Time
}

func main() {
    dsn := "host=localhost user=user password=password dbname=testdb port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // 自动迁移
    db.AutoMigrate(&User{})

    // 创建
    user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
    db.Create(&user)
    fmt.Printf("Created user ID: %d\n", user.ID)

    // 查询
    var foundUser User
    db.First(&foundUser, user.ID)
    fmt.Printf("Found: %s <%s>\n", foundUser.Name, foundUser.Email)

    // 更新
    db.Model(&foundUser).Update("Age", 26)

    // 删除
    db.Delete(&foundUser)
}
```

## GORM 核心特性

### 自动迁移

```go
// 自动创建/更新表结构
db.AutoMigrate(&User{}, &Post{}, &Comment{})
```

### 关联关系

```go
// 一对多
type User struct {
    ID    uint
    Posts []Post `gorm:"foreignKey:AuthorID"`
}

type Post struct {
    ID       uint
    AuthorID uint
    Author   User
}

// 多对多
type Post struct {
    ID   uint
    Tags []Tag `gorm:"many2many:post_tags;"`
}
```

### 链式查询

```go
db.Where("age > ?", 18).
   Order("created_at DESC").
   Limit(10).
   Find(&users)
```

### 钩子

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.UUID = uuid.New().String()
    return nil
}
```

## GORM vs database/sql

| 特性 | database/sql | GORM |
|------|-------------|------|
| 学习曲线 | 低 | 中 |
| 代码量 | 多 | 少 |
| 灵活性 | 高 | 中 |
| 性能 | 最优 | 略有开销 |
| 关联关系 | 手动 | 自动 |
| 迁移 | 外部工具 | 内置 |
| 适用场景 | 复杂查询、高性能 | 快速开发、CRUD |

## 最佳实践

| 实践 | 说明 |
|------|------|
| 定义约束 | 使用 struct tag 明确约束 |
| 使用软删除 | 继承 gorm.DeletedAt |
| 合理预加载 | 避免 N+1 查询 |
| 使用事务 | 多操作放入事务 |
| 监控查询 | 开启日志或使用 Logger |
| 复杂查询用原生 SQL | 避免过度依赖 ORM |

## 常见陷阱

### 1. N+1 查询

```go
// ❌ N+1 问题
var users []User
db.Find(&users)
for _, user := range users {
    db.Model(&user).Association("Posts").Find(&user.Posts)  // N 次查询
}

// ✅ 使用 Preload
db.Preload("Posts").Find(&users)  // 2 次查询
```

### 2. 忘记检查错误

```go
// ❌ 忽略错误
db.First(&user, 1)

// ✅ 检查错误
if err := db.First(&user, 1).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 处理未找到
    }
    return err
}
```

### 3. 更新零值

```go
// ❌ 不会更新为 0
db.Model(&user).Update("Age", 0)

// ✅ 使用 Select 强制更新
db.Model(&user).Select("Age").Update("Age", 0)
```

**下一节**：[GORM 入门](./01-gorm-intro.md) - 开始学习 GORM 基础
