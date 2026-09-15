package model

import (
	"time"
)

// Friend maps the friends table.
type Friend struct {
	ID          int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID      *int64    `gorm:"column:user_id;type:int"`
	Name        *string   `gorm:"column:name;type:varchar(255)"`
	TargetID    *int64    `gorm:"column:target_id;type:int"`
	Remark      *string   `gorm:"column:remark;type:varchar(255)"`
	RoleName    *string   `gorm:"column:role_name;type:varchar(255)"`
	Active      *int      `gorm:"column:active;type:int;default:0"`
	AlreadyBind *bool     `gorm:"column:already_bind;type:tinyint(1);default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (Friend) TableName() string { return "friends" }
