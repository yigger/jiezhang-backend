package model

import (
	"time"
)

// Asset maps the assets table.
type Asset struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:name;type:varchar(255)"`
	Amount        float64   `gorm:"column:amount;type:decimal(12,2);not null;default:0.00;index:index_assets_on_amount,priority:1"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
	ParentID      int64     `gorm:"column:parent_id;type:int;default:0;index:index_assets_on_parent_id,priority:1"`
	Type          string    `gorm:"column:type;type:varchar(255);default:'deposit';index:index_assets_on_type,priority:1"`
	Lock          *int      `gorm:"column:lock;type:int;default:0"`
	IconPath      string    `gorm:"column:icon_path;type:varchar(255)"`
	Remark        string    `gorm:"column:remark;type:text"`
	CreatorID     int64     `gorm:"column:creator_id;type:int"`
	Frequent      int       `gorm:"column:frequent;type:int;default:0"`
	Order         int       `gorm:"column:order;type:int;default:0;index:index_assets_on_order,priority:1"`
	AccountBookID int64     `gorm:"column:account_book_id;type:int"`
}

func (Asset) TableName() string { return "assets" }
