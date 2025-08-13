package entity

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/helper"
)

type Collection struct {
	ID           int
	Title        string
	Scale        Scale
	RelaseType   RelaseType
	Status       COLLECTION_STATUS
	Series       Series
	BuiltAt      time.Time
	Cover        string
	Pictures     []Pictures
	Description  string
	helper.Model `gorm:"embedded"`
}

type CollectionType struct {
	ID                 int
	CollectionTypeName string
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
	ID    int
	Name  string
	Grade GUNPLA_GRADE
}

type COLLECTION_STATUS string

const (
	Built   = 0
	Backlog = 1
)

type RelaseType struct {
	ID              int
	ReleaseTypeName string
}

type Series struct {
	ID         int
	SeriesName string
}

type Pictures struct {
	ID           int
	CollectionID int
	Url          string
}
