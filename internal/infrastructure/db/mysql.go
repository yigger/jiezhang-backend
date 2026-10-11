package db

import (
	"time"
	_ "time/tzdata"

	driver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQL(dsn string) (*gorm.DB, error) {
	return openMySQL(dsn, &gorm.Config{})
}

// NewMySQLQuiet prevents importing personal bill values into SQL error logs.
func NewMySQLQuiet(dsn string) (*gorm.DB, error) {
	options, err := driver.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	// DATETIME has no zone. Keep the exported China wall time even on a UTC host.
	options.Loc, options.ParseTime = location, true
	return openMySQL(options.FormatDSN(), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

func openMySQL(dsn string, options *gorm.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), options)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return db, nil
}
