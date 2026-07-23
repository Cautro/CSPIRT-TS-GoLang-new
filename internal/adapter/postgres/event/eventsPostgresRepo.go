package repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"context"

	models "cspirt/internal/domain/event"
	"cspirt/internal/domain/event/repo"
	ratingModels "cspirt/internal/domain/rating"
	userModels "cspirt/internal/domain/user"
	"cspirt/internal/controller/utils"
	"cspirt/pkg/logger"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.EventsRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ActivateDueEvents() error {
	now := time.Now().Format("2006-01-02 15:04:05")

	_, err := r.db.Exec(`
		UPDATE events
		SET Status = 'active'
		WHERE Status IN ('scheduled', 'pending')
		AND StartedAt <= $1
	`, now)

	return err
}

func (r *postgresRepository) GetEventsByID(ctx context.Context, eventID int) (*models.Event, error) {
	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	rows, err := r.db.Query(`
		SELECT Id, Title, Status, RatingReward, Description, CreatedAt, StartedAt, Players, Classes
		FROM events
		WHERE Id = $1
		LIMIT 1
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events, err := scanEvents(rows)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	return &events[0], nil
}

func (r *postgresRepository) GetEvents(ctx context.Context) ([]models.Event, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_events",
		Message: "Getting all events",
	})

	rows, err := r.db.QueryContext(ctx, `
		SELECT Id, Title, Status, RatingReward, Description, CreatedAt, StartedAt, Players, Classes
		FROM events
		ORDER BY Id
	`)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_events",
			Message: "failed to query events: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (r *postgresRepository) AddEvent(ctx context.Context, event models.Event) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event",
		Message: "adding event",
	})

	event.Players = normalizePositiveIDs(event.Players)
	classes := normalizePositiveIDs(event.Classes)
	if len(classes) == 0 {
		var err error
		classes, err = r.getClassIDsByPlayerIDsLocked(event.Players)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event",
				Message: "failed to resolve event classes: " + err.Error(),
			})
			return err
		}
	}

	playersJSON, err := marshalIDs(event.Players)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to marshal event players: " + err.Error(),
		})
		return err
	}
	classesJSON, err := marshalIDs(classes)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to marshal event classes: " + err.Error(),
		})
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	createdAt := event.CreatedAt.Format(time.RFC3339)

	var eventID int64
	err = tx.QueryRow(`
		INSERT INTO events (Title, Status, RatingReward, Description, CreatedAt, StartedAt, Players, Classes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING Id
	`, event.Title, event.Status, event.BaseRatingReward, event.Description, createdAt, event.StartedAt, playersJSON, classesJSON).Scan(&eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to insert event: " + err.Error(),
		})
		return err
	}

	if err := insertEventPlayers(tx, int(eventID), event.Players); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to insert event players: " + err.Error(),
		})
		return err
	}
	if err := insertEventClasses(tx, int(eventID), classes); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to insert event classes: " + err.Error(),
		})
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to commit event: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event",
		Message: "event inserted",
	})
	return nil
}

func (r *postgresRepository) GetEventParams(ctx context.Context, eventID int) ([]models.EventParams, error) {
	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	rows, err := r.db.Query(`
		SELECT EventID, ExtraRatingReward, Reason, ClassID
		FROM event_params
		WHERE EventID = $1
		ORDER BY Id
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	params := make([]models.EventParams, 0)
	for rows.Next() {
		var param models.EventParams
		if err := rows.Scan(&param.EventID, &param.ExtraRatingReward, &param.Reason, &param.ClassID); err != nil {
			return nil, err
		}
		params = append(params, param)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return params, nil
}

func (r *postgresRepository) DeleteEventParams(ctx context.Context, eventID int) error {
	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_params",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM event_params
		WHERE EventID = $1
	`, eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_params",
			Message: "failed to delete event params: " + err.Error(),
		})
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_params",
			Message: "failed to commit event params deletion: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_event_params",
		Message: "event params deleted",
	})
	return nil
}

func (r *postgresRepository) AddEventParams(ctx context.Context, eventID int, params *models.EventParams) error {
	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	if params == nil {
		return errors.New("invalid event params")
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event_params",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO event_params (EventID, ExtraRatingReward, Reason, ClassID)
		VALUES ($1, $2, $3, $4)
	`, eventID, params.ExtraRatingReward, params.Reason, params.ClassID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event_params",
			Message: "failed to insert event params: " + err.Error(),
		})
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event_params",
			Message: "failed to commit event params: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event_params",
		Message: "event params inserted",
	})
	return nil
}

