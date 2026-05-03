package handlers

import (
	eventModels "cspirt/internal/events/models"
	"cspirt/internal/logger"
	"cspirt/internal/storage"
	"cspirt/internal/users/models"
	sr "cspirt/internal/users/service"
	u "cspirt/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUsersHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		currentUser, err := s.GetUserByLogin(login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current user"})
			return
		}

		if currentUser == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userIDStr := c.Query("id")
		if userIDStr == "" {
			userService := sr.NewUsersService(s, s.Secret)

			users, err := userService.GetUsersHandlerService()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
				return
			}

			c.JSON(http.StatusOK, users)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil || userID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		needUser, err := s.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}

		if needUser == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		SafeNeedUser := u.UserToSafeUser(*needUser)

		notes, err := s.NotesRepo.GetNotesByUserId(SafeNeedUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes"})
			return
		}

		complaints, err := s.ComplaintsRepo.GetComplaintsByUserId(SafeNeedUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve complaints"})
			return
		}

		var classTeacher *models.SafeUser

		if SafeNeedUser.ClassID > 0 {
			classTeacher, err = s.ClassRepo.GetClassTeacherByID(SafeNeedUser.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class teacher"})
				return
			}
		}

		events, err := s.EventsRepo.GetEventsByUserID(SafeNeedUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
			return
		}

		answerResponse := models.UserWithFullInfo{
			User:         SafeNeedUser,
			Notes:        notes,
			Complaints:   complaints,
			ClassTeacher: classTeacher,
			Events:       events,
		}

		c.JSON(http.StatusOK, answerResponse)
	}
}

func AddUserHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "add_user",
				Message: "invalid login or token",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		targetUser, err := s.GetUserByLogin(login)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		if targetUser == nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "add_user",
				Login:   login,
				Message: "user not found",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if !u.CanManageClasses(targetUser.Role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "add_user",
				Login:   login,
				Role:    targetUser.Role,
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userService := sr.NewUsersService(s, s.Secret)
		if err := userService.AddUserHandlerService(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   login,
			Role:    targetUser.Role,
			Message: "user added successfully: " + user.Login,
		})

		c.JSON(http.StatusOK, gin.H{"message": "User added successfully"})
	}
}

func DeleteUserHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.Param("id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		if login == "" {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "delete_user",
				Message: "invalid login or token",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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
				Action:  "delete_user",
				Login:   login,
				Message: "user not found",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if !u.CanManageClasses(foundUser.Role) {
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

		userService := sr.NewUsersService(s, s.Secret)

		if err := userService.DeleteUserHandlerService(idInt, *foundUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "delete_user",
			Login:   login,
			Role:    foundUser.Role,
			Message: "user deleted successfully",
		})

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

func GetMeHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "get_me",
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
				Action:  "get_me",
				Login:   login,
				Message: "user not found",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		notes, err := s.NotesRepo.GetNotesByUserId(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes"})
			return
		}

		complaints, err := s.ComplaintsRepo.GetComplaintsByUserId(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve complaints"})
			return
		}

		resp := u.UserToSafeUser(*user)

		var classTeacher *models.SafeUser

		if user.ClassID > 0 {
			classTeacher, err = s.ClassRepo.GetClassTeacherByID(user.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class teacher"})
				return
			}
		}

		answerResponse := models.UserWithFullInfo{
			User:         resp,
			Notes:        notes,
			Complaints:   complaints,
			ClassTeacher: classTeacher,
			Events:       []eventModels.Event{},
		}

		c.JSON(http.StatusOK, answerResponse)
	}
}
