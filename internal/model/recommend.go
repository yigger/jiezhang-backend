package model

import (
	"time"
)

// Recommend maps the recommends table.
type Recommend struct {
	ID          int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID      *int64    `gorm:"column:user_id;type:int"`
	ShareTicket *string   `gorm:"column:share_ticket;type:varchar(255)"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (Recommend) TableName() string { return "recommends" }
