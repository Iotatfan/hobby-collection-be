package service

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"testing"
	"time"

	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/collection/repository"
)

type shelvesRepositoryStub struct {
	repository.CollectionRepository
	calls int
	err   error
}

func (r *shelvesRepositoryStub) GetCollectionShelves(_ context.Context, req entity.CollectionShelvesRequest) (entity.CollectionShelvesResponse, error) {
	r.calls++
	return entity.CollectionShelvesResponse{GunplaShelf: entity.CollectionShelfResponse{Count: req.Offset + req.Limit}}, r.err
}
func (*shelvesRepositoryStub) InvalidateCollectionFilterCache() {}

type shelvesCacheStub struct {
	values map[string][]byte
	ttl    time.Duration
	getErr error
	setErr error
}

func (c *shelvesCacheStub) Get(_ context.Context, key string, dest any) error {
	if c.getErr != nil {
		return c.getErr
	}
	b, ok := c.values[key]
	if !ok {
		return errors.New("cache miss")
	}
	return json.Unmarshal(b, dest)
}
func (c *shelvesCacheStub) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.setErr != nil {
		return c.setErr
	}
	b, err := json.Marshal(value)
	if err == nil {
		c.values[key], c.ttl = b, ttl
	}
	return err
}
func (c *shelvesCacheStub) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		delete(c.values, key)
	}
	return nil
}
func (c *shelvesCacheStub) DeleteByPattern(ctx context.Context, pattern string) error {
	for key := range c.values {
		if matches, _ := path.Match(pattern, key); matches {
			_ = c.Delete(ctx, key)
		}
	}
	return nil
}

func TestGetCollectionShelvesCache(t *testing.T) {
	ctx := context.Background()
	repo := &shelvesRepositoryStub{}
	cache := &shelvesCacheStub{values: map[string][]byte{}}
	s := &collectionService{collectionRepo: repo, redis: cache}
	requests := []entity.CollectionShelvesRequest{
		{},
		{Category: "gunpla", Limit: 12},
		{Category: "figure", Limit: 12},
		{Category: "other-model-kit", Limit: 12},
		{Category: "gunpla", Limit: 6},
		{Category: "gunpla", Limit: 12, Offset: 12},
	}
	for _, req := range requests {
		resp, err := s.GetCollectionShelves(ctx, req)
		if err != nil || resp.GunplaShelf.Count != req.Offset+req.Limit {
			t.Fatalf("request=%+v response=%+v err=%v", req, resp, err)
		}
	}
	if repo.calls != len(requests) || len(cache.values) != len(requests) || cache.ttl != time.Hour {
		t.Fatal("cache variants or TTL incorrect")
	}
	for _, req := range requests {
		resp, err := s.GetCollectionShelves(ctx, req)
		if err != nil || resp.GunplaShelf.Count != req.Offset+req.Limit {
			t.Fatalf("cached response=%+v err=%v", resp, err)
		}
	}
	if repo.calls != len(requests) {
		t.Fatal("cache hits called repository")
	}
	cache.values["unrelated"] = []byte("{}")
	s.invalidateCollectionCaches(ctx, 1)
	if len(cache.values) != 1 || cache.values["unrelated"] == nil {
		t.Fatalf("invalidation left keys: %v", cache.values)
	}
	for _, req := range requests {
		if _, err := s.GetCollectionShelves(ctx, req); err != nil {
			t.Fatal(err)
		}
	}
	if repo.calls != 2*len(requests) {
		t.Fatal("invalidation did not force reload")
	}
}

func TestGetCollectionShelvesRepositoryError(t *testing.T) {
	want := errors.New("database unavailable")
	cache := &shelvesCacheStub{values: map[string][]byte{}}
	s := &collectionService{collectionRepo: &shelvesRepositoryStub{err: want}, redis: cache}
	_, err := s.GetCollectionShelves(context.Background(), entity.CollectionShelvesRequest{Category: "gunpla", Limit: 12})
	if !errors.Is(err, want) || len(cache.values) != 0 {
		t.Fatalf("err=%v cache=%v", err, cache.values)
	}
}

func TestGetCollectionShelvesCacheUnavailable(t *testing.T) {
	cache := &shelvesCacheStub{values: map[string][]byte{}, getErr: errors.New("read failed"), setErr: errors.New("write failed")}
	repo := &shelvesRepositoryStub{}
	s := &collectionService{collectionRepo: repo, redis: cache}
	resp, err := s.GetCollectionShelves(context.Background(), entity.CollectionShelvesRequest{Category: "gunpla", Limit: 12})
	if err != nil || resp.GunplaShelf.Count != 12 || repo.calls != 1 {
		t.Fatalf("response=%+v calls=%d err=%v", resp, repo.calls, err)
	}
}
