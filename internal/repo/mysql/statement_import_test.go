package mysql

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

func expectImportCatalog(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT .* FROM `account_books`.*FOR UPDATE").WithArgs(int64(2), 1).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(2, 1))
	mock.ExpectQuery("SELECT .* FROM `account_book_collaborators`").WithArgs(int64(2), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT .* FROM `categories`").WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT .* FROM `assets`").WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT .*`import_key`.*`import_fingerprint`.* FROM `statements`.*import_key IS NOT NULL").WithArgs(int64(2), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id", "account_book_id", "user_id", "import_key", "import_fingerprint"}).AddRow(30, 2, 1, "original-key", "original-fingerprint"))
}

func TestImportMetadataAndNewNodesRollBackTogether(t *testing.T) {
	repository, mock := mockStatementDB(t)
	imports := NewStatementImportRepository(repository.db)
	mock.ExpectBegin()
	expectImportCatalog(mock)
	mock.ExpectExec("INSERT INTO `categories`").WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec("INSERT INTO `assets`").WillReturnResult(sqlmock.NewResult(20, 1))
	mock.ExpectQuery("SELECT .*amount.* FROM `assets`.*FOR UPDATE").WithArgs(int64(20), int64(2), 1).WillReturnRows(sqlmock.NewRows([]string{"amount"}).AddRow(88.88))
	failure := errors.New("duplicate import key")
	mock.ExpectExec("INSERT INTO `statements` .*`import_key`,`import_fingerprint`").WillReturnError(failure)
	mock.ExpectRollback()
	err := imports.WithinImportTransaction(context.Background(), 2, 1, func(tx repo.StatementImportMutation, c repo.ImportCatalog) error {
		if c.Book.ID != 2 || len(c.Statements) != 1 || c.Statements[0].ImportKey == nil || *c.Statements[0].ImportKey != "original-key" {
			t.Fatal("wrong locked book or missing import identity")
		}
		if _, err := tx.CreateCategory(context.Background(), model.Category{Name: "早餐", AccountBookID: 2, Type: "expend"}); err != nil {
			return err
		}
		if _, err := tx.CreateAsset(context.Background(), model.Asset{Name: "支付宝", AccountBookID: 2, Type: "deposit"}); err != nil {
			return err
		}
		flag := false
		key, fingerprint := "original-key", "original-fingerprint"
		_, err := tx.CreateStatement(context.Background(), model.Statement{UserID: 1, AccountBookID: 2, CategoryID: 10, AssetID: 20, Amount: 12.30, Type: "expend", BalanceApplied: &flag, ImportKey: &key, ImportFingerprint: &fingerprint}, repo.BalanceEffect{})
		return err
	})
	if !errors.Is(err, failure) {
		t.Fatalf("error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHistoricalCreateNeverUpdatesWalletAmount(t *testing.T) {
	repository, mock := mockStatementDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*amount.* FROM `assets`.*FOR UPDATE").WithArgs(int64(20), int64(2), 1).WillReturnRows(sqlmock.NewRows([]string{"amount"}).AddRow(88.88))
	mock.ExpectExec("INSERT INTO `statements`").WillReturnResult(sqlmock.NewResult(30, 1))
	mock.ExpectExec("UPDATE `categories`.*frequent").WithArgs(1, int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `assets`.*frequent").WithArgs(1, int64(20)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	flag := false
	err := repository.WithinTransaction(context.Background(), func(tx repo.Mutation) error {
		_, err := tx.Create(context.Background(), model.Statement{UserID: 1, AccountBookID: 2, AssetID: 20, CategoryID: 10, Amount: 12.30, Type: "expend", BalanceApplied: &flag}, repo.BalanceEffect{})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHistoricalEditRejectsAssetOutsideBookWithZeroDelta(t *testing.T) {
	repository, mock := mockStatementDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `statements`.*FOR UPDATE").WithArgs(int64(30), int64(2), 1).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "asset_id", "category_id", "balance_applied"}).AddRow(30, 1, 20, 10, false))
	mock.ExpectQuery("SELECT .*amount.* FROM `assets`.*FOR UPDATE").WithArgs(int64(99), int64(2), 1).WillReturnRows(sqlmock.NewRows([]string{"amount"}))
	mock.ExpectRollback()
	err := repository.WithinTransaction(context.Background(), func(tx repo.Mutation) error {
		return tx.UpdateByID(context.Background(), 30, 2, model.Statement{AssetID: 99, CategoryID: 10}, repo.BalanceEffect{}, repo.BalanceEffect{})
	})
	if err == nil {
		t.Fatalf("foreign asset accepted: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
