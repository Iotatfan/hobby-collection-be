package repository

import (
	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/helper"

	"gorm.io/gorm"
)

type CollectionRepository interface {
	GetCollectionByID(id int) (collectionEntity.Collection, error)
	GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionList, error)
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

func (r *collectionRepository) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionList, error) {
	collectionList := collectionEntity.CollectionList{}
	db := r.db.Model(&collectionEntity.Collection{})

	if filters.CollectionTypeID >= 0 || filters.GradeID >= 0 {
		db.Joins("left join collection_type on collection_type.id = collection.type_id").Where("grade_id = ? ", filters.CollectionTypeID)
	}

	result := db.Preload("ReleaseType").Preload("Pictures").Find(&collectionList.Collections)
	if result.Error != nil {
		return collectionEntity.CollectionList{}, helper.DBError{ErrorMsg: result.Error}
	}

	return collectionList, nil
}
