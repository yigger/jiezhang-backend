package mysql

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type BudgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) *BudgetRepository { return &BudgetRepository{db: db} }

func (r *BudgetRepository) GetAccountBookBudget(ctx context.Context, accountBookID int64) (float64, error) {
	var row struct {
		Budget float64 `gorm:"column:budget"`
	}
	err := r.db.WithContext(ctx).
		Table("account_books").
		Select("COALESCE(budget, 0) AS budget").
		Where("id = ?", accountBookID).
		Take(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Budget, nil
}

func (r *BudgetRepository) SumExpendByMonth(ctx context.Context, accountBookID int64, year int, month int) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := r.db.WithContext(ctx).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND type = 'expend' AND year = ? AND month = ?", accountBookID, year, month).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Amount, nil
}

func (r *BudgetRepository) ListExpendParentCategories(ctx context.Context, accountBookID int64) ([]tablemodel.Category, error) {
	rows := make([]tablemodel.Category, 0)
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("id, name, icon_path, parent_id, COALESCE(budget, 0) AS budget, type").
		Where("account_book_id = ? AND type = 'expend' AND parent_id = 0", accountBookID).
		Order("`order` ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *BudgetRepository) ListChildCategoriesByParentID(ctx context.Context, accountBookID int64, parentID int64) ([]tablemodel.Category, error) {
	rows := make([]tablemodel.Category, 0)
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("id, name, icon_path, parent_id, COALESCE(budget, 0) AS budget, type").
		Where("account_book_id = ? AND parent_id = ?", accountBookID, parentID).
		Order("`order` ASC, id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *BudgetRepository) ListChildCategoryIDsByParentID(ctx context.Context, accountBookID int64, parentID int64) ([]int64, error) {
	ids := make([]int64, 0)
	err := r.db.WithContext(ctx).
		Table("categories").
		Where("account_book_id = ? AND parent_id = ?", accountBookID, parentID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *BudgetRepository) FindCategoryByID(ctx context.Context, accountBookID int64, categoryID int64) (tablemodel.Category, error) {
	var row tablemodel.Category
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("id, name, icon_path, parent_id, COALESCE(budget, 0) AS budget, type").
		Where("account_book_id = ? AND id = ?", accountBookID, categoryID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.Category{}, repo.ErrBudgetCategoryNotFound
		}
		return tablemodel.Category{}, err
	}
	return row, nil
}

func (r *BudgetRepository) SumStatementsByCategoryIDsAndMonth(ctx context.Context, accountBookID int64, categoryIDs []int64, year int, month int) (float64, error) {
	if len(categoryIDs) == 0 {
		return 0, nil
	}
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	query := r.db.WithContext(ctx).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND category_id IN ?", accountBookID, categoryIDs)
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	if month > 0 {
		query = query.Where("month = ?", month)
	}
	if err := query.Scan(&row).Error; err != nil {
		return 0, err
	}
	return row.Amount, nil
}

func (r *BudgetRepository) SumParentCategoryBudget(ctx context.Context, accountBookID int64) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := r.db.WithContext(ctx).
		Table("categories").
		Select("COALESCE(SUM(budget), 0) AS amount").
		Where("account_book_id = ? AND parent_id = 0", accountBookID).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.Amount, nil
}

func (r *BudgetRepository) UpdateAccountBookBudget(ctx context.Context, accountBookID int64, amount float64) error {
	res := r.db.WithContext(ctx).
		Table("account_books").
		Where("id = ?", accountBookID).
		Updates(map[string]interface{}{"budget": amount, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrAccountBookNotFound
	}
	return nil
}

func (r *BudgetRepository) UpdateCategoryBudget(ctx context.Context, accountBookID int64, categoryID int64, amount float64) error {
	res := r.db.WithContext(ctx).
		Table("categories").
		Where("account_book_id = ? AND id = ?", accountBookID, categoryID).
		Updates(map[string]interface{}{"budget": amount, "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrBudgetCategoryNotFound
	}
	return nil
}
