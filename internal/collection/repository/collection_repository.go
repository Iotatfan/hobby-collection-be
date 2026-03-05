package repository

import (
	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/helper"

	"gorm.io/gorm"
)

type CollectionRepository interface {
	GetCollectionByID(id int) (collectionEntity.Collection, error)
	GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionList, error)
	GetPicturesByCollectionID(id int) ([]collectionEntity.Picture, error)
	UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.Collection, error)
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
	err := r.db.Model(&collectionEntity.Collection{}).Joins("CollectionType").Preload("CollectionType.Grade").Joins("Series").Joins("ReleaseType").Preload("Pictures").Find(&collection, id).Error
	if err != nil {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: err}
	}

	return collection, nil
}

func (r *collectionRepository) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionList, error) {
	collectionList := collectionEntity.CollectionList{}
	db := r.db.Model(&collectionEntity.Collection{})

	if filters.CollectionTypeID >= 0 || filters.GradeID >= 0 {
		// db.Joins("left join collection_types on collection_types.id = collections.type_id").Where("collection_types.grade_id = ? ", filters.CollectionTypeID)

	}

	result := db.Joins("CollectionType").Preload("CollectionType.Grade").Joins("Series").Joins("ReleaseType").Find(&collectionList.Collections)
	if result.Error != nil {
		return collectionEntity.CollectionList{}, helper.DBError{ErrorMsg: result.Error}
	}

	return collectionList, nil
}

func (r *collectionRepository) GetPicturesByCollectionID(id int) ([]collectionEntity.Picture, error) {
	pictures := []collectionEntity.Picture{}
	err := r.db.Model(&collectionEntity.Picture{}).Where("collection_id = ?", id).Find(&pictures).Error
	if err != nil {
		return []collectionEntity.Picture{}, helper.DBError{ErrorMsg: err}
	}
	return pictures, nil
}

func (r *collectionRepository) UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.Collection, error) {
	collection := collectionEntity.Collection{
		TypeID:         payload.Type.ID,
		Title:          payload.Title,
		ReleaseTypeID:  payload.ReleaseType.ID,
		Status:         payload.Status,
		SeriesID:       payload.Series.ID,
		Cover:          payload.Cover,
		Description:    payload.Description,
		CollectionType: payload.Type,
		ReleaseType:    payload.ReleaseType,
		Series:         payload.Series,
	}

	if !payload.BuiltAt.IsZero() {
		builtAt := payload.BuiltAt
		collection.BuiltAt = &builtAt
	}

	pictures := make([]collectionEntity.Picture, 0, len(payload.Pictures))
	for _, picture := range payload.Pictures {
		if picture.Url == "" {
			continue
		}
		pictures = append(pictures, collectionEntity.Picture{Url: picture.Url})
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&collection).Error; err != nil {
			return err
		}

		if len(pictures) == 0 {
			return nil
		}

		for i := range pictures {
			pictures[i].CollectionID = collection.ID
		}

		if err := tx.Create(&pictures).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: err}
	}

	return collection, nil
}
