package usecase

import (
	complaintModels "cspirt/internal/domain/complaint"
	"cspirt/internal/domain/complaint/repo"
	"cspirt/pkg/logger"
	userModels "cspirt/internal/domain/user"
	"errors"
	"time"
	"context"
)

type ComplaintUsecase struct {
	complaints repo.ComplaintRepository
}

func NewComplaintsUsecase(complaints repo.ComplaintRepository) *ComplaintUsecase {
	return &ComplaintUsecase{
		complaints: complaints,
	}
}

func (s *ComplaintUsecase) GetComplaintsByClassID(ctx context.Context, classID int) ([]userModels.Complaint, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	result, err := s.complaints.GetComplaintsByClassID(ctx, classID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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

func (s *ComplaintUsecase) GetAllComplaints(ctx context.Context) ([]userModels.Complaint, error) {
	result, err := s.complaints.GetAllComplaints(ctx)

	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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

func (s *ComplaintUsecase) AddNewComplaint(ctx context.Context, login string, in *complaintModels.AddNewComplaintResponse, user *userModels.SafeUser) error {
	if in == nil {
		return errors.New("invalid input")
	}
	if user == nil {
		return errors.New("user not found")
	}

	err := s.complaints.AddComplaint(ctx, login, userModels.Complaint{
		ID:        in.ID,
		TargetID:  in.TargetID,
		Content:   in.Content,
		CreatedAt: time.Now(),
	}, *user)

	if err != nil {
		return err
	}

	return nil
}

func (s *ComplaintUsecase) DeleteComplaint(ctx context.Context, id int, user userModels.SafeUser) error {
	err := s.complaints.DeleteComplaint(ctx, id, user)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_complaint",
			Message: "Error to delete complaint",
		})
		return errors.New("server error")
	}

	return nil
}
