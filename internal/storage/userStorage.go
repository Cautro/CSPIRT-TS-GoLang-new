package storage

import (
	"cspirt/internal/models"
	"encoding/json"
)

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

	query := `INSERT INTO users (Name, FullName, LastName, Login, Password, Rating, Role, Class, Notes, Complaints) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query,
		user.Name, 
		user.FullName,
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

	rows, err := s.db.Query("SELECT FullName, Id, Name, LastName, Login, Password, Rating, Role, Class FROM users")
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

func (s *Storage) GetUserByLogin(login string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Getting user by login", "login", login)

	row, err := s.db.Query("SELECT Id, Name, LastName, Login, Password, Rating, Role, Class, Notes, Complaints FROM users WHERE Login = ?", login)
	if err != nil {
		s.log.Warn("failed to query user by login", "login", login, "error", err)
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		var u models.User
		var notesJSON string
		var complaintsJSON string
		if err := row.Scan(
			&u.ID,
			&u.Name,
			&u.LastName,
			&u.FullName,
			&u.Login,
			&u.Password,
			&u.Rating,
			&u.Role,
			&u.Class,
			&notesJSON,
			&complaintsJSON,
		); err != nil {
			s.log.Warn("failed to scan user by login", "login", login, "error", err)
			return nil, err
		}

		if notesJSON != "" {
			if err := json.Unmarshal([]byte(notesJSON), &u.Notes); err != nil {
				s.log.Warn("failed to unmarshal notes for user", "login", login, "error", err)
				return nil, err
			}
		}

		if complaintsJSON != "" {
			if err := json.Unmarshal([]byte(complaintsJSON), &u.Complaints); err != nil {
				s.log.Warn("failed to unmarshal complaints for user", "login", login, "error", err)
				return nil, err
			}
		}
		return &u, nil
	}

	s.log.Warn("user not found", "login", login)
	return nil, nil
}