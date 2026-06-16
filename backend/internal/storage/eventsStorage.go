package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cspirt/internal/events/models"
	"cspirt/internal/logger"
	userModels "cspirt/internal/users/models"
)

func (s *Storage) ActivateDueEvents() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Format("2006-01-02 15:04:05")

	_, err := s.db.Exec(`
		UPDATE events
		SET Status = 'active'
		WHERE Status IN ('scheduled', 'pending')
		AND StartedAt <= ?
	`, now)

	return err
}

func (s *Storage) GetEventsByID(eventID int) (*models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	rows, err := s.db.Query(`
		SELECT Id, Title, Status, RatingReward, Description, CreatedAt, StartedAt, Players, Classes
		FROM events
		WHERE Id = ?
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

func scanEvent(row *sql.Row) (*models.Event, error) {
	event := &models.Event{}
	err := row.Scan(&event.ID, &event.Title, &event.Status, &event.BaseRatingReward, &event.Description, &event.CreatedAt, &event.StartedAt, &event.Players, &event.Classes)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *Storage) GetEvents() ([]models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_events",
		Message: "Getting all events",
	})

	rows, err := s.db.Query(`
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

func (s *Storage) AddEvent(event models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event",
		Message: "adding event",
	})

	event.Players = normalizePositiveIDs(event.Players)
	classes := normalizePositiveIDs(event.Classes)
	if len(classes) == 0 {
		var err error
		classes, err = s.getClassIDsByPlayerIDsLocked(event.Players)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event",
				Message: "failed to resolve event classes: " + err.Error(),
			})
			return err
		}
	}

	playersJSON, err := marshalPlayerIDs(event.Players)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to marshal event players: " + err.Error(),
		})
		return err
	}
	classesJSON, err := marshalClassIDs(classes)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to marshal event classes: " + err.Error(),
		})
		return err
	}

	tx, err := s.db.Begin()
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

	result, err := tx.Exec(`
		INSERT INTO events (Title, Status, RatingReward, Description, CreatedAt, StartedAt, Players, Classes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, event.Title, event.Status, event.BaseRatingReward, event.Description, createdAt, event.StartedAt, playersJSON, classesJSON)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to insert event: " + err.Error(),
		})
		return err
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to get inserted event id: " + err.Error(),
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

func (s *Storage) GetEventParams(eventID int) ([]models.EventParams, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	rows, err := s.db.Query(`
		SELECT EventID, ExtraRatingReward, Reason, ClassID
		FROM event_params
		WHERE EventID = ?
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

func (s *Storage) DeleteEventParams(eventID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	tx, err := s.db.Begin()
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
		WHERE EventID = ?
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

func (s *Storage) AddEventParams(eventID int, params *models.EventParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	if params == nil {
		return errors.New("invalid event params")
	}

	tx, err := s.db.Begin()
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
		VALUES (?, ?, ?, ?)
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

func (s *Storage) GetEventPlayers(eventID int) ([]userModels.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	var exists int
	if err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM events
			WHERE Id = ?
		)
	`, eventID).Scan(&exists); err != nil {
		return nil, err
	}

	if exists == 0 {
		return nil, errors.New("event not found")
	}

	rows, err := s.db.Query(`
		SELECT u.Id, u.Name, u.FullName, u.LastName, u.Login, u.Rating, u.Role, u.Class, u.ClassID
		FROM event_players ep
		JOIN users u ON u.Id = ep.player_id
		WHERE ep.event_id = ?
		ORDER BY u.LastName, u.Name, u.Login
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSafeUsers(rows)
}

func (s *Storage) GetEventPlayersCount(eventID int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	players, err := s.getEventPlayersLocked(eventID)
	if err != nil {
		return 0, err
	}

	return len(players), nil
}

func (s *Storage) EventComplete(eventID int, ratingReward int, _ int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "complete_event",
		Message: "completing event",
	})

	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	classIDs, err := s.getEventClassIDsForRatingLocked(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "complete_event",
			Message: "failed to resolve event classes: " + err.Error(),
		})
		return err
	}

	tx, err := s.db.Begin()
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
		SET Status = 'completed', RatingReward = ?
		WHERE Id = ? AND Status != 'completed'
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
			WHERE ep.event_id = ? AND u.classID = ?
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
				SET Rating = MAX(0, MIN(5000, Rating + ?))
				WHERE Id = ?
			`, ratingReward, playerID)
			if err != nil {
				return err
			}
		}
	}

	for classID, extraRatingReward := range classRewards {
		_, err = tx.Exec(`
			UPDATE classes
			SET ClassTotalRating = ClassTotalRating + ?
			WHERE Id = ?
		`, extraRatingReward, classID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, classID := range appendUniqueIDs(classIDs, rewardClassIDs) {
		if err := s.syncClassByIDLocked(classID); err != nil {
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
		WHERE EventID = ? AND ClassID > 0
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

func (s *Storage) GetEventsByUserID(userID int) ([]models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT e.Id, e.Title, e.Status, e.RatingReward, e.Description, e.CreatedAt, e.StartedAt, e.Players, e.Classes
		FROM events e
		JOIN event_players ep ON ep.event_id = e.Id
		WHERE ep.player_id = ?
		ORDER BY e.Id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (s *Storage) GetEventsByClassID(classID int) ([]models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT e.Id, e.Title, e.Status, e.RatingReward, e.Description, e.CreatedAt, e.StartedAt, e.Players, e.Classes
		FROM events e
		JOIN event_classes ec ON ec.event_id = e.Id
		WHERE ec.class_id = ?
		ORDER BY e.Id
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (s *Storage) DeleteEvent(eventID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_event",
		Message: "deleting event",
	})

	tx, err := s.db.Begin()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to start transaction: " + err.Error(),
		})
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM event_players WHERE event_id = ?`, eventID); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to delete event players: " + err.Error(),
		})
		return err
	}
	if _, err := tx.Exec(`DELETE FROM event_classes WHERE event_id = ?`, eventID); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to delete event classes: " + err.Error(),
		})
		return err
	}

	result, err := tx.Exec(`DELETE FROM events WHERE Id = ?`, eventID)
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

