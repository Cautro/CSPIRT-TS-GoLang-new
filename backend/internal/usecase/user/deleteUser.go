package usecase

import (
	ratingModels "cspirt/internal/domain/rating"
	"errors"
)

var ErrPermissionDenied = errors.New("permission denied")

func (s *UsersUsecase) DeleteUserHandlerService(id int) error {
	user, err := s.userRepo.GetUserByID(id); if err != nil { return err }
	check := s.checkUserRole(user.Login, string(ratingModels.RoleOwner)); if check != nil { return check }

	err = s.userRepo.DeleteUser(id)
	if err == errors.New("user not found") {
		return errors.New("user not found")
	} else if err != nil {
		return err
	}

	return nil
}

func (s *UsersUsecase) checkUserRole(login string, roles ...string) error {
	user, err := s.userRepo.GetUserByLogin(login)
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
