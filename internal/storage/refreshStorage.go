package storage

import (
	"cspirt/internal/models"
	"time"
	"database/sql"
)

func (s *Storage) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, userID, token, expiresAt)

	return err
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

	return err
}