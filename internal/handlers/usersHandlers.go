package handlers

import (
	ServiceUsers "cspirt/internal/service/users"
	"github.com/gin-gonic/gin"
	"cspirt/internal/models"
)

func GetUsersHandler(s *ServiceUsers.UsersService) gin.HandlerFunc {
	return s.GetUsersHandler()
}

func AddUserHandler(s *ServiceUsers.UsersService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		if err := s.AddUser(user); err != nil {
			c.JSON(500, gin.H{"error": "Failed to add user"})
			return
		}

		c.JSON(200, gin.H{"message": "User added successfully"})
	}
}
