package handler

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iotatfan/hobby-collection-be/internal/common"
	"github.com/iotatfan/hobby-collection-be/internal/config"
)

const (
	defaultAdminTokenTTL = 24 * time.Hour
	minAdminPasswordLen  = 16
)

type AdminHandler struct{}

type issueAdminTokenRequest struct {
	Password string `json:"password" binding:"required"`
}

type issueAdminTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewAdminHandler() AdminHandler {
	return AdminHandler{}
}

func (h *AdminHandler) IssueToken(c *gin.Context) {
	req := issueAdminTokenRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, err)
		return
	}

	cfg := config.GetConfig()
	adminPassword := strings.TrimSpace(cfg.Admin.Password)
	if len(adminPassword) < minAdminPasswordLen {
		common.ErrorResponse(c, common.ServiceError{
			ErrorMsg: "admin password must be configured with at least 16 characters",
			Code:     http.StatusInternalServerError,
		})
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(adminPassword)) != 1 {
		common.ErrorResponse(c, common.JWTError{ErrorMsg: "invalid admin password"})
		return
	}

	secret := strings.TrimSpace(cfg.JWT.Secret)
	if secret == "" {
		common.ErrorResponse(c, common.ServiceError{
			ErrorMsg: "jwt secret is not configured",
			Code:     http.StatusInternalServerError,
		})
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(resolveAdminTokenTTL(cfg.JWT.Access))
	claims := jwt.MapClaims{
		"sub":  "admin",
		"iss":  cfg.JWT.Issuer,
		"aud":  []string{cfg.JWT.Audience},
		"role": cfg.JWT.RequiredRole,
		"iat":  now.Unix(),
		"exp":  expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		common.ErrorResponse(c, common.ServiceError{
			ErrorMsg: "failed to sign admin token",
			Code:     http.StatusInternalServerError,
		})
		return
	}

	common.SuccessResponse(c, issueAdminTokenResponse{
		Token:     signedToken,
		ExpiresAt: expiresAt,
	}, http.StatusOK)
}

func resolveAdminTokenTTL(accessMinutes int) time.Duration {
	if accessMinutes <= 0 {
		return defaultAdminTokenTTL
	}

	return time.Duration(accessMinutes) * time.Minute
}
