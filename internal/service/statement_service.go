package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	statementdto "github.com/yigger/jiezhang-backend/internal/service/statement"
)

type ValidateError struct {
	Message string
}

func (e ValidateError) Error() string {
	return fmt.Sprintf("validate error: %s", e.Message)
}

type StatementService struct {
	statementRepo repository.StatementRepository
	queryRepo     repository.StatementQueryRepository
	categoryRepo  repository.CategoryRepository
	assetRepo     repository.AssetRepository
	rowMapper     statementdto.RowMapper
}

func NewStatementService(statementRepo repository.StatementRepository, queryRepo repository.StatementQueryRepository, categoryRepo repository.CategoryRepository, assetRepo repository.AssetRepository, rowMapper statementdto.RowMapper) StatementService {
	return StatementService{
		statementRepo: statementRepo,
		queryRepo:     queryRepo,
		categoryRepo:  categoryRepo,
		assetRepo:     assetRepo,
		rowMapper:     rowMapper,
	}
}

var (
	ErrStatementPermissionDenied = errors.New("statement permission denied")
	ErrStatementInvalidInput     = errors.New("statement invalid input")
)

// 账单列表
func (s StatementService) GetStatements(ctx context.Context, input statementdto.ListInput) ([]statementdto.ListItem, error) {
	filter := repository.StatementListFilter{
		UserID:            input.UserID,
		AccountBookID:     input.AccountBookID,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		ParentCategoryIDs: input.ParentCategoryIDs,
		ExceptIDs:         input.ExceptIDs,
		OrderBy:           input.OrderBy,
		Limit:             input.Limit,
		Offset:            input.Offset,
	}
	rows, err := s.queryRepo.ListRowsWithRelations(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s StatementService) SearchStatements(ctx context.Context, accountBookID int64, keyword string) ([]statementdto.ListItem, error) {
	filter := repository.StatementListFilter{
		AccountBookID: accountBookID,
		Keyword:       keyword,
		OrderBy:       "created_at desc",
	}
	rows, err := s.queryRepo.ListRowsWithRelations(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]statementdto.ListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s StatementService) CreateStatement(ctx context.Context, input statementdto.WriteInput) (statementdto.ListItem, error) {
	record, err := normalizeStatementWriteInput(input)
	if err != nil {
		return statementdto.ListItem{}, err
	}

	statementID, err := s.statementRepo.Create(ctx, record)
	if err != nil {
		return statementdto.ListItem{}, err
	}

	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, statementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	return s.rowMapper.ToListItem(row), nil
}

func (s StatementService) UpdateStatement(ctx context.Context, input statementdto.UpdateInput) (statementdto.ListItem, error) {
	ownerID, err := s.statementRepo.GetOwnerID(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	if ownerID != input.UserID {
		return statementdto.ListItem{}, ErrStatementPermissionDenied
	}

	currentRow, err := s.queryRepo.GetRowByIDWithRelations(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}

	merged := s.mergeStatementPatch(currentRow, input)
	record, err := normalizeStatementWriteInput(merged)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	if err := s.statementRepo.UpdateByID(ctx, input.StatementID, input.AccountBookID, record); err != nil {
		return statementdto.ListItem{}, err
	}
	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return statementdto.ListItem{}, err
	}
	return s.rowMapper.ToListItem(row), nil
}

func (s StatementService) DeleteStatement(ctx context.Context, statementID int64, userID int64, accountBookID int64) error {
	ownerID, err := s.statementRepo.GetOwnerID(ctx, statementID, accountBookID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return ErrStatementPermissionDenied
	}
	return s.statementRepo.DeleteByID(ctx, statementID, accountBookID)
}

func normalizeStatementWriteInput(input statementdto.WriteInput) (repository.StatementWriteRecord, error) {
	statementType := strings.TrimSpace(input.Type)
	if statementType == "" {
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid statement type"}
	}
	if input.Amount <= 0 {
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid amount"}
	}

	occurredAt, err := parseStatementDateTime(input.Date, input.Time)
	if err != nil {
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid date or time"}
	}

	assetID := input.AssetID
	targetAssetID := int64PtrOrNil(input.ToAssetID)
	if statementType == "transfer" || statementType == "repayment" {
		assetID = input.FromAssetID
		if assetID <= 0 || input.ToAssetID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid asset ID"}
		}
		targetAssetID = int64PtrOrNil(input.ToAssetID)
	}

	switch statementType {
	case "expend", "income":
		// 只有收入和支出需要检验分类和资产ID
		if assetID <= 0 || input.CategoryID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid asset or category ID"}
		}
	case "transfer", "repayment":
		// 转账和还款需要检验转入和转出资产的 ID
		fromAssetID := input.FromAssetID
		toAssetID := input.ToAssetID
		if fromAssetID <= 0 || toAssetID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid from or to asset ID"}
		}
	case "loan_in", "loan_out", "reimburse", "payment_proxy":
		// 借贷、报销、代付需要检验资产ID，分类ID可以为空
		if assetID <= 0 {
			return repository.StatementWriteRecord{}, ValidateError{Message: "invalid asset ID"}
		}
	default:
		return repository.StatementWriteRecord{}, ValidateError{Message: "invalid statement type"}
	}

	return repository.StatementWriteRecord{
		UserID:        input.UserID,
		AccountBookID: input.AccountBookID,
		Type:          statementType,
		Amount:        input.Amount,
		Description:   input.Description,
		Mood:          input.Mood,
		CategoryID:    input.CategoryID,
		AssetID:       assetID,
		TargetAssetID: targetAssetID,
		PayeeID:       int64PtrOrNil(input.PayeeID),
		TargetObject:  input.TargetObject,
		Location:      input.Location,
		Nation:        input.Nation,
		Province:      input.Province,
		City:          input.City,
		District:      input.District,
		Street:        input.Street,
		OccurredAt:    occurredAt,
	}, nil
}

