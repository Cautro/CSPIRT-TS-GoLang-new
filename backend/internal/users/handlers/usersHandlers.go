package handlers

import (
	"cspirt/internal/logger"
	ratMod "cspirt/internal/rating/models"
	"cspirt/internal/storage"
	"cspirt/internal/users/models"
	sr "cspirt/internal/users/service"
	u "cspirt/internal/utils"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUsersHandler returns users and detailed user info.
// @Summary List users
// @Description Returns a list of users or full details for a specific user by query parameter.
// @Tags users
// @Produce json
// @Param id query int false "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users [get]
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

// UpdateAvatarHandler updates a user's avatar.
// @Summary Update avatar
// @Description Updates the avatar of a user by query parameter.
// @Tags users
// @Accept json
// @Produce json
// @Param id query int true "User ID"
// @Param request body models.UpdateAvatarRequest true "Avatar payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/user/update/avatar [patch]
func UpdateAvatarHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request, id is empty"})
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		}

		var in models.UpdateAvatarRequest
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		userService := sr.NewUsersService(s, s.Secret)
		err = userService.UpdateAvatar(in, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// LogoutHandler clears the authentication cookies and invalidates the refresh token.
// @Summary Logout
// @Description Logs the user out and clears authentication cookies.
// @Tags users
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/user/logout [patch]
func LogoutHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie("refresh_token"); err == nil {
			if err := s.DeleteRefreshToken(token); err != nil {
				logger.WriteSafe(logger.LogEntry{
					Level:   "error",
					Action:  "logout",
					Message: "failed to delete refresh token from db: " + err.Error(),
				})
			}
		}

		c.SetCookie("access_token", "", -1, "/backend/api", "", cookieSecure(), true)
		c.SetCookie("refresh_token", "", -1, "/backend/api", "", cookieSecure(), true)

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "logout",
			Message: "user logged out successfully",
		})

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "1"
}

// AddUserHandler creates a new user.
// @Summary Create user
// @Description Creates a new user from the request body.
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.User true "User payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/user/add [patch]
func AddUserHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			logger.WriteSafe(logger.LogEntry{
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
			logger.WriteSafe(logger.LogEntry{
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
			logger.WriteSafe(logger.LogEntry{
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

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   login,
			Role:    targetUser.Role,
			Message: "user added successfully: " + user.Login,
		})

		c.JSON(http.StatusOK, gin.H{"message": "User added successfully"})
	}
}

// UpdateUserHandler updates a user's profile.
// @Summary Update user
// @Description Updates a user by query parameter and request body.
// @Tags users
// @Accept json
// @Produce json
// @Param id query int true "User ID"
// @Param request body models.SafeUser true "Updated user payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/user/update [patch]
func UpdateUserHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.Query("id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}
		if login == "" {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_user",
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
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_user",
				Login:   login,
				Message: "user not found",
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var user models.SafeUser

		if err := c.ShouldBindJSON(&user); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_user",
				Login:   login,
				Role:    targetUser.Role,
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		check := u.CheckUserRole(s, login, string(ratMod.RoleOwner))
		if check != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "only owner can update user"})
			return
		}

		userService := sr.NewUsersService(s, s.Secret)
		if err = userService.UpdateUserHandlerService(idInt, user, login); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_user",
			Login:   login,
			Role:    targetUser.Role,
			Message: "user updated successfully: " + user.Login,
		})

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

// DeleteUserHandler deletes a user by ID.
// @Summary Delete user
// @Description Deletes a user by ID.
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/user/delete/{id} [delete]
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
			logger.WriteSafe(logger.LogEntry{
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
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_user",
				Login:   login,
				Message: "user not found",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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

		userService := sr.NewUsersService(s, s.Secret)

		if err := userService.DeleteUserHandlerService(idInt, *foundUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_user",
			Login:   login,
			Role:    foundUser.Role,
			Message: "user deleted successfully",
		})

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

// GetMeHandler returns the current authenticated user profile with related data.
// @Summary Get current user profile
// @Description Returns the current authenticated user and related notes, complaints, events, and class teacher data.
// @Tags users
// @Produce json
// @Success 200 {object} models.UserWithFullInfo
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/me [get]
func GetMeHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			logger.WriteSafe(logger.LogEntry{
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
			logger.WriteSafe(logger.LogEntry{
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

		events, err := s.EventsRepo.GetEventsByUserID(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
			return
		}

		answerResponse := models.UserWithFullInfo{
			User:         resp,
			Notes:        notes,
			Complaints:   complaints,
			ClassTeacher: classTeacher,
			Events:       events,
		}

		c.JSON(http.StatusOK, answerResponse)
	}
}

// GetStaffHandler returns all staff users.
// @Summary List staff users
// @Description Returns the list of staff users.
// @Tags users
// @Produce json
// @Success 200 {array} models.SafeUser
// @Failure 500 {object} map[string]string
// @Router /api/users/get/staff [get]
func GetStaffHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		staff, err := s.GetOnlyStaffUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve staff"})
			return
		}

		c.JSON(http.StatusOK, staff)
	}
}
