package handler

import (
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	collectionService "github.com/iotatfan/hobby-collection-be/internal/collection/service"
	"github.com/iotatfan/hobby-collection-be/internal/common"
)

type CollectiontHandler struct {
	collectionService collectionService.CollectionService
}

func NewCollectionHandler(s collectionService.CollectionService) CollectiontHandler {
	return CollectiontHandler{collectionService: s}
}

func (h *CollectiontHandler) GetCollectionByID(c *gin.Context) {
	startedAt := time.Now()
	defer func() {
		log.Printf("[http] path=%s method=%s id=%s duration=%s status=%d", c.FullPath(), c.Request.Method, c.Param("id"), time.Since(startedAt), c.Writer.Status())
	}()

	inputID := c.Param("id")
	id, err := strconv.Atoi(inputID)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}

	collection, err := h.collectionService.GetCollectionByID(id)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	common.SuccessResponse(c, collection, http.StatusOK)
}

func (h *CollectiontHandler) GetCollectionList(c *gin.Context) {
	startedAt := time.Now()
	defer func() {
		log.Printf("[http] path=%s method=%s raw_query=%q duration=%s status=%d", c.FullPath(), c.Request.Method, c.Request.URL.RawQuery, time.Since(startedAt), c.Writer.Status())
	}()

	filters := entity.CollectionFilterRequest{}
	err := c.ShouldBindQuery(&filters)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}

	releaseTypeIDs, hasReleaseTypeIDs, err := parseFlexibleIntQueryArray(c, "release_type_id", "release_type_ids")
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	if hasReleaseTypeIDs {
		filters.ReleaseTypeIDs = releaseTypeIDs
	}

	result, err := h.collectionService.GetCollectionList(filters)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	common.SuccessResponse(c, result, http.StatusOK)
}

func (h *CollectiontHandler) GetCollectionDrawer(c *gin.Context) {
	result, err := h.collectionService.GetCollectionDrawer()
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	common.SuccessResponse(c, result, http.StatusOK)
}

func (h *CollectiontHandler) GetCollectionFilter(c *gin.Context) {
	startedAt := time.Now()
	defer func() {
		log.Printf("[http] path=%s method=%s duration=%s status=%d", c.FullPath(), c.Request.Method, time.Since(startedAt), c.Writer.Status())
	}()

	result, err := h.collectionService.GetCollectionFilter()
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	common.SuccessResponse(c, result, http.StatusOK)
}

func (h *CollectiontHandler) UploadCollection(c *gin.Context) {
	req := entity.UploadCollectionRequest{}
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err)
		return
	}

	req.Cover = readFormFileIfExists(c, "cover")
	req.Pictures = readMultipartFiles(c, "pictures", "pictures[]")

	addonNames, hasAddonNames := readFlexibleFormArray(c, "addon_names")
	if hasAddonNames {
		req.AddonNames = addonNames
	}

	addonManufacturerIDs, hasAddonManufacturerIDs, err := parseFlexibleIntFormArray(c, "addons_manufacturer_id")
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	if hasAddonManufacturerIDs {
		req.AddonManufacturerID = addonManufacturerIDs
	}

	if req.Cover == nil {
		common.ErrorResponse(c, common.ValError{ErrorMsg: errors.New("the field Cover is required")})
		return
	}
	if len(req.Pictures) == 0 {
		common.ErrorResponse(c, common.ValError{ErrorMsg: errors.New("the field Pictures is required")})
		return
	}

	result, err := h.collectionService.UploadCollection(req)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}

	common.SuccessResponse(c, result, http.StatusCreated)
}

