package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service/helper"
)

var (
	ErrCategoryPermissionDenied = errors.New("category permission denied")
	ErrCategoryInvalidInput     = errors.New("category invalid input")
)

type CategoryService struct {
	repo       repository.CategoryRepository
	urlBuilder CategoryURLBuilder
}

type CategoryURLBuilder interface {
	BuildPublicURL(raw string) string
}

func NewCategoryService(repo repository.CategoryRepository, urlBuilder CategoryURLBuilder) CategoryService {
	return CategoryService{repo: repo, urlBuilder: urlBuilder}
}

type CategoryItem struct {
	ID       int64          `json:"id"`
	Name     string         `json:"name"`
	Order    int            `json:"order"`
	IconPath string         `json:"icon_path"`
	ParentID int64          `json:"parent_id"`
	Type     string         `json:"type"`
	Amount   string         `json:"amount,omitempty"`
	IconURL  string         `json:"icon_url,omitempty"`
	Childs   []CategoryItem `json:"childs,omitempty"`
}

type CategoryHeader struct {
	Month      string `json:"month"`
	Year       string `json:"year"`
	All        string `json:"all"`
	ParentName string `json:"parent_name,omitempty"`
}

type CategoryListResponse struct {
	Header     CategoryHeader `json:"header"`
	Categories []CategoryItem `json:"categories"`
}

type CategoryShowResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Order      int    `json:"order"`
	IconPath   string `json:"icon_path"`
	ParentID   int64  `json:"parent_id"`
	Type       string `json:"type"`
	ParentName string `json:"parent_name"`
	IconURL    string `json:"icon_url"`
}

type CategoryWriteInput struct {
	UserID        int64
	AccountBookID int64
	Name          string
	ParentID      int64
	IconPath      string
	Type          string
}

type CategoryStatementsMonthItem struct {
	Year   int                      `json:"year"`
	Month  int                      `json:"month"`
	Childs []CategoryStatementChild `json:"childs"`
}

