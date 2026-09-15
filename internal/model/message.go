package model

import (
	"time"
)

// Message maps the messages table.
type Message struct {
	ID          int64      `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	Title       string     `gorm:"column:title;type:varchar(255)"`
	FromUserID  *int64     `gorm:"column:from_user_id;type:int"`
	TargetID    int64      `gorm:"column:target_id;type:int"`
	TargetType  int        `gorm:"column:target_type;type:int"`
	Content     string     `gorm:"column:content;type:text"`
	ContentType string     `gorm:"column:content_type;type:varchar(255);default:'md'"`
	AvatarURL   string     `gorm:"column:avatar_url;type:varchar(255)"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:datetime;not null"`
	AlreadyRead int        `gorm:"column:already_read;type:int;default:0"`
	PageURL     string     `gorm:"column:page_url;type:varchar(255)"`
	SubTitle    string     `gorm:"column:sub_title;type:text"`
	Date        *time.Time `gorm:"column:date;type:datetime"`
}

func (Message) TableName() string { return "messages" }
