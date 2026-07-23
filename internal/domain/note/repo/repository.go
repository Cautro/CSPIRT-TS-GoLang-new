package repo

import (
	models "cspirt/internal/domain/user" 
	"context"
)

type NoteRepository interface {
	GetAllNotes(ctx context.Context) ([]models.Note, error)
	AddNote(ctx context.Context, login string, note models.Note, user models.SafeUser) error
	DeleteNote(ctx context.Context, id int, user models.SafeUser) error
	GetNoteByID(ctx context.Context, id int) ([]models.Note, error)
	GetNotesByUserId(ctx context.Context, userId int) ([]models.Note, error)
	GetNotesByClassID(ctx context.Context, classID int) ([]models.Note, error)
}