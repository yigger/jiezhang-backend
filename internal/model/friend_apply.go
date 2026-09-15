package model

import (
	"time"
)

// FriendApply maps the friend_applies table.
type FriendApply struct {
	ID        int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	SourceID  *int64    `gorm:"column:source_id;type:int"`
	TargetID  *int64    `gorm:"column:target_id;type:int"`
	Remark    *string   `gorm:"column:remark;type:varchar(255)"`
	Status    *int      `gorm:"column:status;type:int;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (FriendApply) TableName() string { return "friend_applies" }
