package repo

import (
	"context"
	userEntity "cspirt/internal/domain/user"
	entity "cspirt/internal/domain/moderate"
)

type ModerateRepo interface {
	CheckToWaitStatus(id int) (bool, error)

	GetAllWaitNotes(ctx context.Context) ([]userEntity.Note, error)
	GetAllWaitComplaints(ctx context.Context) ([]userEntity.Complaint, error)
	UpdateNoteModerateStatus(ctx context.Context, id int, input entity.ModerateDTO) error
	UpdateComplaintModerateStatus(ctx context.Context, id int, input entity.ModerateDTO) error 
}