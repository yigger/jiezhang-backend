package model

import (
	"encoding/json"
	"time"
)

// ErrorLog maps the error_logs table.
type ErrorLog struct {
	ID        int64           `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID    *int64          `gorm:"column:user_id;type:int"`
	Message   json.RawMessage `gorm:"column:message;type:json"`
	CreatedAt time.Time       `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time       `gorm:"column:updated_at;type:datetime;not null"`
}

func (ErrorLog) TableName() string { return "error_logs" }
