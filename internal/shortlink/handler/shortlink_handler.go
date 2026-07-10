package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/config"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/entity"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/service"
)

type ShortLinkHandler struct {
	service service.ShortLinkService
}

func NewShortLinkHandler(service service.ShortLinkService) *ShortLinkHandler {
	return &ShortLinkHandler{
		service: service,
	}
}

func (h *ShortLinkHandler) CreateShortLink(c *gin.Context) {
	var req entity.CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expiredAt := time.Now().Add(240 * time.Hour) // Set expiration to 10 days from now
	resp, err := h.service.CreateShortLink(c.Request.Context(), req.OriginalURL, &expiredAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *ShortLinkHandler) GetShortLink(c *gin.Context) {
	shortCode := c.Param("short_code")

	originalURL, err := h.service.GetShortLinkByCode(c.Request.Context(), shortCode)
	if err != nil {
		c.Redirect(http.StatusFound, config.GetConfig().ShortLink.FeURL+"/notfound")
		return
	}
	c.Redirect(http.StatusFound, originalURL)
}
