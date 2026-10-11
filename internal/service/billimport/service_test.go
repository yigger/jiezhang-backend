package billimport

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

type tableReader struct{ table Table }

func (r tableReader) Read(context.Context, io.Reader, string) ([]Table, error) {
	return []Table{r.table}, nil
}
func alipayTable(rows ...[]string) Table {
	head := []string{"交易时间", "交易分类", "交易对方", "对方账号", "商品说明", "收/支", "金额", "收/付款方式", "交易状态", "交易订单号", "商家订单号", "备注", ""}
	all := append([][]string{{"支付宝交易明细"}, head}, rows...)
	var table Table
	for _, row := range all {
		var cells []Cell
		for _, v := range row {
			cells = append(cells, Cell{Text: v})
		}
		table.Rows = append(table.Rows, cells)
	}
	return table
}
func ordinary(id string) []string {
	return []string{"2026-01-01 08:02:03", "餐饮美食", "示例商户", "", "早餐", "支出", "12.30", "余额", "交易成功", id, "merchant-id", "", ""}
}

type memoryRepository struct {
	catalog                repo.ImportCatalog
	statements             []model.Statement
	transaction, committed bool
	failure                error
}

func baseline() *memoryRepository {
	return &memoryRepository{catalog: repo.ImportCatalog{
		Book:       model.AccountBook{ID: 2, UserID: 1},
		Categories: []model.Category{{ID: 10, AccountBookID: 2, Name: "伙食餐饮", Type: "expend"}, {ID: 11, AccountBookID: 2, ParentID: 10, Name: "早饭", Type: "expend"}},
		Assets:     []model.Asset{{ID: 20, AccountBookID: 2, Name: "虚拟账户", Type: "deposit"}, {ID: 21, AccountBookID: 2, ParentID: 20, Name: "支付宝", Type: "deposit", Amount: 88.88}},
	}}
}
func (m *memoryRepository) Catalog(context.Context, int64, int64) (repo.ImportCatalog, error) {
	catalog := m.catalog
	catalog.Statements = append(catalog.Statements, m.statements...)
	return catalog, nil
}
func (m *memoryRepository) WithinImportTransaction(ctx context.Context, _ int64, _ int64, fn func(repo.StatementImportMutation, repo.ImportCatalog) error) error {
	copy := *m
	copy.catalog.Categories = append([]model.Category{}, m.catalog.Categories...)
	copy.catalog.Assets = append([]model.Asset{}, m.catalog.Assets...)
	copy.catalog.Statements = append([]model.Statement{}, m.catalog.Statements...)
	copy.statements = append([]model.Statement{}, m.statements...)
	copy.transaction = true
	catalog, _ := copy.Catalog(ctx, 0, 0)
	if err := fn(&copy, catalog); err != nil {
		return err
	}
	copy.transaction, copy.committed = false, true
	*m = copy
	return nil
}
func (m *memoryRepository) CreateCategory(_ context.Context, row model.Category) (int64, error) {
	if !m.transaction || row.ParentID < 0 {
		panic("category outside transaction or unresolved parent")
	}
	row.ID = int64(100 + len(m.catalog.Categories))
	m.catalog.Categories = append(m.catalog.Categories, row)
	return row.ID, nil
}
func (m *memoryRepository) CreateAsset(_ context.Context, row model.Asset) (int64, error) {
	if !m.transaction || row.ParentID < 0 || row.Amount != 0 {
		panic("invalid auto-created asset")
	}
	row.ID = int64(200 + len(m.catalog.Assets))
	m.catalog.Assets = append(m.catalog.Assets, row)
	return row.ID, nil
}
func (m *memoryRepository) CreateStatement(_ context.Context, row model.Statement, effect repo.BalanceEffect) (int64, error) {
	if !m.transaction || row.CategoryID <= 0 || row.AssetID <= 0 || effect != (repo.BalanceEffect{}) || row.BalanceApplied == nil || *row.BalanceApplied || row.ImportKey == nil || row.ImportFingerprint == nil {
		panic("historical write changed balance or lost marker")
	}
	if m.failure != nil {
		return 0, m.failure
	}
	row.ID = int64(300 + len(m.statements))
	m.statements = append(m.statements, row)
	return row.ID, nil
}
func runImport(t *testing.T, m *memoryRepository, table Table, apply bool, mapping Mapping) (Report, error) {
	t.Helper()
	return New(m, tableReader{table}).Import(context.Background(), Input{UserID: 1, AccountBookID: 2, Files: []File{{Reader: strings.NewReader(""), Format: "csv"}}, Apply: apply, Mapping: mapping})
}

