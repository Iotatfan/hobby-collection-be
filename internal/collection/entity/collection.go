package entity

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/common"
)

// Table
type Collection struct {
	ID             int               `gorm:"primaryKey;autoIncrement;column:id"`
	GradeID        int               `gorm:"column:grade_id"`
	ScaleID        int               `gorm:"column:scale_id"`
	CollectionType CollectionType    `gorm:"-"`
	Title          string            `gorm:"column:title"`
	ReleaseTypeID  int               `gorm:"column:release_type"`
	ReleaseType    ReleaseType       `gorm:"foreignKey:ReleaseTypeID;default:0"`
	Status         COLLECTION_STATUS `gorm:"column:status"`
	ManufacturerID int               `gorm:"column:manufacturer"`
	Manufacturer   Manufacturer      `gorm:"foreignKey:ManufacturerID;default:0"`
	SeriesID       int               `gorm:"column:series_id;default:0"`
	Series         Series            `gorm:"foreignKey:SeriesID"`
	BuiltAt        *time.Time        `gorm:"column:built_at"`
	AcquiredAt     *time.Time        `gorm:"column:acquired_at"`
	Cover          string            `gorm:"column:cover"`
	Pictures       *[]Picture        `gorm:"foreignKey:CollectionID"`
	Addons         *[]Addon          `gorm:"foreignKey:CollectionID"`
	Description    string
	MetadataTags   []MetadataTags `gorm:"many2many:collection_metadata_tags;"`
	DisplaySize    DISPLAY_SIZE   `gorm:"column:display_size"`
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
	Scale              Scale  `gorm:"-"`
	Grade              Grade  `gorm:"foreignKey:CollectionTypeID;references:ID"`
	common.Model       `gorm:"embedded"`
}

type Grade struct {
	ID               int    `gorm:"primaryKey;column:id"`
	Name             string `gorm:"column:name"`
	ShortName        string `gorm:"column:short_name"`
	CollectionTypeID int    `gorm:"column:collection_type_id"`
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
	CollectionID int    `gorm:"column:collection_id;index"`
	Url          string `gorm:"column:url"`
	common.Model `gorm:"embedded"`
}

type Manufacturer struct {
	ID               int    `gorm:"primaryKey;autoIncrement;column:id"`
	ManufacturerName string `gorm:"column:name"`
	common.Model     `gorm:"embedded"`
}

type MetadataTags struct {
	ID           int               `gorm:"primaryKey;autoIncrement;column:id"`
	Slug         string            `gorm:"column:slug;uniqueIndex"`
	Name         string            `gorm:"column:name"`
	Type         METADATA_TAG_TYPE `gorm:"column:type"`
	common.Model `gorm:"embedded"`
}

// Non Table

type COLLECTION_STATUS string

const (
	Wishlist = 0
	Backlog  = 1
	Owned    = 2
	Built    = 3
)

type METADATA_TAG_TYPE int

const (
	Modification = 0
	Feature      = 1
)

type DISPLAY_SIZE string

const (
	SmallWide  = "small_wide" // for 1/144 gunpla
	SmallTall  = "small_tall"
	MediumWide = "medium_wide" // for 1/100 and big 1/144 gunpla
	MediumTall = "medium_tall"
	LargeWide  = "large_wide" // for 1/60 and big 1/100 gunpla
	LargeTall  = "large_tall"
)
