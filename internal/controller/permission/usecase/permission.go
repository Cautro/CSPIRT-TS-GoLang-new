package service

import (
	userModels "cspirt/internal/domain/user"
	userRepo "cspirt/internal/domain/user/repo"
	ratingModels "cspirt/internal/domain/rating"
	"cspirt/pkg/logger"
	"errors"
	"net/http"
	"strings"
	"context"

	"github.com/gin-gonic/gin"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrAccessDenied = errors.New("access denied")
)

type Usecase struct {
	users userRepo.UserRepository
}

func New(users userRepo.UserRepository) *Usecase {
	return &Usecase{users: users}
}

func (s *Usecase) CheckUserRole(ctx context.Context, login string, roles ...string) error {
	user, err := s.users.GetUserByLogin(ctx, login)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "check_user_role",
			Login:   login,
			Message: "failed to retrieve user: " + err.Error(),
		})
		return err
	}

	if user == nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "check_user_role",
			Login:   login,
			Message: "user not found",
		})
		return ErrUserNotFound
	}

	userRole := strings.ToLower(strings.TrimSpace(user.Role))

	for _, role := range roles {
		allowedRole := strings.ToLower(strings.TrimSpace(role))

		if userRole == allowedRole {
			return nil
		}
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "check_user_role",
		Login:   login,
		Role:    user.Role,
		Message: "user does not have the required role",
	})

	return ErrAccessDenied
}

func (s *Usecase) CheckPublicRole(ctx context.Context, login string) error {
	user, err := s.users.GetUserByLogin(ctx, login)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "check_public_role",
			Login:   login,
			Message: "failed to retrieve user: " + err.Error(),
		})
		return err
	}

	if user == nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "check_public_role",
			Login:   login,
			Message: "user not found",
		})
		return ErrUserNotFound
	}

	return nil
}

func (s *Usecase) AuthenticatedUser(ctx context.Context, c *gin.Context, action string) (*userModels.User, bool) {
	login := c.GetString("Login")
	if login == "" {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  action,
			Message: "invalid login or token",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}

	user, err := s.users.GetUserByLogin(ctx, login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return nil, false
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}

	return user, true
}

func CanManageClasses(role string) bool {
	return strings.EqualFold(role, string(ratingModels.RoleAdmin)) ||
		strings.EqualFold(role, string(ratingModels.RoleOwner))
}

func CanReadClass(user *userModels.User, classID int) bool {
	if CanManageClasses(user.Role) {
		return true
	}

	return user.ClassID == classID
}
