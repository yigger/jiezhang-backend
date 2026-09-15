package model

// SchemaMigration maps the schema_migrations table.
type SchemaMigration struct {
	Version string `gorm:"column:version;type:varchar(255);not null;primaryKey;autoIncrement:false"`
}

func (SchemaMigration) TableName() string { return "schema_migrations" }
