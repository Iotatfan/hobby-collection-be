package shortlink

import (
	"context"
	"log"
	"os"

	webrisk "cloud.google.com/go/webrisk/apiv1"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/handler"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/repository"
	"github.com/iotatfan/hobby-collection-be/internal/shortlink/service"
	"github.com/iotatfan/hobby-collection-be/pkg/cache"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

func Register(db *gorm.DB, redisClient *cache.RedisCache) *handler.ShortLinkHandler {
	sR := repository.NewShortLinkRepository(db)
	
	var threatDetector service.ThreatDetector
	apiKey := os.Getenv("WEBRISK_API_KEY")
	if apiKey != "" {
		client, err := webrisk.NewClient(context.Background(), option.WithAPIKey(apiKey))
		if err != nil {
			log.Fatalf("Failed to initialize Web Risk client: %v", err)
		}
		threatDetector = service.NewWebRiskDetector(client)
		log.Println("Web Risk threat detection enabled via API Key")
	} else {
		// Initialize with nil for local dev without credentials
		threatDetector = service.NewWebRiskDetector(nil)
		log.Println("Web Risk threat detection disabled (WEBRISK_API_KEY not set)")
	}
	
	sS := service.NewShortLinkService(sR, redisClient, threatDetector)
	return handler.NewShortLinkHandler(sS)
}
