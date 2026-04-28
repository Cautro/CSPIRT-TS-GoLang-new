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

	Secret string
}

func (s *Storage) init() error {
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
		Class TEXT,
		Complaints TEXT,
		Notes TEXT
	);`

	_, err := s.db.Exec(query)
	return err
}

func NewStorage(path string, jwt_secret string) (*Storage, error) {
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

	if err := st.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return st, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