func parseStatementDateTime(dateStr string, timeStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)
	if dateStr == "" {
		now := time.Now()
		dateStr = now.Format("2006-01-02")
	}
	if timeStr == "" {
		timeStr = "00:00:00"
	}

	layouts := []string{"2006-01-02 15:04:05", "2006-01-02 15:04"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, dateStr+" "+timeStr, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ValidateError{Message: "invalid date or time"}
}

func int64PtrOrNil(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	n := v
	return &n
}

func (s StatementService) mergeStatementPatch(current repository.StatementListRowRecord, update statementdto.UpdateInput) statementdto.WriteInput {
	input := statementdto.WriteInput{
		StatementID:   update.StatementID,
		UserID:        update.UserID,
		AccountBookID: update.AccountBookID,
		Type:          current.Type,
		Amount:        current.Amount,
		Description:   current.Description,
		Mood:          current.Mood,
		CategoryID:    current.CategoryID,
		AssetID:       current.AssetID,
		FromAssetID:   current.AssetID,
		TargetObject:  current.TargetObject,
		Location:      current.Location,
		Nation:        current.Nation,
		Province:      current.Province,
		City:          current.City,
		District:      current.District,
		Street:        current.Street,
		Date:          current.CreatedAt.Format("2006-01-02"),
		Time:          current.CreatedAt.Format("15:04:05"),
	}
	if current.TargetAssetID > 0 {
		input.ToAssetID = current.TargetAssetID
	}
	if current.PayeeID > 0 {
		input.PayeeID = current.PayeeID
	}

	p := update.Patch
	if p.Type != nil {
		input.Type = strings.TrimSpace(*p.Type)
	}
	if p.Amount != nil {
		input.Amount = *p.Amount
	}
	if p.Description != nil {
		input.Description = *p.Description
	}
	if p.Mood != nil {
		input.Mood = *p.Mood
	}
	if p.CategoryID != nil {
		input.CategoryID = *p.CategoryID
	}
	if p.AssetID != nil {
		input.AssetID = *p.AssetID
	}
	if p.FromAssetID != nil {
		input.FromAssetID = *p.FromAssetID
	}
	if p.ToAssetID != nil {
		input.ToAssetID = *p.ToAssetID
	}
	if p.PayeeID != nil {
		input.PayeeID = *p.PayeeID
	}
	if p.TargetObject != nil {
		input.TargetObject = *p.TargetObject
	}
	if p.Location != nil {
		input.Location = *p.Location
	}
	if p.Nation != nil {
		input.Nation = *p.Nation
	}
	if p.Province != nil {
		input.Province = *p.Province
	}
	if p.City != nil {
		input.City = *p.City
	}
	if p.District != nil {
		input.District = *p.District
	}
	if p.Street != nil {
		input.Street = *p.Street
	}
	if p.Date != nil {
		input.Date = strings.TrimSpace(*p.Date)
	}
	if p.Time != nil {
		input.Time = strings.TrimSpace(*p.Time)
	}

	return input
}

