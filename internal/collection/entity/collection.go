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

type CollectionType struct {
	ID                 int    `gorm:"primaryKey;column:id" json:"id"`
	CollectionTypeName string `gorm:"column:name" json:"name"`
	Scale              Scale  `gorm:"foreignKey:ScaleID" json:"scale"`
	helper.Model       `gorm:"embedded"`
}

type GUNPLA_GRADE string

const (
	EG = 0
	HG = 1
	RG = 2
	MG = 3
	PG = 4
)

type Scale struct {
	ID           int          `gorm:"primaryKey;column:id" json:"id"`
	Name         string       `gorm:"column:name" json:"name" binding:"required"`
	Grade        GUNPLA_GRADE `json:"grade"`
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
