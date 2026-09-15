package mysql

import (
	"context"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type statementMutation struct{ *StatementRepository }

func (r *StatementRepository) WithinTransaction(ctx context.Context, fn func(repo.Mutation) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(&statementMutation{&StatementRepository{db: tx}}) })
}
func (r *statementMutation) LockCurrent(ctx context.Context, id, bookID int64) (tablemodel.Statement, error) {
	return r.getStatementForUpdate(r.db.WithContext(ctx), id, bookID)
}