type GetCategoriesInput struct {
	AccountBookID int64
	Type          string
}

type StatementFrequentCategoryItem struct {
	ID       int64                        `json:"id"`
	Name     string                       `json:"name"`
	IconPath string                       `json:"icon_path"`
	Parent   *StatementCategoryParentItem `json:"parent"`
}

type StatementCategoryParentItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementCategoryChildItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type StatementCategoryTreeItem struct {
	ID       int64                        `json:"id"`
	Name     string                       `json:"name"`
	IconPath string                       `json:"icon_path"`
	Childs   []StatementCategoryChildItem `json:"childs"`
}

type StatementCategoriesResult struct {
	Frequent   []StatementFrequentCategoryItem `json:"frequent"`
	Categories []StatementCategoryTreeItem     `json:"categories"`
}

func (s StatementService) GetCategories(ctx context.Context, input GetCategoriesInput) (StatementCategoriesResult, error) {
	filter := repository.CategoryListFilter{
		AccountBookID: input.AccountBookID,
		Type:          input.Type,
	}
	parents, err := s.categoryRepo.ListParents(ctx, filter)
	if err != nil {
		return StatementCategoriesResult{}, err
	}
	parentIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentIDs = append(parentIDs, p.ID)
	}

	children, err := s.categoryRepo.ListChildrenByParentIDs(ctx, filter, parentIDs)
	if err != nil {
		return StatementCategoriesResult{}, err
	}
	frequents, err := s.categoryRepo.ListFrequentChildren(ctx, filter, 10)
	if err != nil {
		return StatementCategoriesResult{}, err
	}

	childrenByParent := make(map[int64][]StatementCategoryChildItem, len(parents))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], StatementCategoryChildItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: s.rowMapper.BuildPublicURL(child.IconPath),
		})
	}

	categories := make([]StatementCategoryTreeItem, 0, len(parents))
	for _, p := range parents {
		childs := childrenByParent[p.ID]
		if childs == nil {
			childs = []StatementCategoryChildItem{}
		}
		categories = append(categories, StatementCategoryTreeItem{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: s.rowMapper.BuildPublicURL(p.IconPath),
			Childs:   childs,
		})
	}

	frequentItems := make([]StatementFrequentCategoryItem, 0, len(frequents))
	for _, f := range frequents {
		var parent *StatementCategoryParentItem
		if f.HasParent {
			parent = &StatementCategoryParentItem{
				ID:   f.ParentID,
				Name: f.ParentName,
			}
		}
		frequentItems = append(frequentItems, StatementFrequentCategoryItem{
			ID:       f.ID,
			Name:     f.Name,
			IconPath: s.rowMapper.BuildPublicURL(f.IconPath),
			Parent:   parent,
		})
	}

	return StatementCategoriesResult{
		Frequent:   frequentItems,
		Categories: categories,
	}, nil
}

type StatementFrequentAssetItem struct {
	ID       int64                     `json:"id"`
	Name     string                    `json:"name"`
	IconPath string                    `json:"icon_path"`
	Parent   *StatementAssetParentItem `json:"parent"`
}

type StatementAssetParentItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatementAssetChildItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IconPath string `json:"icon_path"`
}

type StatementAssetTreeItem struct {
	ID       int64                     `json:"id"`
	Name     string                    `json:"name"`
	IconPath string                    `json:"icon_path"`
	Childs   []StatementAssetChildItem `json:"childs"`
}

type StatementAssetsResult struct {
	Frequent   []StatementFrequentAssetItem `json:"frequent"`
	Categories []StatementAssetTreeItem     `json:"categories"`
}

type StatementDefaultCategoryAssetItem struct {
	CategoryName string `json:"category_name"`
	AssetName    string `json:"asset_name"`
	CategoryID   int64  `json:"category_id"`
	AssetID      int64  `json:"asset_id"`
}

