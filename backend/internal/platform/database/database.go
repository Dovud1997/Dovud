package database

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dsn string, env string) (*gorm.DB, error) {
	cfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}
	if env == "local" || env == "dev" {
		cfg.Logger = logger.Default.LogMode(logger.Info)
	}

	var dialector gorm.Dialector
	if len(dsn) >= 7 && dsn[:7] == "sqlite:" {
		dialector = sqlite.Open(dsn[7:])
	} else {
		dialector = postgres.Open(dsn)
	}

	db, err := gorm.Open(dialector, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}
