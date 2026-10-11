package repo

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/model"
	"time"
)

var ErrInsightRecordNotFound = errors.New("关联记录不存在或不属于当前账簿")

// InsightsStorage keeps analysis metadata separate from financial writes.
type InsightsStorage interface {
	ListProjects(context.Context, int64) ([]model.InsightProject, error)
	ListScheduledFixedCosts(context.Context, time.Time) ([]model.InsightFixedCost, error)
	ListFixedCosts(context.Context, int64) ([]model.InsightFixedCost, error)
	ListAnnotations(context.Context, int64) ([]model.InsightStatementAnnotation, error)
	ListPortfolioSnapshots(context.Context, int64) ([]model.InsightPortfolioSnapshot, error)
	ListMembers(context.Context, int64) ([]model.AccountBookCollaborator, error)
	Transact(context.Context, func(InsightsStorageTx) error) error
}
type InsightsStorageTx interface {
	LockBook(int64) (model.AccountBook, error)
	LockStatement(int64, int64) (model.Statement, error)
	Project(int64, int64) (model.InsightProject, error)
	Category(int64, int64) (model.Category, error)
	FixedCost(int64, int64) (model.InsightFixedCost, error)
	FixedCosts(int64) ([]model.InsightFixedCost, error)
	Members(int64) ([]model.AccountBookCollaborator, error)
	Assets(int64) ([]model.Asset, error)
	Annotation(int64, int64) (model.InsightStatementAnnotation, bool, error)
	Payee(int64, int64) (model.Payee, error)
	SaveMerchantAlias(*model.InsightMerchantAlias) error
	SaveProject(*model.InsightProject) error
	SaveFixedCost(*model.InsightFixedCost) error
	SaveAnnotation(*model.InsightStatementAnnotation) error
	CreateSnapshot(*model.InsightPortfolioSnapshot) error
}

// FixedCostLedgerTx joins recurring metadata to the existing financial transaction.
type FixedCostLedgerTx interface {
	CreateFixedCostStatement(context.Context, model.Statement, BalanceEffect) (int64, error)
	FixedCostRun(int64, time.Time) (bool, error)
	SaveFixedCostRun(*model.InsightFixedCostRun) error
}
