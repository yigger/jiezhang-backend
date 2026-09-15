package mysql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestCategoryParentsApplyDirectQueryArguments(t *testing.T) {
	for _, categoryType := range []string{"expend", ""} {
		t.Run("type="+categoryType, func(t *testing.T) {
			statements, mock := mockStatementDB(t)
			query := mock.ExpectQuery("SELECT .*categories.*")
			if categoryType == "" {
				query.WithArgs(int64(9))
			} else {
				query.WithArgs(int64(9), categoryType)
			}
			query.WillReturnRows(sqlmock.NewRows([]string{"id", "name", "icon_path"}).AddRow(4, "餐饮", "food.png"))
			rows, err := NewCategoryRepository(statements.db).ListParents(context.Background(), 9, categoryType)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].ID != 4 || rows[0].Name != "餐饮" {
				t.Fatalf("rows: %+v", rows)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
