package model

import (
	"time"
)

// Category maps the categories table.
type Category struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID        int64     `gorm:"column:user_id;type:int"`
	Name          string    `gorm:"column:name;type:varchar(255)"`
	ParentID      int64     `gorm:"column:parent_id;type:int;not null;default:0;index:index_categories_on_parent_id,priority:1"`
	Order         int       `gorm:"column:order;type:int;not null;default:0;index:index_categories_on_order,priority:1"`
	IconPath      string    `gorm:"column:icon_path;type:varchar(255)"`
	Color         *string   `gorm:"column:color;type:varchar(255)"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
	Type          string    `gorm:"column:type;type:varchar(255);index:index_categories_on_type,priority:1"`
	Lock          *int      `gorm:"column:lock;type:int;default:0"`
	Budget        float64   `gorm:"column:budget;type:decimal(12,2);default:0.00"`
	Frequent      int       `gorm:"column:frequent;type:int;default:0"`
	IsMess        *int      `gorm:"column:is_mess;type:int;default:0"`
	AccountBookID int64     `gorm:"column:account_book_id;type:int"`
	IsSystem      *bool     `gorm:"column:is_system;type:tinyint(1);default:0"`
	SpecialType   string    `gorm:"column:special_type;type:varchar(255)"`
}

func (Category) TableName() string { return "categories" }
