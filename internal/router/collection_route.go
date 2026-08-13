package router

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/handler"
	"github.com/iotatfan/hobby-collection-be/internal/middleware"
)

func SetCollectionRoutes(g *gin.Engine, cH handler.CollectionHandler) {
	g.GET("/collection/drawer", cH.GetCollectionDrawer)
	g.GET("/collection/filter", cH.GetCollectionFilter)
	g.GET("/collection/statistics", cH.GetCollectionStatistics)
	g.GET("/collection/shelves", cH.GetCollectionShelves)
	g.GET("/collection/:id", cH.GetCollectionByID)
	g.GET("/collection", cH.GetCollectionList)
	g.POST("/create_collection", middleware.JWTAuth(), cH.UploadCollection)
	g.PATCH("/collection/:id", middleware.JWTAuth(), cH.UpdateCollection)
	g.DELETE("/collection/:id", middleware.JWTAuth(), cH.DeleteCollection)
}
