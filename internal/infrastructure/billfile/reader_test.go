package billfile

import (
	"bytes"
	"context"
	"encoding/csv"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"github.com/yigger/jiezhang-backend/internal/service/billimport"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type previewRepository struct{}

func (previewRepository) Catalog(context.Context, int64, int64) (repo.ImportCatalog, error) {
	return repo.ImportCatalog{Book: model.AccountBook{ID: 2, UserID: 1}}, nil
}
func (previewRepository) WithinImportTransaction(context.Context, int64, int64, func(repo.StatementImportMutation, repo.ImportCatalog) error) error {
	panic("preview attempted database mutation")
}

func TestGB18030CSVAndWeChatXLSX(t *testing.T) {
	var csvData bytes.Buffer
	w := csv.NewWriter(&csvData)
	_ = w.Write([]string{"支付宝交易明细"})
	_ = w.Write([]string{"交易时间", "交易分类", "交易对方", "商品说明", "收/支", "金额", "收/付款方式", "交易状态", "交易订单号", "备注", ""})
	_ = w.Write([]string{"2026-01-01 08:02:03", "餐饮美食", "示例商户", "早餐,套餐", "支出", "12.30", "余额", "交易成功", "00000012345678901234567890\t", "备注\n第二行", ""})
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	gb, err := simplifiedchinese.GB18030.NewEncoder().Bytes(csvData.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	service := billimport.New(previewRepository{}, Reader{})
	report, err := service.Import(context.Background(), billimport.Input{UserID: 1, AccountBookID: 2, Files: []billimport.File{{Reader: bytes.NewReader(gb), Format: "csv"}}})
	if err != nil || report.Ready != 1 || report.Errors != 0 || report.Rows[0].Row != 3 {
		t.Fatalf("CSV: %+v %v", report, err)
	}

	for _, date1904 := range []bool{false, true} {
		f := excelize.NewFile()
		if err := f.SetWorkbookProps(&excelize.WorkbookPropsOptions{Date1904: &date1904}); err != nil {
			t.Fatal(err)
		}
		head := []any{"交易时间", "交易类型", "交易对方", "商品", "收/支", "金额(元)", "支付方式", "当前状态", "交易单号", "商户单号", "备注"}
		if err := f.SetSheetRow("Sheet1", "A18", &head); err != nil {
			t.Fatal(err)
		}
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		if date1904 {
			base = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)
		}
		serial := time.Date(2026, 1, 1, 8, 2, 3, 0, time.UTC).Sub(base).Hours() / 24
		row := []any{serial, "商户消费", "示例商户", "早餐", "支出", 12.30, "零钱", "支付成功", "00000012345678901234567890", "/", "/"}
		if err := f.SetSheetRow("Sheet1", "A19", &row); err != nil {
			t.Fatal(err)
		}
		data, err := f.WriteToBuffer()
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
		report, err = service.Import(context.Background(), billimport.Input{UserID: 1, AccountBookID: 2, Files: []billimport.File{{Reader: bytes.NewReader(data.Bytes()), Format: "xlsx"}}})
		if err != nil || report.Ready != 1 || report.Errors != 0 || report.Rows[0].Row != 19 {
			t.Fatalf("XLSX date1904=%v: %+v %v", date1904, report, err)
		}
	}
}

func TestReaderRejectsOversizeAndCancelledInput(t *testing.T) {
	if _, err := (Reader{}).Read(context.Background(), strings.NewReader(strings.Repeat("x", billimport.MaxFileBytes+1)), "csv"); err == nil {
		t.Fatal("oversize file accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Reader{}).Read(ctx, strings.NewReader(""), "csv"); err != context.Canceled {
		t.Fatalf("cancellation: %v", err)
	}
}

// Opt-in, read-only parsing of locally supplied samples. Never copies source
// records into fixtures, prints records, or connects to a database.
func TestLocalSamplesPreview(t *testing.T) {
	directory := os.Getenv("BILL_IMPORT_SAMPLE_DIR")
	if directory == "" {
		t.Skip("BILL_IMPORT_SAMPLE_DIR is not configured")
	}
	var files []*os.File
	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()
	input := billimport.Input{UserID: 1, AccountBookID: 2}
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !(strings.HasPrefix(name, "支付宝交易明细(") && ext == ".csv") && !(strings.HasPrefix(name, "微信支付账单流水文件(") && ext == ".xlsx") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		files = append(files, f)
		input.Files = append(input.Files, billimport.File{Reader: f, Format: strings.TrimPrefix(ext, ".")})
		return nil
	})
	if err != nil {
		t.Fatal("cannot open local samples")
	}
	if len(input.Files) != 2 {
		t.Fatal("expected the two supplied bill samples")
	}
	report, err := billimport.New(previewRepository{}, Reader{}).Import(context.Background(), input)
	reasons := map[string]int{}
	for _, row := range report.Rows {
		if row.Status == "error" {
			reasons[row.Reason]++
		}
	}
	if len(reasons) > 0 {
		t.Logf("Error reason counts: %v", reasons)
	}
	if err != nil || report.Errors != 0 || report.Ready == 0 || report.Skipped == 0 {
		t.Fatalf("sample preview counts: ready=%d skipped=%d errors=%d; %v", report.Ready, report.Skipped, report.Errors, err)
	}
	t.Logf("Read-only sample preview: ready=%d skipped=%d errors=%d", report.Ready, report.Skipped, report.Errors)
}
