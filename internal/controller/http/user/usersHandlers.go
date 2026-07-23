package handlers

import (
	"cspirt/pkg/logger"
	permissionService "cspirt/internal/controller/permission/usecase"
	ratMod "cspirt/internal/domain/rating"
	models "cspirt/internal/domain/user"
	authUsecase "cspirt/internal/usecase/auth"
	sr "cspirt/internal/usecase/user"

	"net/http"
	"os"
	"strconv"
	"context"
	"time"

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
func GetUsersHandler(userService *sr.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		currentUser, err := userService.GetUserByLogin(ctx, login)
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
			users, err := userService.GetUsersHandlerService(ctx)
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

		fullUserInfo, err := userService.GetFullUserInfo(ctx, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve full user info"})
			return
		}

		c.JSON(http.StatusOK, fullUserInfo)
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
func UpdateAvatarHandler(userService *sr.UsersUsecase) gin.HandlerFunc {
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

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err = userService.UpdateAvatar(ctx, in, id)
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
func LogoutHandler(userService *sr.UsersUsecase, authService *authUsecase.AuthUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if login := c.GetString("Login"); login != "" {
			userService.InvalidateUserFullInfo(ctx, login)
		}

		if token, err := c.Cookie("refresh_token"); err == nil {
			if err := userService.DeleteRefreshToken(ctx, token); err != nil {
				logger.WriteSafe(logger.LogEntry{
					Level:   "error",
					Action:  "logout",
					Message: "failed to delete refresh token from db: " + err.Error(),
				})
			}
		}

		if token, err := c.Cookie("access_token"); err == nil {
			if err := authService.Logout(token); err != nil {
				logger.WriteSafe(logger.LogEntry{
					Level:   "error",
					Action:  "logout",
					Message: "failed to blacklist access token: " + err.Error(),
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
func AddUserHandler(userService *sr.UsersUsecase) gin.HandlerFunc {
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

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		targetUser, err := userService.GetUserByLogin(ctx, login)
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

		if !permissionService.CanManageClasses(targetUser.Role) {
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

		if err := userService.AddUserHandlerService(ctx, user); err != nil {
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
func UpdateUserHandler(userService *sr.UsersUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
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

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		targetUser, err := userService.GetUserByLogin(ctx, login)
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

		check := perm.CheckUserRole(ctx, login, string(ratMod.RoleOwner))
		if check != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "only owner can update user"})
			return
		}

		if err = userService.UpdateUserHandlerService(ctx, idInt, user, login); err != nil {
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
func DeleteUserHandler(userService *sr.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.Param("id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		if login == "" {
			logger.WriteSafe(logger.LogEntry{ Level: "info", Action: "delete_user", Message: "invalid login or token" })
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := userService.DeleteUserHandlerService(ctx, idInt, login); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_user",
			Login:   login,
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
func GetMeHandler(userService *sr.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			logger.WriteSafe(logger.LogEntry{ Level: "info", Action: "get_me", Message: "invalid login or token" })
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		fullInfo, err := userService.GetFullUserInfoByLogin(ctx, login)
		if err != nil {
			c.JSON(403, gin.H{"error": "Bad request"})
			return
		}

		c.JSON(http.StatusOK, fullInfo)
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
func GetStaffHandler(userService *sr.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		staff, err := userService.GetOnlyStaffUsers(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve staff"})
			return
		}

		c.JSON(http.StatusOK, staff)
	}
}
