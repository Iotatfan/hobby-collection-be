package service

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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
	GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error)
	GetCollectionFilterDrawer() (collectionEntity.CollectionFilterDrawerResponse, error)
	UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.CollectionDetailResponse, error)
	UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest) (collectionEntity.CollectionDetailResponse, error)
}

type collectionService struct {
	collectionRepo collectionRepository.CollectionRepository
	cld            *cloudinary.Cloudinary
}

const (
	coverUploadFolder      = "Hobby/Cover"
	collectionUploadFolder = "Hobby/Collection"
	addonUploadFolder      = "Hobby/Addons"
)

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

	return mapCollectionReponse(collection, getPictures(collection), getAddons(collection)), nil
}

func (s *collectionService) GetCollectionList(filters collectionEntity.CollectionFilter) (collectionEntity.CollectionListResponse, error) {
	return s.collectionRepo.GetCollectionList(filters)
}

func (s *collectionService) GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error) {
	return s.collectionRepo.GetCollectionDrawer()
}

func (s *collectionService) GetCollectionFilterDrawer() (collectionEntity.CollectionFilterDrawerResponse, error) {
	return s.collectionRepo.GetCollectionFilterDrawer()
}

func (s *collectionService) UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.CollectionDetailResponse, error) {
	log.Printf("[upload] start title=%q grade_id=%d release_type_id=%d manufacturer_id=%d series_id=%d pictures=%d", payload.Title, payload.GradeID, payload.ReleaseTypeID, payload.ManufacturerID, payload.SeriesID, len(payload.Pictures))

	if payload.Cover != nil {
		coverURL, err := s.uploadImage(payload.Cover, coverUploadFolder)
		if err != nil {
			log.Printf("[upload] cover upload failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.CoverURL = coverURL
		log.Printf("[upload] cover uploaded url=%s", payload.CoverURL)
	}

	for i := range payload.Pictures {
		if payload.Pictures[i] == nil {
			continue
		}
		pictureURL, err := s.uploadImage(payload.Pictures[i], collectionUploadFolder)
		if err != nil {
			log.Printf("[upload] picture[%d] upload failed: %v", i, err)
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.PictureURLs = append(payload.PictureURLs, pictureURL)
		log.Printf("[upload] picture[%d] uploaded url=%s", i, pictureURL)
	}

	if len(payload.AddonNames) > 0 || len(payload.AddonPictures) > 0 {
		if len(payload.AddonNames) != len(payload.AddonPictures) {
			return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("addon_names and addon_pictures must have the same length")}
		}
		for i := range payload.AddonPictures {
			if payload.AddonPictures[i] == nil {
				return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("addon_pictures contains an empty file")}
			}
			addonName := strings.TrimSpace(payload.AddonNames[i])
			if addonName == "" {
				return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("addon_names contains an empty name")}
			}
			addonURL, err := s.uploadImage(payload.AddonPictures[i], addonUploadFolder)
			if err != nil {
				log.Printf("[upload] addon_picture[%d] upload failed: %v", i, err)
				return collectionEntity.CollectionDetailResponse{}, err
			}
			payload.AddonPictureURLs = append(payload.AddonPictureURLs, addonURL)
			log.Printf("[upload] addon_picture[%d] uploaded url=%s", i, addonURL)
		}
	}

	log.Printf("[upload] saving to db cover_url=%s picture_urls=%d", payload.CoverURL, len(payload.PictureURLs))
	collection, err := s.collectionRepo.UploadCollection(payload)
	if err != nil {
		log.Printf("[upload] db save failed: %v", err)
		return collectionEntity.CollectionDetailResponse{}, err
	}
	log.Printf("[upload] db save success collection_id=%d", collection.ID)

	collection, err = s.collectionRepo.GetCollectionByID(collection.ID)
	if err != nil {
		return collectionEntity.CollectionDetailResponse{}, err
	}

	return mapCollectionReponse(collection, getPictures(collection), getAddons(collection)), nil
}

