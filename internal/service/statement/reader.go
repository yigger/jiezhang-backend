package statement

import (
	"context"
	"fmt"
	"strings"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	"github.com/yigger/jiezhang-backend/internal/types"
)

// 账单列表
func (s Reader) GetStatements(ctx context.Context, input ListInput) ([]types.StatementListItem, error) {
	filter := repo.StatementListFilter{
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
	rows, err := helperservice.AssembleListRows(ctx, s.queryRepo, filter)
	if err != nil {
		return nil, err
	}

	items := make([]types.StatementListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s Reader) SearchStatements(ctx context.Context, accountBookID int64, keyword string) ([]types.StatementListItem, error) {
	filter := repo.StatementListFilter{
		AccountBookID: accountBookID,
		Keyword:       keyword,
		OrderBy:       "created_at desc",
	}
	rows, err := helperservice.AssembleListRows(ctx, s.queryRepo, filter)
	if err != nil {
		return nil, err
	}

	items := make([]types.StatementListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.rowMapper.ToListItem(row))
	}

	return items, nil
}

func (s Reader) GetCategories(ctx context.Context, input GetCategoriesInput) (types.StatementCategoriesResult, error) {
	parents, err := s.categoryRepo.ListParents(ctx, input.AccountBookID, input.Type)
	if err != nil {
		return types.StatementCategoriesResult{}, err
	}
	parentIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentIDs = append(parentIDs, p.ID)
	}

	children, err := s.categoryRepo.ListChildrenByParentIDs(ctx, input.AccountBookID, input.Type, parentIDs)
	if err != nil {
		return types.StatementCategoriesResult{}, err
	}
	frequents, err := s.categoryRepo.ListFrequentChildren(ctx, input.AccountBookID, input.Type, 10)
	if err != nil {
		return types.StatementCategoriesResult{}, err
	}

	childrenByParent := make(map[int64][]types.StatementCategoryChildItem, len(parents))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], types.StatementCategoryChildItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: s.rowMapper.BuildPublicURL(child.IconPath),
		})
	}

	categories := make([]types.StatementCategoryTreeItem, 0, len(parents))
	for _, p := range parents {
		childs := childrenByParent[p.ID]
		if childs == nil {
			childs = []types.StatementCategoryChildItem{}
		}
		categories = append(categories, types.StatementCategoryTreeItem{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: s.rowMapper.BuildPublicURL(p.IconPath),
			Childs:   childs,
		})
	}

	frequentItems := make([]types.StatementFrequentCategoryItem, 0, len(frequents))
	frequentParentIDs := make([]int64, 0, len(frequents))
	for _, item := range frequents {
		if item.ParentID > 0 {
			frequentParentIDs = append(frequentParentIDs, item.ParentID)
		}
	}
	parentRows, err := s.categoryRepo.ListByIDs(ctx, frequentParentIDs)
	if err != nil {
		return types.StatementCategoriesResult{}, err
	}
	parentNames := make(map[int64]string, len(parentRows))
	for _, parent := range parentRows {
		parentNames[parent.ID] = parent.Name
	}
	for _, f := range frequents {
		var parent *types.StatementCategoryParentItem
		if f.ParentID > 0 {
			parent = &types.StatementCategoryParentItem{
				ID:   f.ParentID,
				Name: parentNames[f.ParentID],
			}
		}
		frequentItems = append(frequentItems, types.StatementFrequentCategoryItem{
			ID:       f.ID,
			Name:     f.Name,
			IconPath: s.rowMapper.BuildPublicURL(f.IconPath),
			Parent:   parent,
		})
	}

	return types.StatementCategoriesResult{
		Frequent:   frequentItems,
		Categories: categories,
	}, nil
}

func (s Reader) GetAssets(ctx context.Context, input GetCategoriesInput) (types.StatementAssetsResult, error) {

	var assetResult []types.StatementAssetTreeItem
	parents, err := s.assetRepo.ListParents(ctx, input.AccountBookID)
	if err != nil {
		return types.StatementAssetsResult{}, err
	}
	parentsIDs := make([]int64, 0, len(parents))
	for _, p := range parents {
		parentsIDs = append(parentsIDs, p.ID)
	}

	children, err := s.assetRepo.ListChildrenByParentIDs(ctx, input.AccountBookID, parentsIDs)
	if err != nil {
		return types.StatementAssetsResult{}, err
	}
	childrenByParent := make(map[int64][]types.StatementAssetChildItem, len(parents))
	for _, child := range children {
		childrenByParent[child.ParentID] = append(childrenByParent[child.ParentID], types.StatementAssetChildItem{
			ID:       child.ID,
			Name:     child.Name,
			IconPath: s.rowMapper.BuildPublicURL(child.IconPath),
		})
	}

	for _, p := range parents {
		childs := childrenByParent[p.ID]
		if childs == nil {
			childs = []types.StatementAssetChildItem{}
		}
		assetResult = append(assetResult, types.StatementAssetTreeItem{
			ID:       p.ID,
			Name:     p.Name,
			IconPath: s.rowMapper.BuildPublicURL(p.IconPath),
			Childs:   childs,
		})
	}

	frequentResult := make([]types.StatementFrequentAssetItem, 0)
	frequent, err := s.assetRepo.ListFrequentChildren(ctx, input.AccountBookID, 10)
	if err != nil {
		return types.StatementAssetsResult{}, err
	}
	parentIDs := make([]int64, 0, len(frequent))
	for _, item := range frequent {
		if item.ParentID > 0 {
			parentIDs = append(parentIDs, item.ParentID)
		}
	}
	parentRows, err := s.assetRepo.ListByIDs(ctx, parentIDs)
	if err != nil {
		return types.StatementAssetsResult{}, err
	}
	parentNames := make(map[int64]string, len(parentRows))
	for _, parent := range parentRows {
		parentNames[parent.ID] = parent.Name
	}
	for _, f := range frequent {
		var parent *types.StatementAssetParentItem
		if f.ParentID > 0 {
			parent = &types.StatementAssetParentItem{
				ID:   f.ParentID,
				Name: parentNames[f.ParentID],
			}
		}
		frequentResult = append(frequentResult, types.StatementFrequentAssetItem{
			ID:       f.ID,
			Name:     f.Name,
			IconPath: s.rowMapper.BuildPublicURL(f.IconPath),
			Parent:   parent,
		})
	}
	return types.StatementAssetsResult{
		Frequent:   frequentResult,
		Categories: assetResult,
	}, nil
}

