package entity

import "time"

type CreateShortLinkResponse struct {
	ShortCode string     `json:"short_code"`
	ExpiredAt *time.Time `json:"expired_at"`
}

type GetShortLinkResponse struct {
	OriginalURL string `json:"original_url"`
}
