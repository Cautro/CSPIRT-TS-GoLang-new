package handlers

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	sr "cspirt/internal/service/auth"
	"cspirt/internal/storage"
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

		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "login",
			Login:   input.Login,
			Message: "login successful",
		})

		c.JSON(200, gin.H{"token": result.Token})
	}
}
