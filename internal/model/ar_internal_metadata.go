package model

import (
	"time"
)

// ARInternalMetadata maps the ar_internal_metadata table.
type ARInternalMetadata struct {
	Key       string    `gorm:"column:key;type:varchar(255);not null;primaryKey;autoIncrement:false"`
	Value     *string   `gorm:"column:value;type:varchar(255)"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (ARInternalMetadata) TableName() string { return "ar_internal_metadata" }
