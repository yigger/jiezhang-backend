package statement

import (
	"context"
	"strings"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	helperservice "github.com/yigger/jiezhang-backend/internal/service/helper"
	"github.com/yigger/jiezhang-backend/internal/types"
)

func (s Writer) normalizeStatementWriteInput(ctx context.Context, input WriteInput) (tablemodel.Statement, error) {
	statementType := strings.TrimSpace(input.Type)
	if statementType == "" {
		return tablemodel.Statement{}, ValidateError{Message: "invalid statement type"}
	}
	if input.Amount <= 0 {
		return tablemodel.Statement{}, ValidateError{Message: "invalid amount"}
	}

	occurredAt, err := parseStatementDateTime(input.Date, input.Time)
	if err != nil {
		return tablemodel.Statement{}, ValidateError{Message: "invalid date or time"}
	}

	assetID := input.AssetID
	targetAssetID := int64PtrOrNil(input.ToAssetID)
	if statementType == "transfer" || statementType == "repayment" {
		assetID = input.FromAssetID
		if assetID <= 0 || input.ToAssetID <= 0 {
			return tablemodel.Statement{}, ValidateError{Message: "invalid asset ID"}
		}
		targetAssetID = int64PtrOrNil(input.ToAssetID)
	}

	categoryID := input.CategoryID
	specialTypes := map[string]bool{"transfer": true, "repayment": true, "loan_in": true, "loan_out": true, "reimburse": true, "payment_proxy": true}
	if specialTypes[statementType] {
		specialID, err := s.categoryRepo.FindBySpecialType(ctx, statementType)
		if err != nil {
			return tablemodel.Statement{}, ValidateError{Message: "special category not found for type: " + statementType}
		}
		categoryID = specialID
	}

	switch statementType {
	case "expend", "income":
		if assetID <= 0 || categoryID <= 0 {
			return tablemodel.Statement{}, ValidateError{Message: "invalid asset or category ID"}
		}
	case "transfer", "repayment":
		fromAssetID := input.FromAssetID
		toAssetID := input.ToAssetID
		if fromAssetID <= 0 || toAssetID <= 0 {
			return tablemodel.Statement{}, ValidateError{Message: "invalid from or to asset ID"}
		}
	case "loan_in", "loan_out", "reimburse", "payment_proxy":
		if assetID <= 0 {
			return tablemodel.Statement{}, ValidateError{Message: "invalid asset ID"}
		}
	default:
		return tablemodel.Statement{}, ValidateError{Message: "invalid statement type"}
	}

	return tablemodel.Statement{
		UserID:        input.UserID,
		AccountBookID: input.AccountBookID,
		Type:          statementType,
		Amount:        input.Amount,
		Description:   input.Description,
		Mood:          input.Mood,
		CategoryID:    categoryID,
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
		CreatedAt:     occurredAt,
		Year:          occurredAt.Year(),
		Month:         int(occurredAt.Month()),
		Day:           occurredAt.Day(),
		TimeText:      occurredAt.Format("15:04"),
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
	return time.Time{}, ValidateError{Message: "invalid date or time"}
}

func int64PtrOrNil(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	n := v
	return &n
}

func (s Writer) mergeStatementPatch(current tablemodel.Statement, update UpdateInput) WriteInput {
	input := WriteInput{
		StatementID:   update.StatementID,
		UserID:        update.UserID,
		AccountBookID: update.AccountBookID,
		Type:          current.Type,
		Amount:        current.Amount,
		Description:   current.Description,
		Mood:          current.Mood,
		CategoryID:    current.CategoryID,
		AssetID:       current.AssetID,
		FromAssetID:   current.AssetID,
		TargetObject:  current.TargetObject,
		Location:      current.Location,
		Nation:        current.Nation,
		Province:      current.Province,
		City:          current.City,
		District:      current.District,
		Street:        current.Street,
		Date:          current.CreatedAt.Format("2006-01-02"),
		Time:          current.CreatedAt.Format("15:04:05"),
	}
	if current.TargetAssetID != nil && *current.TargetAssetID > 0 {
		input.ToAssetID = *current.TargetAssetID
	}
	if current.PayeeID != nil && *current.PayeeID > 0 {
		input.PayeeID = *current.PayeeID
	}

	p := update.Patch
	if p.Type != nil {
		input.Type = strings.TrimSpace(*p.Type)
	}
	if p.Amount != nil {
		input.Amount = *p.Amount
	}
	if p.Description != nil {
		input.Description = *p.Description
	}
	if p.Mood != nil {
		input.Mood = *p.Mood
	}
	if p.CategoryID != nil {
		input.CategoryID = *p.CategoryID
	}
	if p.AssetID != nil {
		input.AssetID = *p.AssetID
	}
	if p.FromAssetID != nil {
		input.FromAssetID = *p.FromAssetID
	}
	if p.ToAssetID != nil {
		input.ToAssetID = *p.ToAssetID
	}
	if p.PayeeID != nil {
		input.PayeeID = *p.PayeeID
	}
	if p.TargetObject != nil {
		input.TargetObject = *p.TargetObject
	}
	if p.Location != nil {
		input.Location = *p.Location
	}
	if p.Nation != nil {
		input.Nation = *p.Nation
	}
	if p.Province != nil {
		input.Province = *p.Province
	}
	if p.City != nil {
		input.City = *p.City
	}
	if p.District != nil {
		input.District = *p.District
	}
	if p.Street != nil {
		input.Street = *p.Street
	}
	if p.Date != nil {
		input.Date = strings.TrimSpace(*p.Date)
	}
	if p.Time != nil {
		input.Time = strings.TrimSpace(*p.Time)
	}

	return input
}

func (s Writer) RemoveAvatar(ctx context.Context, accountBookID int64, statementID int64, avatarID int64) error {
	if accountBookID <= 0 || statementID <= 0 || avatarID <= 0 {
		return ErrStatementInvalidInput
	}
	return s.statementRepo.DeleteAvatarByID(ctx, accountBookID, statementID, avatarID)
}

func (s Writer) CreateStatement(ctx context.Context, input WriteInput) (types.StatementListItem, error) {
	record, err := s.normalizeStatementWriteInput(ctx, input)
	if err != nil {
		return types.StatementListItem{}, err
	}
	var id int64
	err = s.transaction.WithinTransaction(ctx, func(tx repo.Mutation) error {
		var e error
		id, e = tx.Create(ctx, record, statementEffect(record.Type, record.Amount))
		if e != nil {
			return e
		}
		if input.ProjectID > 0 {
			projectTx, ok := tx.(repo.ProjectMutation)
			if !ok {
				return ValidateError{Message: "project entries unavailable"}
			}
			if record.Type != "income" && record.Type != "expend" {
				return ValidateError{Message: "invalid project entry type"}
			}
			return projectTx.AttachProject(ctx, input.AccountBookID, id, input.ProjectID, input.ConsumerID, input.UserID)
		}
		return nil
	})
	if err != nil {
		return types.StatementListItem{}, err
	}
	row, err := helperservice.AssembleSingleRow(ctx, s.queryRepo, id, input.AccountBookID)
	if err != nil {
		return types.StatementListItem{}, err
	}
	return s.rowMapper.ToListItem(row), nil
}
func (s Writer) UpdateStatement(ctx context.Context, input UpdateInput) (types.StatementListItem, error) {
	err := s.transaction.WithinTransaction(ctx, func(tx repo.Mutation) error {
		current, err := tx.LockCurrent(ctx, input.StatementID, input.AccountBookID)
		if err != nil {
			return err
		}
		if current.UserID != input.UserID {
			return ErrStatementPermissionDenied
		}
		record, err := s.normalizeStatementWriteInput(ctx, s.mergeStatementPatch(current, input))
		if err != nil {
			return err
		}
		return tx.UpdateByID(ctx, input.StatementID, input.AccountBookID, record, statementEffect(current.Type, current.Amount), statementEffect(record.Type, record.Amount))
	})
	if err != nil {
		return types.StatementListItem{}, err
	}
	row, err := helperservice.AssembleSingleRow(ctx, s.queryRepo, input.StatementID, input.AccountBookID)
	if err != nil {
		return types.StatementListItem{}, err
	}
	return s.rowMapper.ToListItem(row), nil
}
func (s Writer) DeleteStatement(ctx context.Context, id, userID, bookID int64) error {
	return s.transaction.WithinTransaction(ctx, func(tx repo.Mutation) error {
		current, err := tx.LockCurrent(ctx, id, bookID)
		if err != nil {
			return err
		}
		if current.UserID != userID {
			return ErrStatementPermissionDenied
		}
		return tx.DeleteByID(ctx, id, bookID, statementEffect(current.Type, current.Amount))
	})
}
