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
	"sync"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	collectionEntity "github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	collectionRepository "github.com/iotatfan/hobby-collection-be/internal/collection/repository"
	"github.com/iotatfan/hobby-collection-be/internal/common"
)

type CollectionService interface {
	GetCollectionByID(id int) (collectionEntity.CollectionDetailResponse, error)
	GetCollectionList(filters collectionEntity.CollectionFilterRequest) (collectionEntity.CollectionListResponse, error)
	GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error)
	GetCollectionFilter() (collectionEntity.CollectionFilterResponse, error)
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
	uploadWorkerCount      = 4
	defaultCollectionSort  = "latest_built"
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

	return mapCollectionResponse(collection, getPictures(collection), getAddons(collection)), nil
}

func (s *collectionService) GetCollectionList(filters collectionEntity.CollectionFilterRequest) (collectionEntity.CollectionListResponse, error) {
	if strings.TrimSpace(filters.Sort) == "" {
		filters.Sort = defaultCollectionSort
	}
	return s.collectionRepo.GetCollectionList(filters)
}

func (s *collectionService) GetCollectionDrawer() (collectionEntity.CollectionDrawerResponse, error) {
	return s.collectionRepo.GetCollectionDrawer()
}

func (s *collectionService) GetCollectionFilter() (collectionEntity.CollectionFilterResponse, error) {
	return s.collectionRepo.GetCollectionFilter()
}

func (s *collectionService) UploadCollection(payload collectionEntity.UploadCollectionRequest) (collectionEntity.CollectionDetailResponse, error) {
	uploadedOnThisRequest := make([]string, 0)
	cleanupUploaded := true
	defer func() {
		if cleanupUploaded && len(uploadedOnThisRequest) > 0 {
			s.deleteCloudinaryByPublicID(uploadedOnThisRequest)
		}
	}()

	log.Printf("[upload] start title=%q grade_id=%d scale_id=%d release_type_id=%d manufacturer_id=%d series_id=%d pictures=%d", payload.Title, payload.GradeID, payload.ScaleID, payload.ReleaseTypeID, payload.ManufacturerID, payload.SeriesID, len(payload.Pictures))

	if payload.Cover != nil {
		coverURL, err := s.uploadSingleImage(context.Background(), payload.Cover, coverUploadFolder)
		if err != nil {
			log.Printf("[upload] cover upload failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.CoverURL = coverURL
		uploadedOnThisRequest = append(uploadedOnThisRequest, coverURL)
		log.Printf("[upload] cover uploaded url=%s", payload.CoverURL)
	}

	pictureURLs, err := s.uploadImageBatch(context.Background(), payload.Pictures, collectionUploadFolder, false)
	if err != nil {
		log.Printf("[upload] picture upload failed: %v", err)
		return collectionEntity.CollectionDetailResponse{}, err
	}
	payload.PictureURLs = pictureURLs
	uploadedOnThisRequest = append(uploadedOnThisRequest, pictureURLs...)

	if len(payload.AddonNames) > 0 {
		if len(payload.AddonManufacturerID) > 0 && len(payload.AddonManufacturerID) != len(payload.AddonNames) {
			return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("addons_manufacturer_id must have the same length as addon_names")}
		}
		for i := range payload.AddonNames {
			addonName := strings.TrimSpace(payload.AddonNames[i])
			if addonName == "" {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("addon_names contains an empty name")}
			}
			if len(payload.AddonManufacturerID) > 0 && payload.AddonManufacturerID[i] <= 0 {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("addons_manufacturer_id contains an invalid manufacturer id")}
			}
		}
	}

	log.Printf("[upload] saving to db cover_url=%s picture_urls=%d", payload.CoverURL, len(payload.PictureURLs))
	collection, err := s.collectionRepo.UploadCollection(payload)
	if err != nil {
		log.Printf("[upload] db save failed: %v", err)
		return collectionEntity.CollectionDetailResponse{}, err
	}
	log.Printf("[upload] db save success collection_id=%d", collection.ID)
	cleanupUploaded = false

	collection, err = s.collectionRepo.GetCollectionByID(collection.ID)
	if err != nil {
		return collectionEntity.CollectionDetailResponse{}, err
	}

	return mapCollectionResponse(collection, getPictures(collection), getAddons(collection)), nil
}

