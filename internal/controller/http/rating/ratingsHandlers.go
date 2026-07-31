package handlers

import (
	"context"
	models "cspirt/internal/domain/rating"
	userModels "cspirt/internal/domain/user"
	rating "cspirt/internal/usecase/rating"
	usersvc "cspirt/internal/usecase/user"
	"cspirt/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetRatingsHandler returns the current user's rating.
// @Summary Get rating
// @Description Returns the authenticated user's current rating.
// @Tags rating
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/rating [get]
func GetRatingsHandler(users *usersvc.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		login := c.GetString("Login")
		if login == "" {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "get_rating",
				Message: "invalid login or token",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		user, err := users.GetUserByLogin(ctx, login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}
		if user == nil {
			logger.WriteSafe(logger.LogEntry{
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

// UpdateRatingsHandler updates a user's rating.
// @Summary Update rating
// @Description Updates the rating of a target user using the request body.
// @Tags rating
// @Accept json
// @Produce json
// @Param request body models.RatingInput true "Rating update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/rating/update [patch]
func UpdateRatingsHandler(rs *rating.RatingsUsecase, users *usersvc.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.RatingInput
		if err := c.ShouldBindJSON(&input); err != nil {
			logger.WriteSafe(logger.LogEntry{ Level: "info", Action: "update_rating", Message: "invalid input: " + err.Error() })
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		login := c.GetString("Login")
		if login == "" {
			logger.WriteSafe(logger.LogEntry{ Level: "info", Action: "update_rating", Message: "invalid login or token" })
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if login == input.TargetLogin {
			logger.WriteSafe(logger.LogEntry{ Level: "info", Action: "update_rating", Login: login, Message: "cannot rate yourself" })
			c.JSON(400, gin.H{"error": "Cannot rate yourself"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		user, err := users.GetUserByLogin(ctx, login)
		if err != nil || user == nil {
			c.JSON(500, gin.H{"error": "Login invalid or user dont found"})
			return
		}

		needUser := &userModels.SafeUser{
			ID:       user.ID,
			Name:     user.Name,
			LastName: user.LastName,
			FullName: user.FullName,
			Login:    user.Login,
			Role:     user.Role,
			Class:    user.Class,
			ClassID:  user.ClassID,
			Rating:   user.Rating,
		}

		if err := rs.UpdateRating(ctx, login, &input, needUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		targetUser, err := users.GetUserByLogin(ctx, input.TargetLogin)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve updated target user"})
			return
		}
		if targetUser == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Target user not found"})
			return
		}

		c.JSON(200, gin.H{
			"message":    "Rating updated successfully",
			"target":     targetUser.Login,
			"new_rating": targetUser.Rating,
		})
	}
}

func UpdateClassRatingHandler(usecase *rating.RatingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context)  {
		userIdStr := c.GetString("Id")
		if userIdStr == "" {
			c.JSON(401, gin.H{"error":"Unauthorized"})
			return 
		}
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		var input models.ClassRatingInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		if err := usecase.UpdateClassRating(c.Request.Context(), input.ClassId, userId, input.Delta); err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		c.JSON(200, gin.H{"status":"ok"})
	}
}