package payee

import (
	"context"
	"errors"
	"strings"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
)

var ErrPayeeInvalidInput = errors.New("payee invalid input")

type PayeeService struct {
	repo repo.PayeeRepository
}

func NewPayeeService(repo repo.PayeeRepository) PayeeService {
	return PayeeService{repo: repo}
}

func (s PayeeService) List(ctx context.Context, accountBookID int64) ([]types.PayeeListItem, error) {
	rows, err := s.repo.ListByAccountBookID(ctx, accountBookID)
	if rows == nil {
		return nil, err
	}
	items := make([]types.PayeeListItem, len(rows))
	for i, row := range rows {
		items[i] = types.PayeeListItem{ID: row.ID, Name: row.Name}
	}
	return items, err
}

func (s PayeeService) Create(ctx context.Context, userID int64, accountBookID int64, name string) (tablemodel.Payee, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return tablemodel.Payee{}, ErrPayeeInvalidInput
	}
	return s.repo.Create(ctx, tablemodel.Payee{
		Name:          name,
		UserID:        userID,
		AccountBookID: accountBookID,
	})
}

func (s PayeeService) Update(ctx context.Context, payeeID int64, userID int64, name string) (tablemodel.Payee, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return tablemodel.Payee{}, ErrPayeeInvalidInput
	}
	return s.repo.UpdateNameByID(ctx, payeeID, userID, name)
}

func (s PayeeService) Delete(ctx context.Context, payeeID int64, userID int64) error {
	return s.repo.DeleteByID(ctx, payeeID, userID)
}
