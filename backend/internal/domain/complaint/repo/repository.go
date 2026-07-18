package repo

import (
	models "cspirt/internal/domain/user"
	"context"
)

type ComplaintRepository interface {
	GetAllComplaints(ctx context.Context) ([]models.Complaint, error)
	AddComplaint(ctx context.Context, login string, complaint models.Complaint, user models.SafeUser) error
	DeleteComplaint(ctx context.Context, id int, user models.SafeUser) error
	GetComplaintByID(ctx context.Context, id int) ([]models.Complaint, error)
	GetComplaintsByUserId(ctx context.Context, User_id int) ([]models.Complaint, error)
	GetComplaintsByClassID(ctx context.Context, classID int) ([]models.Complaint, error)
}
