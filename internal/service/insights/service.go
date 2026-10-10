package insights

import (
	"context"
	"errors"
	"fmt"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
	"math"
	"sort"
	"strings"
	"time"
)

var ErrInvalidYear = errors.New("年份无效，不能晚于今年")

type Service struct {
	repo repo.InsightsRepository
	now  func() time.Time
}

func New(r repo.InsightsRepository) *Service { return &Service{repo: r, now: time.Now} }
func NormalizeName(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), ""))
}
func cents(value float64) int64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
		return 0
	}
	return int64(math.Round(value * 100))
}

type group struct {
	item         types.InsightGroup
	ids          map[int64]bool
	categories   map[int64]*types.InsightCategory
	expenseCount int
}

func newGroup(key, name string) *group {
	return &group{item: types.InsightGroup{Key: key, Name: name, IDs: []int64{}, Categories: []types.InsightCategory{}, StatementIDs: []int64{}}, ids: map[int64]bool{}, categories: map[int64]*types.InsightCategory{}}
}
func addAmounts(a *types.InsightAmounts, kind string, amount int64) {
	if kind == "income" {
		a.IncomeCents += amount
	} else if kind == "expend" {
		a.ExpendCents += amount
	}
	a.Count++
	a.BalanceCents = a.IncomeCents - a.ExpendCents
}
func (g *group) add(row repo.InsightStatementRecord, amount int64, id int64) {
	if !g.ids[id] {
		g.ids[id] = true
		g.item.IDs = append(g.item.IDs, id)
	}
	addAmounts(&g.item.InsightAmounts, row.Type, amount)
	g.item.StatementIDs = append(g.item.StatementIDs, row.ID)
	date := row.CreatedAt.Format("2006-01-02")
	if date > g.item.LastDate {
		g.item.LastDate = date
	}
	if row.Type == "expend" {
		g.expenseCount++
		cat := g.categories[row.CategoryID]
		if cat == nil {
			cat = &types.InsightCategory{ID: row.CategoryID, Name: row.CategoryName}
			if cat.Name == "" {
				cat.Name = "未分类"
			}
			g.categories[row.CategoryID] = cat
		}
		cat.AmountCents += amount
		cat.Count++
	}
}
func sortedGroups(groups map[string]*group) []types.InsightGroup {
	result := make([]types.InsightGroup, 0, len(groups))
	for _, g := range groups {
		if g.expenseCount > 0 {
			g.item.AverageCents = int64(math.Round(float64(g.item.ExpendCents) / float64(g.expenseCount)))
		}
		for _, cat := range g.categories {
			g.item.Categories = append(g.item.Categories, *cat)
		}
		sort.Slice(g.item.Categories, func(i, j int) bool {
			a, b := g.item.Categories[i], g.item.Categories[j]
			if a.AmountCents == b.AmountCents {
				return a.ID < b.ID
			}
			return a.AmountCents > b.AmountCents
		})
		sort.Slice(g.item.IDs, func(i, j int) bool { return g.item.IDs[i] < g.item.IDs[j] })
		result = append(result, g.item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ExpendCents == result[j].ExpendCents {
			return result[i].Key < result[j].Key
		}
		return result[i].ExpendCents > result[j].ExpendCents
	})
	return result
}
func (s *Service) Report(ctx context.Context, bookID int64, year int) (types.InsightReport, error) {
	now := s.now()
	if year < 1900 || year > now.Year() {
		return types.InsightReport{}, ErrInvalidYear
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(1, 0, 0)
	if year == now.Year() {
		end = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	}
	rows, err := s.repo.ListRows(ctx, bookID, start, end)
	if err != nil {
		return types.InsightReport{}, err
	}
	return build(rows, year, start, end, now), nil
}
func build(rows []repo.InsightStatementRecord, year int, start, end, now time.Time) types.InsightReport {
	report := types.InsightReport{Year: year, AsOf: now.Format("2006-01-02"), StartDate: start.Format("2006-01-02"), EndDate: end.AddDate(0, 0, -1).Format("2006-01-02"), Months: make([]types.InsightMonth, 12), Merchants: []types.InsightGroup{}, Members: []types.InsightGroup{}, LargeExpenses: []types.InsightLargeExpense{}, RecurringCandidates: []types.InsightRecurringCandidate{}}
	for i := range report.Months {
		month := time.Date(year, time.Month(i+1), 1, 0, 0, 0, 0, now.Location())
		report.Months[i] = types.InsightMonth{Month: i + 1, Started: !month.After(now), Complete: !month.AddDate(0, 1, 0).After(now)}
	}
	merchants, members := map[string]*group{}, map[string]*group{}
	recurring := map[string][]repo.InsightStatementRecord{}
	for _, row := range rows {
		if row.CreatedAt.Before(start) || !row.CreatedAt.Before(end) || (row.Type != "income" && row.Type != "expend") {
			continue
		}
		amount := cents(row.Amount)
		if amount == 0 {
			continue
		}
		addAmounts(&report.InsightAmounts, row.Type, amount)
		addAmounts(&report.Months[int(row.CreatedAt.Month())-1].InsightAmounts, row.Type, amount)
		memberKey := fmt.Sprint(row.UserID)
		if members[memberKey] == nil {
			name := row.MemberName
			if name == "" {
				name = "成员 " + memberKey
			}
			members[memberKey] = newGroup(memberKey, name)
		}
		members[memberKey].add(row, amount, row.UserID)
		if row.Type != "expend" {
			continue
		}
		merchantKey := NormalizeName(row.MerchantName)
		name := strings.TrimSpace(row.MerchantName)
		if merchantKey == "" {
			merchantKey = "unassigned"
			name = "未指定商家"
		} else {
			merchantKey = "name:" + merchantKey
		}
		if merchants[merchantKey] == nil {
			merchants[merchantKey] = newGroup(merchantKey, name)
		}
		var payeeID int64
		if row.PayeeID != nil {
			payeeID = *row.PayeeID
		}
		merchants[merchantKey].add(row, amount, payeeID)
		report.LargeExpenses = append(report.LargeExpenses, types.InsightLargeExpense{ID: row.ID, AmountCents: amount, Date: row.CreatedAt.Format("2006-01-02"), Category: row.CategoryName, Description: row.Description})
		title := NormalizeName(row.Description)
		if title == "" {
			title = NormalizeName(row.MerchantName)
		}
		if title != "" {
			key := fmt.Sprintf("%d:%d:%s", row.CategoryID, row.AssetID, title)
			recurring[key] = append(recurring[key], row)
		}
	}
	report.Merchants = sortedGroups(merchants)
	report.Members = sortedGroups(members)
	sort.Slice(report.LargeExpenses, func(i, j int) bool {
		a, b := report.LargeExpenses[i], report.LargeExpenses[j]
		if a.AmountCents == b.AmountCents {
			return a.ID < b.ID
		}
		return a.AmountCents > b.AmountCents
	})
	if len(report.LargeExpenses) > 3 {
		report.LargeExpenses = report.LargeExpenses[:3]
	}
	for key, items := range recurring {
		byMonth := map[string]bool{}
		for _, item := range items {
			byMonth[item.CreatedAt.Format("2006-01")] = true
		}
		if len(byMonth) < 3 || len(items) != len(byMonth) {
			continue
		}
		sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
		latest := items[0]
		latestCents := cents(latest.Amount)
		minDay, maxDay := 31, 1
		for _, item := range items {
			day := item.CreatedAt.Day()
			if day < minDay {
				minDay = day
			}
			if day > maxDay {
				maxDay = day
			}
		}
		if maxDay-minDay > 5 {
			continue
		}
		consistent := true
		for _, item := range items {
			if math.Abs(float64(cents(item.Amount)-latestCents)) > float64(latestCents)*0.2 {
				consistent = false
				break
			}
		}
		if !consistent {
			continue
		}
		name := strings.TrimSpace(latest.Description)
		if name == "" {
			name = strings.TrimSpace(latest.MerchantName)
		}
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		report.RecurringCandidates = append(report.RecurringCandidates, types.InsightRecurringCandidate{Key: key, Name: name, AmountCents: latestCents, AnnualCents: latestCents * 12, Months: len(byMonth), DueDay: latest.CreatedAt.Day(), CategoryID: latest.CategoryID, AssetID: latest.AssetID, StatementIDs: ids})
	}
	sort.Slice(report.RecurringCandidates, func(i, j int) bool { return report.RecurringCandidates[i].Key < report.RecurringCandidates[j].Key })
	return report
}
