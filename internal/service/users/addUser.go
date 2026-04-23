package users

import (
	"cspirt/internal/models"
)

func (s *UsersService) AddUserHandlerService(user models.User) error {
	err := s.users.AddUser(user)
	if err != nil {
		return err
	}
	s.log.Info("User added successfully", "login", user.Login)
	return nil
}
