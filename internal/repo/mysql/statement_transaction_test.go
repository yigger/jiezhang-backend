package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func mockStatementDB(t *testing.T) (*StatementRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	gdb, err := gorm.Open(mysqldriver.New(mysqldriver.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return NewStatementRepository(gdb), mock
}
func TestTransferCommitAndRollback(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "commit"
		if fail {
			name = "rollback_target_failure"
		}
		t.Run(name, func(t *testing.T) {
			repository, mock := mockStatementDB(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*amount.* FROM `assets`.*FOR UPDATE").WithArgs(int64(10), int64(3), 1).WillReturnRows(sqlmock.NewRows([]string{"amount"}).AddRow(100))
			mock.ExpectExec("INSERT INTO `statements`").WillReturnResult(sqlmock.NewResult(8, 1))
			mock.ExpectExec("UPDATE `assets`").WithArgs(float64(-25), int64(10), int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
			target := mock.ExpectExec("UPDATE `assets`").WithArgs(float64(25), int64(11), int64(3))
			failure := errors.New("target failed")
			if fail {
				target.WillReturnError(failure)
				mock.ExpectRollback()
			} else {
				target.WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
			targetID := int64(11)
			err := repository.WithinTransaction(context.Background(), func(tx repo.Mutation) error {
				_, err := tx.Create(context.Background(), tablemodel.Statement{UserID: 2, AccountBookID: 3, AssetID: 10, TargetAssetID: &targetID, CategoryID: 4, Type: "transfer", Amount: 25, CreatedAt: time.Now()}, repo.BalanceEffect{Source: -25, Target: 25, HasTarget: true})
				return err
			})
			if fail && !errors.Is(err, failure) {
				t.Fatalf("error=%v", err)
			}
			if !fail && err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestDeleteRollsBackBalanceOnFailure(t *testing.T) {
	repository, mock := mockStatementDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `statements`.*FOR UPDATE").WithArgs(int64(8), int64(3), 1).WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "type", "amount"}).AddRow(8, 10, "expend", 25))
	mock.ExpectExec("UPDATE `assets`").WithArgs(float64(25), int64(10), int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	failure := errors.New("delete failed")
	mock.ExpectExec("DELETE FROM `statements`").WithArgs(int64(8), int64(3)).WillReturnError(failure)
	mock.ExpectRollback()
	err := repository.WithinTransaction(context.Background(), func(tx repo.Mutation) error {
		return tx.DeleteByID(context.Background(), 8, 3, repo.BalanceEffect{Source: -25})
	})
	if !errors.Is(err, failure) {
		t.Fatalf("error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestLockCurrentScopesByBook(t *testing.T) {
	repository, mock := mockStatementDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `statements`.*FOR UPDATE").WithArgs(int64(8), int64(3), 1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	err := repository.WithinTransaction(context.Background(), func(tx repo.Mutation) error { _, e := tx.LockCurrent(context.Background(), 8, 3); return e })
	if !errors.Is(err, repo.ErrStatementNotFound) {
		t.Fatalf("error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