func TestHistoricalPreviewApplyAndRepeat(t *testing.T) {
	m := baseline()
	table := alipayTable(ordinary("000000000000000000000001"))
	report, err := runImport(t, m, table, false, Mapping{})
	if err != nil || report.Ready != 1 || report.Rows[0].CategoryID != 11 || len(report.Creations) != 0 || len(m.statements) != 0 || m.committed {
		t.Fatalf("preview: %+v %v", report, err)
	}
	report, err = runImport(t, m, table, true, Mapping{})
	if err != nil || !report.Applied || report.Imported != 1 || m.catalog.Assets[1].Amount != 88.88 || m.statements[0].CreatedAt.Format("15:04:05 -0700") != "08:02:03 +0800" {
		t.Fatalf("apply: %+v %v", report, err)
	}
	report, err = runImport(t, m, table, true, Mapping{})
	if err != nil || report.Duplicate != 1 || report.Imported != 0 || len(m.statements) != 1 {
		t.Fatalf("repeat: %+v %v", report, err)
	}
	// User edits must not change the original source identity or fingerprint.
	m.statements[0].Amount, m.statements[0].TargetObject = 99, "用户修改的对象"
	report, err = runImport(t, m, table, true, Mapping{})
	if err != nil || report.Duplicate != 1 || report.Imported != 0 || m.statements[0].Amount != 99 {
		t.Fatalf("repeat after edit: %+v %v", report, err)
	}
	// With identity stored on the statement, deleting it permits a fresh import.
	m.statements = nil
	report, err = runImport(t, m, table, true, Mapping{})
	if err != nil || report.Duplicate != 0 || report.Imported != 1 || len(m.statements) != 1 {
		t.Fatalf("repeat after deletion: %+v %v", report, err)
	}
}

func TestDistinctOrdersWithSameTimeAmountAndCounterpartyBothImport(t *testing.T) {
	m := baseline()
	report, err := runImport(t, m, alipayTable(ordinary("order-a"), ordinary("order-b")), true, Mapping{})
	if err != nil || report.Imported != 2 || report.Duplicate != 0 || len(m.statements) != 2 || *m.statements[0].ImportKey == *m.statements[1].ImportKey {
		t.Fatalf("distinct purchases collapsed: %+v %v", report, err)
	}
}

