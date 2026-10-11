package calendarjournal

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/types"
	"strings"
	"testing"
	"time"
)

type fakeRepo struct {
	access     bool
	expenses   []string
	rows       []model.CalendarJournal
	saved      *model.CalendarJournal
	revoked    []string
	book, user int64
}

func (f *fakeRepo) CanAccess(_ context.Context, b, u int64) (bool, error) {
	f.book, f.user = b, u
	return f.access, nil
}
func (f *fakeRepo) List(_ context.Context, b, u int64, _, _ string) ([]model.CalendarJournal, error) {
	f.book, f.user = b, u
	return f.rows, nil
}
func (f *fakeRepo) ExpenseDates(context.Context, int64, string, string) ([]string, error) {
	return f.expenses, nil
}
func (f *fakeRepo) Save(_ context.Context, r *model.CalendarJournal) error { f.saved = r; return nil }
func (f *fakeRepo) RevokeZeroExpense(_ context.Context, b, u int64, dates []string) error {
	f.book, f.user = b, u
	f.revoked = dates
	return nil
}
func fixture() (*Service, *fakeRepo) {
	r := &fakeRepo{access: true}
	s := New(r)
	s.now = func() time.Time { return time.Date(2026, 10, 8, 18, 0, 0, 0, time.UTC) }
	return s, r
}
func TestSaveValidatesDatesMoodLengthAndPermissions(t *testing.T) {
	cases := []struct {
		name string
		in   types.CalendarJournalInput
		want error
	}{
		{"invalid date", types.CalendarJournalInput{Date: "2026-02-30"}, ErrInvalid},
		{"future", types.CalendarJournalInput{Date: "2026-10-10"}, ErrFuture},
		{"unknown mood", types.CalendarJournalInput{Date: "2026-10-08", Mood: "unknown"}, ErrInvalid},
		{"long unicode", types.CalendarJournalInput{Date: "2026-10-08", Note: strings.Repeat("好", 201)}, ErrInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, r := fixture()
			_, err := s.Save(context.Background(), 8, 2, tc.in)
			if !errors.Is(err, tc.want) || r.saved != nil {
				t.Fatalf("err=%v saved=%v", err, r.saved)
			}
		})
	}
	s, r := fixture()
	r.access = false
	_, err := s.Save(context.Background(), 8, 2, types.CalendarJournalInput{Date: "2026-10-08"})
	if !errors.Is(err, ErrForbidden) || r.saved != nil {
		t.Fatal(err)
	}
}
func TestSavePersonalScopeUnicodeAndShanghaiDate(t *testing.T) {
	s, r := fixture()
	in := types.CalendarJournalInput{Date: "2026-10-09", Mood: "calm", Note: "  " + strings.Repeat("好", 200) + "  "}
	// Whitespace also counts toward the input limit.
	in.Note = strings.Repeat("好", 200)
	got, err := s.Save(context.Background(), 8, 2, in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Note != in.Note || r.saved.UserID != 2 || r.saved.AccountBookID != 8 {
		t.Fatalf("%+v", r.saved)
	}
}
func TestZeroExpenseRejectsExpenseButOrdinaryJournalCanStillSave(t *testing.T) {
	s, r := fixture()
	r.expenses = []string{"2026-10-08"}
	in := types.CalendarJournalInput{Date: "2026-10-08", ZeroExpense: true}
	if _, err := s.Save(context.Background(), 8, 2, in); !errors.Is(err, ErrExpense) {
		t.Fatal(err)
	}
	in.ZeroExpense = false
	in.Note = "午饭很好吃"
	if _, err := s.Save(context.Background(), 8, 2, in); err != nil {
		t.Fatal(err)
	}
}
func TestMonthRevokesOnlyConfirmedDaysWithExpenseAndPreservesNote(t *testing.T) {
	s, r := fixture()
	r.expenses = []string{"2026-10-08"}
	r.rows = []model.CalendarJournal{{Date: "2026-10-08", Mood: "happy", Note: "留住今天", ZeroExpense: true}, {Date: "2026-10-07", ZeroExpense: true}}
	got, err := s.Month(context.Background(), 8, 2, "2026-10")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.RevokedDates) != 1 || got.Entries[0].ZeroExpense || got.Entries[0].Note != "留住今天" || !got.Entries[1].ZeroExpense || len(r.revoked) != 1 || r.user != 2 || r.book != 8 {
		t.Fatalf("%+v", got)
	}
}
func TestMonthDeniesAccessAndRejectsMalformedMonth(t *testing.T) {
	s, r := fixture()
	if _, err := s.Month(context.Background(), 8, 2, "2026-13"); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	r.access = false
	if _, err := s.Month(context.Background(), 8, 2, "2026-10"); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
}
