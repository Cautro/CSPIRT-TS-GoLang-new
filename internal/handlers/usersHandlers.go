package handlers

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	sr "cspirt/internal/service/users"
	"cspirt/internal/storage"
	u "cspirt/internal/utils/auth"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUsersHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userService := sr.NewUsersService(s, s.Secret)

		user, err := s.GetUserByLogin(c.GetString("Login"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if strings.EqualFold(user.Role, string(models.RoleAdmin)) ||
			strings.EqualFold(user.Role, string(models.RoleUser)) ||
			strings.EqualFold(user.Role, string(models.RoleHelper)) {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "get_users",
				Login:   user.Login,
				Role:    user.Role,
				Message: "users, helpers and admins view the list of users in the them class",
			})

			users, err := userService.GetUsersByClassIDHandlerService(user.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
				return
			}

			c.JSON(http.StatusOK, users)
			return
		}

		users, err := userService.GetUsersHandlerService()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}

		c.JSON(http.StatusOK, users)
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

		check, err := u.CheckUserRole(s, login, string(models.RoleAdmin), string(models.RoleOwner))
		if err != nil || !check {
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

		note, err := s.NotesRepo.GetNotesByUserId(user.ID)
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

		AnswerResponse := models.UserWithFullInfo{
			User:       resp,
			Notes:      note,
			Complaints: complaints,
		}

		c.JSON(http.StatusOK, AnswerResponse)
	}
}
