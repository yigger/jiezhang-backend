package model

import (
	"time"
)

// Feedback maps the feedbacks table.
type Feedback struct {
	ID        int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id;type:int"`
	Content   string    `gorm:"column:content;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`
	Type      int       `gorm:"column:type;type:int;default:0"`
}

func (Feedback) TableName() string { return "feedbacks" }
