package handler

import (
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
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorResponse(c, err)
		return
	}

	result, err := h.collectionService.UploadCollection(req)
	if err != nil {
		helper.ErrorResponse(c, err)
		return
	}

	helper.SuccessResponse(c, result, http.StatusCreated)
}
