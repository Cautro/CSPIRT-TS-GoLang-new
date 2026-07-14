package utils

import (
	models "cspirt/internal/domain/user"
)

func UserToSafeUser(u models.User) *models.SafeUser {
	needUser := &models.SafeUser{
		ID:       u.ID,
		Avatar:   u.Avatar,
		Name:     u.Name,
		LastName: u.LastName,
		FullName: u.FullName,
		Login:    u.Login,
		Rating:   u.Rating,
		Role:     u.Role,
		Class:    u.Class,
		ClassID:  u.ClassID,
	}

	return needUser
}
