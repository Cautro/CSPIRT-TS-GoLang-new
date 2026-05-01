package handlers

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	complaintsservice "cspirt/internal/service/complaintservice"
	"cspirt/internal/storage"
	u "cspirt/internal/utils/auth"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetComplaintsHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		complaints := complaintsservice.NewComplaintsService(s, s.Secret)
		result, err := complaints.GetAllComplaints()
		if err != nil || result == nil {
			c.JSON(500, gin.H{"error": "Server error"})
			return
		}

		c.JSON(200, gin.H{"All_complaints": result})
	}
}

func AddcomplaintHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		complaints := complaintsservice.NewComplaintsService(s, s.Secret)
		user, err := s.GetUserByLogin(login)
		if err != nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "server error: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Server error"})
			return
		}
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var in models.AddNewComplaintResponse
		if err := c.ShouldBindJSON(&in); err != nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		needUser := &models.SafeUser{
			ID:       user.ID,
			Name:     user.Name,
			LastName: user.LastName,
			FullName: user.FullName,
			Login:    user.Login,
			Rating:   user.Rating,
			Role:     user.Role,
			Class:    user.Class,
		}

		if err := complaints.AddNewComplaint(login, &in, needUser); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		writeLog(logger.LogEntry{
			Level:   "info",
			Login:   login,
			Class:   user.Class,
			Role:    user.Role,
			Action:  "added_complaint",
			Message: "Added new complaint",
		})
		c.JSON(200, gin.H{"message": "Complaint added"})
	}
}

func DeletecomplaintHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.Param("id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		foundUser, err := s.GetUserByLogin(login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}

		if foundUser == nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "delete_note",
				Login:   login,
				Message: "note not found",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "Server error"})
			return
		}

		checkRole, err := u.CheckUserRole(s, login, string(models.RoleAdmin), string(models.RoleOwner))
		if err != nil || !checkRole {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "delete_user",
				Login:   login,
				Role:    foundUser.Role,
				Class:   foundUser.Class,
				Message: "User without need roles trying to delete user",
			})
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
			return
		}

		needUser := u.UserToSafeUser(*foundUser)
		complaints := complaintsservice.NewComplaintsService(s, s.Secret)
		if err := complaints.DeleteComplaint(idInt, *needUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Complaint deleted"})
	}
}
