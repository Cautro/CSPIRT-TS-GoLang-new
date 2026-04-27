package users

import (
	"cspirt/internal/models"
)

func (s *UsersService) DeleteUserHandlerService(user models.User) error {
	err := s.users.DeleteUser(user)
	
	if err != nil {
		return err
	}

	s.log.Info("User deleted successfully", "login", user.Login)
	return nil
}