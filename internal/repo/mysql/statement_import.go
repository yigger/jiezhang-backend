package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatementImportRepository struct{ db *gorm.DB }

func NewStatementImportRepository(db *gorm.DB) *StatementImportRepository {
	return &StatementImportRepository{db: db}
}

func (r *StatementImportRepository) Catalog(ctx context.Context, book, user int64) (repo.ImportCatalog, error) {
	return importCatalog(r.db.WithContext(ctx), book, user, false)
}

func importCatalog(db *gorm.DB, book, user int64, locked bool) (repo.ImportCatalog, error) {
	var c repo.ImportCatalog
	query := db.Where("id = ?", book)
	if locked {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&c.Book).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c, repo.ErrAccountBookNotFound
		}
		return c, err
	}
	if err := db.Where("account_book_id = ? AND user_id = ?", book, user).Find(&c.Members).Error; err != nil {
		return c, err
	}
	// Check membership before reading private financial records. The service repeats authorization.
	if c.Book.UserID != user && len(c.Members) == 0 {
		return repo.ImportCatalog{}, repo.ErrAccountBookNotFound
	}
	if err := db.Where("account_book_id = ?", book).Order("id").Find(&c.Categories).Error; err != nil {
		return c, err
	}
	if err := db.Where("account_book_id = ?", book).Order("id").Find(&c.Assets).Error; err != nil {
		return c, err
	}
	if err := db.Select("id", "account_book_id", "user_id", "import_key", "import_fingerprint").Where("account_book_id = ? AND user_id = ? AND import_key IS NOT NULL", book, user).Find(&c.Statements).Error; err != nil {
		return c, err
	}
	return c, nil
}

func (r *StatementImportRepository) WithinImportTransaction(ctx context.Context, book, user int64, fn func(repo.StatementImportMutation, repo.ImportCatalog) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		catalog, err := importCatalog(tx, book, user, true)
		if err != nil {
			return err
		}
		return fn(&importMutation{db: tx}, catalog)
	})
}

type importMutation struct{ db *gorm.DB }

func (r *importMutation) CreateCategory(ctx context.Context, row model.Category) (int64, error) {
	row.CreatedAt, row.UpdatedAt = time.Now(), time.Now()
	err := r.db.WithContext(ctx).Create(&row).Error
	return row.ID, err
}
func (r *importMutation) CreateAsset(ctx context.Context, row model.Asset) (int64, error) {
	row.CreatedAt, row.UpdatedAt = time.Now(), time.Now()
	err := r.db.WithContext(ctx).Create(&row).Error
	return row.ID, err
}
func (r *importMutation) CreateStatement(ctx context.Context, row model.Statement, effect repo.BalanceEffect) (int64, error) {
	return (&statementMutation{StatementRepository: &StatementRepository{db: r.db}}).Create(ctx, row, effect)
}