func (s Reader) GetDefaultCategoryAsset(ctx context.Context, input GetCategoriesInput) (*types.StatementDefaultCategoryAssetItem, error) {
	record, err := s.queryRepo.GetLatestCategoryAssetByType(ctx, input.AccountBookID, input.Type)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	rows, err := helperservice.AssembleRows(ctx, s.queryRepo, input.AccountBookID, []tablemodel.Statement{*record})
	if err != nil {
		return nil, err
	}
	view := rows[0]
	categoryName, assetName := view.CategoryName, view.AssetName
	if view.CategoryParentID > 0 {
		categoryName = view.CategoryParentName + " -> " + categoryName
	}
	if view.AssetParentID > 0 {
		assetName = view.AssetParentName + " -> " + assetName
	}
	return &types.StatementDefaultCategoryAssetItem{
		CategoryName: categoryName,
		AssetName:    assetName,
		CategoryID:   record.CategoryID,
		AssetID:      record.AssetID,
	}, nil
}

func (s Reader) GetTargetObjects(ctx context.Context, input GetCategoriesInput) ([]string, error) {
	return s.queryRepo.ListDistinctTargetObjectsByType(ctx, input.AccountBookID, input.Type)
}

func (s Reader) CategoriesGuess(ctx context.Context, input GetCategoriesInput) ([]types.StatementFrequentCategoryItem, error) {
	statementType := input.Type
	if statementType == "" {
		statementType = "expend"
	}

	filter := repo.CategoryGuessFilter{
		AccountBookID: input.AccountBookID,
		StatementType: statementType,
		Now:           time.Now(),
		Limit:         3,
	}
	rows, err := s.categoryRepo.ListGuessedFrequentByStatementType(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]types.StatementFrequentCategoryItem, 0, len(rows))
	parentIDs := make([]int64, 0, len(rows))
	for _, item := range rows {
		if item.ParentID > 0 {
			parentIDs = append(parentIDs, item.ParentID)
		}
	}
	parentRows, err := s.categoryRepo.ListByIDs(ctx, parentIDs)
	if err != nil {
		return nil, err
	}
	parentNames := make(map[int64]string, len(parentRows))
	for _, parent := range parentRows {
		parentNames[parent.ID] = parent.Name
	}
	for _, row := range rows {
		var parent *types.StatementCategoryParentItem
		if _, hasParent := parentNames[row.ParentID]; hasParent {
			parent = &types.StatementCategoryParentItem{
				ID:   row.ParentID,
				Name: parentNames[row.ParentID],
			}
		}
		items = append(items, types.StatementFrequentCategoryItem{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: s.rowMapper.BuildPublicURL(row.IconPath),
			Parent:   parent,
		})
	}
	return items, nil
}

