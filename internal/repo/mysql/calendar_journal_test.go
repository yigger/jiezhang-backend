package mysql

import (
	"context"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/yigger/jiezhang-backend/internal/model"
	"regexp"
	"testing"
	"time"
)

func TestCalendarJournalListIsPersonalAndBookScoped(t *testing.T) {
	db, mock := mockStatementDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `calendar_journals` WHERE account_book_id = ? AND user_id = ? AND date >= ? AND date < ? ORDER BY date ASC")).WithArgs(int64(8), int64(2), "2026-10-01", "2026-11-01").WillReturnRows(sqlmock.NewRows([]string{"date", "note"}).AddRow("2026-10-08", "日记"))
	rows, err := NewCalendarJournalRepository(db.db).List(context.Background(), 8, 2, "2026-10-01", "2026-11-01")
	if err != nil || len(rows) != 1 || rows[0].Note != "日记" {
		t.Fatalf("%v %+v", err, rows)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestCalendarJournalUpsertWritesFalseAndEmptyFields(t *testing.T) {
	db, mock := mockStatementDB(t)
	now := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `calendar_journals`.*ON DUPLICATE KEY UPDATE `mood`=VALUES\\(`mood`\\),`note`=VALUES\\(`note`\\),`zero_expense`=VALUES\\(`zero_expense`\\),`updated_at`=VALUES\\(`updated_at`\\)").WithArgs(int64(8), int64(2), "2026-10-08", "", "", false, now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := NewCalendarJournalRepository(db.db).Save(context.Background(), &model.CalendarJournal{AccountBookID: 8, UserID: 2, Date: "2026-10-08", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestCalendarJournalStampRevocationDoesNotOverwriteText(t *testing.T) {
	db, mock := mockStatementDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `calendar_journals` SET `zero_expense`=\\?,`updated_at`=\\? WHERE account_book_id = \\? AND user_id = \\? AND date IN \\(\\?\\) AND zero_expense = \\?").WithArgs(false, sqlmock.AnyArg(), int64(8), int64(2), "2026-10-08", true).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err := NewCalendarJournalRepository(db.db).RevokeZeroExpense(context.Background(), 8, 2, []string{"2026-10-08"})
	if err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCalendarJournalExpenseLookupUsesBookAndLocalDateAndOnlyExpenseType(t *testing.T) {
	db, mock := mockStatementDB(t)
	mock.ExpectQuery("SELECT DISTINCT CONCAT.*FROM `statements` WHERE account_book_id = \\? AND type = 'expend'.*").WithArgs(int64(8), "2026-10-01", "2026-11-01").WillReturnRows(sqlmock.NewRows([]string{"date"}).AddRow("2026-10-08"))
	rows, err := NewCalendarJournalRepository(db.db).ExpenseDates(context.Background(), 8, "2026-10-01", "2026-11-01")
	if err != nil || len(rows) != 1 || rows[0] != "2026-10-08" {
		t.Fatalf("%v %v", rows, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
