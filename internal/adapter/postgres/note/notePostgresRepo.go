package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cspirt/internal/domain/note/repo"
	models "cspirt/internal/domain/user"
	"cspirt/pkg/logger"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.NoteRepository {
	return &postgresRepository{db: db}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNote(s rowScanner) (models.Note, error) {
	var (
		n                models.Note
		createdAt        any
		moderateAt       sql.NullTime
		moderatorID      sql.NullInt64
		moderationStatus sql.NullString
	)

	err := s.Scan(
		&n.ID,
		&n.TargetID,
		&n.TargetName,
		&n.AuthorID,
		&n.AuthorName,
		&n.Content,
		&createdAt,
		&moderateAt,
		&moderatorID,
		&moderationStatus,
	)
	if err != nil {
		return models.Note{}, err
	}

	parsedTime, err := parseEventTime(createdAt)
	if err != nil {
		return models.Note{}, err
	}
	n.CreatedAt = parsedTime

	if moderateAt.Valid {
		n.ModerateAt = moderateAt.Time
	}
	if moderatorID.Valid {
		n.ModeratorId = int(moderatorID.Int64)
	}
	if moderationStatus.Valid {
		n.ModerationStatus = moderationStatus.String
	}

	return n, nil
}

// AddNote — создаёт заметку со статусом по умолчанию "wait"
func (r *postgresRepository) AddNote(ctx context.Context, login string, note models.Note, user models.SafeUser) error {
	note.Content = strings.TrimSpace(note.Content)
	if note.TargetID <= 0 || note.AuthorID <= 0 {
		return errors.New("target and author are required")
	}
	if note.Content == "" {
		return errors.New("content is required")
	}

	var moderatorID any
	if note.ModeratorId > 0 {
		moderatorID = note.ModeratorId
	}

	var moderateAt any
	if !note.ModerateAt.IsZero() {
		moderateAt = note.ModerateAt
	}

	createdAt := note.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	moderationStatus := note.ModerationStatus
	if moderationStatus == "" {
		moderationStatus = string(models.WaitStatus) // "wait"
	}

	query := `
		INSERT INTO notes
		(TargetID, AuthorID, TargetName, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		note.TargetID,
		note.AuthorID,
		note.TargetName,
		note.AuthorName,
		note.Content,
		createdAt,
		moderateAt,
		moderatorID,
		moderationStatus,
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

// UpdateModerationStatus — сменяет статус модерации (поддерживает: wait, cancel, deleted, success)
func (r *postgresRepository) UpdateModerationStatus(ctx context.Context, id int, status string, moderatorID int, user models.SafeUser) error {
	validStatuses := map[string]bool{
		string(models.WaitStatus):    true, // "wait"
		string(models.CancelStatus):  true, // "cancel"
		string(models.DeleteStatus):  true, // "deleted"
		string(models.SuccessStatus): true, // "success"
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid moderation status: %s", status)
	}

	query := `
		UPDATE notes
		SET ModerationStatus = $1, ModerateAt = $2, ModeratorId = $3
		WHERE Id = $4
	`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), moderatorID, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "update_moderation_status",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to update moderation status: " + err.Error(),
		})
		return err
	}

	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("note not found")
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "update_moderation_status",
		Login:   user.Login,
		Role:    user.Role,
		Message: fmt.Sprintf("note %d status updated to %s", id, status),
	})

	return nil
}

// DeleteNote — мягкое удаление (проставляет статус "deleted")
func (r *postgresRepository) DeleteNote(ctx context.Context, id int, user models.SafeUser) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_note",
		Message: "deleting note",
		Login:   user.Login,
		Role:    user.Role,
	})

	query := `
		UPDATE notes
		SET ModerationStatus = $1, ModerateAt = $2
		WHERE Id = $3
	`
	result, err := r.db.ExecContext(ctx, query, string(models.DeleteStatus), time.Now(), id)
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
		Message: "note soft deleted",
	})
	return nil
}

// GetNotesByModerationStatus — получает заметки с нужным статусом модерации
func (r *postgresRepository) GetNotesByModerationStatus(ctx context.Context, status string) ([]models.Note, error) {
	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM notes
		WHERE ModerationStatus = $1
		ORDER BY Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_notes_by_status",
			Message: "failed to query notes by status: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *postgresRepository) GetAllNotes(ctx context.Context) ([]models.Note, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_notes",
		Message: "getting all notes",
	})

	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM notes
		ORDER BY Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
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
		n, err := scanNote(rows)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_notes",
				Message: "failed to scan note: " + err.Error(),
			})
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *postgresRepository) GetNotesByUserId(ctx context.Context, userID int) ([]models.Note, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_note_by_user_id",
		Message: "getting needed note by user id",
	})

	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM notes
		WHERE TargetID = $1
		ORDER BY Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
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
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *postgresRepository) GetNotesByClassID(ctx context.Context, classID int) ([]models.Note, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	query := `
		SELECT n.Id, n.TargetID, n.TargetName, n.AuthorID, n.AuthorName, n.Content, n.CreatedAt, n.ModerateAt, n.ModeratorId, n.ModerationStatus
		FROM notes n
		JOIN users u ON u.Id = n.TargetID
		WHERE u.ClassID = $1
		ORDER BY n.Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, classID)
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
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *postgresRepository) GetNoteByID(ctx context.Context, id int) ([]models.Note, error) {
	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM notes
		WHERE Id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
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