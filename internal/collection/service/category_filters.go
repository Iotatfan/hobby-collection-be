package service

import (
	"context"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
)

func (s *collectionService) GetCategoryFilters(ctx context.Context, category string) (entity.CategoryFilterResponse, error) {
	key := cacheKeyCollectionFilter + ":" + category
	var result entity.CategoryFilterResponse
	s.filterCacheMu.Lock()
	generation := s.filterGeneration
	if s.redis != nil && s.redis.Get(ctx, key, &result) == nil {
		s.filterCacheMu.Unlock()
		return result, nil
	}
	s.filterCacheMu.Unlock()
	result, err := s.collectionRepo.GetCategoryFilters(ctx, category)
	if err != nil {
		return result, err
	}
	s.filterCacheMu.Lock()
	defer s.filterCacheMu.Unlock()
	// A mutation must not be undone by a metadata request started before it.
	if s.redis != nil && generation == s.filterGeneration {
		_ = s.redis.Set(ctx, key, result, 5*time.Minute)
	}
	return result, nil
}