func (s StatementService) GetAssets(ctx context.Context, input GetCategoriesInput) (StatementAssetsResult, error) {
	filter := repository.AssetFilter{
		AccountBookID: input.AccountBookID,
		Type:          input.Type,
	}

	var assetResult []StatementAssetTreeItem
	parents, err := s.assetRepo.ListParents(ctx, filter)
	if err != nil {
		return StatementAssetsResult{}, err
	}
	parentsIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentsIDs = append(parentsIDs, p.ID)
	}

	children, err := s.assetRepo.ListChildrenByParentIDs(ctx, filter, parentsIDs)
	if err != nil {
		return StatementAssetsResult{}, err
	}
	childrenByParent := make(map[int64][]StatementAssetChildItem, len(parents))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], StatementAssetChildItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: s.rowMapper.BuildPublicURL(child.IconPath),
		})
	}

	for _, p := range parents {
		childs := childrenByParent[p.ID]
		if childs == nil {
			childs = []StatementAssetChildItem{}
		}
		assetResult = append(assetResult, StatementAssetTreeItem{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: s.rowMapper.BuildPublicURL(p.IconPath),
			Childs:   childs,
		})
	}

	frequentResult := make([]StatementFrequentAssetItem, 0)
	frequent, err := s.assetRepo.ListFrequentChildren(ctx, filter, 10)
	if err != nil {
		return StatementAssetsResult{}, err
	}
	for _, f := range frequent {
		var parent *StatementAssetParentItem
		if f.HasParent {
			parent = &StatementAssetParentItem{
				ID:   f.ParentID,
				Name: f.ParentName,
			}
		}
		frequentResult = append(frequentResult, StatementFrequentAssetItem{
			ID:       f.ID,
			Name:     f.Name,
			IconPath: s.rowMapper.BuildPublicURL(f.IconPath),
			Parent:   parent,
		})
	}
	return StatementAssetsResult{
		Frequent:   frequentResult,
		Categories: assetResult,
	}, nil
}

func (s StatementService) GetDefaultCategoryAsset(ctx context.Context, input GetCategoriesInput) (*StatementDefaultCategoryAssetItem, error) {
	record, err := s.queryRepo.GetLatestCategoryAssetByType(ctx, input.AccountBookID, input.Type)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	return &StatementDefaultCategoryAssetItem{
		CategoryName: record.CategoryName,
		AssetName:    record.AssetName,
		CategoryID:   record.CategoryID,
		AssetID:      record.AssetID,
	}, nil
}

func (s StatementService) GetTargetObjects(ctx context.Context, input GetCategoriesInput) ([]string, error) {
	return s.queryRepo.ListDistinctTargetObjectsByType(ctx, input.AccountBookID, input.Type)
}

func (s StatementService) CategoriesGuess(ctx context.Context, input GetCategoriesInput) ([]StatementFrequentCategoryItem, error) {
	statementType := input.Type
	if statementType == "" {
		statementType = "expend"
	}

	filter := repository.CategoryGuessFilter{
		AccountBookID: input.AccountBookID,
		StatementType: statementType,
		Now:           time.Now(),
		Limit:         3,
	}
	rows, err := s.categoryRepo.ListGuessedFrequentByStatementType(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]StatementFrequentCategoryItem, 0, len(rows))
	for _, row := range rows {
		var parent *StatementCategoryParentItem
		if row.HasParent {
			parent = &StatementCategoryParentItem{
				ID:   row.ParentID,
				Name: row.ParentName,
			}
		}
		items = append(items, StatementFrequentCategoryItem{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: s.rowMapper.BuildPublicURL(row.IconPath),
			Parent:   parent,
		})
	}
	return items, nil
}

func (s StatementService) AssetsGuess(ctx context.Context, input GetCategoriesInput) ([]StatementFrequentAssetItem, error) {
	rows, err := s.assetRepo.ListGuessedFrequentByStatementTime(ctx, repository.AssetGuessFilter{
		AccountBookID: input.AccountBookID,
		Now:           time.Now(),
		Limit:         3,
	})
	if err != nil {
		return nil, err
	}

	items := make([]StatementFrequentAssetItem, 0, len(rows))
	for _, row := range rows {
		var parent *StatementAssetParentItem
		if row.HasParent {
			parent = &StatementAssetParentItem{
				ID:   row.ParentID,
				Name: row.ParentName,
			}
		}
		items = append(items, StatementFrequentAssetItem{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: s.rowMapper.BuildPublicURL(row.IconPath),
			Parent:   parent,
		})
	}
	return items, nil
}

func (s StatementService) GetStatementByID(ctx context.Context, statementID int64, accountBookID int64) (statementdto.DetailItem, error) {
	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, statementID, accountBookID)
	if err != nil {
		return statementdto.DetailItem{}, err
	}

	return s.rowMapper.ToDetailItem(row), nil
}
