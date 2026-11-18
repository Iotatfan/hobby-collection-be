package service

import (
	"context"
	"testing"

	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/collection/repository"
)

type categoryFiltersRepository struct {
	repository.CollectionRepository
	calls      int
	duringRead func()
}

func (r *categoryFiltersRepository) GetCategoryFilters(_ context.Context, category string) (entity.CategoryFilterResponse, error) {
	r.calls++
	if r.duringRead != nil {
		r.duringRead()
	}
	return entity.CategoryFilterResponse{Category: category, TotalCount: 18}, nil
}
func (*categoryFiltersRepository) InvalidateCollectionFilterCache() {}

func TestCategoryFiltersCacheAndInvalidation(t *testing.T) {
	ctx := context.Background()
	repo := &categoryFiltersRepository{}
	cache := &shelvesCacheStub{values: map[string][]byte{}}
	s := &collectionService{collectionRepo: repo, redis: cache}
	for _, category := range []string{"gunpla", "figure", "other-model-kit", "gunpla"} {
		result, err := s.GetCategoryFilters(ctx, category)
		if err != nil || result.Category != category {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	}
	if repo.calls != 3 {
		t.Fatalf("calls=%d", repo.calls)
	}
	cache.values[cacheKeyCollectionFilter] = []byte("{}")
	s.invalidateCollectionCaches(ctx, 1)
	if len(cache.values) != 0 {
		t.Fatalf("collection invalidation left filter keys: %v", cache.values)
	}
	_, _ = s.GetCategoryFilters(ctx, "gunpla")
	s.invalidateCatalogCaches(ctx)
	if len(cache.values) != 0 {
		t.Fatal("catalog invalidation left scoped keys")
	}
	repo.duringRead = func() { s.invalidateCatalogCaches(ctx) }
	_, _ = s.GetCategoryFilters(ctx, "gunpla")
	if len(cache.values) != 0 {
		t.Fatal("pre-mutation request repopulated metadata cache")
	}
}
