package repo

import (
	"time"
	"context"
	entity "cspirt/internal/domain/user"
)

type UserRepository interface {
	DeleteUser(ctx context.Context, id int) error
	AddUser(ctx context.Context, user entity.User) error
	SaveUser(ctx context.Context, user entity.SafeUser) error
	GetAllUsers(ctx context.Context) ([]entity.SafeUser, error)
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	GetUsersByClassID(ctx context.Context, classID int) ([]entity.SafeUser, error)
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	SaveRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	UpdateUser(ctx context.Context, id int, req entity.UpdateUserRequest, login string) error
	UpdateAvatar(ctx context.Context, input entity.UpdateAvatarRequest, id int) error
	GetOnlyStaffUsers(ctx context.Context) ([]entity.SafeUser, error)
}

