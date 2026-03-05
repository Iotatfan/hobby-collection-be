package entity

import "time"

type UploadCollectionRequest struct {
	ID          int               `json:"id"`
	Title       string            `json:"title" binding:"required"`
	Type        CollectionType    `json:"type" binding:"required"`
	ReleaseType ReleaseType       `json:"release_type" binding:"required"`
	Status      COLLECTION_STATUS `json:"status"`
	Series      Series            `json:"series"`
	BuiltAt     time.Time         `json:"built_at"`
	Cover       string            `json:"cover"`
	Pictures    []Picture         `json:"pictures"`
	Description string            `json:"description"`
}
