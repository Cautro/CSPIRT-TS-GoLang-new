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

	RatingRepo     repo.RatingRepository
	NotesRepo      repo.NoteRepository
	ComplaintsRepo repo.ComplaintRepository
	ClassRepo      repo.ClassRepository

	Secret string
}

func (s *Storage) initUserStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS users (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Name TEXT NOT NULL,
		FullName TEXT NOT NULL DEFAULT '[]',
		LastName TEXT NOT NULL,
		Login TEXT NOT NULL UNIQUE,
		Password TEXT NOT NULL,
		Rating INTEGER NOT NULL DEFAULT 500,
		Role TEXT NOT NULL,
		Class TEXT NOT NULL
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
		TargetID INTEGER NOT NULL,
		AuthorID INTEGER NOT NULL,
		Content TEXT NOT NULL,
		CreatedAt TEXT NOT NULL,
		FOREIGN KEY (TargetID) REFERENCES users(Id) ON DELETE CASCADE,
		FOREIGN KEY (AuthorID) REFERENCES users(Id) ON DELETE CASCADE
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
		TargetID INTEGER NOT NULL,
		AuthorID INTEGER NOT NULL,
		Content TEXT NOT NULL,
		CreatedAt TEXT NOT NULL,
		FOREIGN KEY (TargetID) REFERENCES users(Id) ON DELETE CASCADE,
		FOREIGN KEY (AuthorID) REFERENCES users(Id) ON DELETE CASCADE
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

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(1)

	st := &Storage{
		db:     db,
		mu:     sync.Mutex{},
		Secret: jwt_secret,
	}

	st.RatingRepo = st
	st.NotesRepo = st
	st.ComplaintsRepo = st
	st.ClassRepo = st

	if err := st.initUserStorage(); err != nil {
		db.Close()
		return nil, err
	}
	if err := st.initClassStorage(); err != nil {
		db.Close()
		return nil, err
	}
	if err := st.initNoteStorage(); err != nil {
		db.Close()
		return nil, err
	}
	if err := st.initComplaintStorage(); err != nil {
		db.Close()
		return nil, err
	}
	if err := st.initHTTPOnlyStorage(); err != nil {
		db.Close()
		return nil, err
	}
	if err := st.syncAllClasses(); err != nil {
		db.Close()
		return nil, err
	}

	return st, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
