package mysql

import (
	"context"
	"github.com/yigger/jiezhang-backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CalendarJournalRepository struct{ db *gorm.DB }

func NewCalendarJournalRepository(db *gorm.DB) *CalendarJournalRepository {
	return &CalendarJournalRepository{db}
}
func (r *CalendarJournalRepository) CanAccess(ctx context.Context, book, user int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AccountBook{}).Where("id = ? AND (user_id = ? OR EXISTS (SELECT 1 FROM account_book_collaborators c WHERE c.account_book_id = account_books.id AND c.user_id = ?))", book, user, user).Count(&count).Error
	return count > 0, err
}
func (r *CalendarJournalRepository) List(ctx context.Context, book, user int64, start, end string) ([]model.CalendarJournal, error) {
	rows := []model.CalendarJournal{}
	err := r.db.WithContext(ctx).Where("account_book_id = ? AND user_id = ? AND date >= ? AND date < ?", book, user, start, end).Order("date ASC").Find(&rows).Error
	return rows, err
}
func (r *CalendarJournalRepository) ExpenseDates(ctx context.Context, book int64, start, end string) ([]string, error) {
	rows := []string{}
	err := r.db.WithContext(ctx).Model(&model.Statement{}).Distinct("CONCAT(LPAD(year,4,'0'),'-',LPAD(month,2,'0'),'-',LPAD(day,2,'0'))").Where("account_book_id = ? AND type = 'expend' AND CONCAT(LPAD(year,4,'0'),'-',LPAD(month,2,'0'),'-',LPAD(day,2,'0')) >= ? AND CONCAT(LPAD(year,4,'0'),'-',LPAD(month,2,'0'),'-',LPAD(day,2,'0')) < ?", book, start, end).Pluck("CONCAT(LPAD(year,4,'0'),'-',LPAD(month,2,'0'),'-',LPAD(day,2,'0'))", &rows).Error
	return rows, err
}
func (r *CalendarJournalRepository) Save(ctx context.Context, row *model.CalendarJournal) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "account_book_id"}, {Name: "user_id"}, {Name: "date"}}, DoUpdates: clause.AssignmentColumns([]string{"mood", "note", "zero_expense", "updated_at"})}).Create(row).Error
}
func (r *CalendarJournalRepository) RevokeZeroExpense(ctx context.Context, book, user int64, dates []string) error {
	if len(dates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.CalendarJournal{}).Where("account_book_id = ? AND user_id = ? AND date IN ? AND zero_expense = ?", book, user, dates, true).Update("zero_expense", false).Error
}
