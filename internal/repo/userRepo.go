package repo

import (
	"cspirt/internal/models"
	"time"
)

type UserRepository interface {
	DeleteUser(id int, user models.SafeUser) error
	AddUser(user models.User) error
	SaveUser(user models.SafeUser) error
	GetAllUsers() ([]models.SafeUser, error)
	GetUserByLogin(login string) (*models.User, error)
	GetUsersByClass(class string) ([]models.SafeUser, error)
	GetUserByID(id int) (*models.User, error)
	SaveRefreshToken(userID int, token string, expiresAt time.Time) error
	GetRefreshToken(token string) (*models.RefreshToken, error)
	DeleteRefreshToken(token string) error
}
