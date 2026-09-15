package model

import (
	"time"
)

// UserAsset maps the user_assets table.
type UserAsset struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:name;type:varchar(255)"`
	Path          string    `gorm:"column:path;type:varchar(255)"`
	Type          string    `gorm:"column:type;type:varchar(255)"`
	ImageableType string    `gorm:"column:imageable_type;type:varchar(255);index:index_user_assets_on_imageable_type_and_imageable_id,priority:1,length:191"`
	ImageableID   int64     `gorm:"column:imageable_id;type:bigint;index:index_user_assets_on_imageable_type_and_imageable_id,priority:2"`
	Score         int       `gorm:"column:score;type:int;default:0"`
	Locked        int       `gorm:"column:locked;type:int;default:0"`
	System        int       `gorm:"column:system;type:int;default:0"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (UserAsset) TableName() string { return "user_assets" }
