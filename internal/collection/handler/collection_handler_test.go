package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/collection/service"
)

type shelvesServiceStub struct {
	service.CollectionService
	request entity.CollectionShelvesRequest
	called  bool
}

func (s *shelvesServiceStub) GetCollectionShelves(_ context.Context, req entity.CollectionShelvesRequest) (entity.CollectionShelvesResponse, error) {
	s.called, s.request = true, req
	return entity.CollectionShelvesResponse{
		GunplaShelf: entity.CollectionShelfResponse{ID: 2, Name: "Gunpla", Items: []entity.ShelfItemResponse{}, Count: 25},
	}, nil
}

func TestGetCollectionShelves(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		want        entity.CollectionShelvesRequest
		bad         bool
	}{
		{name: "preview"},
		{name: "gunpla defaults", query: "category=gunpla", want: entity.CollectionShelvesRequest{Category: "gunpla", Limit: 12}},
		{name: "gunpla explicit defaults", query: "category=gunpla&limit=12&offset=0", want: entity.CollectionShelvesRequest{Category: "gunpla", Limit: 12}},
		{name: "figure", query: "category=figure&limit=12&offset=12", want: entity.CollectionShelvesRequest{Category: "figure", Limit: 12, Offset: 12}},
		{name: "other", query: "category=other-model-kit&limit=12&offset=24", want: entity.CollectionShelvesRequest{Category: "other-model-kit", Limit: 12, Offset: 24}},
		{name: "min limit", query: "category=gunpla&limit=1&offset=0", want: entity.CollectionShelvesRequest{Category: "gunpla", Limit: 1}},
		{name: "max limit", query: "category=gunpla&limit=100", want: entity.CollectionShelvesRequest{Category: "gunpla", Limit: 100}},
		{name: "offset only", query: "offset=0", bad: true},
		{name: "limit only", query: "limit=12", bad: true},
		{name: "empty category", query: "category=", bad: true},
		{name: "unsupported category", query: "category=backlog", bad: true},
		{name: "empty limit", query: "category=gunpla&limit=", bad: true},
		{name: "zero limit", query: "category=gunpla&limit=0", bad: true},
		{name: "negative limit", query: "category=gunpla&limit=-1", bad: true},
		{name: "large limit", query: "category=gunpla&limit=101", bad: true},
		{name: "nonnumeric limit", query: "category=gunpla&limit=x", bad: true},
		{name: "empty offset", query: "category=gunpla&offset=", bad: true},
		{name: "negative offset", query: "category=gunpla&offset=-1", bad: true},
		{name: "fractional offset", query: "category=gunpla&offset=1.5", bad: true},
		{name: "overflow offset", query: "category=gunpla&offset=9999999999999999999999999", bad: true},
		{name: "invalid encoded category", query: "category=%ZZ", bad: true},
		{name: "invalid encoded limit", query: "category=gunpla&limit=%ZZ", bad: true},
		{name: "invalid encoded offset", query: "category=gunpla&offset=%ZZ", bad: true},
		{name: "invalid semicolon", query: "category=gunpla&offset=0;12", bad: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &shelvesServiceStub{}
			h := NewCollectionHandler(stub)
			router := gin.New()
			router.GET("/collection/shelves", h.GetCollectionShelves)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("GET", "/collection/shelves?"+tc.query, nil))
			if tc.bad {
				if w.Code != 400 || stub.called {
					t.Fatalf("status=%d service called=%v", w.Code, stub.called)
				}
				return
			}
			if w.Code != 200 || !stub.called || stub.request != tc.want {
				t.Fatalf("status=%d request=%+v want=%+v", w.Code, stub.request, tc.want)
			}
			var body struct {
				Data entity.CollectionShelvesResponse `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Data.GunplaShelf.Count != 25 || body.Data.GunplaShelf.Items == nil {
				t.Fatalf("unexpected envelope: %s", w.Body.String())
			}
		})
	}
}
