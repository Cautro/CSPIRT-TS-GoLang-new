package utils

import (
	"cspirt/internal/storage"

	"strings"
	"log/slog"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Login string `json:"Login"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func (c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			slog.Warn("Authorization header missing")
			c.Abort()
			return
		}

		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
			slog.Warn("Invalid Authorization header")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		tok, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				slog.Warn("Unexpected signing method")
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(storage.Secret), nil
		})
		if err != nil || !tok.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			slog.Warn("Invalid or expired token: %v", err)
			c.Abort()
			return
		}

		c.Set("Login", claims.Login)
		c.Next()
	}
}
