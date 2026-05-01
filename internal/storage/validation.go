package storage

import (
	"cspirt/internal/models"
	"errors"
	"strings"
)

func normalizeLogin(login string) string {
	return strings.TrimSpace(login)
}

func normalizeClassName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

func normalizeRole(role string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case strings.ToLower(string(models.RoleUser)):
		return string(models.RoleUser), nil
	case strings.ToLower(string(models.RoleHelper)):
		return string(models.RoleHelper), nil
	case strings.ToLower(string(models.RoleAdmin)):
		return string(models.RoleAdmin), nil
	case strings.ToLower(string(models.RoleOwner)):
		return string(models.RoleOwner), nil
	default:
		return "", errors.New("invalid role")
	}
}

func isTeacherCandidate(role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	return role == strings.ToLower(string(models.RoleAdmin)) ||
		role == strings.ToLower(string(models.RoleOwner)) ||
		role == strings.ToLower(string(models.RoleHelper))
}

func trimUserInput(user *models.User) {
	user.Name = strings.TrimSpace(user.Name)
	user.LastName = strings.TrimSpace(user.LastName)
	user.Login = normalizeLogin(user.Login)
	user.Class = normalizeClassName(user.Class)

	for i := range user.FullName {
		user.FullName[i].Name = strings.TrimSpace(user.FullName[i].Name)
		user.FullName[i].LastName = strings.TrimSpace(user.FullName[i].LastName)
	}
}
