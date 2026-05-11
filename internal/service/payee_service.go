package service

import (
	"context"
	"errors"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

var ErrPayeeInvalidInput = errors.New("payee invalid input")

type PayeeService struct {
	repo repository.PayeeRepository
}

func NewPayeeService(repo repository.PayeeRepository) PayeeService {
	return PayeeService{repo: repo}
}

func (s PayeeService) List(ctx context.Context, accountBookID int64) ([]domain.Payee, error) {
	return s.repo.ListByAccountBookID(ctx, accountBookID)
}

func (s PayeeService) Create(ctx context.Context, userID int64, accountBookID int64, name string) (domain.Payee, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Payee{}, ErrPayeeInvalidInput
	}
	return s.repo.Create(ctx, domain.Payee{
		Name:          name,
		UserID:        userID,
		AccountBookID: accountBookID,
	})
}

func (s PayeeService) Update(ctx context.Context, payeeID int64, userID int64, name string) (domain.Payee, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Payee{}, ErrPayeeInvalidInput
	}
	return s.repo.UpdateNameByID(ctx, payeeID, userID, name)
}

func (s PayeeService) Delete(ctx context.Context, payeeID int64, userID int64) error {
	return s.repo.DeleteByID(ctx, payeeID, userID)
}
