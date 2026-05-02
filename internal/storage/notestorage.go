package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"errors"
	"strings"
	"time"
)

func (s *Storage) AddNote(login string, note models.Note, user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	note.Content = strings.TrimSpace(note.Content)
	if note.TargetID <= 0 || note.AuthorID <= 0 {
		return errors.New("target and author are required")
	}
	if note.Content == "" {
		return errors.New("content is required")
	}
	if note.CreatedAt == "" {
		note.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	query := `
		INSERT INTO notes
		(TargetID, AuthorID, TargetName, AuthorName, Content, CreatedAt)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		note.TargetID,
		note.AuthorID,
		note.TargetName,
		note.AuthorName,
		note.Content,
		note.CreatedAt,
	)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_note",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to insert note: " + err.Error(),
		})
	}

	return err
}

func (s *Storage) DeleteNote(id int, user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_note",
		Message: "deleting note",
		Login:   user.Login,
		Role:    user.Role,
	})

	query := `DELETE FROM notes WHERE Id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "delete_note",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to delete note: " + err.Error(),
		})
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("note not found")
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_note",
		Login:   user.Login,
		Role:    user.Role,
		Message: "note deleted",
	})
	return err
}

func (s *Storage) GetAllNotes() ([]models.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_notes",
		Message: "getting all notes",
	})

	rows, err := s.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM notes
	`)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_notes",
			Message: "failed to query notes: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)
	for rows.Next() {
		var n models.Note

		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&n.CreatedAt,
		); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_notes",
				Message: "failed to scan note: " + err.Error(),
			})
			return nil, err
		}

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_notes",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return notes, nil
}

func (s *Storage) GetNotesByUserId(User_id int) ([]models.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_note_by_user_id",
		Message: "getting needed note by user id",
	})

	rows, err := s.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM notes
		WHERE TargetID = ?
	`, User_id)

	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_note_by_user_id",
			Message: "failed to query notes: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)

	for rows.Next() {
		var n models.Note

		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&n.CreatedAt,
		); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_note_by_user_id",
				Message: "Server error: " + err.Error(),
			})
			return []models.Note{}, err
		}

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_note_by_user_id",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return notes, nil
}

func (s *Storage) GetNotesByClassID(classID int) ([]models.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	if err := s.syncAllClassesLocked(); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT n.Id, n.TargetID, n.TargetName, n.AuthorID, n.AuthorName, n.Content, n.CreatedAt
		FROM notes n
		JOIN users u ON u.Id = n.TargetID
		WHERE u.ClassID = ?
		ORDER BY n.Id DESC
	`, classID)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_notes_by_class",
			Message: "failed to query notes by class: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)

	for rows.Next() {
		var n models.Note

		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&n.CreatedAt,
		); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_notes_by_class",
				Message: "failed to scan note: " + err.Error(),
			})
			return nil, err
		}

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_notes_by_class",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return notes, nil
}

func (s *Storage) GetNoteByID(id int) ([]models.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM notes
		WHERE Id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)
	for rows.Next() {
		var n models.Note
		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&n.CreatedAt,
		); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}