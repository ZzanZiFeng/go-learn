package models

import (
	"time"

	"gorm.io/gorm"
)

// Todo 待办事项模型
type Todo struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	UserID      uint           `json:"user_id" gorm:"index;not null"`
	Title       string         `json:"title" gorm:"size:200;not null"`
	Description string         `json:"description" gorm:"size:1000"`
	Completed   bool           `json:"completed" gorm:"default:false"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Todo) TableName() string {
	return "todos"
}
