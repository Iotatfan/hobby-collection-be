package entity

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/common"
)

// Table
type Collection struct {
	ID             int               `gorm:"primaryKey;autoIncrement;column:id"`
	GradeID        int               `gorm:"column:grade_id;index:idx_collections_grade_deleted,composite:grade_id"`
	CollectionType CollectionType    `gorm:"-"`
	Title          string            `gorm:"column:title"`
	ReleaseTypeID  int               `gorm:"column:release_type;index:idx_collections_release_type_deleted,composite:release_type"`
	ReleaseType    ReleaseType       `gorm:"foreignKey:ReleaseTypeID;default:0"`
	Status         COLLECTION_STATUS `gorm:"column:status"`
	ManufacturerID int               `gorm:"column:manufacturer;index:idx_collections_manufacturer_deleted,composite:manufacturer"`
	Manufacturer   Manufacturer      `gorm:"foreignKey:ManufacturerID;default:0"`
	SeriesID       int               `gorm:"column:series_id;default:0;index:idx_collections_series_deleted,composite:series_id"`
	Series         Series            `gorm:"foreignKey:SeriesID"`
	BuiltAt        *time.Time        `gorm:"column:built_at"`
	AcquiredAt     *time.Time        `gorm:"column:acquired_at"`
	Cover          string            `gorm:"column:cover"`
	Pictures       *[]Picture        `gorm:"foreignKey:CollectionID"`
	Addons         *[]Addon          `gorm:"foreignKey:CollectionID"`
	Description    string
	common.Model   `gorm:"embedded"`
}

type Addon struct {
	ID             int          `gorm:"primaryKey;autoIncrement;column:id"`
	AddonName      string       `gorm:"column:addon_name"`
	ManufacturerID int          `gorm:"column:manufacturer"`
	Manufacturer   Manufacturer `gorm:"foreignKey:ManufacturerID;default:0"`
	CollectionID   int          `gorm:"column:collection_id;index"`
	common.Model   `gorm:"embedded"`
}

type CollectionType struct {
	ID                 int    `gorm:"primaryKey;autoIncrement;column:id"`
	CollectionTypeName string `gorm:"column:name"`
	Scale              string `gorm:"-"`
	Grade              Grade  `gorm:"foreignKey:CollectionTypeID;references:ID"`
	common.Model       `gorm:"embedded"`
}

type Grade struct {
	ID               int    `gorm:"primaryKey;column:id"`
	Name             string `gorm:"column:name"`
	ShortName        string `gorm:"column:short_name"`
	ScaleID          int    `gorm:"column:scale_id"`
	Scale            Scale  `gorm:"foreignKey:ScaleID"`
	CollectionTypeID int    `gorm:"column:collection_type_id;index:idx_grades_collection_type_deleted,composite:collection_type_id"`
	common.Model     `gorm:"embedded"`
}

type Scale struct {
	ID           int    `gorm:"primaryKey;autoIncrement;column:id"`
	Name         string `gorm:"column:name"`
	common.Model `gorm:"embedded"`
}

type ReleaseType struct {
	ID              int    `gorm:"primaryKey;autoIncrement;column:id"`
	ReleaseTypeName string `gorm:"column:name"`
	common.Model    `gorm:"embedded"`
}

type Series struct {
	ID           int    `gorm:"primaryKey;autoIncrement;column:id"`
	SeriesName   string `gorm:"column:name"`
	common.Model `gorm:"embedded"`
}

type Picture struct {
	ID           int    `gorm:"primaryKey;autoIncrement;column:id"`
	CollectionID int    `gorm:"column:collection_id"`
	Url          string `gorm:"column:url"`
	common.Model `gorm:"embedded"`
}

type Manufacturer struct {
	ID               int    `gorm:"primaryKey;autoIncrement;column:id"`
	ManufacturerName string `gorm:"column:name"`
	common.Model     `gorm:"embedded"`
}

// Non Table

type COLLECTION_STATUS string

const (
	Whishlist = 0
	Backlog   = 1
	Owned     = 2
	Built     = 3
)
