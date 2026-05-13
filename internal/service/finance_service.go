package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type FinanceService struct {
	financeRepo    repository.FinanceRepository
	statementQuery repository.StatementQueryRepository
	rowMapper      statementdto.RowMapper
}

func NewFinanceService(financeRepo repository.FinanceRepository, statementQuery repository.StatementQueryRepository, rowMapper statementdto.RowMapper) FinanceService {
	return FinanceService{
		financeRepo:    financeRepo,
		statementQuery: statementQuery,
		rowMapper:      rowMapper,
	}
}

type WalletResponse struct {
	Header        WalletHeader      `json:"header"`
	List          []WalletParent    `json:"list"`
	AmountVisible bool              `json:"amount_visible"`
	Receivables   WalletTypeSummary `json:"receivables"`
	Payables      WalletTypeSummary `json:"payables"`
}

type WalletHeader struct {
	TotalAsset     string `json:"total_asset"`
	NetWorth       string `json:"net_worth"`
	TotalLiability string `json:"total_liability"`
}

type WalletParent struct {
	Name   string             `json:"name"`
	Amount string             `json:"amount"`
	Childs []WalletChildAsset `json:"childs"`
}

type WalletChildAsset struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	IconPath string `json:"icon_path"`
}

type WalletTypeSummary struct {
	Name   string                 `json:"name"`
	Amount string                 `json:"amount"`
	Childs []WalletTypeAmountItem `json:"childs"`
}

type WalletTypeAmountItem struct {
	CategoryID int64  `json:"category_id"`
	Name       string `json:"name"`
	Amount     string `json:"amount"`
}

type WalletInformation struct {
	Name          string  `json:"name"`
	Income        string  `json:"income"`
	Expend        string  `json:"expend"`
	Surplus       string  `json:"surplus"`
	SourceSurplus float64 `json:"source_surplus"`
}

type WalletTimelineResponse struct {
	Status int                  `json:"status"`
	Data   []WalletTimelineItem `json:"data"`
}

type WalletTimelineItem struct {
	ExpendAmount float64 `json:"expend_amount"`
	IncomeAmount float64 `json:"income_amount"`
	Surplus      float64 `json:"surplus"`
	Year         int     `json:"year"`
	Month        int     `json:"month"`
	Hidden       int     `json:"hidden"`
}

func (s FinanceService) GetWallet(ctx context.Context, accountBookID int64) (WalletResponse, error) {
	assets, err := s.financeRepo.ListAssets(ctx, accountBookID)
	if err != nil {
		return WalletResponse{}, err
	}

	var totalAsset float64
	var totalLiability float64
	for _, a := range assets {
		if a.ParentID <= 0 {
			continue
		}
		switch a.Type {
		case "deposit":
			totalAsset += a.Amount
		case "debt":
			totalLiability += a.Amount
		}
	}

	parents := make([]repository.FinanceAssetRecord, 0)
	assetsByParent := make(map[int64][]WalletChildAsset, len(assets))
	parentAmount := make(map[int64]float64, len(assets))

	for _, a := range assets {
		if a.ParentID == 0 {
			parents = append(parents, a)
			continue
		}
		parentAmount[a.ParentID] += a.Amount
		assetsByParent[a.ParentID] = append(assetsByParent[a.ParentID], WalletChildAsset{
			ID:       a.ID,
			Name:     a.Name,
			Amount:   financeMoneyFormat(a.Amount),
			IconPath: s.rowMapper.BuildPublicURL(a.IconPath),
		})
	}

	list := make([]WalletParent, 0, len(parents))
	for _, p := range parents {
		children := assetsByParent[p.ID]
		if children == nil {
			children = []WalletChildAsset{}
		}
		list = append(list, WalletParent{
			Name:   p.Name,
			Amount: financeMoneyFormat(parentAmount[p.ID]),
			Childs: children,
		})
	}

	receivableTypes := []string{"reimburse", "payment_proxy", "loan_out"}
	payableTypes := []string{"loan_in"}

	receivableTotal, err := s.financeRepo.SumStatementAmountByTypes(ctx, accountBookID, receivableTypes)
	if err != nil {
		return WalletResponse{}, err
	}
	payableTotal, err := s.financeRepo.SumStatementAmountByTypes(ctx, accountBookID, payableTypes)
	if err != nil {
		return WalletResponse{}, err
	}

	specialRows, err := s.financeRepo.ListSpecialCategoryByTypes(ctx, accountBookID, append(receivableTypes, payableTypes...))
	if err != nil {
		return WalletResponse{}, err
	}
	specialCategoryMap := make(map[string]int64, len(specialRows))
	for _, row := range specialRows {
		specialCategoryMap[row.SpecialType] = row.ID
	}

	receivableItems, err := s.listStatementTypeAmounts(ctx, accountBookID, receivableTypes, specialCategoryMap)
	if err != nil {
		return WalletResponse{}, err
	}
	payableItems, err := s.listStatementTypeAmounts(ctx, accountBookID, payableTypes, specialCategoryMap)
	if err != nil {
		return WalletResponse{}, err
	}

	return WalletResponse{
		Header: WalletHeader{
			TotalAsset:     financeMoneyFormat(totalAsset),
			NetWorth:       financeMoneyFormat(totalAsset - totalLiability),
			TotalLiability: financeMoneyFormat(totalLiability),
		},
		List:          list,
		AmountVisible: true,
		Receivables: WalletTypeSummary{
			Name:   "应收款项",
			Amount: financeMoneyFormat(receivableTotal),
			Childs: receivableItems,
		},
		Payables: WalletTypeSummary{
			Name:   "应付款项",
			Amount: financeMoneyFormat(payableTotal),
			Childs: payableItems,
		},
	}, nil
}