func (s *collectionService) UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest) (collectionEntity.CollectionDetailResponse, error) {
	log.Printf("[update] start collection_id=%d new_pictures=%d keep_picture_ids=%d", id, len(payload.NewPictures), len(payload.ExistingPictureIDs))

	deletePictureIDs := make([]int, 0)
	deletePicturePublicIDs := make([]string, 0)
	if payload.ExistingPictureIDsPresent {
		currentPictures, err := s.collectionRepo.GetPicturesByCollectionID(id)
		if err != nil {
			log.Printf("[update] load current pictures failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}

		currentIDMap := make(map[int]struct{}, len(currentPictures))
		for _, picture := range currentPictures {
			currentIDMap[picture.ID] = struct{}{}
		}

		keepIDMap := make(map[int]struct{}, len(payload.ExistingPictureIDs))
		for _, pictureID := range payload.ExistingPictureIDs {
			if _, exists := currentIDMap[pictureID]; !exists {
				return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("one or more existing_picture_ids do not belong to this collection")}
			}
			keepIDMap[pictureID] = struct{}{}
		}

		for _, picture := range currentPictures {
			if _, keep := keepIDMap[picture.ID]; keep {
				continue
			}
			deletePictureIDs = append(deletePictureIDs, picture.ID)
			if picture.Url != "" {
				deletePicturePublicIDs = append(deletePicturePublicIDs, picture.Url)
			}
		}
	}

	if payload.Cover != nil {
		coverURL, err := s.uploadImage(payload.Cover, coverUploadFolder)
		if err != nil {
			log.Printf("[update] cover upload failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.CoverURL = coverURL
	}

	for i := range payload.NewPictures {
		if payload.NewPictures[i] == nil {
			continue
		}
		pictureURL, err := s.uploadImage(payload.NewPictures[i], collectionUploadFolder)
		if err != nil {
			log.Printf("[update] picture[%d] upload failed: %v", i, err)
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.NewPictureURLs = append(payload.NewPictureURLs, pictureURL)
	}

	deleteAddonIDs := make([]int, 0)
	deleteAddonPublicIDs := make([]string, 0)
	currentAddons := []collectionEntity.Addon{}
	if payload.ExistingAddonIDsPresent || len(payload.UpdateAddonIDs) > 0 {
		addons, err := s.collectionRepo.GetAddonsByCollectionID(id)
		if err != nil {
			log.Printf("[update] load current addons failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}
		currentAddons = addons
	}

	if payload.ExistingAddonIDsPresent {
		currentIDMap := make(map[int]collectionEntity.Addon, len(currentAddons))
		for _, addon := range currentAddons {
			currentIDMap[addon.ID] = addon
		}

		keepIDMap := make(map[int]struct{}, len(payload.ExistingAddonIDs))
		for _, addonID := range payload.ExistingAddonIDs {
			if _, exists := currentIDMap[addonID]; !exists {
				return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("one or more existing_addon_ids do not belong to this collection")}
			}
			keepIDMap[addonID] = struct{}{}
		}

		for _, addon := range currentAddons {
			if _, keep := keepIDMap[addon.ID]; keep {
				continue
			}
			deleteAddonIDs = append(deleteAddonIDs, addon.ID)
			if addon.Picture != "" {
				deleteAddonPublicIDs = append(deleteAddonPublicIDs, addon.Picture)
			}
		}
	}

	if len(payload.UpdateAddonIDs) > 0 {
		if len(payload.UpdateAddonNames) > 0 && len(payload.UpdateAddonNames) != len(payload.UpdateAddonIDs) {
			return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("update_addon_names must have the same length as update_addon_ids")}
		}
		if len(payload.UpdateAddonPictures) > 0 && len(payload.UpdateAddonPictures) != len(payload.UpdateAddonIDs) {
			return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("update_addon_pictures must have the same length as update_addon_ids")}
		}

		currentByID := make(map[int]collectionEntity.Addon, len(currentAddons))
		for _, addon := range currentAddons {
			currentByID[addon.ID] = addon
		}

		payload.UpdateAddonPictureURLs = make([]string, 0, len(payload.UpdateAddonIDs))
		for i := range payload.UpdateAddonIDs {
			addonID := payload.UpdateAddonIDs[i]
			current, ok := currentByID[addonID]
			if !ok {
				return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("one or more update_addon_ids do not belong to this collection")}
			}

			if len(payload.UpdateAddonNames) > 0 {
				if strings.TrimSpace(payload.UpdateAddonNames[i]) == "" {
					return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("update_addon_names contains an empty name")}
				}
			}

			// picture update is optional; keep placeholder alignment for repository
			if len(payload.UpdateAddonPictures) == 0 {
				payload.UpdateAddonPictureURLs = append(payload.UpdateAddonPictureURLs, "")
				continue
			}
			if payload.UpdateAddonPictures[i] == nil {
				payload.UpdateAddonPictureURLs = append(payload.UpdateAddonPictureURLs, "")
				continue
			}

			addonURL, err := s.uploadImage(payload.UpdateAddonPictures[i], addonUploadFolder)
			if err != nil {
				log.Printf("[update] addon_picture[%d] upload failed: %v", i, err)
				return collectionEntity.CollectionDetailResponse{}, err
			}
			payload.UpdateAddonPictureURLs = append(payload.UpdateAddonPictureURLs, addonURL)
			if current.Picture != "" {
				deleteAddonPublicIDs = append(deleteAddonPublicIDs, current.Picture)
			}
		}
	}

	if len(payload.NewAddonNames) > 0 || len(payload.NewAddonPictures) > 0 {
		if len(payload.NewAddonNames) != len(payload.NewAddonPictures) {
			return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("new_addon_names and new_addon_pictures must have the same length")}
		}
		for i := range payload.NewAddonPictures {
			if payload.NewAddonPictures[i] == nil {
				return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("new_addon_pictures contains an empty file")}
			}
			addonName := strings.TrimSpace(payload.NewAddonNames[i])
			if addonName == "" {
				return collectionEntity.CollectionDetailResponse{}, helper.ValError{ErrorMsg: errors.New("new_addon_names contains an empty name")}
			}
			addonURL, err := s.uploadImage(payload.NewAddonPictures[i], addonUploadFolder)
			if err != nil {
				log.Printf("[update] addon_picture[%d] upload failed: %v", i, err)
				return collectionEntity.CollectionDetailResponse{}, err
			}
			payload.NewAddonPictureURLs = append(payload.NewAddonPictureURLs, addonURL)
		}
	}

	if _, err := s.collectionRepo.UpdateCollection(id, payload, deletePictureIDs, deleteAddonIDs); err != nil {
		log.Printf("[update] db update failed: %v", err)
		return collectionEntity.CollectionDetailResponse{}, err
	}

	s.deleteCloudinaryByPublicID(deletePicturePublicIDs)
	s.deleteCloudinaryByPublicID(deleteAddonPublicIDs)

	collection, err := s.collectionRepo.GetCollectionByID(id)
	if err != nil {
		return collectionEntity.CollectionDetailResponse{}, err
	}

	return mapCollectionReponse(collection, getPictures(collection), getAddons(collection)), nil
}

