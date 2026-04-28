package rating

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"cspirt/internal/repo"
	u "cspirt/internal/utils/auth"
	"errors"
)

type RatingsService struct {
	users     repo.UserRepository
	jwtSecret string
}

type MyRatingResponce struct {
	Login string
}

type RatingResponceResult struct {
	Rating int
}

func NewRatingsService(users repo.UserRepository, jwtSecret string) *RatingsService {
	return &RatingsService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (s *RatingsService) UpdateRating(login string, in *models.RatingInput) error {
	targetUser, err := s.users.GetUserByLogin(in.TargetLogin)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "update_rating",
			Login:   login,
			Message: "failed to retrieve target user: " + err.Error(),
		})
		return err
	}
	if targetUser == nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "update_rating",
			Login:   login,
			Message: "target user not found: " + in.TargetLogin,
		})
		return errors.New("target user not found")
	}

	user, err := s.users.GetUserByLogin(login)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "update_rating",
			Login:   login,
			Message: "failed to retrieve current user: " + err.Error(),
		})
		return err
	}
	if user == nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "update_rating",
			Login:   login,
			Message: "current user not found",
		})
		return errors.New("user not found")
	}

	check, err := u.CheckUserRole(s.users, login, string(models.RoleAdmin), string(models.RoleOwner))
	if err != nil || !check {
		return errors.New("only admins and owners can update ratings")
	}

	targetUser.Rating += in.Rating

	if targetUser.Rating < 0 {
		targetUser.Rating = 0
	} else if targetUser.Rating > 5000 {
		targetUser.Rating = 5000
	}

	if err := s.users.SaveUser(*targetUser); err != nil {
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "update_rating",
		Login:   login,
		Role:    user.Role,
		Message: "rating updated for user: " + in.TargetLogin,
	})

	return nil
}
