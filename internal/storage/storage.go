package storage

import (
	classRepo "cspirt/internal/class/repo"
	complaintRepo "cspirt/internal/complaints/repo"
	eventsRepo "cspirt/internal/events/repo"
	noteRepo "cspirt/internal/note/repo"
	ratingRepo "cspirt/internal/rating/repo"

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
	EventsRepo     eventsRepo.EventsRepository

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

	if err := s.ensureColumn("users", "ClassID", "INTEGER"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_id ON users(ClassID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_role_id ON users(ClassID, Role, Id);`); err != nil {
		return err
	}

	return nil
}

func (s *Storage) initEventsStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventsQuery := `
	CREATE TABLE IF NOT EXISTS events (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Title TEXT NOT NULL,
		Status TEXT NOT NULL,
		RatingReward INTEGER NOT NULL DEFAULT 0,
		Description TEXT NOT NULL,
		CreatedAt TEXT NOT NULL,
		StartedAt TEXT NOT NULL,
		Players TEXT NOT NULL,
		Classes TEXT NOT NULL DEFAULT '[]'
	);`

	if _, err := s.db.Exec(eventsQuery); err != nil {
		return err
	}
	if err := s.ensureColumn("events", "RatingReward", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn("events", "Classes", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		return err
	}

	eventPlayersQuery := `
	CREATE TABLE IF NOT EXISTS event_players (
		event_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		PRIMARY KEY (event_id, player_id),
		FOREIGN KEY (event_id) REFERENCES events(Id) ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES users(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(eventPlayersQuery); err != nil {
		return err
	}

	eventClassesQuery := `
	CREATE TABLE IF NOT EXISTS event_classes (
		event_id INTEGER NOT NULL,
		class_id INTEGER NOT NULL,
		PRIMARY KEY (event_id, class_id),
		FOREIGN KEY (event_id) REFERENCES events(Id) ON DELETE CASCADE,
		FOREIGN KEY (class_id) REFERENCES classes(Id) ON DELETE CASCADE
	);`

	if _, err := s.db.Exec(eventClassesQuery); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_event_players_player_id ON event_players(player_id);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_event_classes_class_id ON event_classes(class_id);`); err != nil {
		return err
	}

	return nil
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

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_notes_target_id ON notes(TargetID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_notes_author_id ON notes(AuthorID);`); err != nil {
		return err
	}

	return nil
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

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);`); err != nil {
		return err
	}

	return nil
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

	if _, err := s.db.Exec(query); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_complaints_target_id ON complaints(TargetID);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_complaints_author_id ON complaints(AuthorID);`); err != nil {
		return err
	}

	return nil
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

	if err := configureSQLite(db); err != nil {
		db.Close()
		return nil, err
	}

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

func configureSQLite(db *sql.DB) error {
	pragmas := []string{
		`PRAGMA foreign_keys = ON`,
		`PRAGMA journal_mode = WAL`,
		`PRAGMA synchronous = NORMAL`,
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA temp_store = MEMORY`,
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return err
		}
	}

	return nil
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
