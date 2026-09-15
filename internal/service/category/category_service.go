package category

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	"github.com/yigger/jiezhang-backend/internal/types"
)

var (
	ErrCategoryPermissionDenied = errors.New("category permission denied")
	ErrCategoryInvalidInput     = errors.New("category invalid input")
)

var (
	categoryIconsOnce  sync.Once
	categoryIconsCache []map[string]string
)

type CategoryService struct {
	repo       repo.CategoryRepository
	urlBuilder repo.URLBuilder
}

func NewCategoryService(repo repo.CategoryRepository, urlBuilder repo.URLBuilder) CategoryService {
	return CategoryService{repo: repo, urlBuilder: urlBuilder}
}

type CategoryWriteInput struct {
	UserID        int64
	AccountBookID int64
	Name          string
	ParentID      int64
	IconPath      string
	Type          string
}

func (s CategoryService) ListByParent(ctx context.Context, accountBookID int64, statementType string, parentID int64) (types.CategoryListResponse, error) {
	statementType = strings.TrimSpace(statementType)
	rows, err := s.repo.ListByParent(ctx, accountBookID, statementType, parentID)
	if err != nil {
		return types.CategoryListResponse{}, err
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	amountMap := make(map[int64]float64)
	if parentID == 0 {
		amountRows, amountErr := s.repo.ListStatementAmountByParentIDs(ctx, accountBookID, ids)
		if amountErr != nil {
			return types.CategoryListResponse{}, amountErr
		}
		amountMap = make(map[int64]float64, len(amountRows))
		for _, item := range amountRows {
			amountMap[item.CategoryID] = item.Amount
		}
	} else {
		amountRows, amountErr := s.repo.ListStatementAmountByCategoryIDs(ctx, accountBookID, ids)
		if amountErr != nil {
			return types.CategoryListResponse{}, amountErr
		}
		amountMap = make(map[int64]float64, len(amountRows))
		for _, item := range amountRows {
			amountMap[item.CategoryID] = item.Amount
		}
	}

	categories := make([]types.CategoryItem, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, types.CategoryItem{
			ID:       row.ID,
			Name:     row.Name,
			Order:    row.Order,
			IconPath: row.IconPath,
			ParentID: row.ParentID,
			Type:     row.Type,
			Amount:   categoryMoneyFormat(amountMap[row.ID]),
			IconURL:  s.buildPublicURL(row.IconPath),
		})
	}

	now := time.Now()
	sumCategoryIDs := ids
	if parentID == 0 {
		sumCategoryIDs = nil
	}

	monthExpend, err := s.repo.SumStatements(ctx, accountBookID, statementType, sumCategoryIDs, now.Year(), int(now.Month()))
	if err != nil {
		return types.CategoryListResponse{}, err
	}
	yearExpend, err := s.repo.SumStatements(ctx, accountBookID, statementType, sumCategoryIDs, now.Year(), 0)
	if err != nil {
		return types.CategoryListResponse{}, err
	}
	allExpend, err := s.repo.SumStatements(ctx, accountBookID, statementType, sumCategoryIDs, 0, 0)
	if err != nil {
		return types.CategoryListResponse{}, err
	}

	res := types.CategoryListResponse{
		Header: types.CategoryHeader{
			Month: categoryMoneyFormat(monthExpend),
			Year:  categoryMoneyFormat(yearExpend),
			All:   categoryMoneyFormat(allExpend),
		},
		Categories: categories,
	}

	if parentID > 0 {
		parent, parentErr := s.repo.FindByID(ctx, accountBookID, parentID)
		if parentErr == nil {
			res.Header.ParentName = parent.Name
		}
	}

	return res, nil
}

func (s CategoryService) ListTree(ctx context.Context, accountBookID int64, statementType string) ([]types.CategoryItem, error) {
	statementType = strings.TrimSpace(statementType)
	parents, err := s.repo.ListParents(ctx, accountBookID, statementType)
	if err != nil {
		return nil, err
	}
	parentIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentIDs = append(parentIDs, p.ID)
	}
	children, err := s.repo.ListChildrenByParentIDs(ctx, accountBookID, statementType, parentIDs)
	if err != nil {
		return nil, err
	}

	amountRows, err := s.repo.ListStatementAmountByParentIDs(ctx, accountBookID, parentIDs)
	if err != nil {
		return nil, err
	}
	amountMap := make(map[int64]float64, len(amountRows))
	for _, item := range amountRows {
		amountMap[item.CategoryID] = item.Amount
	}

	childrenByParent := make(map[int64][]types.CategoryItem, len(parentIDs))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], types.CategoryItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: child.IconPath,
			ParentID: child.ParentID,
			Amount:   "0.00",
			IconURL:  s.buildPublicURL(child.IconPath),
		})
	}

	res := make([]types.CategoryItem, 0, len(parents))
	for _, parent := range parents {
		childs := childrenByParent[parent.ID]
		if childs == nil {
			childs = []types.CategoryItem{}
		}
		res = append(res, types.CategoryItem{
			ID:       parent.ID,
			Name:     parent.Name,
			IconPath: parent.IconPath,
			ParentID: 0,
			Type:     strings.TrimSpace(statementType),
			Amount:   categoryMoneyFormat(amountMap[parent.ID]),
			IconURL:  s.buildPublicURL(parent.IconPath),
			Childs:   childs,
		})
	}
	return res, nil
}