func TestImportIdentitySeparatesPlatformsAndScopes(t *testing.T) {
	if importKey("alipay", "same-order") == importKey("wechat", "same-order") {
		t.Fatal("different platforms share an import key")
	}
	m := baseline()
	key, fingerprint := importKey("alipay", "order-a"), "other-content"
	m.catalog.Statements = []model.Statement{
		{AccountBookID: 99, UserID: 1, ImportKey: &key, ImportFingerprint: &fingerprint},
		{AccountBookID: 2, UserID: 99, ImportKey: &key, ImportFingerprint: &fingerprint},
		{AccountBookID: 2, UserID: 1}, // Ordinary statements have NULL import metadata.
	}
	report, err := runImport(t, m, alipayTable(ordinary("order-a")), false, Mapping{})
	if err != nil || report.Ready != 1 || report.Errors != 0 || report.Duplicate != 0 {
		t.Fatalf("unrelated statement affected deduplication: %+v %v", report, err)
	}
}
func TestCreationReusesPlannedParentsAndLeaves(t *testing.T) {
	m := baseline()
	m.catalog.Categories = nil
	m.catalog.Assets = nil
	report, err := runImport(t, m, alipayTable(ordinary("a"), ordinary("b")), true, Mapping{})
	if err != nil || report.Imported != 2 || len(m.catalog.Categories) != 2 || len(m.catalog.Assets) != 2 || m.statements[0].CategoryID != m.statements[1].CategoryID || m.statements[0].AssetID != m.statements[1].AssetID {
		t.Fatalf("creation duplicated nodes: %+v %v", report, err)
	}
}
func TestStatementFailureRollsBackEntireBatch(t *testing.T) {
	m := baseline()
	m.catalog.Categories = nil
	m.catalog.Assets = nil
	m.failure = errors.New("write failed")
	report, err := runImport(t, m, alipayTable(ordinary("a")), true, Mapping{})
	if !errors.Is(err, ErrStorage) || report.Applied || report.Imported != 0 || len(m.catalog.Categories) != 0 || len(m.catalog.Assets) != 0 || len(m.statements) != 0 {
		t.Fatalf("partial commit: %+v %v", report, err)
	}
}
func TestSourceConflictRejectsWholeBatch(t *testing.T) {
	m := baseline()
	if _, err := runImport(t, m, alipayTable(ordinary("a")), true, Mapping{}); err != nil {
		t.Fatal(err)
	}
	row := ordinary("a")
	row[6] = "13.30"
	report, err := runImport(t, m, alipayTable(row, ordinary("b")), true, Mapping{})
	if !errors.Is(err, ErrInvalid) || report.Errors != 1 || len(m.statements) != 1 || report.Rows[0].Reason != "source_order_changed" {
		t.Fatalf("conflict: %+v %v", report, err)
	}
}
func TestSpecialTypesSkipWithoutCreatingNodes(t *testing.T) {
	m := baseline()
	var rows [][]string
	for i, category := range []string{"转账红包", "信用借还", "账户存取", "投资理财", "退款"} {
		row := ordinary(string(rune('a' + i)))
		row[1] = category
		rows = append(rows, row)
	}
	closed := ordinary("closed")
	closed[8] = "交易关闭"
	rows = append(rows, closed)
	partial := ordinary("partial")
	partial[8] = "已退款¥3.00"
	rows = append(rows, partial)
	mixed := ordinary("mixed")
	mixed[7] = "余额&红包"
	rows = append(rows, mixed)
	report, err := runImport(t, m, alipayTable(rows...), true, Mapping{})
	if err != nil || report.Skipped != len(rows) || report.Imported != 0 || len(report.Creations) != 0 {
		t.Fatalf("special rows imported: %+v %v", report, err)
	}
	row := ordinary("phone")
	row[1], row[4] = "充值缴费", "手机话费"
	report, err = runImport(t, m, alipayTable(row), false, Mapping{})
	if err != nil || report.Ready != 1 {
		t.Fatalf("ordinary top-up skipped: %+v %v", report, err)
	}
}
func TestAuthorizationAndCreationPrivileges(t *testing.T) {
	for _, member := range []bool{false, true} {
		m := baseline()
		m.catalog.Book.UserID = 9
		if member {
			m.catalog.Members = []model.AccountBookCollaborator{{AccountBookID: 2, UserID: 1, Role: "member"}}
		}
		report, err := runImport(t, m, alipayTable(ordinary("a")), false, Mapping{})
		if !member && !errors.Is(err, ErrInvalid) {
			t.Fatal("nonmember authorized")
		}
		if member && (err != nil || report.Ready != 1) {
			t.Fatalf("member could not reuse: %+v %v", report, err)
		}
		m.catalog.Categories = nil
		report, err = runImport(t, m, alipayTable(ordinary("b")), true, Mapping{})
		if !errors.Is(err, ErrInvalid) || report.Applied || len(m.statements) != 0 {
			t.Fatalf("unauthorized creation: %+v %v", report, err)
		}
	}
}
func TestMappingSimilarityAndAmbiguity(t *testing.T) {
	m := baseline()
	row := ordinary("a")
	m.catalog.Categories = append(m.catalog.Categories, model.Category{ID: 12, AccountBookID: 2, ParentID: 10, Name: "早餐", Type: "expend"})
	report, err := runImport(t, m, alipayTable(row), false, Mapping{})
	if err != nil || report.Rows[0].CategoryID != 12 {
		t.Fatalf("exact name did not win: %+v %v", report, err)
	}
	m.catalog.Categories[2].Name = "早点"
	report, err = runImport(t, m, alipayTable(row), true, Mapping{})
	if !errors.Is(err, ErrInvalid) || report.Errors != 1 || len(m.statements) != 0 {
		t.Fatalf("ambiguous aliases merged: %+v %v", report, err)
	}
	mapping := Mapping{Categories: map[string]Target{"expend:餐饮美食": {ID: 11}}}
	report, err = runImport(t, m, alipayTable(row), true, mapping)
	if err != nil || report.Rows[0].CategoryID != 11 {
		t.Fatalf("explicit mapping failed: %+v %v", report, err)
	}
	if categoryName("早餐") == categoryName("晚餐") || assetName("建设银行(1234)") == assetName("建设银行(5678)") {
		t.Fatal("different meals/accounts merged")
	}
}
func TestParserRejectsInvalidRowsAndPreservesIDs(t *testing.T) {
	m := baseline()
	for _, bad := range []string{"NaN", "Inf", "-1", "1.001", "10000000000", "1e3"} {
		row := ordinary("a")
		row[6] = bad
		report, err := runImport(t, m, alipayTable(row), true, Mapping{})
		if !errors.Is(err, ErrInvalid) || report.Errors != 1 || len(m.statements) != 0 {
			t.Fatalf("accepted amount %q", bad)
		}
	}
	zero := ordinary("zero")
	zero[6] = "0.00"
	zeroReport, zeroErr := runImport(t, m, alipayTable(zero), true, Mapping{})
	if zeroErr != nil || zeroReport.Skipped != 1 || zeroReport.Rows[0].Reason != "zero_amount" {
		t.Fatal("zero payment must be explicitly skipped")
	}
	row := ordinary("00000012345678901234567890\t")
	table := alipayTable(row)
	entries, err := New(m, tableReader{table}).parse(context.Background(), File{Reader: strings.NewReader(""), Format: "csv"}, 1)
	if err != nil || entries[0].ID != "00000012345678901234567890" {
		t.Fatalf("ID damaged: %v", err)
	}
	table.Rows[2][9].Numeric = true
	report, err := runImport(t, m, table, true, Mapping{})
	if !errors.Is(err, ErrInvalid) || report.Rows[0].Reason != "order_id_must_be_text" {
		t.Fatalf("numeric long ID accepted: %+v %v", report, err)
	}
	table = alipayTable(ordinary("formula"))
	table.Rows[2][6].Formula = true
	report, err = runImport(t, m, table, true, Mapping{})
	if !errors.Is(err, ErrInvalid) || report.Rows[0].Reason != "formula_not_allowed" {
		t.Fatal("formula accepted")
	}
}
func TestExcelDatesIgnoreHostTimezone(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	for _, date1904 := range []bool{false, true} {
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		if date1904 {
			base = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)
		}
		want := time.Date(2026, 1, 1, 8, 2, 3, 0, loc)
		wallTime := time.Date(2026, 1, 1, 8, 2, 3, 0, time.UTC)
		cell := Cell{Text: strconv.FormatFloat(wallTime.Sub(base).Hours()/24, 'f', -1, 64), Numeric: true}
		got, err := parseDate(cell, date1904, loc)
		if err != nil || !got.Equal(want) {
			t.Fatalf("date1904=%v got %v want %v", date1904, got, want)
		}
	}
}

