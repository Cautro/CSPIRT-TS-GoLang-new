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

func (s *RatingsService) UpdateRating(login string, in *models.RatingInput, user *models.SafeUser) error {
	if in == nil {
		return errors.New("invalid input")
	}
	if in.Rating < -5000 || in.Rating > 5000 {
		return errors.New("rating change must be between -5000 and 5000")
	}

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
		writeLog(logger.LogEntry{
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

	needTargetUser := &models.SafeUser{
		ID:       targetUser.ID,
		Name:     targetUser.Name,
		LastName: targetUser.LastName,
		FullName: targetUser.FullName,
		Login:    targetUser.Login,
		Role:     targetUser.Role,
		Class:    targetUser.Class,
		Rating:   targetUser.Rating,
	}

	if err := s.users.SaveUser(*needTargetUser); err != nil {
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "update_rating",
		Login:   login,
		Role:    user.Role,
		Message: "rating updated for user: " + in.TargetLogin + ", reason: " + in.Reason,
	})

	return nil
}
