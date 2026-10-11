package mysql

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InsightsStorage struct{ db *gorm.DB }

func NewInsightsStorage(db *gorm.DB) *InsightsStorage { return &InsightsStorage{db: db} }
func (s *InsightsStorage) ListProjects(ctx context.Context, book int64) ([]model.InsightProject, error) {
	rows := []model.InsightProject{}
	err := s.db.WithContext(ctx).Where("account_book_id = ?", book).Order("archived ASC, id DESC").Find(&rows).Error
	return rows, err
}
func (s *InsightsStorage) ListFixedCosts(ctx context.Context, book int64) ([]model.InsightFixedCost, error) {
	rows := []model.InsightFixedCost{}
	err := s.db.WithContext(ctx).Where("account_book_id = ?", book).Order("active DESC, id DESC").Find(&rows).Error
	return rows, err
}
func (s *InsightsStorage) ListAnnotations(ctx context.Context, book int64) ([]model.InsightStatementAnnotation, error) {
	rows := []model.InsightStatementAnnotation{}
	err := s.db.WithContext(ctx).Where("account_book_id = ?", book).Find(&rows).Error
	return rows, err
}
func (s *InsightsStorage) ListPortfolioSnapshots(ctx context.Context, book int64) ([]model.InsightPortfolioSnapshot, error) {
	rows := []model.InsightPortfolioSnapshot{}
	err := s.db.WithContext(ctx).Where("account_book_id = ?", book).Order("created_at ASC, id ASC").Find(&rows).Error
	return rows, err
}
func (s *InsightsStorage) ListMembers(ctx context.Context, book int64) ([]model.AccountBookCollaborator, error) {
	rows := []model.AccountBookCollaborator{}
	err := s.db.WithContext(ctx).Where("account_book_id = ?", book).Find(&rows).Error
	return rows, err
}
func (s *InsightsStorage) Transact(ctx context.Context, fn func(repo.InsightsStorageTx) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(&insightsStorageTx{db: tx}) })
}

type insightsStorageTx struct{ db *gorm.DB }

func (t *insightsStorageTx) LockBook(id int64) (model.AccountBook, error) {
	var row model.AccountBook
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&row).Error
	return row, insightLookupError(err)
}
func (t *insightsStorageTx) LockStatement(book, id int64) (model.Statement, error) {
	var row model.Statement
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("account_book_id = ? AND id = ?", book, id).Take(&row).Error
	return row, insightLookupError(err)
}
func (t *insightsStorageTx) Project(book, id int64) (model.InsightProject, error) {
	var row model.InsightProject
	err := t.db.Where("account_book_id = ? AND id = ?", book, id).Take(&row).Error
	return row, insightLookupError(err)
}
func (t *insightsStorageTx) FixedCost(book, id int64) (model.InsightFixedCost, error) {
	var row model.InsightFixedCost
	err := t.db.Where("account_book_id = ? AND id = ?", book, id).Take(&row).Error
	return row, insightLookupError(err)
}
func (t *insightsStorageTx) Members(book int64) ([]model.AccountBookCollaborator, error) {
	rows := []model.AccountBookCollaborator{}
	err := t.db.Where("account_book_id = ?", book).Find(&rows).Error
	return rows, err
}
func (t *insightsStorageTx) Assets(book int64) ([]model.Asset, error) {
	rows := []model.Asset{}
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("account_book_id = ? AND parent_id > 0", book).Order("id ASC").Find(&rows).Error
	return rows, err
}
func (t *insightsStorageTx) SaveProject(row *model.InsightProject) error {
	if row.ID == 0 {
		return t.db.Create(row).Error
	}
	return t.db.Model(&model.InsightProject{}).Where("id = ? AND account_book_id = ?", row.ID, row.AccountBookID).Select("icon", "color", "name", "budget_cents", "archived", "participant_ids", "start_date", "end_date", "updated_at").Updates(row).Error
}

