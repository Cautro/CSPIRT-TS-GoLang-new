package repo

import (
	models "cspirt/internal/domain/user"
	"cspirt/internal/domain/note/repo"
	"cspirt/pkg/logger"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.NoteRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) AddNote(login string, note models.Note, user models.SafeUser) error {
	note.Content = strings.TrimSpace(note.Content)
	if note.TargetID <= 0 || note.AuthorID <= 0 {
		return errors.New("target and author are required")
	}
	if note.Content == "" {
		return errors.New("content is required")
	}

	query := `
		INSERT INTO notes
		(TargetID, AuthorID, TargetName, AuthorName, Content, CreatedAt)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		query,
		note.TargetID,
		note.AuthorID,
		note.TargetName,
		note.AuthorName,
		note.Content,
		note.CreatedAt,
	)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_note",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to insert note: " + err.Error(),
		})
	}

	return err
}

func (r *postgresRepository) DeleteNote(id int, user models.SafeUser) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_note",
		Message: "deleting note",
		Login:   user.Login,
		Role:    user.Role,
	})

	query := `DELETE FROM notes WHERE Id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_note",
		Login:   user.Login,
		Role:    user.Role,
		Message: "note deleted",
	})
	return err
}

func (r *postgresRepository) GetAllNotes() ([]models.Note, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_notes",
		Message: "getting all notes",
	})

	rows, err := r.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM notes
	`)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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
		var createdAt interface{}

		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&createdAt,
		); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_notes",
				Message: "failed to scan note: " + err.Error(),
			})
			return nil, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = parsedTime

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_notes",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return notes, nil
}

func (r *postgresRepository) GetNotesByUserId(User_id int) ([]models.Note, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_note_by_user_id",
		Message: "getting needed note by user id",
	})

	rows, err := r.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM notes
		WHERE TargetID = $1
	`, User_id)

	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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
		var createdAt interface{}

		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&createdAt,
		); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_note_by_user_id",
				Message: "Server error: " + err.Error(),
			})
			return []models.Note{}, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = parsedTime

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_note_by_user_id",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return notes, nil
}

func (r *postgresRepository) GetNotesByClassID(classID int) ([]models.Note, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	rows, err := r.db.Query(`
		SELECT n.Id, n.TargetID, n.TargetName, n.AuthorID, n.AuthorName, n.Content, n.CreatedAt
		FROM notes n
		JOIN users u ON u.Id = n.TargetID
		WHERE u.ClassID = $1
		ORDER BY n.Id DESC
	`, classID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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
		var createdAt interface{}

		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&createdAt,
		); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_notes_by_class",
				Message: "failed to scan note: " + err.Error(),
			})
			return nil, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = parsedTime

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_notes_by_class",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return notes, nil
}

func (r *postgresRepository) GetNoteByID(id int) ([]models.Note, error) {
	rows, err := r.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM notes
		WHERE Id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)
	for rows.Next() {
		var n models.Note
		var createdAt interface{}
		if err := rows.Scan(
			&n.ID,
			&n.TargetID,
			&n.TargetName,
			&n.AuthorID,
			&n.AuthorName,
			&n.Content,
			&createdAt,
		); err != nil {
			return nil, err
		}
		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = parsedTime
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func parseEventTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		return parseEventTimeString(v)
	case []byte:
		return parseEventTimeString(string(v))
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported event time type %T", value)
	}
}

func parseEventTimeString(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if monotonicIndex := strings.Index(value, " m="); monotonicIndex >= 0 {
		value = value[:monotonicIndex]
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid event time %q", value)
}
