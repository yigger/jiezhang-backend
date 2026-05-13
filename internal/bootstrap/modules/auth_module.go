package modules

import (
	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/wechat"
	"github.com/yigger/jiezhang-backend/internal/repository"
	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
)

type AuthModule struct {
	Handler      handler.AuthHandler
	SessionCache sessioncache.Cache
}

func BuildAuthModule(cfg config.Config, users repository.UserRepository, cache sessioncache.Cache) AuthModule {
	sessionCache := cache
	if sessionCache == nil {
		sessionCache = sessioncache.NewMemoryCache()
	}
	wechatClient := wechat.NewHTTPClient(cfg.MiniProgramAppID, cfg.MiniProgramSecret)

	checkOpenIDService := authservice.NewCheckOpenIDService(
		users,
		wechatClient,
		cfg.SessionTokenSecret,
		sessionCache,
	)
	authHandler := handler.NewAuthHandler(checkOpenIDService)

	return AuthModule{
		Handler:      authHandler,
		SessionCache: sessionCache,
	}
}
