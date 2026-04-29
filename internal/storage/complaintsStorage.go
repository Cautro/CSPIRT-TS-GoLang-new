package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
)

func (s *Storage) AddComplaints(login string, complaint models.Complaint, user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
		INSERT INTO complaints
		(TargetID, AuthorID, Content, CreatedAt)
		VALUES (?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		complaint.TargetID,
		complaint.AuthorID,
		complaint.Content,
		complaint.CreatedAt,
	)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_complaint",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to insert complaint: " + err.Error(),
		})
	}

	return err
}