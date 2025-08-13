package service

import (
	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	collectionRepository "github.com/iotatfan/hobby-collection-be/internal/collection/repository"
)

type CollectionService interface {
	GetCollectionByID(id int) (collectionEntity.CollectionDetailResponse, error)
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

	result := collectionEntity.CollectionDetailResponse{
		ID:          collection.ID,
		Title:       collection.Title,
		Scale:       collection.Scale,
		RelaseType:  collection.RelaseType,
		Status:      collection.Status,
		Series:      collection.Series,
		BuiltAt:     collection.BuiltAt.Local(),
		Cover:       collection.Cover,
		Description: collection.Description,
	}

	return result, nil
}
