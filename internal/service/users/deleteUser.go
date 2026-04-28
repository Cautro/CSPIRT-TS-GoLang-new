package users

import (
	"cspirt/internal/models"
)

func (s *UsersService) DeleteUserHandlerService(user models.User) error {
	return s.users.DeleteUser(user)
}
