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
		`CREATE INDEX IF NOT EXISTS idx_addons_collection_deleted ON addons(collection_id, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_pictures_collection_deleted ON pictures(collection_id, deleted_at)`,
	}
	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Fatalf("create index failed: %v", err)
		}
	}

	log.Println("migration complete")
}
