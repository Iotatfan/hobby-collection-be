package repository

import (
	"context"
	"errors"

	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/common"
	"gorm.io/gorm"
)

func (r *collectionRepository) GetCategoryFilters(ctx context.Context, category string) (entity.CategoryFilterResponse, error) {
	result := entity.CategoryFilterResponse{
		Category: category,
		Grades:   []entity.GunplaGradeFilterResponse{}, Scales: []entity.ScaleResponse{},
		ReleaseTypes: []entity.ReleaseTypeFilterResponse{}, Manufacturers: []entity.ManufacturerResponse{},
		Series: []entity.SeriesResponse{},
	}
	if category != "gunpla" && category != "figure" && category != "other-model-kit" {
		return result, common.ValError{ErrorMsg: errors.New("unsupported shelf category")}
	}
	// Match shelf eligibility, including Other Model Kit's grouping across types.
	base := func() *gorm.DB {
		db := r.db.WithContext(ctx).Table("collections c").
			Joins("JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
			Joins("JOIN collection_types ct ON ct.id = g.collection_type_id AND ct.deleted_at IS NULL").
			Where("c.deleted_at IS NULL AND CAST(c.status AS int) <> ?", entity.Backlog)
		switch category {
		case "gunpla":
			return db.Where("ct.id = ?", collectionTypeGunplaID)
		case "figure":
			return db.Where("ct.id = ?", collectionTypeFiguresID)
		default:
			return db.Where("ct.id NOT IN ?", []int{collectionTypeFiguresID, collectionTypeGunplaID})
		}
	}
	if err := base().Count(&result.TotalCount).Error; err != nil {
		return result, common.DBError{ErrorMsg: err}
	}
	if category == "gunpla" {
		if err := base().Select("DISTINCT g.id, g.name, g.short_name").Order("g.short_name ASC").Scan(&result.Grades).Error; err != nil {
			return result, common.DBError{ErrorMsg: err}
		}
		if err := base().Joins("JOIN release_types rt ON rt.id = c.release_type AND rt.deleted_at IS NULL").
			Select("DISTINCT rt.id, rt.name AS release_type_name").Order("release_type_name ASC").Scan(&result.ReleaseTypes).Error; err != nil {
			return result, common.DBError{ErrorMsg: err}
		}
	} else {
		if err := base().Joins("JOIN scales sc ON sc.id = c.scale_id AND sc.deleted_at IS NULL").
			Select("DISTINCT sc.id, sc.name").Order("sc.name ASC").Scan(&result.Scales).Error; err != nil {
			return result, common.DBError{ErrorMsg: err}
		}
		if err := base().Joins("JOIN manufacturers m ON m.id = c.manufacturer AND m.deleted_at IS NULL").
			Select("DISTINCT m.id, m.name AS manufacturer_name").Order("manufacturer_name ASC").Scan(&result.Manufacturers).Error; err != nil {
			return result, common.DBError{ErrorMsg: err}
		}
	}
	if err := base().Joins("JOIN series s ON s.id = c.series_id AND s.deleted_at IS NULL").
		Select("DISTINCT s.id, s.name AS series_name").Order("series_name ASC").Scan(&result.Series).Error; err != nil {
		return result, common.DBError{ErrorMsg: err}
	}
	return result, nil
}
