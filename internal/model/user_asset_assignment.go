package model

import (
	"time"
)

// UserAssetAssignment maps the users_assets table.
type UserAssetAssignment struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID        *int64    `gorm:"column:user_id;type:int"`
	AssetID       *string   `gorm:"column:asset_id;type:varchar(255)"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
	AccountBookID *int64    `gorm:"column:account_book_id;type:int"`
	AssignType    *int      `gorm:"column:assign_type;type:int;default:1"`
}

func (UserAssetAssignment) TableName() string { return "users_assets" }
