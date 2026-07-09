package collection

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	CollectionHandler "github.com/iotatfan/hobby-collection-be/internal/collection/handler"
	collectionRepository "github.com/iotatfan/hobby-collection-be/internal/collection/repository"
	collectionService "github.com/iotatfan/hobby-collection-be/internal/collection/service"
	"github.com/iotatfan/hobby-collection-be/internal/router"
	"github.com/iotatfan/hobby-collection-be/pkg/cache"
	"gorm.io/gorm"
)

func Register(g *gin.Engine, db *gorm.DB, cld *cloudinary.Cloudinary, redisClient *cache.RedisCache) {
	colR := collectionRepository.NewCollectionRepository(db)
	colS := collectionService.NewCollectionService(colR, cld, redisClient)
	colH := CollectionHandler.NewCollectionHandler(colS)
	router.SetCollectionRoutes(g, colH)
}