func (s CategoryService) Show(ctx context.Context, accountBookID int64, id int64) (types.CategoryShowResponse, error) {
	row, err := s.repo.FindByID(ctx, accountBookID, id)
	if err != nil {
		return types.CategoryShowResponse{}, err
	}

	parentName := ""
	if row.ParentID > 0 {
		parent, parentErr := s.repo.FindByID(ctx, accountBookID, row.ParentID)
		if parentErr == nil {
			parentName = parent.Name
		}
	}

	return types.CategoryShowResponse{
		ID:         row.ID,
		Name:       row.Name,
		Order:      row.Order,
		IconPath:   row.IconPath,
		ParentID:   row.ParentID,
		Type:       row.Type,
		ParentName: parentName,
		IconURL:    s.buildPublicURL(row.IconPath),
	}, nil
}

func (s CategoryService) Create(ctx context.Context, input CategoryWriteInput) error {
	record, err := s.normalizeWriteInput(input)
	if err != nil {
		return err
	}
	ok, err := s.repo.CanAdmin(ctx, input.AccountBookID, input.UserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCategoryPermissionDenied
	}
	_, err = s.repo.Create(ctx, record)
	return err
}

func (s CategoryService) Update(ctx context.Context, id int64, input CategoryWriteInput) error {
	record, err := s.normalizeWriteInput(input)
	if err != nil {
		return err
	}
	ok, err := s.repo.CanAdmin(ctx, input.AccountBookID, input.UserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCategoryPermissionDenied
	}
	return s.repo.UpdateByID(ctx, id, input.AccountBookID, record)
}

func (s CategoryService) Delete(ctx context.Context, id int64, accountBookID int64, userID int64) error {
	ok, err := s.repo.CanAdmin(ctx, accountBookID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCategoryPermissionDenied
	}
	return s.repo.DeleteByID(ctx, id, accountBookID)
}

func (s CategoryService) ListCategoryIcons() ([]map[string]string, error) {
	var loadErr error
	categoryIconsOnce.Do(func() {
		dir := filepath.Join("public", "images", "category")
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			categoryIconsCache = []map[string]string{}
			return
		}
		entries, err := filepath.Glob(filepath.Join(dir, "*"))
		if err != nil {
			loadErr = err
			return
		}
		sort.Strings(entries)
		items := make([]map[string]string, 0, len(entries))
		for _, path := range entries {
			name := filepath.Base(path)
			if strings.TrimSpace(name) == "" {
				continue
			}
			raw := "/images/category/" + name
			items = append(items, map[string]string{
				"id":  raw,
				"url": s.buildPublicURL(raw),
			})
		}
		categoryIconsCache = items
	})
	if loadErr != nil {
		return nil, loadErr
	}
	return categoryIconsCache, nil
}

func (s CategoryService) ListStatementsByCategory(ctx context.Context, accountBookID int64, categoryID int64) ([]types.CategoryStatementsMonthItem, error) {
	rows, err := s.repo.ListStatementsByCategory(ctx, accountBookID, categoryID)
	if err != nil {
		return nil, err
	}

	categoryRows, err := s.repo.ListByIDs(ctx, []int64{categoryID})
	if err != nil {
		return nil, err
	}
	var category tablemodel.Category
	if len(categoryRows) > 0 {
		category = categoryRows[0]
	}
	assetIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		assetIDs = append(assetIDs, row.AssetID)
	}
	assets, err := s.repo.BatchGetAssets(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	assetNames := make(map[int64]string, len(assets))
	for _, asset := range assets {
		assetNames[asset.ID] = asset.Name
	}
	type monthKey struct {
		Year  int
		Month int
	}
	monthMap := make(map[monthKey][]types.CategoryStatementChild)
	monthOrder := make([]monthKey, 0)

	for _, row := range rows {
		key := monthKey{Year: row.Year, Month: row.Month}
		if _, ok := monthMap[key]; !ok {
			monthOrder = append(monthOrder, key)
		}
		monthMap[key] = append(monthMap[key], types.CategoryStatementChild{
			ID:          row.ID,
			Day:         row.Day,
			Week:        weekCN(row.CreatedAt),
			Type:        row.Type,
			Category:    category.Name,
			IconPath:    s.buildPublicURL(category.IconPath),
			Description: row.Description,
			Money:       categoryMoneyFormat(row.Amount),
			TimeStr:     row.CreatedAt.Format("01-02 15:04"),
			Asset:       assetNames[row.AssetID],
		})
	}

	sort.Slice(monthOrder, func(i, j int) bool {
		if monthOrder[i].Year == monthOrder[j].Year {
			return monthOrder[i].Month > monthOrder[j].Month
		}
		return monthOrder[i].Year > monthOrder[j].Year
	})

	res := make([]types.CategoryStatementsMonthItem, 0, len(monthOrder))
	for _, key := range monthOrder {
		res = append(res, types.CategoryStatementsMonthItem{
			Year:   key.Year,
			Month:  key.Month,
			Childs: monthMap[key],
		})
	}
	return res, nil
}

func (s CategoryService) normalizeWriteInput(input CategoryWriteInput) (tablemodel.Category, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return tablemodel.Category{}, ErrCategoryInvalidInput
	}
	statementType := strings.TrimSpace(input.Type)
	if statementType == "" {
		statementType = "expend"
	}
	return tablemodel.Category{
		UserID:        input.UserID,
		AccountBookID: input.AccountBookID,
		Name:          name,
		ParentID:      input.ParentID,
		IconPath:      strings.TrimSpace(input.IconPath),
		Type:          statementType,
	}, nil
}

func (s CategoryService) buildPublicURL(raw string) string {
	return s.urlBuilder.BuildPublicURL(raw)
}

func categoryMoneyFormat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func weekCN(t time.Time) string {
	return helperservice.WeekdayCN(t.Weekday())
}
