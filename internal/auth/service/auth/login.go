package service

import (
	// "cspirt/internal/logger"
	"cspirt/internal/auth/models"
	"cspirt/internal/users/repo"
	"cspirt/internal/utils"
	// "crypto/rand"
	// "strings"
	"errors"
	"strings"
	"time"
)

type AuthService struct {
	users     repo.UserRepository
	jwtSecret string
}

type LoginResult struct {
	Token        string
	RefreshToken string
}

func NewAuthService(users repo.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Login(in models.LoginInput) (LoginResult, error) {
	in.Login = strings.TrimSpace(in.Login)
	if in.Login == "" || in.Password == "" {
		return LoginResult{}, nil
	}

	user, err := s.users.GetUserByLogin(in.Login)
	if err != nil {
		return LoginResult{}, err
	}

	if user == nil {
		return LoginResult{}, nil
	}

	if !utils.CheckPasswordHash(in.Password, user.Password) {
		return LoginResult{}, nil
	}

	accessToken, err := utils.GenerateToken(in.Login, s.jwtSecret)
	if err != nil {
		return LoginResult{}, err
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := s.users.SaveRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (LoginResult, error) {
	session, err := s.users.GetRefreshToken(refreshToken)
	if err != nil {
		return LoginResult{}, err
	}

	if session == nil {
		return LoginResult{}, errors.New("invalid refresh token")
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.users.DeleteRefreshToken(refreshToken)
		return LoginResult{}, errors.New("refresh token expired")
	}

	user, err := s.users.GetUserByID(session.UserID)
	if err != nil {
		return LoginResult{}, err
	}

	if user == nil {
		return LoginResult{}, errors.New("user not found")
	}

	token, err := utils.GenerateToken(user.Login, s.jwtSecret)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token}, nil
}
