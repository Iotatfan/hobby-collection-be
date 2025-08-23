package service

import (
	"time"

	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	collectionRepository "github.com/iotatfan/hobby-collection-be/internal/collection/repository"
)

type CollectionService interface {
	GetCollectionByID(id int) (collectionEntity.CollectionDetailResponse, error)
	GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error)
}

type collectionService struct {
	collectionRepo collectionRepository.CollectionRepository
}

func NewCollectionService(collectionRepo collectionRepository.CollectionRepository) CollectionService {
	return &collectionService{
		collectionRepo: collectionRepo,
	}
}

func (s *collectionService) GetCollectionByID(id int) (collectionEntity.CollectionDetailResponse, error) {
	collection, err := s.collectionRepo.GetCollectionByID(id)
	if err != nil {
		return collectionEntity.CollectionDetailResponse{}, err
	}

	builtAt := time.Time{}
	if collection.BuiltAt != nil {
		builtAt = collection.BuiltAt.Local()
	}

	result := collectionEntity.CollectionDetailResponse{
		ID:          collection.ID,
		Title:       collection.Title,
		Type:        collection.CollectionType,
		ReleaseType: collection.ReleaseType,
		Status:      collection.Status,
		Series:      collection.Series,
		BuiltAt:     builtAt,
		Cover:       collection.Cover,
		Description: collection.Description,
	}

	return result, nil
}

func (s *collectionService) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error) {
	queryResult, err := s.collectionRepo.GetCollectionList(filters)
	if err != nil {
		return collectionEntity.CollectionListResponse{}, err
	}
	result := collectionEntity.CollectionListResponse{}

	for _, collection := range queryResult.Collections {
		newResult := collectionEntity.CollectionDetailResponse{
			ID:          collection.ID,
			Title:       collection.Title,
			Type:        collection.CollectionType,
			ReleaseType: collection.ReleaseType,
			Status:      collection.Status,
			Series:      collection.Series,
			BuiltAt:     collection.BuiltAt.Local(),
			Cover:       collection.Cover,
			Description: collection.Description,
		}

		result.Collections = append(result.Collections, newResult)
	}

	return result, nil
}
