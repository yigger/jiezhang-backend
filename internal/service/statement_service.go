package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service/helper"
)

type StatementService struct {
	queryRepo    repository.StatementQueryRepository
	categoryRepo repository.CategoryRepository
	assetRepo    repository.AssetRepository
}

func NewStatementService(queryRepo repository.StatementQueryRepository, categoryRepo repository.CategoryRepository, assetRepo repository.AssetRepository) StatementService {
	return StatementService{queryRepo: queryRepo, categoryRepo: categoryRepo, assetRepo: assetRepo}
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
		items = append(items, repository.StatementListItem{
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
		})
	}

	return items, nil
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

func (s StatementService) GetAssets(ctx context.Context, input GetCategoriesInput) ([]StatementAssetsResult, error) {
	filter := repository.AssetFilter{
		AccountBookID: input.AccountBookID,
		Type:          input.Type,
	}

	var assetResult []StatementAssetTreeItem
	parents, err := s.assetRepo.ListParents(ctx, filter)
	if err != nil {
		return nil, err
	}
	parentsIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentsIDs = append(parentsIDs, p.ID)
	}

	children, err := s.assetRepo.ListChildrenByParentIDs(ctx, filter, parentsIDs)
	if err != nil {
		return nil, err
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
		return nil, err
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
	return []StatementAssetsResult{
		{
			Frequent:   frequentResult,
			Categories: assetResult,
		},
	}, nil
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
