package storage

import (
	"cspirt/internal/events/models"
	"cspirt/internal/logger"
)

func (s *Storage) GetEvents() ([]models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level: "info",
		Action: "get_all_events",
		Message: "Getting all events",
	})

	rows, err := s.db.Query(`
		SELECT id, title, status, description, created_at, started_at, players 
		FROM events
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

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Status, &event.Description, &event.CreatedAt, &event.StartedAt, &event.Players); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) AddEvent(event models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
		INSERT INTO events (title, status, description, created_at, started_at, players)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, event.Title, event.Status, event.Description, event.CreatedAt, event.StartedAt, event.Players)
	return err
}

func (s *Storage) GetEventsByUserID(userID int) ([]models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
		SELECT id, title, status, description, created_at, started_at, players 
		FROM events
		WHERE user_id = ?
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Status, &event.Description, &event.CreatedAt, &event.StartedAt, &event.Players); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) DeleteEvent(eventID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
		DELETE FROM events
		WHERE id = ?
	`

	_, err := s.db.Exec(query, eventID)
	return err
}

func (s *Storage) AddPlayersToEvent(eventID int, playerIDs []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, playerID := range playerIDs {
		query := `
			INSERT INTO event_players (event_id, player_id)
			VALUES (?, ?)
		`
		_, err := s.db.Exec(query, eventID, playerID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) DeletePlayersFromEvent(eventID int, playerIDs []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, playerID := range playerIDs {
		query := `
			DELETE FROM event_players
			WHERE event_id = ? AND player_id = ?
		`
		_, err := s.db.Exec(query, eventID, playerID)
		if err != nil {
			return err
		}
	}
	return nil
}