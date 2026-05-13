package statement

import (
	"fmt"
	"strings"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"github.com/yigger/jiezhang-backend/internal/service/helper"
)

type URLBuilder interface {
	BuildPublicURL(raw string) string
}

type RowMapper struct {
	urlBuilder URLBuilder
}

func NewRowMapper(urlBuilder URLBuilder) RowMapper {
	return RowMapper{urlBuilder: urlBuilder}
}

func (m RowMapper) BuildPublicURL(raw string) string {
	if m.urlBuilder == nil {
		return raw
	}
	return m.urlBuilder.BuildPublicURL(raw)
}

func (m RowMapper) ToListItem(row repository.StatementListRowRecord) ListItem {
	return ListItem{
		BaseItem: BaseItem{
			ID:           row.ID,
			Type:         row.Type,
			Amount:       row.Amount,
			Description:  row.Description,
			CategoryID:   row.CategoryID,
			AssetID:      row.AssetID,
			Title:        helper.StatementTitle(row),
			TargetObject: row.TargetObject,
			Mood:         row.Mood,
			Money:        fmt.Sprintf("%.2f", row.Amount),
			Category:     row.CategoryName,
			IconPath:     m.BuildPublicURL(row.IconPath),
			Asset:        row.AssetName,
			Date:         row.CreatedAt.Format("2006-01-02"),
			Time:         row.CreatedAt.Format("15:04:05"),
			TimeStr:      row.CreatedAt.Format("01-02 15:04"),
			Week:         helper.WeekdayCN(row.CreatedAt.Weekday()),
			Payee: Payee{
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
	}
}

func (m RowMapper) ToDetailItem(row repository.StatementListRowRecord) DetailItem {
	item := m.ToListItem(row)
	return DetailItem{
		BaseItem:    item.BaseItem,
		Location:    item.Location,
		Province:    item.Province,
		City:        item.City,
		Street:      item.Street,
		MonthDay:    item.MonthDay,
		HasPic:      item.HasPic,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		UploadFiles: []interface{}{},
	}
}

func NormalizeStatementType(raw string, fallback string) string {
	statementType := strings.TrimSpace(raw)
	if statementType == "" {
		return fallback
	}
	return statementType
}
