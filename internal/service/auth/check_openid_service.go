package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/wechat"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

var ErrLoginFailed = errors.New("login failed")

type CheckOpenIDService struct {
	users       repository.UserRepository
	wechat      wechat.Client
	tokenSecret string
	cache       sessioncache.Cache
}

func NewCheckOpenIDService(
	users repository.UserRepository,
	wechatClient wechat.Client,
	tokenSecret string,
	cache sessioncache.Cache,
) CheckOpenIDService {
	return CheckOpenIDService{
		users:       users,
		wechat:      wechatClient,
		tokenSecret: tokenSecret,
		cache:       cache,
	}
}

func (s CheckOpenIDService) Execute(ctx context.Context, code string) (string, error) {
	session, err := s.exchangeCodeWithRetry(ctx, strings.TrimSpace(code), 4)
	if err != nil {
		return "", ErrLoginFailed
	}

	user, err := s.users.FindByOpenID(ctx, session.OpenID)
	if err != nil {
		if !errors.Is(err, repository.ErrUserNotFound) {
			return "", err
		}
		user = domain.User{
			OpenID:     session.OpenID,
			SessionKey: session.SessionKey,
		}
		user, err = s.users.Create(ctx, user)
		if err != nil {
			return "", err
		}
	}

	cacheKey := user.RedisSessionKey()
	if cached, ok := s.cache.Get(cacheKey); ok && strings.TrimSpace(cached) != "" {
		return cached, nil
	}

	thirdSession, err := s.generateSecureToken(user.ID, session.SessionKey)
	if err != nil {
		return "", err
	}

	user.SessionKey = session.SessionKey
	user.ThirdSession = thirdSession
	if _, err := s.users.Save(ctx, user); err != nil {
		return "", err
	}

	s.cache.Set(cacheKey, thirdSession, 48*time.Hour)
	return thirdSession, nil
}

func (s CheckOpenIDService) exchangeCodeWithRetry(ctx context.Context, code string, maxRetries int) (wechat.SessionResponse, error) {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		resp, err := s.wechat.Code2Session(ctx, code)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if !isTimeout(err) {
			break
		}
	}

	return wechat.SessionResponse{}, lastErr
}

func (s CheckOpenIDService) generateSecureToken(userID int64, sessionKey string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	raw := fmt.Sprintf(
		"%d|%s|%d|%s|%s",
		userID,
		sessionKey,
		time.Now().Unix(),
		hex.EncodeToString(randomBytes),
		s.tokenSecret,
	)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]), nil
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	return errors.Is(err, context.DeadlineExceeded)
}
