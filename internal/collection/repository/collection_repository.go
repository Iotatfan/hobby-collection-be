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
	err := r.db.Model(&collectionEntity.Collection{}).
		Joins("CollectionType").
		Preload("CollectionType.Grade").
		Joins("Series").
		Joins("ReleaseType").
		Joins("Manufacturer").
		Preload("Pictures").
		First(&collection, id).Error
	if err != nil {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: err}
	}

	return collection, nil
}

func (r *collectionRepository) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionList, error) {
	collectionList := collectionEntity.CollectionList{}
	db := r.db.Model(&collectionEntity.Collection{})

	if filters.CollectionTypeID > 0 {
		db = db.Where("collections.type_id = ?", filters.CollectionTypeID)
	}
	if filters.GradeID > 0 {
		db = db.Joins("JOIN collection_types ct_filter ON ct_filter.id = collections.type_id").Where("ct_filter.grade_id = ?", filters.GradeID)
	}

	limit := filters.Limit
	offset := filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	result := db.Joins("CollectionType").
		Preload("CollectionType.Grade").
		Joins("Series").
		Joins("ReleaseType").
		Joins("Manufacturer").
		Order("collections.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&collectionList.Collections)
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
		TypeID:         payload.TypeID,
		Title:          payload.Title,
		ReleaseTypeID:  payload.ReleaseTypeID,
		ManufacturerID: payload.ManufacturerID,
		Status:         payload.Status,
		SeriesID:       payload.SeriesID,
		Cover:          payload.CoverURL,
		Description:    payload.Description,
	}

	if !payload.BuiltAt.IsZero() {
		builtAt := payload.BuiltAt
		collection.BuiltAt = &builtAt
	}

	pictures := make([]collectionEntity.Picture, 0, len(payload.PictureURLs))
	for _, pictureURL := range payload.PictureURLs {
		if pictureURL == "" {
			continue
		}
		pictures = append(pictures, collectionEntity.Picture{Url: pictureURL})
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
