package repo

import (
	entity "cspirt/internal/domain/globalEvent"
	"context"
)

type GlobalEventRepo interface {
	GetAllGlobalEvents(ctx context.Context) ([]entity.GlobalEventOutput, error)
	AddInfoGlobalEvent(ctx context.Context, input entity.GlobalEventInfoDTO) error
	AddQuizGlobalEvent(ctx context.Context, input entity.GlobalEventQuizDTO) error
	DeleteInfoGlobalEvent(ctx context.Context, id int) error 
	DeleteQuizGlobalEvent(ctx context.Context, id int) error
	Vote(ctx context.Context, idUser, eventId, voteItemId int) error
}