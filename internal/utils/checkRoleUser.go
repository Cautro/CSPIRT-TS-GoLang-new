package utils

import (
	"cspirt/internal/logger"
	"cspirt/internal/users/models"
	"errors"
	"strings"
)

type UserProvider interface {
	GetUserByLogin(login string) (*models.User, error)
}

func CheckUserRole(provider UserProvider, login string, roles ...string) (bool, error) {
	user, err := provider.GetUserByLogin(login)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "check_user_role",
			Login:   login,
			Message: "failed to retrieve user: " + err.Error(),
		})
		return false, err
	}

	if user == nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "check_user_role",
			Login:   login,
			Message: "user not found",
		})
		return false, errors.New("user not found")
	}

	userRole := strings.ToLower(strings.TrimSpace(user.Role))

	for _, role := range roles {
		allowedRole := strings.ToLower(strings.TrimSpace(role))

		if userRole == allowedRole {
			return true, nil
		}
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "check_user_role",
		Login:   login,
		Role:    user.Role,
		Message: "user does not have the required role",
	})

	return false, errors.New("access denied")
}
