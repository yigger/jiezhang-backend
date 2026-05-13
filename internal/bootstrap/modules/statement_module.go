package modules

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/http/handler"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/urlbuilder"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repository/mysql"
	"github.com/yigger/jiezhang-backend/internal/service"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

func BuildStatementModule(db *gorm.DB, cfg config.Config, cache sessioncache.Cache) (handler.StatementsHandler, error) {
	statementRepo, err := mysqlrepo.NewStatementRepository(db)
	if err != nil {
		return handler.StatementsHandler{}, fmt.Errorf("init statement repository: %w", err)
	}
	categoryRepo, err := mysqlrepo.NewCategoryRepository(db)
	if err != nil {
		return handler.StatementsHandler{}, fmt.Errorf("init category repository: %w", err)
	}
	assetRepo, err := mysqlrepo.NewAssetRepository(db)
	if err != nil {
		return handler.StatementsHandler{}, fmt.Errorf("init asset repository: %w", err)
	}
	userRepo, err := mysqlrepo.NewUserRepository(db)
	if err != nil {
		return handler.StatementsHandler{}, fmt.Errorf("init user repository: %w", err)
	}

	publicURLBuilder := urlbuilder.NewPublicURLBuilder(cfg.PublicBaseURL)
	rowMapper := statementdto.NewRowMapper(publicURLBuilder)
	statementService := service.NewStatementServiceWithSession(
		statementRepo,
		statementRepo,
		categoryRepo,
		assetRepo,
		userRepo,
		cache,
		cfg.SessionTokenSecret,
		rowMapper,
	)
	statementsHandler := handler.NewStatementsHandler(statementService)
	return statementsHandler, nil
}
