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

	indexStatements := []string{
		"CREATE INDEX IF NOT EXISTS idx_collections_grade_id ON collections (grade_id)",
		"CREATE INDEX IF NOT EXISTS idx_collections_release_type ON collections (release_type)",
		"CREATE INDEX IF NOT EXISTS idx_collections_manufacturer ON collections (manufacturer)",
		"CREATE INDEX IF NOT EXISTS idx_collections_series_id ON collections (series_id)",
		"CREATE INDEX IF NOT EXISTS idx_collections_status ON collections (status)",
		"CREATE INDEX IF NOT EXISTS idx_pictures_collection_id_deleted_at ON pictures (collection_id, deleted_at)",
		"CREATE INDEX IF NOT EXISTS idx_addons_collection_id_deleted_at ON addons (collection_id, deleted_at)",
	}
	for _, stmt := range indexStatements {
		if err := db.Exec(stmt).Error; err != nil {
			log.Fatalf("create index failed: %v", err)
		}
	}

	log.Println("migration complete")
}
