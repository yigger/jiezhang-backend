package service

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

type AccountBookService struct {
	repo repository.AccountBookRepository
}

func NewAccountBookService(repo repository.AccountBookRepository) AccountBookService {
	return AccountBookService{repo: repo}
}

func (s AccountBookService) FindByID(ctx context.Context, id int64, userId int64) (domain.AccountBook, error) {
	return s.repo.FindByID(ctx, id, userId)
}
