package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

type PayeeRepository struct {
	db *gorm.DB
}

type payeeModel struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:name"`
	UserID        int64     `gorm:"column:user_id"`
	AccountBookID int64     `gorm:"column:account_book_id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (payeeModel) TableName() string {
	return "payees"
}

func NewPayeeRepository(db *gorm.DB) (*PayeeRepository, error) {
	return &PayeeRepository{db: db}, nil
}

func (r *PayeeRepository) ListByAccountBookID(ctx context.Context, accountBookID int64) ([]domain.Payee, error) {
	var models []payeeModel
	if err := r.db.WithContext(ctx).
		Where("account_book_id = ?", accountBookID).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]domain.Payee, 0, len(models))
	for _, model := range models {
		items = append(items, toPayeeDomain(model))
	}
	return items, nil
}

func (r *PayeeRepository) FindByIDAndUserID(ctx context.Context, payeeID int64, userID int64) (domain.Payee, error) {
	var model payeeModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", payeeID, userID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Payee{}, repository.ErrPayeeNotFound
		}
		return domain.Payee{}, err
	}
	return toPayeeDomain(model), nil
}

func (r *PayeeRepository) Create(ctx context.Context, payee domain.Payee) (domain.Payee, error) {
	model := payeeModel{
		Name:          payee.Name,
		UserID:        payee.UserID,
		AccountBookID: payee.AccountBookID,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Payee{}, err
	}
	return toPayeeDomain(model), nil
}

func (r *PayeeRepository) UpdateNameByID(ctx context.Context, payeeID int64, userID int64, name string) (domain.Payee, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&payeeModel{}).
		Where("id = ? AND user_id = ?", payeeID, userID).
		Updates(map[string]interface{}{
			"name":       name,
			"updated_at": now,
		})
	if res.Error != nil {
		return domain.Payee{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.Payee{}, repository.ErrPayeeNotFound
	}
	return r.FindByIDAndUserID(ctx, payeeID, userID)
}

func (r *PayeeRepository) DeleteByID(ctx context.Context, payeeID int64, userID int64) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", payeeID, userID).
		Delete(&payeeModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrPayeeNotFound
	}
	return nil
}

func toPayeeDomain(model payeeModel) domain.Payee {
	return domain.Payee{
		ID:            model.ID,
		Name:          model.Name,
		UserID:        model.UserID,
		AccountBookID: model.AccountBookID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}
