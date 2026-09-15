package repo

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrStatementNotFound = errors.New("statement not found")
var ErrStatementAvatarNotFound = errors.New("statement avatar not found")

// StatementQueryRepository is for read-side complex queries.
type StatementQueryRepository interface {
	BatchGetStatements(context.Context, []int64) ([]tablemodel.Statement, error)
	ListSimpleRows(ctx context.Context, filter StatementListFilter) ([]tablemodel.Statement, error)
	BatchLookup
	GetSimpleRowByID(ctx context.Context, statementID int64, accountBookID int64) (tablemodel.Statement, error)
	GetLatestCategoryAssetByType(ctx context.Context, accountBookID int64, statementType string) (*tablemodel.Statement, error)
	ListDistinctTargetObjectsByType(ctx context.Context, accountBookID int64, statementType string) ([]string, error)
	ListAvatarRows(ctx context.Context, accountBookID int64) ([]tablemodel.UserAsset, error)
	ListAvatarsByStatementID(ctx context.Context, statementID int64) ([]tablemodel.UserAsset, error)
}

type StatementListFilter struct {
	UserID            int64
	AccountBookID     int64
	AssetID           int64
	Type              string
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
	Keyword           string
}

type StatementExportFilter struct {
	AccountBookID int64
	StartDate     time.Time
	EndDate       time.Time
	Limit         int
}

// BatchLookup fetches table rows needed by the service assembler.
type BatchLookup interface {
	BatchGetCategories(context.Context, []int64) ([]tablemodel.Category, error)
	BatchGetAssets(context.Context, []int64) ([]tablemodel.Asset, error)
	BatchGetPayees(context.Context, []int64) ([]tablemodel.Payee, error)
	BatchGetCollaboratorRemarks(context.Context, int64, []int64) ([]tablemodel.AccountBookCollaborator, error)
	BatchStatementAvatars(context.Context, []int64) ([]tablemodel.UserAsset, error)
}
