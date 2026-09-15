package model

import (
	"time"
)

// AssetSnapshot maps the asset_snapshots table.
type AssetSnapshot struct {
	ID          int64      `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID      *int64     `gorm:"column:user_id;type:int"`
	StatementID *int64     `gorm:"column:statement_id;type:int"`
	AssetID     *int64     `gorm:"column:asset_id;type:int"`
	Year        *int       `gorm:"column:year;type:int"`
	Month       *int       `gorm:"column:month;type:int"`
	Day         *int       `gorm:"column:day;type:int"`
	Date        *time.Time `gorm:"column:date;type:date"`
	BeforeValue *float64   `gorm:"column:before_value;type:decimal(12,2)"`
	AfterValue  *float64   `gorm:"column:after_value;type:decimal(12,2)"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:datetime;not null"`
}

func (AssetSnapshot) TableName() string { return "asset_snapshots" }
