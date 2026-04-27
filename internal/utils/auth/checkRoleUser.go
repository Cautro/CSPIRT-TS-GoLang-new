package utils

import (
	"log/slog"
	"errors"
	"cspirt/internal/models"
)

type UserProvider interface {
    GetUserByLogin(login string) (*models.User, error)
}

func CheckUserRole(provider UserProvider, login string, roles ...string) error {
	user, err := provider.GetUserByLogin(login)
	if err != nil {
		return err
	}
	if user == nil {
		slog.Error("user not found")
		return errors.New("user not found")
	}

	for _, role := range roles {
		if user.Role == role {
			return errors.New("access denied")
		}
	}

	slog.Error("user does not have the required role")
	return errors.New("access denied")
}