func (s FinanceService) GetWalletInformation(ctx context.Context, accountBookID int64, assetID int64) (WalletInformation, error) {
	asset, err := s.financeRepo.FindAssetByID(ctx, assetID, accountBookID)
	if err != nil {
		return WalletInformation{}, err
	}
	sums, err := s.financeRepo.SumIncomeExpendByAsset(ctx, accountBookID, assetID)
	if err != nil {
		return WalletInformation{}, err
	}

	return WalletInformation{
		Name:          asset.Name,
		Income:        financeMoneyFormat(sums.Income),
		Expend:        financeMoneyFormat(sums.Expend),
		Surplus:       financeMoneyFormat(asset.Amount),
		SourceSurplus: asset.Amount,
	}, nil
}

func (s FinanceService) GetWalletTimeline(ctx context.Context, accountBookID int64, assetID int64) (WalletTimelineResponse, error) {
	rows, err := s.financeRepo.ListAssetTimeline(ctx, accountBookID, assetID)
	if err != nil {
		return WalletTimelineResponse{}, err
	}

	items := make([]WalletTimelineItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, WalletTimelineItem{
			ExpendAmount: row.ExpendAmount,
			IncomeAmount: row.IncomeAmount,
			Surplus:      row.IncomeAmount - row.ExpendAmount,
			Year:         row.Year,
			Month:        row.Month,
			Hidden:       1,
		})
	}

	return WalletTimelineResponse{Status: 200, Data: items}, nil
}

func (s FinanceService) GetWalletStatementList(ctx context.Context, accountBookID int64, assetID int64, year int, month int) ([]statementdto.ListItem, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	rows, err := s.statementQuery.ListRowsWithRelations(ctx, repository.StatementListFilter{
		AccountBookID: accountBookID,
		AssetID:       assetID,
		StartDate:     &startDate,
		EndDate:       &endDate,
		OrderBy:       "created_at",
		Limit:         1000,
		Offset:        0,
	})
	if err != nil {
		return nil, err
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}
	return items, nil
}

func (s FinanceService) listStatementTypeAmounts(ctx context.Context, accountBookID int64, statementTypes []string, specialCategoryMap map[string]int64) ([]WalletTypeAmountItem, error) {
	rows, err := s.financeRepo.ListStatementSumsByTypes(ctx, accountBookID, statementTypes)
	if err != nil {
		return nil, err
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].StatementType < rows[j].StatementType
	})

	items := make([]WalletTypeAmountItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, WalletTypeAmountItem{
			CategoryID: specialCategoryMap[row.StatementType],
			Name:       statementTypeCN(row.StatementType),
			Amount:     financeMoneyFormat(row.Amount),
		})
	}
	return items, nil
}

func statementTypeCN(statementType string) string {
	switch statementType {
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
		return statementType
	}
}

func financeMoneyFormat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
