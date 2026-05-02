package handlers

import (
	"cspirt/internal/logger"
	"cspirt/internal/auth/models"
	sr "cspirt/internal/auth/service/auth"
	"cspirt/internal/storage"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func LoginHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "login",
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		authService := sr.NewAuthService(s, s.Secret)
		result, err := authService.Login(input)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if result.Token == "" {
			c.JSON(401, gin.H{"error": "Invalid login or password"})
			return
		}

		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(
			"refresh_token",
			result.RefreshToken,
			3600*24*7,
			"/api/refresh",
			"",
			os.Getenv("COOKIE_SECURE") == "1",
			true,
		)

		c.JSON(200, gin.H{
			"accessToken": result.Token,
		})

		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "login",
			Login:   input.Login,
			Message: "login successful",
		})
	}
}

func RefreshHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(401, gin.H{"error": "refresh token missing"})
			return
		}

		authService := sr.NewAuthService(s, s.Secret)

		result, err := authService.Refresh(refreshToken)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"token": result.Token,
		})
	}
}
