package rating

import (
	"cspirt/internal/repo"
	"log/slog"
)

type RatingsService struct {
	users     repo.UserRepository
	jwtSecret string
	log 	  *slog.Logger
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
		log:       slog.Default(),
	}
}


func (s *RatingsService) UpdateRating(login string, rating int) error {
	user, err := s.users.GetUserByLogin(login)
	if err != nil || user == nil {
		return err
	}

	user.Rating += rating

	if user.Rating < 0 {
		user.Rating = 0
	} else if user.Rating > 5000 {
		user.Rating = 5000
	}

	return s.users.SaveUser(*user)
}
