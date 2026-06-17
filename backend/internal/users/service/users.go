package users

import (
	"cspirt/internal/users/models"
	"cspirt/internal/users/repo"
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

func (s *UsersService) GetUsersByClassIDHandlerService(classID int) ([]models.SafeUser, error) {
	users, err := s.users.GetUsersByClassID(classID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UsersService) UpdateAvatar(input models.UpdateAvatarRequest, id int) error {
	err := s.users.UpdateAvatar(input, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersService) UpdateUserHandlerService(userID int, user models.SafeUser, login string) (error) {
	NeedUser := models.UpdateUserRequest{
		Name: &user.Name,
		LastName: &user.LastName,
		Avatar: &user.Avatar,
		FullName: &user.FullName,
		Login: &user.Login,
		Rating: &user.Rating,
		Role: &user.Role,
		Class: &user.Class,
		ClassID: &user.ClassID,
	}
	err := s.users.UpdateUser(userID, NeedUser, login)
	if err != nil {
		return err
	}
	return nil
}