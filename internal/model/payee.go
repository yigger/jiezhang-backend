package model

import (
	"time"
)

// Payee maps the payees table.
type Payee struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement" format:"int64"`
	Name          string    `gorm:"column:name;type:varchar(255)"`
	AccountBookID int64     `gorm:"column:account_book_id;type:bigint;index:index_payees_on_account_book_id,priority:1" format:"int64"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
	UserID        int64     `gorm:"column:user_id;type:int" format:"int64"`
}

func (Payee) TableName() string { return "payees" }
