package repo

import "cspirt/internal/events/models"

type EventsRepository interface {
	AddEvent(event models.Event) error
	GetEventsByUserID(userID int) ([]models.Event, error)
	DeleteEvent(eventID int) error
	GetEvents() ([]models.Event, error)
	AddPlayersToEvent(eventID int, playerIDs []int) error
	DeletePlayersFromEvent(eventID int, playerIDs []int) error
}