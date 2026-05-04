package users

import (
	"cspirt/internal/logger"
	"cspirt/internal/users/models"
	utils "cspirt/internal/utils"
	"errors"
)

func (s *UsersService) DeleteUserHandlerService(id int, u models.User) error {
	safeUser := utils.UserToSafeUser(u)

	err := s.users.DeleteUser(id, *safeUser)
	if err == errors.New("user not found") {
		return errors.New("user not found")
	} else if err != nil {
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_user",
		Message: "Deleted user by " + u.Name + " " + u.LastName + " with role " + u.Role,
		Role:    u.Role,
		Class:   u.Class,
	})
	return nil
}