func getPictures(collection collectionEntity.Collection) []collectionEntity.Picture {
	if collection.Pictures == nil {
		return nil
	}
	return *collection.Pictures
}

func getAddons(collection collectionEntity.Collection) []collectionEntity.Addon {
	if collection.Addons == nil {
		return nil
	}
	return *collection.Addons
}

func (s *collectionService) uploadImage(fileHeader *multipart.FileHeader, folder string) (string, error) {
	if s.cld == nil {
		return "", helper.ServiceError{ErrorMsg: "cloudinary client is not configured", Code: http.StatusInternalServerError}
	}
	log.Printf("[upload] uploading file name=%q size=%d", fileHeader.Filename, fileHeader.Size)

	file, err := fileHeader.Open()
	if err != nil {
		return "", helper.ServiceError{ErrorMsg: err.Error(), Code: http.StatusBadRequest}
	}
	defer file.Close()

	result, err := s.cld.Upload.Upload(context.Background(), file, uploader.UploadParams{
		Transformation: "f_auto,q_auto:good,w_800,c_limit",
		Folder:         folder,
	})
	if err != nil {
		return "", helper.ServiceError{ErrorMsg: err.Error(), Code: http.StatusInternalServerError}
	}
	if result.SecureURL == "" {
		return "", helper.ServiceError{ErrorMsg: errors.New("cloudinary returned empty secure url").Error(), Code: http.StatusInternalServerError}
	}

	cachePath, err := toCloudinaryCachePath(result.SecureURL)
	if err != nil {
		return "", helper.ServiceError{ErrorMsg: err.Error(), Code: http.StatusInternalServerError}
	}

	return cachePath, nil
}

