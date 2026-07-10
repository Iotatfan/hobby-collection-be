package shortlink

import (
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/handler"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/repository"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/service"
	"github.com/iotatfan/hobby-collection-be/pkg/cache"
	"gorm.io/gorm"
)

func Register(db *gorm.DB, redisClient *cache.RedisCache) *handler.ShortLinkHandler {
	sR := repository.NewShortLinkRepository(db)
	sS := service.NewShortLinkService(sR, redisClient)
	return handler.NewShortLinkHandler(sS)
}
