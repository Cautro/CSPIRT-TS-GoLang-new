package handlers

import (
	"cspirt/internal/storage"
	"log/slog"
	"github.com/gin-gonic/gin"
)

func GetRatingsHandler(s *storage.Storage) gin.HandlerFunc {
	return func (c *gin.Context)  {
		login := c.GetString("Login")
		if login == "" {
			slog.Error("Invalid login or token")
			return 
		}

		user, err := s.GetUserByLogin(login)
		if err != nil {
			slog.Error("error", err)
			return
		} 

		c.JSON(200, gin.H{"Rating": user.Rating})
	}
}