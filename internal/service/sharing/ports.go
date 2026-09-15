package sharing

import (
	"context"
	"errors"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
)

var (
	ErrStatementInvalidToken = errors.New("statement share token invalid")
	ErrStatementDecodeFailed = errors.New("statement share token decode failed")
)

type Cache interface {
	Get(string) (string, bool)
	Set(string, string, time.Duration)
}
type Users interface {
	FindByID(context.Context, int64) (tablemodel.User, error)
}
type Reader interface {
	GetStatements(context.Context, statement.ListInput) ([]types.StatementListItem, error)
}
type Service struct {
	reader    Reader
	userRepo  Users
	cache     Cache
	codec     Codec
	rowMapper statement.RowMapper
}

func New(reader Reader, users Users, cache Cache, codec Codec, mapper statement.RowMapper) *Service {
	return &Service{reader: reader, userRepo: users, cache: cache, codec: codec, rowMapper: mapper}
}

type Codec interface {
	Encrypt([]byte) (string, error)
	Decrypt(string) ([]byte, error)
}
