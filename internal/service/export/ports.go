package export

import (
	"context"
	"errors"
	"github.com/yigger/jiezhang-backend/internal/types"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

var ErrStatementExportLimited = errors.New("statement export daily limit reached")

type Cache interface {
	Get(string) (string, bool)
	Set(string, string, time.Duration)
}
type Query interface {
	repo.BatchLookup
	ListExportRows(context.Context, repo.StatementExportFilter) ([]tablemodel.Statement, error)
}
type Service struct {
	queryRepo Query
	cache     Cache
	renderer  Renderer
}

func New(query Query, cache Cache, renderer Renderer) *Service {
	return &Service{queryRepo: query, cache: cache, renderer: renderer}
}

type Renderer interface {
	Render([]types.ExportRow) ([]byte, error)
}
