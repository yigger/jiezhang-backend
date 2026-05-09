package service

import (
	"context"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

type StatementService struct {
	queryRepo repository.StatementQueryRepository
}

func NewStatementService(queryRepo repository.StatementQueryRepository) StatementService {
	return StatementService{queryRepo: queryRepo}
}

type StatementListInput struct {
	UserID            int64
	AccountBookID     int64
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
}

func (s StatementService) GetStatements(ctx context.Context, input StatementListInput) ([]repository.StatementListItem, error) {
	filter := repository.StatementListFilter{
		UserID:            input.UserID,
		AccountBookID:     input.AccountBookID,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		ParentCategoryIDs: input.ParentCategoryIDs,
		ExceptIDs:         input.ExceptIDs,
		OrderBy:           input.OrderBy,
		Limit:             input.Limit,
		Offset:            input.Offset,
	}
	return s.queryRepo.ListWithRelations(ctx, filter)
}
