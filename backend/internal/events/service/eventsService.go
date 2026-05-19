package service

import (
	"cspirt/internal/events/models"
	"cspirt/internal/events/repo"
	"cspirt/internal/logger"
	userModels "cspirt/internal/users/models"
	"errors"
	"strings"
	"time"
)

type EventsService struct {
	events repo.EventsRepository
}

func NewEventsService(events repo.EventsRepository, jwtSecret string) *EventsService {
	return &EventsService{
		events: events,
	}
}

func (s *EventsService) GetEvents() ([]models.Event, error) {
	events, err := s.events.GetEvents()
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_events",
			Message: "failed to get events: " + err.Error(),
		})
		return nil, err
	}
	if events == nil {
		return []models.Event{}, nil
	}

	return events, nil
}

func (s *EventsService) GetEventsByUserID(userID int) ([]models.Event, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	events, err := s.events.GetEventsByUserID(userID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_events_by_user",
			Message: "failed to get events by user: " + err.Error(),
		})
		return nil, err
	}
	if events == nil {
		return []models.Event{}, nil
	}

	return events, nil
}

func (s *EventsService) GetEventsByClassID(classID int) ([]models.Event, error) {
	events, err := s.events.GetEventsByClassID(classID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_events_by_class",
			Message: "failed to get events by class: " + err.Error(),
		})
		return nil, err
	}
	if events == nil {
		return []models.Event{}, nil
	}

	return events, nil
}

func (s *EventsService) AddEvent(event models.Event) error {
	event.Status = strings.ToLower(strings.TrimSpace(event.Status))
	if event.Status == "" {
		event.Status = "scheduled"
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	if event.StartedAt == "" {
		return errors.New("started at is required")
	}

	if err := s.events.AddEvent(event); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event",
			Message: "failed to add event: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event",
		Message: "event added",
	})
	return nil
}

func (s *EventsService) GetEventParams(eventID int) ([]models.EventParams, error) {
	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	result, err := s.events.GetEventParams(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_event_params",
			Message: "failed to get event params: " + err.Error(),
		})
		return nil, err
	}
	if result == nil {
		return []models.EventParams{}, nil
	}

	return result, nil
}

func (s *EventsService) AddEventParams(EventId int, params *models.EventParams) error {
	if EventId <= 0 {
		return errors.New("invalid event id")
	}
	if params == nil {
		return errors.New("invalid event params")
	}

	if err := s.events.AddEventParams(EventId, params); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event_params",
			Message: "failed to add event params: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event_params",
		Message: "event params added",
	})
	return nil
}

func (s *EventsService) DeleteEventParams(eventID int) error {
	if eventID <= 0 {
		return errors.New("invalid event id")
	}

	if err := s.events.DeleteEventParams(eventID); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_params",
			Message: "failed to delete event params: " + err.Error(),
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

func (s *EventsService) GetEventsByEventID(eventID int) (*models.Event, error) {
	if eventID <= 0 {
		return nil, errors.New("invalid event id")
	}

	return s.events.GetEventsByID(eventID)
}

func (s *EventsService) DeleteEvent(eventID int) error {
	if err := s.events.DeleteEvent(eventID); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event",
			Message: "failed to delete event: " + err.Error(),
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

func (s *EventsService) AddPlayersToEvent(eventID int, playerIDs []int, login string) error {
	if err := s.events.AddPlayersToEvent(eventID, playerIDs, login); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_event_players",
			Message: "failed to add players to event: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "add_event_players",
		Message: "players added to event",
	})
	return nil
}

func (s *EventsService) DeletePlayersFromEvent(eventID int, playerIDs []int) error {
	if err := s.events.DeletePlayersFromEvent(eventID, playerIDs); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_event_players",
			Message: "failed to delete players from event: " + err.Error(),
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

func (s *EventsService) GetEventPlayers(eventID int) ([]userModels.SafeUser, error) {
	players, err := s.events.GetEventPlayers(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_event_players",
			Message: "failed to get event players: " + err.Error(),
		})
		return nil, err
	}
	if players == nil {
		return []userModels.SafeUser{}, nil
	}

	return players, nil
}

func (s *EventsService) GetEventPlayersCount(eventID int) (int, error) {
	count, err := s.events.GetEventPlayersCount(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_event_players_count",
			Message: "failed to get event players count: " + err.Error(),
		})
		return 0, err
	}

	return count, nil
}

func (s *EventsService) EventComplete(eventID int, ratingReward int, classReward int) error {
	if err := s.events.EventComplete(eventID, ratingReward, classReward); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "complete_event",
			Message: "failed to complete event: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "complete_event",
		Message: "event completed",
	})
	return nil
}
