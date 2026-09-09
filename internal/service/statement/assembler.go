package statement

import (
	"context"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

// batchLookupRepo defines the batch-lookup operations shared by both
// StatementQueryRepository and SuperStatementRepository.
type batchLookupRepo interface {
	BatchGetCategories(ctx context.Context, ids []int64) ([]repository.CategoryBatchRecord, error)
	BatchGetAssets(ctx context.Context, ids []int64) ([]repository.AssetBatchRecord, error)
	BatchGetPayees(ctx context.Context, ids []int64) ([]repository.PayeeBatchRecord, error)
	BatchGetCollaboratorRemarks(ctx context.Context, accountBookID int64, userIDs []int64) ([]repository.CollaboratorRemarkRecord, error)
	BatchCheckHasPic(ctx context.Context, statementIDs []int64) ([]repository.StatementHasPicRecord, error)
}

// AssembleListRows performs a split-query approach: first fetches statement rows
// from a simple single-table query, then batch-fetches related data (categories,
// assets, payees, remarks, has_pic) and assembles the full StatementListRowRecord
// in Go. This avoids the expensive multi-table JOIN used by the old
// ListRowsWithRelations.
func AssembleListRows(ctx context.Context, queryRepo repository.StatementQueryRepository, filter repository.StatementListFilter) ([]repository.StatementListRowRecord, error) {
	simpleRows, err := queryRepo.ListSimpleRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(simpleRows) == 0 {
		return nil, nil
	}

	return assembleRows(ctx, queryRepo, filter.AccountBookID, simpleRows)
}

// AssembleSingleRow assembles a single statement row with its related data
// using simple single-table queries instead of a multi-table JOIN.
func AssembleSingleRow(ctx context.Context, queryRepo repository.StatementQueryRepository, statementID int64, accountBookID int64) (repository.StatementListRowRecord, error) {
	simpleRow, err := queryRepo.GetSimpleRowByID(ctx, statementID, accountBookID)
	if err != nil {
		return repository.StatementListRowRecord{}, err
	}

	rows, err := assembleRows(ctx, queryRepo, accountBookID, []repository.StatementSimpleRowRecord{simpleRow})
	if err != nil {
		return repository.StatementListRowRecord{}, err
	}
	return rows[0], nil
}

// AssembleSuperListRows is the super-statement variant of AssembleListRows.
func AssembleSuperListRows(ctx context.Context, repo repository.SuperStatementRepository, filter repository.SuperStatementFilter) ([]repository.StatementListRowRecord, error) {
	simpleRows, err := repo.ListSimpleRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(simpleRows) == 0 {
		return nil, nil
	}

	return assembleRows(ctx, repo, filter.AccountBookID, simpleRows)
}

// assembleRows is the shared assembly logic.
func assembleRows(ctx context.Context, lookup batchLookupRepo, accountBookID int64, simpleRows []repository.StatementSimpleRowRecord) ([]repository.StatementListRowRecord, error) {
	// Collect foreign keys.
	categoryIDs := make(map[int64]struct{})
	assetIDs := make(map[int64]struct{})
	payeeIDs := make(map[int64]struct{})
	userIDs := make(map[int64]struct{})
	statementIDs := make([]int64, 0, len(simpleRows))

	for _, row := range simpleRows {
		categoryIDs[row.CategoryID] = struct{}{}
		assetIDs[row.AssetID] = struct{}{}
		if row.TargetAssetID > 0 {
			assetIDs[row.TargetAssetID] = struct{}{}
		}
		if row.PayeeID > 0 {
			payeeIDs[row.PayeeID] = struct{}{}
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
	simpleRows []repository.StatementSimpleRowRecord,
	categoryMap map[int64]categoryInfo,
	assetMap map[int64]assetInfo,
	payeeMap map[int64]string,
	remarkMap map[int64]string,
	hasPicMap map[int64]bool,
) []repository.StatementListRowRecord {
	items := make([]repository.StatementListRowRecord, 0, len(simpleRows))
	for _, row := range simpleRows {
		cat := categoryMap[row.CategoryID]
		ast := assetMap[row.AssetID]

		items = append(items, repository.StatementListRowRecord{
			ID:                 row.ID,
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
			PayeeID:            row.PayeeID,
			PayeeName:          payeeMap[row.PayeeID],
			TargetAssetID:      row.TargetAssetID,
			TargetAssetName:    assetMap[row.TargetAssetID].Name,
			TargetObject:       row.TargetObject,
			CategoryParentName: cat.ParentName,
			AssetParentName:    ast.ParentName,
		})
	}
	return items
}

type categoryInfo struct {
	Name       string
	IconPath   string
	ParentName string
}

type assetInfo struct {
	Name       string
	IconPath   string
	ParentName string
}

func batchCategoryMap(ctx context.Context, lookup batchLookupRepo, idSet map[int64]struct{}) (map[int64]categoryInfo, error) {
	ids := keysFromSet(idSet)
	rows, err := lookup.BatchGetCategories(ctx, ids)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]categoryInfo, len(rows))
	for _, r := range rows {
		m[r.ID] = categoryInfo{Name: r.Name, IconPath: r.IconPath, ParentName: r.ParentName}
	}
	return m, nil
}

func batchAssetMap(ctx context.Context, lookup batchLookupRepo, idSet map[int64]struct{}) (map[int64]assetInfo, error) {
	ids := keysFromSet(idSet)
	rows, err := lookup.BatchGetAssets(ctx, ids)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]assetInfo, len(rows))
	for _, r := range rows {
		m[r.ID] = assetInfo{Name: r.Name, IconPath: r.IconPath, ParentName: r.ParentName}
	}
	return m, nil
}

func batchPayeeMap(ctx context.Context, lookup batchLookupRepo, idSet map[int64]struct{}) (map[int64]string, error) {
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

func batchRemarkMap(ctx context.Context, lookup batchLookupRepo, accountBookID int64, idSet map[int64]struct{}) (map[int64]string, error) {
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

func batchHasPicMap(ctx context.Context, lookup batchLookupRepo, statementIDs []int64) (map[int64]bool, error) {
	rows, err := lookup.BatchCheckHasPic(ctx, statementIDs)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]bool, len(rows))
	for _, r := range rows {
		m[r.StatementID] = r.HasPic
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
