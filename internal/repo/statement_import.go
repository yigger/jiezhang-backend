package repo

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/model"
)

type ImportCatalog struct {
	Book       model.AccountBook
	Members    []model.AccountBookCollaborator
	Categories []model.Category
	Assets     []model.Asset
	Statements []model.Statement
}

type StatementImportRepository interface {
	Catalog(context.Context, int64, int64) (ImportCatalog, error)
	WithinImportTransaction(context.Context, int64, int64, func(StatementImportMutation, ImportCatalog) error) error
}

type StatementImportMutation interface {
	CreateCategory(context.Context, model.Category) (int64, error)
	CreateAsset(context.Context, model.Asset) (int64, error)
	CreateStatement(context.Context, model.Statement, BalanceEffect) (int64, error)
}
