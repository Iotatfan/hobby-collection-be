package entity

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/common"
)

type ShortLink struct {
	ID          int    `gorm:"primaryKey;type:integer"`
	OriginalURL string `gorm:"type:text"`
	ShortCode   string `gorm:"type:varchar(10);uniqueIndex"`
	// OwnerID      int    `gorm:"type:integer" json:"owner_id"` // Should be used for multi-user support in the future
	Duration     int        `gorm:"type:int"`
	ExpiredAt    *time.Time `gorm:"type:timestamp"`
	IsMalicious  bool       `gorm:"type:boolean"`
	common.Model `gorm:"embedded"`
}
