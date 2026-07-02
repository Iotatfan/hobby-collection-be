package entity

import (
	"time"
)

type CollectionDetailResponse struct {
	ID            int                    `json:"id"`
	Title         string                 `json:"title"`
	Type          CollectionTypeResponse `json:"type"`
	ReleaseType   ReleaseTypeResponse    `json:"release_type"`
	Manufacturer  ManufacturerResponse   `json:"manufacturer"`
	Status        COLLECTION_STATUS      `json:"status"`
	Series        SeriesResponse         `json:"series"`
	BuiltAt       *time.Time             `json:"built_at"`
	AcquiredAt    *time.Time             `json:"acquired_at"`
	Cover         string                 `json:"cover"`
	Pictures      []string               `json:"pictures"`
	Addons        []AddonResponse        `json:"addons"`
	Modifications []MetadataTagResponse  `json:"modifications"`
	Features      []MetadataTagResponse  `json:"features"`
	Description   string                 `json:"description"`
	DisplaySize   DISPLAY_SIZE           `json:"display_size"`
}

type CollectionListResponse struct {
	Collections []CollectionListItemResponse `json:"collections"`
}

type CollectionListItemResponse struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Type        CollectionTypeResponse `json:"type"`
	ReleaseType ReleaseTypeResponse    `json:"release_type"`
	Status      COLLECTION_STATUS      `json:"status"`
	Series      SeriesResponse         `json:"series"`
	BuiltAt     *time.Time             `json:"built_at"`
	AcquiredAt  *time.Time             `json:"acquired_at"`
	Cover       string                 `json:"cover"`
}

type CollectionTypeResponse struct {
	ID                 int           `json:"id"`
	CollectionTypeName string        `json:"name"`
	Scale              string        `json:"scale"`
	Grade              GradeResponse `json:"grade"`
}

type CollectionDrawerResponse struct {
	Grades        []GradeDrawerItem           `json:"grades"`
	Scales        []ScaleResponse             `json:"scales"`
	ReleaseTypes  []ReleaseTypeFilterResponse `json:"release_types"`
	Manufacturers []ManufacturerResponse      `json:"manufacturers"`
	Series        []SeriesResponse            `json:"series"`
	Modifications []MetadataTagResponse       `json:"modifications"`
	Features      []MetadataTagResponse       `json:"features"`
}

type GradeDrawerItem struct {
	GradeID            int    `json:"grade_id"`
	CollectionTypeName string `json:"collection_type_name"`
	GradeShortName     string `json:"grade_short_name"`
}

type CollectionFilterResponse struct {
	CollectionTypes []CollectionTypeFilterResponse `json:"collection_types"`
	ReleaseTypes    []ReleaseTypeFilterResponse    `json:"release_types"`
	GunplaGrades    []GunplaGradeFilterResponse    `json:"gunpla_grades"`
	FiguresScales   []FiguresScaleFilterResponse   `json:"figures_scales"`
}

type CollectionTypeFilterResponse struct {
	ID                 int    `json:"id"`
	CollectionTypeName string `json:"name"`
}

type CollectionTypeDrawer struct {
	ID                 int             `json:"id"`
	CollectionTypeName string          `json:"name"`
	Grades             []GradeResponse `json:"grades"`
}

type GradeResponse struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	ShortName        string `json:"short_name"`
	CollectionTypeID int    `json:"collection_type_id"`
}

type GunplaGradeFilterResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type ScaleResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type FiguresScaleFilterResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ManufacturerResponse struct {
	ID               int    `json:"id"`
	ManufacturerName string `json:"name"`
}

type ReleaseTypeResponse struct {
	ID              int    `json:"id"`
	ReleaseTypeName string `json:"name"`
}

type ReleaseTypeFilterResponse struct {
	ID              int    `json:"id"`
	ReleaseTypeName string `json:"name"`
}

type SeriesResponse struct {
	ID         int    `json:"id"`
	SeriesName string `json:"name"`
}

type AddonResponse struct {
	ID           int                  `json:"id"`
	AddonName    string               `json:"name"`
	CollectionID int                  `json:"collection_id"`
	Manufacturer ManufacturerResponse `json:"manufacturer"`
}

type MetadataTagResponse struct {
	ID   int               `json:"id"`
	Slug string            `json:"slug"`
	Name string            `json:"name"`
	Type METADATA_TAG_TYPE `json:"type"`
}

type StatisticResponse struct {
	TotalCount     int `json:"total_count"`
	CompletedCount int `json:"completed_count"`
	BacklogCount   int `json:"backlog_count"`
	LimitedCount   int `json:"limited_count"` // limited/exclusive/p-bandai items count
}

type ShelfItemResponse struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Type        CollectionTypeResponse `json:"type"`
	Status      COLLECTION_STATUS      `json:"status"`
	Cover       string                 `json:"cover"`
	DisplaySize DISPLAY_SIZE           `json:"display_size"`
}

type CollectionShelfResponse struct {
	ID    int                 `json:"id"`
	Name  string              `json:"name"`
	Items []ShelfItemResponse `json:"items"`
	Count int                 `json:"count"`
}

type CollectionShelvesResponse struct {
	GunplaShelf        CollectionShelfResponse `json:"gunpla_shelf"`
	FigureShelf        CollectionShelfResponse `json:"figure_shelf"`
	OtherModelKitShelf CollectionShelfResponse `json:"other_model_kit_shelf"`
	BacklogShelf       CollectionShelfResponse `json:"backlog_shelf"`
}
