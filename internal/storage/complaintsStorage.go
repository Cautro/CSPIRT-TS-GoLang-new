package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
)

func (s *Storage) AddComplaint(login string, complaint models.Complaint, user models.SafeUser) error {
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
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "add_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "complaint inserted",
	})

	return nil
}

func (s *Storage) DeleteComplaint(id int, user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "deleting complaint",
	})

	query := `DELETE FROM complaints WHERE Id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "delete_complaint",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to delete complaint: " + err.Error(),
		})
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "complaint deleted",
	})

	return nil
}

func (s *Storage) GetAllComplaints() ([]models.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_complaints",
		Message: "getting all complaints",
	})

	rows, err := s.db.Query(`
		SELECT Id, TargetID, AuthorID, Content, CreatedAt
		FROM complaints
	`)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_complaints",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var complaint models.Complaint

		if err := rows.Scan(
			&complaint.ID,
			&complaint.TargetID,
			&complaint.AuthorID,
			&complaint.Content,
			&complaint.CreatedAt,
		); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_complaints",
				Message: "failed to scan complaint: " + err.Error(),
			})
			return nil, err
		}

		complaints = append(complaints, complaint)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_complaints",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}

func (s *Storage) GetComplaintByID(id int) ([]models.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_complaint_by_id",
		Message: "getting needed complaint by id",
	})

	rows, err := s.db.Query(`
		SELECT Id, TargetID, AuthorID, Content, CreatedAt
		FROM complaints
		WHERE Id = ?
	`, id)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaint_by_id",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var complaint models.Complaint

		if err := rows.Scan(
			&complaint.ID,
			&complaint.TargetID,
			&complaint.AuthorID,
			&complaint.Content,
			&complaint.CreatedAt,
		); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_complaint_by_id",
				Message: "failed to scan complaint: " + err.Error(),
			})
			return nil, err
		}

		complaints = append(complaints, complaint)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaint_by_id",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}