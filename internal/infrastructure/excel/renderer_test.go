package excel

import (
	"bytes"
	"github.com/yigger/jiezhang-backend/internal/types"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestExportWorkbookContract(t *testing.T) {
	data, err := (Renderer{}).Render([]types.ExportRow{{Category: "午餐", ParentCategory: "餐饮", TypeName: "支出", Asset: "现金", Description: "=SUM(1,2)", Amount: 12.5}})
	if err != nil {
		t.Fatal(err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	sheet := book.GetSheetName(0)
	for cell, want := range map[string]string{"A1": "子分类", "H1": "更新时间", "A2": "午餐", "E2": "=SUM(1,2)", "F2": "12.5"} {
		got, err := book.GetCellValue(sheet, cell)
		if err != nil || got != want {
			t.Fatalf("%s=%q err=%v; want %q", cell, got, err, want)
		}
	}
	formula, err := book.GetCellFormula(sheet, "E2")
	if err != nil || formula != "" {
		t.Fatalf("description became formula: %s %v", formula, err)
	}
}
