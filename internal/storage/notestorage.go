package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
)

func (s *Storage) AddNote(login string, note models.Note, user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
		INSERT INTO notes
		(TargetID, AuthorID, Content, CreatedAt)
		VALUES (?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		note.TargetID,
		note.AuthorID,
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

func (s *Storage) GetAllNotes() ([]models.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_notes",
		Message: "getting all notes",
	})

	rows, err := s.db.Query(`
		SELECT Id, TargetID, AuthorID, Content, CreatedAt
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
			&n.AuthorID,
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