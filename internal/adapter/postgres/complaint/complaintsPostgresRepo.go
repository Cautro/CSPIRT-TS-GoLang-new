package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cspirt/internal/domain/complaint/repo"
	models "cspirt/internal/domain/user"
	"cspirt/pkg/logger"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.ComplaintRepository {
	return &postgresRepository{db: db}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanComplaint(s rowScanner) (models.Complaint, error) {
	var (
		c                models.Complaint
		createdAt        any
		moderateAt       sql.NullTime
		moderatorID      sql.NullInt64
		moderationStatus sql.NullString
	)

	err := s.Scan(
		&c.ID,
		&c.TargetID,
		&c.TargetName,
		&c.AuthorID,
		&c.AuthorName,
		&c.Content,
		&createdAt,
		&moderateAt,
		&moderatorID,
		&moderationStatus,
	)
	if err != nil {
		return models.Complaint{}, err
	}

	parsedTime, err := parseEventTime(createdAt)
	if err != nil {
		return models.Complaint{}, err
	}
	c.CreatedAt = parsedTime

	if moderateAt.Valid {
		c.ModerateAt = moderateAt.Time
	}
	if moderatorID.Valid {
		c.ModeratorId = int(moderatorID.Int64)
	}
	if moderationStatus.Valid {
		c.ModerationStatus = moderationStatus.String
	}

	return c, nil
}

func (r *postgresRepository) AddComplaint(ctx context.Context, login string, complaint models.Complaint, user models.SafeUser) error {
	complaint.Content = strings.TrimSpace(complaint.Content)
	if complaint.TargetID <= 0 || complaint.AuthorID <= 0 {
		return errors.New("target and author are required")
	}
	if complaint.Content == "" {
		return errors.New("content is required")
	}

	var moderatorID any
	if complaint.ModeratorId > 0 {
		moderatorID = complaint.ModeratorId
	}

	var moderateAt any
	if !complaint.ModerateAt.IsZero() {
		moderateAt = complaint.ModerateAt
	}

	createdAt := complaint.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	moderationStatus := complaint.ModerationStatus
	if moderationStatus == "" {
		moderationStatus = string(models.WaitStatus)
	}

	query := `
		INSERT INTO complaints
		(TargetID, AuthorID, TargetName, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		complaint.TargetID,
		complaint.AuthorID,
		complaint.TargetName,
		complaint.AuthorName,
		complaint.Content,
		createdAt,
		moderateAt,
		moderatorID,
		moderationStatus,
	)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_complaint",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to insert complaint: " + err.Error(),
		})
	}

	return err
}

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
		UPDATE complaints
		SET ModerationStatus = $1, ModerateAt = $2, ModeratorId = $3
		WHERE Id = $4
	`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), moderatorID, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "update_complaint_moderation_status",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to update moderation status: " + err.Error(),
		})
		return err
	}

	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("complaint not found")
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "update_complaint_moderation_status",
		Login:   user.Login,
		Role:    user.Role,
		Message: fmt.Sprintf("complaint %d status updated to %s", id, status),
	})

	return nil
}

func (r *postgresRepository) DeleteComplaint(ctx context.Context, id int, user models.SafeUser) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_complaint",
		Message: "deleting complaint",
		Login:   user.Login,
		Role:    user.Role,
	})

	query := `
		UPDATE complaints
		SET ModerationStatus = $1, ModerateAt = $2
		WHERE Id = $3
	`
	result, err := r.db.ExecContext(ctx, query, string(models.DeleteStatus), time.Now(), id)
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
		Message: "complaint soft deleted",
	})
	return nil
}

func (r *postgresRepository) GetComplaintsByModerationStatus(ctx context.Context, status string) ([]models.Complaint, error) {
	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM complaints
		WHERE ModerationStatus = $1
		ORDER BY Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaints_by_status",
			Message: "failed to query complaints by status: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)
	for rows.Next() {
		c, err := scanComplaint(rows)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return complaints, nil
}

func (r *postgresRepository) GetAllComplaints(ctx context.Context) ([]models.Complaint, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_complaints",
		Message: "getting all complaints",
	})

	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM complaints
		ORDER BY Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_complaints",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)
	for rows.Next() {
		c, err := scanComplaint(rows)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_complaints",
				Message: "failed to scan complaint: " + err.Error(),
			})
			return nil, err
		}
		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return complaints, nil
}

func (r *postgresRepository) GetComplaintsByUserId(ctx context.Context, userID int) ([]models.Complaint, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_complaint_by_user_id",
		Message: "getting needed complaint by user id",
	})

	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM complaints
		WHERE TargetID = $1
		ORDER BY Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaint_by_user_id",
			Message: "failed to query complaints: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)
	for rows.Next() {
		c, err := scanComplaint(rows)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return complaints, nil
}

func (r *postgresRepository) GetComplaintsByClassID(ctx context.Context, classID int) ([]models.Complaint, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	query := `
		SELECT c.Id, c.TargetID, c.TargetName, c.AuthorID, c.AuthorName, c.Content, c.CreatedAt, c.ModerateAt, c.ModeratorId, c.ModerationStatus
		FROM complaints c
		JOIN users u ON u.Id = c.TargetID
		WHERE u.ClassID = $1
		ORDER BY c.Id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, classID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_complaints_by_class",
			Message: "failed to query complaints by class: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)
	for rows.Next() {
		c, err := scanComplaint(rows)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return complaints, nil
}

func (r *postgresRepository) GetComplaintByID(ctx context.Context, id int) ([]models.Complaint, error) {
	query := `
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM complaints
		WHERE Id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	complaints := make([]models.Complaint, 0)
	for rows.Next() {
		c, err := scanComplaint(rows)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return complaints, nil
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