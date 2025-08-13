package entity

import (
	"time"
)

type CollectionDetailResponse struct {
	ID          int
	Title       string
	Scale       Scale
	RelaseType  RelaseType
	Status      COLLECTION_STATUS
	Series      Series
	BuiltAt     time.Time
	Cover       string
	Pictures    []Pictures
	Description string
}

type CollectionSearchResponse struct {
	ID          int
	Title       string
	Scale       Scale
	RelaseType  RelaseType
	Status      COLLECTION_STATUS
	Series      Series
	BuiltAt     time.Time
	Cover       string
	Description string
}
