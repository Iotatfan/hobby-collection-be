package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/common"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/entity"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/repository"
	"gorm.io/gorm"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const maxRetries = 5
const cacheKeyPrefix = "shortlink:"
const maxExpiryDays = 365

type cachedShortLink struct {
	OriginalURL string     `json:"original_url"`
	ExpiredAt   *time.Time `json:"expired_at"`
}

type shortLinkCache interface {
	Get(context.Context, string, any) error
	Set(context.Context, string, any, time.Duration) error
	Delete(context.Context, ...string) error
}

type ShortLinkService interface {
	CreateShortLink(ctx context.Context, originalURL string, duration int, customURL string) (entity.CreateShortLinkResponse, error)
	GetShortLinkByCode(ctx context.Context, shortCode string) (string, error)
}

type shortLinkService struct {
	repo           repository.ShortLinkRepository
	redis          shortLinkCache
	threatDetector ThreatDetector
}

func NewShortLinkService(repo repository.ShortLinkRepository, redis shortLinkCache, threatDetector ThreatDetector) ShortLinkService {
	return &shortLinkService{
		repo:           repo,
		redis:          redis,
		threatDetector: threatDetector,
	}
}

func (s *shortLinkService) CreateShortLink(ctx context.Context, originalURL string, duration int, customURL string) (entity.CreateShortLinkResponse, error) {
	if err := validateDestination(originalURL); err != nil {
		return entity.CreateShortLinkResponse{}, err
	}
	if duration < 1 || duration > maxExpiryDays {
		return entity.CreateShortLinkResponse{}, common.ValError{ErrorMsg: errors.New("expiry_days must be between 1 and 365")}
	}
	if len(customURL) > 10 {
		return entity.CreateShortLinkResponse{}, common.ValError{ErrorMsg: errors.New("custom_url must be 10 characters or fewer")}
	}

	if customURL == "" {
		existingLink, err := s.repo.GetShortLinkByUrl(ctx, originalURL, duration)
		if err == nil {
			if existingLink.IsMalicious {
				return entity.CreateShortLinkResponse{}, common.ValError{ErrorMsg: errors.New("url is malicious and cannot be shortened")}
			}
			return entity.CreateShortLinkResponse{ShortCode: existingLink.ShortCode, ExpiredAt: existingLink.ExpiredAt}, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.CreateShortLinkResponse{}, common.DBError{ErrorMsg: err}
		}
	}

	var isMalicious bool
	var err error
	if s.threatDetector != nil {
		isMalicious, err = s.threatDetector.IsMaliciousURL(ctx, originalURL)
		if err != nil {
			return entity.CreateShortLinkResponse{}, fmt.Errorf("failed to check url threat: %w", err)
		}
	}
	expiredAt := time.Now().Add(time.Duration(duration) * 24 * time.Hour)
	if isMalicious {
		return entity.CreateShortLinkResponse{}, common.ValError{ErrorMsg: errors.New("url is malicious and cannot be shortened")}
	}
	if customURL != "" {
		err = s.repo.CreateShortLink(ctx, originalURL, customURL, duration, &expiredAt)
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return entity.CreateShortLinkResponse{}, common.ValError{ErrorMsg: errors.New("custom_url is already in use")}
			}
			return entity.CreateShortLinkResponse{}, common.DBError{ErrorMsg: err}
		}
		return entity.CreateShortLinkResponse{ShortCode: customURL, ExpiredAt: &expiredAt}, nil
	}

	for i := 0; i < maxRetries; i++ {
		shortCode, err := generateShortCode(7)
		if err != nil {
			return entity.CreateShortLinkResponse{}, err
		}

		err = s.repo.CreateShortLink(ctx, originalURL, shortCode, duration, &expiredAt)
		if err != nil {
			if isUniqueContraintViolationDetected(err) {
				continue
			}
			return entity.CreateShortLinkResponse{}, common.DBError{ErrorMsg: fmt.Errorf("failed to create short link: %w", err)}
		}

		if s.redis != nil {
			cacheValue := cachedShortLink{OriginalURL: originalURL, ExpiredAt: &expiredAt}
			if err := s.redis.Set(ctx, cacheKeyPrefix+shortCode, cacheValue, cacheDuration(&expiredAt)); err != nil {
				log.Printf("short link cache set failed: %v", err)
			}
		}
		return entity.CreateShortLinkResponse{ShortCode: shortCode, ExpiredAt: &expiredAt}, nil
	}

	return entity.CreateShortLinkResponse{}, fmt.Errorf("failed to create short link after %d retries", maxRetries)
}

func (s *shortLinkService) GetShortLinkByCode(ctx context.Context, shortCode string) (string, error) {
	var cached cachedShortLink
	cacheHit := s.redis != nil && s.redis.Get(ctx, cacheKeyPrefix+shortCode, &cached) == nil && cached.ExpiredAt != nil && time.Now().Before(*cached.ExpiredAt)
	if !cacheHit {
		link, err := s.repo.GetShortLinkByCode(ctx, shortCode)
		if err != nil {
			return "", err
		}
		cached = cachedShortLink{OriginalURL: link.OriginalURL, ExpiredAt: link.ExpiredAt}
	}

	if err := validateDestination(cached.OriginalURL); err != nil {
		return "", err
	}
	if s.threatDetector != nil {
		malicious, err := s.threatDetector.IsMaliciousURL(ctx, cached.OriginalURL)
		if err != nil {
			return "", fmt.Errorf("failed to check url threat: %w", err)
		}
		if malicious {
			if err := s.repo.MarkLinkAsMalicious(ctx, shortCode); err != nil {
				log.Printf("short link malicious flag update failed: %v", err)
			}
			if s.redis != nil {
				if err := s.redis.Delete(ctx, cacheKeyPrefix+shortCode); err != nil {
					log.Printf("short link cache delete failed: %v", err)
				}
			}
			return "", common.ValError{ErrorMsg: errors.New("short link is malicious")}
		}
	}

	if !cacheHit && s.redis != nil && cached.ExpiredAt != nil {
		if err := s.redis.Set(ctx, cacheKeyPrefix+shortCode, cached, cacheDuration(cached.ExpiredAt)); err != nil {
			log.Printf("short link cache set failed: %v", err)
		}
	}
	return cached.OriginalURL, nil
}

func validateDestination(destination string) error {
	parsed, err := url.Parse(destination)
	if err != nil || parsed.Hostname() == "" || (!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) {
		return common.ValError{ErrorMsg: errors.New("original_url must be a valid HTTP or HTTPS URL with a host")}
	}
	return nil
}

func cacheDuration(expiredAt *time.Time) time.Duration {
	if expiredAt == nil {
		return 0
	}
	duration := time.Until(*expiredAt)
	if duration <= 0 {
		return time.Second
	}
	return duration
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
