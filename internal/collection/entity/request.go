package entity

import "time"

type UploadCollectionRequest struct {
	ID          int
	Title       string
	Scale       Scale
	RelaseType  RelaseType
	Status      COLLECTION_STATUS
	Series      Series
	BuiltAt     time.Time
	Cover       string
	Pictures    []Picture
	Description string
}
