package repo

import (
	ratingModels "cspirt/internal/domain/rating"
	"cspirt/internal/domain/complaint/repo"
	models "cspirt/internal/domain/user"
	"cspirt/internal/controller/utils"
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

func New(db *sql.DB) repo.ComplaintRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) AddComplaint(login string, complaint models.Complaint, user models.SafeUser) error {
	if login == "" {
		return errors.New("invalid login or token")
	}

	complaint.Content = strings.TrimSpace(complaint.Content)
	if complaint.TargetID <= 0 {
		return errors.New("target and author are required")
	}
	if complaint.Content == "" {
		return errors.New("content is required")
	}

	targetUser, err := r.getUserByIDLocked(complaint.TargetID)
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
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	CreateAt := time.Now()

	_, err = r.db.Exec(
		query,
		complaint.TargetID,
		targetUser.Name,
		user.ID,
		user.Name,
		complaint.Content,
		CreateAt,
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

func (r *postgresRepository) DeleteComplaint(id int, user models.SafeUser) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_complaint",
		Login:   user.Login,
		Role:    user.Role,
		Message: "deleting complaint",
	})

	query := `DELETE FROM complaints WHERE Id = $1`

	check, err := r.hasUserRoleLocked(user.Login, string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner))
	if err != nil {
		return err
	}
	if !check {
		return errors.New("only admins can delete complaints")
	}

	result, err := r.db.Exec(query, id)
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

func (r *postgresRepository) GetAllComplaints() ([]models.Complaint, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_complaints",
		Message: "getting all complaints",
	})

	rows, err := r.db.Query(`
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

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var complaint models.Complaint
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

func (r *postgresRepository) GetComplaintByID(id int) ([]models.Complaint, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_complaint_by_id",
		Message: "getting needed complaint by id",
	})

	rows, err := r.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM complaints
		WHERE Id = $1
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

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var complaint models.Complaint
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

func (r *postgresRepository) GetComplaintsByUserId(User_id int) ([]models.Complaint, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_complaint_by_user_id",
		Message: "getting needed complaint by user id",
	})

	rows, err := r.db.Query(`
		SELECT Id, TargetID, TargetName, AuthorID, AuthorName, Content, CreatedAt
		FROM complaints
		WHERE TargetID = $1
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

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var c models.Complaint
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
			return []models.Complaint{}, err
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

func (r *postgresRepository) GetComplaintsByClassID(classID int) ([]models.Complaint, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	rows, err := r.db.Query(`
		SELECT c.Id, c.TargetID, c.TargetName, c.AuthorID, c.AuthorName, c.Content, c.CreatedAt
		FROM complaints c
		JOIN users u ON u.Id = c.TargetID
		WHERE u.ClassID = $1
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

	complaints := make([]models.Complaint, 0)

	for rows.Next() {
		var c models.Complaint
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

func (r *postgresRepository) getUserByIDLocked(id int) (*models.User, error) {
	row := r.db.QueryRow(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Id = $1
	`, id)

	var user models.User
	var fullNameJSON sql.NullString

	err := row.Scan(
		&user.ID,
		&user.Avatar,
		&user.Name,
		&fullNameJSON,
		&user.LastName,
		&user.Login,
		&user.Password,
		&user.Rating,
		&user.Role,
		&user.Class,
		&user.ClassID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresRepository) hasUserRoleLocked(login string, roles ...string) (bool, error) {
	var role string
	err := r.db.QueryRow(`SELECT Role FROM users WHERE Login = $1`, strings.TrimSpace(login)).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("user not found")
		}
		return false, err
	}

	userRole := strings.ToLower(strings.TrimSpace(role))
	for _, r := range roles {
		if userRole == strings.ToLower(strings.TrimSpace(r)) {
			return true, nil
		}
	}

	return false, nil
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
