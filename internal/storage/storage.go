package storage

import (
	ratingRepo "cspirt/internal/rating/repo"
	noteRepo "cspirt/internal/note/repo"
	complaintRepo "cspirt/internal/complaints/repo"
	classRepo "cspirt/internal/class/repo"
	eventsRepo "cspirt/internal/events/repo"

	_ "modernc.org/sqlite"

	"database/sql"
	"os"
	"path/filepath"
	"sync"
)

type Storage struct {
	db *sql.DB
	mu sync.Mutex

	RatingRepo     ratingRepo.RatingRepository
	NotesRepo      noteRepo.NoteRepository
	ComplaintsRepo complaintRepo.ComplaintRepository
	ClassRepo      classRepo.ClassRepository
	EventsRepo	   eventsRepo.EventsRepository

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
		Class TEXT NOT NULL,
		ClassID INTEGER
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	return s.ensureColumn("users", "ClassID", "INTEGER")
}

func (s *Storage) initEventsStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS events (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Title TEXT NOT NULL,
		Status TEXT NOT NULL,
		Description TEXT NOT NULL,
		CreatedAt TEXT NOT NULL,
		StartedAt TEXT NOT NULL,
		Players TEXT NOT NULL
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
		TargetName TEXT NOT NULL,
		AuthorName TEXT NOT NULL,
		Content TEXT NOT NULL,
		CreatedAt TEXT NOT NULL
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
		TargetName TEXT NOT NULL,
		AuthorID INTEGER NOT NULL,
		AuthorName TEXT NOT NULL,
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
	st.EventsRepo = st

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
	if err := st.initEventsStorage(); err != nil {
		db.Close()
		return nil, err
	}

	return st, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) ensureColumn(table string, column string, definition string) error {
	rows, err := s.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = s.db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}
