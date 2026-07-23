package service

import (
	"context"
	"time"

	entity "cspirt/internal/domain/auth"
	utils "cspirt/internal/controller/http/middleware-JWT"
	"cspirt/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

// Logout revokes accessToken by adding it to the Redis blacklist for the
// remainder of its natural lifetime, so AuthMiddleware rejects it on the next
// request even though the JWT signature is still valid.
//
// If accessToken is empty, malformed, or Redis is unavailable, Logout logs
// the issue and returns nil - a failed revocation must never prevent the
// user from logging out (see the "Fail-open" section in
// internal/adapter/redis/README.md).
func (s *AuthUsecase) Logout(accessToken string) error {
	if accessToken == "" || s.cache == nil {
		return nil
	}

	claims := &utils.Claims{}
	_, _, err := jwt.NewParser().ParseUnverified(accessToken, claims)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "logout",
			Message: "could not parse access token for blacklisting: " + err.Error(),
		})
		return nil
	}

	var ttl time.Duration
	if claims.ExpiresAt != nil {
		ttl = time.Until(claims.ExpiresAt.Time)
	}
	if ttl <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.cache.Set(ctx, entity.BlacklistTokenKey(accessToken), true, ttl); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "logout",
			Login:   claims.Login,
			Message: "failed to blacklist access token: " + err.Error(),
		})
	}

	return nil
}