func (s *collectionService) UpdateCollection(id int, payload collectionEntity.UpdateCollectionRequest) (collectionEntity.CollectionDetailResponse, error) {
	uploadedOnThisRequest := make([]string, 0)
	cleanupUploaded := true
	defer func() {
		if cleanupUploaded && len(uploadedOnThisRequest) > 0 {
			s.deleteCloudinaryByPublicID(uploadedOnThisRequest)
		}
	}()

	log.Printf("[update] start collection_id=%d new_pictures=%d delete_picture_urls=%d", id, len(payload.NewPictures), len(payload.DeletedPictureURLs))

	deletePictureIDs := make([]int, 0)
	deletePicturePublicIDs := make([]string, 0)
	deleteCoverPublicIDs := make([]string, 0)
	if payload.DeletedPictureURLsPresent {
		currentPictures, err := s.collectionRepo.GetPicturesByCollectionID(id)
		if err != nil {
			log.Printf("[update] load current pictures failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}

		publicIDToPicture := make(map[string]collectionEntity.Picture, len(currentPictures))
		for _, picture := range currentPictures {
			publicID := cloudinaryValueToPublicID(picture.Url)
			if publicID == "" {
				continue
			}
			publicIDToPicture[publicID] = picture
		}

		seenPublicIDs := make(map[string]struct{}, len(payload.DeletedPictureURLs))
		for _, pictureURL := range payload.DeletedPictureURLs {
			publicID := cloudinaryValueToPublicID(pictureURL)
			if publicID == "" {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("one or more deleted_picture_urls are invalid")}
			}
			if _, seen := seenPublicIDs[publicID]; seen {
				continue
			}
			seenPublicIDs[publicID] = struct{}{}

			picture, exists := publicIDToPicture[publicID]
			if !exists {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("one or more deleted_picture_urls do not belong to this collection")}
			}

			deletePictureIDs = append(deletePictureIDs, picture.ID)
			if picture.Url != "" {
				deletePicturePublicIDs = append(deletePicturePublicIDs, picture.Url)
			}
		}
	}

	if payload.Cover != nil {
		currentCollection, err := s.collectionRepo.GetCollectionByID(id)
		if err != nil {
			log.Printf("[update] load current collection failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}

		coverURL, err := s.uploadSingleImage(context.Background(), payload.Cover, coverUploadFolder)
		if err != nil {
			log.Printf("[update] cover upload failed: %v", err)
			return collectionEntity.CollectionDetailResponse{}, err
		}
		payload.CoverURL = coverURL
		uploadedOnThisRequest = append(uploadedOnThisRequest, coverURL)

		if strings.TrimSpace(currentCollection.Cover) != "" {
			deleteCoverPublicIDs = append(deleteCoverPublicIDs, currentCollection.Cover)
		}
	}

	newPictureURLs, err := s.uploadImageBatch(context.Background(), payload.NewPictures, collectionUploadFolder, false)
	if err != nil {
		log.Printf("[update] picture upload failed: %v", err)
		return collectionEntity.CollectionDetailResponse{}, err
	}
	payload.NewPictureURLs = newPictureURLs
	uploadedOnThisRequest = append(uploadedOnThisRequest, newPictureURLs...)

	deleteAddonIDs := make([]int, 0)
	deleteAddonIDMap := make(map[int]struct{})
	currentAddons := []collectionEntity.Addon{}
	if payload.ExistingAddonIDsPresent || payload.DeletedAddonIDsPresent || len(payload.UpdateAddonIDs) > 0 {
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
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("one or more existing_addon_ids do not belong to this collection")}
			}
			keepIDMap[addonID] = struct{}{}
		}

		for _, addon := range currentAddons {
			if _, keep := keepIDMap[addon.ID]; keep {
				continue
			}
			if _, exists := deleteAddonIDMap[addon.ID]; exists {
				continue
			}
			deleteAddonIDMap[addon.ID] = struct{}{}
			deleteAddonIDs = append(deleteAddonIDs, addon.ID)
		}
	}

	if payload.DeletedAddonIDsPresent {
		currentIDMap := make(map[int]collectionEntity.Addon, len(currentAddons))
		for _, addon := range currentAddons {
			currentIDMap[addon.ID] = addon
		}

		for _, addonID := range payload.DeletedAddonIDs {
			if _, exists := currentIDMap[addonID]; !exists {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("one or more deleted_addon_ids do not belong to this collection")}
			}
			if _, exists := deleteAddonIDMap[addonID]; exists {
				continue
			}
			deleteAddonIDMap[addonID] = struct{}{}
			deleteAddonIDs = append(deleteAddonIDs, addonID)
		}
	}

	if len(payload.UpdateAddonIDs) > 0 {
		if len(payload.UpdateAddonNames) > 0 && len(payload.UpdateAddonNames) != len(payload.UpdateAddonIDs) {
			return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("update_addon_names must have the same length as update_addon_ids")}
		}
		if len(payload.UpdateAddonManufacturerID) > 0 && len(payload.UpdateAddonManufacturerID) != len(payload.UpdateAddonIDs) {
			return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("update_addons_manufacturer_id must have the same length as update_addon_ids")}
		}
		currentByID := make(map[int]collectionEntity.Addon, len(currentAddons))
		for _, addon := range currentAddons {
			currentByID[addon.ID] = addon
		}

		for i := range payload.UpdateAddonIDs {
			addonID := payload.UpdateAddonIDs[i]
			_, ok := currentByID[addonID]
			if !ok {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("one or more update_addon_ids do not belong to this collection")}
			}
			if _, addonIsForDelete := deleteAddonIDMap[addonID]; addonIsForDelete {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("one or more update_addon_ids overlap with addons scheduled for deletion")}
			}

			if len(payload.UpdateAddonNames) > 0 {
				if strings.TrimSpace(payload.UpdateAddonNames[i]) == "" {
					return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("update_addon_names contains an empty name")}
				}
			}
			if len(payload.UpdateAddonManufacturerID) > 0 && payload.UpdateAddonManufacturerID[i] <= 0 {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("update_addons_manufacturer_id contains an invalid manufacturer id")}
			}
		}
	}

	if len(payload.NewAddonNames) > 0 {
		if len(payload.NewAddonManufacturerID) > 0 && len(payload.NewAddonManufacturerID) != len(payload.NewAddonNames) {
			return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("new_addons_manufacturer_id must have the same length as new_addon_names")}
		}
		for i := range payload.NewAddonNames {
			addonName := strings.TrimSpace(payload.NewAddonNames[i])
			if addonName == "" {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("new_addon_names contains an empty name")}
			}
			if len(payload.NewAddonManufacturerID) > 0 && payload.NewAddonManufacturerID[i] <= 0 {
				return collectionEntity.CollectionDetailResponse{}, common.ValError{ErrorMsg: errors.New("new_addons_manufacturer_id contains an invalid manufacturer id")}
			}
		}
	}

	if _, err := s.collectionRepo.UpdateCollection(id, payload, deletePictureIDs, deleteAddonIDs); err != nil {
		log.Printf("[update] db update failed: %v", err)
		return collectionEntity.CollectionDetailResponse{}, err
	}
	cleanupUploaded = false

	s.deleteCloudinaryByPublicID(deleteCoverPublicIDs)
	s.deleteCloudinaryByPublicID(deletePicturePublicIDs)

	collection, err := s.collectionRepo.GetCollectionByID(id)
	if err != nil {
		return collectionEntity.CollectionDetailResponse{}, err
	}

	return mapCollectionResponse(collection, getPictures(collection), getAddons(collection)), nil
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

