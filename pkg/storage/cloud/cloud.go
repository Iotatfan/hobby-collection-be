package cloud

import (
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/iotatfan/hobby-collection-be/internal/config"
)

func NewCld(cfg *config.CloudinaryConfig) *cloudinary.Cloudinary {
	cld, err := cloudinary.NewFromParams(cfg.Name, cfg.Key, cfg.Secret)
	if err != nil {
		panic(fmt.Errorf("database error: %w", err))
	}
	return cld
}
