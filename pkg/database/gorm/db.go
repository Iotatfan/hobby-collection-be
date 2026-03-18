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

	ensurePerformanceIndexes(db)

	return db
}

func ensurePerformanceIndexes(db *gorm.DB) {
	indexSQL := []string{
		"CREATE INDEX IF NOT EXISTS idx_collections_deleted_id_desc ON collections (deleted_at, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_pictures_collection_deleted_created_id ON pictures (collection_id, deleted_at, created_at DESC, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_addons_collection_deleted_created_id ON addons (collection_id, deleted_at, created_at DESC, id DESC)",
	}

	for _, sql := range indexSQL {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("[db] failed to ensure performance index: %v", err)
		}
	}
}
