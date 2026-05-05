package storage

import (
	"cspirt/internal/users/models"
	ratingModels "cspirt/internal/rating/models"
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
	case strings.ToLower(string(ratingModels.RoleAdmin)):
		return string(ratingModels.RoleAdmin), nil
	case strings.ToLower(string(ratingModels.RoleUser)):
		return string(ratingModels.RoleUser), nil
	case strings.ToLower(string(ratingModels.RoleHelper)):
		return string(ratingModels.RoleHelper), nil
	case strings.ToLower(string(ratingModels.RoleOwner)):
		return string(ratingModels.RoleOwner), nil
	default:
		return "", errors.New("invalid role")
	}
}

func isTeacherCandidate(role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	return role == strings.ToLower(string(ratingModels.RoleAdmin)) ||
		role == strings.ToLower(string(ratingModels.RoleOwner)) ||
		role == strings.ToLower(string(ratingModels.RoleHelper))
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
