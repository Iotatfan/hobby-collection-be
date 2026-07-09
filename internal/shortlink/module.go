package shortlink

import (
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/handler"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/repository"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/service"
	"gorm.io/gorm"
)

func Register(db *gorm.DB) *handler.ShortLinkHandler {
	sR := repository.NewShortLinkRepository(db)
	sS := service.NewShortLinkService(sR)
	return handler.NewShortLinkHandler(sS)
}
