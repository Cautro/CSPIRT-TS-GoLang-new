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
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "auth_middleware",
				Message: "authorization header missing",
			})
			c.Abort()
			return
		}

		parts := strings.Fields(auth)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "auth_middleware",
				Message: "invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		tok, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				writeLog(logger.LogEntry{
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
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "auth_middleware",
				Message: message,
			})
			c.Abort()
			return
		}

		if claims.Login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			writeLog(logger.LogEntry{
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
