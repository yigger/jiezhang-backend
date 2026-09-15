package export

import (
	"context"
	"fmt"
	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	statement "github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
	"strings"
	"time"
)

type StatementExportInput struct {
	AccountBookID int64
	UserID        int64
	Range         string
}

type StatementExportRowItem = types.ExportRow

func (s Service) ExportCheck(_ context.Context, userID int64) (types.StatementExportCheckResult, error) {
	if userID <= 0 {
		return types.StatementExportCheckResult{}, statement.ErrStatementInvalidInput
	}
	count, err := s.getTodayExportCount(userID)
	if err != nil {
		return types.StatementExportCheckResult{}, err
	}
	if count >= 5 {
		return types.StatementExportCheckResult{}, ErrStatementExportLimited
	}
	return types.StatementExportCheckResult{TodayCount: count}, nil
}

func (s Service) ExportRows(ctx context.Context, input StatementExportInput) (types.StatementExportResult, error) {
	if input.AccountBookID <= 0 || input.UserID <= 0 {
		return types.StatementExportResult{}, statement.ErrStatementInvalidInput
	}
	count, err := s.getTodayExportCount(input.UserID)
	if err != nil {
		return types.StatementExportResult{}, err
	}
	if count >= 5 {
		return types.StatementExportResult{}, ErrStatementExportLimited
	}
	if err := s.setTodayExportCount(input.UserID, count+1); err != nil {
		return types.StatementExportResult{}, err
	}

	end := time.Now()
	start := end.AddDate(0, -1, 0)
	switch strings.TrimSpace(input.Range) {
	case "3months":
		start = end.AddDate(0, -3, 0)
	case "all":
		start = end.AddDate(-100, 0, 0)
	case "1month", "":
	default:
		start = end.AddDate(0, -1, 0)
	}

	rows, err := s.queryRepo.ListExportRows(ctx, repo.StatementExportFilter{
		AccountBookID: input.AccountBookID,
		StartDate:     start,
		EndDate:       end,
		Limit:         3000,
	})
	if err != nil {
		return types.StatementExportResult{}, err
	}

	views, err := helperservice.AssembleRows(ctx, s.queryRepo, input.AccountBookID, rows)
	if err != nil {
		return types.StatementExportResult{}, err
	}
	items := make([]StatementExportRowItem, 0, len(rows))
	for _, row := range views {
		items = append(items, StatementExportRowItem{
			Category:       row.CategoryName,
			ParentCategory: row.CategoryParentName,
			Type:           row.Type,
			TypeName:       statementTypeCNForExport(row.Type),
			Asset:          row.AssetName,
			Description:    row.Description,
			Amount:         row.Amount,
			CreatedAt:      row.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      row.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return types.StatementExportResult{Rows: items}, nil
}

func (s Service) ExportExcelFile(ctx context.Context, input StatementExportInput) ([]byte, error) {
	result, err := s.ExportRows(ctx, input)
	if err != nil {
		return nil, err
	}
	return s.renderer.Render(result.Rows)
}

func statementExportCacheKey(userID int64, now time.Time) string {
	return fmt.Sprintf("export_excel_limit_%d_%s", userID, now.Format("20060102"))
}

func (s Service) getTodayExportCount(userID int64) (int, error) {
	if s.cache == nil {
		return 0, nil
	}
	key := statementExportCacheKey(userID, time.Now())
	raw, ok := s.cache.Get(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	count, err := parseCounterInt(strings.TrimSpace(raw))
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (s Service) setTodayExportCount(userID int64, count int) error {
	if s.cache == nil {
		return nil
	}
	if count < 0 {
		count = 0
	}
	key := statementExportCacheKey(userID, time.Now())
	ttl := 24 * time.Hour
	s.cache.Set(key, fmt.Sprintf("%d", count), ttl)
	return nil
}

func parseCounterInt(v string) (int, error) {
	n64, err := parseInt64(v)
	if err != nil {
		return 0, err
	}
	return int(n64), nil
}

func statementTypeCNForExport(statementType string) string {
	switch strings.TrimSpace(statementType) {
	case "income":
		return "收入"
	case "expend":
		return "支出"
	case "transfer":
		return "转账"
	case "repayment":
		return "还款"
	case "loan_in":
		return "借入"
	case "loan_out":
		return "借出"
	case "reimburse":
		return "报销"
	case "payment_proxy":
		return "代付"
	default:
		return strings.TrimSpace(statementType)
	}
}

func parseInt64(v string) (int64, error) {
	var n int64
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return 0, statement.ErrStatementInvalidInput
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}
