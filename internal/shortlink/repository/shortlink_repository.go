package repository

import (
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/shortlink/entity"
	"gorm.io/gorm"
)

type ShortLinkRepository interface {
	CreateShortLink(originalURL string, shortCode string, expiredAt *time.Time, isMalicious bool) error
	GetShortLinkByCode(shortCode string) (string, error)
	CheckShortCodeExists(shortCode string) (bool, error)
	GetShortLinkByUrl(url string) (entity.ShortLink, error)
	MarkLinkAsMalicious(shortCode string) error
}

type shortLinkRepository struct {
	db *gorm.DB
}

func NewShortLinkRepository(db *gorm.DB) ShortLinkRepository {
	return &shortLinkRepository{
		db: db,
	}
}

func (r *shortLinkRepository) CreateShortLink(originalURL string, shortCode string, expiredAt *time.Time, isMalicious bool) error {
	shortLink := entity.ShortLink{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		ExpiredAt:   expiredAt,
		IsMalicious: isMalicious,
	}
	err := r.db.Create(&shortLink).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *shortLinkRepository) GetShortLinkByCode(shortCode string) (string, error) {
	var shortLink entity.ShortLink
	err := r.db.Where("short_code = ? AND expired_at > ?", shortCode, time.Now()).First(&shortLink).Error
	if err != nil {
		return "", err
	}

	if shortLink.IsMalicious {
		// Do not return original URL if it's marked malicious
		return "", gorm.ErrRecordNotFound
	}

	return shortLink.OriginalURL, nil
}

func (r *shortLinkRepository) CheckShortCodeExists(shortCode string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.ShortLink{}).Where("short_code = ?", shortCode).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *shortLinkRepository) GetShortLinkByUrl(url string) (entity.ShortLink, error) {
	var shortLink entity.ShortLink
	err := r.db.Where("original_url = ? AND expired_at > ?", url, time.Now()).First(&shortLink).Error
	if err != nil {
		return entity.ShortLink{}, err
	}
	return shortLink, nil
}

func (r *shortLinkRepository) MarkLinkAsMalicious(shortCode string) error {
	return r.db.Model(&entity.ShortLink{}).Where("short_code = ?", shortCode).Update("is_malicious", true).Error
}
