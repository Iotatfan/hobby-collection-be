package handle

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	CollectionHandler "github.com/iotatfan/hobby-collection-be/internal/collection/handler"
	collectionRepository "github.com/iotatfan/hobby-collection-be/internal/collection/repository"
	collectionService "github.com/iotatfan/hobby-collection-be/internal/collection/service"
	"github.com/iotatfan/hobby-collection-be/internal/route"
	"gorm.io/gorm"
)

func SetupCollection(g *gin.Engine, db *gorm.DB, cld *cloudinary.Cloudinary) {
	colR := collectionRepository.NewCollectionRepository(db)
	colS := collectionService.NewCollectionService(colR, cld)
	colH := CollectionHandler.NewCollectionHandler(colS)
	route.SetCollectionRoutes(g, colH)
}
