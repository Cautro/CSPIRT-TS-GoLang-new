package notes

import (
	"cspirt/internal/logger"
	"cspirt/internal/users/models"
	noteModels "cspirt/internal/note/models"
	"cspirt/internal/note/repo"
	"errors"
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

	err := s.notes.AddNote(login, models.Note{
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
