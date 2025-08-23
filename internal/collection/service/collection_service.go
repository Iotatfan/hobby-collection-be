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

	return mapCollectionReponse(collection), nil
}

func (s *collectionService) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error) {
	queryResult, err := s.collectionRepo.GetCollectionList(filters)
	if err != nil {
		return collectionEntity.CollectionListResponse{}, err
	}
	result := collectionEntity.CollectionListResponse{}

	for _, collection := range queryResult.Collections {
		result.Collections = append(result.Collections, mapCollectionReponse(collection))
	}

	return result, nil
}

func mapCollectionReponse(collection collectionEntity.Collection) collectionEntity.CollectionDetailResponse {

	builtAt := time.Time{}
	if collection.BuiltAt != nil {
		builtAt = collection.BuiltAt.Local()
	}

	collectionTypeResp := collectionEntity.CollectionTypeResponse{
		ID:                 collection.CollectionType.ID,
		CollectionTypeName: collection.CollectionType.CollectionTypeName,
		Scale:              collection.CollectionType.Scale,
		Grade:              collection.CollectionType.Grade,
	}

	result := collectionEntity.CollectionDetailResponse{
		ID:          collection.ID,
		Title:       collection.Title,
		Type:        collectionTypeResp,
		ReleaseType: collection.ReleaseType,
		Status:      collection.Status,
		Series:      collection.Series,
		BuiltAt:     builtAt,
		Cover:       collection.Cover,
		Description: collection.Description,
	}

	return result
}