func TestIncomeAndSameFileDuplicates(t *testing.T) {
	m := baseline()
	income := ordinary("salary")
	income[1], income[4], income[5], income[6] = "收入", "示例收入", "收入", "100.01"
	report, err := runImport(t, m, alipayTable(income, ordinary("meal"), ordinary("meal")), true, Mapping{})
	if err != nil || report.Imported != 2 || report.Duplicate != 1 || m.statements[0].Type != "income" || m.statements[0].Amount != 100.01 || m.catalog.Assets[1].Amount != 88.88 {
		t.Fatalf("income/duplicate: %+v %v", report, err)
	}
}

func TestBankCardAliasesPreserveAccountIdentity(t *testing.T) {
	if assetName("中国工商银行储蓄卡（1234）") != assetName("工行尾号1234") {
		t.Fatal("equivalent card names not recognized")
	}
	if assetName("工行尾号1234") == assetName("工行尾号5678") {
		t.Fatal("different cards merged")
	}
	m := baseline()
	m.catalog.Assets[1].Type = ""
	report, err := runImport(t, m, alipayTable(ordinary("legacy-zero")), false, Mapping{})
	if err != nil || len(report.Creations) != 0 || report.Rows[0].AssetID != 21 {
		t.Fatalf("legacy empty type created duplicate: %+v %v", report, err)
	}
}
