package repo

import "cspirt/internal/models"

type NoteRepository interface {
	GetAllNotes() ([]models.Note, error)
	AddNote(login string, note models.Note, user models.SafeUser) error
}