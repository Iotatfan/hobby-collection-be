package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	GetCollectionShelves(ctx context.Context) (collectionEntity.CollectionShelvesResponse, error)
	GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error)
	GetCollectionFilter() (collectionEntity.CollectionFilterResponse, error)
	GetPicturesByCollectionID(id int) ([]collectionEntity.Picture, error)
	GetAddonsByCollectionID(id int) ([]collectionEntity.Addon, error)
	GetMetadataTagsByIDs(ids []int) ([]collectionEntity.MetadataTags, error)
	GetCollectionStatistics() (collectionEntity.StatisticResponse, error)
	UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.Collection, error)
	UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest, deletePictureIDs []int, deleteAddonIDs []int) (collectionEntity.Collection, error)
}

type collectionRepository struct {
	db                   *gorm.DB
	filterCacheMu        sync.RWMutex
	filterDrawerCache    collectionEntity.CollectionFilterResponse
	filterDrawerCachedAt time.Time
	filterDrawerCacheTTL time.Duration
}

const (
	collectionTypeFiguresID = 1
	collectionTypeGunplaID  = 2
)

const (
	unknownScaleFiguresScaleID = 1
)

func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{
		db:                   db,
		filterDrawerCacheTTL: 5 * time.Minute,
	}
}

func (r *collectionRepository) GetCollectionByID(id int) (collectionEntity.Collection, error) {
	startedAt := time.Now()
	defer func() {
		log.Printf("[repo.collection.detail] collection_id=%d duration=%s", id, time.Since(startedAt))
	}()

	type collectionDetailRow struct {
		ID               int                                `gorm:"column:id"`
		Title            string                             `gorm:"column:title"`
		Status           collectionEntity.COLLECTION_STATUS `gorm:"column:status"`
		BuiltAt          *time.Time                         `gorm:"column:built_at"`
		AcquiredAt       *time.Time                         `gorm:"column:acquired_at"`
		Cover            string                             `gorm:"column:cover"`
		Description      string                             `gorm:"column:description"`
		DisplaySize      collectionEntity.DISPLAY_SIZE      `gorm:"column:display_size"`
		TypeID           int                                `gorm:"column:type_id"`
		TypeName         string                             `gorm:"column:type_name"`
		GradeID          int                                `gorm:"column:grade_id"`
		GradeName        string                             `gorm:"column:grade_name"`
		GradeShortName   string                             `gorm:"column:grade_short_name"`
		ScaleID          sql.NullInt64                      `gorm:"column:scale_id"`
		ScaleName        sql.NullString                     `gorm:"column:scale_name"`
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
	type metadataTagRow struct {
		ID   int    `gorm:"column:id"`
		Slug string `gorm:"column:slug"`
		Name string `gorm:"column:name"`
		Type int    `gorm:"column:type"`
	}

	row := collectionDetailRow{}
	mainQueryStartedAt := time.Now()
	result := r.db.Table("collections c").
		Select(`
			c.id,
			c.title,
			c.status,
			c.built_at,
			c.acquired_at,
			c.cover,
			c.description,
			c.display_size,
			ct.id as type_id,
			ct.name as type_name,
			g.id as grade_id,
			g.name as grade_name,
			g.short_name as grade_short_name,
			sc.id as scale_id,
			sc.name as scale_name,
			rt.id as release_type_id,
			rt.name as release_type_name,
			m.id as manufacturer_id,
			m.name as manufacturer_name,
			s.id as series_id,
			s.name as series_name
		`).
		Joins("JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
		Joins("JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
		Joins("LEFT JOIN scales sc ON sc.id = c.scale_id AND sc.deleted_at IS NULL").
		Joins("LEFT JOIN release_types rt ON rt.id = c.release_type AND rt.deleted_at IS NULL").
		Joins("LEFT JOIN manufacturers m ON m.id = c.manufacturer AND m.deleted_at IS NULL").
		Joins("LEFT JOIN series s ON s.id = c.series_id AND s.deleted_at IS NULL").
		Where("c.id = ? AND c.deleted_at IS NULL", id).
		Limit(1).
		Scan(&row)
	log.Printf("[repo.collection.detail.main_query] collection_id=%d duration=%s rows=%d err=%v", id, time.Since(mainQueryStartedAt), result.RowsAffected, result.Error)
	if result.Error != nil {
		return collectionEntity.Collection{}, common.DBError{ErrorMsg: result.Error}
	}
	if result.RowsAffected == 0 {
		return collectionEntity.Collection{}, common.DBError{ErrorMsg: gorm.ErrRecordNotFound}
	}

	pictures := []collectionEntity.Picture{}
	addons := []collectionEntity.Addon{}
	addonRows := []addonDetailRow{}
	metadataTagRows := []metadataTagRow{}
	var picturesErr error
	var addonsErr error
	var metadataTagsErr error
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		queryStartedAt := time.Now()
		picturesErr = r.db.Model(&collectionEntity.Picture{}).
			Select("id", "collection_id", "url").
			Where("collection_id = ? AND deleted_at IS NULL", id).
			Order("created_at DESC").
			Order("id DESC").
			Find(&pictures).Error
		log.Printf("[repo.collection.detail.pictures_query] collection_id=%d duration=%s rows=%d err=%v", id, time.Since(queryStartedAt), len(pictures), picturesErr)
	}()

	go func() {
		defer wg.Done()
		queryStartedAt := time.Now()
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
		log.Printf("[repo.collection.detail.addons_query] collection_id=%d duration=%s rows=%d err=%v", id, time.Since(queryStartedAt), len(addonRows), addonsErr)
	}()

	go func() {
		defer wg.Done()
		queryStartedAt := time.Now()
		metadataTagsErr = r.db.Table("metadata_tags mt").
			Select(`
				mt.id,
				mt.slug AS slug,
				mt.name,
				mt.type
			`).
			Joins("JOIN collection_metadata_tags cmt ON cmt.metadata_tags_id = mt.id").
			Where("cmt.collection_id = ? AND mt.deleted_at IS NULL", id).
			Order("mt.name ASC").
			Scan(&metadataTagRows).Error
		log.Printf("[repo.collection.detail.metadata_tags_query] collection_id=%d duration=%s rows=%d err=%v", id, time.Since(queryStartedAt), len(metadataTagRows), metadataTagsErr)
	}()

	wg.Wait()
	if picturesErr != nil || addonsErr != nil || metadataTagsErr != nil {
		return collectionEntity.Collection{}, common.DBError{
			ErrorMsg: errors.Join(
				wrapErr("load pictures", picturesErr),
				wrapErr("load addons", addonsErr),
				wrapErr("load metadata tags", metadataTagsErr),
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

	metadataTags := make([]collectionEntity.MetadataTags, 0, len(metadataTagRows))
	for _, metadataTagRow := range metadataTagRows {
		metadataTags = append(metadataTags, collectionEntity.MetadataTags{
			ID:   metadataTagRow.ID,
			Slug: metadataTagRow.Slug,
			Name: metadataTagRow.Name,
			Type: collectionEntity.METADATA_TAG_TYPE(metadataTagRow.Type),
		})
	}

	collection := collectionEntity.Collection{
		ID:             row.ID,
		GradeID:        row.GradeID,
		ScaleID:        nullInt(row.ScaleID),
		Title:          row.Title,
		ReleaseTypeID:  nullInt(row.ReleaseTypeID),
		Status:         row.Status,
		ManufacturerID: nullInt(row.ManufacturerID),
		SeriesID:       nullInt(row.SeriesID),
		BuiltAt:        row.BuiltAt,
		AcquiredAt:     row.AcquiredAt,
		Cover:          row.Cover,
		Description:    row.Description,
		DisplaySize:    row.DisplaySize,
		CollectionType: collectionEntity.CollectionType{
			ID:                 row.TypeID,
			CollectionTypeName: row.TypeName,
			Scale: collectionEntity.Scale{
				ID:   nullInt(row.ScaleID),
				Name: nullString(row.ScaleName),
			},
			Grade: collectionEntity.Grade{
				ID:               row.GradeID,
				Name:             row.GradeName,
				ShortName:        row.GradeShortName,
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
		Pictures:     &pictures,
		Addons:       &addons,
		MetadataTags: metadataTags,
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
	DisplaySize     collectionEntity.DISPLAY_SIZE      `gorm:"column:display_size"`
	TypeID          int                                `gorm:"column:type_id"`
	TypeName        string                             `gorm:"column:type_name"`
	GradeID         int                                `gorm:"column:grade_id"`
	GradeName       string                             `gorm:"column:grade_name"`
	GradeShortName  string                             `gorm:"column:grade_short_name"`
	ScaleID         sql.NullInt64                      `gorm:"column:scale_id"`
	ScaleName       sql.NullString                     `gorm:"column:scale_name"`
	ReleaseTypeID   sql.NullInt64                      `gorm:"column:release_type_id"`
	ReleaseTypeName sql.NullString                     `gorm:"column:release_type_name"`
	SeriesID        sql.NullInt64                      `gorm:"column:series_id"`
	SeriesName      sql.NullString                     `gorm:"column:series_name"`
	TotalCount      int                                `gorm:"column:total_count"`
}

func (r *collectionRepository) GetCollectionList(filters collectionEntity.CollectionFilterRequest) (collectionEntity.CollectionListResponse, error) {
	startedAt := time.Now()
	defer func() {
		log.Printf("[repo.collection.list] duration=%s filters={collection_type_id:%d grade_id:%d scale_id:%d release_type_ids:%v manufacturer_id:%d series_id:%d status:%q sort:%q limit:%d offset:%d}",
			time.Since(startedAt),
			filters.CollectionTypeID,
			filters.GradeID,
			filters.ScaleID,
			filters.ReleaseTypeIDs,
			filters.ManufacturerID,
			filters.SeriesID,
			filters.Status,
			filters.Sort,
			filters.Limit,
			filters.Offset,
		)
	}()

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
			g.id as grade_id,
			g.name as grade_name,
			g.short_name as grade_short_name,
			sc.id as scale_id,
			sc.name as scale_name,
			rt.id as release_type_id,
			rt.name as release_type_name,
			s.id as series_id,
			s.name as series_name
		`).
		Joins("JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
		Joins("JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
		Joins("LEFT JOIN scales sc ON sc.id = c.scale_id AND sc.deleted_at IS NULL").
		Joins("LEFT JOIN release_types rt ON rt.id = c.release_type AND rt.deleted_at IS NULL").
		Joins("LEFT JOIN series s ON s.id = c.series_id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL")

	if filters.CollectionTypeID > 0 {
		db = db.Where("g.collection_type_id = ?", filters.CollectionTypeID)
	}
	if filters.GradeID > 0 {
		db = db.Where("c.grade_id = ?", filters.GradeID)
	}
	if filters.ScaleID > 0 {
		db = db.Where("c.scale_id = ?", filters.ScaleID)
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

	queryStartedAt := time.Now()
	result := db.Order(orderBy1).
		Order(orderBy2).
		Limit(limit).
		Offset(offset).
		Scan(&rows)
	log.Printf("[repo.collection.list.query] duration=%s rows=%d order_by_1=%q order_by_2=%q err=%v", time.Since(queryStartedAt), len(rows), orderBy1, orderBy2, result.Error)
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
				Scale:              nullString(row.ScaleName),
				Grade: collectionEntity.GradeResponse{
					ID:               row.GradeID,
					Name:             row.GradeName,
					ShortName:        row.GradeShortName,
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

func (r *collectionRepository) GetCollectionShelves(ctx context.Context) (collectionEntity.CollectionShelvesResponse, error) {
	startedAt := time.Now()
	defer func() {
		log.Printf("[repo.collection.shelves] duration=%s", time.Since(startedAt))
	}()

	const shelfLimit = 6

	type shelfQuery struct {
		name  string
		id    int
		where string
		args  []any
	}

	queries := []shelfQuery{
		{
			name:  "Gunpla",
			id:    collectionTypeGunplaID,
			where: "ct.id = ? AND CAST(c.status AS int) <> ?",
			args:  []any{collectionTypeGunplaID, collectionEntity.Backlog},
		},
		{
			name:  "Figure",
			id:    collectionTypeFiguresID,
			where: "ct.id = ? AND CAST(c.status AS int) <> ?",
			args:  []any{collectionTypeFiguresID, collectionEntity.Backlog},
		},
		{
			name:  "Other Model Kit",
			id:    0,
			where: "ct.id NOT IN ? AND CAST(c.status AS int) <> ?",
			args:  []any{[]int{collectionTypeFiguresID, collectionTypeGunplaID}, collectionEntity.Backlog},
		},
		{
			name:  "Backlog",
			id:    collectionEntity.Backlog,
			where: "CAST(c.status AS int) = ?",
			args:  []any{collectionEntity.Backlog},
		},
	}

	shelves := make([]collectionEntity.CollectionShelfResponse, len(queries))
	var wg sync.WaitGroup
	var queryErrs []error
	var queryErrMu sync.Mutex

	for i, query := range queries {
		wg.Add(1)
		go func(index int, shelf shelfQuery) {
			defer wg.Done()

			rows := []collectionListItemRow{}
			queryStartedAt := time.Now()
			result := r.db.WithContext(ctx).
				Table("collections c").
				Select(`
					c.id,
					c.title,
					c.status,
					c.cover,
					c.display_size,
					ct.id as type_id,
					ct.name as type_name,
					g.id as grade_id,
					g.name as grade_name,
					g.short_name as grade_short_name,
					sc.id as scale_id,
					sc.name as scale_name,
					COUNT(*) OVER() AS total_count
				`).
				Joins("JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
				Joins("JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
				Joins("LEFT JOIN scales sc ON sc.id = c.scale_id AND sc.deleted_at IS NULL").
				Where("c.deleted_at IS NULL").
				Where(shelf.where, shelf.args...).
				Order("COALESCE(c.built_at, c.acquired_at, c.created_at) DESC NULLS LAST").
				Order("c.id DESC").
				Limit(shelfLimit).
				Scan(&rows)
			log.Printf("[repo.collection.shelves.query] shelf=%q duration=%s rows=%d err=%v", shelf.name, time.Since(queryStartedAt), len(rows), result.Error)
			if result.Error != nil {
				queryErrMu.Lock()
				queryErrs = append(queryErrs, wrapErr("load "+shelf.name+" shelf", result.Error))
				queryErrMu.Unlock()
				return
			}

			var total int
			if len(rows) > 0 {
				total = rows[0].TotalCount
			}

			shelves[index] = collectionEntity.CollectionShelfResponse{
				ID:    shelf.id,
				Name:  shelf.name,
				Items: mapShelfItems(rows),
				Count: total,
			}
		}(i, query)
	}

	wg.Wait()
	if len(queryErrs) > 0 {
		return collectionEntity.CollectionShelvesResponse{}, common.DBError{ErrorMsg: errors.Join(queryErrs...)}
	}

	return collectionEntity.CollectionShelvesResponse{
		GunplaShelf:        shelves[0],
		FigureShelf:        shelves[1],
		OtherModelKitShelf: shelves[2],
		BacklogShelf:       shelves[3],
	}, nil
}

func mapShelfItems(rows []collectionListItemRow) []collectionEntity.ShelfItemResponse {
	items := make([]collectionEntity.ShelfItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, collectionEntity.ShelfItemResponse{
			ID:    row.ID,
			Title: row.Title,
			Type: collectionEntity.CollectionTypeResponse{
				ID:                 row.TypeID,
				CollectionTypeName: row.TypeName,
				Scale:              nullString(row.ScaleName),
				Grade: collectionEntity.GradeResponse{
					ID:               row.GradeID,
					Name:             row.GradeName,
					ShortName:        row.GradeShortName,
					CollectionTypeID: row.TypeID,
				},
			},
			Status:      row.Status,
			Cover:       row.Cover,
			DisplaySize: row.DisplaySize,
		})
	}

	return items
}

func getCollectionListSort(sort string) (string, string) {
	sort = strings.TrimSpace(strings.ToLower(sort))

	switch sort {
	case "latest", "latest_built":
		return "COALESCE(c.built_at, c.acquired_at) DESC NULLS LAST", "c.id DESC"
	case "name", "name_asc":
		return "c.title ASC", "c.id ASC"
	case "name_desc":
		return "c.title DESC", "c.id DESC"
	default:
		return "COALESCE(c.built_at, c.acquired_at) DESC NULLS LAST", "c.id DESC"
	}
}

func (r *collectionRepository) GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error) {
	drawer := collectionEntity.CollectionDrawerResponse{}
	type gradeRow struct {
		ID                 int            `gorm:"column:id"`
		ShortName          string         `gorm:"column:short_name"`
		CollectionTypeName sql.NullString `gorm:"column:collection_type_name"`
	}

	grades := []gradeRow{}
	if err := r.db.Table("grades g").
		Select(`
			g.id,
			g.short_name,
			ct.name as collection_type_name
		`).
		Joins("LEFT JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
		Where("g.deleted_at IS NULL").
		Order("ct.name ASC, g.short_name ASC").
		Find(&grades).Error; err != nil {
		return collectionEntity.CollectionDrawerResponse{}, common.DBError{ErrorMsg: err}
	}

	drawer.Grades = make([]collectionEntity.GradeDrawerItem, 0, len(grades))
	for _, grade := range grades {
		drawer.Grades = append(drawer.Grades, collectionEntity.GradeDrawerItem{
			GradeID:            grade.ID,
			CollectionTypeName: nullString(grade.CollectionTypeName),
			GradeShortName:     grade.ShortName,
		})
	}

	scales := []collectionEntity.Scale{}
	releaseTypes := []collectionEntity.ReleaseType{}
	manufacturers := []collectionEntity.Manufacturer{}
	series := []collectionEntity.Series{}
	metadataTags := []collectionEntity.MetadataTags{}

	var scalesErr error
	var releaseTypesErr error
	var manufacturersErr error
	var seriesErr error
	var metadataTagsErr error
	var drawerWG sync.WaitGroup
	drawerWG.Add(5)

	go func() {
		defer drawerWG.Done()
		scalesErr = r.db.Model(&collectionEntity.Scale{}).
			Order("name ASC").
			Find(&scales).Error
	}()

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

	go func() {
		defer drawerWG.Done()
		metadataTagsErr = r.db.Model(&collectionEntity.MetadataTags{}).
			Order("type ASC, name ASC").
			Find(&metadataTags).Error
	}()

	drawerWG.Wait()
	if scalesErr != nil || releaseTypesErr != nil || manufacturersErr != nil || seriesErr != nil || metadataTagsErr != nil {
		return collectionEntity.CollectionDrawerResponse{}, common.DBError{
			ErrorMsg: errors.Join(
				wrapErr("load scales", scalesErr),
				wrapErr("load release types", releaseTypesErr),
				wrapErr("load manufacturers", manufacturersErr),
				wrapErr("load series", seriesErr),
				wrapErr("load metadata tags", metadataTagsErr),
			),
		}
	}

	drawer.Scales = make([]collectionEntity.ScaleResponse, 0, len(scales))
	for _, scale := range scales {
		drawer.Scales = append(drawer.Scales, collectionEntity.ScaleResponse{
			ID:   scale.ID,
			Name: scale.Name,
		})
	}

	drawer.ReleaseTypes = make([]collectionEntity.ReleaseTypeFilterResponse, 0, len(releaseTypes))
	for _, releaseType := range releaseTypes {
		drawer.ReleaseTypes = append(drawer.ReleaseTypes, collectionEntity.ReleaseTypeFilterResponse{
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

	drawer.Modifications = make([]collectionEntity.MetadataTagResponse, 0)
	drawer.Features = make([]collectionEntity.MetadataTagResponse, 0)
	for _, metadataTag := range metadataTags {
		resp := collectionEntity.MetadataTagResponse{
			ID:   metadataTag.ID,
			Slug: metadataTag.Slug,
			Name: metadataTag.Name,
			Type: metadataTag.Type,
		}

		switch metadataTag.Type {
		case collectionEntity.Feature:
			drawer.Features = append(drawer.Features, resp)
		default:
			drawer.Modifications = append(drawer.Modifications, resp)
		}
	}

	return drawer, nil
}

func (r *collectionRepository) GetCollectionFilter() (collectionEntity.CollectionFilterResponse, error) {
	if cached, ok := r.getCachedCollectionFilter(); ok {
		log.Printf("[repo.collection.filter] cache=hit duration=0s")
		return cached, nil
	}

	startedAt := time.Now()
	drawer := collectionEntity.CollectionFilterResponse{}

	type collectionTypeRow struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	rows := []collectionTypeRow{}
	collectionTypesStartedAt := time.Now()
	if err := r.db.Table("collection_types").
		Select("id", "name").
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&rows).Error; err != nil {
		log.Printf("[repo.collection.filter.collection_types_query] duration=%s rows=%d err=%v", time.Since(collectionTypesStartedAt), len(rows), err)
		return collectionEntity.CollectionFilterResponse{}, common.DBError{ErrorMsg: err}
	}
	log.Printf("[repo.collection.filter.collection_types_query] duration=%s rows=%d err=<nil>", time.Since(collectionTypesStartedAt), len(rows))

	releaseTypes := []collectionEntity.ReleaseType{}
	releaseTypesStartedAt := time.Now()
	if err := r.db.Model(&collectionEntity.ReleaseType{}).
		Order("name ASC").
		Find(&releaseTypes).Error; err != nil {
		log.Printf("[repo.collection.filter.release_types_query] duration=%s rows=%d err=%v", time.Since(releaseTypesStartedAt), len(releaseTypes), err)
		return collectionEntity.CollectionFilterResponse{}, common.DBError{ErrorMsg: err}
	}
	log.Printf("[repo.collection.filter.release_types_query] duration=%s rows=%d err=<nil>", time.Since(releaseTypesStartedAt), len(releaseTypes))

	gunplaGrades := []collectionEntity.Grade{}
	if err := r.db.Model(&collectionEntity.Grade{}).
		Joins("JOIN collection_types ct ON ct.id = grades.collection_type_id AND ct.deleted_at IS NULL").
		Where("grades.deleted_at IS NULL").
		Where("ct.id = ?", collectionTypeGunplaID).
		Order("grades.short_name ASC").
		Find(&gunplaGrades).Error; err != nil {
		return collectionEntity.CollectionFilterResponse{}, common.DBError{ErrorMsg: err}
	}

	figuresScales := []collectionEntity.Scale{}
	if err := r.db.Table("scales s").
		Select("DISTINCT s.id, s.name").
		Joins("JOIN collections c ON c.scale_id = s.id AND c.deleted_at IS NULL").
		Joins("JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
		Joins("JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
		Where("s.deleted_at IS NULL").
		Where("ct.id = ?", collectionTypeFiguresID).
		Where("s.id <> ?", unknownScaleFiguresScaleID).
		Order("s.name ASC").
		Find(&figuresScales).Error; err != nil {
		return collectionEntity.CollectionFilterResponse{}, common.DBError{ErrorMsg: err}
	}

	drawer.CollectionTypes = make([]collectionEntity.CollectionTypeFilterResponse, 0, len(rows))
	for _, row := range rows {
		drawer.CollectionTypes = append(drawer.CollectionTypes, collectionEntity.CollectionTypeFilterResponse{
			ID:                 row.ID,
			CollectionTypeName: row.Name,
		})
	}

	drawer.ReleaseTypes = make([]collectionEntity.ReleaseTypeFilterResponse, 0, len(releaseTypes))
	for _, releaseType := range releaseTypes {
		drawer.ReleaseTypes = append(drawer.ReleaseTypes, collectionEntity.ReleaseTypeFilterResponse{
			ID:              releaseType.ID,
			ReleaseTypeName: releaseType.ReleaseTypeName,
		})
	}

	drawer.GunplaGrades = make([]collectionEntity.GunplaGradeFilterResponse, 0, len(gunplaGrades))
	for _, grade := range gunplaGrades {
		drawer.GunplaGrades = append(drawer.GunplaGrades, collectionEntity.GunplaGradeFilterResponse{
			ID:        grade.ID,
			Name:      grade.Name,
			ShortName: grade.ShortName,
		})
	}

	drawer.FiguresScales = make([]collectionEntity.FiguresScaleFilterResponse, 0, len(figuresScales))
	for _, scale := range figuresScales {
		drawer.FiguresScales = append(drawer.FiguresScales, collectionEntity.FiguresScaleFilterResponse{
			ID:   scale.ID,
			Name: scale.Name,
		})
	}

	r.setCachedCollectionFilterDrawer(drawer)
	log.Printf("[repo.collection.filter] cache=miss duration=%s", time.Since(startedAt))

	return drawer, nil
}

func (r *collectionRepository) getCachedCollectionFilter() (collectionEntity.CollectionFilterResponse, bool) {
	r.filterCacheMu.RLock()
	defer r.filterCacheMu.RUnlock()

	if r.filterDrawerCachedAt.IsZero() || time.Since(r.filterDrawerCachedAt) > r.filterDrawerCacheTTL {
		return collectionEntity.CollectionFilterResponse{}, false
	}

	return cloneCollectionFilterResponse(r.filterDrawerCache), true
}

func (r *collectionRepository) setCachedCollectionFilterDrawer(drawer collectionEntity.CollectionFilterResponse) {
	r.filterCacheMu.Lock()
	defer r.filterCacheMu.Unlock()

	r.filterDrawerCache = cloneCollectionFilterResponse(drawer)
	r.filterDrawerCachedAt = time.Now()
}

func cloneCollectionFilterResponse(drawer collectionEntity.CollectionFilterResponse) collectionEntity.CollectionFilterResponse {
	cloned := collectionEntity.CollectionFilterResponse{
		CollectionTypes: make([]collectionEntity.CollectionTypeFilterResponse, len(drawer.CollectionTypes)),
		ReleaseTypes:    make([]collectionEntity.ReleaseTypeFilterResponse, len(drawer.ReleaseTypes)),
		GunplaGrades:    make([]collectionEntity.GunplaGradeFilterResponse, len(drawer.GunplaGrades)),
		FiguresScales:   make([]collectionEntity.FiguresScaleFilterResponse, len(drawer.FiguresScales)),
	}

	copy(cloned.CollectionTypes, drawer.CollectionTypes)
	copy(cloned.ReleaseTypes, drawer.ReleaseTypes)
	copy(cloned.GunplaGrades, drawer.GunplaGrades)
	copy(cloned.FiguresScales, drawer.FiguresScales)

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

func (r *collectionRepository) GetCollectionStatistics() (collectionEntity.StatisticResponse, error) {
	var statistic collectionEntity.StatisticResponse

	err := r.db.Model(&collectionEntity.Collection{}).
		Select("COUNT(*) as total_count, COUNT(*) FILTER (WHERE CAST(status AS int) IN (?, ?)) AS completed_count, COUNT(*) FILTER (WHERE CAST(status AS int) = ?) AS backlog_count, COUNT(*) FILTER (WHERE release_type IN (?, ?, ?)) AS limited_count",
			collectionEntity.Built, collectionEntity.Owned, collectionEntity.Backlog, 2, 3, 4).
		Where("deleted_at IS NULL").Scan(&statistic).Error
	if err != nil {
		return collectionEntity.StatisticResponse{}, common.DBError{ErrorMsg: err}
	}

	return statistic, nil
}

func (r *collectionRepository) UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.Collection, error) {
	collection := collectionEntity.Collection{
		GradeID:        payload.GradeID,
		ScaleID:        payload.ScaleID,
		Title:          payload.Title,
		ReleaseTypeID:  payload.ReleaseTypeID,
		ManufacturerID: payload.ManufacturerID,
		Status:         payload.Status,
		SeriesID:       payload.SeriesID,
		Cover:          payload.CoverURL,
		Description:    payload.Description,
		DisplaySize:    payload.DisplaySize,
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

	metadataTags, err := r.GetMetadataTagsByIDs(payload.MetadataTagIDs)
	if err != nil {
		return collectionEntity.Collection{}, err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
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

		if len(metadataTags) > 0 {
			if err := tx.Model(&collection).Association("MetadataTags").Replace(metadataTags); err != nil {
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

	metadataTags := make([]collectionEntity.MetadataTags, 0)
	if payload.MetadataTagIDsPresent {
		var err error
		metadataTags, err = r.GetMetadataTagsByIDs(payload.MetadataTagIDs)
		if err != nil {
			return collectionEntity.Collection{}, err
		}
	}

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
		if payload.ScaleID != nil {
			updates["scale_id"] = *payload.ScaleID
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
		if payload.DisplaySize != nil {
			updates["display_size"] = *payload.DisplaySize
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

		if payload.MetadataTagIDsPresent {
			if err := tx.Model(&collection).Association("MetadataTags").Replace(metadataTags); err != nil {
				return err
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

func (r *collectionRepository) GetMetadataTagsByIDs(ids []int) ([]collectionEntity.MetadataTags, error) {
	if len(ids) == 0 {
		return []collectionEntity.MetadataTags{}, nil
	}

	metadataTags := make([]collectionEntity.MetadataTags, 0, len(ids))
	if err := r.db.Where("id IN ?", ids).Find(&metadataTags).Error; err != nil {
		return nil, common.DBError{ErrorMsg: err}
	}

	if len(metadataTags) != len(ids) {
		return nil, common.ValError{ErrorMsg: errors.New("one or more metadata_tags do not exist")}
	}

	return metadataTags, nil
}
