package model

import "time"

type CalendarJournal struct {
	ID            int64     `gorm:"column:id;type:bigint;not null;primaryKey;autoIncrement"`
	AccountBookID int64     `gorm:"column:account_book_id;type:bigint;not null;uniqueIndex:idx_calendar_journal_day,priority:1"`
	UserID        int64     `gorm:"column:user_id;type:bigint;not null;uniqueIndex:idx_calendar_journal_day,priority:2"`
	Date          string    `gorm:"column:date;type:varchar(10);not null;uniqueIndex:idx_calendar_journal_day,priority:3"`
	Mood          string    `gorm:"column:mood;type:varchar(16);not null;default:''"`
	Note          string    `gorm:"column:note;type:varchar(800);not null;default:''"`
	ZeroExpense   bool      `gorm:"column:zero_expense;type:tinyint(1);not null;default:0"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (CalendarJournal) TableName() string { return "calendar_journals" }
