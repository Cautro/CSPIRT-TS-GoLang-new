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

// ErrTooManyLoginAttempts is returned by Login when the per-login failed
// attempt counter (tracked in Redis) has been exceeded.
var ErrTooManyLoginAttempts = errors.New("too many login attempts, try again later")

func loginRateLimitKey(login string) string {
	return "ratelimit:login:" + login
}

// checkLoginRateLimit increments the failed-attempt counter for login and
// reports whether the caller is currently locked out.
//
// If Redis is unavailable this fails open (returns false, nil) so a cache
// outage never blocks legitimate logins - see the "Fail-open" section in
// internal/adapter/redis/README.md.
func (s *AuthUsecase) checkLoginRateLimit(login string) bool {
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

// resetLoginRateLimit clears the failed-attempt counter after a successful login.
func (s *AuthUsecase) resetLoginRateLimit(login string) {
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
