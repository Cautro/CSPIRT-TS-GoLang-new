package storage

import (
	"cspirt/internal/models"
	_ "modernc.org/sqlite"

	// "bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"log/slog"
)

type Storage struct {
	db  *sql.DB
	log *slog.Logger
	mu  sync.Mutex

	Secret  string
}

func (s *Storage) init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS users (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Name TEXT,
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

func NewStorage(path string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	st := &Storage{db: db}
	if err := st.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return st, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) GetAllUsers() ([]models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Getting all users")

	rows, err := s.db.Query(`
		SELECT Id, Name, LastName, Login, Password, Rating, Role, Class, Notes, Complaints
		FROM users
	`)
	if err != nil {
		s.log.Error("failed to query users", "error", err)
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		var notesJSON string
		var complaintsJSON string

		if err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.LastName,
			&u.Login,
			&u.Password,
			&u.Rating,
			&u.Role,
			&u.Class,
			&notesJSON,
			&complaintsJSON,
		); err != nil {
			s.log.Error("failed to scan user", "error", err)
			return nil, err
		}

		if notesJSON != "" {
			if err := json.Unmarshal([]byte(notesJSON), &u.Notes); err != nil {
				s.log.Error("failed to unmarshal notes", "login", u.Login, "error", err)
				return nil, err
			}
		}

		if complaintsJSON != "" {
			if err := json.Unmarshal([]byte(complaintsJSON), &u.Complaints); err != nil {
				s.log.Error("failed to unmarshal complaints", "login", u.Login, "error", err)
				return nil, err
			}
		}

		users = append(users, u)
	}

	return users, rows.Err()
}

func (s *Storage) AddUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Adding user", "login", user.Login)

	notesJSON, err := json.Marshal(user.Notes)
	if err != nil {
		s.log.Error("failed to marshal notes", "login", user.Login, "error", err)
		return err
	}
	complaintsJSON, err := json.Marshal(user.Complaints)
	if err != nil {
		s.log.Error("failed to marshal complaints", "login", user.Login, "error", err)
		return err
	}

	query := `INSERT INTO users (Name, 
	LastName, 
	Login,
	Password,
	Rating,
	Role,
	Class,
	Notes,
	Complaints) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query,
		user.Name, 
        user.LastName, 
        user.Login, 
        user.Password, 
        user.Rating, 
        user.Role, 
        user.Class, 
		string(notesJSON),
        string(complaintsJSON),
	)

	if err != nil {
		s.log.Error("failed to insert user", "login", user.Login, "error", err)
		return err
	}

	s.log.Info("user added", "login", user.Login)
	return nil
}

func (s *Storage) SaveUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Saving user", "login", user.Login)

	notesJSON, err := json.Marshal(user.Notes)
	if err != nil {
		s.log.Error("failed to marshal notes", "login", user.Login, "error", err)
		return err
	}
	complaintsJSON, err := json.Marshal(user.Complaints)
	if err != nil {
		s.log.Error("failed to marshal complaints", "login", user.Login, "error", err)
		return err
	}

	query := `UPDATE users SET Name = ?,
	LastName = ?,
	Login = ?,
	Password = ?,
	Rating = ?,
	Role = ?,
	Class = ?,
	Notes = ?,
	Complaints = ?
	WHERE Id = ?`

	_, err = s.db.Exec(query,
		user.Name,
		user.LastName,
		user.Login,
		user.Password,
		user.Rating,
		user.Role,
		user.Class,
		string(notesJSON),
		string(complaintsJSON),
		user.ID,
	)
	if err != nil {
		s.log.Error("failed to update user", "login", user.Login, "error", err)
		return err
	}

	s.log.Info("user saved", "login", user.Login)
	return err
}

func (s *Storage) DeleteUser(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Deleting user", "id", id)

	query := `DELETE FROM users WHERE Id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		s.log.Error("failed to delete user", "id", id, "error", err)
		return err
	}

	s.log.Info("user deleted", "id", id)
	return err
}

func (s *Storage) ReadUsers() ([]models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Reading all users")

	rows, err := s.db.Query("SELECT Id, Name, LastName, Login, Password, Rating, Role, Class FROM users")
	if err != nil {
		s.log.Error("failed to query users", "error", err)
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user []byte
		if err := rows.Scan(&user); err != nil {
			s.log.Error("failed to scan user", "error", err)
			return nil, err
		}

		for rows.Next() {
			var u models.User
			if err := rows.Scan(
				&u.ID,
				&u.Name,
				&u.LastName,
				&u.Login,
				&u.Password,
				&u.Rating,
				&u.Role,
				&u.Class,
			); err != nil {
				s.log.Error("failed to scan user", "error", err)
				return nil, err
			}
			users = append(users, u)
		}
	}

	return users, rows.Err()
}