package repository

import (
	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/helper"

	"gorm.io/gorm"
)

type CollectionRepository interface {
	GetCollectionByID(id int) (collectionEntity.Collection, error)
}

type collectionRepository struct {
	db *gorm.DB
}

func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{
		db: db,
	}
}

func (r *collectionRepository) GetCollectionByID(id int) (collectionEntity.Collection, error) {
	collection := collectionEntity.Collection{}
	err := r.db.Preload("Scale").Preload("ReleaseType").Preload("Pictures").Take(&collection, id).Error
	if err != nil {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: err}
	}

	err = r.db.Model(&collectionEntity.Collection{}).Where("id = ?", collection.ID).Error
	if err != nil {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: err}
	}
	return collection, nil
}
