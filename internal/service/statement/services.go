package statement

import (
	"errors"

	"github.com/yigger/jiezhang-backend/internal/repo"
)

var (
	ErrStatementPermissionDenied = errors.New("statement permission denied")
	ErrStatementInvalidInput     = errors.New("statement invalid input")
)

type Reader struct {
	queryRepo    repo.StatementQueryRepository
	categoryRepo repo.CategoryRepository
	assetRepo    repo.AssetRepository
	rowMapper    RowMapper
}

func NewReader(query repo.StatementQueryRepository, categories repo.CategoryRepository, assets repo.AssetRepository, mapper RowMapper) *Reader {
	return &Reader{queryRepo: query, categoryRepo: categories, assetRepo: assets, rowMapper: mapper}
}

type Writer struct {
	statementRepo repo.AvatarWriter
	transaction   repo.Transactor
	queryRepo     repo.DetailQuery
	categoryRepo  repo.CategoryRepository
	rowMapper     RowMapper
}

func NewWriter(tx repo.Transactor, avatars repo.AvatarWriter, query repo.DetailQuery, categories repo.CategoryRepository, mapper RowMapper) *Writer {
	return &Writer{statementRepo: avatars, transaction: tx, queryRepo: query, categoryRepo: categories, rowMapper: mapper}
}
