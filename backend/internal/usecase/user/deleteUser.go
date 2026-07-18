package usecase

import (
	ratingModels "cspirt/internal/domain/rating"
	"errors"
	"context"
)

var ErrPermissionDenied = errors.New("permission denied")

func (s *UsersUsecase) DeleteUserHandlerService(ctx context.Context, id int, login string) error {
	check := s.checkUserRole(ctx, login, string(ratingModels.RoleOwner)); if check != nil { return check }

	// Resolve the target's login before deletion so we can drop its /me cache.
	s.invalidateUserByID(ctx, id)

	err := s.userRepo.DeleteUser(ctx, id)
	if err == errors.New("user not found") {
		return errors.New("user not found")
	} else if err != nil {
		return err
	}

	return nil
}

func (s *UsersUsecase) checkUserRole(ctx context.Context, login string, roles ...string) error {
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if user.Role == role {
			return nil
		}
	}

	return ErrPermissionDenied
}
