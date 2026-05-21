package modules

import (
	"fmt"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/signedurl"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/wechat"
	"github.com/yigger/jiezhang-backend/internal/repository"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
	authservice "github.com/yigger/jiezhang-backend/internal/service/auth"
	"gorm.io/gorm"
)

type AuthModule struct {
	Handler      handler.AuthHandler
	SessionCache sessioncache.Cache
}

func BuildAuthModule(cfg config.Config, db *gorm.DB, users repository.UserRepository, cache sessioncache.Cache) (AuthModule, error) {
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

	uploadRepo, err := mysqlrepo.NewUploadRepository(db)
	if err != nil {
		return AuthModule{}, fmt.Errorf("init upload repository: %w", err)
	}
	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	if err != nil {
		return AuthModule{}, fmt.Errorf("init statement repository: %w", err)
	}
	signer := signedurl.NewSigner(cfg.SessionTokenSecret)
	uploadURLBuilder := urlbuilder.NewPublicURLBuilderWithSigner(cfg.PublicBaseURL, signer)
	uploadService := service.NewUploadService(users, uploadRepo, statementRepo, uploadURLBuilder)
	authHandler := handler.NewAuthHandler(checkOpenIDService, uploadService)

	return AuthModule{
		Handler:      authHandler,
		SessionCache: sessionCache,
	}, nil
}
