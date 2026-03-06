package service

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	collectionRepository "github.com/iotatfan/hobby-collection-be/internal/collection/repository"
	"github.com/iotatfan/hobby-collection-be/internal/helper"
)

type CollectionService interface {
	GetCollectionByID(id int) (collectionEntity.CollectionDetailResponse, error)
	GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error)
	UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.CollectionDetailResponse, error)
}

type collectionService struct {
	collectionRepo collectionRepository.CollectionRepository
	cld            *cloudinary.Cloudinary
}

func NewCollectionService(collectionRepo collectionRepository.CollectionRepository, cld *cloudinary.Cloudinary) CollectionService {
	return &collectionService{
		collectionRepo: collectionRepo,
		cld:            cld,
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
	if payload.Cover != "" {
		coverURL, err := s.uploadImage(payload.Cover)
		if err != nil {
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.Cover = coverURL
	}

	for i := range payload.Pictures {
		if payload.Pictures[i].Url == "" {
			continue
		}
		pictureURL, err := s.uploadImage(payload.Pictures[i].Url)
		if err != nil {
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.Pictures[i].Url = pictureURL
	}

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

func (s *collectionService) uploadImage(file string) (string, error) {
	if s.cld == nil {
		return "", helper.ServiceError{ErrorMsg: "cloudinary client is not configured", Code: http.StatusInternalServerError}
	}

	result, err := s.cld.Upload.Upload(context.Background(), file, uploader.UploadParams{})
	if err != nil {
		return "", helper.ServiceError{ErrorMsg: err.Error(), Code: http.StatusInternalServerError}
	}

	return result.SecureURL, nil
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