func (r *postgresRepository) GetEventPlayers(ctx context.Context, eventID int) ([]userModels.SafeUser, error) {
	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	var exists int
	if err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM events
			WHERE Id = $1
		)
	`, eventID).Scan(&exists); err != nil {
		return nil, err
	}

	if exists == 0 {
		return nil, errors.New("event not found")
	}

	rows, err := r.db.Query(`
		SELECT u.Id, u.Name, u.FullName, u.LastName, u.Login, u.Rating, u.Role, u.Class, u.ClassID
		FROM event_players ep
		JOIN users u ON u.Id = ep.player_id
		WHERE ep.event_id = $1
		ORDER BY u.LastName, u.Name, u.Login
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSafeUsersNoAvatar(rows)
}

func (r *postgresRepository) GetEventPlayersCount(ctx context.Context, eventID int) (int, error) {
	players, err := r.getEventPlayersLocked(eventID)
	if err != nil {
		return 0, err
	}

	return len(players), nil
}

func (r *postgresRepository) EventComplete(ctx context.Context, eventID int, ratingReward int, _ int) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "complete_event",
		Message: "completing event",
	})

	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	classIDs, err := r.getEventClassIDsForRatingLocked(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "complete_event",
			Message: "failed to resolve event classes: " + err.Error(),
		})
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "complete_event",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE events
		SET Status = 'completed', RatingReward = $1
		WHERE Id = $2 AND Status != 'completed'
	`, ratingReward, eventID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("event not found or already completed")
	}

	classRewards, rewardClassIDs, err := getEventClassRewardsTx(tx, eventID)
	if err != nil {
		return err
	}

	for _, classID := range classIDs {
		rows, err := tx.Query(`
			SELECT ep.player_id
			FROM event_players ep
			JOIN users u ON ep.player_id = u.Id
			WHERE ep.event_id = $1 AND u.classID = $2
		`, eventID, classID)
		if err != nil {
			return err
		}

		var participantIDs []int
		for rows.Next() {
			var pID int
			if err := rows.Scan(&pID); err != nil {
				rows.Close()
				return err
			}
			participantIDs = append(participantIDs, pID)
		}
		rows.Close()

		participantsCount := len(participantIDs)
		if participantsCount == 0 {
			continue
		}

		for _, playerID := range participantIDs {
			_, err := tx.Exec(`
				UPDATE users
				SET Rating = GREATEST(0, LEAST(5000, Rating + $1))
				WHERE Id = $2
			`, ratingReward, playerID)
			if err != nil {
				return err
			}
		}
	}

	for classID, extraRatingReward := range classRewards {
		_, err = tx.Exec(`
			UPDATE classes
			SET ClassTotalRating = ClassTotalRating + $1
			WHERE Id = $2
		`, extraRatingReward, classID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, classID := range appendUniqueIDs(classIDs, rewardClassIDs) {
		if err := r.syncClassByIDLocked(classID); err != nil {
			return err
		}
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "complete_event",
		Message: "event completed",
	})
	return nil
}

func getEventClassRewardsTx(tx *sql.Tx, eventID int) (map[int]int, []int, error) {
	rows, err := tx.Query(`
		SELECT ClassID, SUM(ExtraRatingReward)
		FROM event_params
		WHERE EventID = $1 AND ClassID > 0
		GROUP BY ClassID
	`, eventID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	rewards := make(map[int]int)
	classIDs := make([]int, 0)
	for rows.Next() {
		var classID int
		var reward int
		if err := rows.Scan(&classID, &reward); err != nil {
			return nil, nil, err
		}
		if reward == 0 {
			continue
		}
		rewards[classID] += reward
		classIDs = appendUniqueIDs(classIDs, []int{classID})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return rewards, classIDs, nil
}

func (r *postgresRepository) GetEventsByUserID(ctx context.Context, userID int) ([]models.Event, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.Id, e.Title, e.Status, e.RatingReward, e.Description, e.CreatedAt, e.StartedAt, e.Players, e.Classes
		FROM events e
		JOIN event_players ep ON ep.event_id = e.Id
		WHERE ep.player_id = $1
		ORDER BY e.Id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (r *postgresRepository) GetEventsByClassID(ctx context.Context, classID int) ([]models.Event, error) {
	rows, err := r.db.Query(`
		SELECT e.Id, e.Title, e.Status, e.RatingReward, e.Description, e.CreatedAt, e.StartedAt, e.Players, e.Classes
		FROM events e
		JOIN event_classes ec ON ec.event_id = e.Id
		WHERE ec.class_id = $1
		ORDER BY e.Id
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (r *postgresRepository) DeleteEvent(ctx context.Context, eventID int) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_event",
		Message: "deleting event",
	})

	tx, err := r.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM event_players WHERE event_id = $1`, eventID); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to delete event players: " + err.Error(),
		})
		return err
	}
	if _, err := tx.Exec(`DELETE FROM event_classes WHERE event_id = $1`, eventID); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to delete event classes: " + err.Error(),
		})
		return err
	}

	result, err := tx.Exec(`DELETE FROM events WHERE Id = $1`, eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to delete event: " + err.Error(),
		})
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_event",
			Message: "event not found",
		})
		return errors.New("event not found")
	}

	if err := tx.Commit(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to commit event delete: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_event",
		Message: "event deleted",
	})
	return nil
}

func (r *postgresRepository) AddPlayersToEvent(ctx context.Context, eventID int, playerIDs []int, login string) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event_players",
		Message: "adding players to event",
	})

	user, err := r.getUserByLoginLocked(login)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.ClassID <= 0 && user.Role != string(ratingModels.RoleOwner) {
		return errors.New("user has no class")
	}

	class, err := r.getClassByIDLocked(user.ClassID)
	if err != nil {
		return err
	}

	if class == nil && user.Role != string(ratingModels.RoleOwner) {
		return errors.New("class not found")
	}

	playerIDs = normalizePositiveIDs(playerIDs)
	if len(playerIDs) == 0 {
		return nil
	}

	for _, playerID := range playerIDs {
		player, err := r.getUserByIDLocked(playerID)
		if err != nil {
			return err
		}
		if player == nil {
			return errors.New("player not found")
		}
	}

	currentPlayers, err := r.getEventPlayersLocked(eventID)
	if err != nil {
		return err
	}

	currentClasses, err := r.getEventClassesLocked(eventID)
	if err != nil {
		return err
	}

	players := appendUniqueIDs(currentPlayers, playerIDs)
	playersJSON, err := marshalIDs(players)
	if err != nil {
		return err
	}

	var classID int
	if class != nil {
		classID = class.id
	}
	classes := appendUniqueIDs(currentClasses, []int{classID})
	classesJSON, err := marshalIDs(classes)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := insertEventPlayers(tx, eventID, playerIDs); err != nil {
		return err
	}

	if err := insertEventClasses(tx, eventID, []int{classID}); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`UPDATE events SET Players = $1, Classes = $2 WHERE Id = $3`,
		playersJSON, classesJSON, eventID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresRepository) UpdateEventParams(ctx context.Context, eventID int, params *models.EventParams) error {
	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	if params == nil {
		return errors.New("invalid event params")
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "update_event_params",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE event_params
		SET ExtraRatingReward = $1, Reason = $2, ClassID = $3
		WHERE EventID = $4 AND ClassID = $5
	`, params.ExtraRatingReward, params.Reason, params.ClassID, eventID, params.ClassID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "update_event_params",
			Message: "failed to update event params: " + err.Error(),
		})
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "update_event_params",
			Message: "failed to get affected rows: " + err.Error(),
		})
		return err
	}
	if affected == 0 {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_event_params",
			Message: "event params not found, inserting new params",
		})
		if err := r.AddEventParams(ctx, eventID, params); err != nil {
			return err
		}
	} else {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_event_params",
			Message: fmt.Sprintf("event params updated, affected rows: %d", affected),
		})
	}

	if err := tx.Commit(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "update_event_params",
			Message: "failed to commit event params update: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "update_event_params",
		Message: "event params updated successfully",
	})
	return nil
}

