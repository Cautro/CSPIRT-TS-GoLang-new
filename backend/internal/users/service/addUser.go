package users

import (
	"cspirt/internal/users/models"
)

func (s *UsersService) AddUserHandlerService(user models.User) error {
	return s.users.AddUser(user)
}
