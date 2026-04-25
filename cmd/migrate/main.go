package main

import (
	"log"

	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/config"
	"github.com/iotatfan/hobby-collection-be/pkg/database/gorm"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("config error: %v", err)
	}

	db := gorm.NewDB(&config.GetConfig().Postgres)
	if err := db.AutoMigrate(
		&entity.Scale{},
		&entity.Grade{},
		&entity.CollectionType{},
		&entity.ReleaseType{},
		&entity.Manufacturer{},
		&entity.Series{},
		&entity.Collection{},
		&entity.Picture{},
		&entity.Addon{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_collections_grade_deleted ON collections(grade_id, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_collections_release_type_deleted ON collections(release_type, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_collections_manufacturer_deleted ON collections(manufacturer, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_collections_series_deleted ON collections(series_id, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_grades_collection_type_deleted ON grades(collection_type_id, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_scales_deleted ON scales(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_release_types_deleted ON release_types(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_manufacturers_deleted ON manufacturers(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_series_deleted ON series(deleted_at)`,
	}
	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Fatalf("create index failed: %v", err)
		}
	}

	log.Println("migration complete")
}
