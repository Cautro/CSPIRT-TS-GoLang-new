package utils

import (
	"cspirt/internal/logger"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Login string `json:"Login"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, errMessage := accessTokenFromRequest(c)
		if errMessage != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMessage})
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "auth_middleware",
				Message: strings.ToLower(errMessage),
			})
			c.Abort()
			return
		}

		claims := &Claims{}

		tok, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.WriteSafe(logger.LogEntry{
					Level:   "info",
					Action:  "auth_middleware",
					Message: "unexpected signing method",
				})
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !tok.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			message := "invalid or expired token"
			if err != nil {
				message += ": " + err.Error()
			}
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "auth_middleware",
				Message: message,
			})
			c.Abort()
			return
		}

		if claims.Login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "auth_middleware",
				Message: "token login is empty",
			})
			c.Abort()
			return
		}

		c.Set("Login", claims.Login)
		c.Next()
	}
}

func accessTokenFromRequest(c *gin.Context) (string, string) {
	auth := c.GetHeader("Authorization")
	if auth != "" {
		parts := strings.Fields(auth)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", "Invalid Authorization header"
		}

		return parts[1], ""
	}

	tokenString, err := c.Cookie(AccessTokenCookieName)
	if err != nil || strings.TrimSpace(tokenString) == "" {
		return "", "Authorization header or access token cookie missing"
	}

	return tokenString, ""
}
