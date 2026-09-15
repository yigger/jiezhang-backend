package model

import (
	"time"
)

// AssetLog maps the asset_logs table.
type AssetLog struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID        *int64    `gorm:"column:user_id;type:int"`
	Type          *int      `gorm:"column:type;type:int"`
	From          *int      `gorm:"column:from;type:int"`
	To            *int      `gorm:"column:to;type:int"`
	Amount        *float64  `gorm:"column:amount;type:decimal(12,2)"`
	Residue       *float64  `gorm:"column:residue;type:decimal(12,2)"`
	Description   *string   `gorm:"column:description;type:text"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
	AccountBookID *int64    `gorm:"column:account_book_id;type:int"`
}

func (AssetLog) TableName() string { return "asset_logs" }
