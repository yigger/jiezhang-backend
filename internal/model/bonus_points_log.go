package model

import (
	"time"
)

// BonusPointsLog maps the bonus_points_logs table.
type BonusPointsLog struct {
	ID        int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	Year      *int      `gorm:"column:year;type:int"`
	Month     *int      `gorm:"column:month;type:int"`
	Day       *int      `gorm:"column:day;type:int"`
	Type      *int      `gorm:"column:type;type:int;default:0"`
	Point     *int      `gorm:"column:point;type:int;default:0"`
	UserID    *int64    `gorm:"column:user_id;type:int"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (BonusPointsLog) TableName() string { return "bonus_points_logs" }
