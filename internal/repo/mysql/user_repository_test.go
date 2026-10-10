package mysql

import (
	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"gorm.io/gorm"
	"strings"
	"testing"
)

func TestUserWritesUseNicknameWithoutPhantomNameColumn(t *testing.T) {
	db, _ := mockStatementDB(t)
	nickname, session := "test-user", "test-session"
	for _, operation := range []string{"create", "save"} {
		t.Run(operation, func(t *testing.T) {
			user := tablemodel.User{ID: 1, OpenID: "test-openid", Nickname: &nickname, ThirdSession: &session}
			tx := db.db.Session(&gorm.Session{DryRun: true, SkipDefaultTransaction: true})
			if operation == "create" {
				tx = tx.Create(&user)
			} else {
				tx = tx.Save(&user)
			}
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}
			query := tx.Statement.SQL.String()
			if strings.Contains(query, "`name`") || !strings.Contains(query, "`nickname`") || !strings.Contains(query, "`third_session`") {
				t.Fatal("user write must persist nickname and session without an unmapped name column")
			}
		})
	}
}
