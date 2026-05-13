package service

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func buildStatementsExcel(rows []StatementExportRowItem) ([]byte, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"子分类", "父分类", "类型", "资产", "备注", "金额", "创建时间", "更新时间"}

	for i, v := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, v); err != nil {
			return nil, err
		}
	}

	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet, "A1", "H1", style); err != nil {
		return nil, err
	}

	for idx, row := range rows {
		line := idx + 2
		values := []interface{}{
			row.Category,
			row.ParentCategory,
			row.TypeName,
			row.Asset,
			row.Description,
			row.Amount,
			row.CreatedAt,
			row.UpdatedAt,
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, line)
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return nil, err
			}
		}
	}

	_ = f.SetColWidth(sheet, "A", "A", 15)
	_ = f.SetColWidth(sheet, "B", "B", 15)
	_ = f.SetColWidth(sheet, "C", "C", 10)
	_ = f.SetColWidth(sheet, "D", "D", 15)
	_ = f.SetColWidth(sheet, "E", "E", 30)
	_ = f.SetColWidth(sheet, "F", "F", 15)
	_ = f.SetColWidth(sheet, "G", "G", 20)
	_ = f.SetColWidth(sheet, "H", "H", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}
