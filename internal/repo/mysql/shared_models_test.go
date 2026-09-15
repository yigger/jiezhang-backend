package mysql

import (
	"context"
	"github.com/yigger/jiezhang-backend/internal/model"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestFinanceQueriesReturnSharedAssetAndCategoryModels(t *testing.T) {
	statements, mock := mockStatementDB(t)
	repo := NewFinanceRepository(statements.db)
	mock.ExpectQuery("SELECT .* FROM .assets.").WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "amount", "parent_id", "icon_path", "type"}).AddRow(3, "现金", 12.5, 1, "cash.png", "deposit"))
	assets, err := repo.ListAssets(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets[0].ID != 3 || assets[0].Amount != 12.5 || assets[0].Type != "deposit" || assets[0].ParentID != 1 {
		t.Fatalf("assets: %+v", assets)
	}
	mock.ExpectQuery("SELECT special_type, id FROM .categories.").WithArgs("loan_in").
		WillReturnRows(sqlmock.NewRows([]string{"special_type", "id"}).AddRow("loan_in", 8))
	categories, err := repo.ListSpecialCategoryByTypes(context.Background(), []string{"loan_in"})
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 || categories[0].ID != 8 || categories[0].SpecialType != "loan_in" {
		t.Fatalf("categories: %+v", categories)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestUserModelKeepsNullableColumns(t *testing.T) {
	statements, mock := mockStatementDB(t)
	mock.ExpectQuery("SELECT .* FROM .users.").WithArgs(int64(5), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname", "avatar_url", "session_key", "third_session", "account_book_id"}).AddRow(5, nil, nil, nil, "session", 9))
	user, err := NewUserRepository(statements.db).FindByID(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != 5 || user.AccountBookID != 9 || user.Nickname != nil || user.AvatarURL != nil || user.SessionKey != nil || user.ThirdSession == nil || *user.ThirdSession != "session" {
		t.Fatalf("nullable fields: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCompletedTableMappingsScanDatabaseValues(t *testing.T) {
	statements, mock := mockStatementDB(t)
	mock.ExpectQuery("SELECT .* FROM .month_charts.").WillReturnRows(
		sqlmock.NewRows([]string{"id", "dashboard", "expend_compare", "user_id"}).
			AddRow(1, []byte(`{"amount":12.50}`), nil, nil))
	var charts []model.MonthChart
	if err := statements.db.Find(&charts).Error; err != nil {
		t.Fatal(err)
	}
	if len(charts) != 1 || string(charts[0].Dashboard) != `{"amount":12.50}` || charts[0].ExpendCompare != nil || charts[0].UserID != nil {
		t.Fatalf("JSON/null mapping: %+v", charts)
	}
	mock.ExpectQuery("SELECT .* FROM .users_assets.").WillReturnRows(
		sqlmock.NewRows([]string{"id", "asset_id", "assign_type"}).AddRow(2, "asset-abc", nil))
	var assignments []model.UserAssetAssignment
	if err := statements.db.Find(&assignments).Error; err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 1 || assignments[0].AssetID == nil || *assignments[0].AssetID != "asset-abc" || assignments[0].AssignType != nil {
		t.Fatalf("assignment mapping: %+v", assignments)
	}
	mock.ExpectQuery("SELECT .* FROM .messages.").WillReturnRows(
		sqlmock.NewRows([]string{"id", "target_id", "from_user_id"}).AddRow(3, 42, 7))
	var messages []model.Message
	if err := statements.db.Find(&messages).Error; err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || messages[0].TargetID != 42 || messages[0].FromUserID == nil || *messages[0].FromUserID != 7 {
		t.Fatalf("message mapping: %+v", messages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