func (h *CollectiontHandler) UpdateCollection(c *gin.Context) {
	inputID := c.Param("id")
	id, err := strconv.Atoi(inputID)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}

	req := entity.UpdateCollectionRequest{}
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResponse(c, err)
		return
	}

	req.Cover = readFormFileIfExists(c, "cover")
	req.NewPictures = readMultipartFiles(c, "new_pictures", "new_pictures[]")

	newAddonNames, hasNewAddonNames := readFlexibleFormArray(c, "new_addon_names")
	if hasNewAddonNames {
		req.NewAddonNames = newAddonNames
	}

	newAddonManufacturerIDs, hasNewAddonManufacturerIDs, err := parseFlexibleIntFormArray(c, "new_addons_manufacturer_id")
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	if hasNewAddonManufacturerIDs {
		req.NewAddonManufacturerID = newAddonManufacturerIDs
	}

	updateAddonIDs, hasUpdateAddonIDs, err := parseFlexibleIntFormArray(c, "update_addon_ids")
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	if hasUpdateAddonIDs {
		req.UpdateAddonIDs = updateAddonIDs
	}

	updateAddonNames, hasUpdateAddonNames := readFlexibleFormArray(c, "update_addon_names")
	if hasUpdateAddonNames {
		req.UpdateAddonNames = updateAddonNames
	}

	updateAddonManufacturerIDs, hasUpdateAddonManufacturerIDs, err := parseFlexibleIntFormArray(c, "update_addons_manufacturer_id")
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	if hasUpdateAddonManufacturerIDs {
		req.UpdateAddonManufacturerID = updateAddonManufacturerIDs
	}

	deletedURLs, hasDeletedURLs := readFlexibleFormArray(c, "deleted_picture_urls")
	if hasDeletedURLs {
		req.DeletedPictureURLsPresent = true
		req.DeletedPictureURLs = trimNonEmptyStrings(deletedURLs)
	}

	existingAddonIDs, hasExistingAddonIDs, err := parseFlexibleIntFormArray(c, "existing_addon_ids")
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	if hasExistingAddonIDs {
		req.ExistingAddonIDsPresent = true
		req.ExistingAddonIDs = existingAddonIDs
	}

	deletedAddonIDs, hasDeletedAddonIDs, err := parseFlexibleIntFormArray(c, "deleted_addon_ids")
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}
	if hasDeletedAddonIDs {
		req.DeletedAddonIDsPresent = true
		req.DeletedAddonIDs = deletedAddonIDs
	}

	result, err := h.collectionService.UpdateCollection(id, req)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}

	common.SuccessResponse(c, result, http.StatusOK)
}

func readFormFileIfExists(c *gin.Context, key string) *multipart.FileHeader {
	file, err := c.FormFile(key)
	if err != nil {
		return nil
	}
	return file
}

func readMultipartFiles(c *gin.Context, keys ...string) []*multipart.FileHeader {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return nil
	}

	files := make([]*multipart.FileHeader, 0)
	for _, key := range keys {
		formFiles, ok := form.File[key]
		if !ok || len(formFiles) == 0 {
			continue
		}
		files = append(files, formFiles...)
	}

	return files
}

func readFlexibleFormArray(c *gin.Context, key string) ([]string, bool) {
	values, exists := c.GetPostFormArray(key)
	if exists {
		return values, true
	}

	values, exists = c.GetPostFormArray(key + "[]")
	if exists {
		return values, true
	}

	value, exists := c.GetPostForm(key)
	if exists {
		return strings.Split(value, ","), true
	}

	return nil, false
}

func parseFlexibleIntFormArray(c *gin.Context, key string) ([]int, bool, error) {
	values, exists := readFlexibleFormArray(c, key)
	if !exists {
		return nil, false, nil
	}

	result := make([]int, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		parsed, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, true, err
		}
		result = append(result, parsed)
	}

	return result, true, nil
}

func readFlexibleQueryArray(c *gin.Context, keys ...string) ([]string, bool) {
	for _, key := range keys {
		values, exists := c.GetQueryArray(key)
		if exists {
			return values, true
		}

		value := c.Query(key)
		if strings.TrimSpace(value) != "" {
			return strings.Split(value, ","), true
		}
	}

	return nil, false
}

func parseFlexibleIntQueryArray(c *gin.Context, keys ...string) ([]int, bool, error) {
	values, exists := readFlexibleQueryArray(c, keys...)
	if !exists {
		return nil, false, nil
	}

	result := make([]int, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		parsed, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, true, err
		}
		result = append(result, parsed)
	}

	return result, true, nil
}

func trimNonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
