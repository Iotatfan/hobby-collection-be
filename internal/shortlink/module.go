package shortlink

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/router"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/handler"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/repository"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/service"
	"gorm.io/gorm"
)

func Register(g *gin.Engine, db *gorm.DB) {
	sR := repository.NewShortLinkRepository(db)
	sS := service.NewShortLinkService(sR)
	sH := handler.NewShortLinkHandler(sS)
	router.SetShortLinkRoutes(g, sH)
}