func (s *collectionService) uploadSingleImage(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {
	if s.cld == nil {
		return "", common.ServiceError{ErrorMsg: "cloudinary client is not configured", Code: http.StatusInternalServerError}
	}
	log.Printf("[upload] uploading file name=%q size=%d", fileHeader.Filename, fileHeader.Size)

	file, err := fileHeader.Open()
	if err != nil {
		return "", common.ServiceError{ErrorMsg: err.Error(), Code: http.StatusBadRequest}
	}
	defer file.Close()

	result, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Transformation: "f_auto,q_auto:good,w_1920,c_limit",
		Folder:         folder,
	})
	if err != nil {
		return "", common.ServiceError{ErrorMsg: err.Error(), Code: http.StatusInternalServerError}
	}
	if result.SecureURL == "" {
		return "", common.ServiceError{ErrorMsg: errors.New("cloudinary returned empty secure url").Error(), Code: http.StatusInternalServerError}
	}

	cachePath, err := toCloudinaryCachePath(result.SecureURL)
	if err != nil {
		return "", common.ServiceError{ErrorMsg: err.Error(), Code: http.StatusInternalServerError}
	}

	return cachePath, nil
}

func (s *collectionService) uploadImageBatch(parentCtx context.Context, files []*multipart.FileHeader, folder string, preserveAlignment bool) ([]string, error) {
	if len(files) == 0 {
		return []string{}, nil
	}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	workerCount := uploadWorkerCount
	if len(files) < workerCount {
		workerCount = len(files)
	}

	type uploadResult struct {
		index int
		url   string
		err   error
	}

	jobs := make(chan int)
	results := make(chan uploadResult, len(files))

	var wg sync.WaitGroup
	var cancelOnce sync.Once
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case idx, ok := <-jobs:
					if !ok {
						return
					}
					if files[idx] == nil {
						results <- uploadResult{index: idx}
						continue
					}
					url, err := s.uploadSingleImage(ctx, files[idx], folder)
					if err != nil {
						cancelOnce.Do(cancel)
					}
					results <- uploadResult{index: idx, url: url, err: err}
				}
			}
		}()
	}

	go func() {
		for idx := range files {
			select {
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				close(results)
				return
			case jobs <- idx:
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	ordered := make([]string, len(files))
	var firstErr error
	for result := range results {
		if result.err != nil && firstErr == nil {
			firstErr = result.err
			cancelOnce.Do(cancel)
			continue
		}
		ordered[result.index] = result.url
	}
	if firstErr != nil {
		uploadedBeforeFailure := make([]string, 0, len(ordered))
		for _, value := range ordered {
			if strings.TrimSpace(value) == "" {
				continue
			}
			uploadedBeforeFailure = append(uploadedBeforeFailure, value)
		}
		if len(uploadedBeforeFailure) > 0 {
			s.deleteCloudinaryByPublicID(uploadedBeforeFailure)
		}
		return nil, firstErr
	}

	if preserveAlignment {
		return ordered, nil
	}

	compact := make([]string, 0, len(ordered))
	for _, value := range ordered {
		if strings.TrimSpace(value) == "" {
			continue
		}
		compact = append(compact, value)
	}

	return compact, nil
}

func (s *collectionService) deleteCloudinaryByPublicID(publicIDs []string) {
	if s.cld == nil || len(publicIDs) == 0 {
		return
	}

	for _, storedValue := range publicIDs {
		if storedValue == "" {
			continue
		}
		publicID := cloudinaryValueToPublicID(storedValue)
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

func cloudinaryValueToPublicID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		cachePath, err := toCloudinaryCachePath(trimmed)
		if err != nil {
			return ""
		}
		return cachePathToPublicID(cachePath)
	}

	return cachePathToPublicID(trimmed)
}

func mapCollectionResponse(collection collectionEntity.Collection, pictures []collectionEntity.Picture, addons []collectionEntity.Addon) collectionEntity.CollectionDetailResponse {

	var builtAt *time.Time
	if collection.BuiltAt != nil {
		localBuiltAt := collection.BuiltAt.Local()
		builtAt = &localBuiltAt
	}

	var acquiredAt *time.Time
	if collection.AcquiredAt != nil {
		localAcquiredAt := collection.AcquiredAt.Local()
		acquiredAt = &localAcquiredAt
	}

	collectionTypeResp := collectionEntity.CollectionTypeResponse{
		ID:                 collection.CollectionType.ID,
		CollectionTypeName: collection.CollectionType.CollectionTypeName,
		Scale: collection.CollectionType.Scale.Name,
		Grade: collectionEntity.GradeResponse{
			ID:               collection.CollectionType.Grade.ID,
			Name:             collection.CollectionType.Grade.Name,
			ShortName:        collection.CollectionType.Grade.ShortName,
			CollectionTypeID: collection.CollectionType.Grade.CollectionTypeID,
		},
	}

	var picturesResp []string
	for _, picture := range pictures {
		picturesResp = append(picturesResp, picture.Url)
	}

	addonsResp := make([]collectionEntity.AddonResponse, 0, len(addons))
	for _, addon := range addons {
		addonsResp = append(addonsResp, collectionEntity.AddonResponse{
			ID:           addon.ID,
			AddonName:    addon.AddonName,
			CollectionID: addon.CollectionID,
			Manufacturer: collectionEntity.ManufacturerResponse{
				ID:               addon.ManufacturerID,
				ManufacturerName: addon.Manufacturer.ManufacturerName,
			},
		})
	}

	result := collectionEntity.CollectionDetailResponse{
		ID:    collection.ID,
		Title: collection.Title,
		Type:  collectionTypeResp,
		ReleaseType: collectionEntity.ReleaseTypeResponse{
			ID:              collection.ReleaseType.ID,
			ReleaseTypeName: collection.ReleaseType.ReleaseTypeName,
		},
		Manufacturer: collectionEntity.ManufacturerResponse{
			ID:               collection.Manufacturer.ID,
			ManufacturerName: collection.Manufacturer.ManufacturerName,
		},
		Status: collection.Status,
		Series: collectionEntity.SeriesResponse{
			ID:         collection.Series.ID,
			SeriesName: collection.Series.SeriesName,
		},
		BuiltAt:     builtAt,
		AcquiredAt:  acquiredAt,
		Cover:       collection.Cover,
		Description: collection.Description,
		Pictures:    picturesResp,
		Addons:      addonsResp,
	}

	return result
}
