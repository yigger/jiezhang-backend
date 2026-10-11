package repo

import (
	"context"
	"github.com/yigger/jiezhang-backend/internal/model"
)

type CalendarJournalRepository interface {
	CanAccess(context.Context, int64, int64) (bool, error)
	List(context.Context, int64, int64, string, string) ([]model.CalendarJournal, error)
	ExpenseDates(context.Context, int64, string, string) ([]string, error)
	Save(context.Context, *model.CalendarJournal) error
	RevokeZeroExpense(context.Context, int64, int64, []string) error
}
