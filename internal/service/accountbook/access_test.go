package accountbook

import (
	"context"
	"errors"
	"testing"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

type membershipRepo struct{ calls int }

func (r *membershipRepo) FindByID(_ context.Context, book, user int64) (tablemodel.AccountBook, error) {
	r.calls++
	if book != 3 || user != 7 {
		return tablemodel.AccountBook{}, ErrAccessDenied
	}
	return tablemodel.AccountBook{ID: book}, nil
}
func TestBookAuthorization(t *testing.T) {
	repo := &membershipRepo{}
	s := NewAccessService(repo)
	for _, tc := range []struct {
		book, user int64
		allowed    bool
	}{{3, 7, true}, {3, 8, false}, {4, 7, false}, {0, 7, false}, {3, 0, false}} {
		book, err := s.Authorize(context.Background(), tc.book, tc.user)
		if tc.allowed {
			if err != nil || book.ID != 3 {
				t.Fatalf("%v %v", book, err)
			}
		} else if !errors.Is(err, ErrAccessDenied) {
			t.Fatal(err)
		}
	}
	if repo.calls != 3 {
		t.Fatalf("invalid IDs reached repository: %d calls", repo.calls)
	}
}
