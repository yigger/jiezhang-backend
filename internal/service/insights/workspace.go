package insights

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/types"
	"math"
	"sort"
	"time"
)

func (s *StorageService) Workspace(ctx context.Context, book model.AccountBook, user int64) (types.InsightWorkspace, error) {
	projects, e := s.storage.ListProjects(ctx, book.ID)
	if e != nil {
		return types.InsightWorkspace{}, e
	}
	fixed, e := s.storage.ListFixedCosts(ctx, book.ID)
	if e != nil {
		return types.InsightWorkspace{}, e
	}
	annotations, e := s.storage.ListAnnotations(ctx, book.ID)
	if e != nil {
		return types.InsightWorkspace{}, e
	}
	snapshots, e := s.storage.ListPortfolioSnapshots(ctx, book.ID)
	if e != nil {
		return types.InsightWorkspace{}, e
	}
	members, e := s.storage.ListMembers(ctx, book.ID)
	if e != nil {
		return types.InsightWorkspace{}, e
	}
	now := s.now()
	rows, e := s.rows.ListRows(ctx, book.ID, time.Date(1900, 1, 1, 0, 0, 0, 0, now.Location()), time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()))
	if e != nil {
		return types.InsightWorkspace{}, e
	}
	return assembleWorkspace(book, user, projects, fixed, annotations, snapshots, members, rows), nil
}
func assembleWorkspace(book model.AccountBook, user int64, projects []model.InsightProject, fixed []model.InsightFixedCost, annotations []model.InsightStatementAnnotation, snapshots []model.InsightPortfolioSnapshot, members []model.AccountBookCollaborator, rows []repo.InsightStatementRecord) types.InsightWorkspace {
	out := types.InsightWorkspace{EditableStatementIDs: []int64{}, Projects: []types.InsightProjectSummary{}, FixedCosts: []types.InsightFixedCostSummary{}, Annotations: []types.InsightAnnotation{}, Members: []types.InsightMemberOption{}, Payers: []types.InsightGroup{}, Burdens: []types.InsightGroup{}, Portfolio: []types.InsightPortfolioPoint{}}
	_, owner := authorized(book, members, user)
	names := map[int64]string{}
	currentMembers := map[int64]bool{book.UserID: true}
	for _, m := range members {
		names[m.UserID] = m.Remark
		currentMembers[m.UserID] = true
	}
	if _, ok := names[book.UserID]; !ok {
		names[book.UserID] = ""
	}
	for _, r := range rows {
		if names[r.UserID] == "" && r.MemberName != "" {
			names[r.UserID] = r.MemberName
		}
	}
	for id, name := range names {
		if name == "" {
			name = fmt.Sprintf("成员 %d", id)
		}
		names[id] = name
		if currentMembers[id] {
			out.Members = append(out.Members, types.InsightMemberOption{ID: id, Name: name})
		}
	}
	sort.Slice(out.Members, func(i, j int) bool { return out.Members[i].ID < out.Members[j].ID })
	annotationByID := map[int64]model.InsightStatementAnnotation{}
	for _, a := range annotations {
		annotationByID[a.StatementID] = a
	}
	projectGroups := map[int64]*group{}
	projectPayers := map[int64]map[string]*group{}
	payers, burdens := map[string]*group{}, map[string]*group{}
	for _, p := range projects {
		projectGroups[p.ID] = newGroup(fmt.Sprint(p.ID), p.Name)
		projectPayers[p.ID] = map[string]*group{}
	}
	addMember := func(groups map[string]*group, id int64, row repo.InsightStatementRecord, amount int64) {
		key := fmt.Sprint(id)
		if groups[key] == nil {
			name := names[id]
			if name == "" {
				name = fmt.Sprintf("成员 %d", id)
			}
			groups[key] = newGroup(key, name)
		}
		groups[key].add(row, amount, id)
	}
	for _, r := range rows {
		if r.Type != "income" && r.Type != "expend" {
			continue
		}
		if owner || r.UserID == user {
			out.EditableStatementIDs = append(out.EditableStatementIDs, r.ID)
		}
		amount := cents(r.Amount)
		if amount == 0 {
			continue
		}
		a, has := annotationByID[r.ID]
		var splits []types.InsightAllocation
		valid := true
		if has {
			if len(a.Allocations) > 0 {
				if e := json.Unmarshal(a.Allocations, &splits); e != nil {
					valid = false
				}
			}
			var sum int64
			seen := map[int64]bool{}
			for _, v := range splits {
				if seen[v.MemberID] || v.AmountCents < 0 {
					valid = false
				}
				seen[v.MemberID] = true
				sum += v.AmountCents
			}
			if len(splits) > 0 && (r.Type != "expend" || sum != amount) {
				valid = false
			}
			if splits == nil {
				splits = []types.InsightAllocation{}
			}
			out.Annotations = append(out.Annotations, types.InsightAnnotation{InsightAnnotationInput: types.InsightAnnotationInput{StatementID: r.ID, ConsumerID: a.ConsumerID, ProjectID: a.ProjectID, PayerID: a.PayerID, FixedCostID: a.FixedCostID, Allocations: splits}, CanEdit: owner || r.UserID == user, SplitValid: valid})
		}
		if a.ProjectID != nil && projectGroups[*a.ProjectID] != nil {
			projectGroups[*a.ProjectID].add(r, amount, r.UserID)
		}
		if r.Type != "expend" {
			continue
		}
		if a.PayerID == nil {
			out.UnknownPayerCents += amount
		} else {
			addMember(payers, *a.PayerID, r, amount)
			if a.ProjectID != nil && projectPayers[*a.ProjectID] != nil {
				addMember(projectPayers[*a.ProjectID], *a.PayerID, r, amount)
			}
		}
		if !valid || len(splits) == 0 {
			out.UnallocatedCents += amount
			if !valid {
				out.InvalidSplitCount++
			}
		} else {
			for _, v := range splits {
				if v.AmountCents > 0 {
					addMember(burdens, v.MemberID, r, v.AmountCents)
				}
			}
		}
	}
	for _, p := range projects {
		participants := []int64{}
		_ = json.Unmarshal(p.ParticipantIDs, &participants)
		var startDate, endDate *string
		if p.StartDate != nil {
			v := p.StartDate.Format("2006-01-02")
			startDate = &v
		}
		if p.EndDate != nil {
			v := p.EndDate.Format("2006-01-02")
			endDate = &v
		}
		summary := sortedGroups(map[string]*group{fmt.Sprint(p.ID): projectGroups[p.ID]})[0]
		out.Projects = append(out.Projects, types.InsightProjectSummary{InsightProjectInput: types.InsightProjectInput{ID: p.ID, Icon: p.Icon, Color: p.Color, ParticipantIDs: participants, StartDate: startDate, EndDate: endDate, Name: p.Name, BudgetCents: p.BudgetCents, Archived: p.Archived}, CreatorID: p.CreatorID, CanEdit: owner || p.CreatorID == user, Summary: summary, Payers: sortedGroups(projectPayers[p.ID])})
	}
	for _, f := range fixed {
		if f.IntervalMonths < 1 || f.IntervalMonths > 12 {
			continue
		}
		monthly := int64(math.Round(float64(f.AmountCents) / float64(f.IntervalMonths)))
		annual := int64(math.Round(float64(f.AmountCents) * 12 / float64(f.IntervalMonths)))
		var nextDate *string
		if f.NextRunDate != nil {
			text := f.NextRunDate.Format("2006-01-02")
			nextDate = &text
		}
		out.FixedCosts = append(out.FixedCosts, types.InsightFixedCostSummary{NextRunDate: nextDate, InsightFixedCostInput: types.InsightFixedCostInput{ID: f.ID, Name: f.Name, AmountCents: f.AmountCents, CategoryID: f.CategoryID, AssetID: f.AssetID, IntervalMonths: f.IntervalMonths, DueDay: f.DueDay, CandidateKey: f.CandidateKey, Active: f.Active}, CreatorID: f.CreatorID, CanEdit: owner || f.CreatorID == user, MonthlyCents: monthly, AnnualCents: annual})
		if f.Active {
			out.MonthlyFixedCents += monthly
			out.AnnualFixedCents += annual
		}
	}
	out.Payers = sortedGroups(payers)
	out.Burdens = sortedGroups(burdens)
	out.Portfolio = portfolioPoints(snapshots)
	return out
}
func portfolioPoints(snapshots []model.InsightPortfolioSnapshot) []types.InsightPortfolioPoint {
	result := []types.InsightPortfolioPoint{}
	previous := map[int64]types.InsightAssetBalance{}
	var previousNet int64
	for _, s := range snapshots {
		point := types.InsightPortfolioPoint{ID: s.ID, Date: s.CreatedAt.Format(time.RFC3339), Note: s.Note, AssetsCents: s.AssetsCents, LiabilitiesCents: s.LiabilitiesCents, NetCents: s.AssetsCents - s.LiabilitiesCents, Balances: []types.InsightAssetBalance{}, Changes: []types.InsightAssetChange{}}
		if e := json.Unmarshal(s.Balances, &point.Balances); e != nil {
			continue
		}
		current := map[int64]types.InsightAssetBalance{}
		for _, a := range point.Balances {
			current[a.ID] = a
		}
		if len(result) > 0 {
			delta := point.NetCents - previousNet
			point.DeltaCents = &delta
			all := map[int64]bool{}
			for id := range previous {
				all[id] = true
			}
			for id := range current {
				all[id] = true
			}
			for id := range all {
				before, bok := previous[id]
				after, aok := current[id]
				bnet, anet := before.AmountCents, after.AmountCents
				if before.Type == "debt" {
					bnet = -bnet
				}
				if after.Type == "debt" {
					anet = -anet
				}
				if anet == bnet && bok && aok {
					continue
				}
				name := after.Name
				if !aok {
					name = before.Name
				}
				reason := "余额变化"
				if !bok {
					reason = "新增资产纳入快照"
				} else if !aok {
					reason = "资产已移出快照"
				} else if before.Type != after.Type {
					reason = "资产类型变化"
				}
				var beforeAmount, afterAmount *int64
				if bok {
					value := before.AmountCents
					beforeAmount = &value
				}
				if aok {
					value := after.AmountCents
					afterAmount = &value
				}
				point.Changes = append(point.Changes, types.InsightAssetChange{ID: id, Name: name, BeforeCents: beforeAmount, AfterCents: afterAmount, NetDeltaCents: anet - bnet, Reason: reason})
			}
			sort.Slice(point.Changes, func(i, j int) bool { return point.Changes[i].ID < point.Changes[j].ID })
		}
		result = append(result, point)
		previous = current
		previousNet = point.NetCents
	}
	return result
}
