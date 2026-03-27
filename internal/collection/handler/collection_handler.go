package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	filters := entity.CollectionFilterRequest{}
	err := c.ShouldBindQuery(&filters)
	if err != nil {
		common.ErrorResponse(c, err)
		return
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

func (h *CollectiontHandler) GetCollectionFilterDrawer(c *gin.Context) {
	result, err := h.collectionService.GetCollectionFilterDrawer()
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

	cover, err := c.FormFile("cover")
	if err == nil {
		req.Cover = cover
	}

	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if pictures, ok := form.File["pictures"]; ok && len(pictures) > 0 {
			req.Pictures = pictures
		}
		if pictures, ok := form.File["pictures[]"]; ok && len(pictures) > 0 {
			req.Pictures = append(req.Pictures, pictures...)
		}

		if addonPictures, ok := form.File["addon_pictures"]; ok && len(addonPictures) > 0 {
			req.AddonPictures = addonPictures
		}
		if addonPictures, ok := form.File["addon_pictures[]"]; ok && len(addonPictures) > 0 {
			req.AddonPictures = append(req.AddonPictures, addonPictures...)
		}
	}

	addonNames, hasAddonNames := c.GetPostFormArray("addon_names")
	if !hasAddonNames {
		addonNames, hasAddonNames = c.GetPostFormArray("addon_names[]")
	}
	if !hasAddonNames {
		if raw, exists := c.GetPostForm("addon_names"); exists {
			hasAddonNames = true
			addonNames = strings.Split(raw, ",")
		}
	}
	if hasAddonNames {
		req.AddonNames = addonNames
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

	cover, err := c.FormFile("cover")
	if err == nil {
		req.Cover = cover
	}

	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if pictures, ok := form.File["new_pictures"]; ok && len(pictures) > 0 {
			req.NewPictures = pictures
		}
		if pictures, ok := form.File["new_pictures[]"]; ok && len(pictures) > 0 {
			req.NewPictures = append(req.NewPictures, pictures...)
		}

		if addonPictures, ok := form.File["new_addon_pictures"]; ok && len(addonPictures) > 0 {
			req.NewAddonPictures = addonPictures
		}
		if addonPictures, ok := form.File["new_addon_pictures[]"]; ok && len(addonPictures) > 0 {
			req.NewAddonPictures = append(req.NewAddonPictures, addonPictures...)
		}

		if addonPictures, ok := form.File["update_addon_pictures"]; ok && len(addonPictures) > 0 {
			req.UpdateAddonPictures = addonPictures
		}
		if addonPictures, ok := form.File["update_addon_pictures[]"]; ok && len(addonPictures) > 0 {
			req.UpdateAddonPictures = append(req.UpdateAddonPictures, addonPictures...)
		}
	}

	newAddonNames, hasNewAddonNames := c.GetPostFormArray("new_addon_names")
	if !hasNewAddonNames {
		newAddonNames, hasNewAddonNames = c.GetPostFormArray("new_addon_names[]")
	}
	if !hasNewAddonNames {
		if raw, exists := c.GetPostForm("new_addon_names"); exists {
			hasNewAddonNames = true
			newAddonNames = strings.Split(raw, ",")
		}
	}
	if hasNewAddonNames {
		req.NewAddonNames = newAddonNames
	}

	updateAddonIDsRaw, hasUpdateAddonIDs := c.GetPostFormArray("update_addon_ids")
	if !hasUpdateAddonIDs {
		updateAddonIDsRaw, hasUpdateAddonIDs = c.GetPostFormArray("update_addon_ids[]")
	}
	if !hasUpdateAddonIDs {
		if raw, exists := c.GetPostForm("update_addon_ids"); exists {
			hasUpdateAddonIDs = true
			updateAddonIDsRaw = strings.Split(raw, ",")
		}
	}
	if hasUpdateAddonIDs {
		req.UpdateAddonIDs = make([]int, 0, len(updateAddonIDsRaw))
		for _, v := range updateAddonIDsRaw {
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				continue
			}
			addonID, parseErr := strconv.Atoi(trimmed)
			if parseErr != nil {
				common.ErrorResponse(c, parseErr)
				return
			}
			req.UpdateAddonIDs = append(req.UpdateAddonIDs, addonID)
		}
	}

	updateAddonNames, hasUpdateAddonNames := c.GetPostFormArray("update_addon_names")
	if !hasUpdateAddonNames {
		updateAddonNames, hasUpdateAddonNames = c.GetPostFormArray("update_addon_names[]")
	}
	if !hasUpdateAddonNames {
		if raw, exists := c.GetPostForm("update_addon_names"); exists {
			hasUpdateAddonNames = true
			updateAddonNames = strings.Split(raw, ",")
		}
	}
	if hasUpdateAddonNames {
		req.UpdateAddonNames = updateAddonNames
	}

	deletedURLs, hasDeletedURLs := c.GetPostFormArray("deleted_picture_urls")
	if !hasDeletedURLs {
		deletedURLs, hasDeletedURLs = c.GetPostFormArray("deleted_picture_urls[]")
	}
	if !hasDeletedURLs {
		if raw, exists := c.GetPostForm("deleted_picture_urls"); exists {
			hasDeletedURLs = true
			deletedURLs = strings.Split(raw, ",")
		}
	}
	if hasDeletedURLs {
		req.DeletedPictureURLsPresent = true
		req.DeletedPictureURLs = make([]string, 0, len(deletedURLs))
		for _, v := range deletedURLs {
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				continue
			}
			req.DeletedPictureURLs = append(req.DeletedPictureURLs, trimmed)
		}
	}

	existingAddonIDsRaw, hasExistingAddonIDs := c.GetPostFormArray("existing_addon_ids")
	if !hasExistingAddonIDs {
		existingAddonIDsRaw, hasExistingAddonIDs = c.GetPostFormArray("existing_addon_ids[]")
	}
	if !hasExistingAddonIDs {
		if raw, exists := c.GetPostForm("existing_addon_ids"); exists {
			hasExistingAddonIDs = true
			existingAddonIDsRaw = strings.Split(raw, ",")
		}
	}
	if hasExistingAddonIDs {
		req.ExistingAddonIDsPresent = true
		req.ExistingAddonIDs = make([]int, 0, len(existingAddonIDsRaw))
		for _, v := range existingAddonIDsRaw {
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				continue
			}
			addonID, parseErr := strconv.Atoi(trimmed)
			if parseErr != nil {
				common.ErrorResponse(c, parseErr)
				return
			}
			req.ExistingAddonIDs = append(req.ExistingAddonIDs, addonID)
		}
	}

	result, err := h.collectionService.UpdateCollection(id, req)
	if err != nil {
		common.ErrorResponse(c, err)
		return
	}

	common.SuccessResponse(c, result, http.StatusOK)
}
