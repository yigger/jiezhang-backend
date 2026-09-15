package model

import (
	"time"
)

// OperateLog maps the operate_logs table.
type OperateLog struct {
	ID         int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID     *int64    `gorm:"column:user_id;type:int"`
	Year       *int      `gorm:"column:year;type:int"`
	Month      *int      `gorm:"column:month;type:int"`
	Day        *int      `gorm:"column:day;type:int"`
	PageRoute  *string   `gorm:"column:page_route;type:varchar(512)"`
	SessionKey *string   `gorm:"column:session_key;type:varchar(255)"`
	IP         *string   `gorm:"column:ip;type:varchar(255)"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (OperateLog) TableName() string { return "operate_logs" }
