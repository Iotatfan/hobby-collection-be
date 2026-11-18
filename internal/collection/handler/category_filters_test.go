package handler

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/collection/service"
)

type categoryFilterStub struct {
	service.CollectionService
	category string
	legacy   bool
}

func (s *categoryFilterStub) GetCategoryFilters(_ context.Context, category string) (entity.CategoryFilterResponse, error) {
	s.category = category
	return entity.CategoryFilterResponse{Category: category, TotalCount: 18, Grades: []entity.GunplaGradeFilterResponse{}, Scales: []entity.ScaleResponse{}, ReleaseTypes: []entity.ReleaseTypeFilterResponse{}, Manufacturers: []entity.ManufacturerResponse{}, Series: []entity.SeriesResponse{}}, nil
}

func (s *categoryFilterStub) GetCollectionFilter(context.Context) (entity.CollectionFilterResponse, error) {
	s.legacy = true
	return entity.CollectionFilterResponse{}, nil
}

func TestCategoryFilterContract(t *testing.T) {
	for _, query := range []string{"", "category=gunpla", "category=figure", "category=other-model-kit", "category=", "category=backlog", "category=gunpla&category=figure"} {
		t.Run(query, func(t *testing.T) {
			stub := &categoryFilterStub{}
			h := NewCollectionHandler(stub)
			router := gin.New()
			router.GET("/collection/filter", h.GetCollectionFilter)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", "/collection/filter?"+query, nil))
			invalid := query == "category=" || query == "category=backlog" || strings.Contains(query, "&")
			if invalid {
				if response.Code != 400 || stub.category != "" || stub.legacy {
					t.Fatalf("invalid query reached service: %s", response.Body)
				}
			} else if query == "" {
				if response.Code != 200 || !stub.legacy || strings.Contains(response.Body.String(), "total_count") {
					t.Fatalf("legacy response changed: %s", response.Body)
				}
			} else if response.Code != 200 || stub.category == "" || !strings.Contains(response.Body.String(), `"series":[]`) || !strings.Contains(response.Body.String(), `"total_count":18`) {
				t.Fatalf("invalid scoped response: %s", response.Body)
			}
		})
	}
}
