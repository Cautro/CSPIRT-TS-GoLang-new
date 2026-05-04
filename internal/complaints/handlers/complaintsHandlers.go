package handlers

import (
	srClass "cspirt/internal/class/service"
	complaintModels "cspirt/internal/complaints/models"
	complaintsservice "cspirt/internal/complaints/service"
	"cspirt/internal/logger"
	ratingModels "cspirt/internal/rating/models"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"
	u "cspirt/internal/utils"
	"errors"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetComplaintsHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := u.AuthenticatedUser(c, s, "get_notes")
		if !ok {
			return
		}

		err := u.CheckUserRole(s, user.Login, string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner), string(ratingModels.RoleHelper))
		if err != nil {
			if errors.Is(err, u.ErrAccessDenied) || errors.Is(err, u.ErrUserNotFound) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		complaintService := complaintsservice.NewComplaintsService(s, s.Secret)

		classIDStr := c.Query("class")
		if classIDStr != "" {
			classID, err := strconv.Atoi(classIDStr)
			if err != nil || classID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
				return
			}

			classService := srClass.NewClassService(s, s.Secret)

			class, err := classService.GetClassByID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
				return
			}
			if class == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
				return
			}

			if !u.CanReadClass(user, classID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this class"})
				return
			}

			result, err := complaintService.GetComplaintsByClassID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"Complaints": result})
			return
		}

		if !u.CanManageClasses(user.Role) {
			if user.ClassID <= 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "User has no class"})
				return
			}

			result, err := complaintService.GetComplaintsByClassID(user.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"Complaints": result})
			return
		}

		result, err := complaintService.GetAllComplaints()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"All_Complaints": result})
	}
}

func AddcomplaintHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		complaintService := complaintsservice.NewComplaintsService(s, s.Secret)
		user, err := s.GetUserByLogin(login)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
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

		var in complaintModels.AddNewComplaintResponse
		if err := c.ShouldBindJSON(&in); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		needUser := &userModels.SafeUser{
			ID:       user.ID,
			Name:     user.Name,
			LastName: user.LastName,
			FullName: user.FullName,
			Login:    user.Login,
			Rating:   user.Rating,
			Role:     user.Role,
			Class:    user.Class,
			ClassID:  user.ClassID,
		}

		if err := complaintService.AddNewComplaint(login, &in, needUser); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
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
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_note",
				Login:   login,
				Message: "note not found",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "Server error"})
			return
		}

		if !u.CanManageClasses(foundUser.Role) {
			logger.WriteSafe(logger.LogEntry{
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
