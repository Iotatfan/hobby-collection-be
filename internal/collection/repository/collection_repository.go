package repository

import (
	"time"

	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/helper"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CollectionRepository interface {
	GetCollectionByID(id int) (collectionEntity.Collection, error)
	GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error)
	GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error)
	GetPicturesByCollectionID(id int) ([]collectionEntity.Picture, error)
	UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.Collection, error)
	UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest, deletePictureIDs []int) (collectionEntity.Collection, error)
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
	type collectionDetailRow struct {
		ID               int                                `gorm:"column:id"`
		Title            string                             `gorm:"column:title"`
		Status           collectionEntity.COLLECTION_STATUS `gorm:"column:status"`
		BuiltAt          *time.Time                         `gorm:"column:built_at"`
		Cover            string                             `gorm:"column:cover"`
		Description      string                             `gorm:"column:description"`
		TypeID           int                                `gorm:"column:type_id"`
		TypeName         string                             `gorm:"column:type_name"`
		TypeScale        string                             `gorm:"column:type_scale"`
		GradeID          int                                `gorm:"column:grade_id"`
		GradeScaleID     int                                `gorm:"column:grade_scale_id"`
		GradeName        string                             `gorm:"column:grade_name"`
		GradeShortName   string                             `gorm:"column:grade_short_name"`
		ReleaseTypeID    int                                `gorm:"column:release_type_id"`
		ReleaseTypeName  string                             `gorm:"column:release_type_name"`
		ManufacturerID   int                                `gorm:"column:manufacturer_id"`
		ManufacturerName string                             `gorm:"column:manufacturer_name"`
		SeriesID         int                                `gorm:"column:series_id"`
		SeriesName       string                             `gorm:"column:series_name"`
	}

	row := collectionDetailRow{}
	result := r.db.Table("collections c").
		Select(`
			c.id,
			c.title,
			c.status,
			c.built_at,
			c.cover,
			c.description,
			ct.id as type_id,
			ct.name as type_name,
			COALESCE(sc.name, '') as type_scale,
			COALESCE(g.id, 0) as grade_id,
			COALESCE(g.scale_id, 0) as grade_scale_id,
			COALESCE(g.name, '') as grade_name,
			COALESCE(g.short_name, '') as grade_short_name,
			COALESCE(rt.id, 0) as release_type_id,
			COALESCE(rt.name, '') as release_type_name,
			COALESCE(m.id, 0) as manufacturer_id,
			COALESCE(m.name, '') as manufacturer_name,
			COALESCE(s.id, 0) as series_id,
			COALESCE(s.name, '') as series_name
		`).
		Joins("JOIN collection_types ct ON ct.id = c.type_id AND ct.deleted_at IS NULL").
		Joins("LEFT JOIN grades g ON g.collection_type_id = ct.id AND g.deleted_at IS NULL").
		Joins("LEFT JOIN scales sc ON sc.id = g.scale_id AND sc.deleted_at IS NULL").
		Joins("LEFT JOIN release_types rt ON rt.id = c.release_type AND rt.deleted_at IS NULL").
		Joins("LEFT JOIN manufacturers m ON m.id = c.manufacturer AND m.deleted_at IS NULL").
		Joins("LEFT JOIN series s ON s.id = c.series_id AND s.deleted_at IS NULL").
		Where("c.id = ? AND c.deleted_at IS NULL", id).
		Limit(1).
		Scan(&row)
	if result.Error != nil {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: result.Error}
	}
	if result.RowsAffected == 0 {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: gorm.ErrRecordNotFound}
	}

	pictures := []collectionEntity.Picture{}
	if err := r.db.Model(&collectionEntity.Picture{}).
		Select("id", "collection_id", "url").
		Where("collection_id = ? AND deleted_at IS NULL", id).
		Order("created_at DESC").
		Order("id DESC").
		Find(&pictures).Error; err != nil {
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: err}
	}

	collection := collectionEntity.Collection{
		ID:             row.ID,
		TypeID:         row.TypeID,
		Title:          row.Title,
		ReleaseTypeID:  row.ReleaseTypeID,
		Status:         row.Status,
		ManufacturerID: row.ManufacturerID,
		SeriesID:       row.SeriesID,
		BuiltAt:        row.BuiltAt,
		Cover:          row.Cover,
		Description:    row.Description,
		CollectionType: collectionEntity.CollectionType{
			ID:                 row.TypeID,
			CollectionTypeName: row.TypeName,
			Scale:              row.TypeScale,
			Grade: collectionEntity.Grade{
				ID:               row.GradeID,
				Name:             row.GradeName,
				ShortName:        row.GradeShortName,
				ScaleID:          row.GradeScaleID,
				CollectionTypeID: row.TypeID,
			},
		},
		ReleaseType: collectionEntity.ReleaseType{
			ID:              row.ReleaseTypeID,
			ReleaseTypeName: row.ReleaseTypeName,
		},
		Manufacturer: collectionEntity.Manufacturer{
			ID:               row.ManufacturerID,
			ManufacturerName: row.ManufacturerName,
		},
		Series: collectionEntity.Series{
			ID:         row.SeriesID,
			SeriesName: row.SeriesName,
		},
		Pictures: &pictures,
	}

	return collection, nil
}

