package entity

import (
	"time"
)

type CollectionDetailResponse struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Type        CollectionType    `json:"collection_type"`
	RelaseType  RelaseType        `json:"release_type"`
	Status      COLLECTION_STATUS `json:"status"`
	Series      Series            `json:"series"`
	BuiltAt     time.Time         `json:"built_at"`
	Cover       string            `json:"cover"`
	Pictures    []Picture         `json:"pictures"`
	Description string            `json:"description"`
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
