package repo

import "cspirt/internal/models"

type NoteRepository interface {
	GetAllNotes() ([]models.Note, error)
	AddNote(login string, note models.Note, user models.SafeUser) error
	DeleteNote(id int, user models.SafeUser) error
	GetNoteByID(id int) ([]models.Note, error)
	GetNotesByUserId(User_id int) ([]models.Note, error)
}