func (r *postgresRepository) UpdateEvent(ctx context.Context, eventID int, updatedEvent *models.Event) error {

	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	if updatedEvent == nil {
		return errors.New("invalid event data")
	}

	currentEvent, err := r.GetEventsByID(ctx, eventID)
	if err != nil {
		return err
	}
	if currentEvent == nil {
		return errors.New("event not found")
	}

	updatedEvent.ID = currentEvent.ID
	if updatedEvent.CreatedAt.IsZero() {
		updatedEvent.CreatedAt = currentEvent.CreatedAt
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE events
		SET Title = $1, Status = $2, RatingReward = $3, Description = $4, CreatedAt = $5, StartedAt = $6, Players = $7, Classes = $8
		WHERE Id = $9
	`, updatedEvent.Title, updatedEvent.Status, updatedEvent.BaseRatingReward, updatedEvent.Description, updatedEvent.CreatedAt.Format(time.RFC3339), updatedEvent.StartedAt, updatedEvent.Players, updatedEvent.Classes, eventID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresRepository) DeletePlayersFromEvent(ctx context.Context, eventID int, playerIDs []int) error {

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_event_players",
		Message: "deleting players from event",
	})

	currentPlayers, err := r.getEventPlayersLocked(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to get current event players: " + err.Error(),
		})
		return err
	}

	players := removeIDs(currentPlayers, playerIDs)
	playersJSON, err := marshalIDs(players)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to marshal event players: " + err.Error(),
		})
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	for _, playerID := range playerIDs {
		if _, err := tx.Exec(`
			DELETE FROM event_players
			WHERE event_id = $1 AND player_id = $2
		`, eventID, playerID); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_event_players",
				Message: "failed to delete event player: " + err.Error(),
			})
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE events SET Players = $1 WHERE Id = $2`, playersJSON, eventID); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to update event players: " + err.Error(),
		})
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to commit event players: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_event_players",
		Message: "players deleted from event",
	})
	return nil
}

