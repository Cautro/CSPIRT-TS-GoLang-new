package notes

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"cspirt/internal/repo"
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

	if err != nil || result == nil {
		writeLog(logger.LogEntry{
			Level: "Error",
			Action: "getting_all_notes",
			Message: "Error by getting all notes",
		})
		return []models.Note{}, nil
	}

	return result, nil
}

func (s *NoteService) AddNewNote(login string, in *models.AddNewNoteResponse, user *models.SafeUser) (error) {
	result := s.notes.AddNote(login, models.Note{
		ID: in.ID,
		TargetID: in.TargetID,
		AuthorID: in.AuthorID,
		Content: in.Content,
		CreatedAt: in.CreatedAt,
	}, *user)

	if result == nil {
		return errors.New("Failed to create new note")
	}

	return nil
}

func (s *NoteService) DeleteNote(id int, user models.SafeUser) error {
	err := s.notes.DeleteNote(id, user)
	if err != nil {
		writeLog(logger.LogEntry{
			Level: "info",
			Action: "delete_note",
			Message: "Error to delete note",
		})
		return errors.New("Server error")
	}
	return nil
}