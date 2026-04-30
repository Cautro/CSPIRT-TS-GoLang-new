package complaints

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"cspirt/internal/repo"
	"errors"
)

type ComplaintService struct {
	complaints repo.ComplaintRepository
}

func NewComplaintsService(complaints repo.ComplaintRepository, jwtSecret string) *ComplaintService {
	return &ComplaintService{
		complaints: complaints,
	}
}

func (s *ComplaintService) GetAllComplaints() ([]models.Complaint, error) {
	result, err := s.complaints.GetAllComplaints()

	if err != nil || result == nil {
		writeLog(logger.LogEntry{
			Level:   "Error",
			Action:  "getting_all_complaints",
			Message: "Error by getting all complaints",
		})
		return []models.Complaint{}, nil
	}

	return result, nil
}

func (s *ComplaintService) AddNewComplaint(login string, in *models.AddNewComplaintResponse, user *models.SafeUser) error {
	result := s.complaints.AddComplaint(login, models.Complaint{
		ID:        in.ID,
		TargetID:  in.TargetID,
		AuthorID:  in.AuthorID,
		Content:   in.Content,
		CreatedAt: in.CreatedAt,
	}, *user)

	if result == nil {
		return errors.New("failed to create new complaint")
	}

	return nil
}

func (s *ComplaintService) DeleteComplaint(id int, user models.SafeUser) error {
	err := s.complaints.DeleteComplaint(id, user)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "delete_complaint",
			Message: "Error to delete complaint",
		})
		return errors.New("server error")
	}

	return nil
}