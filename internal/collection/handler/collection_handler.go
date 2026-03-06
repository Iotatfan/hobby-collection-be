package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	collectionService "github.com/iotatfan/hobby-collection-be/internal/collection/service"
	"github.com/iotatfan/hobby-collection-be/internal/helper"
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
		helper.ErrorResponse(c, err)
		return
	}

	collection, err := h.collectionService.GetCollectionByID(id)
	if err != nil {
		helper.ErrorResponse(c, err)
		return
	}
	helper.SuccessResponse(c, collection, http.StatusOK)
}

func (h *CollectiontHandler) GetCollectionList(c *gin.Context) {
	filters := entity.CollectionFilter{}
	err := c.ShouldBindQuery(&filters)
	if err != nil {
		helper.ErrorResponse(c, err)
		return
	}
	result, err := h.collectionService.GetCollectionList(filters)
	if err != nil {
		helper.ErrorResponse(c, err)
		return
	}
	helper.SuccessResponse(c, result, http.StatusOK)
}

func (h *CollectiontHandler) UploadCollection(c *gin.Context) {
	req := entity.UploadCollectionRequest{}
	if err := c.ShouldBind(&req); err != nil {
		helper.ErrorResponse(c, err)
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
	}

	if req.Cover == nil {
		helper.ErrorResponse(c, helper.ValError{ErrorMsg: errors.New("the field Cover is required")})
		return
	}
	if len(req.Pictures) == 0 {
		helper.ErrorResponse(c, helper.ValError{ErrorMsg: errors.New("the field Pictures is required")})
		return
	}

	result, err := h.collectionService.UploadCollection(req)
	if err != nil {
		helper.ErrorResponse(c, err)
		return
	}

	helper.SuccessResponse(c, result, http.StatusCreated)
}
