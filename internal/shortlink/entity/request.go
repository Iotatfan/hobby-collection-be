package entity

type CreateShortLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ExpiryDays  int    `json:"expiry_days" binding:"required"`
}
