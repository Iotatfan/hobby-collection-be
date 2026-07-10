package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/shortlink/entity"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/repository"
	"github.com/iotatfan/hobby-collection-be/pkg/cache"
	"gorm.io/gorm"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const maxRetries = 5
const cacheKeyPrefix = "shortlink:"
const cacheTTL = 24 * time.Hour

type ShortLinkService interface {
	CreateShortLink(ctx context.Context, originalURL string, expiredAt *time.Time) (entity.CreateShortLinkResponse, error)
	GetShortLinkByCode(ctx context.Context, shortCode string) (string, error)
}

type shortLinkService struct {
	repo  repository.ShortLinkRepository
	redis *cache.RedisCache
}

func NewShortLinkService(repo repository.ShortLinkRepository, redis *cache.RedisCache) ShortLinkService {
	return &shortLinkService{
		repo:  repo,
		redis: redis,
	}
}

func (s *shortLinkService) CreateShortLink(ctx context.Context, originalURL string, expiredAt *time.Time) (entity.CreateShortLinkResponse, error) {
	existingLink, err := s.repo.GetShortLinkByUrl(originalURL)
	if err == nil {
		return entity.CreateShortLinkResponse{ShortCode: existingLink.ShortCode, ExpiredAt: existingLink.ExpiredAt}, nil
	}

	for i := 0; i < maxRetries; i++ {
		shortCode, err := generateShortCode(7)
		if err != nil {
			return entity.CreateShortLinkResponse{}, err
		}

		err = s.repo.CreateShortLink(originalURL, shortCode, expiredAt)
		if err != nil {
			if isUniqueContraintViolationDetected(err) {
				continue
			}
			return entity.CreateShortLinkResponse{}, fmt.Errorf("failed to create short link after %d retries", maxRetries)
		}

		err = s.redis.Set(ctx, cacheKeyPrefix+shortCode, originalURL, cacheTTL)
		if err != nil {
			return entity.CreateShortLinkResponse{}, err
		}
		return entity.CreateShortLinkResponse{ShortCode: shortCode, ExpiredAt: expiredAt}, nil
	}

	return entity.CreateShortLinkResponse{}, fmt.Errorf("failed to create short link after %d retries", maxRetries)
}

func (s *shortLinkService) GetShortLinkByCode(ctx context.Context, shortCode string) (string, error) {
	var resp string
	err := s.redis.Get(ctx, cacheKeyPrefix+shortCode, &resp)
	if err == nil {
		return resp, nil
	}

	url, err := s.repo.GetShortLinkByCode(shortCode)
	if err != nil {
		return "", err
	}

	_ = s.redis.Set(ctx, cacheKeyPrefix+shortCode, url, cacheTTL)
	return url, nil
}

func generateShortCode(length int) (string, error) {
	code := make([]byte, length)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[num.Int64()]
	}
	return string(code), nil
}

func isUniqueContraintViolationDetected(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
