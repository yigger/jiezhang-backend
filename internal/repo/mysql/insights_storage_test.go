package mysql

import (
	"context"
	"errors"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"testing"
)

func TestInsightsStorageBookScopedLists(t *testing.T) {
	db, mock := mockStatementDB(t)
	storage := NewInsightsStorage(db.db)
	mock.ExpectQuery("SELECT .*insight_projects.*WHERE account_book_id = \\?.*").WithArgs(int64(8)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_book_id", "name", "budget_cents"}).AddRow(7, 8, "旅行", 12345))
	rows, err := storage.ListProjects(context.Background(), 8)
	if err != nil || len(rows) != 1 || rows[0].AccountBookID != 8 || rows[0].BudgetCents != 12345 {
		t.Fatalf("projects: %+v %v", rows, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestInsightMetadataTransactionRollsBackFailedUseCase(t *testing.T) {
	db, mock := mockStatementDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*statements.*WHERE account_book_id = \\? AND id = \\?.*FOR UPDATE").WithArgs(int64(8), int64(5), 1).WillReturnRows(sqlmock.NewRows([]string{"id", "account_book_id", "user_id", "amount"}).AddRow(5, 8, 2, 12.34))
	mock.ExpectRollback()
	failure := errors.New("invalid allocation")
	err := NewInsightsStorage(db.db).Transact(context.Background(), func(tx repo.InsightsStorageTx) error {
		row, e := tx.LockStatement(8, 5)
		if e != nil {
			return e
		}
		if row.ID != 5 {
			return errors.New("wrong statement")
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
