package service

import (
	// "cspirt/internal/logger"
	"cspirt/internal/domain/auth"
	cacheRepo "cspirt/internal/domain/cache/repo"
	"cspirt/internal/domain/user/repo"
	"cspirt/internal/controller/http/middleware-JWT"
	// "crypto/rand"
	// "strings"
	"errors"
	"strings"
	"time"
)

type AuthUsecase struct {
	users     repo.UserRepository
	cache     cacheRepo.CacheRepository
	jwtSecret string
}

type LoginResult struct {
	Token        string
	RefreshToken string
}

// NewAuthService builds the auth usecase. cache may be nil, in which case
// rate limiting and token revocation are silently disabled (see
// internal/adapter/redis/README.md).
func NewAuthService(users repo.UserRepository, jwtSecret string, cache cacheRepo.CacheRepository) *AuthUsecase {
	return &AuthUsecase{
		users:     users,
		cache:     cache,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthUsecase) Login(in entity.LoginInput) (LoginResult, error) {
	in.Login = strings.TrimSpace(in.Login)
	if in.Login == "" || in.Password == "" {
		return LoginResult{}, nil
	}

	if s.checkLoginRateLimit(in.Login) {
		return LoginResult{}, ErrTooManyLoginAttempts
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

	accessToken, err := utils.GenerateToken(user.ID, in.Login, user.Role, s.jwtSecret)
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

	s.resetLoginRateLimit(in.Login)

	return LoginResult{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthUsecase) Refresh(refreshToken string) (LoginResult, error) {
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

	token, err := utils.GenerateToken(user.ID, user.Login, user.Role, s.jwtSecret)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token}, nil
}
