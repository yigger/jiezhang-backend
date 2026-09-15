package model

import (
	"time"
)

// Statement maps the statements table.
type Statement struct {
	ID            int64      `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	UserID        int64      `gorm:"column:user_id;type:int;not null;index:index_statements_on_user_id_and_asset_id,priority:1;index:index_statements_on_user_id_and_category_id,priority:1;index:index_statements_on_user_id_and_type,priority:1"`
	CategoryID    int64      `gorm:"column:category_id;type:int;index:index_statements_on_user_id_and_category_id,priority:2"`
	AssetID       int64      `gorm:"column:asset_id;type:int;not null;index:index_statements_on_user_id_and_asset_id,priority:2"`
	Amount        float64    `gorm:"column:amount;type:decimal(12,2)"`
	Refund        float64    `gorm:"column:refund;type:decimal(12,2);default:0.00"`
	Type          string     `gorm:"column:type;type:varchar(255);not null;index:index_statements_on_type,priority:1;index:index_statements_on_user_id_and_type,priority:2"`
	Description   string     `gorm:"column:description;type:text"`
	Year          int        `gorm:"column:year;type:int;index:index_statements_on_year_and_month_and_day_and_time,priority:1"`
	Month         int        `gorm:"column:month;type:int;index:index_statements_on_year_and_month_and_day_and_time,priority:2"`
	Day           int        `gorm:"column:day;type:int;index:index_statements_on_year_and_month_and_day_and_time,priority:3"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;type:datetime;not null"`
	TimeText      string     `gorm:"column:time;type:time;index:index_statements_on_year_and_month_and_day_and_time,priority:4"`
	Residue       float64    `gorm:"column:residue;type:decimal(12,2);default:0.00"`
	Location      string     `gorm:"column:location;type:text"`
	Nation        string     `gorm:"column:nation;type:varchar(255)"`
	Province      string     `gorm:"column:province;type:varchar(255)"`
	City          string     `gorm:"column:city;type:varchar(255)"`
	District      string     `gorm:"column:district;type:varchar(255)"`
	Street        string     `gorm:"column:street;type:varchar(255)"`
	TargetAssetID *int64     `gorm:"column:target_asset_id;type:int"`
	AccountBookID int64      `gorm:"column:account_book_id;type:int"`
	Ref           *string    `gorm:"column:ref;type:varchar(255)"`
	RefAt         *time.Time `gorm:"column:ref_at;type:datetime"`
	Mood          string     `gorm:"column:mood;type:varchar(255)"`
	PayeeID       *int64     `gorm:"column:payee_id;type:int"`
	TargetObject  string     `gorm:"column:target_object;type:varchar(255)"`
}

func (Statement) TableName() string { return "statements" }
