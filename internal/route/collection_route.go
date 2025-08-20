package route

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/handler"
)

func SetCollectionRoutes(g *gin.Engine, cH handler.CollectiontHandler) {
	g.GET("/product/:id", cH.GetCollectionByID)
}
