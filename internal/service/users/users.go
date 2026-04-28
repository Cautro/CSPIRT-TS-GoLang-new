package users

import (
	"cspirt/internal/models"
	"cspirt/internal/repo"
)

type UsersService struct {
	users repo.UserRepository
}

func NewUsersService(users repo.UserRepository, jwtSecret string) *UsersService {
	return &UsersService{
		users: users,
	}
}

func (s *UsersService) GetUsersHandlerService() ([]models.SafeUser, error) {
	users, err := s.users.GetAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}
