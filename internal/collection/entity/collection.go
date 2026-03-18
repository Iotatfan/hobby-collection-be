package entity

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/helper"
)

// Table
type Collection struct {
	ID             int               `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GradeID        int               `gorm:"column:grade_id"`
	CollectionType CollectionType    `gorm:"-" json:"type"`
	Title          string            `gorm:"column:title" json:"title" binding:"required"`
	ReleaseTypeID  int               `gorm:"column:release_type"`
	ReleaseType    ReleaseType       `gorm:"foreignKey:ReleaseTypeID;default:0"  json:"release_type"`
	Status         COLLECTION_STATUS `gorm:"column:status" json:"status"`
	ManufacturerID int               `gorm:"column:manufacturer"`
	Manufacturer   Manufacturer      `gorm:"foreignKey:ManufacturerID;default:0"  json:"manufacturer"`
	SeriesID       int               `gorm:"column:series_id;default:0"`
	Series         Series            `gorm:"foreignKey:SeriesID" json:"series"`
	BuiltAt        *time.Time        `gorm:"column:built_at" json:"built_at"`
	Cover          string            `gorm:"column:cover" json:"cover"`
	Pictures       *[]Picture        `gorm:"foreignKey:CollectionID" json:"pictures"`
	Addons         *[]Addon          `gorm:"foreignKey:CollectionID" json:"addons"`
	Description    string
	helper.Model   `gorm:"embedded"`
}

type Addon struct {
	ID           int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AddonName    string `gorm:"column:addon_name" json:"addon_name" binding:"required"`
	CollectionID int    `gorm:"column:collection_id" json:"collection_id" binding:"required"`
	Picture      string `gorm:"column:picture" json:"picture" binding:"required"`
	helper.Model `gorm:"embedded"`
}

type CollectionType struct {
	ID                 int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CollectionTypeName string `gorm:"column:name" json:"name" binding:"required"`
	Scale              string `gorm:"-" json:"scale"`
	Grade              Grade  `gorm:"foreignKey:CollectionTypeID;references:ID" json:"grade"`

	helper.Model `gorm:"embedded"`
}

type Grade struct {
	ID               int    `gorm:"primaryKey;column:id" json:"id"`
	Name             string `gorm:"column:name" json:"name" binding:"required"`
	ShortName        string `gorm:"column:short_name" json:"short_name"`
	ScaleID          int    `gorm:"column:scale_id" json:"scale_id"`
	Scale            Scale  `gorm:"foreignKey:ScaleID" json:"scale_data"`
	CollectionTypeID int    `gorm:"column:collection_type_id" json:"collection_type_id"`
	helper.Model     `gorm:"embedded"`
}

type Scale struct {
	ID           int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name         string `gorm:"column:name" json:"name" binding:"required"`
	helper.Model `gorm:"embedded"`
}

type ReleaseType struct {
	ID              int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ReleaseTypeName string `gorm:"column:name" json:"name" binding:"required"`
	helper.Model    `gorm:"embedded"`
}

type Series struct {
	ID           int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SeriesName   string `gorm:"column:name" json:"name" binding:"required"`
	helper.Model `gorm:"embedded"`
}

type Picture struct {
	ID           int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CollectionID int    `gorm:"column:collection_id" json:"collection_id" binding:"required"`
	Url          string `gorm:"column:url" json:"url" binding:"required"`
	helper.Model `gorm:"embedded"`
}

type Manufacturer struct {
	ID               int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ManufacturerName string `gorm:"column:name" json:"name" binding:"required"`
	helper.Model     `gorm:"embedded"`
}

// Non Table

type CollectionList struct {
	Collections []Collection
}

type CollectionFilter struct {
	CollectionTypeID int               `form:"collection_type_id"`
	GradeID          int               `form:"grade_id"`
	ReleaseTypeID    int               `form:"release_type_id"`
	ManufacturerID   int               `form:"manufacturer_id"`
	SeriesID         int               `form:"series_id"`
	Status           COLLECTION_STATUS `form:"status"`
	Limit            int               `form:"limit"`
	Offset           int               `form:"offset"`
}

type COLLECTION_STATUS string

const (
	Whishlist = 0
	Backlog   = 1
	Owned     = 2
	Built     = 3
)
