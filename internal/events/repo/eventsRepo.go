package repo

import (
	eventModels "cspirt/internal/events/models"
	userModels "cspirt/internal/users/models"
)

type EventsRepository interface {
	AddEvent(event eventModels.Event) error
	GetEventsByUserID(userID int) ([]eventModels.Event, error)
	GetEventsByClassID(classID int) ([]eventModels.Event, error)
	GetEventsByID(eventID int) (*eventModels.Event, error)
	DeleteEvent(eventID int) error
	GetEvents() ([]eventModels.Event, error)
	AddPlayersToEvent(eventID int, playerIDs []int) error
	DeletePlayersFromEvent(eventID int, playerIDs []int) error
	GetEventPlayers(eventID int) ([]userModels.SafeUser, error)
	GetEventPlayersCount(eventID int) (int, error)
	EventComplete(eventID int, ratingReward int) error
}
