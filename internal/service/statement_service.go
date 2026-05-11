package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service/helper"
)

type StatementService struct {
	statementRepo repository.StatementRepository
	queryRepo     repository.StatementQueryRepository
	categoryRepo  repository.CategoryRepository
	assetRepo     repository.AssetRepository
}

func NewStatementService(statementRepo repository.StatementRepository, queryRepo repository.StatementQueryRepository, categoryRepo repository.CategoryRepository, assetRepo repository.AssetRepository) StatementService {
	return StatementService{statementRepo: statementRepo, queryRepo: queryRepo, categoryRepo: categoryRepo, assetRepo: assetRepo}
}

type StatementListInput struct {
	UserID            int64
	AccountBookID     int64
	StartDate         *time.Time
	EndDate           *time.Time
	ParentCategoryIDs []int64
	ExceptIDs         []int64
	OrderBy           string
	Limit             int
	Offset            int
}

var (
	ErrStatementPermissionDenied = errors.New("statement permission denied")
	ErrStatementInvalidInput     = errors.New("statement invalid input")
)

type StatementWriteInput struct {
	StatementID   int64
	UserID        int64
	AccountBookID int64

	Type         string
	Amount       float64
	Description  string
	Mood         string
	CategoryID   int64
	AssetID      int64
	FromAssetID  int64
	ToAssetID    int64
	PayeeID      int64
	TargetObject string

	Location string
	Nation   string
	Province string
	City     string
	District string
	Street   string

	Date string
	Time string
}

// 账单列表
func (s StatementService) GetStatements(ctx context.Context, input StatementListInput) ([]repository.StatementListItem, error) {
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

	items := make([]repository.StatementListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapStatementRowToItem(row))
	}

	return items, nil
}

func (s StatementService) CreateStatement(ctx context.Context, input StatementWriteInput) (repository.StatementListItem, error) {
	record, err := normalizeStatementWriteInput(input)
	if err != nil {
		return repository.StatementListItem{}, err
	}

	statementID, err := s.statementRepo.Create(ctx, record)
	if err != nil {
		return repository.StatementListItem{}, err
	}

	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, statementID, input.AccountBookID)
	if err != nil {
		return repository.StatementListItem{}, err
	}
	return mapStatementRowToItem(row), nil
}

func (s StatementService) UpdateStatement(ctx context.Context, input StatementWriteInput) (repository.StatementListItem, error) {
	ownerID, err := s.statementRepo.GetOwnerID(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return repository.StatementListItem{}, err
	}
	if ownerID != input.UserID {
		return repository.StatementListItem{}, ErrStatementPermissionDenied
	}

	record, err := normalizeStatementWriteInput(input)
	if err != nil {
		return repository.StatementListItem{}, err
	}
	if err := s.statementRepo.UpdateByID(ctx, input.StatementID, input.AccountBookID, record); err != nil {
		return repository.StatementListItem{}, err
	}
	row, err := s.queryRepo.GetRowByIDWithRelations(ctx, input.StatementID, input.AccountBookID)
	if err != nil {
		return repository.StatementListItem{}, err
	}
	return mapStatementRowToItem(row), nil
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

func normalizeStatementWriteInput(input StatementWriteInput) (repository.StatementWriteRecord, error) {
	statementType := strings.TrimSpace(input.Type)
	if statementType == "" {
		return repository.StatementWriteRecord{}, ErrStatementInvalidInput
	}
	if input.Amount <= 0 {
		return repository.StatementWriteRecord{}, ErrStatementInvalidInput
	}

	occurredAt, err := parseStatementDateTime(input.Date, input.Time)
	if err != nil {
		return repository.StatementWriteRecord{}, ErrStatementInvalidInput
	}

	assetID := input.AssetID
	targetAssetID := int64PtrOrNil(input.ToAssetID)
	if statementType == "transfer" || statementType == "repayment" {
		assetID = input.FromAssetID
		if assetID <= 0 || input.ToAssetID <= 0 {
			return repository.StatementWriteRecord{}, ErrStatementInvalidInput
		}
		targetAssetID = int64PtrOrNil(input.ToAssetID)
	}
	if assetID <= 0 || input.CategoryID <= 0 {
		return repository.StatementWriteRecord{}, ErrStatementInvalidInput
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
	return time.Time{}, ErrStatementInvalidInput
}

func int64PtrOrNil(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	n := v
	return &n
}

func mapStatementRowToItem(row repository.StatementListRowRecord) repository.StatementListItem {
	return repository.StatementListItem{
		StatementBaseItem: repository.StatementBaseItem{
			ID:           row.ID,
			Type:         row.Type,
			Amount:       row.Amount,
			Description:  row.Description,
			Title:        helper.StatementTitle(row),
			TargetObject: row.AssetName,
			Mood:         row.Mood,
			Money:        fmt.Sprintf("%.2f", row.Amount),
			Category:     row.CategoryName,
			IconPath:     row.IconPath,
			Asset:        row.AssetName,
			Date:         row.CreatedAt.Format("2006-01-02"),
			Time:         row.CreatedAt.Format("15:04:05"),
			TimeStr:      row.CreatedAt.Format("01-02 15:04"),
			Week:         helper.WeekdayCN(row.CreatedAt.Weekday()),
			Payee: repository.Payee{
				ID:   row.PayeeID,
				Name: row.PayeeName,
			},
			Remark: row.Remark,
		},
		Location:  row.Location,
		Province:  row.Province,
		City:      row.City,
		Street:    row.Street,
		MonthDay:  row.CreatedAt.Format("01-02"),
		HasPic:    row.HasPic,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
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
			IconPath: child.IconPath,
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
			IconPath: p.IconPath,
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
			IconPath: f.IconPath,
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
			IconPath: child.IconPath,
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
			IconPath: p.IconPath,
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
			IconPath: f.IconPath,
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
			IconPath: row.IconPath,
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
			IconPath: row.IconPath,
			Parent:   parent,
		})
	}
	return items, nil
}
