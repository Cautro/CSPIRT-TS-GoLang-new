package service

import (
	"cspirt/internal/models"
	"cspirt/internal/repo"
	"cspirt/internal/utils/auth"
	"log/slog"
)

type AuthService struct {
	users     repo.UserRepository
	jwtSecret string
	log 	  *slog.Logger
}

type LoginResult struct {
	Token string `json:"token"`
}

func NewAuthService(users repo.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
		log:       slog.Default(),
	}
}

func (s *AuthService) Login(in models.LoginInput) (LoginResult, error) {
	user, err := s.users.GetUserByLogin(in.Login)
	if err != nil {
		return LoginResult{}, err
	}

	if !utils.CheckPasswordHash(in.Password, user.Password) {
		return LoginResult{}, nil
	}

	if user == nil {
		return LoginResult{}, nil
	}

	token, err := utils.GenerateToken(in.Login, s.jwtSecret)
	if err != nil {
		s.log.Error("failed to generate JWT", "login", in.Login, "error", err)
		return LoginResult{}, err
	}

	return LoginResult{Token: token}, nil
}