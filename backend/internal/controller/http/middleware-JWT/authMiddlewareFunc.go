package utils

import (
	entity "cspirt/internal/domain/auth"
	cacheRepo "cspirt/internal/domain/cache/repo"
	"cspirt/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Id int `json:"Id"`
	Login string `json:"Login"`
	Role string `json:"Role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates the JWT on every request. cache may be nil, in
// which case revoked-token checking is silently disabled (see
// internal/adapter/redis/README.md).
func AuthMiddleware(jwtSecret string, cache cacheRepo.CacheRepository) gin.HandlerFunc {
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

		if isTokenBlacklisted(cache, tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "auth_middleware",
				Login:   claims.Login,
				Message: "token revoked",
			})
			c.Abort()
			return
		}

		c.Set("Login", claims.Login)
		c.Next()
	}
}

// isTokenBlacklisted reports whether tokenString was revoked via logout.
// Fails open (returns false) on a nil cache or a Redis error so an outage
// never locks every user out - see internal/adapter/redis/README.md.
func isTokenBlacklisted(cache cacheRepo.CacheRepository, tokenString string) bool {
	if cache == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	exists, err := cache.Exists(ctx, entity.BlacklistTokenKey(tokenString))
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "auth_middleware",
			Message: "redis unavailable, failing open: " + err.Error(),
		})
		return false
	}

	return exists
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
