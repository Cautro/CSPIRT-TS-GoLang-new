package service

import (
	"context"
	"errors"
	"time"

	"cspirt/pkg/logger"
)

const (
	maxLoginAttempts   = 5
	loginLockoutWindow = 15 * time.Minute
)

var ErrTooManyLoginAttempts = errors.New("too many login attempts, try again later")

func loginRateLimitKey(login string) string {
	return "ratelimit:login:" + login
}

func (s *AuthUsecase) checkLoginRateLimit(ctx context.Context, login string) bool {
	if s.cache == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := s.cache.Increment(ctx, loginRateLimitKey(login), loginLockoutWindow)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "login_rate_limit",
			Login:   login,
			Message: "redis unavailable, failing open: " + err.Error(),
		})
		return false
	}

	return count > maxLoginAttempts
}

func (s *AuthUsecase) resetLoginRateLimit(ctx context.Context, login string) {
	if s.cache == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.cache.Delete(ctx, loginRateLimitKey(login)); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "login_rate_limit",
			Login:   login,
			Message: "failed to reset counter: " + err.Error(),
		})
	}
}
