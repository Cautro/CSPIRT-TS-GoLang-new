package usecase

import (
	"cspirt/pkg/logger"
	permissionService "cspirt/internal/controller/permission/usecase"
	models "cspirt/internal/domain/rating"
	"cspirt/internal/domain/rating/repo"
	userModels "cspirt/internal/domain/user"
	usersRepo "cspirt/internal/domain/user/repo"
	"errors"
)

type RatingsUsecase struct {
	users     usersRepo.UserRepository
	rating    repo.RatingRepository
}

type MyRatingResponce struct {
	Login string
}

type RatingResponceResult struct {
	Rating int
}

func NewRatingsUsecase(rating repo.RatingRepository, users usersRepo.UserRepository) *RatingsUsecase {
	return &RatingsUsecase{
		rating:    rating,
		users:     users,
	}
}

func (s *RatingsUsecase) UpdateRating(login string, in *models.RatingInput, user *userModels.SafeUser) error {
	if in == nil {
		return errors.New("invalid input")
	}
	if in.Rating < -5000 || in.Rating > 5000 {
		return errors.New("rating change must be between -5000 and 5000")
	}

	targetUser, err := s.users.GetUserByLogin(in.TargetLogin)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "update_rating",
			Login:   login,
			Message: "failed to retrieve target user: " + err.Error(),
		})
		return err
	}
	if targetUser == nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_rating",
			Login:   login,
			Message: "target user not found: " + in.TargetLogin,
		})
		return errors.New("target user not found")
	}

	if user == nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_rating",
			Login:   login,
			Message: "current user not found",
		})
		return errors.New("user not found")
	}

	if !permissionService.CanManageClasses(user.Role) {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_rating",
			Role:    user.Role,
			Class:   user.Class,
			Login:   login,
			Message: "User or helper try to update rating",
		})
		return errors.New("only admins and owners can update ratings")
	}

	targetUser.Rating += in.Rating

	if targetUser.Rating < 0 {
		targetUser.Rating = 0
	} else if targetUser.Rating > 5000 {
		targetUser.Rating = 5000
	}

	needTargetUser := &userModels.SafeUser{
		ID:       targetUser.ID,
		Name:     targetUser.Name,
		LastName: targetUser.LastName,
		FullName: targetUser.FullName,
		Login:    targetUser.Login,
		Role:     targetUser.Role,
		Class:    targetUser.Class,
		ClassID:  targetUser.ClassID,
		Rating:   targetUser.Rating,
	}

	if err := s.users.SaveUser(*needTargetUser); err != nil {
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "update_rating",
		Login:   login,
		Role:    user.Role,
		Message: "rating updated for user: " + in.TargetLogin + ", reason: " + in.Reason,
	})

	return nil
}
