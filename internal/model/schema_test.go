package model_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/yigger/jiezhang-backend/internal/model"
	"gorm.io/gorm/schema"
)

// This fixture contains captured metadata and explicitly marked planned migration metadata, never application rows.
func TestModelsMatchDatabaseSchema(t *testing.T) {
	data, err := os.ReadFile("testdata/schema_snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Tables []struct {
			Name string `json:"TABLE_NAME"`
		} `json:"tables"`
		Columns []struct {
			Table    string  `json:"TABLE_NAME"`
			Name     string  `json:"COLUMN_NAME"`
			Type     string  `json:"COLUMN_TYPE"`
			Nullable string  `json:"IS_NULLABLE"`
			Key      string  `json:"COLUMN_KEY"`
			Extra    string  `json:"EXTRA"`
			Default  *string `json:"COLUMN_DEFAULT"`
		} `json:"columns"`
		Indexes []struct {
			Table     string `json:"TABLE_NAME"`
			Name      string `json:"INDEX_NAME"`
			Column    string `json:"COLUMN_NAME"`
			NonUnique int    `json:"NON_UNIQUE"`
			Sequence  int    `json:"SEQ_IN_INDEX"`
			Length    *int   `json:"SUB_PART"`
		} `json:"indexes"`
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatal(err)
	}
	models := []any{
		&model.CalendarJournal{},
		&model.InsightMerchantAlias{},
		&model.InsightProject{},
		&model.InsightFixedCost{},
		&model.InsightFixedCostRun{},
		&model.InsightStatementAnnotation{},
		&model.InsightPortfolioSnapshot{},
		&model.AccountBookCollaborator{},
		&model.AccountBookFriendRef{},
		&model.AccountBook{},
		&model.ARInternalMetadata{},
		&model.AssetLog{},
		&model.AssetSnapshot{},
		&model.Asset{},
		&model.BonusPointsLog{},
		&model.Category{},
		&model.ErrorLog{},
		&model.Feedback{},
		&model.FriendApply{},
		&model.Friend{},
		&model.Message{},
		&model.MonthChart{},
		&model.OperateLog{},
		&model.Payee{},
		&model.PreOrder{},
		&model.Recommend{},
		&model.SchemaMigration{},
		&model.Statement{},
		&model.UserAsset{},
		&model.User{},
		&model.UserAssetAssignment{},
	}
	parsed := map[string]*schema.Schema{}
	for _, value := range models {
		s, err := schema.Parse(value, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			t.Fatal(err)
		}
		if parsed[s.Table] != nil {
			t.Fatalf("duplicate model for %s", s.Table)
		}
		parsed[s.Table] = s
	}
	if len(parsed) != len(snapshot.Tables) {
		t.Fatalf("models=%d, database tables=%d", len(parsed), len(snapshot.Tables))
	}
	for _, table := range snapshot.Tables {
		t.Run(table.Name, func(t *testing.T) {
			s := parsed[table.Name]
			if s == nil {
				t.Fatal("missing table model")
			}
			columns := 0
			for _, col := range snapshot.Columns {
				if col.Table != table.Name {
					continue
				}
				columns++
				f := s.FieldsByDBName[col.Name]
				if f == nil {
					t.Errorf("missing column %s", col.Name)
					continue
				}
				if f.TagSettings["TYPE"] != col.Type || f.NotNull != (col.Nullable == "NO") || f.PrimaryKey != (col.Key == "PRI") || f.AutoIncrement != strings.Contains(col.Extra, "auto_increment") {
					t.Errorf("%s: type/nullability/primary key/auto increment differs from database: %v", col.Name, f.TagSettings)
				}
				gotDefault, hasDefault := f.TagSettings["DEFAULT"]
				if hasDefault != (col.Default != nil) || (col.Default != nil && strings.Trim(gotDefault, "'") != *col.Default) {
					t.Errorf("%s: default differs from database: %q", col.Name, gotDefault)
				}
			}
			if len(s.DBNames) != columns {
				t.Errorf("model columns=%d, database columns=%d", len(s.DBNames), columns)
			}
			type indexColumn struct {
				Name   string
				Length int
			}
			type indexDef struct {
				Unique  bool
				Columns []indexColumn
			}
			expected := map[string]indexDef{}
			for _, idx := range snapshot.Indexes {
				if idx.Table != table.Name || idx.Name == "PRIMARY" {
					continue
				}
				value := expected[idx.Name]
				value.Unique = idx.NonUnique == 0
				length := 0
				if idx.Length != nil {
					length = *idx.Length
				}
				if idx.Sequence != len(value.Columns)+1 {
					t.Fatal("snapshot index columns are out of order")
				}
				value.Columns = append(value.Columns, indexColumn{idx.Column, length})
				expected[idx.Name] = value
			}
			actual := map[string]indexDef{}
			for _, idx := range s.ParseIndexes() {
				value := indexDef{Unique: idx.Class == "UNIQUE"}
				for _, f := range idx.Fields {
					value.Columns = append(value.Columns, indexColumn{f.DBName, f.Length})
				}
				actual[idx.Name] = value
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("indexes differ: got %#v, want %#v", actual, expected)
			}
		})
	}
}
