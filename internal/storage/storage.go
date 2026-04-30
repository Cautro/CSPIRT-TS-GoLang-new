package storage

import (
	"cspirt/internal/repo"

	_ "modernc.org/sqlite"

	"database/sql"
	"os"
	"path/filepath"
	"sync"
)

type Storage struct {
	db *sql.DB
	mu sync.Mutex

	RatingRepo repo.RatingRepository
	NotesRepo repo.NoteRepository
	ComplaintsRepo repo.ComplaintRepository

	Secret string
}

func (s *Storage) initUserStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS users (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Name TEXT,
		FullName TEXT,
		LastName TEXT,
		Login TEXT UNIQUE,
		Password TEXT,
		Rating INTEGER,
		Role TEXT,
		Class TEXT
	);`

	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) initNoteStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS notes (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		TargetID INTEGER,
		AuthorID INTEGER,
		Content TEXT,
		CreatedAt TEXT
	);`

	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) initHTTPOnlyStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(Id) ON DELETE CASCADE
	);`

	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) initComplaintStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS complaints (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		TargetID INTEGER,
		AuthorID INTEGER,
		Content TEXT,
		CreatedAt TEXT
	);`

	_, err := s.db.Exec(query)
	return err
}

func NewUserStorage(path string, jwt_secret string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	st := &Storage{
		db:     db,
		mu:     sync.Mutex{},
		Secret: jwt_secret,
	}

	if err := st.initUserStorage(); err != nil { return nil, err }
    if err := st.initNoteStorage(); err != nil { return nil, err }
    if err := st.initComplaintStorage(); err != nil { return nil, err }
	if err := st.initHTTPOnlyStorage(); err != nil { return nil, err }

	return st, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
