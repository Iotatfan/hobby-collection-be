package entity

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/common"
)

type ShortLink struct {
	ID          int    `gorm:"primaryKey;type:integer" json:"id"`
	OriginalURL string `gorm:"type:text" json:"original_url"`
	ShortCode   string `gorm:"type:varchar(10);uniqueIndex" json:"short_code"`
	// OwnerID      int    `gorm:"type:integer" json:"owner_id"` // Should be used for multi-user support in the future
	ExpiredAt    *time.Time `gorm:"type:timestamp" json:"expired_at"`
	common.Model `gorm:"embedded"`
}
