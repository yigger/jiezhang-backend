package model

import (
	"time"
)

// AccountBookCollaborator maps the account_book_collaborators table.
type AccountBookCollaborator struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID int64     `gorm:"column:account_book_id;type:int;uniqueIndex:idx_account_book_collaborators_account_book_id_user_id,priority:1"`
	UserID        int64     `gorm:"column:user_id;type:int;uniqueIndex:idx_account_book_collaborators_account_book_id_user_id,priority:2"`
	Role          string    `gorm:"column:role;type:varchar(255);default:'member'"`
	Remark        string    `gorm:"column:remark;type:varchar(255)"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (AccountBookCollaborator) TableName() string { return "account_book_collaborators" }
