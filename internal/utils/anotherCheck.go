package utils

import (
	"cspirt/internal/logger"
	"cspirt/internal/users/models"
	ratingModels "cspirt/internal/rating/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserGetter interface {
    GetUserByLogin(login string) (*models.User, error)
}

func IsSystemRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "owner":
		return true
	default:
		return false
	}
}

func AuthenticatedUser(c *gin.Context, s UserGetter, action string) (*models.User, bool) {
	login := c.GetString("Login")
	if login == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  action,
			Message: "invalid login or token",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}

	user, err := s.GetUserByLogin(login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return nil, false
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}

	return user, true
}

func CanManageClasses(role string) bool {
	return strings.EqualFold(role, string(ratingModels.RoleAdmin)) ||
		strings.EqualFold(role, string(ratingModels.RoleOwner))
}

func CanReadClass(user *models.User, classID int) bool {
	if CanManageClasses(user.Role) {
		return true
	}

	return user.ClassID == classID
}