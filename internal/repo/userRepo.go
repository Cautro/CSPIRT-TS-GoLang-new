package repo

import "cspirt/internal/models"

type UserRepository interface {
	DeleteUser(user models.User) error
	AddUser(user models.User) error
	SaveUser(user models.SafeUser) error
	GetAllUsers() ([]models.SafeUser, error)
	GetUserByLogin(login string) (*models.User, error)
	GetUsersByClass(class string) ([]models.SafeUser, error)
}
