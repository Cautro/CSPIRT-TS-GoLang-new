package repo

import (
	"time"
	entity "cspirt/internal/domain/user"
)

type UserRepository interface {
	DeleteUser(id int) error
	AddUser(user entity.User) error
	SaveUser(user entity.SafeUser) error
	GetAllUsers() ([]entity.SafeUser, error)
	GetUserByLogin(login string) (*entity.User, error)
	GetUsersByClassID(classID int) ([]entity.SafeUser, error)
	GetUserByID(id int) (*entity.User, error)
	SaveRefreshToken(userID int, token string, expiresAt time.Time) error
	GetRefreshToken(token string) (*entity.RefreshToken, error)
	DeleteRefreshToken(token string) error
	UpdateUser(id int, req entity.UpdateUserRequest, login string) error
	UpdateAvatar(input entity.UpdateAvatarRequest, id int) error
	GetOnlyStaffUsers() ([]entity.SafeUser, error)
}

