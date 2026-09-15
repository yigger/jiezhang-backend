package budget

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
)

var (
	ErrBudgetInvalidInput = errors.New("budget invalid input")
)

type BudgetService struct {
	repo       repo.BudgetRepository
	urlBuilder repo.URLBuilder
}

func NewBudgetService(repo repo.BudgetRepository, urlBuilder repo.URLBuilder) BudgetService {
	return BudgetService{repo: repo, urlBuilder: urlBuilder}
}

type BudgetUpdateInput struct {
	Type       string
	Amount     string
	CategoryID *int64
}

func (s BudgetService) Summary(ctx context.Context, accountBookID int64, year int, month int) (types.BudgetSummary, error) {
	allBudget, err := s.repo.GetAccountBookBudget(ctx, accountBookID)
	if err != nil {
		return types.BudgetSummary{}, err
	}
	used, err := s.repo.SumExpendByMonth(ctx, accountBookID, year, month)
	if err != nil {
		return types.BudgetSummary{}, err
	}

	amountStr := "0"
	surplus := 0.0
	if allBudget != 0 {
		amountStr = budgetMoneyFormat(allBudget)
		surplus = allBudget - used
	}

	return types.BudgetSummary{
		SourceAmount: allBudget,
		Amount:       amountStr,
		Used:         used,
		Surplus:      surplus,
	}, nil
}

func (s BudgetService) ParentList(ctx context.Context, accountBookID int64, year int, month int) ([]types.BudgetParentItem, error) {
	categories, err := s.repo.ListExpendParentCategories(ctx, accountBookID)
	if err != nil {
		return nil, err
	}

	items := make([]types.BudgetParentItem, 0, len(categories))
	for _, c := range categories {
		childIDs, err := s.repo.ListChildCategoryIDsByParentID(ctx, accountBookID, c.ID)
		if err != nil {
			return nil, err
		}
		ids := []int64{c.ID}
		ids = append(ids, childIDs...)
		usedAmount, err := s.repo.SumStatementsByCategoryIDsAndMonth(ctx, accountBookID, ids, year, month)
		if err != nil {
			return nil, err
		}
		usePercent, surplusPercent := calcBudgetPercent(usedAmount, c.Budget)
		items = append(items, types.BudgetParentItem{
			ID:             c.ID,
			Name:           c.Name,
			IconPath:       s.buildPublicURL(c.IconPath),
			SourceAmount:   c.Budget,
			Amount:         budgetMoneyFormatOrUnset(c.Budget),
			Surplus:        budgetMoneyFormat(zeroIfUnset(c.Budget, c.Budget-usedAmount)),
			UsedAmount:     usedAmount,
			UsePercent:     usePercent,
			SurplusPercent: surplusPercent,
		})
	}
	return items, nil
}

func (s BudgetService) CategoryDetail(ctx context.Context, accountBookID int64, categoryID int64, year int, month int) (types.BudgetCategoryDetail, error) {
	category, err := s.repo.FindCategoryByID(ctx, accountBookID, categoryID)
	if err != nil {
		return types.BudgetCategoryDetail{}, err
	}

	childs, err := s.repo.ListChildCategoriesByParentID(ctx, accountBookID, category.ID)
	if err != nil {
		return types.BudgetCategoryDetail{}, err
	}

	childIDs, err := s.repo.ListChildCategoryIDsByParentID(ctx, accountBookID, category.ID)
	if err != nil {
		return types.BudgetCategoryDetail{}, err
	}
	ids := []int64{category.ID}
	ids = append(ids, childIDs...)
	usedAmount, err := s.repo.SumStatementsByCategoryIDsAndMonth(ctx, accountBookID, ids, year, month)
	if err != nil {
		return types.BudgetCategoryDetail{}, err
	}
	usePercent, surplusPercent := calcBudgetPercent(usedAmount, category.Budget)

	childItems := make([]types.BudgetChildItem, 0, len(childs))
	for _, c := range childs {
		childUsed, sumErr := s.repo.SumStatementsByCategoryIDsAndMonth(ctx, accountBookID, []int64{c.ID}, year, month)
		if sumErr != nil {
			return types.BudgetCategoryDetail{}, sumErr
		}
		childUsePercent, childSurplusPercent := calcBudgetPercent(childUsed, c.Budget)
		childItems = append(childItems, types.BudgetChildItem{
			ID:             c.ID,
			Name:           c.Name,
			IconPath:       s.buildPublicURL(c.IconPath),
			SourceAmount:   c.Budget,
			Amount:         budgetMoneyFormat(c.Budget),
			Surplus:        budgetMoneyFormat(zeroIfUnset(c.Budget, c.Budget-childUsed)),
			UsePercent:     childUsePercent,
			UsedAmount:     childUsed,
			SurplusPercent: childSurplusPercent,
		})
	}

	return types.BudgetCategoryDetail{
		Root: types.BudgetRoot{
			ID:             category.ID,
			Name:           category.Name,
			IconPath:       s.buildPublicURL(category.IconPath),
			SourceAmount:   category.Budget,
			UsedAmount:     budgetMoneyFormat(usedAmount),
			Amount:         budgetMoneyFormat(category.Budget),
			Surplus:        budgetMoneyFormat(zeroIfUnset(category.Budget, category.Budget-usedAmount)),
			UsePercent:     usePercent,
			SurplusPercent: surplusPercent,
		},
		Childs: childItems,
	}, nil
}

