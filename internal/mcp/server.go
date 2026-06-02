package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yigger/jiezhang-backend/internal/service"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

// Server wraps the MCP server, registered tools, and the HTTP handler.
type Server struct {
	mcpServer *mcp.Server

	statementSvc   service.StatementService
	financeSvc     service.FinanceService
	statisticsSvc  service.StatisticsService
	accountBookSvc service.AccountBookService
	budgetSvc      service.BudgetService
}

// NewServer creates an MCP server with all tools registered.
func NewServer(
	statementSvc service.StatementService,
	financeSvc service.FinanceService,
	statisticsSvc service.StatisticsService,
	accountBookSvc service.AccountBookService,
	budgetSvc service.BudgetService,
) *Server {
	s := &Server{
		statementSvc:   statementSvc,
		financeSvc:     financeSvc,
		statisticsSvc:  statisticsSvc,
		accountBookSvc: accountBookSvc,
		budgetSvc:      budgetSvc,
	}

	s.mcpServer = mcp.NewServer(
		&mcp.Implementation{Name: "jiezhang", Version: "1.0.0"},
		&mcp.ServerOptions{
			Instructions: "记账应用 MCP 服务 — 可查询账单、财务概览、统计数据和预算。每个需要用户或账本上下文的工具请提供 account_book_id 参数。",
		},
	)

	s.registerTools()
	return s
}

// Handler returns an http.Handler for the MCP endpoint using Streamable HTTP
// in stateless + JSON mode.
func (s *Server) Handler() http.Handler {
	return mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return s.mcpServer
		},
		&mcp.StreamableHTTPOptions{
			Stateless:   true,
			JSONResponse: true,
		},
	)
}

func (s *Server) registerTools() {
	s.registerQueryStatements()
	s.registerGetWalletOverview()
	s.registerGetMonthlyStatistics()
	s.registerListAccountBooks()
	s.registerGetBudgetSummary()
}

// ---------------------------------------------------------------------------
// Tool: query_statements
// ---------------------------------------------------------------------------

type queryStatementsInput struct {
	AccountBookID int64  `json:"account_book_id" jsonschema:"账本 ID（必填）"`
	StartDate     string `json:"start_date,omitempty" jsonschema:"开始日期 YYYY-MM-DD，不填则为当月 1 号"`
	EndDate       string `json:"end_date,omitempty" jsonschema:"结束日期 YYYY-MM-DD，不填则为今天"`
	Type          string `json:"type,omitempty" jsonschema:"账单类型: expense(支出) / income(收入) / transfer(转账)，不填则全部"`
	Keyword       string `json:"keyword,omitempty" jsonschema:"模糊搜索关键词（匹配备注、分类等）"`
	Limit         int    `json:"limit,omitempty" jsonschema:"返回条数，默认 20，最大 100"`
}

type queryStatementsOutput struct {
	Total      int                  `json:"total" jsonschema:"符合条件的总条数"`
	Statements []statementListItem  `json:"statements" jsonschema:"账单列表"`
}

