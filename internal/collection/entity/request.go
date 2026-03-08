package entity

import (
	"mime/multipart"
	"time"
)

type UploadCollectionRequest struct {
	Title          string                  `form:"title" binding:"required"`
	TypeID         int                     `form:"type_id" binding:"required"`
	ReleaseTypeID  int                     `form:"release_type_id" binding:"required"`
	ManufacturerID int                     `form:"manufacturer_id" binding:"required"`
	Status         COLLECTION_STATUS       `form:"status"`
	SeriesID       int                     `form:"series_id"`
	BuiltAt        time.Time               `form:"built_at" time_format:"2006-01-02T15:04:05Z07:00"`
	Cover          *multipart.FileHeader   `form:"cover"`
	Pictures       []*multipart.FileHeader `form:"pictures"`
	Description    string                  `form:"description"`
	CoverURL       string                  `form:"-" json:"-"`
	PictureURLs    []string                `form:"-" json:"-"`
}

type UpdateCollectionRequest struct {
	Title                     *string                 `form:"title"`
	TypeID                    *int                    `form:"type_id"`
	ReleaseTypeID             *int                    `form:"release_type_id"`
	ManufacturerID            *int                    `form:"manufacturer_id"`
	Status                    *COLLECTION_STATUS      `form:"status"`
	SeriesID                  *int                    `form:"series_id"`
	BuiltAt                   *time.Time              `form:"built_at" time_format:"2006-01-02T15:04:05Z07:00"`
	Description               *string                 `form:"description"`
	Cover                     *multipart.FileHeader   `form:"cover"`
	NewPictures               []*multipart.FileHeader `form:"new_pictures"`
	ExistingPictureIDs        []int                   `form:"existing_picture_ids"`
	ExistingPictureIDsPresent bool                    `form:"-" json:"-"`
	CoverURL                  string                  `form:"-" json:"-"`
	NewPictureURLs            []string                `form:"-" json:"-"`
}
