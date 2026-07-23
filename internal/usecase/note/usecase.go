package usecase

import (
	"cspirt/pkg/logger"
	noteModels "cspirt/internal/domain/note"
	"cspirt/internal/domain/note/repo"
	models "cspirt/internal/domain/user"
	"errors"
	"time"
	"context"
)

type NoteUsecase struct {
	notes repo.NoteRepository
}

func NewNoteUsecase(notes repo.NoteRepository) *NoteUsecase {
	return &NoteUsecase{
		notes: notes,
	}
}

func (s *NoteUsecase) GetAllNotes(ctx context.Context) ([]models.Note, error) {
	result, err := s.notes.GetAllNotes(ctx)

	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "getting_all_notes",
			Message: "Error by getting all notes",
		})
		return nil, err
	}
	if result == nil {
		return []models.Note{}, nil
	}

	return result, nil
}

func (s *NoteUsecase) GetNotesByClassID(ctx context.Context, classID int) ([]models.Note, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	result, err := s.notes.GetNotesByClassID(ctx, classID)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "getting_notes_by_class",
			Message: "Error by getting notes by class",
		})
		return nil, err
	}

	if result == nil {
		return []models.Note{}, nil
	}

	return result, nil
}

func (s *NoteUsecase) AddNewNote(ctx context.Context, login string, in *noteModels.AddNewNoteResponse, user *models.SafeUser) error {
	if in == nil {
		return errors.New("invalid input")
	}
	if user == nil {
		return errors.New("user not found")
	}
	if in.TargetID <= 0 {
		return errors.New("target is required")
	}
	if in.Content == "" {
		return errors.New("content is required")
	}

	authorName := user.Name + " " + user.LastName

	err := s.notes.AddNote(ctx, login, models.Note{
		TargetID:  in.TargetID,
		AuthorID:  user.ID,
		AuthorName: authorName,
		TargetName: in.TargetName,
		Content:   in.Content,
		CreatedAt: time.Now(),
	}, *user)

	if err != nil {
		return err
	}

	return nil
}

func (s *NoteUsecase) DeleteNote(ctx context.Context, id int, user models.SafeUser) error {
	err := s.notes.DeleteNote(ctx, id, user)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_note",
			Message: "Error to delete note",
		})
		return errors.New("server error")
	}
	return nil
}