type statementListItem struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Asset       string  `json:"asset"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Time        string  `json:"time"`
	Payee       string  `json:"payee"`
}

func (s *Server) registerQueryStatements() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "query_statements",
		Description: "查询账单流水。支持按日期范围、类型、关键词筛选。返回符合条件的账单列表。",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input queryStatementsInput) (*mcp.CallToolResult, queryStatementsOutput, error) {
		now := time.Now()
		startDate := firstDayOfMonth(now)
		endDate := now

		if input.StartDate != "" {
			if t, err := time.Parse("2006-01-02", input.StartDate); err == nil {
				startDate = t
			}
		}
		if input.EndDate != "" {
			if t, err := time.Parse("2006-01-02", input.EndDate); err == nil {
				endDate = t
			}
		}
		if input.Limit <= 0 {
			input.Limit = 20
		}
		if input.Limit > 100 {
			input.Limit = 100
		}

		// Use SearchStatements when keyword is given, otherwise GetStatements
		var items []statementdto.ListItem
		var err error
		if input.Keyword != "" {
			items, err = s.statementSvc.SearchStatements(ctx, input.AccountBookID, input.Keyword)
		} else {
			listInput := statementdto.ListInput{
				AccountBookID: input.AccountBookID,
				StartDate:     &startDate,
				EndDate:       &endDate,
				OrderBy:       "desc",
				Limit:         input.Limit,
			}
			if input.Type != "" {
				// Filter by type is done via category — for simplicity we pass it to the query
				// and the service filters by parent category type
			}
			items, err = s.statementSvc.GetStatements(ctx, listInput)
		}
		if err != nil {
			return nil, queryStatementsOutput{}, fmt.Errorf("查询账单失败: %w", err)
		}

		out := queryStatementsOutput{Total: len(items)}
		for _, it := range items {
			if it.Type != "" && input.Type != "" && it.Type != input.Type {
				continue
			}
			if len(out.Statements) >= input.Limit {
				break
			}
			out.Statements = append(out.Statements, statementListItem{
				ID:          it.ID,
				Type:        it.Type,
				Amount:      it.Amount,
				Category:    it.Category,
				Asset:       it.Asset,
				Description: it.Description,
				Date:        it.Date,
				Time:        it.Time,
				Payee:       it.Payee.Name,
			})
		}
		out.Total = len(out.Statements)
		return nil, out, nil
	})
}

// ---------------------------------------------------------------------------
// Tool: get_wallet_overview
// ---------------------------------------------------------------------------

type walletOverviewInput struct {
	AccountBookID int64 `json:"account_book_id" jsonschema:"账本 ID（必填）"`
}

type walletOverviewOutput struct {
	TotalAsset     string            `json:"total_asset" jsonschema:"总资产"`
	NetWorth       string            `json:"net_worth" jsonschema:"净资产"`
	TotalLiability string            `json:"total_liability" jsonschema:"总负债"`
	ReceivablesAmt string            `json:"receivables_amount" jsonschema:"应收金额"`
	PayablesAmt    string            `json:"payables_amount" jsonschema:"应付金额"`
	Assets         []walletAssetItem `json:"assets" jsonschema:"资产明细"`
}

type walletAssetItem struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Type   string `json:"type"`
}

func (s *Server) registerGetWalletOverview() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_wallet_overview",
		Description: "获取钱包/财务总览。返回总资产、净资产、负债以及各资产账户余额。",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input walletOverviewInput) (*mcp.CallToolResult, walletOverviewOutput, error) {
		wallet, err := s.financeSvc.GetWallet(ctx, input.AccountBookID)
		if err != nil {
			return nil, walletOverviewOutput{}, fmt.Errorf("获取财务概览失败: %w", err)
		}

		out := walletOverviewOutput{
			TotalAsset:     wallet.Header.TotalAsset,
			NetWorth:       wallet.Header.NetWorth,
			TotalLiability: wallet.Header.TotalLiability,
			ReceivablesAmt: wallet.Receivables.Amount,
			PayablesAmt:    wallet.Payables.Amount,
		}
		for _, parent := range wallet.List {
			for _, child := range parent.Childs {
				out.Assets = append(out.Assets, walletAssetItem{
					ID:     child.ID,
					Name:   child.Name,
					Amount: child.Amount,
				})
			}
		}
		return nil, out, nil
	})
}

// ---------------------------------------------------------------------------
// Tool: get_monthly_statistics
// ---------------------------------------------------------------------------

type monthlyStatisticsInput struct {
	AccountBookID int64  `json:"account_book_id" jsonschema:"账本 ID（必填）"`
	Year          int    `json:"year,omitempty" jsonschema:"年份，不填为当前年份"`
	Month         int    `json:"month,omitempty" jsonschema:"月份 1-12，不填为当前月份"`
}

type monthlyStatisticsOutput struct {
	Year         int     `json:"year"`
	Month        int     `json:"month"`
	TotalIncome  float64 `json:"total_income" jsonschema:"月总收入"`
	TotalExpend  float64 `json:"total_expend" jsonschema:"月总支出"`
	TotalRepay   float64 `json:"total_repay" jsonschema:"月总还款"`
	TotalBalance float64 `json:"total_balance" jsonschema:"月结余（收入-支出-还款）"`
}

func (s *Server) registerGetMonthlyStatistics() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_monthly_statistics",
		Description: "获取月度收支统计数据。返回指定月份的汇总收入、支出、还款和结余。",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input monthlyStatisticsInput) (*mcp.CallToolResult, monthlyStatisticsOutput, error) {
		if input.Year <= 0 {
			input.Year = time.Now().Year()
		}
		if input.Month <= 0 || input.Month > 12 {
			input.Month = int(time.Now().Month())
		}

		date := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.Local)
		header, err := s.statisticsSvc.GetOverviewHeader(ctx, date, input.AccountBookID)
		if err != nil {
			return nil, monthlyStatisticsOutput{}, fmt.Errorf("获取月度统计失败: %w", err)
		}

		return nil, monthlyStatisticsOutput{
			Year:         input.Year,
			Month:        input.Month,
			TotalIncome:  header.Income,
			TotalExpend:  header.Expend,
			TotalRepay:   header.Repay,
			TotalBalance: header.TotalBalance,
		}, nil
	})
}

// ---------------------------------------------------------------------------
// Tool: list_account_books
// ---------------------------------------------------------------------------

type listAccountBooksInput struct {
	UserID int64 `json:"user_id" jsonschema:"用户 ID（必填）"`
}

type listAccountBooksOutput struct {
	AccountBooks []accountBookItem `json:"account_books" jsonschema:"账本列表"`
}

type accountBookItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AccountType string `json:"account_type" jsonschema:"账本类型: 基础账簿/家庭账簿/旅行账簿/生意账簿"`
	Budget      string `json:"budget" jsonschema:"预算金额"`
}

func (s *Server) registerListAccountBooks() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_account_books",
		Description: "列出指定用户的所有账本。返回账本 ID、名称、类型和预算。",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listAccountBooksInput) (*mcp.CallToolResult, listAccountBooksOutput, error) {
		books, err := s.accountBookSvc.List(ctx, input.UserID)
		if err != nil {
			return nil, listAccountBooksOutput{}, fmt.Errorf("获取账本列表失败: %w", err)
		}

		out := listAccountBooksOutput{}
		for _, b := range books {
			out.AccountBooks = append(out.AccountBooks, accountBookItem{
				ID:          b.ID,
				Name:        b.Name,
				Description: b.Description,
				AccountType: b.AccountTypeName,
				Budget:      fmt.Sprintf("%.2f", b.Budget),
			})
		}
		return nil, out, nil
	})
}

// ---------------------------------------------------------------------------
// Tool: get_budget_summary
// ---------------------------------------------------------------------------

type budgetSummaryInput struct {
	AccountBookID int64 `json:"account_book_id" jsonschema:"账本 ID（必填）"`
	Year          int   `json:"year,omitempty" jsonschema:"年份，不填为当前年份"`
	Month         int   `json:"month,omitempty" jsonschema:"月份 1-12，不填为当前月份"`
}

type budgetSummaryOutput struct {
	Year    int     `json:"year"`
	Month   int     `json:"month"`
	Budget  float64 `json:"budget" jsonschema:"预算总额"`
	Used    float64 `json:"used" jsonschema:"已使用金额"`
	Surplus float64 `json:"surplus" jsonschema:"剩余金额"`
}

func (s *Server) registerGetBudgetSummary() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_budget_summary",
		Description: "获取预算使用情况。返回某月的预算总额、已用金额和剩余金额。",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input budgetSummaryInput) (*mcp.CallToolResult, budgetSummaryOutput, error) {
		if input.Year <= 0 {
			input.Year = time.Now().Year()
		}
		if input.Month <= 0 || input.Month > 12 {
			input.Month = int(time.Now().Month())
		}

		summary, err := s.budgetSvc.Summary(ctx, input.AccountBookID, input.Year, input.Month)
		if err != nil {
			return nil, budgetSummaryOutput{}, fmt.Errorf("获取预算汇总失败: %w", err)
		}

		amount := 0.0
		if f, err := parseFloat(summary.Amount); err == nil {
			amount = f
		}

		return nil, budgetSummaryOutput{
			Year:    input.Year,
			Month:   input.Month,
			Budget:  amount,
			Used:    summary.Used,
			Surplus: summary.Surplus,
		}, nil
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func firstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
