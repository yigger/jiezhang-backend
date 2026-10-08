package calendarjournal

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("手帐内容无效")
var ErrForbidden = errors.New("没有权限访问此账簿")
var ErrFuture = errors.New("只能记录今天或过去的日子")
var ErrExpense = errors.New("当天已有支出，不能标记零消费")

type Service struct {
	repo repo.CalendarJournalRepository
	now  func() time.Time
}

func New(r repo.CalendarJournalRepository) *Service { return &Service{r, time.Now} }
func (s *Service) authorize(ctx context.Context, book, user int64) error {
	ok, err := s.repo.CanAccess(ctx, book, user)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
func (s *Service) Month(ctx context.Context, book, user int64, month string) (types.CalendarJournalMonth, error) {
	result := types.CalendarJournalMonth{Entries: []types.CalendarJournalInput{}, RevokedDates: []string{}}
	date, err := time.Parse("2006-01", month)
	if err != nil {
		return result, ErrInvalid
	}
	if err = s.authorize(ctx, book, user); err != nil {
		return result, err
	}
	start, end := date.Format("2006-01-02"), date.AddDate(0, 1, 0).Format("2006-01-02")
	expenses, err := s.repo.ExpenseDates(ctx, book, start, end)
	if err != nil {
		return result, err
	}
	rows, err := s.repo.List(ctx, book, user, start, end)
	if err != nil {
		return result, err
	}
	spent := map[string]bool{}
	for _, d := range expenses {
		spent[d] = true
	}
	for _, row := range rows {
		if row.ZeroExpense && spent[row.Date] {
			result.RevokedDates = append(result.RevokedDates, row.Date)
			row.ZeroExpense = false
		}
		result.Entries = append(result.Entries, types.CalendarJournalInput{Date: row.Date, Mood: row.Mood, Note: row.Note, ZeroExpense: row.ZeroExpense})
	}
	if err = s.repo.RevokeZeroExpense(ctx, book, user, result.RevokedDates); err != nil {
		return result, err
	}
	return result, nil
}
func (s *Service) Save(ctx context.Context, book, user int64, in types.CalendarJournalInput) (types.CalendarJournalInput, error) {
	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil || date.Format("2006-01-02") != in.Date || utf8.RuneCountInString(in.Note) > 200 {
		return in, ErrInvalid
	}
	switch in.Mood {
	case "", "happy", "calm", "tired", "low", "spark":
	default:
		return in, ErrInvalid
	}
	// Calendar dates are interpreted in the application's Shanghai timezone.
	zone := time.FixedZone("Asia/Shanghai", 8*60*60)
	if in.Date > s.now().In(zone).Format("2006-01-02") {
		return in, ErrFuture
	}
	if err = s.authorize(ctx, book, user); err != nil {
		return in, err
	}
	if in.ZeroExpense {
		expenses, e := s.repo.ExpenseDates(ctx, book, in.Date, date.AddDate(0, 0, 1).Format("2006-01-02"))
		if e != nil {
			return in, e
		}
		if len(expenses) > 0 {
			return in, ErrExpense
		}
	}
	in.Note = strings.TrimSpace(in.Note)
	now := s.now()
	row := model.CalendarJournal{AccountBookID: book, UserID: user, Date: in.Date, Mood: in.Mood, Note: in.Note, ZeroExpense: in.ZeroExpense, CreatedAt: now, UpdatedAt: now}
	return in, s.repo.Save(ctx, &row)
}
