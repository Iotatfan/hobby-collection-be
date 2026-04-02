package entity

import (
	"mime/multipart"
	"time"
)

type UploadCollectionRequest struct {
	Title               string                  `form:"title" binding:"required"`
	GradeID             int                     `form:"grade_id" binding:"required"`
	ReleaseTypeID       int                     `form:"release_type_id" binding:"required"`
	ManufacturerID      int                     `form:"manufacturer_id" binding:"required"`
	Status              COLLECTION_STATUS       `form:"status"`
	SeriesID            int                     `form:"series_id"`
	BuiltAt             time.Time               `form:"built_at" time_format:"2006-01-02T15:04:05Z07:00"`
	AcquiredAt          time.Time               `form:"acquired_at" time_format:"2006-01-02T15:04:05Z07:00"`
	Cover               *multipart.FileHeader   `form:"cover"`
	Pictures            []*multipart.FileHeader `form:"pictures"`
	AddonNames          []string                `form:"addon_names"`
	AddonManufacturerID []int                   `form:"addons_manufacturer_id"`
	Description         string                  `form:"description"`
	CoverURL            string                  `form:"-" json:"-"`
	PictureURLs         []string                `form:"-" json:"-"`
}

type UpdateCollectionRequest struct {
	Title                     *string                 `form:"title"`
	GradeID                   *int                    `form:"grade_id"`
	ReleaseTypeID             *int                    `form:"release_type_id"`
	ManufacturerID            *int                    `form:"manufacturer_id"`
	Status                    *COLLECTION_STATUS      `form:"status"`
	SeriesID                  *int                    `form:"series_id"`
	BuiltAt                   *time.Time              `form:"built_at" time_format:"2006-01-02T15:04:05Z07:00"`
	AcquiredAt                *time.Time              `form:"acquired_at" time_format:"2006-01-02T15:04:05Z07:00"`
	Description               *string                 `form:"description"`
	Cover                     *multipart.FileHeader   `form:"cover"`
	NewPictures               []*multipart.FileHeader `form:"new_pictures"`
	DeletedPictureURLs        []string                `form:"deleted_picture_urls"`
	DeletedPictureURLsPresent bool                    `form:"-" json:"-"`
	NewAddonNames             []string                `form:"new_addon_names"`
	NewAddonManufacturerID    []int                   `form:"new_addons_manufacturer_id"`
	ExistingAddonIDs          []int                   `form:"existing_addon_ids"`
	ExistingAddonIDsPresent   bool                    `form:"-" json:"-"`
	DeletedAddonIDs           []int                   `form:"deleted_addon_ids"`
	DeletedAddonIDsPresent    bool                    `form:"-" json:"-"`
	UpdateAddonIDs            []int                   `form:"update_addon_ids"`
	UpdateAddonNames          []string                `form:"update_addon_names"`
	UpdateAddonManufacturerID []int                   `form:"update_addons_manufacturer_id"`
	CoverURL                  string                  `form:"-" json:"-"`
	NewPictureURLs            []string                `form:"-" json:"-"`
}

type CollectionFilterRequest struct {
	CollectionTypeID int               `form:"collection_type_id"`
	GradeID          int               `form:"grade_id"`
	ReleaseTypeIDs   []int             `form:"release_type_id"`
	ManufacturerID   int               `form:"manufacturer_id"`
	SeriesID         int               `form:"series_id"`
	Status           COLLECTION_STATUS `form:"status"`
	Sort             string            `form:"sort"`
	Limit            int               `form:"limit"`
	Offset           int               `form:"offset"`
}
