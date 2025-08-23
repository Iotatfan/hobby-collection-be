package entity

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/helper"
)

type Collection struct {
	ID           int               `gorm:"primaryKey;column:id" json:"id"`
	TypeID       int               `gorm:"column:type_id"`
	Type         CollectionType    `gorm:"foreignKey:TypeID" json:"collection_type"`
	Title        string            `gorm:"column:title" json:"title" binding:"required"`
	RelaseTypeID int               `gorm:"column:release_type"`
	RelaseType   RelaseType        `gorm:"foreignKey:RelaseTypeID"  json:"release_type"`
	Status       COLLECTION_STATUS `gorm:"column:status" json:"status"`
	SeriesID     int               `gorm:"column:series_id"`
	Series       Series            `gorm:"foreignKey:SeriesID" json:"series"`
	BuiltAt      time.Time         `gorm:"column:built_at" json:"built_at"`
	Cover        string            `gorm:"column:cover" json:"cover"`
	Pictures     []Picture         `gorm:"foreignKey:CollectionID" json:"pcitures"`
	Description  string
	helper.Model `gorm:"embedded"`
}

type CollectionList struct {
	Collections []Collection
}

type CollectionFilter struct {
	CollectionTypeID int
	GradeID          int
}

type CollectionType struct {
	ID                 int    `gorm:"primaryKey;column:id" json:"id"`
	CollectionTypeName string `gorm:"column:name" json:"name" binding:"required"`
	Scale              string `gorm:"column:scale" json:"scale" binding:"required"`
	GradeID            int    `gorm:"column:grade_id"`
	Grade              Grade  `gorm:"foreignKey:GradeID" json:"grade"`

	helper.Model `gorm:"embedded"`
}

type Grade struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	Name         string `gorm:"column:name" json:"name" binding:"required"`
	ShortName    string `gorm:"column:short_name" json:"short_name"`
	helper.Model `gorm:"embedded"`
}

type COLLECTION_STATUS string

const (
	Whishlist = 0
	Backlog   = 1
	Owned     = 2
	Built     = 3
)

type RelaseType struct {
	ID              int    `gorm:"primaryKey;column:id" json:"id"`
	ReleaseTypeName string `gorm:"column:name" json:"name" binding:"required"`
	helper.Model    `gorm:"embedded"`
}

type Series struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	SeriesName   string `gorm:"column:name" json:"name" binding:"required"`
	helper.Model `gorm:"embedded"`
}

type Picture struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	CollectionID int    `gorm:"column:collection_id" json:"collection_id" binding:"required"`
	Url          string `gorm:"column:url" json:"url" binding:"required"`
	helper.Model `gorm:"embedded"`
}
