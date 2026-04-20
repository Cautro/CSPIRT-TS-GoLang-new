package repo

import "cspirt/internal/models"

type UserRepository interface {
	DeleteUser(id int) error
	AddUser(user models.User) error
	SaveUser(user models.User) error
	ReadUsers() ([]models.User, error)
	GetAllUsers() ([]models.User, error)
	GetUserByLogin(login string) (*models.User, error)
}