func (s *collectionService) deleteCloudinaryByPublicID(publicIDs []string) {
	if s.cld == nil || len(publicIDs) == 0 {
		return
	}

	for _, storedValue := range publicIDs {
		if storedValue == "" {
			continue
		}
		publicID := cachePathToPublicID(storedValue)
		if publicID == "" {
			log.Printf("[cloudinary] skip delete: invalid stored value=%s", storedValue)
			continue
		}
		if _, err := s.cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicID}); err != nil {
			log.Printf("[cloudinary] failed to delete public_id=%s: %v", publicID, err)
		}
	}
}

var cloudinaryVersionSegment = regexp.MustCompile(`^v\d+$`)

func toCloudinaryCachePath(secureURL string) (string, error) {
	parsedURL, err := url.Parse(secureURL)
	if err != nil {
		return "", errors.New("failed to parse cloudinary secure url")
	}

	path := parsedURL.Path
	uploadMarker := "/upload/"
	uploadIndex := strings.Index(path, uploadMarker)
	if uploadIndex < 0 {
		return "", errors.New("cloudinary secure url format is invalid")
	}

	tail := path[uploadIndex+len(uploadMarker):]
	segments := strings.Split(tail, "/")
	versionIndex := -1
	for i, segment := range segments {
		if cloudinaryVersionSegment.MatchString(segment) {
			versionIndex = i
			break
		}
	}
	if versionIndex < 0 || versionIndex >= len(segments)-1 {
		return "", errors.New("cloudinary version/public id is missing")
	}

	publicIDSegments := append([]string{}, segments[versionIndex+1:]...)
	last := publicIDSegments[len(publicIDSegments)-1]
	if dotIndex := strings.LastIndex(last, "."); dotIndex > 0 {
		last = last[:dotIndex]
	}
	if last == "" {
		return "", errors.New("cloudinary public id is invalid")
	}
	publicIDSegments[len(publicIDSegments)-1] = last

	return "/" + segments[versionIndex] + "/" + strings.Join(publicIDSegments, "/"), nil
}

func cachePathToPublicID(storedValue string) string {
	trimmed := strings.TrimSpace(storedValue)
	if trimmed == "" {
		return ""
	}

	trimmed = strings.TrimPrefix(trimmed, "/")
	segments := strings.Split(trimmed, "/")
	if len(segments) >= 2 && cloudinaryVersionSegment.MatchString(segments[0]) {
		return strings.Join(segments[1:], "/")
	}

	return strings.TrimPrefix(trimmed, "/")
}

func mapCollectionReponse(collection collectionEntity.Collection, pictures []collectionEntity.Picture, addons []collectionEntity.Addon) collectionEntity.CollectionDetailResponse {

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
		ID:           collection.ID,
		Title:        collection.Title,
		Type:         collectionTypeResp,
		ReleaseType:  collection.ReleaseType,
		Manufacturer: collection.Manufacturer,
		Status:       collection.Status,
		Series:       collection.Series,
		BuiltAt:      builtAt,
		Cover:        collection.Cover,
		Description:  collection.Description,
		Pictures:     picturesResp,
		Addons:       addons,
	}

	return result
}
