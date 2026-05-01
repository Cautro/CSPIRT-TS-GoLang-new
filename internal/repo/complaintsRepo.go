package repo

import "cspirt/internal/models"

type ComplaintRepository interface {
	GetAllComplaints() ([]models.Complaint, error)
	AddComplaint(login string, complaint models.Complaint, user models.SafeUser) error
	DeleteComplaint(id int, user models.SafeUser) error
	GetComplaintByID(id int) ([]models.Complaint, error)
	GetComplaintsByUserId(User_id int) ([]models.Complaint, error)
}
