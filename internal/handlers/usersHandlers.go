package handlers

import (
	"cspirt/internal/storage"
	sr "cspirt/internal/service/users"
	// "cspirt/internal/repo"
	"log/slog"
	"fmt"
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User added successfully"})
	}
}

func DeleteUserHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context)  {
		login := c.GetString("Login")
		var user models.User

		foundUser, err := s.GetUserByLogin(login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}

		if foundUser == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		switch foundUser.Role {
		case "admin":
			// Admins can delete any user
			slog.Info("Admin user deleting a user", "admin", login)
			break
		case "owner":
			// Owners can delete any user
			slog.Info("Owner user deleting a user", "owner", login)
			break
		case "user":
			c.JSON(http.StatusForbidden, gin.H{"error": "Users cannot delete users"})
			return
		case "helper":
			c.JSON(http.StatusForbidden, gin.H{"error": "Helpers cannot delete users"})
			return
		default:
			c.JSON(http.StatusForbidden, gin.H{"error": "Unknown role"})
			return
		}


		if err := c.ShouldBindBodyWithJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userService := sr.NewUsersService(s, s.Secret)
		if err := userService.DeleteUserHandlerService(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

func GetMeHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context)  {
		login := c.GetString("Login")
		fmt.Print(login)
		user, err := s.GetUserByLogin(login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}