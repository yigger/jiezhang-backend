package mysql

import (
	"context"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
	"time"
)

func TestInsightQueryScopesBookAndTimeAndMapsEmbeddedStatement(t *testing.T) {
	db, mock := mockStatementDB(t)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT s.id, s.user_id, s.category_id, s.asset_id, s.amount, s.type, s.description, s.created_at, s.payee_id, ")+".*"+regexp.QuoteMeta("COALESCE(NULLIF(ac.remark, ''), NULLIF(u.nickname, ''), CONCAT('成员 ', s.user_id)) AS member_name")+".*LEFT JOIN payees p ON p.id = s.payee_id AND p.account_book_id = s.account_book_id.*WHERE s.account_book_id = \\? AND s.created_at >= \\? AND s.created_at < \\? ORDER BY s.created_at DESC, s.id DESC").WithArgs(int64(8), start, end).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "type", "created_at", "category_name", "merchant_name", "member_name"}).AddRow(99, 3, 12.34, "expend", start, "餐饮", "店铺", "成员"))
	rows, err := NewInsightsRepository(db.db).ListRows(context.Background(), 8, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != 99 || rows[0].UserID != 3 || rows[0].Amount != 12.34 || rows[0].MemberName != "成员" {
		t.Fatalf("projection: %+v", rows)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
