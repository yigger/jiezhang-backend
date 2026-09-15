package accountbook

import (
	"context"
	"errors"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

var ErrAccessDenied = errors.New("account book access denied")

type AccessRepository interface {
	FindByID(context.Context, int64, int64) (tablemodel.AccountBook, error)
}
type AccessService struct{ books AccessRepository }

func NewAccessService(books AccessRepository) *AccessService { return &AccessService{books: books} }
func (s *AccessService) Authorize(ctx context.Context, bookID, userID int64) (tablemodel.AccountBook, error) {
	if bookID <= 0 || userID <= 0 {
		return tablemodel.AccountBook{}, ErrAccessDenied
	}
	return s.books.FindByID(ctx, bookID, userID)
}
