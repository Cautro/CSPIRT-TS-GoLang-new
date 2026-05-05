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

var (
	ErrUserNotFound = errors.New("user not found")
	ErrAccessDenied = errors.New("access denied")
)

func CheckUserRole(provider UserProvider, login string, roles ...string) error {
	user, err := provider.GetUserByLogin(login)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "check_user_role",
			Login:   login,
			Message: "failed to retrieve user: " + err.Error(),
		})
		return err
	}

	if user == nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "check_user_role",
			Login:   login,
			Message: "user not found",
		})
		return ErrUserNotFound
	}

	userRole := strings.ToLower(strings.TrimSpace(user.Role))

	for _, role := range roles {
		allowedRole := strings.ToLower(strings.TrimSpace(role))

		if userRole == allowedRole {
			return nil
		}
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "check_user_role",
		Login:   login,
		Role:    user.Role,
		Message: "user does not have the required role",
	})

	return ErrAccessDenied
}
