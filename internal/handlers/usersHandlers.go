package handlers

import (
	"cspirt/internal/storage"
	sr "cspirt/internal/service/users"
	// "cspirt/internal/repo"
	"net/http"
	"github.com/gin-gonic/gin"
	"cspirt/internal/models"
)

func GetUsersHandler(s *storage.Storage) gin.HandlerFunc {
	userService := sr.NewUsersService(s, s.Secret)
	return userService.GetUsersHandlerService()
}

func AddUserHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userService := sr.NewUsersService(s, s.Secret)
		if err := userService.AddUserHandlerService(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User added successfully"})
	}
}
