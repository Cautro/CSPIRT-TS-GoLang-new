package storage

import (
	"cspirt/internal/models"
	utils "cspirt/internal/utils/auth"
	"encoding/json"
	"database/sql"
)

func (s *Storage) AddUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.Login == "" {
		s.log.Error("login is required")
		return nil
	} else if user.Password == "" {
		s.log.Error("password is required", "login", user.Login)
		return nil
	} else if user.Role == "" {
		s.log.Error("role is required", "login", user.Login)
		return nil
	} else if user.Role != "admin" && user.Role != "owner" && user.Role != "user" && user.Role != "helper" {
		s.log.Error("invalid role", "login", user.Login, "role", user.Role)
		return nil
	} else if user.Class == "" {
		s.log.Error("class is required", "login", user.Login)
		return nil
	} else if user.Rating >= 5000 {
		s.log.Error("rating must be less than 5000", "login", user.Login, "rating", user.Rating)
		return nil
	} else if user.Rating < 0 {
		s.log.Error("rating must be non-negative", "login", user.Login, "rating", user.Rating)
		return nil
	} else if user.Rating == 0 {
		user.Rating = 500
	} else if user.FullName == nil {
		s.log.Error("full name is required", "login", user.Login)
		return nil
	} else if user.Name == "" {
		s.log.Error("name is required", "login", user.Login)
		return nil
	} else if user.LastName == "" {
		s.log.Error("last name is required", "login", user.Login)
		return nil
	}


	fullNameJSON, err := json.Marshal(user.FullName)
	if err != nil {
		return err
	}

	notesJSON, err := json.Marshal(user.Notes)
	if err != nil {
		return err
	}

	complaintsJSON, err := json.Marshal(user.Complaints)
	if err != nil {
		return err
	}

	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users
		(Name, FullName, LastName, Login, Password, Rating, Role, Class, Notes, Complaints)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(
		query,
		user.Name,
		string(fullNameJSON),
		user.LastName,
		user.Login,
		passwordHash,
		user.Rating,
		user.Role,
		user.Class,
		string(notesJSON),
		string(complaintsJSON),
	)

	return err
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

func (s *Storage) DeleteUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log.Info("Deleting user", "id", user.ID)

	query := `DELETE FROM users WHERE Id = ?`
	_, err := s.db.Exec(query, user.ID)
	if err != nil {
		s.log.Error("failed to delete user", "id", user.ID, "error", err)
		return err
	}

	s.log.Info("user deleted", "id", user.ID)
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
		SELECT Id, Name, LastName, Login, Rating, Role, Class, Notes, Complaints
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

	row := s.db.QueryRow(`
		SELECT Id, Name, FullName, LastName, Login, Password, Rating, Role, Class, Notes, Complaints
		FROM users
		WHERE Login = ?
	`, login)

	var u models.User
	var fullNameJSON string
	var notesJSON string
	var complaintsJSON string

	err := row.Scan(
		&u.ID,
		&u.Name,
		&fullNameJSON,
		&u.LastName,
		&u.Login,
		&u.Password,
		&u.Rating,
		&u.Role,
		&u.Class,
		&notesJSON,
		&complaintsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if fullNameJSON != "" {
		if err := json.Unmarshal([]byte(fullNameJSON), &u.FullName); err != nil {
			return nil, err
		}
	}
	if notesJSON != "" {
		if err := json.Unmarshal([]byte(notesJSON), &u.Notes); err != nil {
			return nil, err
		}
	}
	if complaintsJSON != "" {
		if err := json.Unmarshal([]byte(complaintsJSON), &u.Complaints); err != nil {
			return nil, err
		}
	}

	return &u, nil
}

