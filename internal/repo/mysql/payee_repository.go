package mysql

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type PayeeRepository struct {
	db *gorm.DB
}

func NewPayeeRepository(db *gorm.DB) *PayeeRepository { return &PayeeRepository{db: db} }

func (r *PayeeRepository) ListByAccountBookID(ctx context.Context, accountBookID int64) ([]tablemodel.Payee, error) {
	models := make([]tablemodel.Payee, 0)
	if err := r.db.WithContext(ctx).
		Where("account_book_id = ?", accountBookID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}

func (r *PayeeRepository) FindByIDAndUserID(ctx context.Context, payeeID int64, userID int64) (tablemodel.Payee, error) {
	var model tablemodel.Payee
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", payeeID, userID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tablemodel.Payee{}, repo.ErrPayeeNotFound
		}
		return tablemodel.Payee{}, err
	}
	return model, nil
}

func (r *PayeeRepository) Create(ctx context.Context, payee tablemodel.Payee) (tablemodel.Payee, error) {
	model := tablemodel.Payee{
		Name:          payee.Name,
		UserID:        payee.UserID,
		AccountBookID: payee.AccountBookID,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return tablemodel.Payee{}, err
	}
	return model, nil
}

func (r *PayeeRepository) UpdateNameByID(ctx context.Context, payeeID int64, userID int64, name string) (tablemodel.Payee, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&tablemodel.Payee{}).
		Where("id = ? AND user_id = ?", payeeID, userID).
		Updates(map[string]interface{}{
			"name":       name,
			"updated_at": now,
		})
	if res.Error != nil {
		return tablemodel.Payee{}, res.Error
	}
	if res.RowsAffected == 0 {
		return tablemodel.Payee{}, repo.ErrPayeeNotFound
	}
	return r.FindByIDAndUserID(ctx, payeeID, userID)
}

func (r *PayeeRepository) DeleteByID(ctx context.Context, payeeID int64, userID int64) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", payeeID, userID).
		Delete(&tablemodel.Payee{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repo.ErrPayeeNotFound
	}
	return nil
}
