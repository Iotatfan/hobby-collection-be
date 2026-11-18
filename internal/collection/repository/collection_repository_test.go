//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetCollectionShelves(t *testing.T) {
	dsn := os.Getenv("COLLECTION_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("COLLECTION_TEST_POSTGRES_DSN must identify a disposable PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	// All parallel shelf queries share one connection and its temporary tables.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, statement := range []string{
		`CREATE TEMP TABLE collection_types (id int PRIMARY KEY, name text, deleted_at timestamptz)`,
		`CREATE TEMP TABLE grades (id int PRIMARY KEY, name text, short_name text, collection_type_id int, deleted_at timestamptz)`,
		`CREATE TEMP TABLE scales (id int PRIMARY KEY, name text, deleted_at timestamptz)`,
		`CREATE TEMP TABLE collections (id int PRIMARY KEY, title text, status text, cover text, display_size text, grade_id int, scale_id int, release_type int, manufacturer int, series_id int, built_at timestamptz, acquired_at timestamptz, created_at timestamptz, deleted_at timestamptz)`,
		`INSERT INTO collection_types VALUES (1,'Figure',NULL),(2,'Gunpla',NULL),(3,'Other',NULL),(4,'Another',NULL),(5,'Deleted',now())`,
		`INSERT INTO grades VALUES (1,'Figure','F',1,NULL),(2,'Gunpla','G',2,NULL),(3,'Other','O',3,NULL),(4,'Another','A',4,NULL),(5,'Deleted type','D',5,NULL),(6,'Deleted grade','D',2,now())`,
		`INSERT INTO scales VALUES (1,'Scale',NULL),(2,'Deleted',now())`,
		`INSERT INTO collections (id,title,status,grade_id,scale_id,created_at)
		 SELECT type_id*100+n, 'item', '2', type_id, 1, '2020-01-01'::timestamptz
		 FROM generate_series(1,4) AS type_id CROSS JOIN generate_series(1,15) AS n`,
		// Each timestamp fallback takes priority in order; deleted scale remains eligible.
		`UPDATE collections SET built_at='2023-01-01', acquired_at='2030-01-01' WHERE id=201`,
		`UPDATE collections SET acquired_at='2022-01-01' WHERE id=202`,
		`UPDATE collections SET created_at='2021-01-01', scale_id=2 WHERE id=203`,
		`UPDATE collections SET created_at=NULL WHERE id=204`,
		`INSERT INTO collections (id,title,status,grade_id,created_at,deleted_at) VALUES
		 (901,'backlog',1,2,now(),NULL), (902,'deleted',2,2,now(),now()),
		 (903,'deleted grade',2,6,now(),NULL),(904,'deleted type',2,5,now(),NULL),
		 (905,'figure backlog',1,1,now(),NULL),(906,'other backlog',1,3,now(),NULL)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := NewCollectionRepository(db)
	gunpla := []int{201, 202, 203, 215, 214, 213, 212, 211, 210, 209, 208, 207, 206, 205, 204}
	figure := []int{115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101}
	other := []int{415, 414, 413, 412, 411, 410, 409, 408, 407, 406, 405, 404, 403, 402, 401, 315, 314, 313, 312, 311, 310, 309, 308, 307, 306, 305, 304, 303, 302, 301}
	check := func(t *testing.T, shelf entity.CollectionShelfResponse, want []int, count int) {
		t.Helper()
		ids := make([]int, 0, len(shelf.Items))
		for _, item := range shelf.Items {
			ids = append(ids, item.ID)
		}
		if shelf.Items == nil || shelf.Count != count || !reflect.DeepEqual(ids, want) {
			t.Fatalf("items=%v count=%d; want items=%v count=%d", ids, shelf.Count, want, count)
		}
	}
	t.Run("previews", func(t *testing.T) {
		resp, err := repo.GetCollectionShelves(context.Background(), entity.CollectionShelvesRequest{})
		if err != nil {
			t.Fatal(err)
		}
		check(t, resp.GunplaShelf, gunpla[:6], 15)
		check(t, resp.FigureShelf, figure[:6], 15)
		check(t, resp.OtherModelKitShelf, other[:6], 30)
		check(t, resp.BacklogShelf, []int{906, 905, 901}, 3)
	})
	for _, category := range []string{"gunpla", "figure", "other-model-kit"} {
		for _, offset := range []int{0, 12, 15, 24, 30, 100} {
			t.Run(fmt.Sprintf("%s/%d", category, offset), func(t *testing.T) {
				resp, err := repo.GetCollectionShelves(context.Background(), entity.CollectionShelvesRequest{Category: category, Limit: 12, Offset: offset})
				if err != nil {
					t.Fatal(err)
				}
				selected, want := resp.GunplaShelf, gunpla
				switch category {
				case "figure":
					selected, want = resp.FigureShelf, figure
				case "other-model-kit":
					selected, want = resp.OtherModelKitShelf, other
				}
				start := min(offset, len(want))
				check(t, selected, want[start:min(start+12, len(want))], len(want))
				if category != "gunpla" {
					check(t, resp.GunplaShelf, gunpla[:6], 15)
				}
				if category != "figure" {
					check(t, resp.FigureShelf, figure[:6], 15)
				}
				if category != "other-model-kit" {
					check(t, resp.OtherModelKitShelf, other[:6], 30)
				}
			})
		}
	}

	t.Run("category filters", func(t *testing.T) {
		if err := db.Exec("UPDATE collections SET release_type=2, manufacturer=7, series_id=8 WHERE id IN (201,202,901)").Error; err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			name  string
			req   entity.CollectionShelvesRequest
			ids   []int
			count int
		}{
			{"release set", entity.CollectionShelvesRequest{ReleaseTypeIDs: []int{2, 3}}, []int{201, 202}, 2},
			{"manufacturer", entity.CollectionShelvesRequest{ManufacturerID: 7}, []int{201, 202}, 2},
			{"series", entity.CollectionShelvesRequest{SeriesID: 8}, []int{201, 202}, 2},
			{"grade", entity.CollectionShelvesRequest{GradeID: 2}, gunpla[:12], 15},
			{"scale", entity.CollectionShelvesRequest{ScaleID: 2}, []int{203}, 1},
			{"conflicting type", entity.CollectionShelvesRequest{CollectionTypeID: 1}, []int{}, 0},
			{"combined", entity.CollectionShelvesRequest{CollectionTypeID: 2, GradeID: 2, ScaleID: 1, ReleaseTypeIDs: []int{2}, ManufacturerID: 7, SeriesID: 8}, []int{201, 202}, 2},
			{"no matches", entity.CollectionShelvesRequest{SeriesID: 999}, []int{}, 0},
			{"terminal page", entity.CollectionShelvesRequest{ManufacturerID: 7, Offset: 12}, []int{}, 2},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tc.req.Category, tc.req.Limit = "gunpla", 12
				resp, err := repo.GetCollectionShelves(context.Background(), tc.req)
				if err != nil {
					t.Fatal(err)
				}
				check(t, resp.GunplaShelf, tc.ids, tc.count)
				check(t, resp.FigureShelf, figure[:6], 15)
				check(t, resp.OtherModelKitShelf, other[:6], 30)
				check(t, resp.BacklogShelf, []int{906, 905, 901}, 3)
			})
		}
	})
	t.Run("category metadata", func(t *testing.T) {
		for _, statement := range []string{
			`CREATE TEMP TABLE release_types (id int PRIMARY KEY, name text, deleted_at timestamptz)`,
			`CREATE TEMP TABLE manufacturers (id int PRIMARY KEY, name text, deleted_at timestamptz)`,
			`CREATE TEMP TABLE series (id int PRIMARY KEY, name text, deleted_at timestamptz)`,
			`INSERT INTO release_types VALUES (1,'Regular',NULL),(2,'Deleted',now()),(3,'Backlog only',NULL),(4,'Unused',NULL)`,
			`INSERT INTO manufacturers VALUES (1,'Brand',NULL),(2,'Backlog only',NULL),(3,'Unused',NULL)`,
			`INSERT INTO series VALUES (1,'Late page series',NULL),(2,'Backlog only',NULL),(3,'Unused',NULL)`,
			`UPDATE collections SET release_type=1, series_id=1 WHERE id=204`,
			`UPDATE collections SET release_type=2 WHERE id=202`,
			`UPDATE collections SET release_type=3, series_id=2, manufacturer=2 WHERE id IN (901,905,906,902)`,
			`UPDATE collections SET manufacturer=1, series_id=1 WHERE id IN (101,301,401)`,
		} {
			if err := db.Exec(statement).Error; err != nil {
				t.Fatal(err)
			}
		}
		gunpla, err := repo.GetCategoryFilters(context.Background(), "gunpla")
		if err != nil || gunpla.TotalCount != 15 || len(gunpla.Grades) != 1 || len(gunpla.ReleaseTypes) != 1 || gunpla.ReleaseTypes[0].ID != 1 || len(gunpla.Series) != 1 || gunpla.Series[0].ID != 1 {
			t.Fatalf("gunpla=%+v err=%v", gunpla, err)
		}
		for _, category := range []string{"figure", "other-model-kit"} {
			options, err := repo.GetCategoryFilters(context.Background(), category)
			wantCount := int64(15)
			if category == "other-model-kit" {
				wantCount = 30
			}
			if err != nil || options.TotalCount != wantCount || len(options.Scales) != 1 || len(options.Manufacturers) != 1 || options.Manufacturers[0].ID != 1 || len(options.Series) != 1 {
				t.Fatalf("category=%s options=%+v err=%v", category, options, err)
			}
		}
	})
	t.Run("empty category", func(t *testing.T) {
		if err := db.Exec("DELETE FROM collections WHERE grade_id IN (3,4)").Error; err != nil {
			t.Fatal(err)
		}
		options, err := repo.GetCategoryFilters(context.Background(), "other-model-kit")
		if err != nil || options.TotalCount != 0 || options.Series == nil || len(options.Series) != 0 || len(options.Scales) != 0 || len(options.Manufacturers) != 0 {
			t.Fatalf("empty options=%+v err=%v", options, err)
		}
		for _, offset := range []int{0, 12} {
			resp, err := repo.GetCollectionShelves(context.Background(), entity.CollectionShelvesRequest{Category: "other-model-kit", Limit: 12, Offset: offset})
			if err != nil {
				t.Fatal(err)
			}
			check(t, resp.OtherModelKitShelf, []int{}, 0)
		}
	})
}
