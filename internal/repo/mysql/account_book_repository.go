package mysql

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type AccountBookRepository struct {
	db *gorm.DB
}

func NewAccountBookRepository(db *gorm.DB) *AccountBookRepository {
	return &AccountBookRepository{db: db}
}

func (r *AccountBookRepository) FindByID(ctx context.Context, id int64, userID int64) (tablemodel.AccountBook, error) {
	row, err := r.FindAccessibleByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repo.ErrAccountBookNotFound) {
			return tablemodel.AccountBook{}, repo.ErrAccountBookNotFound
		}
		return tablemodel.AccountBook{}, err
	}
	return row, nil
}

func (r *AccountBookRepository) ListAccessible(ctx context.Context, userID int64) ([]tablemodel.AccountBook, error) {
	rows := make([]tablemodel.AccountBook, 0)
	err := r.db.WithContext(ctx).
		Table("account_books ab").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = ab.id").
		Where("ab.user_id = ? OR abc.user_id = ?", userID, userID).
		Select("ab.id, ab.user_id, ab.account_type, ab.name, COALESCE(ab.description, '') AS description, COALESCE(ab.budget, 0) AS budget, ab.created_at, ab.updated_at").
		Group("ab.id").
		Order("ab.id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *AccountBookRepository) FindAccessibleByID(ctx context.Context, id int64, userID int64) (tablemodel.AccountBook, error) {
	var row tablemodel.AccountBook
	err := r.db.WithContext(ctx).
		Table("account_books ab").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = ab.id").
		Where("ab.id = ? AND (ab.user_id = ? OR abc.user_id = ?)", id, userID, userID).
		Select("ab.id, ab.user_id, ab.account_type, ab.name, COALESCE(ab.description, '') AS description, COALESCE(ab.budget, 0) AS budget, ab.created_at, ab.updated_at").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.AccountBook{}, repo.ErrAccountBookNotFound
		}
		return tablemodel.AccountBook{}, err
	}

	return row, nil
}

func (r *AccountBookRepository) Create(ctx context.Context, input repo.AccountBookCreateInput) (tablemodel.AccountBook, error) {
	var created tablemodel.AccountBook
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var innerErr error
		created, innerErr = createAccountBookInTx(tx, input)
		return innerErr
	})
	if err != nil {
		return tablemodel.AccountBook{}, err
	}

	return created, nil
}

func (r *AccountBookRepository) UpdateByID(ctx context.Context, id int64, input repo.AccountBookUpdateInput) error {
	res := r.db.WithContext(ctx).
		Table("account_books").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":         input.Name,
			"description":  input.Description,
			"account_type": input.AccountType,
			"updated_at":   time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrAccountBookNotFound
	}
	return nil
}

func (r *AccountBookRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tables := []string{"statements", "categories", "assets", "payees", "account_book_collaborators"}
		for _, table := range tables {
			if err := tx.Table(table).Where("account_book_id = ?", id).Delete(&struct{}{}).Error; err != nil {
				return err
			}
		}
		res := tx.Table("account_books").Where("id = ?", id).Delete(&struct{}{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return repo.ErrAccountBookNotFound
		}
		return nil
	})
}

func (r *AccountBookRepository) SwitchDefaultByUserID(ctx context.Context, userID int64, accountBookID int64) error {
	res := r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", userID).
		Update("account_book_id", accountBookID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrUserNotFound
	}
	return nil
}
