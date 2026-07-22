package postgres

import (
	"context"
	"errors"
	"database/sql"
	"encoding/json"
	"time"
	"strconv"

	entity "cspirt/internal/domain/globalEvent"
	repo "cspirt/internal/domain/globalEvent/repo"

	valid "github.com/go-playground/validator"
	//"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.GlobalEventRepo {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) GetAllGlobalEvents(ctx context.Context) ([]entity.GlobalEventOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	infoEvents := make([]entity.GlobalEventInfoEntity, 0)
	infoRows, err := r.db.QueryContext(ctx, "SELECT id, title, description FROM global_events_info")
	if err != nil {
		return nil, err
	}
	defer infoRows.Close()

	for infoRows.Next() {
		var info entity.GlobalEventInfoEntity
		if err := infoRows.Scan(&info.ID, &info.Title, &info.Description); err != nil {
			return nil, err
		}
		infoEvents = append(infoEvents, info)
	}
	if err := infoRows.Err(); err != nil {
		return nil, err
	}

	quizzes := make([]entity.GlobalEventQuizEntity, 0)
	qRows, err := r.db.QueryContext(ctx, "SELECT id, title, description, options FROM global_event_quizzes")
	if err != nil {
		return nil, err
	}
	defer qRows.Close()

	for qRows.Next() {
		var quiz entity.GlobalEventQuizEntity
		var optionsRaw []byte

		if err := qRows.Scan(&quiz.ID, &quiz.Title, &quiz.Description, &optionsRaw); err != nil {
			return nil, err
		}

		if len(optionsRaw) > 0 {
			if err := json.Unmarshal(optionsRaw, &quiz.Options); err != nil {
				return nil, err
			}
		}

		if quiz.Options == nil {
			quiz.Options = make([]entity.QuizOption, 0)
		}

		quizzes = append(quizzes, quiz)
	}
	if err := qRows.Err(); err != nil {
		return nil, err
	}

	return []entity.GlobalEventOutput{
		{
			InfoEvents: infoEvents,
			Quizzes:       quizzes,
		},
	}, nil
}

func (r *postgresRepository) AddInfoGlobalEvent(ctx context.Context, input entity.GlobalEventInfoDTO) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	validalor := valid.New()
	if err := validalor.StructCtx(ctx, input); err != nil {
		return err
	}

	query := `
	INSERT INTO global_events_info
	(title, description)
	VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(ctx, 
		query,
		input.Title,
		input.Description,
	); if err != nil { return err }

	return nil
}

func (r *postgresRepository) AddQuizGlobalEvent(ctx context.Context, input entity.GlobalEventQuizDTO) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	validator := valid.New()
	if err := validator.StructCtx(ctx, input); err != nil {
		return err
	}

	optionsJSON, err := json.Marshal(input.Option)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO global_event_quizzes (title, description, options)
		VALUES ($1, $2, $3)
	`

	_, err = r.db.ExecContext(ctx,
		query,
		input.Title,
		input.Description,
		optionsJSON,
	); if err != nil { return err }

	return nil
}

func (r *postgresRepository) DeleteInfoGlobalEvent(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if id <= 0 { return ErrBadRequest }

	query := `DELETE FROM global_events_info WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil { return ErrServer }

	return nil
}

func (r *postgresRepository) DeleteQuizGlobalEvent(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if id <= 0 { return ErrBadRequest }

	query := `DELETE FROM global_events_quiz WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil { return ErrServer }

	return nil
}

func (r *postgresRepository) Vote(ctx context.Context, idUser, eventId, voteItemId int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if idUser <= 0 || eventId <= 0 || voteItemId < 0 {
		return ErrBadRequest
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Фиксируем голос (все аргументы — int)
	insertVoteQuery := `
		INSERT INTO global_event_quiz_votes (quiz_id, user_id, option_index)
		VALUES ($1, $2, $3)
	`
	if _, err := tx.ExecContext(ctx, insertVoteQuery, eventId, idUser, voteItemId); err != nil {
		return errors.New("already voted or invalid request")
	}

	// 2. Подготовка строки для JSONB path
	voteItemStr := strconv.Itoa(voteItemId)

	// $1 передаётся как string (TEXT), $2 как int (INTEGER)
	updateQuizQuery := `
		UPDATE global_event_quizzes
		SET options = jsonb_set(
			options,
			ARRAY[$1, 'votes'],
			to_jsonb(COALESCE((options->($1::int)->>'votes')::int, 0) + 1)
		)
		WHERE id = $2
	`
	res, err := tx.ExecContext(ctx, updateQuizQuery, voteItemStr, eventId)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrBadRequest
	}

	return tx.Commit()
}