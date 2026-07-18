package usecase

import (
	"context"
	entity "cspirt/internal/domain/user"
)

func (s *UsersUsecase) AddUserHandlerService(ctx context.Context, user entity.User) error {
	return s.userRepo.AddUser(ctx, user)
}