func (t *insightsStorageTx) SaveFixedCost(row *model.InsightFixedCost) error {
	// SQL DATE is a calendar value; bind it as text instead of converting an instant
	// through the connection timezone (which could move a midnight to yesterday).
	var next any
	if row.NextRunDate != nil {
		next = row.NextRunDate.Format("2006-01-02")
	}
	if row.ID == 0 {
		date := row.NextRunDate
		row.NextRunDate = nil
		err := t.db.Create(row).Error
		row.NextRunDate = date
		if err != nil {
			return err
		}
		if next == nil {
			return nil
		}
		return t.db.Model(&model.InsightFixedCost{}).Where("id = ? AND account_book_id = ?", row.ID, row.AccountBookID).UpdateColumn("next_run_date", next).Error
	}
	return t.db.Model(&model.InsightFixedCost{}).Where("id = ? AND account_book_id = ?", row.ID, row.AccountBookID).Updates(map[string]any{"name": row.Name, "amount_cents": row.AmountCents, "category_id": row.CategoryID, "asset_id": row.AssetID, "interval_months": row.IntervalMonths, "due_day": row.DueDay, "candidate_key": row.CandidateKey, "active": row.Active, "next_run_date": next, "updated_at": row.UpdatedAt}).Error
}

func (t *insightsStorageTx) SaveAnnotation(row *model.InsightStatementAnnotation) error {
	return t.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "statement_id"}}, DoUpdates: clause.AssignmentColumns([]string{"project_id", "consumer_id", "payer_id", "allocations", "fixed_cost_id", "updated_at"})}).Create(row).Error
}
func (t *insightsStorageTx) CreateSnapshot(row *model.InsightPortfolioSnapshot) error {
	return t.db.Create(row).Error
}

func (t *insightsStorageTx) Category(book, id int64) (model.Category, error) {
	var row model.Category
	err := t.db.Where("id = ? AND (account_book_id = ? OR account_book_id = 0)", id, book).Take(&row).Error
	return row, insightLookupError(err)
}

func (t *insightsStorageTx) FixedCosts(book int64) ([]model.InsightFixedCost, error) {
	rows := []model.InsightFixedCost{}
	err := t.db.Where("account_book_id = ?", book).Find(&rows).Error
	return rows, err
}
func (t *insightsStorageTx) Payee(book, id int64) (model.Payee, error) {
	var row model.Payee
	err := t.db.Where("account_book_id = ? AND id = ?", book, id).Take(&row).Error
	return row, insightLookupError(err)
}
func (t *insightsStorageTx) SaveMerchantAlias(row *model.InsightMerchantAlias) error {
	return t.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "payee_id"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"})}).Create(row).Error
}

func (t *insightsStorageTx) Annotation(book, statement int64) (model.InsightStatementAnnotation, bool, error) {
	var row model.InsightStatementAnnotation
	err := t.db.Where("account_book_id = ? AND statement_id = ?", book, statement).Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return row, false, nil
	}
	return row, err == nil, err
}

func insightLookupError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return repo.ErrInsightRecordNotFound
	}
	return err
}

func (s *InsightsStorage) ListScheduledFixedCosts(ctx context.Context, through time.Time) ([]model.InsightFixedCost, error) {
	rows := []model.InsightFixedCost{}
	err := s.db.WithContext(ctx).Where("active = ? AND next_run_date IS NOT NULL AND next_run_date <= ?", true, through.Format("2006-01-02")).Order("account_book_id, id").Find(&rows).Error
	return rows, err
}
func (t *insightsStorageTx) CreateFixedCostStatement(ctx context.Context, input model.Statement, effect repo.BalanceEffect) (int64, error) {
	return (&statementMutation{StatementRepository: NewStatementRepository(t.db)}).Create(ctx, input, effect)
}
func (t *insightsStorageTx) FixedCostRun(rule int64, date time.Time) (bool, error) {
	var n int64
	month := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	err := t.db.Model(&model.InsightFixedCostRun{}).Where("fixed_cost_id = ? AND due_date >= ? AND due_date < ?", rule, month.Format("2006-01-02"), month.AddDate(0, 1, 0).Format("2006-01-02")).Count(&n).Error
	return n > 0, err
}
func (t *insightsStorageTx) SaveFixedCostRun(row *model.InsightFixedCostRun) error {
	return t.db.Model(&model.InsightFixedCostRun{}).Create(map[string]any{"fixed_cost_id": row.FixedCostID, "due_date": row.DueDate.Format("2006-01-02"), "statement_id": row.StatementID, "created_at": row.CreatedAt}).Error
}
