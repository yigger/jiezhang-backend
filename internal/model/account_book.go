package model

import (
	"time"
)

// AccountBook maps the account_books table.
type AccountBook struct {
	ID          int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID      int64     `gorm:"column:user_id;type:int"`
	AccountType int       `gorm:"column:account_type;type:int;default:0"`
	Name        string    `gorm:"column:name;type:varchar(255)"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;not null"`
	Description string    `gorm:"column:description;type:varchar(255)"`
	Budget      float64   `gorm:"column:budget;type:decimal(12,2);default:0.00"`
}

func (AccountBook) TableName() string { return "account_books" }
