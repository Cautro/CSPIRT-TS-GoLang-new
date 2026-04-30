package utils

import (
	"cspirt/internal/models"
)

func UserToSafeUser(u models.User) *models.SafeUser {
	needUser := &models.SafeUser{
		ID: u.ID,
		Name: u.Name,
		LastName: u.LastName,
		FullName: u.FullName,
		Login: u.Login,
		Rating: u.Rating,
		Role: u.Role,
		Class: u.Class,
	}

	return needUser
} 