package repo

import (
	userModels "cspirt/internal/domain/user"
	entity "cspirt/internal/domain/event"
	"context"
)

type EventsRepository interface {
	ActivateDueEvents() error
	AddEvent(ctx context.Context, event entity.Event) error
	GetEventsByUserID(ctx context.Context, userID int) ([]entity.Event, error)
	GetEventsByClassID(ctx context.Context, classID int) ([]entity.Event, error)
	GetEventsByID(ctx context.Context, eventID int) (*entity.Event, error)
	DeleteEvent(ctx context.Context, eventID int) error
	GetEvents(ctx context.Context) ([]entity.Event, error)
	AddPlayersToEvent(ctx context.Context, eventID int, playerIDs []int, login string) error
	DeletePlayersFromEvent(ctx context.Context, eventID int, playerIDs []int) error
	GetEventPlayers(ctx context.Context, eventID int) ([]userModels.SafeUser, error)
	GetEventPlayersCount(ctx context.Context, eventID int) (int, error)
	EventComplete(ctx context.Context, eventID int, ratingReward int, classReward int) error
	AddEventParams(ctx context.Context, eventID int, params *entity.EventParams) error
	GetEventParams(ctx context.Context, eventID int) ([]entity.EventParams, error)
	DeleteEventParams(ctx context.Context, eventID int) error
	UpdateEventParams(ctx context.Context, eventID int, params *entity.EventParams) error
	UpdateEvent(ctx context.Context, eventID int, updatedEvent *entity.Event) error
}
