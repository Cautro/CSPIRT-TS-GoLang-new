package handlers

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"cspirt/internal/service/rating"
	"cspirt/internal/storage"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRatingsHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "get_rating",
				Message: "invalid login or token",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		user, err := s.GetUserByLogin(login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}
		if user == nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "get_rating",
				Login:   login,
				Message: "user not found",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(200, gin.H{"Rating": user.Rating})
	}
}

func UpdateRatingsHandler(rs *rating.RatingsService, s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.RatingInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "update_rating",
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		login := c.GetString("Login")
		if login == "" {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "update_rating",
				Message: "invalid login or token",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if login == input.TargetLogin {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "update_rating",
				Login:   login,
				Message: "cannot rate yourself",
			})
			c.JSON(400, gin.H{"error": "Cannot rate yourself"})
			return
		}

		if err := rs.UpdateRating(login, &input); err != nil {
			c.JSON(500, gin.H{"error": "Failed to update rating"})
			return
		}

		user, err := s.GetUserByLogin(login)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve updated user"})
			return
		}
		if user == nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "update_rating",
				Login:   login,
				Message: "user not found after rating update",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(200, gin.H{"message": "Rating updated successfully", "new_rating": user.Rating})
	}
}
