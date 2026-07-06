package router

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/handler"
)

func SetShortLinkRoutes(g *gin.Engine, sH *handler.ShortLinkHandler) {
	g.POST("/shortlink", sH.CreateShortLink)
	g.GET("/shortlink/:short_code", sH.GetShortLink)
}