func (s *Storage) AddPlayersToEvent(eventID int, playerIDs []int, login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:  "info",
		Action: "add_event_players",
		Message: "adding players to event",
	})

	user, err := s.getUserByLoginLocked(login)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.ClassID <= 0 {
		return errors.New("user has no class")
	}

	class, err := s.getClassByIDLocked(user.ClassID)
	if err != nil {
		return err
	}

	if class == nil {
		return errors.New("class not found")
	}

	playerIDs = normalizePositiveIDs(playerIDs)
	if len(playerIDs) == 0 {
		return nil
	}

	for _, playerID := range playerIDs {
		player, err := s.getUserByIDLocked(playerID)
		if err != nil {
			return err
		}
		if player == nil {
			return errors.New("player not found")
		}
	}

	currentPlayers, err := s.getEventPlayersLocked(eventID)
	if err != nil {
		return err
	}

	currentClasses, err := s.getEventClassesLocked(eventID)
	if err != nil {
		return err
	}

	players := appendUniquePlayerIDs(currentPlayers, playerIDs)
	playersJSON, err := marshalPlayerIDs(players)
	if err != nil {
		return err
	}

	classes := appendUniqueIDs(currentClasses, []int{class.ID})
	classesJSON, err := marshalClassIDs(classes)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := insertEventPlayers(tx, eventID, playerIDs); err != nil {
		return err
	}

	if err := insertEventClasses(tx, eventID, []int{class.ID}); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`UPDATE events SET Players = ?, Classes = ? WHERE Id = ?`,
		playersJSON, classesJSON, eventID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) UpdateEventParams(eventID int, params *models.EventParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	if params == nil {
		return errors.New("invalid event params")
	}

	tx, err := s.db.Begin()
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
		SET ExtraRatingReward = ?, Reason = ?, ClassID = ?
		WHERE EventID = ? AND ClassID = ?
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
		if err := s.AddEventParams(eventID, params); err != nil {
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

func (s *Storage) UpdateEvent(eventID int, updatedEvent *models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	if updatedEvent == nil {
		return errors.New("invalid event data")
	}

	currentEvent, err := s.GetEventsByID(eventID)
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

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		UPDATE events
		SET Title = ?, Status = ?, RatingReward = ?, Description = ?, CreatedAt = ?, StartedAt = ?, Players = ?, Classes = ?
		WHERE Id = ?
	`, updatedEvent.Title, updatedEvent.Status, updatedEvent.BaseRatingReward, updatedEvent.Description, updatedEvent.CreatedAt.Format(time.RFC3339), updatedEvent.StartedAt, updatedEvent.Players, updatedEvent.Classes, eventID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) DeletePlayersFromEvent(eventID int, playerIDs []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_event_players",
		Message: "deleting players from event",
	})

	currentPlayers, err := s.getEventPlayersLocked(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to get current event players: " + err.Error(),
		})
		return err
	}

	players := removePlayerIDs(currentPlayers, playerIDs)
	playersJSON, err := marshalPlayerIDs(players)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to marshal event players: " + err.Error(),
		})
		return err
	}

	tx, err := s.db.Begin()
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
			WHERE event_id = ? AND player_id = ?
		`, eventID, playerID); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_event_players",
				Message: "failed to delete event player: " + err.Error(),
			})
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE events SET Players = ? WHERE Id = ?`, playersJSON, eventID); err != nil {
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

func (s *Storage) getEventPlayersLocked(eventID int) ([]int, error) {
	var playersJSON string
	err := s.db.QueryRow(`
		SELECT Players
		FROM events
		WHERE Id = ?
	`, eventID).Scan(&playersJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	return unmarshalPlayerIDs(playersJSON)
}

func (s *Storage) getEventClassesLocked(eventID int) ([]int, error) {
	var classesJSON string
	err := s.db.QueryRow(`
		SELECT Classes
		FROM events
		WHERE Id = ?
	`, eventID).Scan(&classesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	return unmarshalClassIDs(classesJSON)
}

func (s *Storage) getEventClassIDsForRatingLocked(eventID int) ([]int, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT u.ClassID
		FROM users u
		JOIN event_players ep ON ep.player_id = u.Id
		WHERE ep.event_id = ? AND u.ClassID > 0
		UNION
		SELECT class_id
		FROM event_classes
		WHERE event_id = ?
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

func (s *Storage) getClassIDsByPlayerIDsLocked(playerIDs []int) ([]int, error) {
	playerIDs = normalizePositiveIDs(playerIDs)
	if len(playerIDs) == 0 {
		return []int{}, nil
	}

	classIDs := make([]int, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		var classID int
		err := s.db.QueryRow(`
			SELECT ClassID
			FROM users
			WHERE Id = ? AND ClassID > 0
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
			INSERT OR IGNORE INTO event_players (event_id, player_id)
			VALUES (?, ?)
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
			INSERT OR IGNORE INTO event_classes (event_id, class_id)
			VALUES (?, ?)
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

		players, err := unmarshalPlayerIDs(playersJSON)
		if err != nil {
			return nil, err
		}
		event.Players = players

		classes, err := unmarshalClassIDs(classesJSON)
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

func marshalPlayerIDs(playerIDs []int) (string, error) {
	return marshalIDs(playerIDs)
}

func unmarshalPlayerIDs(playersJSON string) ([]int, error) {
	return unmarshalIDs(playersJSON)
}

func marshalClassIDs(classIDs []int) (string, error) {
	return marshalIDs(classIDs)
}

func unmarshalClassIDs(classesJSON string) ([]int, error) {
	return unmarshalIDs(classesJSON)
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

func appendUniquePlayerIDs(current []int, additions []int) []int {
	return appendUniqueIDs(current, additions)
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

func removePlayerIDs(current []int, removals []int) []int {
	return removeIDs(current, removals)
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
