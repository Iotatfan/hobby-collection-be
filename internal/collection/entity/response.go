package entity

import (
	"time"
)

type CollectionDetailResponse struct {
	ID           int                    `json:"id"`
	Title        string                 `json:"title"`
	Type         CollectionTypeResponse `json:"type"`
	ReleaseType  ReleaseType            `json:"release_type"`
	Manufacturer Manufacturer           `json:"manufacturer"`
	Status       COLLECTION_STATUS      `json:"status"`
	Series       Series                 `json:"series"`
	BuiltAt      time.Time              `json:"built_at"`
	Cover        string                 `json:"cover"`
	Pictures     []string               `json:"pictures"`
	Description  string                 `json:"description"`
}

type CollectionListResponse struct {
	Collections []CollectionListItemResponse `json:"collections"`
}

type CollectionListItemResponse struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Type        CollectionTypeResponse `json:"type"`
	ReleaseType ReleaseType            `json:"release_type"`
	Status      COLLECTION_STATUS      `json:"status"`
	Series      Series                 `json:"series"`
	BuiltAt     time.Time              `json:"built_at"`
	Cover       string                 `json:"cover"`
}

type CollectionTypeResponse struct {
	ID                 int    `json:"id"`
	CollectionTypeName string `json:"name"`
	Scale              string `json:"scale"`
	Grade              Grade  `json:"grade"`
}

type CollectionDrawerResponse struct {
	CollectionTypes []CollectionTypeDrawer `json:"collection_types"`
	ReleaseTypes    []ReleaseType          `json:"release_types"`
	Manufacturers   []Manufacturer         `json:"manufacturers"`
	Series          []Series               `json:"series"`
}

type CollectionTypeDrawer struct {
	ID                 int     `json:"id"`
	CollectionTypeName string  `json:"name"`
	Grades             []Grade `json:"grades"`
}
