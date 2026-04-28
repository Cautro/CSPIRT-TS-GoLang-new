package service

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"cspirt/internal/repo"
	"cspirt/internal/utils/auth"
)

type AuthService struct {
	users     repo.UserRepository
	jwtSecret string
}

type LoginResult struct {
	Token string `json:"token"`
}

func NewAuthService(users repo.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Login(in models.LoginInput) (LoginResult, error) {
	user, err := s.users.GetUserByLogin(in.Login)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "login",
			Login:   in.Login,
			Message: "failed to retrieve user: " + err.Error(),
		})
		return LoginResult{}, err
	}

	if user == nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "login",
			Login:   in.Login,
			Message: "user not found",
		})
		return LoginResult{}, nil
	}

	if !utils.CheckPasswordHash(in.Password, user.Password) {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "login",
			Login:   in.Login,
			Role:    user.Role,
			Message: "invalid login or password",
		})
		return LoginResult{}, nil
	}

	token, err := utils.GenerateToken(in.Login, s.jwtSecret)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "login",
			Login:   in.Login,
			Role:    user.Role,
			Message: "failed to generate token: " + err.Error(),
		})
		return LoginResult{}, err
	}

	return LoginResult{Token: token}, nil
}
