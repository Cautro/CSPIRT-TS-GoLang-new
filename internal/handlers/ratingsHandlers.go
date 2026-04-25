package handlers

import (
	"cspirt/internal/storage"
	"log/slog"

	"github.com/gin-gonic/gin"
	"cspirt/internal/service/rating"
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

func UpdateRatingsHandler(rs *rating.RatingsService, s *storage.Storage) gin.HandlerFunc {
	return func (c *gin.Context)  {
		var input struct {
			Rating int `json:"rating"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		login := c.GetString("Login")
		if login == "" {
			slog.Error("Invalid login or token")
			return 
		}

		err := rs.UpdateRating(login, input.Rating)
		if err != nil {
			slog.Error("error", err)
			c.JSON(500, gin.H{"error": "Failed to update rating"})
			return
		}

		user, err := s.GetUserByLogin(login)
		if err != nil {
			slog.Error("error", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve updated user"})
			return
		}

		c.JSON(200, gin.H{"message": "Rating updated successfully", "new_rating": user.Rating})
	}
}