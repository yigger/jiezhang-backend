package auth

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/types"
	"testing"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
)

type fakeUsers struct {
	user  tablemodel.User
	err   error
	calls int
}

func (f *fakeUsers) FindByID(context.Context, int64) (tablemodel.User, error) {
	f.calls++
	return f.user, f.err
}
func (f *fakeUsers) FindByThirdSession(context.Context, string) (tablemodel.User, error) {
	f.calls++
	return f.user, f.err
}

type fakeCache map[string]string

func (f fakeCache) Get(key string) (string, bool) { v, ok := f[key]; return v, ok }
func TestSessionAuthentication(t *testing.T) {
	u := tablemodel.User{ID: 7}
	for _, tc := range []struct {
		name, app, key, cached string
		missing                bool
		want                   error
		calls                  int
	}{
		{"valid", "app", "session", "session", false, nil, 1},
		{"wrong app", "other", "session", "session", false, ErrInvalidAppID, 0},
		{"missing key", "app", "", "session", false, ErrSessionExpired, 0},
		{"expired", "app", "session", "", false, ErrSessionExpired, 1},
		{"rotated", "app", "session", "new-session", false, ErrSessionExpired, 1},
		{"unknown user", "app", "session", "session", true, ErrSessionExpired, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			users := &fakeUsers{user: u}
			if tc.missing {
				users.err = errors.New("not found")
			}
			s := NewSessionService(users, fakeCache{u.RedisSessionKey(): tc.cached}, "app", false)
			got, err := s.Authenticate(context.Background(), tc.app, tc.key)
			if !errors.Is(err, tc.want) || users.calls != tc.calls {
				t.Fatalf("user=%v err=%v queries=%d", got, err, users.calls)
			}
			if err == nil && got.ID != u.ID {
				t.Fatal("wrong principal")
			}
		})
	}
}
func TestPrincipalRequired(t *testing.T) {
	if _, err := CurrentUser(context.Background()); !errors.Is(err, ErrSessionExpired) {
		t.Fatal(err)
	}
	u, err := CurrentUser(WithUser(context.Background(), types.UserContext{ID: 7}))
	if err != nil || u.ID != 7 {
		t.Fatalf("%v %v", u, err)
	}
}