type collectionListItemRow struct {
	ID              int                                `gorm:"column:id"`
	Title           string                             `gorm:"column:title"`
	Status          collectionEntity.COLLECTION_STATUS `gorm:"column:status"`
	BuiltAt         *time.Time                         `gorm:"column:built_at"`
	Cover           string                             `gorm:"column:cover"`
	TypeID          int                                `gorm:"column:type_id"`
	TypeName        string                             `gorm:"column:type_name"`
	TypeScale       string                             `gorm:"column:type_scale"`
	GradeID         int                                `gorm:"column:grade_id"`
	GradeScaleID    int                                `gorm:"column:grade_scale_id"`
	GradeName       string                             `gorm:"column:grade_name"`
	GradeShortName  string                             `gorm:"column:grade_short_name"`
	ReleaseTypeID   int                                `gorm:"column:release_type_id"`
	ReleaseTypeName string                             `gorm:"column:release_type_name"`
	SeriesID        int                                `gorm:"column:series_id"`
	SeriesName      string                             `gorm:"column:series_name"`
}

func (r *collectionRepository) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error) {
	rows := []collectionListItemRow{}
	db := r.db.Table("collections c").
		Select("c.id").
		Where("c.deleted_at IS NULL")

	if filters.CollectionTypeID > 0 {
		db = db.Where("c.type_id = ?", filters.CollectionTypeID)
	}
	if filters.GradeID > 0 {
		db = db.Where("c.type_id IN (?)",
			r.db.Model(&collectionEntity.Grade{}).
				Select("collection_type_id").
				Where("id = ? AND deleted_at IS NULL", filters.GradeID),
		)
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

	ids := []int{}
	result := db.Order("c.id DESC").
		Limit(limit).
		Offset(offset).
		Pluck("c.id", &ids)
	if result.Error != nil {
		return collectionEntity.CollectionListResponse{}, helper.DBError{ErrorMsg: result.Error}
	}
	if len(ids) == 0 {
		return collectionEntity.CollectionListResponse{
			Collections: []collectionEntity.CollectionListItemResponse{},
		}, nil
	}

	result = r.db.Table("collections c").
		Select(`
			c.id,
			c.title,
			c.status,
			c.built_at,
			c.cover,
			ct.id as type_id,
			ct.name as type_name,
			COALESCE(sc.name, '') as type_scale,
			COALESCE(g.id, 0) as grade_id,
			COALESCE(g.scale_id, 0) as grade_scale_id,
			COALESCE(g.name, '') as grade_name,
			COALESCE(g.short_name, '') as grade_short_name,
			COALESCE(rt.id, 0) as release_type_id,
			COALESCE(rt.name, '') as release_type_name,
			COALESCE(s.id, 0) as series_id,
			COALESCE(s.name, '') as series_name
		`).
		Joins("JOIN collection_types ct ON ct.id = c.type_id AND ct.deleted_at IS NULL").
		Joins("LEFT JOIN grades g ON g.collection_type_id = ct.id AND g.deleted_at IS NULL").
		Joins("LEFT JOIN scales sc ON sc.id = g.scale_id AND sc.deleted_at IS NULL").
		Joins("LEFT JOIN release_types rt ON rt.id = c.release_type AND rt.deleted_at IS NULL").
		Joins("LEFT JOIN series s ON s.id = c.series_id AND s.deleted_at IS NULL").
		Where("c.id IN ?", ids).
		Order(clause.Expr{SQL: "c.id DESC"}).
		Scan(&rows)
	if result.Error != nil {
		return collectionEntity.CollectionListResponse{}, helper.DBError{ErrorMsg: result.Error}
	}

	response := collectionEntity.CollectionListResponse{
		Collections: make([]collectionEntity.CollectionListItemResponse, 0, len(rows)),
	}
	for _, row := range rows {
		builtAt := time.Time{}
		if row.BuiltAt != nil {
			builtAt = row.BuiltAt.Local()
		}

		response.Collections = append(response.Collections, collectionEntity.CollectionListItemResponse{
			ID:    row.ID,
			Title: row.Title,
			Type: collectionEntity.CollectionTypeResponse{
				ID:                 row.TypeID,
				CollectionTypeName: row.TypeName,
				Scale:              row.TypeScale,
				Grade: collectionEntity.Grade{
					ID:               row.GradeID,
					Name:             row.GradeName,
					ShortName:        row.GradeShortName,
					ScaleID:          row.GradeScaleID,
					CollectionTypeID: row.TypeID,
				},
			},
			ReleaseType: collectionEntity.ReleaseType{
				ID:              row.ReleaseTypeID,
				ReleaseTypeName: row.ReleaseTypeName,
			},
			Status: row.Status,
			Series: collectionEntity.Series{
				ID:         row.SeriesID,
				SeriesName: row.SeriesName,
			},
			BuiltAt: builtAt,
			Cover:   row.Cover,
		})
	}

	return response, nil
}

func (r *collectionRepository) GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error) {
	drawer := collectionEntity.CollectionDrawerResponse{}

	if err := r.db.Model(&collectionEntity.CollectionType{}).
		Preload("Grade.Scale").
		Order("name ASC").
		Find(&drawer.CollectionTypes).Error; err != nil {
		return collectionEntity.CollectionDrawerResponse{}, helper.DBError{ErrorMsg: err}
	}
	for i := range drawer.CollectionTypes {
		drawer.CollectionTypes[i].Scale = drawer.CollectionTypes[i].Grade.Scale.Name
	}

	if err := r.db.Model(&collectionEntity.ReleaseType{}).
		Order("name ASC").
		Find(&drawer.ReleaseTypes).Error; err != nil {
		return collectionEntity.CollectionDrawerResponse{}, helper.DBError{ErrorMsg: err}
	}

	if err := r.db.Model(&collectionEntity.Manufacturer{}).
		Order("name ASC").
		Find(&drawer.Manufacturers).Error; err != nil {
		return collectionEntity.CollectionDrawerResponse{}, helper.DBError{ErrorMsg: err}
	}

	if err := r.db.Model(&collectionEntity.Series{}).
		Order("name ASC").
		Find(&drawer.Series).Error; err != nil {
		return collectionEntity.CollectionDrawerResponse{}, helper.DBError{ErrorMsg: err}
	}

	return drawer, nil
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

func (r *collectionRepository) UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest, deletePictureIDs []int) (collectionEntity.Collection, error) {
	collection := collectionEntity.Collection{}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&collection, id).Error; err != nil {
			return err
		}

		updates := map[string]any{}
		if payload.Title != nil {
			updates["title"] = *payload.Title
		}
		if payload.TypeID != nil {
			updates["type_id"] = *payload.TypeID
		}
		if payload.ReleaseTypeID != nil {
			updates["release_type"] = *payload.ReleaseTypeID
		}
		if payload.ManufacturerID != nil {
			updates["manufacturer"] = *payload.ManufacturerID
		}
		if payload.Status != nil {
			updates["status"] = *payload.Status
		}
		if payload.SeriesID != nil {
			updates["series_id"] = *payload.SeriesID
		}
		if payload.BuiltAt != nil {
			updates["built_at"] = *payload.BuiltAt
		}
		if payload.Description != nil {
			updates["description"] = *payload.Description
		}
		if payload.CoverURL != "" {
			updates["cover"] = payload.CoverURL
		}

		if len(updates) > 0 {
			if err := tx.Model(&collectionEntity.Collection{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}

		if len(deletePictureIDs) > 0 {
			if err := tx.Where("collection_id = ? AND id IN ?", id, deletePictureIDs).Delete(&collectionEntity.Picture{}).Error; err != nil {
				return err
			}
		}

		if len(payload.NewPictureURLs) > 0 {
			newPictures := make([]collectionEntity.Picture, 0, len(payload.NewPictureURLs))
			for _, pictureURL := range payload.NewPictureURLs {
				if pictureURL == "" {
					continue
				}
				newPictures = append(newPictures, collectionEntity.Picture{
					CollectionID: id,
					Url:          pictureURL,
				})
			}

			if len(newPictures) > 0 {
				if err := tx.Create(&newPictures).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		if valErr, ok := err.(helper.ValError); ok {
			return collectionEntity.Collection{}, valErr
		}
		return collectionEntity.Collection{}, helper.DBError{ErrorMsg: err}
	}

	return collection, nil
}
