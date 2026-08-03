package moderate

import (
	"context"
	entity "cspirt/internal/domain/moderate"
	"cspirt/internal/domain/moderate/repo"
	userEntity "cspirt/internal/domain/user"
	"database/sql"
	"time"
)

type postgresRepository struct {
	db sql.DB
}

func New(db *sql.DB) repo.ModerateRepo {
	return &postgresRepository{}
}

func (r *postgresRepository) GetAllWaitNotes(ctx context.Context) ([]userEntity.Note, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT Id, TargetID, AuthorID, TargetName, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM notes
		WHERE ModerationStatus = 'wait'`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return []userEntity.Note{}, err
	}
	defer rows.Close()

	return r.scanWaitNotes(rows)
}

func (r *postgresRepository) GetAllWaitComplaints(ctx context.Context) ([]userEntity.Complaint, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT Id, TargetID, AuthorID, TargetName, AuthorName, Content, CreatedAt, ModerateAt, ModeratorId, ModerationStatus
		FROM complaints
		WHERE ModerationStatus = 'wait'`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return []userEntity.Complaint{}, err
	}
	defer rows.Close()

	return r.scanWaitComplaints(rows)
}

func (r *postgresRepository) UpdateNoteModerateStatus(ctx context.Context, id int, input entity.ModerateDTO) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	query := `UPDATE notes
		SET ModeratorId = $1, ModerationStatus = $2
		WHERE Id = $3`

	_, err := r.db.ExecContext(ctx, query, input.ModeratorId, input.NewStatus, id)
	if err != nil {
		return err
	}
	
	return nil
}

func (r *postgresRepository) UpdateComplaintModerateStatus(ctx context.Context, id int, input entity.ModerateDTO) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	query := `UPDATE complaints
		SET ModeratorId = $1, ModerationStatus = $2
		WHERE Id = $3`

	_, err := r.db.ExecContext(ctx, query, input.ModeratorId, input.NewStatus, id)
	if err != nil {
		return err
	}
	
	return nil
}

// METHODS 

func (r *postgresRepository) scanWaitComplaints(rows *sql.Rows) ([]userEntity.Complaint, error) {
	output := make([]userEntity.Complaint, 0)

	for rows.Next() {
		var waitNotes userEntity.Complaint

		if err := rows.Scan(
			&waitNotes.ID,
			&waitNotes.TargetID,
			&waitNotes.AuthorID,
			&waitNotes.TargetName,
			&waitNotes.AuthorName,
			&waitNotes.Content,
			&waitNotes.CreatedAt,
			&waitNotes.ModerateAt,
			&waitNotes.ModeratorId,
			&waitNotes.ModerationStatus,
		); err != nil {
			return []userEntity.Complaint{}, err
		}

		output = append(output, waitNotes)
	}

	return output, nil
}

func (r *postgresRepository) scanWaitNotes(rows *sql.Rows) ([]userEntity.Note, error) {
	output := make([]userEntity.Note, 0)

	for rows.Next() {
		var waitNotes userEntity.Note

		if err := rows.Scan(
			&waitNotes.ID,
			&waitNotes.TargetID,
			&waitNotes.AuthorID,
			&waitNotes.TargetName,
			&waitNotes.AuthorName,
			&waitNotes.Content,
			&waitNotes.CreatedAt,
			&waitNotes.ModerateAt,
			&waitNotes.ModeratorId,
			&waitNotes.ModerationStatus,
		); err != nil {
			return []userEntity.Note{}, err
		}

		output = append(output, waitNotes)
	}

	return output, nil
}

func (r *postgresRepository) CheckToWaitStatus(id int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT ModerationStatus FROM notes WHERE Id = $1`
	
	var status string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&status)
	if err != nil {
		return false, err 
	}

	return status == "wait", nil
}