package notes

import (
	"cspirt/internal/logger"
	noteModels "cspirt/internal/note/models"
	"cspirt/internal/note/repo"
	"cspirt/internal/users/models"
	"errors"
	"time"
)

type NoteService struct {
	notes repo.NoteRepository
}

func NewNoteService(notes repo.NoteRepository, jwtSecret string) *NoteService {
	return &NoteService{
		notes: notes,
	}
}

func (s *NoteService) GetAllNotes() ([]models.Note, error) {
	result, err := s.notes.GetAllNotes()

	if err != nil {
		writeLog(logger.LogEntry{
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

func (s *NoteService) GetNotesByClassID(classID int) ([]models.Note, error) {
	if classID <= 0 {
		return nil, errors.New("invalid class id")
	}

	result, err := s.notes.GetNotesByClassID(classID)
	if err != nil {
		writeLog(logger.LogEntry{
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

func (s *NoteService) AddNewNote(login string, in *noteModels.AddNewNoteResponse, user *models.SafeUser) error {
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

	err := s.notes.AddNote(login, models.Note{
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

func (s *NoteService) DeleteNote(id int, user models.SafeUser) error {
	err := s.notes.DeleteNote(id, user)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "delete_note",
			Message: "Error to delete note",
		})
		return errors.New("server error")
	}
	return nil
}
