package billimport

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repo"
)

const MaxFileBytes = 20 << 20
const MaxRows = 50000

var ErrInvalid = errors.New("invalid bill import")
var ErrStorage = errors.New("bill import storage failed")

// File accepts streams so a future HTTP adapter need not expose filesystem paths.
type File struct {
	Reader io.Reader
	Format string // csv or xlsx
	Source string // empty, alipay or wechat
}
type Cell struct {
	Text    string
	Numeric bool
	Formula bool
}
type Table struct {
	Name      string
	Rows      [][]Cell
	Excel1904 bool
}
type TableReader interface {
	Read(context.Context, io.Reader, string) ([]Table, error)
}

// Target can select an existing ID or a name/path to resolve or create.
type Target struct {
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Parent       string `json:"parent,omitempty"`
	strictParent bool
}
type CategoryRule struct {
	Type     string   `json:"type"`
	Contains []string `json:"contains"`
	Target   Target   `json:"target"`
}
type Mapping struct {
	Assets     map[string]Target `json:"assets,omitempty"`     // source:payment method
	Categories map[string]Target `json:"categories,omitempty"` // type:source category
	Rules      []CategoryRule    `json:"rules,omitempty"`      // first matching rule wins
}
type Input struct {
	UserID        int64
	AccountBookID int64
	Files         []File
	Mapping       Mapping
	Apply         bool
}
type RowResult struct {
	File       int    `json:"file"`
	Row        int    `json:"row"`
	Source     string `json:"source"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
	CategoryID int64  `json:"category_id,omitempty"`
	AssetID    int64  `json:"asset_id,omitempty"`
}
type Creation struct {
	ID       int64  `json:"id"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	ParentID int64  `json:"parent_id,omitempty"`
}
type Report struct {
	Applied   bool        `json:"applied"`
	Imported  int         `json:"imported"`
	Ready     int         `json:"ready"`
	Duplicate int         `json:"duplicate"`
	Skipped   int         `json:"skipped"`
	Errors    int         `json:"errors"`
	Creations []Creation  `json:"creations"`
	Rows      []RowResult `json:"rows"`
}
type entry struct {
	File, Row                                                                int
	Source, ID, Kind, Category, Counterparty, Product, Payment, Status, Note string
	Occurred                                                                 time.Time
	Cents                                                                    int64
	Skip, Invalid, Fingerprint                                               string
}
type Service struct {
	repository repo.StatementImportRepository
	reader     TableReader
}

func New(repository repo.StatementImportRepository, reader TableReader) *Service {
	return &Service{repository: repository, reader: reader}
}
