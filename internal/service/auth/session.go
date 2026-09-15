package auth

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/types"
	"strings"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
)

var (
	ErrInvalidAppID   = errors.New("invalid appid")
	ErrSessionExpired = errors.New("session key overdue")
	ErrDevUserMissing = errors.New("[dev] user=1 not found")
)

type SessionUsers interface {
	FindByID(context.Context, int64) (tablemodel.User, error)
	FindByThirdSession(context.Context, string) (tablemodel.User, error)
}
type SessionCache interface{ Get(string) (string, bool) }
type SessionService struct {
	users       SessionUsers
	cache       SessionCache
	appID       string
	development bool
}

func NewSessionService(users SessionUsers, cache SessionCache, appID string, development bool) *SessionService {
	return &SessionService{users: users, cache: cache, appID: appID, development: development}
}
func (s *SessionService) Authenticate(ctx context.Context, appID, key string) (types.UserContext, error) {
	if s.development {
		u, e := s.users.FindByID(ctx, 1)
		if e != nil {
			return types.UserContext{}, ErrDevUserMissing
		}
		return helperservice.ToUserContext(u), nil
	}
	if strings.TrimSpace(appID) == "" || strings.TrimSpace(appID) != s.appID {
		return types.UserContext{}, ErrInvalidAppID
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return types.UserContext{}, ErrSessionExpired
	}
	u, e := s.users.FindByThirdSession(ctx, key)
	if e != nil {
		return types.UserContext{}, ErrSessionExpired
	}
	value, ok := s.cache.Get(u.RedisSessionKey())
	if !ok || value == "" || value != key {
		return types.UserContext{}, ErrSessionExpired
	}
	return helperservice.ToUserContext(u), nil
}

type principalKey struct{}

func WithUser(ctx context.Context, user types.UserContext) context.Context {
	return context.WithValue(ctx, principalKey{}, user)
}
func CurrentUser(ctx context.Context) (types.UserContext, error) {
	u, ok := ctx.Value(principalKey{}).(types.UserContext)
	if !ok || u.ID <= 0 {
		return types.UserContext{}, ErrSessionExpired
	}
	return u, nil
}
