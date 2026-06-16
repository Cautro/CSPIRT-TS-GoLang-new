package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/users/models"
	"database/sql"
	"time"
)

func (s *Storage) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, userID, token, expiresAt)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "save_refresh_token",
			Message: "failed to save refresh token: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "save_refresh_token",
		Message: "refresh token saved",
	})
	return nil
}

func (s *Storage) GetRefreshToken(token string) (*models.RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.db.QueryRow(`
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = ?
	`, token)

	var rt models.RefreshToken

	err := row.Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_refresh_token",
			Message: "failed to get refresh token: " + err.Error(),
		})
		return nil, err
	}

	return &rt, nil
}

func (s *Storage) DeleteRefreshToken(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		DELETE FROM refresh_tokens
		WHERE token = ?
	`, token)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_refresh_token",
			Message: "failed to delete refresh token: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_refresh_token",
		Message: "refresh token deleted",
	})
	return nil
}