func (s Reader) AssetsGuess(ctx context.Context, input GetCategoriesInput) ([]types.StatementFrequentAssetItem, error) {
	rows, err := s.assetRepo.ListGuessedFrequentByStatementTime(ctx, repo.AssetGuessFilter{
		AccountBookID: input.AccountBookID,
		Now:           time.Now(),
		Limit:         3,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.StatementFrequentAssetItem, 0, len(rows))
	parentIDs := make([]int64, 0, len(rows))
	for _, item := range rows {
		if item.ParentID > 0 {
			parentIDs = append(parentIDs, item.ParentID)
		}
	}
	parentRows, err := s.assetRepo.ListByIDs(ctx, parentIDs)
	if err != nil {
		return nil, err
	}
	parentNames := make(map[int64]string, len(parentRows))
	for _, parent := range parentRows {
		parentNames[parent.ID] = parent.Name
	}
	for _, row := range rows {
		var parent *types.StatementAssetParentItem
		if _, hasParent := parentNames[row.ParentID]; hasParent {
			parent = &types.StatementAssetParentItem{
				ID:   row.ParentID,
				Name: parentNames[row.ParentID],
			}
		}
		items = append(items, types.StatementFrequentAssetItem{
			ID:       row.ID,
			Name:     row.Name,
			IconPath: s.rowMapper.BuildPublicURL(row.IconPath),
			Parent:   parent,
		})
	}
	return items, nil
}

func (s Reader) GetStatementByID(ctx context.Context, statementID int64, accountBookID int64, currentUserID int64) (types.StatementDetailItem, error) {
	row, err := helperservice.AssembleSingleRow(ctx, s.queryRepo, statementID, accountBookID)
	if err != nil {
		return types.StatementDetailItem{}, err
	}

	avatarRows, err := s.queryRepo.ListAvatarsByStatementID(ctx, statementID)
	if err != nil {
		return types.StatementDetailItem{}, err
	}
	uploadFiles := make([]types.StatementUploadFileItem, 0, len(avatarRows))
	for _, a := range avatarRows {
		uploadFiles = append(uploadFiles, types.StatementUploadFileItem{
			ID:  a.ID,
			URL: s.buildAvatarURL(a.Path, row.UserID, a.ImageableID),
		})
	}

	canAdmin, err := s.categoryRepo.CanAdmin(ctx, accountBookID, currentUserID)
	if err != nil {
		canAdmin = false
	}

	return s.rowMapper.ToDetailItem(row, currentUserID, canAdmin, uploadFiles), nil
}

func (s Reader) buildAvatarURL(avatarPath string, userID int64, statementID int64) string {
	if strings.HasPrefix(avatarPath, "/private") {
		return s.rowMapper.BuildPublicURL(avatarPath)
	}
	return s.rowMapper.BuildPublicURL(fmt.Sprintf("/private/%d/statements/%d/%s", userID, statementID, avatarPath))
}

func (s Reader) GetImages(ctx context.Context, accountBookID int64) (types.StatementImagesResult, error) {
	rows, err := s.queryRepo.ListAvatarRows(ctx, accountBookID)
	if err != nil {
		return types.StatementImagesResult{}, err
	}

	statementIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		statementIDs = append(statementIDs, row.ImageableID)
	}
	statements, err := s.queryRepo.BatchGetStatements(ctx, statementIDs)
	if err != nil {
		return types.StatementImagesResult{}, err
	}
	statementMap := make(map[int64]tablemodel.Statement, len(statements))
	for _, st := range statements {
		statementMap[st.ID] = st
	}
	yearOrder := make([]int, 0)
	monthOrderByYear := make(map[int][]int)
	timeline := make(map[int]map[int][]types.StatementImageItem)
	seenImage := make(map[string]struct{})
	avatars := make([]string, 0)

	for _, row := range rows {
		year := statementMap[row.ImageableID].Year
		month := statementMap[row.ImageableID].Month
		path := s.buildAvatarURL(row.Path, statementMap[row.ImageableID].UserID, row.ImageableID)
		imageItem := types.StatementImageItem{
			StatementID: row.ImageableID,
			AvatarID:    row.ID,
			Path:        path,
		}

		if _, ok := timeline[year]; !ok {
			timeline[year] = make(map[int][]types.StatementImageItem)
			yearOrder = append(yearOrder, year)
		}
		if _, ok := timeline[year][month]; !ok {
			timeline[year][month] = make([]types.StatementImageItem, 0)
			monthOrderByYear[year] = append(monthOrderByYear[year], month)
		}
		timeline[year][month] = append(timeline[year][month], imageItem)

		if _, ok := seenImage[path]; !ok && path != "" {
			seenImage[path] = struct{}{}
			avatars = append(avatars, path)
		}
	}

	resultTimeline := make([]types.StatementImageYearGroup, 0, len(yearOrder))
	for _, year := range yearOrder {
		months := monthOrderByYear[year]
		monthGroups := make([]types.StatementImageMonthGroup, 0, len(months))
		for _, month := range months {
			monthGroups = append(monthGroups, types.StatementImageMonthGroup{
				Month: month,
				Data:  dedupeImageItems(timeline[year][month]),
			})
		}
		resultTimeline = append(resultTimeline, types.StatementImageYearGroup{
			Year: year,
			Data: monthGroups,
		})
	}

	return types.StatementImagesResult{
		AvatarTimeline: resultTimeline,
		Avatars:        avatars,
	}, nil
}

func dedupeImageItems(items []types.StatementImageItem) []types.StatementImageItem {
	seen := make(map[int64]struct{}, len(items))
	out := make([]types.StatementImageItem, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item.AvatarID]; ok {
			continue
		}
		seen[item.AvatarID] = struct{}{}
		out = append(out, item)
	}
	return out
}

func statementAvatarSecureURL(baseURL string, statementID int64, avatarID int64) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return fmt.Sprintf("/api/statements/%d/avatars/%d", statementID, avatarID)
	}
	return fmt.Sprintf("%s/api/statements/%d/avatars/%d", strings.TrimRight(baseURL, "/"), statementID, avatarID)
}
