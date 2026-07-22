package usecase

import (
	"context"
	entity "cspirt/internal/domain/globalEvent"
	repo "cspirt/internal/domain/globalEvent/repo"
	"time"
)

type GlobalEventUsecase struct {
	globalEventRepo repo.GlobalEventRepo
}

var Timeout = 5*time.Second

func NewGlobalEventUsecase(globalEventRepo repo.GlobalEventRepo) *GlobalEventUsecase {
	return &GlobalEventUsecase{
		globalEventRepo: globalEventRepo,
	}
}

func (u *GlobalEventUsecase) GetAllGlobalEvents(ctx context.Context) ([]entity.GlobalEventOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	output, err := u.globalEventRepo.GetAllGlobalEvents(ctx); if err != nil { return []entity.GlobalEventOutput{}, err }
	return output, nil
}

func (u *GlobalEventUsecase) AddInfoGlobalEvent(ctx context.Context, input entity.GlobalEventInfoDTO) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	if err := u.globalEventRepo.AddInfoGlobalEvent(ctx, input); err != nil { return err }
    return nil
}

func (u *GlobalEventUsecase) AddQuizGlobalEvent(ctx context.Context, input entity.GlobalEventQuizDTO) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	if err := u.globalEventRepo.AddQuizGlobalEvent(ctx, input); err != nil { return err }

	return nil
}

func (u *GlobalEventUsecase) DeleteInfoGlobalEvent(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	if err := u.globalEventRepo.DeleteInfoGlobalEvent(ctx, id); err != nil { return err }

	return nil
}

func (u *GlobalEventUsecase) DeleteQuizGlobalEvent(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	if err := u.globalEventRepo.DeleteQuizGlobalEvent(ctx, id); err != nil { return err }

	return nil
}

func (u *GlobalEventUsecase) Vote(ctx context.Context, idUser, idQuiz, idVoteItem int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := u.globalEventRepo.Vote(ctx, idUser, idQuiz, idVoteItem); err != nil { return err }

	return nil
}