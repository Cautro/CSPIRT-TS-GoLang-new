package moderate

import (
	"context"
	"cspirt/internal/domain/moderate/repo"
	"errors"
	"time"

	permission "cspirt/internal/controller/permission/usecase"
	entity "cspirt/internal/domain/moderate"
	ratingEntity "cspirt/internal/domain/rating"
	userEntity "cspirt/internal/domain/user"
)

type ModerateUsecase struct {
	moderate repo.ModerateRepo
}

func NewModerateUsecase(moderate repo.ModerateRepo) *ModerateUsecase {
	return &ModerateUsecase{
		moderate: moderate,
	}
}

func (u *ModerateUsecase) GetAllWaitNotes(ctx context.Context) ([]userEntity.Note, error) {
	return u.moderate.GetAllWaitNotes(ctx)
}

func (u *ModerateUsecase) GetAllWaitComplaints(ctx context.Context) ([]userEntity.Complaint, error) {
	return u.moderate.GetAllWaitComplaints(ctx)
}

func (u *ModerateUsecase) UpdateNoteModerateWait(ctx context.Context, login string, id int, input entity.ModerateDTO, permission permission.Usecase) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	checkStatus, err := u.moderate.CheckToWaitStatus(id)
	if err != nil {
		return err
	} else if !checkStatus {
		return errors.New("For bidden")
	}

	checkRole := permission.CheckUserRole(ctx, login, string(ratingEntity.RoleOwner), string(ratingEntity.RoleAdmin))
	if checkRole != nil {
		return checkRole
	}

	if err := u.moderate.UpdateNoteModerateStatus(ctx, id, input); err != nil {
		return err
	}
	return nil
}

func (u *ModerateUsecase) UpdateComplaintModerateWait(ctx context.Context, login string, id int, input entity.ModerateDTO, permission permission.Usecase) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	checkStatus, err := u.moderate.CheckToWaitStatus(id)
	if err != nil {
		return err
	} else if !checkStatus {
		return errors.New("For bidden")
	}

	checkRole := permission.CheckUserRole(ctx, login, string(ratingEntity.RoleOwner), string(ratingEntity.RoleAdmin))
	if checkRole != nil {
		return checkRole
	}

	if err := u.moderate.UpdateComplaintModerateStatus(ctx, id, input); err != nil {
		return err
	}
	return nil
}