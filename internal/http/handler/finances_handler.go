package handler

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/yigger/jiezhang-backend/internal/repository"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type FinancesHandler struct {
	db             *gorm.DB
	statementQuery repository.StatementQueryRepository
	rowMapper      statementdto.RowMapper
}

func NewFinancesHandler(db *gorm.DB, statementQuery repository.StatementQueryRepository, rowMapper statementdto.RowMapper) FinancesHandler {
	return FinancesHandler{
		db:             db,
		statementQuery: statementQuery,
		rowMapper:      rowMapper,
	}
}

func (h FinancesHandler) Wallet(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}

	var assets []walletAssetRow
	if err := h.db.WithContext(c.Request.Context()).
		Table("assets").
		Where("account_book_id = ?", accountBook.ID).
		Order("parent_id ASC, id ASC").
		Scan(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load assets"})
		return
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

	assetsByParent := make(map[int64][]walletChildAsset, len(assets))
	parentAmount := make(map[int64]float64, len(assets))
	parents := make([]walletAssetRow, 0)
	for _, a := range assets {
		if a.ParentID == 0 {
			parents = append(parents, a)
			continue
		}
		parentAmount[a.ParentID] += a.Amount
		assetsByParent[a.ParentID] = append(assetsByParent[a.ParentID], walletChildAsset{
			ID:       a.ID,
			Name:     a.Name,
			Amount:   financeMoneyFormat(a.Amount),
			IconPath: h.rowMapper.BuildPublicURL(a.IconPath),
		})
	}

	list := make([]walletParentAsset, 0, len(parents))
	for _, p := range parents {
		children := assetsByParent[p.ID]
		if children == nil {
			children = []walletChildAsset{}
		}
		list = append(list, walletParentAsset{
			Name:   p.Name,
			Amount: financeMoneyFormat(parentAmount[p.ID]),
			Childs: children,
		})
	}

	receivableTypes := []string{"reimburse", "payment_proxy", "loan_out"}
	payableTypes := []string{"loan_in"}

	receivableTotal, err := h.sumStatementAmountByTypes(c, accountBook.ID, receivableTypes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load receivables"})
		return
	}
	payableTotal, err := h.sumStatementAmountByTypes(c, accountBook.ID, payableTypes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load payables"})
		return
	}

	specialCategoryMap, err := h.loadSpecialCategoryIDMap(c, accountBook.ID, append(receivableTypes, payableTypes...))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load category map"})
		return
	}

	receivableChilds, err := h.sumStatementAmountGroupedByType(c, accountBook.ID, receivableTypes, specialCategoryMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load receivable details"})
		return
	}
	payableChilds, err := h.sumStatementAmountGroupedByType(c, accountBook.ID, payableTypes, specialCategoryMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load payable details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"header": gin.H{
			"total_asset":     financeMoneyFormat(totalAsset),
			"net_worth":       financeMoneyFormat(totalAsset - totalLiability),
			"total_liability": financeMoneyFormat(totalLiability),
		},
		"list":           list,
		"amount_visible": true,
		"receivables": gin.H{
			"name":   "应收款项",
			"amount": financeMoneyFormat(receivableTotal),
			"childs": receivableChilds,
		},
		"payables": gin.H{
			"name":   "应付款项",
			"amount": financeMoneyFormat(payableTotal),
			"childs": payableChilds,
		},
	})
}

func (h FinancesHandler) WalletInformation(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := parseInt64Query(c, "asset_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	var asset walletAssetRow
	err = h.db.WithContext(c.Request.Context()).
		Table("assets").
		Where("id = ? AND account_book_id = ?", assetID, accountBook.ID).
		Take(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"status": 404, "msg": "asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load asset"})
		return
	}

	var sums struct {
		Income float64 `gorm:"column:income"`
		Expend float64 `gorm:"column:expend"`
	}
	if err := h.db.WithContext(c.Request.Context()).
		Table("statements").
		Select("COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income, COALESCE(SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END), 0) AS expend").
		Where("account_book_id = ? AND asset_id = ?", accountBook.ID, assetID).
		Scan(&sums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load statement sums"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":           asset.Name,
		"income":         financeMoneyFormat(sums.Income),
		"expend":         financeMoneyFormat(sums.Expend),
		"surplus":        financeMoneyFormat(asset.Amount),
		"source_surplus": asset.Amount,
	})
}

