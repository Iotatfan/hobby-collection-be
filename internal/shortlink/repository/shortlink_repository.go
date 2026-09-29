package repository

import (
	"context"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/shortlink/entity"
	"gorm.io/gorm"
)

type ShortLinkRepository interface {
	CreateShortLink(ctx context.Context, originalURL string, shortCode string, duration int, expiredAt *time.Time) error
	GetShortLinkByCode(ctx context.Context, shortCode string) (entity.ShortLink, error)
	CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error)
	GetShortLinkByUrl(ctx context.Context, url string, duration int) (entity.ShortLink, error)
	MarkLinkAsMalicious(ctx context.Context, shortCode string) error
}

type shortLinkRepository struct {
	db *gorm.DB
}

func NewShortLinkRepository(db *gorm.DB) ShortLinkRepository {
	return &shortLinkRepository{
		db: db,
	}
}

func (r *shortLinkRepository) CreateShortLink(ctx context.Context, originalURL string, shortCode string, duration int, expiredAt *time.Time) error {
	shortLink := entity.ShortLink{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		ExpiredAt:   expiredAt,
		Duration:    duration,
	}
	err := r.db.WithContext(ctx).Create(&shortLink).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *shortLinkRepository) GetShortLinkByCode(ctx context.Context, shortCode string) (entity.ShortLink, error) {
	var shortLink entity.ShortLink
	err := r.db.WithContext(ctx).Where("short_code = ? AND expired_at > ?", shortCode, time.Now()).First(&shortLink).Error
	if err != nil {
		return entity.ShortLink{}, err
	}

	if shortLink.IsMalicious {
		// Do not return original URL if it's marked malicious
		return entity.ShortLink{}, gorm.ErrRecordNotFound
	}

	return shortLink, nil
}

func (r *shortLinkRepository) CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ShortLink{}).Where("short_code = ?", shortCode).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *shortLinkRepository) GetShortLinkByUrl(ctx context.Context, url string, duration int) (entity.ShortLink, error) {
	var shortLink entity.ShortLink
	err := r.db.WithContext(ctx).Where("original_url = ? AND duration = ? AND expired_at > ?", url, duration, time.Now()).First(&shortLink).Error
	if err != nil {
		return entity.ShortLink{}, err
	}
	return shortLink, nil
}

func (r *shortLinkRepository) MarkLinkAsMalicious(ctx context.Context, shortCode string) error {
	return r.db.WithContext(ctx).Model(&entity.ShortLink{}).Where("short_code = ?", shortCode).Update("is_malicious", true).Error
}
