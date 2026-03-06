package entity

import (
	"mime/multipart"
	"time"
)

type UploadCollectionRequest struct {
	Title         string                  `form:"title" binding:"required"`
	TypeID        int                     `form:"type_id" binding:"required"`
	ReleaseTypeID int                     `form:"release_type_id" binding:"required"`
	Status        COLLECTION_STATUS       `form:"status"`
	SeriesID      int                     `form:"series_id"`
	BuiltAt       time.Time               `form:"built_at" time_format:"2006-01-02T15:04:05Z07:00"`
	Cover         *multipart.FileHeader   `form:"cover"`
	Pictures      []*multipart.FileHeader `form:"pictures"`
	Description   string                  `form:"description"`
	CoverURL      string                  `form:"-" json:"-"`
	PictureURLs   []string                `form:"-" json:"-"`
}