func (r *postgresRepository) getEventPlayersLocked(eventID int) ([]int, error) {
	var playersJSON string
	err := r.db.QueryRow(`
		SELECT Players
		FROM events
		WHERE Id = $1
	`, eventID).Scan(&playersJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	return unmarshalIDs(playersJSON)
}

func (r *postgresRepository) getEventClassesLocked(eventID int) ([]int, error) {
	var classesJSON string
	err := r.db.QueryRow(`
		SELECT Classes
		FROM events
		WHERE Id = $1
	`, eventID).Scan(&classesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	return unmarshalIDs(classesJSON)
}

func (r *postgresRepository) getEventClassIDsForRatingLocked(eventID int) ([]int, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT u.ClassID
		FROM users u
		JOIN event_players ep ON ep.player_id = u.Id
		WHERE ep.event_id = $1 AND u.ClassID > 0
		UNION
		SELECT class_id
		FROM event_classes
		WHERE event_id = $2
	`, eventID, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	classIDs := make([]int, 0)
	for rows.Next() {
		var classID int
		if err := rows.Scan(&classID); err != nil {
			return nil, err
		}
		classIDs = append(classIDs, classID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return classIDs, nil
}

func (r *postgresRepository) getClassIDsByPlayerIDsLocked(playerIDs []int) ([]int, error) {
	playerIDs = normalizePositiveIDs(playerIDs)
	if len(playerIDs) == 0 {
		return []int{}, nil
	}

	classIDs := make([]int, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		var classID int
		err := r.db.QueryRow(`
			SELECT ClassID
			FROM users
			WHERE Id = $1 AND ClassID > 0
		`, playerID).Scan(&classID)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
		classIDs = append(classIDs, classID)
	}

	return normalizePositiveIDs(classIDs), nil
}

func insertEventPlayers(tx *sql.Tx, eventID int, playerIDs []int) error {
	for _, playerID := range playerIDs {
		if playerID <= 0 {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO event_players (event_id, player_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, eventID, playerID); err != nil {
			return err
		}
	}

	return nil
}

func insertEventClasses(tx *sql.Tx, eventID int, classIDs []int) error {
	for _, classID := range classIDs {
		if classID <= 0 {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO event_classes (event_id, class_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, eventID, classID); err != nil {
			return err
		}
	}

	return nil
}

func scanEvents(rows *sql.Rows) ([]models.Event, error) {
	events := make([]models.Event, 0)

	for rows.Next() {
		var event models.Event
		var createdAt interface{}
		var playersJSON string
		var classesJSON string

		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Status,
			&event.BaseRatingReward,
			&event.Description,
			&createdAt,
			&event.StartedAt,
			&playersJSON,
			&classesJSON,
		); err != nil {
			return nil, err
		}

		parsedTime, err := parseEventTime(createdAt)
		if err != nil {
			return nil, err
		}
		event.CreatedAt = parsedTime

		players, err := unmarshalIDs(playersJSON)
		if err != nil {
			return nil, err
		}
		event.Players = players

		classes, err := unmarshalIDs(classesJSON)
		if err != nil {
			return nil, err
		}
		event.Classes = classes

		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
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

func marshalIDs(ids []int) (string, error) {
	ids = normalizePositiveIDs(ids)
	if ids == nil {
		ids = []int{}
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func unmarshalIDs(idsJSON string) ([]int, error) {
	if idsJSON == "" {
		return []int{}, nil
	}

	var ids []int
	if err := json.Unmarshal([]byte(idsJSON), &ids); err != nil {
		return nil, err
	}
	if ids == nil {
		return []int{}, nil
	}

	return normalizePositiveIDs(ids), nil
}

func appendUniqueIDs(current []int, additions []int) []int {
	result := append([]int(nil), current...)
	seen := make(map[int]struct{}, len(current)+len(additions))

	for _, playerID := range current {
		seen[playerID] = struct{}{}
	}
	for _, playerID := range additions {
		if playerID <= 0 {
			continue
		}
		if _, ok := seen[playerID]; ok {
			continue
		}
		seen[playerID] = struct{}{}
		result = append(result, playerID)
	}

	return result
}

func removeIDs(current []int, removals []int) []int {
	remove := make(map[int]struct{}, len(removals))
	for _, playerID := range removals {
		remove[playerID] = struct{}{}
	}

	result := make([]int, 0, len(current))
	for _, playerID := range current {
		if _, ok := remove[playerID]; !ok {
			result = append(result, playerID)
		}
	}

	return result
}

func normalizePositiveIDs(ids []int) []int {
	if len(ids) == 0 {
		return []int{}
	}

	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}

// --- shared user/class lookups (duplicated from users/repo & class/repo;
// repos intentionally don't depend on one another) ---

type minimalClass struct {
	id int
}

func (r *postgresRepository) getClassByIDLocked(id int) (*minimalClass, error) {
	if id <= 0 {
		return nil, nil
	}
	var exists bool
	if err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM classes WHERE id = $1)`, id).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return &minimalClass{id: id}, nil
}

func (r *postgresRepository) getUserByLoginLocked(login string) (*userModels.User, error) {
	row := r.db.QueryRow(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Login = $1
	`, strings.TrimSpace(login))

	return scanUser(row)
}

func (r *postgresRepository) getUserByIDLocked(id int) (*userModels.User, error) {
	row := r.db.QueryRow(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Id = $1
	`, id)

	return scanUser(row)
}

type userScanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(scanner userScanner) (*userModels.User, error) {
	var user userModels.User
	var fullNameJSON sql.NullString

	err := scanner.Scan(
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

	if fullNameJSON.Valid && fullNameJSON.String != "" {
		if err := json.Unmarshal([]byte(fullNameJSON.String), &user.FullName); err != nil {
			return nil, err
		}
	}
	if user.FullName == nil {
		user.FullName = []userModels.FullName{}
	}

	return &user, nil
}

func scanSafeUsersNoAvatar(rows *sql.Rows) ([]userModels.SafeUser, error) {
	users := make([]userModels.SafeUser, 0)

	for rows.Next() {
		var user userModels.SafeUser
		var fullNameJSON sql.NullString

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&fullNameJSON,
			&user.LastName,
			&user.Login,
			&user.Rating,
			&user.Role,
			&user.Class,
			&user.ClassID,
		); err != nil {
			return nil, err
		}

		if fullNameJSON.Valid && fullNameJSON.String != "" {
			if err := json.Unmarshal([]byte(fullNameJSON.String), &user.FullName); err != nil {
				return nil, err
			}
		}
		if user.FullName == nil {
			user.FullName = []userModels.FullName{}
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *postgresRepository) syncClassByIDLocked(classID int) error {
	if classID <= 0 {
		return nil
	}

	var exists bool
	if err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM classes WHERE id = $1)`, classID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return nil
	}

	rows, err := r.db.Query(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		WHERE ClassID = $1
		ORDER BY LastName, Name, Login
	`, classID)
	if err != nil {
		return err
	}

	members := make([]userModels.SafeUser, 0)
	for rows.Next() {
		var user userModels.SafeUser
		var fullNameJSON sql.NullString
		if err := rows.Scan(
			&user.ID,
			&user.Avatar,
			&user.Name,
			&fullNameJSON,
			&user.LastName,
			&user.Login,
			&user.Rating,
			&user.Role,
			&user.Class,
			&user.ClassID,
		); err != nil {
			rows.Close()
			return err
		}
		if fullNameJSON.Valid && fullNameJSON.String != "" {
			if err := json.Unmarshal([]byte(fullNameJSON.String), &user.FullName); err != nil {
				rows.Close()
				return err
			}
		}
		if user.FullName == nil {
			user.FullName = []userModels.FullName{}
		}
		members = append(members, user)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	userTotalRating := 0
	if len(members) > 0 {
		for _, member := range members {
			userTotalRating += member.Rating
		}
		userTotalRating = userTotalRating / len(members)
	}

	membersJSON, err := json.Marshal(members)
	if err != nil {
		return err
	}

	var teacherLogin sql.NullString
	err = r.db.QueryRow(`SELECT TeacherLogin FROM classes WHERE id = $1`, classID).Scan(&teacherLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if teacherLogin.Valid && teacherLogin.String != "" {
		teacher, err := r.getUserByLoginLocked(teacherLogin.String)
		if err != nil {
			return err
		}

		if teacher == nil {
			teacherLogin = sql.NullString{}
		} else if !utils.IsSystemRole(teacher.Role) && teacher.ClassID != classID {
			teacherLogin = sql.NullString{}
		}
	}

	if !teacherLogin.Valid || teacherLogin.String == "" {
		candidate, err := r.findTeacherCandidateLocked(classID)
		if err != nil {
			return err
		}
		if candidate != "" {
			teacherLogin = sql.NullString{String: candidate, Valid: true}
		}
	}

	_, err = r.db.Exec(`
		UPDATE classes
		SET Members = $1, UserTotalRating = $2, TeacherLogin = $3
		WHERE id = $4
	`, string(membersJSON), userTotalRating, teacherLogin, classID)
	return err
}

func (r *postgresRepository) findTeacherCandidateLocked(classID int) (string, error) {
	var login string
	err := r.db.QueryRow(`
		SELECT Login
		FROM users
		WHERE ClassID = $1
		AND LOWER(Role) IN ('admin', 'owner', 'helper')
		ORDER BY
			CASE LOWER(Role)
				WHEN 'admin' THEN 0
				WHEN 'owner' THEN 1
				WHEN 'helper' THEN 2
				ELSE 3
			END,
			Id
		LIMIT 1
	`, classID).Scan(&login)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return login, nil
}
