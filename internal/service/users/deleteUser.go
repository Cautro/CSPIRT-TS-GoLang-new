package users

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	utils "cspirt/internal/utils/auth"
	"errors"
)

func (s *UsersService) DeleteUserHandlerService(id int, u models.User) error {
	checkRole, err := utils.CheckUserRole(s.users, u.Login, string(models.RoleAdmin), string(models.RoleOwner))
		if err != nil || !checkRole {
			writeLog(logger.LogEntry{
				Level:   "info",
				Action:  "delete_user",
				Login:   u.Login,
				Role:    u.Role,
				Class:   u.Class,
				Message: "User without need roles trying to delete user",
			})
			return errors.New("You dont have permisions for thats action")
		}

	safeUser := utils.UserToSafeUser(u)

	err = s.users.DeleteUser(id, *safeUser)

	if err != nil {
		return errors.New("Server error")
	}

	writeLog(logger.LogEntry{
		Level: "info",
		Action: "deleting_note",
		Message: "Deleted note by " + u.Name + " " + u.LastName + "with role " + u.Role,
		Role: u.Role,
		Class: u.Class,
		
	})
	return nil
}
