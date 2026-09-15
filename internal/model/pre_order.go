package model

import (
	"time"
)

// PreOrder maps the pre_orders table.
type PreOrder struct {
	ID        int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	OwnerID   *int64    `gorm:"column:owner_id;type:int"`
	CreatorID *int64    `gorm:"column:creator_id;type:int"`
	Name      *string   `gorm:"column:name;type:varchar(255)"`
	Amount    *float64  `gorm:"column:amount;type:decimal(12,2)"`
	State     *string   `gorm:"column:state;type:varchar(255);default:'pending'"`
	Remark    *string   `gorm:"column:remark;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`
	Address   *string   `gorm:"column:address;type:varchar(255)"`
}

func (PreOrder) TableName() string { return "pre_orders" }
