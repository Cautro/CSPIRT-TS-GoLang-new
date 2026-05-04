package storage

import (
	"cspirt/internal/logger"
	ratingModels "cspirt/internal/rating/models"
	userModels "cspirt/internal/users/models"
	utils "cspirt/internal/utils"
	"errors"
	"strings"
)

func (s *Storage) AddComplaint(login string, complaint userModels.Complaint, user userModels.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if login == "" {
		return errors.New("invalid login or token")
	}


	complaint.Content = strings.TrimSpace(complaint.Content)
	if complaint.TargetID <= 0 || complaint.AuthorID <= 0 {
		return errors.New("target and author are required")
	}
	if complaint.Content == "" {
		return errors.New("content is required")
	}

	targetUser, err := s.getUserByIDLocked(complaint.TargetID)
	if err != nil {
		return err
	}
	if targetUser == nil {
		return errors.New("target user not found")
	}

	if utils.IsSystemRole(targetUser.Role) {
		return errors.New("system users cannot be complained about")
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "adding new complaint",
	})

	if targetUser.Login == login {
		return errors.New("users cannot complain about themselves")
	}
	if complaint.TargetID == complaint.AuthorID {
		return errors.New("author and target cannot be the same")
	}
	if len(complaint.Content) > 1000 {
		return errors.New("content exceeds maximum length of 1000 characters")
	}
	
	query := `
		INSERT INTO complaints
		(TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(
		query,
		complaint.TargetID,
		complaint.TargetName,
		complaint.AuthorID,
		complaint.AuthorName,
		complaint.Content,
		complaint.CreatedAt,
	)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_complaint",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to insert complaint: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "complaint inserted",
	})

	return nil
}

func (s *Storage) DeleteComplaint(id int, user userModels.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "deleting complaint",
	})

	query := `DELETE FROM complaints WHERE Id = ?`

	check, err := s.hasUserRoleLocked(user.Login, string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner))
	if err != nil {
		return err
	}
	if !check {
		return errors.New("only admins can delete complaints")
	}

	result, err := s.db.Exec(query, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "complaint deleted",
	})

	return nil
}

func (s *Storage) GetAllComplaints() ([]userModels.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_complaints",
		Message: "getting all complaints",
	})

	rows, err := s.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM complaints
	`)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_complaints",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]userModels.Complaint, 0)

	for rows.Next() {
		var complaint userModels.Complaint
		var createdAt interface{}

		if err := rows.Scan(
			&complaint.ID,
			&complaint.TargetID,
			&complaint.TargetName,
			&complaint.AuthorID,
			&complaint.AuthorName,
			&complaint.Content,
			&createdAt,
		); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_complaints",
				Message: "failed to scan complaint: " + err.Error(),
			})
			return nil, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		complaint.CreatedAt = parsedTime

		complaints = append(complaints, complaint)
	}

	if err := rows.Err(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_complaints",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}

func (s *Storage) GetComplaintByID(id int) ([]userModels.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
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
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaint_by_id",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]userModels.Complaint, 0)

	for rows.Next() {
		var complaint userModels.Complaint
		var createdAt interface{}

		if err := rows.Scan(
			&complaint.ID,
			&complaint.TargetID,
			&complaint.TargetName,
			&complaint.AuthorID,
			&complaint.AuthorName,
			&complaint.Content,
			&createdAt,
		); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_complaint_by_id",
				Message: "failed to scan complaint: " + err.Error(),
			})
			return nil, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		complaint.CreatedAt = parsedTime

		complaints = append(complaints, complaint)
	}

	if err := rows.Err(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaint_by_id",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}

func (s *Storage) GetComplaintsByUserId(User_id int) ([]userModels.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
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
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaint_by_user_id",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]userModels.Complaint, 0)

	for rows.Next() {
		var c userModels.Complaint
		var createdAt interface{}

		if err := rows.Scan(
			&c.ID,
			&c.TargetID,
			&c.TargetName,
			&c.AuthorID,
			&c.AuthorName,
			&c.Content,
			&createdAt,
		); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_complaint_by_user_id",
				Message: "Server error: " + err.Error(),
			})
			return []userModels.Complaint{}, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		c.CreatedAt = parsedTime

		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_note_by_user_id",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}

func (s *Storage) GetComplaintsByClassID(classID int) ([]userModels.Complaint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	rows, err := s.db.Query(`
		SELECT c.Id, c.TargetID, c.TargetName, c.AuthorID, c.AuthorName, c.Content, c.CreatedAt
		FROM complaints c
		JOIN users u ON u.Id = c.TargetID
		WHERE u.ClassID = ?
		ORDER BY c.Id DESC
	`, classID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaints_by_class",
			Message: "failed to query complaints by class: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]userModels.Complaint, 0)

	for rows.Next() {
		var c userModels.Complaint
		var createdAt interface{}

		if err := rows.Scan(
			&c.ID,
			&c.TargetID,
			&c.TargetName,
			&c.AuthorID,
			&c.AuthorName,
			&c.Content,
			&createdAt,
		); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_complaints_by_class",
				Message: "failed to scan complaint: " + err.Error(),
			})
			return nil, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		c.CreatedAt = parsedTime

		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaints_by_class",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return complaints, nil
}
