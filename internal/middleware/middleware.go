package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iotatfan/hobby-collection-be/internal/common"
	"github.com/iotatfan/hobby-collection-be/internal/config"
	"github.com/iotatfan/hobby-collection-be/internal/text"
)

var requestSeq uint64

const (
	defaultReadRequestsPerMinute  = 120
	defaultWriteRequestsPerMinute = 20
	defaultRateLimitBurst         = 10
	rateLimitWindow               = time.Minute
	minJWTSecretLength            = 32
)

type rateLimitState struct {
	windowStartedAt time.Time
	count           int
}

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string]rateLimitState
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		requests: make(map[string]rateLimitState),
	}
}

func (r *rateLimiter) allow(key string, limit int, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.requests[key]
	if !exists || now.Sub(state.windowStartedAt) >= rateLimitWindow {
		r.requests[key] = rateLimitState{
			windowStartedAt: now,
			count:           1,
		}
		return true
	}

	if state.count >= limit {
		return false
	}

	state.count++
	r.requests[key] = state
	return true
}

var globalRateLimiter = newRateLimiter()

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RateLimit() gin.HandlerFunc {
	cfg := config.GetConfig().RateLimit
	readLimit := cfg.ReadRequestsPerMinute
	writeLimit := cfg.WriteRequestsPerMinute
	burst := cfg.Burst

	if readLimit <= 0 {
		readLimit = defaultReadRequestsPerMinute
	}
	if writeLimit <= 0 {
		writeLimit = defaultWriteRequestsPerMinute
	}
	if burst < 0 {
		burst = 0
	}
	if burst == 0 {
		burst = defaultRateLimitBurst
	}

	return func(c *gin.Context) {
		limit := readLimit
		scope := "read"
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.Request.Method != http.MethodOptions {
			limit = writeLimit
			scope = "write"
		}

		limit += burst
		clientIP := c.ClientIP()
		now := time.Now()
		key := scope + ":" + clientIP
		if !globalRateLimiter.allow(key, limit, now) {
			c.Header("Retry-After", strconv.Itoa(int(rateLimitWindow/time.Second)))
			common.ErrorResponse(c, common.ServiceError{
				ErrorMsg: "rate limit exceeded",
				Code:     http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.ErrorResponse(c, common.JWTError{ErrorMsg: text.NoAuth})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" || tokenString == authHeader {
			common.ErrorResponse(c, common.JWTError{ErrorMsg: text.NoAuth})
			c.Abort()
			return
		}

		jwtCfg := config.GetConfig().JWT
		secret := strings.TrimSpace(jwtCfg.Secret)
		if secret == "" {
			common.ErrorResponse(c, common.ServiceError{ErrorMsg: "jwt secret is not configured", Code: http.StatusInternalServerError})
			c.Abort()
			return
		}
		if len(secret) < minJWTSecretLength {
			common.ErrorResponse(c, common.ServiceError{
				ErrorMsg: fmt.Sprintf("jwt secret must be at least %d characters", minJWTSecretLength),
				Code:     http.StatusInternalServerError,
			})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			switch token.Method.Alg() {
			case jwt.SigningMethodHS256.Alg(), jwt.SigningMethodHS384.Alg(), jwt.SigningMethodHS512.Alg():
				return []byte(secret), nil
			default:
				return nil, common.JWTError{ErrorMsg: text.InvToken}
			}
		}, jwt.WithExpirationRequired(), jwt.WithIssuedAt())
		if err != nil || token == nil || !token.Valid {
			common.ErrorResponse(c, common.JWTError{ErrorMsg: text.InvToken})
			c.Abort()
			return
		}

		if err := validateJWTClaims(claims, jwtCfg); err != nil {
			common.ErrorResponse(c, common.JWTError{ErrorMsg: text.InvToken})
			c.Abort()
			return
		}

		c.Set("jwt_claims", claims)
		c.Next()
	}
}

func validateJWTClaims(claims jwt.MapClaims, cfg config.JWTConfig) error {
	subject, err := claims.GetSubject()
	if err != nil || strings.TrimSpace(subject) == "" {
		return errors.New("jwt subject claim is required")
	}

	if cfg.Issuer != "" {
		issuer, err := claims.GetIssuer()
		if err != nil || issuer != cfg.Issuer {
			return errors.New("jwt issuer claim is invalid")
		}
	}

	if cfg.Audience != "" {
		audiences, err := claims.GetAudience()
		if err != nil || !containsAudience(audiences, cfg.Audience) {
			return errors.New("jwt audience claim is invalid")
		}
	}

	if cfg.RequiredRole != "" && !hasRequiredRole(claims["role"], cfg.RequiredRole) && !hasRequiredRole(claims["roles"], cfg.RequiredRole) {
		return errors.New("jwt role claim is invalid")
	}

	return nil
}

func hasRequiredRole(value any, requiredRole string) bool {
	requiredRole = strings.TrimSpace(requiredRole)
	if requiredRole == "" {
		return true
	}

	switch typed := value.(type) {
	case string:
		for _, item := range strings.Split(typed, ",") {
			if strings.EqualFold(strings.TrimSpace(item), requiredRole) {
				return true
			}
		}
	case []string:
		for _, item := range typed {
			if strings.EqualFold(strings.TrimSpace(item), requiredRole) {
				return true
			}
		}
	case []any:
		for _, item := range typed {
			if role, ok := item.(string); ok && strings.EqualFold(strings.TrimSpace(role), requiredRole) {
				return true
			}
		}
	}

	return false
}

func containsAudience(audiences jwt.ClaimStrings, requiredAudience string) bool {
	requiredAudience = strings.TrimSpace(requiredAudience)
	if requiredAudience == "" {
		return true
	}

	for _, audience := range audiences {
		if audience == requiredAudience {
			return true
		}
	}

	return false
}
