package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"errors"
	"strings"
	"time"
)

func (s *Storage) AddComplaint(login string, complaint models.Complaint, user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	complaint.Content = strings.TrimSpace(complaint.Content)
	if complaint.TargetID <= 0 || complaint.AuthorID <= 0 {
		return errors.New("target and author are required")
	}
	if complaint.Content == "" {
		return errors.New("content is required")
	}
	if complaint.CreatedAt == "" {
		complaint.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	query := `
		INSERT INTO complaints
		(TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		complaint.TargetID,
		complaint.TargetName,
		complaint.AuthorID,
		complaint.AuthorName,
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

	result, err := s.db.Exec(query, id)
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
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("complaint not found")
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
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
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
			&complaint.TargetName,
			&complaint.AuthorID,
			&complaint.AuthorName,
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
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
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
			&complaint.TargetName,
			&complaint.AuthorID,
			&complaint.AuthorName,
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

func (s *Storage) GetComplaintsByUserId(User_id int) ([]models.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_complaint_by_user_id",
		Message: "getting needed complaint by user id",
	})

	rows, err := s.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM complaints
		WHERE TargetID = ?
	`, User_id)

	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaint_by_user_id",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var c models.Complaint

		if err := rows.Scan(
			&c.ID,
			&c.TargetID,
			&c.TargetName,
			&c.AuthorID,
			&c.AuthorName,
			&c.Content,
			&c.CreatedAt,
		); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_complaint_by_user_id",
				Message: "Server error: " + err.Error(),
			})
			return []models.Complaint{}, err
		}

		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_note_by_user_id",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}

func (s *Storage) GetComplaintsByClassID(classID int) ([]models.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	if err := s.syncAllClassesLocked(); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT c.Id, c.TargetID, c.TargetName, c.AuthorID, c.AuthorName, c.Content, c.CreatedAt
		FROM complaints c
		JOIN users u ON u.Id = c.TargetID
		WHERE u.ClassID = ?
		ORDER BY c.Id DESC
	`, classID)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaints_by_class",
			Message: "failed to query complaints by class: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var c models.Complaint

		if err := rows.Scan(
			&c.ID,
			&c.TargetID,
			&c.TargetName,
			&c.AuthorID,
			&c.AuthorName,
			&c.Content,
			&c.CreatedAt,
		); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_complaints_by_class",
				Message: "failed to scan complaint: " + err.Error(),
			})
			return nil, err
		}

		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaints_by_class",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}