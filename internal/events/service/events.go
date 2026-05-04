package service

import (
	"cspirt/internal/events/models"
	"cspirt/internal/events/repo"
	"cspirt/internal/logger"
	userModels "cspirt/internal/users/models"
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

func (s *EventsService) GetEventsByEventID(eventID int) (*models.Event, error) {
	event, err := s.events.GetEventByID(eventID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_events_by_event",
			Message: "failed to get event by ID: " + err.Error(),
		})
		return nil, err
	}
	if event == nil {
		return nil, nil
	}

	return event, nil
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

func (s *EventsService) AddPlayersToEvent(eventID int, playerIDs []int) error {
	if err := s.events.AddPlayersToEvent(eventID, playerIDs); err != nil {
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

func (s *EventsService) EventComplete(eventID int, ratingReward int) error {
	if err := s.events.EventComplete(eventID, ratingReward); err != nil {
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