type CategoryStatementChild struct {
	ID          int64  `json:"id"`
	Day         int    `json:"day"`
	Week        string `json:"week"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	IconPath    string `json:"icon_path"`
	Description string `json:"description"`
	Money       string `json:"money"`
	TimeStr     string `json:"timeStr"`
	Asset       string `json:"asset"`
}

func (s CategoryService) ListByParent(ctx context.Context, accountBookID int64, statementType string, parentID int64) (CategoryListResponse, error) {
	statementType = strings.TrimSpace(statementType)
	rows, err := s.repo.ListByParent(ctx, repository.CategoryListFilter{AccountBookID: accountBookID, Type: statementType}, parentID)
	if err != nil {
		return CategoryListResponse{}, err
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	amountMap := make(map[int64]float64)
	if parentID == 0 {
		amountRows, amountErr := s.repo.ListStatementAmountByParentIDs(ctx, accountBookID, ids)
		if amountErr != nil {
			return CategoryListResponse{}, amountErr
		}
		amountMap = make(map[int64]float64, len(amountRows))
		for _, item := range amountRows {
			amountMap[item.CategoryID] = item.Amount
		}
	} else {
		amountRows, amountErr := s.repo.ListStatementAmountByCategoryIDs(ctx, accountBookID, ids)
		if amountErr != nil {
			return CategoryListResponse{}, amountErr
		}
		amountMap = make(map[int64]float64, len(amountRows))
		for _, item := range amountRows {
			amountMap[item.CategoryID] = item.Amount
		}
	}

	categories := make([]CategoryItem, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, CategoryItem{
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
		return CategoryListResponse{}, err
	}
	yearExpend, err := s.repo.SumStatements(ctx, accountBookID, statementType, sumCategoryIDs, now.Year(), 0)
	if err != nil {
		return CategoryListResponse{}, err
	}
	allExpend, err := s.repo.SumStatements(ctx, accountBookID, statementType, sumCategoryIDs, 0, 0)
	if err != nil {
		return CategoryListResponse{}, err
	}

	res := CategoryListResponse{
		Header: CategoryHeader{
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

func (s CategoryService) ListTree(ctx context.Context, accountBookID int64, statementType string) ([]CategoryItem, error) {
	filter := repository.CategoryListFilter{AccountBookID: accountBookID, Type: strings.TrimSpace(statementType)}
	parents, err := s.repo.ListParents(ctx, filter)
	if err != nil {
		return nil, err
	}
	parentIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentIDs = append(parentIDs, p.ID)
	}
	children, err := s.repo.ListChildrenByParentIDs(ctx, filter, parentIDs)
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

	childrenByParent := make(map[int64][]CategoryItem, len(parentIDs))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], CategoryItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: child.IconPath,
			ParentID: child.ParentID,
			Amount:   "0.00",
			IconURL:  s.buildPublicURL(child.IconPath),
		})
	}

	res := make([]CategoryItem, 0, len(parents))
	for _, parent := range parents {
		childs := childrenByParent[parent.ID]
		if childs == nil {
			childs = []CategoryItem{}
		}
		res = append(res, CategoryItem{
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

func (s CategoryService) Show(ctx context.Context, accountBookID int64, id int64) (CategoryShowResponse, error) {
	row, err := s.repo.FindByID(ctx, accountBookID, id)
	if err != nil {
		return CategoryShowResponse{}, err
	}

	parentName := ""
	if row.ParentID > 0 {
		parent, parentErr := s.repo.FindByID(ctx, accountBookID, row.ParentID)
		if parentErr == nil {
			parentName = parent.Name
		}
	}

	return CategoryShowResponse{
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
	dir := filepath.Join("public", "images", "category")
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return []map[string]string{}, nil
	}
	entries, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
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
	return items, nil
}

func (s CategoryService) ListStatementsByCategory(ctx context.Context, accountBookID int64, categoryID int64) ([]CategoryStatementsMonthItem, error) {
	rows, err := s.repo.ListStatementsByCategory(ctx, accountBookID, categoryID)
	if err != nil {
		return nil, err
	}

	type monthKey struct {
		Year  int
		Month int
	}
	monthMap := make(map[monthKey][]CategoryStatementChild)
	monthOrder := make([]monthKey, 0)

	for _, row := range rows {
		key := monthKey{Year: row.Year, Month: row.Month}
		if _, ok := monthMap[key]; !ok {
			monthOrder = append(monthOrder, key)
		}
		monthMap[key] = append(monthMap[key], CategoryStatementChild{
			ID:          row.ID,
			Day:         row.Day,
			Week:        weekCN(row.CreatedAt),
			Type:        row.Type,
			Category:    row.CategoryName,
			IconPath:    s.buildPublicURL(row.IconPath),
			Description: row.Description,
			Money:       categoryMoneyFormat(row.Amount),
			TimeStr:     row.CreatedAt.Format("01-02 15:04"),
			Asset:       row.AssetName,
		})
	}

	sort.Slice(monthOrder, func(i, j int) bool {
		if monthOrder[i].Year == monthOrder[j].Year {
			return monthOrder[i].Month > monthOrder[j].Month
		}
		return monthOrder[i].Year > monthOrder[j].Year
	})

	res := make([]CategoryStatementsMonthItem, 0, len(monthOrder))
	for _, key := range monthOrder {
		res = append(res, CategoryStatementsMonthItem{
			Year:   key.Year,
			Month:  key.Month,
			Childs: monthMap[key],
		})
	}
	return res, nil
}

func (s CategoryService) normalizeWriteInput(input CategoryWriteInput) (repository.CategoryWriteRecord, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return repository.CategoryWriteRecord{}, ErrCategoryInvalidInput
	}
	statementType := strings.TrimSpace(input.Type)
	if statementType == "" {
		statementType = "expend"
	}
	return repository.CategoryWriteRecord{
		UserID:        input.UserID,
		AccountBookID: input.AccountBookID,
		Name:          name,
		ParentID:      input.ParentID,
		IconPath:      strings.TrimSpace(input.IconPath),
		Type:          statementType,
	}, nil
}

func (s CategoryService) buildPublicURL(raw string) string {
	if s.urlBuilder == nil {
		return raw
	}
	return s.urlBuilder.BuildPublicURL(raw)
}

func categoryMoneyFormat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func weekCN(t time.Time) string {
	return helper.WeekdayCN(t.Weekday())
}
