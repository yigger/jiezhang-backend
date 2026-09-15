package finance

import (
	"context"
	"fmt"
	"sort"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	statementservice "github.com/yigger/jiezhang-backend/internal/service/statement"
	"github.com/yigger/jiezhang-backend/internal/types"
)

type FinanceService struct {
	financeRepo    repo.FinanceRepository
	statementQuery repo.StatementQueryRepository
	rowMapper      statementservice.RowMapper
}

func NewFinanceService(financeRepo repo.FinanceRepository, statementQuery repo.StatementQueryRepository, rowMapper statementservice.RowMapper) FinanceService {
	return FinanceService{
		financeRepo:    financeRepo,
		statementQuery: statementQuery,
		rowMapper:      rowMapper,
	}
}

func (s FinanceService) GetWallet(ctx context.Context, accountBookID int64) (types.WalletResponse, error) {
	assets, err := s.financeRepo.ListAssets(ctx, accountBookID)
	if err != nil {
		return types.WalletResponse{}, err
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

	parents := make([]tablemodel.Asset, 0)
	assetsByParent := make(map[int64][]types.WalletChildAsset, len(assets))
	parentAmount := make(map[int64]float64, len(assets))

	for _, a := range assets {
		if a.ParentID == 0 {
			parents = append(parents, a)
			continue
		}
		parentAmount[a.ParentID] += a.Amount
		assetsByParent[a.ParentID] = append(assetsByParent[a.ParentID], types.WalletChildAsset{
			ID:       a.ID,
			Name:     a.Name,
			Amount:   financeMoneyFormat(a.Amount),
			IconPath: s.rowMapper.BuildPublicURL(a.IconPath),
		})
	}

	list := make([]types.WalletParent, 0, len(parents))
	for _, p := range parents {
		children := assetsByParent[p.ID]
		if children == nil {
			children = []types.WalletChildAsset{}
		}
		list = append(list, types.WalletParent{
			Name:   p.Name,
			Amount: financeMoneyFormat(parentAmount[p.ID]),
			Childs: children,
		})
	}

	receivableTypes := []string{"reimburse", "payment_proxy", "loan_out"}
	payableTypes := []string{"loan_in"}

	receivableTotal, err := s.financeRepo.SumStatementAmountByTypes(ctx, accountBookID, receivableTypes)
	if err != nil {
		return types.WalletResponse{}, err
	}
	payableTotal, err := s.financeRepo.SumStatementAmountByTypes(ctx, accountBookID, payableTypes)
	if err != nil {
		return types.WalletResponse{}, err
	}

	specialRows, err := s.financeRepo.ListSpecialCategoryByTypes(ctx, append(receivableTypes, payableTypes...))
	if err != nil {
		return types.WalletResponse{}, err
	}
	specialCategoryMap := make(map[string]int64, len(specialRows))
	for _, row := range specialRows {
		specialCategoryMap[row.SpecialType] = row.ID
	}

	receivableItems, err := s.listStatementTypeAmounts(ctx, accountBookID, receivableTypes, specialCategoryMap)
	if err != nil {
		return types.WalletResponse{}, err
	}
	payableItems, err := s.listStatementTypeAmounts(ctx, accountBookID, payableTypes, specialCategoryMap)
	if err != nil {
		return types.WalletResponse{}, err
	}

	return types.WalletResponse{
		Header: types.WalletHeader{
			TotalAsset:     financeMoneyFormat(totalAsset),
			NetWorth:       financeMoneyFormat(totalAsset - totalLiability),
			TotalLiability: financeMoneyFormat(totalLiability),
		},
		List:          list,
		AmountVisible: true,
		Receivables: types.WalletTypeSummary{
			Name:   "应收款项",
			Amount: financeMoneyFormat(receivableTotal),
			Childs: receivableItems,
		},
		Payables: types.WalletTypeSummary{
			Name:   "应付款项",
			Amount: financeMoneyFormat(payableTotal),
			Childs: payableItems,
		},
	}, nil
}

func (s FinanceService) GetWalletInformation(ctx context.Context, accountBookID int64, assetID int64) (types.WalletInformation, error) {
	asset, err := s.financeRepo.FindAssetByID(ctx, assetID, accountBookID)
	if err != nil {
		return types.WalletInformation{}, err
	}
	sums, err := s.financeRepo.SumIncomeExpendByAsset(ctx, accountBookID, assetID)
	if err != nil {
		return types.WalletInformation{}, err
	}

	return types.WalletInformation{
		Name:          asset.Name,
		Income:        financeMoneyFormat(sums.Income),
		Expend:        financeMoneyFormat(sums.Expend),
		Surplus:       financeMoneyFormat(asset.Amount),
		SourceSurplus: asset.Amount,
	}, nil
}

func (s FinanceService) GetWalletTimeline(ctx context.Context, accountBookID int64, assetID int64) (types.WalletTimelineResponse, error) {
	rows, err := s.financeRepo.ListAssetTimeline(ctx, accountBookID, assetID)
	if err != nil {
		return types.WalletTimelineResponse{}, err
	}

	items := make([]types.WalletTimelineItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, types.WalletTimelineItem{
			ExpendAmount: row.ExpendAmount,
			IncomeAmount: row.IncomeAmount,
			Surplus:      row.IncomeAmount - row.ExpendAmount,
			Year:         row.Year,
			Month:        row.Month,
			Hidden:       1,
		})
	}

	return types.WalletTimelineResponse{Status: 200, Data: items}, nil
}

func (s FinanceService) GetWalletStatementList(ctx context.Context, accountBookID int64, assetID int64, year int, month int) ([]types.StatementListItem, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	rows, err := helperservice.AssembleListRows(ctx, s.statementQuery, repo.StatementListFilter{
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

	items := make([]types.StatementListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}
	return items, nil
}

func (s FinanceService) listStatementTypeAmounts(ctx context.Context, accountBookID int64, statementTypes []string, specialCategoryMap map[string]int64) ([]types.WalletTypeAmountItem, error) {
	rows, err := s.financeRepo.ListStatementSumsByTypes(ctx, accountBookID, statementTypes)
	if err != nil {
		return nil, err
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].StatementType < rows[j].StatementType
	})

	items := make([]types.WalletTypeAmountItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, types.WalletTypeAmountItem{
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
