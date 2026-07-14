package repo

import (
	userModels "cspirt/internal/domain/user"
	entity "cspirt/internal/domain/event"
)

type EventsRepository interface {
	ActivateDueEvents() error
	AddEvent(event entity.Event) error
	GetEventsByUserID(userID int) ([]entity.Event, error)
	GetEventsByClassID(classID int) ([]entity.Event, error)
	GetEventsByID(eventID int) (*entity.Event, error)
	DeleteEvent(eventID int) error
	GetEvents() ([]entity.Event, error)
	AddPlayersToEvent(eventID int, playerIDs []int, login string) error
	DeletePlayersFromEvent(eventID int, playerIDs []int) error
	GetEventPlayers(eventID int) ([]userModels.SafeUser, error)
	GetEventPlayersCount(eventID int) (int, error)
	EventComplete(eventID int, ratingReward int, classReward int) error
	AddEventParams(eventID int, params *entity.EventParams) error
	GetEventParams(eventID int) ([]entity.EventParams, error)
	DeleteEventParams(eventID int) error
	UpdateEventParams(eventID int, params *entity.EventParams) error
	UpdateEvent(eventID int, updatedEvent *entity.Event) error
}
