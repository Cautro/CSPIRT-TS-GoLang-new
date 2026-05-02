package complaintservice

import (
	complaintModels "cspirt/internal/complaints/models"
	"cspirt/internal/complaints/repo"
	"cspirt/internal/logger"
	userModels "cspirt/internal/users/models"
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

func (s *ComplaintService) GetComplaintsByClassID(classID int) ([]userModels.Complaint, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	result, err := s.complaints.GetComplaintsByClassID(classID)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "getting_complaints_by_class",
			Message: "Error by getting complaints by class",
		})
		return nil, err
	}

	if result == nil {
		return []userModels.Complaint{}, nil
	}

	return result, nil
}

func (s *ComplaintService) GetAllComplaints() ([]userModels.Complaint, error) {
	result, err := s.complaints.GetAllComplaints()

	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "getting_all_complaints",
			Message: "Error by getting all complaints",
		})
		return nil, err
	}
	if result == nil {
		return []userModels.Complaint{}, nil
	}

	return result, nil
}

func (s *ComplaintService) AddNewComplaint(login string, in *complaintModels.AddNewComplaintResponse, user *userModels.SafeUser) error {
	if in == nil {
		return errors.New("invalid input")
	}
	if user == nil {
		return errors.New("user not found")
	}

	err := s.complaints.AddComplaint(login, userModels.Complaint{
		ID:        in.ID,
		TargetID:  in.TargetID,
		AuthorID:  in.AuthorID,
		Content:   in.Content,
		CreatedAt: in.CreatedAt,
	}, *user)

	if err != nil {
		return err
	}

	return nil
}

func (s *ComplaintService) DeleteComplaint(id int, user userModels.SafeUser) error {
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
