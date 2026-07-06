package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/shortlink/entity"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const maxRetries = 5

type ShortLinkService interface {
	CreateShortLink(originalURL string, expiredAt *time.Time) (entity.CreateShortLinkResponse, error)
	GetShortLinkByCode(shortCode string) (string, error)
}

type shortLinkService struct {
	repo repository.ShortLinkRepository
}

func NewShortLinkService(repo repository.ShortLinkRepository) ShortLinkService {
	return &shortLinkService{
		repo: repo,
	}
}

func (s *shortLinkService) CreateShortLink(originalURL string, expiredAt *time.Time) (entity.CreateShortLinkResponse, error) {
	for i := 0; i < maxRetries; i++ {
		shortCode, err := generateShortCode(7)
		if err != nil {
			return entity.CreateShortLinkResponse{}, err
		}

		exists, err := s.repo.CheckShortCodeExists(shortCode)
		if err != nil {
			return entity.CreateShortLinkResponse{}, err
		}
		if !exists {

			return entity.CreateShortLinkResponse{ShortCode: shortCode, ExpiredAt: expiredAt}, s.repo.CreateShortLink(originalURL, shortCode, expiredAt)
		}
	}

	return entity.CreateShortLinkResponse{}, fmt.Errorf("failed to create short link after %d retries", maxRetries)
}

func (s *shortLinkService) GetShortLinkByCode(shortCode string) (string, error) {
	return s.repo.GetShortLinkByCode(shortCode)
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
