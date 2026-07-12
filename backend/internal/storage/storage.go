package storage

import (
	"database/sql"
	"sync"

	classConfig "cspirt/internal/class/config"
	classRepo "cspirt/internal/class/repo"
	complaintRepo "cspirt/internal/complaints/repo"
	eventsRepo "cspirt/internal/events/repo"
	noteRepo "cspirt/internal/note/repo"
	ratingRepo "cspirt/internal/rating/repo"
)

type Storage struct {
	db *sql.DB
	mu sync.Mutex
	
	RatingRepo     ratingRepo.RatingRepository
	NotesRepo      noteRepo.NoteRepository
	ComplaintsRepo complaintRepo.ComplaintRepository
	ClassRepo      classRepo.ClassRepository
	EventsRepo     eventsRepo.EventsRepository

	Secret string

	ParallelsConfig []classConfig.ParallelConfig
}

func NewUserStorage(path string, jwtSecret string) (*Storage, error) {
	db, err := openPostgres(path)
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