func (h FinancesHandler) WalletTimeline(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := parseInt64Query(c, "asset_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	var rows []walletTimelineRow
	if err := h.db.WithContext(c.Request.Context()).
		Table("statements").
		Select(strings.Join([]string{
			"year",
			"month",
			"COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income_amount",
			"COALESCE(SUM(CASE WHEN type = 'expend' THEN amount ELSE 0 END), 0) AS expend_amount",
		}, ", ")).
		Where("account_book_id = ? AND asset_id = ?", accountBook.ID, assetID).
		Group("year, month").
		Order("year DESC, month DESC").
		Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load timeline"})
		return
	}

	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"expend_amount": row.ExpendAmount,
			"income_amount": row.IncomeAmount,
			"surplus":       row.IncomeAmount - row.ExpendAmount,
			"year":          row.Year,
			"month":         row.Month,
			"hidden":        1,
		})
	}
	c.JSON(http.StatusOK, gin.H{"status": 200, "data": items})
}

func (h FinancesHandler) WalletStatementList(c *gin.Context) {
	accountBook, ok := requireAccountBook(c)
	if !ok {
		return
	}
	assetID, err := parseInt64Query(c, "asset_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}
	year, err := parseIntQuery(c, "year")
	if err != nil || year <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	month, err := parseIntQuery(c, "month")
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
		return
	}

	loc := time.Local
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	rows, err := h.statementQuery.ListRowsWithRelations(c.Request.Context(), repository.StatementListFilter{
		AccountBookID: accountBook.ID,
		AssetID:       assetID,
		StartDate:     &startDate,
		EndDate:       &endDate,
		OrderBy:       "created_at",
		Limit:         1000,
		Offset:        0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load statement list"})
		return
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, h.rowMapper.ToListItem(row))
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

type walletAssetRow struct {
	ID       int64   `gorm:"column:id"`
	Name     string  `gorm:"column:name"`
	Amount   float64 `gorm:"column:amount"`
	ParentID int64   `gorm:"column:parent_id"`
	IconPath string  `gorm:"column:icon_path"`
	Type     string  `gorm:"column:type"`
}

type walletChildAsset struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Amount   string `json:"amount"`
	IconPath string `json:"icon_path"`
}

type walletParentAsset struct {
	Name   string             `json:"name"`
	Amount string             `json:"amount"`
	Childs []walletChildAsset `json:"childs"`
}

type walletTypeSumRow struct {
	StatementType string  `gorm:"column:statement_type"`
	Amount        float64 `gorm:"column:amount"`
}

type walletTimelineRow struct {
	Year         int     `gorm:"column:year"`
	Month        int     `gorm:"column:month"`
	IncomeAmount float64 `gorm:"column:income_amount"`
	ExpendAmount float64 `gorm:"column:expend_amount"`
}

func (h FinancesHandler) sumStatementAmountByTypes(c *gin.Context, accountBookID int64, statementTypes []string) (float64, error) {
	var row struct {
		Amount float64 `gorm:"column:amount"`
	}
	err := h.db.WithContext(c.Request.Context()).
		Table("statements").
		Select("COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND type IN ?", accountBookID, statementTypes).
		Scan(&row).Error
	return row.Amount, err
}

func (h FinancesHandler) loadSpecialCategoryIDMap(c *gin.Context, accountBookID int64, statementTypes []string) (map[string]int64, error) {
	type row struct {
		SpecialType string `gorm:"column:special_type"`
		ID          int64  `gorm:"column:id"`
	}
	rows := make([]row, 0)
	err := h.db.WithContext(c.Request.Context()).
		Table("categories").
		Select("special_type, id").
		Where("account_book_id = ? AND special_type IN ?", accountBookID, statementTypes).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	res := make(map[string]int64, len(rows))
	for _, r := range rows {
		res[r.SpecialType] = r.ID
	}
	return res, nil
}

func (h FinancesHandler) sumStatementAmountGroupedByType(c *gin.Context, accountBookID int64, statementTypes []string, specialCategoryMap map[string]int64) ([]gin.H, error) {
	rows := make([]walletTypeSumRow, 0)
	if err := h.db.WithContext(c.Request.Context()).
		Table("statements").
		Select("type AS statement_type, COALESCE(SUM(amount), 0) AS amount").
		Where("account_book_id = ? AND type IN ?", accountBookID, statementTypes).
		Group("type").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].StatementType < rows[j].StatementType
	})

	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, gin.H{
			"category_id": specialCategoryMap[row.StatementType],
			"name":        statementTypeCN(row.StatementType),
			"amount":      financeMoneyFormat(row.Amount),
		})
	}
	return items, nil
}

func parseInt64Query(c *gin.Context, key string) (int64, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, errInvalidParam(key)
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidParam(key)
	}
	return id, nil
}

func parseIntQuery(c *gin.Context, key string) (int, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, errInvalidParam(key)
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, errInvalidParam(key)
	}
	return v, nil
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
