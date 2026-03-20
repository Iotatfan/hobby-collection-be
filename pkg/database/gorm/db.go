package gorm

import (
	"fmt"
	"log"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 10
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

func NewDB(cfg *config.PostgresConfig) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	if err != nil {
		panic(fmt.Errorf("database error: %w", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Errorf("database pool error: %w", err))
	}
	sqlDB.SetMaxOpenConns(defaultMaxOpenConns)
	sqlDB.SetMaxIdleConns(defaultMaxIdleConns)
	sqlDB.SetConnMaxLifetime(defaultConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	return db
}
