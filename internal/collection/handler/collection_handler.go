package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
