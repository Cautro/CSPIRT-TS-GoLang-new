package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	classRepo "cspirt/internal/class/repo"
	complaintRepo "cspirt/internal/complaints/repo"
	eventsRepo "cspirt/internal/events/repo"
	noteRepo "cspirt/internal/note/repo"
	ratingRepo "cspirt/internal/rating/repo"
)

type Storage struct {
	db *sql.DB
	mu sync.Mutex

	// Repositories expose this SQLite adapter through feature-owned interfaces.
	// Services depend on these interfaces instead of depending on SQL directly.
	RatingRepo     ratingRepo.RatingRepository
	NotesRepo      noteRepo.NoteRepository
	ComplaintsRepo complaintRepo.ComplaintRepository
	ClassRepo      classRepo.ClassRepository
	EventsRepo     eventsRepo.EventsRepository

	Secret string
}

// NewUserStorage opens the SQLite adapter and initializes every schema module.
//
// The name is kept for compatibility with existing handlers/tests; new code
// should treat the returned value as the application storage adapter.
func NewUserStorage(path string, jwtSecret string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := openSQLite(path)
	if err != nil {
		return nil, err
	}

	st := &Storage{
		db:     db,
		mu:     sync.Mutex{},
		Secret: jwtSecret,
	}
	st.bindRepositories()

	if err := st.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return st, nil
}

func (s *Storage) bindRepositories() {
	s.RatingRepo = s
	s.NotesRepo = s
	s.ComplaintsRepo = s
	s.ClassRepo = s
	s.EventsRepo = s
}

func (s *Storage) Close() error {
	return s.db.Close()
}
