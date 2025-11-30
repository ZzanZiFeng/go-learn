// examples/gorm/associations/main.go
// GORM 关联关系示例
package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User 用户（Has Many Posts, Has One Profile）
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:255;uniqueIndex"`
	Profile   *Profile  // Has One
	Posts     []Post    `gorm:"foreignKey:AuthorID"` // Has Many
	CreatedAt time.Time
}

// Profile 用户资料（Belongs To User）
type Profile struct {
	ID       uint   `gorm:"primaryKey"`
	UserID   uint   `gorm:"uniqueIndex"` // 外键
	Bio      string `gorm:"type:text"`
	Avatar   string `gorm:"size:500"`
	Location string `gorm:"size:100"`
}

// Post 文章（Belongs To User, Many To Many Tags）
type Post struct {
	ID        uint           `gorm:"primaryKey"`
	Title     string         `gorm:"size:255;not null"`
	Content   string         `gorm:"type:text"`
	AuthorID  uint           `gorm:"index"` // 外键
	Author    *User          // Belongs To
	Tags      []Tag          `gorm:"many2many:post_tags;"` // Many To Many
	Comments  []Comment      `gorm:"foreignKey:PostID"`    // Has Many
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Tag 标签（Many To Many Posts）
type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:50;uniqueIndex"`
	Posts []Post `gorm:"many2many:post_tags;"`
}

// Comment 评论（Belongs To Post, Belongs To User）
type Comment struct {
	ID        uint      `gorm:"primaryKey"`
	Content   string    `gorm:"type:text;not null"`
	PostID    uint      `gorm:"index"`
	Post      *Post     // Belongs To
	UserID    uint      `gorm:"index"`
	User      *User     // Belongs To
	CreatedAt time.Time
}

func main() {
	// 连接 SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	// 自动迁移
	db.AutoMigrate(&User{}, &Profile{}, &Post{}, &Tag{}, &Comment{})

	fmt.Println("=== Create with Associations ===")
	createWithAssociations(db)

	fmt.Println("\n=== Preload Associations ===")
	preloadDemo(db)

	fmt.Println("\n=== Association Operations ===")
	associationOperations(db)

	fmt.Println("\n=== Nested Preload ===")
	nestedPreloadDemo(db)
}

func createWithAssociations(db *gorm.DB) {
	// 创建用户同时创建 Profile
	user := User{
		Name:  "Alice",
		Email: "alice@example.com",
		Profile: &Profile{
			Bio:      "Software Engineer",
			Avatar:   "avatar.jpg",
			Location: "San Francisco",
		},
	}
	db.Create(&user)
	fmt.Printf("Created user %s with profile\n", user.Name)

	// 创建标签
	tags := []Tag{
		{Name: "Go"},
		{Name: "Database"},
		{Name: "Tutorial"},
	}
	db.Create(&tags)
	fmt.Printf("Created %d tags\n", len(tags))

	// 创建带标签的文章
	post := Post{
		Title:    "Learning GORM",
		Content:  "GORM is a great ORM for Go...",
		AuthorID: user.ID,
		Tags:     []Tag{tags[0], tags[1]}, // Go, Database
	}
	db.Create(&post)
	fmt.Printf("Created post '%s' with %d tags\n", post.Title, len(post.Tags))

	// 创建评论
	comment := Comment{
		Content: "Great article!",
		PostID:  post.ID,
		UserID:  user.ID,
	}
	db.Create(&comment)
	fmt.Printf("Created comment on post\n")
}

func preloadDemo(db *gorm.DB) {
	// Preload Has One
	var user User
	db.Preload("Profile").First(&user)
	fmt.Printf("User: %s, Bio: %s\n", user.Name, user.Profile.Bio)

	// Preload Has Many
	db.Preload("Posts").First(&user)
	fmt.Printf("User: %s has %d posts\n", user.Name, len(user.Posts))

	// Preload Belongs To
	var post Post
	db.Preload("Author").First(&post)
	fmt.Printf("Post: %s by %s\n", post.Title, post.Author.Name)

	// Preload Many To Many
	db.Preload("Tags").First(&post)
	fmt.Printf("Post: %s has %d tags\n", post.Title, len(post.Tags))

	// 多个 Preload
	db.Preload("Author").Preload("Tags").Preload("Comments").First(&post)
	fmt.Printf("Post loaded with author, %d tags, %d comments\n",
		len(post.Tags), len(post.Comments))
}

func associationOperations(db *gorm.DB) {
	var user User
	db.First(&user)

	// 添加新文章
	newPost := Post{Title: "Another Post", Content: "Content..."}
	db.Model(&user).Association("Posts").Append(&newPost)
	fmt.Printf("Added post to user, new post ID: %d\n", newPost.ID)

	// 查询关联数量
	count := db.Model(&user).Association("Posts").Count()
	fmt.Printf("User has %d posts\n", count)

	var post Post
	db.First(&post)

	// 添加标签
	var tag Tag
	db.Where("name = ?", "Tutorial").First(&tag)
	db.Model(&post).Association("Tags").Append(&tag)
	fmt.Println("Added Tutorial tag to post")

	// 查询文章的所有标签
	var postTags []Tag
	db.Model(&post).Association("Tags").Find(&postTags)
	fmt.Printf("Post tags: ")
	for _, t := range postTags {
		fmt.Printf("%s ", t.Name)
	}
	fmt.Println()

	// 删除关联（不删除标签本身）
	db.Model(&post).Association("Tags").Delete(&tag)
	fmt.Println("Removed Tutorial tag from post")

	// 替换所有标签
	var goTag, dbTag Tag
	db.Where("name = ?", "Go").First(&goTag)
	db.Where("name = ?", "Database").First(&dbTag)
	db.Model(&post).Association("Tags").Replace(&goTag, &dbTag)
	fmt.Println("Replaced post tags")

	// 清除所有关联
	// db.Model(&post).Association("Tags").Clear()
}

func nestedPreloadDemo(db *gorm.DB) {
	// 嵌套预加载：User -> Posts -> Tags
	var user User
	db.Preload("Posts.Tags").First(&user)

	fmt.Printf("User: %s\n", user.Name)
	for _, post := range user.Posts {
		fmt.Printf("  Post: %s\n", post.Title)
		for _, tag := range post.Tags {
			fmt.Printf("    Tag: %s\n", tag.Name)
		}
	}

	// 条件预加载
	db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(5)
	}).Preload("Posts.Tags").First(&user)

	fmt.Printf("\nUser %s has %d recent posts\n", user.Name, len(user.Posts))

	// 嵌套预加载：Post -> Author -> Profile
	var post Post
	db.Preload("Author.Profile").Preload("Tags").Preload("Comments.User").First(&post)

	fmt.Printf("\nPost: %s\n", post.Title)
	fmt.Printf("  Author: %s (%s)\n", post.Author.Name, post.Author.Profile.Location)
	fmt.Printf("  Tags: %d\n", len(post.Tags))
	fmt.Printf("  Comments: %d\n", len(post.Comments))
}
