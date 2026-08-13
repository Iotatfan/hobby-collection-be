package entity

type CreateShortLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	CustomURL   string `json:"custom_url" binding:"omitempty,alphanum"`
	ExpiryDays  int    `json:"expiry_days" binding:"required"`
}
