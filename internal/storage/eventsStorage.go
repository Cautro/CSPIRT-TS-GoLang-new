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
)

func (s *Storage) GetEvents() ([]models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_events",
		Message: "Getting all events",
	})

	rows, err := s.db.Query(`
		SELECT Id, Title, Status, Description, CreatedAt, StartedAt, Players
		FROM events
		ORDER BY Id
	`)
	if err != nil {
		writeLog(logger.LogEntry{
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

	playersJSON, err := marshalPlayerIDs(event.Players)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO events (Title, Status, Description, CreatedAt, StartedAt, Players)
		VALUES (?, ?, ?, ?, ?, ?)
	`, event.Title, event.Status, event.Description, event.CreatedAt, event.StartedAt, playersJSON)
	if err != nil {
		return err
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	if err := insertEventPlayers(tx, int(eventID), event.Players); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetEventsByUserID(userID int) ([]models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT e.Id, e.Title, e.Status, e.Description, e.CreatedAt, e.StartedAt, e.Players
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

func (s *Storage) DeleteEvent(eventID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM event_players WHERE event_id = ?`, eventID); err != nil {
		return err
	}

	result, err := tx.Exec(`DELETE FROM events WHERE Id = ?`, eventID)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("event not found")
	}

	return tx.Commit()
}

func (s *Storage) AddPlayersToEvent(eventID int, playerIDs []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentPlayers, err := s.getEventPlayersLocked(eventID)
	if err != nil {
		return err
	}

	players := appendUniquePlayerIDs(currentPlayers, playerIDs)
	playersJSON, err := marshalPlayerIDs(players)
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

	if _, err := tx.Exec(`UPDATE events SET Players = ? WHERE Id = ?`, playersJSON, eventID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) DeletePlayersFromEvent(eventID int, playerIDs []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentPlayers, err := s.getEventPlayersLocked(eventID)
	if err != nil {
		return err
	}

	players := removePlayerIDs(currentPlayers, playerIDs)
	playersJSON, err := marshalPlayerIDs(players)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, playerID := range playerIDs {
		if _, err := tx.Exec(`
			DELETE FROM event_players
			WHERE event_id = ? AND player_id = ?
		`, eventID, playerID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE events SET Players = ? WHERE Id = ?`, playersJSON, eventID); err != nil {
		return err
	}

	return tx.Commit()
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

func scanEvents(rows *sql.Rows) ([]models.Event, error) {
	events := make([]models.Event, 0)

	for rows.Next() {
		var event models.Event
		var createdAt interface{}
		var playersJSON string

		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Status,
			&event.Description,
			&createdAt,
			&event.StartedAt,
			&playersJSON,
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
	if playerIDs == nil {
		playerIDs = []int{}
	}

	data, err := json.Marshal(playerIDs)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func unmarshalPlayerIDs(playersJSON string) ([]int, error) {
	if playersJSON == "" {
		return []int{}, nil
	}

	var playerIDs []int
	if err := json.Unmarshal([]byte(playersJSON), &playerIDs); err != nil {
		return nil, err
	}
	if playerIDs == nil {
		return []int{}, nil
	}

	return playerIDs, nil
}

func appendUniquePlayerIDs(current []int, additions []int) []int {
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
