package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"strings"

	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/common"

	"gorm.io/gorm"
)

type CollectionRepository interface {
	GetCollectionByID(id int) (collectionEntity.Collection, error)
	GetCollectionList(filters collectionEntity.CollectionFilterRequest) (collectionEntity.CollectionListResponse, error)
	GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error)
	GetCollectionFilterDrawer() (collectionEntity.CollectionFilterDrawerResponse, error)
	GetPicturesByCollectionID(id int) ([]collectionEntity.Picture, error)
	GetAddonsByCollectionID(id int) ([]collectionEntity.Addon, error)
	UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.Collection, error)
	UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest, deletePictureIDs []int, deleteAddonIDs []int) (collectionEntity.Collection, error)
}

type collectionRepository struct {
	db                   *gorm.DB
	filterDrawerCacheMu  sync.RWMutex
	filterDrawerCache    collectionEntity.CollectionFilterDrawerResponse
	filterDrawerCachedAt time.Time
	filterDrawerCacheTTL time.Duration
}

func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{
		db:                   db,
		filterDrawerCacheTTL: 5 * time.Minute,
	}
}

func (r *collectionRepository) GetCollectionByID(id int) (collectionEntity.Collection, error) {
	type collectionDetailRow struct {
		ID               int                                `gorm:"column:id"`
		Title            string                             `gorm:"column:title"`
		Status           collectionEntity.COLLECTION_STATUS `gorm:"column:status"`
		BuiltAt          *time.Time                         `gorm:"column:built_at"`
		AcquiredAt       *time.Time                         `gorm:"column:acquired_at"`
		Cover            string                             `gorm:"column:cover"`
		Description      string                             `gorm:"column:description"`
		TypeID           int                                `gorm:"column:type_id"`
		TypeName         string                             `gorm:"column:type_name"`
		TypeScale        sql.NullString                     `gorm:"column:type_scale"`
		GradeID          int                                `gorm:"column:grade_id"`
		GradeScaleID     int                                `gorm:"column:grade_scale_id"`
		GradeName        string                             `gorm:"column:grade_name"`
		GradeShortName   string                             `gorm:"column:grade_short_name"`
		ReleaseTypeID    sql.NullInt64                      `gorm:"column:release_type_id"`
		ReleaseTypeName  sql.NullString                     `gorm:"column:release_type_name"`
		ManufacturerID   sql.NullInt64                      `gorm:"column:manufacturer_id"`
		ManufacturerName sql.NullString                     `gorm:"column:manufacturer_name"`
		SeriesID         sql.NullInt64                      `gorm:"column:series_id"`
		SeriesName       sql.NullString                     `gorm:"column:series_name"`
	}
	type addonDetailRow struct {
		ID               int            `gorm:"column:id"`
		AddonName        string         `gorm:"column:addon_name"`
		CollectionID     int            `gorm:"column:collection_id"`
		ManufacturerID   int            `gorm:"column:manufacturer"`
		ManufacturerName sql.NullString `gorm:"column:manufacturer_name"`
	}

	row := collectionDetailRow{}
	result := r.db.Table("collections c").
		Select(`
			c.id,
			c.title,
			c.status,
			c.built_at,
			c.acquired_at,
			c.cover,
			c.description,
			ct.id as type_id,
			ct.name as type_name,
			sc.name as type_scale,
			g.id as grade_id,
			g.scale_id as grade_scale_id,
			g.name as grade_name,
			g.short_name as grade_short_name,
			rt.id as release_type_id,
			rt.name as release_type_name,
			m.id as manufacturer_id,
			m.name as manufacturer_name,
			s.id as series_id,
			s.name as series_name
		`).
		Joins("JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
		Joins("JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
		Joins("LEFT JOIN scales sc ON sc.id = g.scale_id AND sc.deleted_at IS NULL").
		Joins("LEFT JOIN release_types rt ON rt.id = c.release_type AND rt.deleted_at IS NULL").
		Joins("LEFT JOIN manufacturers m ON m.id = c.manufacturer AND m.deleted_at IS NULL").
		Joins("LEFT JOIN series s ON s.id = c.series_id AND s.deleted_at IS NULL").
		Where("c.id = ? AND c.deleted_at IS NULL", id).
		Limit(1).
		Scan(&row)
	if result.Error != nil {
		return collectionEntity.Collection{}, common.DBError{ErrorMsg: result.Error}
	}
	if result.RowsAffected == 0 {
		return collectionEntity.Collection{}, common.DBError{ErrorMsg: gorm.ErrRecordNotFound}
	}

	pictures := []collectionEntity.Picture{}
	addons := []collectionEntity.Addon{}
	addonRows := []addonDetailRow{}
	var picturesErr error
	var addonsErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		picturesErr = r.db.Model(&collectionEntity.Picture{}).
			Select("id", "collection_id", "url").
			Where("collection_id = ? AND deleted_at IS NULL", id).
			Order("created_at DESC").
			Order("id DESC").
			Find(&pictures).Error
	}()

	go func() {
		defer wg.Done()
		addonsErr = r.db.Table("addons a").
			Select(`
				a.id,
				a.addon_name,
				a.collection_id,
				a.manufacturer,
				m.name AS manufacturer_name
			`).
			Joins("LEFT JOIN manufacturers m ON m.id = a.manufacturer AND m.deleted_at IS NULL").
			Where("a.collection_id = ? AND a.deleted_at IS NULL", id).
			Order("a.created_at DESC").
			Order("a.id DESC").
			Scan(&addonRows).Error
	}()

	wg.Wait()
	if picturesErr != nil || addonsErr != nil {
		return collectionEntity.Collection{}, common.DBError{
			ErrorMsg: errors.Join(
				wrapErr("load pictures", picturesErr),
				wrapErr("load addons", addonsErr),
			),
		}
	}
	addons = make([]collectionEntity.Addon, 0, len(addonRows))
	for _, addonRow := range addonRows {
		addons = append(addons, collectionEntity.Addon{
			ID:             addonRow.ID,
			AddonName:      addonRow.AddonName,
			ManufacturerID: addonRow.ManufacturerID,
			CollectionID:   addonRow.CollectionID,
			Manufacturer: collectionEntity.Manufacturer{
				ID:               addonRow.ManufacturerID,
				ManufacturerName: nullString(addonRow.ManufacturerName),
			},
		})
	}

	collection := collectionEntity.Collection{
		ID:             row.ID,
		GradeID:        row.GradeID,
		Title:          row.Title,
		ReleaseTypeID:  nullInt(row.ReleaseTypeID),
		Status:         row.Status,
		ManufacturerID: nullInt(row.ManufacturerID),
		SeriesID:       nullInt(row.SeriesID),
		BuiltAt:        row.BuiltAt,
		AcquiredAt:     row.AcquiredAt,
		Cover:          row.Cover,
		Description:    row.Description,
		CollectionType: collectionEntity.CollectionType{
			ID:                 row.TypeID,
			CollectionTypeName: row.TypeName,
			Scale:              nullString(row.TypeScale),
			Grade: collectionEntity.Grade{
				ID:               row.GradeID,
				Name:             row.GradeName,
				ShortName:        row.GradeShortName,
				ScaleID:          row.GradeScaleID,
				CollectionTypeID: row.TypeID,
			},
		},
		ReleaseType: collectionEntity.ReleaseType{
			ID:              nullInt(row.ReleaseTypeID),
			ReleaseTypeName: nullString(row.ReleaseTypeName),
		},
		Manufacturer: collectionEntity.Manufacturer{
			ID:               nullInt(row.ManufacturerID),
			ManufacturerName: nullString(row.ManufacturerName),
		},
		Series: collectionEntity.Series{
			ID:         nullInt(row.SeriesID),
			SeriesName: nullString(row.SeriesName),
		},
		Pictures: &pictures,
		Addons:   &addons,
	}

	return collection, nil
}

type collectionListItemRow struct {
	ID              int                                `gorm:"column:id"`
	Title           string                             `gorm:"column:title"`
	Status          collectionEntity.COLLECTION_STATUS `gorm:"column:status"`
	BuiltAt         *time.Time                         `gorm:"column:built_at"`
	AcquiredAt      *time.Time                         `gorm:"column:acquired_at"`
	Cover           string                             `gorm:"column:cover"`
	TypeID          int                                `gorm:"column:type_id"`
	TypeName        string                             `gorm:"column:type_name"`
	TypeScale       sql.NullString                     `gorm:"column:type_scale"`
	GradeID         int                                `gorm:"column:grade_id"`
	GradeScaleID    int                                `gorm:"column:grade_scale_id"`
	GradeName       string                             `gorm:"column:grade_name"`
	GradeShortName  string                             `gorm:"column:grade_short_name"`
	ReleaseTypeID   sql.NullInt64                      `gorm:"column:release_type_id"`
	ReleaseTypeName sql.NullString                     `gorm:"column:release_type_name"`
	SeriesID        sql.NullInt64                      `gorm:"column:series_id"`
	SeriesName      sql.NullString                     `gorm:"column:series_name"`
}

func (r *collectionRepository) GetCollectionList(filters collectionEntity.CollectionFilterRequest) (collectionEntity.CollectionListResponse, error) {
	rows := []collectionListItemRow{}
	db := r.db.Table("collections c").
		Select(`
			c.id,
			c.title,
			c.status,
			c.built_at,
			c.acquired_at,
			c.cover,
			ct.id as type_id,
			ct.name as type_name,
			sc.name as type_scale,
			g.id as grade_id,
			g.scale_id as grade_scale_id,
			g.name as grade_name,
			g.short_name as grade_short_name,
			rt.id as release_type_id,
			rt.name as release_type_name,
			s.id as series_id,
			s.name as series_name
		`).
		Joins("JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
		Joins("JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
		Joins("LEFT JOIN scales sc ON sc.id = g.scale_id AND sc.deleted_at IS NULL").
		Joins("LEFT JOIN release_types rt ON rt.id = c.release_type AND rt.deleted_at IS NULL").
		Joins("LEFT JOIN series s ON s.id = c.series_id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL")

	if filters.CollectionTypeID > 0 {
		db = db.Where("g.collection_type_id = ?", filters.CollectionTypeID)
	}
	if filters.GradeID > 0 {
		db = db.Where("c.grade_id = ?", filters.GradeID)
	}
	if len(filters.ReleaseTypeIDs) > 0 {
		db = db.Where("c.release_type IN ?", filters.ReleaseTypeIDs)
	}
	if filters.ManufacturerID > 0 {
		db = db.Where("c.manufacturer = ?", filters.ManufacturerID)
	}
	if filters.SeriesID > 0 {
		db = db.Where("c.series_id = ?", filters.SeriesID)
	}
	if filters.Status != "" {
		db = db.Where("c.status = ?", filters.Status)
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

	orderBy1, orderBy2 := getCollectionListSort(filters.Sort)

	result := db.Order(orderBy1).
		Order(orderBy2).
		Limit(limit).
		Offset(offset).
		Scan(&rows)
	if result.Error != nil {
		return collectionEntity.CollectionListResponse{}, common.DBError{ErrorMsg: result.Error}
	}
	if len(rows) == 0 {
		return collectionEntity.CollectionListResponse{
			Collections: []collectionEntity.CollectionListItemResponse{},
		}, nil
	}

	response := collectionEntity.CollectionListResponse{
		Collections: make([]collectionEntity.CollectionListItemResponse, 0, len(rows)),
	}
	for _, row := range rows {
		var builtAt *time.Time
		if row.BuiltAt != nil {
			localBuiltAt := row.BuiltAt.Local()
			builtAt = &localBuiltAt
		}
		var acquiredAt *time.Time
		if row.AcquiredAt != nil {
			localAcquiredAt := row.AcquiredAt.Local()
			acquiredAt = &localAcquiredAt
		}

		response.Collections = append(response.Collections, collectionEntity.CollectionListItemResponse{
			ID:    row.ID,
			Title: row.Title,
			Type: collectionEntity.CollectionTypeResponse{
				ID:                 row.TypeID,
				CollectionTypeName: row.TypeName,
				Scale:              nullString(row.TypeScale),
				Grade: collectionEntity.GradeResponse{
					ID:               row.GradeID,
					Name:             row.GradeName,
					ShortName:        row.GradeShortName,
					ScaleID:          row.GradeScaleID,
					CollectionTypeID: row.TypeID,
				},
			},
			ReleaseType: collectionEntity.ReleaseTypeResponse{
				ID:              nullInt(row.ReleaseTypeID),
				ReleaseTypeName: nullString(row.ReleaseTypeName),
			},
			Status: row.Status,
			Series: collectionEntity.SeriesResponse{
				ID:         nullInt(row.SeriesID),
				SeriesName: nullString(row.SeriesName),
			},
			BuiltAt:    builtAt,
			AcquiredAt: acquiredAt,
			Cover:      row.Cover,
		})
	}

	return response, nil
}

func getCollectionListSort(sort string) (string, string) {
	sort = strings.TrimSpace(strings.ToLower(sort))

	switch sort {
	case "latest", "latest_built":
		return "c.built_at DESC NULLS LAST", "c.acquired_at DESC NULLS LAST, c.id DESC"
	case "name", "name_asc":
		return "c.title ASC", "c.id ASC"
	case "name_desc":
		return "c.title DESC", "c.id DESC"
	default:
		return "c.built_at DESC NULLS LAST", "c.acquired_at DESC NULLS LAST, c.id DESC"
	}
}

func (r *collectionRepository) GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error) {
	drawer := collectionEntity.CollectionDrawerResponse{}
	type gradeRow struct {
		ID                 int            `gorm:"column:id"`
		ShortName          string         `gorm:"column:short_name"`
		ScaleName          sql.NullString `gorm:"column:scale_name"`
		CollectionTypeName sql.NullString `gorm:"column:collection_type_name"`
	}

	grades := []gradeRow{}
	if err := r.db.Table("grades g").
		Select(`
			g.id,
			g.short_name,
			s.name as scale_name,
			ct.name as collection_type_name
		`).
		Joins("LEFT JOIN scales s ON s.id = g.scale_id AND s.deleted_at IS NULL").
		Joins("LEFT JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
		Where("g.deleted_at IS NULL").
		Order("ct.name ASC, g.short_name ASC, s.name ASC").
		Find(&grades).Error; err != nil {
		return collectionEntity.CollectionDrawerResponse{}, common.DBError{ErrorMsg: err}
	}

	drawer.Grades = make([]collectionEntity.GradeDrawerItem, 0, len(grades))
	for _, grade := range grades {
		drawer.Grades = append(drawer.Grades, collectionEntity.GradeDrawerItem{
			GradeID:            grade.ID,
			CollectionTypeName: nullString(grade.CollectionTypeName),
			GradeShortName:     grade.ShortName,
			Scale:              nullString(grade.ScaleName),
		})
	}

	releaseTypes := []collectionEntity.ReleaseType{}
	manufacturers := []collectionEntity.Manufacturer{}
	series := []collectionEntity.Series{}

	var releaseTypesErr error
	var manufacturersErr error
	var seriesErr error
	var drawerWG sync.WaitGroup
	drawerWG.Add(3)

	go func() {
		defer drawerWG.Done()
		releaseTypesErr = r.db.Model(&collectionEntity.ReleaseType{}).
			Order("name ASC").
			Find(&releaseTypes).Error
	}()

	go func() {
		defer drawerWG.Done()
		manufacturersErr = r.db.Model(&collectionEntity.Manufacturer{}).
			Order("name ASC").
			Find(&manufacturers).Error
	}()

	go func() {
		defer drawerWG.Done()
		seriesErr = r.db.Model(&collectionEntity.Series{}).
			Order("name ASC").
			Find(&series).Error
	}()

	drawerWG.Wait()
	if releaseTypesErr != nil || manufacturersErr != nil || seriesErr != nil {
		return collectionEntity.CollectionDrawerResponse{}, common.DBError{
			ErrorMsg: errors.Join(
				wrapErr("load release types", releaseTypesErr),
				wrapErr("load manufacturers", manufacturersErr),
				wrapErr("load series", seriesErr),
			),
		}
	}

	drawer.ReleaseTypes = make([]collectionEntity.ReleaseTypeResponse, 0, len(releaseTypes))
	for _, releaseType := range releaseTypes {
		drawer.ReleaseTypes = append(drawer.ReleaseTypes, collectionEntity.ReleaseTypeResponse{
			ID:              releaseType.ID,
			ReleaseTypeName: releaseType.ReleaseTypeName,
		})
	}

	drawer.Manufacturers = make([]collectionEntity.ManufacturerResponse, 0, len(manufacturers))
	for _, manufacturer := range manufacturers {
		drawer.Manufacturers = append(drawer.Manufacturers, collectionEntity.ManufacturerResponse{
			ID:               manufacturer.ID,
			ManufacturerName: manufacturer.ManufacturerName,
		})
	}

	drawer.Series = make([]collectionEntity.SeriesResponse, 0, len(series))
	for _, item := range series {
		drawer.Series = append(drawer.Series, collectionEntity.SeriesResponse{
			ID:         item.ID,
			SeriesName: item.SeriesName,
		})
	}

	return drawer, nil
}

func (r *collectionRepository) GetCollectionFilterDrawer() (collectionEntity.CollectionFilterDrawerResponse, error) {
	if cached, ok := r.getCachedCollectionFilterDrawer(); ok {
		return cached, nil
	}

	drawer := collectionEntity.CollectionFilterDrawerResponse{}

	type collectionTypeRow struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	rows := []collectionTypeRow{}
	if err := r.db.Table("collection_types").
		Select("id", "name").
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&rows).Error; err != nil {
		return collectionEntity.CollectionFilterDrawerResponse{}, common.DBError{ErrorMsg: err}
	}

	releaseTypes := []collectionEntity.ReleaseType{}
	if err := r.db.Model(&collectionEntity.ReleaseType{}).
		Order("name ASC").
		Find(&releaseTypes).Error; err != nil {
		return collectionEntity.CollectionFilterDrawerResponse{}, common.DBError{ErrorMsg: err}
	}

	drawer.CollectionTypes = make([]collectionEntity.CollectionTypeFilterItem, 0, len(rows))
	for _, row := range rows {
		drawer.CollectionTypes = append(drawer.CollectionTypes, collectionEntity.CollectionTypeFilterItem{
			ID:                 row.ID,
			CollectionTypeName: row.Name,
		})
	}

	drawer.ReleaseTypes = make([]collectionEntity.ReleaseTypeResponse, 0, len(releaseTypes))
	for _, releaseType := range releaseTypes {
		drawer.ReleaseTypes = append(drawer.ReleaseTypes, collectionEntity.ReleaseTypeResponse{
			ID:              releaseType.ID,
			ReleaseTypeName: releaseType.ReleaseTypeName,
		})
	}

	r.setCachedCollectionFilterDrawer(drawer)

	return drawer, nil
}

func (r *collectionRepository) getCachedCollectionFilterDrawer() (collectionEntity.CollectionFilterDrawerResponse, bool) {
	r.filterDrawerCacheMu.RLock()
	defer r.filterDrawerCacheMu.RUnlock()

	if r.filterDrawerCachedAt.IsZero() || time.Since(r.filterDrawerCachedAt) > r.filterDrawerCacheTTL {
		return collectionEntity.CollectionFilterDrawerResponse{}, false
	}

	return cloneCollectionFilterDrawerResponse(r.filterDrawerCache), true
}

func (r *collectionRepository) setCachedCollectionFilterDrawer(drawer collectionEntity.CollectionFilterDrawerResponse) {
	r.filterDrawerCacheMu.Lock()
	defer r.filterDrawerCacheMu.Unlock()

	r.filterDrawerCache = cloneCollectionFilterDrawerResponse(drawer)
	r.filterDrawerCachedAt = time.Now()
}

func cloneCollectionFilterDrawerResponse(drawer collectionEntity.CollectionFilterDrawerResponse) collectionEntity.CollectionFilterDrawerResponse {
	cloned := collectionEntity.CollectionFilterDrawerResponse{
		CollectionTypes: make([]collectionEntity.CollectionTypeFilterItem, len(drawer.CollectionTypes)),
		ReleaseTypes:    make([]collectionEntity.ReleaseTypeResponse, len(drawer.ReleaseTypes)),
	}

	copy(cloned.CollectionTypes, drawer.CollectionTypes)
	copy(cloned.ReleaseTypes, drawer.ReleaseTypes)

	return cloned
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func nullInt(value sql.NullInt64) int {
	if !value.Valid {
		return 0
	}

	return int(value.Int64)
}

func (r *collectionRepository) GetPicturesByCollectionID(id int) ([]collectionEntity.Picture, error) {
	pictures := []collectionEntity.Picture{}
	err := r.db.Model(&collectionEntity.Picture{}).
		Select("id", "collection_id", "url").
		Where("collection_id = ? AND deleted_at IS NULL", id).
		Find(&pictures).Error
	if err != nil {
		return []collectionEntity.Picture{}, common.DBError{ErrorMsg: err}
	}
	return pictures, nil
}

func (r *collectionRepository) GetAddonsByCollectionID(id int) ([]collectionEntity.Addon, error) {
	addons := []collectionEntity.Addon{}
	err := r.db.Model(&collectionEntity.Addon{}).
		Select("id", "addon_name", "collection_id", "manufacturer", "created_at", "updated_at").
		Where("collection_id = ? AND deleted_at IS NULL", id).
		Find(&addons).Error
	if err != nil {
		return []collectionEntity.Addon{}, common.DBError{ErrorMsg: err}
	}
	return addons, nil
}

func (r *collectionRepository) UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.Collection, error) {
	collection := collectionEntity.Collection{
		GradeID:        payload.GradeID,
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
	if !payload.AcquiredAt.IsZero() {
		acquiredAt := payload.AcquiredAt
		collection.AcquiredAt = &acquiredAt
	}

	pictures := make([]collectionEntity.Picture, 0, len(payload.PictureURLs))
	for _, pictureURL := range payload.PictureURLs {
		if pictureURL == "" {
			continue
		}
		pictures = append(pictures, collectionEntity.Picture{Url: pictureURL})
	}

	addons := make([]collectionEntity.Addon, 0, len(payload.AddonNames))
	if len(payload.AddonNames) > 0 {
		for i := range payload.AddonNames {
			name := strings.TrimSpace(payload.AddonNames[i])
			if name == "" {
				continue
			}
			manufacturerID := 0
			if len(payload.AddonManufacturerID) == len(payload.AddonNames) {
				manufacturerID = payload.AddonManufacturerID[i]
			}
			addons = append(addons, collectionEntity.Addon{
				AddonName:      name,
				ManufacturerID: manufacturerID,
			})
		}
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&collection).Error; err != nil {
			return err
		}

		if len(pictures) > 0 {
			for i := range pictures {
				pictures[i].CollectionID = collection.ID
			}
			if err := tx.Create(&pictures).Error; err != nil {
				return err
			}
		}

		if len(addons) > 0 {
			for i := range addons {
				addons[i].CollectionID = collection.ID
			}
			if err := tx.Create(&addons).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return collectionEntity.Collection{}, common.DBError{ErrorMsg: err}
	}

	return collection, nil
}

func wrapErr(scope string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", scope, err)
}

func (r *collectionRepository) UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest, deletePictureIDs []int, deleteAddonIDs []int) (collectionEntity.Collection, error) {
	collection := collectionEntity.Collection{}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&collection, id).Error; err != nil {
			return err
		}

		updates := map[string]any{}
		if payload.Title != nil {
			updates["title"] = *payload.Title
		}
		if payload.GradeID != nil {
			updates["grade_id"] = *payload.GradeID
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
		if payload.AcquiredAt != nil {
			updates["acquired_at"] = *payload.AcquiredAt
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

		if len(deleteAddonIDs) > 0 {
			if err := tx.Where("collection_id = ? AND id IN ?", id, deleteAddonIDs).Delete(&collectionEntity.Addon{}).Error; err != nil {
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

		if len(payload.UpdateAddonIDs) > 0 {
			for i := range payload.UpdateAddonIDs {
				addonID := payload.UpdateAddonIDs[i]
				if addonID <= 0 {
					continue
				}

				update := map[string]any{}
				if len(payload.UpdateAddonNames) == len(payload.UpdateAddonIDs) {
					name := strings.TrimSpace(payload.UpdateAddonNames[i])
					if name != "" {
						update["addon_name"] = name
					}
				}
				if len(payload.UpdateAddonManufacturerID) == len(payload.UpdateAddonIDs) {
					update["manufacturer"] = payload.UpdateAddonManufacturerID[i]
				}
				if len(update) == 0 {
					continue
				}

				if err := tx.Model(&collectionEntity.Addon{}).
					Where("collection_id = ? AND id = ?", id, addonID).
					Updates(update).Error; err != nil {
					return err
				}
			}
		}

		if len(payload.NewAddonNames) > 0 {
			newAddons := make([]collectionEntity.Addon, 0, len(payload.NewAddonNames))
			for i := range payload.NewAddonNames {
				name := strings.TrimSpace(payload.NewAddonNames[i])
				if name == "" {
					continue
				}
				manufacturerID := 0
				if len(payload.NewAddonManufacturerID) == len(payload.NewAddonNames) {
					manufacturerID = payload.NewAddonManufacturerID[i]
				}
				newAddons = append(newAddons, collectionEntity.Addon{
					CollectionID:   id,
					AddonName:      name,
					ManufacturerID: manufacturerID,
				})
			}
			if len(newAddons) > 0 {
				if err := tx.Create(&newAddons).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		if valErr, ok := err.(common.ValError); ok {
			return collectionEntity.Collection{}, valErr
		}
		return collectionEntity.Collection{}, common.DBError{ErrorMsg: err}
	}

	return collection, nil
}