func (s BudgetService) UpdateAmount(ctx context.Context, accountBookID int64, input BudgetUpdateInput) error {
	amount, err := strconv.ParseFloat(strings.TrimSpace(input.Amount), 64)
	if err != nil {
		return ErrBudgetInvalidInput
	}
	if amount < 0 {
		return ErrBudgetInvalidInput
	}

	if strings.TrimSpace(input.Type) == "user" {
		categoryAmount, err := s.repo.SumParentCategoryBudget(ctx, accountBookID)
		if err != nil {
			return err
		}
		if amount < categoryAmount {
			return errors.New("总预算必须大于分类预算的总和")
		}
		return s.repo.UpdateAccountBookBudget(ctx, accountBookID, amount)
	}

	categoryID := int64(0)
	if input.CategoryID != nil {
		categoryID = *input.CategoryID
	}
	category, err := s.repo.FindCategoryByID(ctx, accountBookID, categoryID)
	if err != nil {
		return err
	}
	if category.ParentID == 0 {
		childs, err := s.repo.ListChildCategoriesByParentID(ctx, accountBookID, category.ID)
		if err != nil {
			return err
		}
		childBudgetSum := 0.0
		for _, child := range childs {
			childBudgetSum += child.Budget
		}
		if amount < childBudgetSum {
			return errors.New("一级分类预算不能少于二级分类的总和")
		}
	} else {
		parent, err := s.repo.FindCategoryByID(ctx, accountBookID, category.ParentID)
		if err == nil && amount > parent.Budget {
			if err := s.repo.UpdateCategoryBudget(ctx, accountBookID, parent.ID, amount); err != nil {
				return err
			}
		}
	}

	if err := s.repo.UpdateCategoryBudget(ctx, accountBookID, category.ID, amount); err != nil {
		return err
	}
	userBudgetAmount, err := s.repo.SumParentCategoryBudget(ctx, accountBookID)
	if err != nil {
		return err
	}
	accountBookBudget, err := s.repo.GetAccountBookBudget(ctx, accountBookID)
	if err != nil {
		return err
	}
	if userBudgetAmount > accountBookBudget {
		if err := s.repo.UpdateAccountBookBudget(ctx, accountBookID, userBudgetAmount); err != nil {
			return err
		}
	}
	return nil
}

func ResolveBudgetYearMonth(yearRaw string, monthRaw string) (int, int) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if v, err := strconv.Atoi(strings.TrimSpace(yearRaw)); err == nil && v > 0 {
		year = v
	}
	if v, err := strconv.Atoi(strings.TrimSpace(monthRaw)); err == nil && v >= 1 && v <= 12 {
		month = v
	}
	return year, month
}

func calcBudgetPercent(usedAmount float64, budget float64) (int, int) {
	if budget <= 0 {
		return 0, 100
	}
	usePercent := int((usedAmount * 100) / budget)
	if usePercent > 100 {
		usePercent = 100
	}
	if usePercent < 0 {
		usePercent = 0
	}
	return usePercent, 100 - usePercent
}

func zeroIfUnset(budget float64, value float64) float64 {
	if budget == 0 {
		return 0
	}
	return value
}

func budgetMoneyFormat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func budgetMoneyFormatOrUnset(v float64) string {
	if v == 0 {
		return "未设置"
	}
	return budgetMoneyFormat(v)
}

func (s BudgetService) buildPublicURL(raw string) string {
	return s.urlBuilder.BuildPublicURL(raw)
}
