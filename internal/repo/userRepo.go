package repo

import "cspirt/internal/models"

type UserRepository interface {
	DeleteUser(user models.User) error
	AddUser(user models.User) error
	SaveUser(user models.User) error
	ReadUsers() ([]models.User, error)
	GetAllUsers() ([]models.User, error)
	GetUserByLogin(login string) (*models.User, error)
}