package entity

type CreateShortLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	CustomURL   string `json:"custom_url" binding:"omitempty,alphanum,max=10"`
	ExpiryDays  int    `json:"expiry_days" binding:"required,min=1,max=365"`
}
