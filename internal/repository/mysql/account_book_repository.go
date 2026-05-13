package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
	"gorm.io/gorm"
)

type AccountBookRepository struct {
	db *gorm.DB
}

func NewAccountBookRepository(db *gorm.DB) (*AccountBookRepository, error) {
	return &AccountBookRepository{db: db}, nil
}

type accountBookRow struct {
	ID          int64     `gorm:"column:id"`
	UserID      int64     `gorm:"column:user_id"`
	AccountType int       `gorm:"column:account_type"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Budget      float64   `gorm:"column:budget"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type accountBookCollaboratorRow struct {
	AccountBookID int64     `gorm:"column:account_book_id"`
	UserID        int64     `gorm:"column:user_id"`
	Role          string    `gorm:"column:role"`
	Remark        string    `gorm:"column:remark"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

type categoryRow struct {
	ID            int64     `gorm:"column:id"`
	UserID        int64     `gorm:"column:user_id"`
	AccountBookID int64     `gorm:"column:account_book_id"`
	Name          string    `gorm:"column:name"`
	ParentID      int64     `gorm:"column:parent_id"`
	Order         int       `gorm:"column:order"`
	IconPath      string    `gorm:"column:icon_path"`
	Type          string    `gorm:"column:type"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

type assetRow struct {
	ID            int64     `gorm:"column:id"`
	AccountBookID int64     `gorm:"column:account_book_id"`
	Name          string    `gorm:"column:name"`
	Amount        float64   `gorm:"column:amount"`
	ParentID      int64     `gorm:"column:parent_id"`
	Type          string    `gorm:"column:type"`
	IconPath      string    `gorm:"column:icon_path"`
	CreatorID     int64     `gorm:"column:creator_id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (r *AccountBookRepository) FindByID(ctx context.Context, id int64, userID int64) (domain.AccountBook, error) {
	row, err := r.FindAccessibleByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountBookNotFound) {
			return domain.AccountBook{}, repository.ErrAccountBookNotFound
		}
		return domain.AccountBook{}, err
	}
	return domain.AccountBook{
		ID:          row.ID,
		UserID:      row.UserID,
		AccountType: row.AccountType,
		Name:        row.Name,
		Description: row.Description,
		Budget:      row.Budget,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *AccountBookRepository) ListAccessible(ctx context.Context, userID int64) ([]repository.AccountBookRecord, error) {
	rows := make([]accountBookRow, 0)
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

	items := make([]repository.AccountBookRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, repository.AccountBookRecord{
			ID:          row.ID,
			UserID:      row.UserID,
			AccountType: row.AccountType,
			Name:        row.Name,
			Description: row.Description,
			Budget:      row.Budget,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *AccountBookRepository) FindAccessibleByID(ctx context.Context, id int64, userID int64) (repository.AccountBookRecord, error) {
	var row accountBookRow
	err := r.db.WithContext(ctx).
		Table("account_books ab").
		Joins("LEFT JOIN account_book_collaborators abc ON abc.account_book_id = ab.id").
		Where("ab.id = ? AND (ab.user_id = ? OR abc.user_id = ?)", id, userID, userID).
		Select("ab.id, ab.user_id, ab.account_type, ab.name, COALESCE(ab.description, '') AS description, COALESCE(ab.budget, 0) AS budget, ab.created_at, ab.updated_at").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repository.AccountBookRecord{}, repository.ErrAccountBookNotFound
		}
		return repository.AccountBookRecord{}, err
	}

	return repository.AccountBookRecord{
		ID:          row.ID,
		UserID:      row.UserID,
		AccountType: row.AccountType,
		Name:        row.Name,
		Description: row.Description,
		Budget:      row.Budget,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *AccountBookRepository) Create(ctx context.Context, input repository.AccountBookCreateInput) (repository.AccountBookRecord, error) {
	var created accountBookRow
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		created = accountBookRow{
			UserID:      input.UserID,
			AccountType: input.AccountType,
			Name:        input.Name,
			Description: input.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Table("account_books").Create(&created).Error; err != nil {
			return err
		}

		collaborator := accountBookCollaboratorRow{
			AccountBookID: created.ID,
			UserID:        input.UserID,
			Role:          "owner",
			Remark:        input.UserNickname,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Table("account_book_collaborators").Create(&collaborator).Error; err != nil {
			return err
		}

		for statementType, parents := range input.Categories {
			for parentOrder, parent := range parents {
				parentCategory := categoryRow{
					UserID:        input.UserID,
					AccountBookID: created.ID,
					Name:          parent.Name,
					IconPath:      parent.IconPath,
					Type:          statementType,
					Order:         parentOrder + 1,
					ParentID:      0,
					CreatedAt:     now,
					UpdatedAt:     now,
				}
				if err := tx.Table("categories").Create(&parentCategory).Error; err != nil {
					return err
				}
				for childOrder, child := range parent.Childs {
					childCategory := categoryRow{
						UserID:        input.UserID,
						AccountBookID: created.ID,
						Name:          child.Name,
						IconPath:      child.IconPath,
						Type:          statementType,
						Order:         childOrder,
						ParentID:      parentCategory.ID,
						CreatedAt:     now,
						UpdatedAt:     now,
					}
					if err := tx.Table("categories").Create(&childCategory).Error; err != nil {
						return err
					}
				}
			}
		}

		for _, asset := range input.Assets {
			assetType := asset.Type
			if assetType == "" {
				assetType = "deposit"
			}
			parentAsset := assetRow{
				AccountBookID: created.ID,
				Name:          asset.Name,
				IconPath:      asset.IconPath,
				Type:          assetType,
				ParentID:      0,
				CreatorID:     input.UserID,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Table("assets").Create(&parentAsset).Error; err != nil {
				return err
			}

			for _, child := range asset.Childs {
				childAsset := assetRow{
					AccountBookID: created.ID,
					Name:          child.Name,
					IconPath:      child.IconPath,
					Type:          assetType,
					ParentID:      parentAsset.ID,
					CreatorID:     input.UserID,
					CreatedAt:     now,
					UpdatedAt:     now,
				}
				if err := tx.Table("assets").Create(&childAsset).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return repository.AccountBookRecord{}, err
	}

	return repository.AccountBookRecord{
		ID:          created.ID,
		UserID:      created.UserID,
		AccountType: created.AccountType,
		Name:        created.Name,
		Description: created.Description,
		Budget:      created.Budget,
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}, nil
}

func (r *AccountBookRepository) UpdateByID(ctx context.Context, id int64, input repository.AccountBookUpdateInput) error {
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
		return repository.ErrAccountBookNotFound
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
			return repository.ErrAccountBookNotFound
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
		return repository.ErrUserNotFound
	}
	return nil
}
