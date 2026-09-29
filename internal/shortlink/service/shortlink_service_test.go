package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/shortlink/entity"
	"gorm.io/gorm"
)

type testShortLinkRepository struct {
	created int
}

func (r *testShortLinkRepository) CreateShortLink(context.Context, string, string, int, *time.Time) error {
	r.created++
	return nil
}

func (r *testShortLinkRepository) GetShortLinkByCode(context.Context, string) (entity.ShortLink, error) {
	return entity.ShortLink{}, gorm.ErrRecordNotFound
}

func (r *testShortLinkRepository) CheckShortCodeExists(context.Context, string) (bool, error) {
	return false, nil
}

func (r *testShortLinkRepository) GetShortLinkByUrl(context.Context, string, int) (entity.ShortLink, error) {
	return entity.ShortLink{}, gorm.ErrRecordNotFound
}

func (r *testShortLinkRepository) MarkLinkAsMalicious(context.Context, string) error {
	return nil
}

type testThreatDetector struct {
	malicious bool
}

func (d testThreatDetector) IsMaliciousURL(context.Context, string) (bool, error) {
	return d.malicious, nil
}

func TestCreateShortLinkRejectsInvalidExpiry(t *testing.T) {
	repo := &testShortLinkRepository{}
	svc := NewShortLinkService(repo, nil, nil)

	for _, expiry := range []int{0, -1, maxExpiryDays + 1} {
		if _, err := svc.CreateShortLink(context.Background(), "https://example.com", expiry, ""); err == nil {
			t.Fatalf("expected expiry %d to be rejected", expiry)
		}
	}
	if repo.created != 0 {
		t.Fatalf("invalid requests must not create links")
	}
}

func TestCreateShortLinkRejectsMaliciousURLBeforePersistence(t *testing.T) {
	repo := &testShortLinkRepository{}
	svc := NewShortLinkService(repo, nil, testThreatDetector{malicious: true})

	_, err := svc.CreateShortLink(context.Background(), "https://example.com", 1, "custom")
	if err == nil {
		t.Fatal("expected malicious URL to be rejected")
	}
	if repo.created != 0 {
		t.Fatal("malicious URL must not be persisted")
	}
}

func TestCacheDurationNeverExpiresAfterLink(t *testing.T) {
	expiredAt := time.Now().Add(2 * time.Hour)
	duration := cacheDuration(&expiredAt)
	if duration <= 0 || duration > 2*time.Hour {
		t.Fatalf("unexpected cache duration: %s", duration)
	}
}

func TestCreateShortLinkPropagatesRepositoryErrors(t *testing.T) {
	repo := &errorShortLinkRepository{}
	svc := NewShortLinkService(repo, nil, nil)

	_, err := svc.CreateShortLink(context.Background(), "https://example.com", 1, "")
	if err == nil || !errors.Is(err, errRepositoryFailure) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

var errRepositoryFailure = errors.New("repository unavailable")

type errorShortLinkRepository struct{ testShortLinkRepository }

func (*errorShortLinkRepository) GetShortLinkByUrl(context.Context, string, int) (entity.ShortLink, error) {
	return entity.ShortLink{}, errRepositoryFailure
}
