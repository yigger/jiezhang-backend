package helper

import (
	"context"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

// AssembleListRows batch-loads related table models and builds service data.
func AssembleListRows(ctx context.Context, queryRepo repo.StatementQueryRepository, filter repo.StatementListFilter) ([]StatementData, error) {
	simpleRows, err := queryRepo.ListSimpleRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(simpleRows) == 0 {
		return nil, nil
	}

	return AssembleRows(ctx, queryRepo, filter.AccountBookID, simpleRows)
}

// AssembleSingleRow assembles a single statement row with its related data
// using simple single-table queries instead of a multi-table JOIN.
func AssembleSingleRow(ctx context.Context, queryRepo repo.DetailQuery, statementID int64, accountBookID int64) (StatementData, error) {
	simpleRow, err := queryRepo.GetSimpleRowByID(ctx, statementID, accountBookID)
	if err != nil {
		return StatementData{}, err
	}

	rows, err := AssembleRows(ctx, queryRepo, accountBookID, []tablemodel.Statement{simpleRow})
	if err != nil {
		return StatementData{}, err
	}
	return rows[0], nil
}

// AssembleSuperListRows is the super-statement variant of AssembleListRows.
func AssembleSuperListRows(ctx context.Context, repo repo.SuperStatementRepository, filter repo.SuperStatementFilter) ([]StatementData, error) {
	simpleRows, err := repo.ListSimpleRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(simpleRows) == 0 {
		return nil, nil
	}

	return AssembleRows(ctx, repo, filter.AccountBookID, simpleRows)
}

// AssembleRows is the shared assembly logic.
func AssembleRows(ctx context.Context, lookup repo.BatchLookup, accountBookID int64, simpleRows []tablemodel.Statement) ([]StatementData, error) {
	// Collect foreign keys.
	categoryIDs := make(map[int64]struct{})
	assetIDs := make(map[int64]struct{})
	payeeIDs := make(map[int64]struct{})
	userIDs := make(map[int64]struct{})
	statementIDs := make([]int64, 0, len(simpleRows))

	for _, row := range simpleRows {
		categoryIDs[row.CategoryID] = struct{}{}
		assetIDs[row.AssetID] = struct{}{}
		if row.TargetAssetID != nil && *row.TargetAssetID > 0 {
			assetIDs[*row.TargetAssetID] = struct{}{}
		}
		if row.PayeeID != nil && *row.PayeeID > 0 {
			payeeIDs[*row.PayeeID] = struct{}{}
		}
		userIDs[row.UserID] = struct{}{}
		statementIDs = append(statementIDs, row.ID)
	}

	// Batch fetch related data.
	categoryMap, err := batchCategoryMap(ctx, lookup, categoryIDs)
	if err != nil {
		return nil, err
	}
	assetMap, err := batchAssetMap(ctx, lookup, assetIDs)
	if err != nil {
		return nil, err
	}
	payeeMap, err := batchPayeeMap(ctx, lookup, payeeIDs)
	if err != nil {
		return nil, err
	}
	remarkMap, err := batchRemarkMap(ctx, lookup, accountBookID, userIDs)
	if err != nil {
		return nil, err
	}
	hasPicMap, err := batchHasPicMap(ctx, lookup, statementIDs)
	if err != nil {
		return nil, err
	}

	return buildListRowRecords(simpleRows, categoryMap, assetMap, payeeMap, remarkMap, hasPicMap), nil
}

func buildListRowRecords(
	simpleRows []tablemodel.Statement,
	categoryMap map[int64]categoryInfo,
	assetMap map[int64]assetInfo,
	payeeMap map[int64]string,
	remarkMap map[int64]string,
	hasPicMap map[int64]bool,
) []StatementData {
	items := make([]StatementData, 0, len(simpleRows))
	for _, row := range simpleRows {
		cat := categoryMap[row.CategoryID]
		ast := assetMap[row.AssetID]

		items = append(items, StatementData{
			ID:                 row.ID,
			CategoryParentID:   cat.ParentID,
			AssetParentID:      ast.ParentID,
			UserID:             row.UserID,
			Type:               row.Type,
			Amount:             row.Amount,
			Description:        row.Description,
			CategoryID:         row.CategoryID,
			AssetID:            row.AssetID,
			Remark:             remarkMap[row.UserID],
			Mood:               row.Mood,
			IconPath:           cat.IconPath,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
			CategoryName:       cat.Name,
			AssetName:          ast.Name,
			Location:           row.Location,
			Nation:             row.Nation,
			Province:           row.Province,
			City:               row.City,
			District:           row.District,
			Street:             row.Street,
			HasPic:             hasPicMap[row.ID],
			Residue:            row.Residue,
			PayeeID:            pointerID(row.PayeeID),
			PayeeName:          payeeMap[pointerID(row.PayeeID)],
			TargetAssetID:      pointerID(row.TargetAssetID),
			TargetAssetName:    assetMap[pointerID(row.TargetAssetID)].Name,
			TargetObject:       row.TargetObject,
			CategoryParentName: cat.ParentName,
			AssetParentName:    ast.ParentName,
		})
	}
	return items
}

type categoryInfo struct {
	ParentID   int64
	Name       string
	IconPath   string
	ParentName string
}

type assetInfo struct {
	ParentID   int64
	Name       string
	IconPath   string
	ParentName string
}

func batchCategoryMap(ctx context.Context, lookup repo.BatchLookup, idSet map[int64]struct{}) (map[int64]categoryInfo, error) {
	ids := keysFromSet(idSet)
	rows, err := lookup.BatchGetCategories(ctx, ids)
	if err != nil {
		return nil, err
	}
	parents := make(map[int64]struct{})
	for _, row := range rows {
		if row.ParentID > 0 {
			parents[row.ParentID] = struct{}{}
		}
	}
	parentRows, err := lookup.BatchGetCategories(ctx, keysFromSet(parents))
	if err != nil {
		return nil, err
	}
	parentNames := make(map[int64]string, len(parentRows))
	for _, parent := range parentRows {
		parentNames[parent.ID] = parent.Name
	}
	m := make(map[int64]categoryInfo, len(rows))
	for _, r := range rows {
		m[r.ID] = categoryInfo{ParentID: r.ParentID, Name: r.Name, IconPath: r.IconPath, ParentName: parentNames[r.ParentID]}
	}
	return m, nil
}

func batchAssetMap(ctx context.Context, lookup repo.BatchLookup, idSet map[int64]struct{}) (map[int64]assetInfo, error) {
	ids := keysFromSet(idSet)
	rows, err := lookup.BatchGetAssets(ctx, ids)
	if err != nil {
		return nil, err
	}
	parents := make(map[int64]struct{})
	for _, row := range rows {
		if row.ParentID > 0 {
			parents[row.ParentID] = struct{}{}
		}
	}
	parentRows, err := lookup.BatchGetAssets(ctx, keysFromSet(parents))
	if err != nil {
		return nil, err
	}
	parentNames := make(map[int64]string, len(parentRows))
	for _, parent := range parentRows {
		parentNames[parent.ID] = parent.Name
	}
	m := make(map[int64]assetInfo, len(rows))
	for _, r := range rows {
		m[r.ID] = assetInfo{ParentID: r.ParentID, Name: r.Name, IconPath: r.IconPath, ParentName: parentNames[r.ParentID]}
	}
	return m, nil
}

func batchPayeeMap(ctx context.Context, lookup repo.BatchLookup, idSet map[int64]struct{}) (map[int64]string, error) {
	ids := keysFromSet(idSet)
	rows, err := lookup.BatchGetPayees(ctx, ids)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]string, len(rows))
	for _, r := range rows {
		m[r.ID] = r.Name
	}
	return m, nil
}

func batchRemarkMap(ctx context.Context, lookup repo.BatchLookup, accountBookID int64, idSet map[int64]struct{}) (map[int64]string, error) {
	ids := keysFromSet(idSet)
	rows, err := lookup.BatchGetCollaboratorRemarks(ctx, accountBookID, ids)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]string, len(rows))
	for _, r := range rows {
		m[r.UserID] = r.Remark
	}
	return m, nil
}

func batchHasPicMap(ctx context.Context, lookup repo.BatchLookup, statementIDs []int64) (map[int64]bool, error) {
	rows, err := lookup.BatchStatementAvatars(ctx, statementIDs)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]bool, len(rows))
	for _, r := range rows {
		m[r.ImageableID] = true
	}
	return m, nil
}

func keysFromSet(idSet map[int64]struct{}) []int64 {
	if len(idSet) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func pointerID(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}
