package router

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/handler"
	"github.com/iotatfan/hobby-collection-be/internal/middleware"
)

func SetCollectionRoutes(g *gin.Engine, cH handler.CollectionHandler) {
	g.POST("/collection/types", middleware.JWTAuth(), cH.CreateCollectionType)
	g.POST("/collection/grades", middleware.JWTAuth(), cH.CreateGrade)
	g.POST("/collection/scales", middleware.JWTAuth(), cH.CreateScale)
	g.POST("/collection/release-types", middleware.JWTAuth(), cH.CreateReleaseType)
	g.POST("/collection/series", middleware.JWTAuth(), cH.CreateSeries)
	g.POST("/collection/manufacturers", middleware.JWTAuth(), cH.CreateManufacturer)
	g.POST("/collection/metadata-tags", middleware.JWTAuth(), cH.CreateMetadataTag)
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
