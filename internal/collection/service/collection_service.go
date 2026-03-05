package service

import (
	"time"

	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	collectionRepository "github.com/iotatfan/hobby-collection-be/internal/collection/repository"
)

type CollectionService interface {
	GetCollectionByID(id int) (collectionEntity.CollectionDetailResponse, error)
	GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error)
	UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.CollectionDetailResponse, error)
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

	pictures, err := s.collectionRepo.GetPicturesByCollectionID(id)
	if err != nil {
		return mapCollectionReponse(collection, nil), nil
	}

	return mapCollectionReponse(collection, pictures), nil
}

func (s *collectionService) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error) {
	queryResult, err := s.collectionRepo.GetCollectionList(filters)
	if err != nil {
		return collectionEntity.CollectionListResponse{}, err
	}
	result := collectionEntity.CollectionListResponse{}

	for _, collection := range queryResult.Collections {
		result.Collections = append(result.Collections, mapCollectionReponse(collection, nil))
	}

	return result, nil
}

func (s *collectionService) UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.CollectionDetailResponse, error) {
	collection, err := s.collectionRepo.UploadCollection(payload)
	if err != nil {
		return collectionEntity.CollectionDetailResponse{}, err
	}

	pictures, err := s.collectionRepo.GetPicturesByCollectionID(collection.ID)
	if err != nil {
		return mapCollectionReponse(collection, nil), nil
	}

	return mapCollectionReponse(collection, pictures), nil
}

func mapCollectionReponse(collection collectionEntity.Collection, pictures []collectionEntity.Picture) collectionEntity.CollectionDetailResponse {

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

	var picturesResp []string
	for _, picture := range pictures {
		picturesResp = append(picturesResp, picture.Url)
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
		Pictures:    picturesResp,
	}

	return result
}
