package insights_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	mysqlrepo "github.com/yigger/jiezhang-backend/internal/repo/mysql"
	"github.com/yigger/jiezhang-backend/internal/service/insights"
	"github.com/yigger/jiezhang-backend/internal/types"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// Runs only against an explicitly supplied disposable local MySQL socket.
// Never reads application credentials or connects to the configured application database.
func TestMySQLInsightLifecycle(t *testing.T) {
	socket := os.Getenv("INSIGHTS_TEST_MYSQL_SOCKET")
	if socket == "" {
		t.Skip("disposable MySQL socket not supplied")
	}
	if !strings.HasPrefix(socket, "/tmp/jiezhang-insights-db.") {
		t.Fatal("integration socket must belong to the disposable analytics fixture")
	}
	root, e := gorm.Open(gormmysql.Open(fmt.Sprintf("root@unix(%s)/?parseTime=true&loc=Local", socket)), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	schema := fmt.Sprintf("insight_test_%d", time.Now().UnixNano())
	if e = root.Exec("CREATE DATABASE " + schema + " CHARACTER SET utf8mb4").Error; e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { root.Exec("DROP DATABASE " + schema); sql, _ := root.DB(); sql.Close() })
	db, e := gorm.Open(gormmysql.Open(fmt.Sprintf("root@unix(%s)/%s?parseTime=true&loc=Local", socket, schema)), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { sql, _ := db.DB(); sql.Close() })
	fixtures := []string{
		"CREATE TABLE account_books (id bigint PRIMARY KEY,user_id int,name varchar(255)) ENGINE=InnoDB",
		"CREATE TABLE account_book_collaborators (id bigint PRIMARY KEY,account_book_id int,user_id int,remark varchar(255),role varchar(255),UNIQUE KEY book_user(account_book_id,user_id)) ENGINE=InnoDB",
		"CREATE TABLE users (id bigint PRIMARY KEY,nickname varchar(255)) ENGINE=InnoDB",
		"CREATE TABLE categories (id bigint PRIMARY KEY,account_book_id int,name varchar(255),type varchar(255)) ENGINE=InnoDB",
		"CREATE TABLE payees (id bigint PRIMARY KEY,account_book_id bigint,name varchar(255)) ENGINE=InnoDB",
		"CREATE TABLE assets (id bigint PRIMARY KEY,account_book_id int,parent_id int,name varchar(255),type varchar(255),amount decimal(12,2)) ENGINE=InnoDB",
		"CREATE TABLE statements (id bigint PRIMARY KEY,account_book_id int,user_id int,category_id int,asset_id int,amount decimal(12,2),type varchar(255),description text,created_at datetime,payee_id int) ENGINE=InnoDB",
		"INSERT INTO account_books VALUES(8,1,'fixture'),(9,3,'other')",
		"INSERT INTO account_book_collaborators VALUES(1,8,1,'所有者','owner'),(2,8,2,'成员','member')",
		"INSERT INTO users VALUES(1,'所有者'),(2,'成员'),(3,'其他')",
		"INSERT INTO categories VALUES(1,8,'订阅','expend')",
		"INSERT INTO payees VALUES(1,8,'Shop')",
		"INSERT INTO assets VALUES(1,8,10,'存款','deposit',100.10),(2,8,20,'信用卡','debt',20.20),(3,9,10,'外部资产','deposit',999.99)",
		"INSERT INTO statements VALUES(5,8,2,1,1,12.34,'expend','会员','2026-01-01 12:00:00',1),(6,9,3,1,3,99.99,'expend','其他账簿','2026-01-01 12:00:00',NULL)",
	}
	for _, query := range fixtures {
		if e = db.Exec(query).Error; e != nil {
			t.Fatal(e)
		}
	}
	first, e := os.ReadFile("../../../migrations/20261006_insights.sql")
	if e != nil {
		t.Fatal(e)
	}
	second, e := os.ReadFile("../../../migrations/20261007_project_details.sql")
	third, e := os.ReadFile("../../../migrations/20261007_fixed_cost_schedule.sql")
	if e != nil {
		t.Fatal(e)
	}
	fourth, e := os.ReadFile("../../../migrations/20261008_project_appearance.sql")
	if e != nil {
		t.Fatal(e)
	}
	migration := append(append(append(first, second...), third...), fourth...)
	if e != nil {
		t.Fatal(e)
	}
	var sqlLines []string
	for _, line := range strings.Split(string(migration), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			sqlLines = append(sqlLines, line)
		}
	}
	for _, statement := range strings.Split(strings.Join(sqlLines, "\n"), ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if e = db.Exec(statement).Error; e != nil {
			t.Fatal(e)
		}
	}
	ctx := context.Background()
	storage := mysqlrepo.NewInsightsStorage(db)
	service := insights.NewStorage(storage, mysqlrepo.NewInsightsRepository(db))
	p, e := service.SaveProject(ctx, 8, 2, types.InsightProjectInput{Name: "旅行", Icon: strptr("jcon-car"), Color: strptr("#abcDEF"), BudgetCents: 2000, ParticipantIDs: []int64{1, 2}, StartDate: strptr("2026-01-01"), EndDate: strptr("2026-01-31")})
	if e != nil {
		t.Fatal(e)
	}
	var persisted model.InsightProject
	if e = db.First(&persisted, p.ID).Error; e != nil || persisted.Icon == nil || *persisted.Icon != "jcon-car" || persisted.Color == nil || *persisted.Color != "#ABCDEF" {
		t.Fatal("project appearance not persisted", e)
	}
	f, e := service.SaveFixedCost(ctx, 8, 2, types.InsightFixedCostInput{Name: "会员", AmountCents: 1234, CategoryID: 1, AssetID: 1, IntervalMonths: 1, DueDay: 1, CandidateKey: "fixture", Active: true})
	if e != nil {
		t.Fatal(e)
	}
	duplicate, e := service.SaveFixedCost(ctx, 8, 2, types.InsightFixedCostInput{Name: "会员", AmountCents: 1234, CategoryID: 1, AssetID: 1, IntervalMonths: 1, DueDay: 1, CandidateKey: "fixture", Active: true})
	if e != nil || duplicate.ID != f.ID {
		t.Fatal("candidate confirmation not idempotent", e)
	}
	payer := int64(1)
	in := types.InsightAnnotationInput{StatementID: 5, ProjectID: &p.ID, PayerID: &payer, ConsumerID: &payer, FixedCostID: &f.ID, Allocations: []types.InsightAllocation{{MemberID: 1, AmountCents: 600}, {MemberID: 2, AmountCents: 634}}}
	if e = service.SaveAnnotation(ctx, 8, 2, in); e != nil {
		t.Fatal(e)
	}
	if e = service.SaveAnnotation(ctx, 8, 3, in); !errors.Is(e, insights.ErrForbidden) {
		t.Fatal("outsider annotation accepted", e)
	}
	in.StatementID = 6
	if e = service.SaveAnnotation(ctx, 8, 1, in); e == nil {
		t.Fatal("cross book annotation accepted")
	}
	in.StatementID = 5
	in.Allocations[0].AmountCents = 1
	if e = service.SaveAnnotation(ctx, 8, 2, in); !errors.Is(e, insights.ErrInvalidInput) {
		t.Fatal("bad split accepted", e)
	}
	snapshot, e := service.CapturePortfolio(ctx, 8, 2, "首次")
	if e != nil {
		t.Fatal(e)
	}
	if snapshot.AssetsCents != 10010 || snapshot.LiabilitiesCents != 2020 {
		t.Fatal("snapshot scope incorrect")
	}
	if e = db.Exec("UPDATE assets SET amount=110.10 WHERE id=1").Error; e != nil {
		t.Fatal(e)
	}
	if _, e = service.CapturePortfolio(ctx, 8, 2, "第二次"); e != nil {
		t.Fatal(e)
	}
	w, e := service.Workspace(ctx, model.AccountBook{ID: 8, UserID: 1}, 1)
	if e != nil {
		t.Fatal(e)
	}
	if len(w.Projects) != 1 || w.Projects[0].Summary.ExpendCents != 1234 || len(w.Payers) != 1 || w.Payers[0].ExpendCents != 1234 || len(w.Portfolio) != 2 || *w.Portfolio[1].DeltaCents != 1000 {
		t.Fatalf("workspace reconciliation failed %+v", w)
	}

	if len(w.Projects[0].ParticipantIDs) != 2 || w.Projects[0].StartDate == nil || *w.Projects[0].StartDate != "2026-01-01" || w.Annotations[0].ConsumerID == nil || *w.Annotations[0].ConsumerID != 1 {
		t.Fatal("project details not persisted", w)
	}
	if e = service.SaveMerchantAliases(ctx, 8, 2, types.InsightMerchantAliasInput{PayeeIDs: []int64{1}, Name: "统一商家"}); e != nil {
		t.Fatal(e)
	}
	report, e := insights.New(mysqlrepo.NewInsightsRepository(db)).Report(ctx, 8, 2026)
	if e != nil || len(report.Merchants) != 1 || report.Merchants[0].Name != "统一商家" {
		t.Fatal("merchant alias not applied", e)
	}
	if e = service.SaveMerchantAliases(ctx, 8, 2, types.InsightMerchantAliasInput{PayeeIDs: []int64{1}, Name: ""}); e != nil {
		t.Fatal(e)
	}
	report, e = insights.New(mysqlrepo.NewInsightsRepository(db)).Report(ctx, 8, 2026)
	if e != nil || report.Merchants[0].Name != "Shop" {
		t.Fatal("merchant alias reset failed", e)
	}
	var payee model.Payee
	db.First(&payee, 1)
	if payee.Name != "Shop" {
		t.Fatal("merchant alias changed original merchant")
	}
	// Expand only this disposable fixture to exercise the existing financial write.
	typ := reflect.TypeOf(model.Statement{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		parts := strings.Split(f.Tag.Get("gorm"), ";")
		col, sqlType := "", ""
		for _, part := range parts {
			if strings.HasPrefix(part, "column:") {
				col = strings.TrimPrefix(part, "column:")
			}
			if strings.HasPrefix(part, "type:") {
				sqlType = strings.TrimPrefix(part, "type:")
			}
		}
		var n int64
		db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=? AND table_name='statements' AND column_name=?", schema, col).Scan(&n)
		if n == 0 {
			if e = db.Exec("ALTER TABLE statements ADD COLUMN `" + col + "` " + sqlType + " NULL").Error; e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, q := range []string{"ALTER TABLE statements MODIFY id bigint NOT NULL AUTO_INCREMENT", "ALTER TABLE assets ADD COLUMN frequent int DEFAULT 0", "ALTER TABLE categories ADD COLUMN frequent int DEFAULT 0"} {
		if e = db.Exec(q).Error; e != nil {
			t.Fatal(e)
		}
	}
	writer := mysqlrepo.NewStatementRepository(db)
	create := func(consumer int64) (int64, error) {
		var id int64
		err := writer.WithinTransaction(ctx, func(tx repo.Mutation) error {
			var err error
			id, err = tx.Create(ctx, model.Statement{UserID: 2, AccountBookID: 8, AssetID: 1, CategoryID: 1, Amount: 5, Type: "expend", CreatedAt: time.Now(), TimeText: "12:00:00"}, repo.BalanceEffect{Source: -5})
			if err != nil {
				return err
			}
			return tx.(repo.ProjectMutation).AttachProject(ctx, 8, id, p.ID, consumer, 2)
		})
		return id, err
	}
	if _, e = create(3); e == nil {
		t.Fatal("invalid consumer accepted")
	}
	var amount float64
	db.Table("assets").Select("amount").Where("id=1").Scan(&amount)
	if amount != 110.10 {
		t.Fatal("failed project entry changed balance", amount)
	}
	var count int64
	db.Table("statements").Count(&count)
	if count != 2 {
		t.Fatal("failed project entry left statement", count)
	}
	id, err := create(2)
	if err != nil {
		t.Fatal(err)
	}
	var annotation model.InsightStatementAnnotation
	if e = db.Where("statement_id=?", id).Take(&annotation).Error; e != nil || annotation.ConsumerID == nil || *annotation.ConsumerID != 2 {
		t.Fatal("atomic project attachment failed", e)
	}
	db.Table("assets").Select("amount").Where("id=1").Scan(&amount)
	if amount != 105.10 {
		t.Fatal("valid entry balance incorrect", amount)
	}
	if _, e = service.SaveFixedCost(ctx, 8, 2, types.InsightFixedCostInput{ID: f.ID, Name: f.Name, AmountCents: f.AmountCents, CategoryID: f.CategoryID, AssetID: f.AssetID, IntervalMonths: f.IntervalMonths, DueDay: f.DueDay, CandidateKey: f.CandidateKey, Active: false}); e != nil {
		t.Fatal(e)
	}
	if _, e = service.SaveProject(ctx, 8, 2, types.InsightProjectInput{ID: p.ID, Name: p.Name, BudgetCents: 0, Archived: true}); e != nil {
		t.Fatal(e)
	}
	w, e = service.Workspace(ctx, model.AccountBook{ID: 8, UserID: 1}, 1)
	if e != nil || w.MonthlyFixedCents != 0 || w.Projects[0].BudgetCents != 0 || !w.Projects[0].Archived {
		t.Fatal("zero and false updates not persisted", e)
	}
	var saved model.InsightStatementAnnotation
	if e = db.First(&saved).Error; e != nil {
		t.Fatal(e)
	}
	var split []types.InsightAllocation
	if e = json.Unmarshal(saved.Allocations, &split); e != nil || split[0].AmountCents != 600 {
		t.Fatal("invalid split persisted")
	}
	var financial struct{ Amount float64 }
	if e = db.Table("statements").Select("amount").Where("id = ?", 5).Take(&financial).Error; e != nil || financial.Amount != 12.34 {
		t.Fatal("analysis altered financial amount")
	}
	t.Run("nightly fixed-cost lifecycle", func(t *testing.T) {
		r, err := service.SaveFixedCost(ctx, 8, 2, types.InsightFixedCostInput{Name: "定期会员", AmountCents: 125, CategoryID: 1, AssetID: 1, IntervalMonths: 1, DueDay: 31, Active: true})
		if err != nil {
			t.Fatal(err)
		}
		if r.NextRunDate == nil {
			t.Fatal("enabled schedule missing date")
		}
		if err = db.Exec("UPDATE insight_fixed_costs SET next_run_date='2026-01-31' WHERE id=?", r.ID).Error; err != nil {
			t.Fatal(err)
		}
		through := time.Date(2026, 2, 28, 0, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*3600))
		var wg sync.WaitGroup
		results := make(chan int, 2)
		failures := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); n, e := service.RunDueFixedCosts(ctx, through); results <- n; failures <- e }()
		}
		wg.Wait()
		close(results)
		close(failures)
		total := 0
		for n := range results {
			total += n
		}
		for e := range failures {
			if e != nil {
				t.Fatal(e)
			}
		}
		if total != 2 {
			t.Fatal("concurrent jobs duplicated or skipped occurrences", total)
		}
		var balance float64
		db.Table("assets").Where("id=1").Select("amount").Scan(&balance)
		if balance != 102.60 {
			t.Fatal("scheduled balance mismatch", balance)
		}
		var runs []model.InsightFixedCostRun
		db.Where("fixed_cost_id=?", r.ID).Order("due_date").Find(&runs)
		if len(runs) != 2 || runs[0].DueDate.Format("2006-01-02") != "2026-01-31" || runs[1].DueDate.Format("2006-01-02") != "2026-02-28" {
			t.Fatal("occurrence dates incorrect", runs)
		}
		if n, e := service.RunDueFixedCosts(ctx, through); e != nil || n != 0 {
			t.Fatal("repeat job generated extra statement", n, e)
		}
		var rule model.InsightFixedCost
		db.First(&rule, r.ID)
		if rule.NextRunDate == nil || rule.NextRunDate.Format("2006-01-02") != "2026-03-31" {
			t.Fatal("March day did not restore", rule)
		}
		var statementCount int64
		db.Table("statements").Count(&statementCount)
		if err = db.Exec("CREATE TRIGGER reject_fixed_run BEFORE INSERT ON insight_fixed_cost_runs FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='fixture failure'").Error; err != nil {
			t.Fatal(err)
		}
		march := time.Date(2026, 3, 31, 0, 0, 0, 0, through.Location())
		if n, e := service.RunDueFixedCosts(ctx, march); e == nil || n != 0 {
			t.Fatal("execution failure did not fail", n, e)
		}
		var after int64
		db.Table("statements").Count(&after)
		if after != statementCount {
			t.Fatal("failed execution left financial row")
		}
		db.Table("assets").Where("id=1").Select("amount").Scan(&balance)
		if balance != 102.60 {
			t.Fatal("failed execution changed wallet", balance)
		}
		db.First(&rule, r.ID)
		if rule.NextRunDate.Format("2006-01-02") != "2026-03-31" {
			t.Fatal("failed execution advanced schedule")
		}
		if err = db.Exec("DROP TRIGGER reject_fixed_run").Error; err != nil {
			t.Fatal(err)
		}
		if n, e := service.RunDueFixedCosts(ctx, march); e != nil || n != 1 {
			t.Fatal("failed occurrence did not recover", n, e)
		}
		db.First(&rule, r.ID)
		if rule.NextRunDate.Format("2006-01-02") != "2026-04-30" {
			t.Fatal("next date mismatch")
		}
		var generated model.Statement
		db.First(&generated, runs[0].StatementID)
		if generated.CategoryID != 1 || generated.AssetID != 1 || generated.Type != "expend" || generated.Amount != 1.25 {
			t.Fatal("incorrect generated statement")
		}
		var label model.InsightStatementAnnotation
		db.Where("statement_id=?", generated.ID).First(&label)
		if label.FixedCostID == nil || *label.FixedCostID != r.ID {
			t.Fatal("generated statement not annotated")
		}
		// Deleting a bill and moving the due day must not recreate the same month.
		db.Delete(&model.Statement{}, generated.ID)
		if err = db.Exec("UPDATE insight_fixed_costs SET next_run_date='2026-01-15' WHERE id=?", r.ID).Error; err != nil {
			t.Fatal(err)
		}
		if n, e := service.RunDueFixedCosts(ctx, march); e != nil || n != 0 {
			t.Fatal("deleted bill regenerated", n, e)
		}
		if _, err = service.SaveFixedCost(ctx, 8, 2, types.InsightFixedCostInput{ID: r.ID, Name: r.Name, AmountCents: r.AmountCents, CategoryID: r.CategoryID, AssetID: r.AssetID, IntervalMonths: 1, DueDay: 31, Active: false}); err != nil {
			t.Fatal(err)
		}
		if n, e := service.RunDueFixedCosts(ctx, march.AddDate(1, 0, 0)); e != nil || n != 0 {
			t.Fatal("paused schedule executed", n, e)
		}
		debtRule, err := service.SaveFixedCost(ctx, 8, 2, types.InsightFixedCostInput{Name: "信用卡固定扣款", AmountCents: 125, CategoryID: 1, AssetID: 2, IntervalMonths: 1, DueDay: 31, Active: true})
		if err != nil {
			t.Fatal(err)
		}
		if err = db.Exec("UPDATE insight_fixed_costs SET next_run_date='2026-01-31' WHERE id=?", debtRule.ID).Error; err != nil {
			t.Fatal(err)
		}
		if n, e := service.RunDueFixedCosts(ctx, through.AddDate(0, 0, -28)); e != nil || n != 1 {
			t.Fatal("selected credit wallet failed", n, e)
		}
		db.Table("assets").Where("id=2").Select("amount").Scan(&balance)
		if balance != 18.95 {
			t.Fatal("selected wallet differs from existing expenditure balance rules", balance)
		}
		if err = db.Exec("UPDATE insight_fixed_costs SET active=0 WHERE id=?", debtRule.ID).Error; err != nil {
			t.Fatal(err)
		}
		resumed, err := service.SaveFixedCost(ctx, 8, 2, types.InsightFixedCostInput{ID: r.ID, Name: r.Name, AmountCents: r.AmountCents, CategoryID: r.CategoryID, AssetID: r.AssetID, IntervalMonths: 1, DueDay: 31, Active: true})
		if err != nil || resumed.NextRunDate == nil {
			t.Fatal("rule could not resume", err)
		}
		today := time.Now().In(time.FixedZone("Asia/Shanghai", 8*3600)).Format("2006-01-02")
		if resumed.NextRunDate.Format("2006-01-02") < today {
			t.Fatal("resume backfilled paused interval")
		}

	})

}

func strptr(s string) *string { return &s }
