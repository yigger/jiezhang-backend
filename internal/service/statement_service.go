package service

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

type StatementService struct {
	repo repository.StatementRepository
}

func NewStatementService(repo repository.StatementRepository) StatementService {
	return StatementService{repo: repo}
}

func (s StatementService) GetUserStatement(userId int64) ([]domain.Statement, error) {
	return s.repo.List(context.Background(), userId)
}
