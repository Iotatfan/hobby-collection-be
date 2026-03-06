package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iotatfan/hobby-collection-be/internal/config"
	"github.com/iotatfan/hobby-collection-be/internal/helper"
	"github.com/iotatfan/hobby-collection-be/internal/text"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helper.ErrorResponse(c, helper.JWTError{ErrorMsg: text.NoAuth})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" || tokenString == authHeader {
			helper.ErrorResponse(c, helper.JWTError{ErrorMsg: text.NoAuth})
			c.Abort()
			return
		}

		secret := config.GetConfig().JWT.Secret
		if secret == "" {
			helper.ErrorResponse(c, helper.ServiceError{ErrorMsg: "jwt secret is not configured", Code: http.StatusInternalServerError})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, helper.JWTError{ErrorMsg: text.InvToken}
			}
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			helper.ErrorResponse(c, helper.JWTError{ErrorMsg: text.InvToken})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			c.Set("jwt_claims", claims)
		}

		c.Next()
	}
}
