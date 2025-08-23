package helper

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime:mili" json:"-"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime:mili" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}
