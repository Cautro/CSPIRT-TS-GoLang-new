package usecase

import (
	entity "cspirt/internal/domain/user"
)

func (s *UsersUsecase) AddUserHandlerService(user entity.User) error {
	return s.userRepo.AddUser(user)
}
