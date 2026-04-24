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

