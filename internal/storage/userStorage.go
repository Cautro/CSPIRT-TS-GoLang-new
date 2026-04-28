package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	utils "cspirt/internal/utils/auth"
	"database/sql"
	"encoding/json"
	"errors"
)

func (s *Storage) AddUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.Login == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Message: "login is required",
		})
		return errors.New("login is required")
	} else if user.Password == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Message: "password is required",
		})
		return errors.New("password is required")
	} else if user.Role == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Message: "role is required",
		})
		return errors.New("role is required")
	} else if user.Role != "admin" && user.Role != "owner" && user.Role != "user" && user.Role != "helper" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "invalid role",
		})
		return errors.New("invalid role")
	} else if user.Class == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "class is required",
		})
		return errors.New("class is required")
	} else if user.Rating >= 5000 {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "rating must be less than 5000",
		})
		return errors.New("rating must be less than 5000")
	} else if user.Rating < 0 {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "rating must be non-negative",
		})
		return errors.New("rating must be non-negative")
	} else if user.Rating == 0 {
		user.Rating = 500
	} else if user.FullName == nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "full name is required",
		})
		return errors.New("full name is required")
	} else if user.Name == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "name is required",
		})
		return errors.New("name is required")
	} else if user.LastName == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "last name is required",
		})
		return errors.New("last name is required")
	}

	fullNameJSON, err := json.Marshal(user.FullName)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to marshal full name: " + err.Error(),
		})
		return err
	}

	notesJSON, err := json.Marshal(user.Notes)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to marshal notes: " + err.Error(),
		})
		return err
	}

	complaintsJSON, err := json.Marshal(user.Complaints)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to marshal complaints: " + err.Error(),
		})
		return err
	}

	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to hash password: " + err.Error(),
		})
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
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to insert user: " + err.Error(),
		})
	}

	return err
}

func (s *Storage) SaveUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "save_user",
		Login:   user.Login,
		Role:    user.Role,
		Message: "saving user",
	})

	notesJSON, err := json.Marshal(user.Notes)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "save_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to marshal notes: " + err.Error(),
		})
		return err
	}
	complaintsJSON, err := json.Marshal(user.Complaints)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "save_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to marshal complaints: " + err.Error(),
		})
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
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "save_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to save user: " + err.Error(),
		})
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "save_user",
		Login:   user.Login,
		Role:    user.Role,
		Message: "user saved",
	})
	return err
}

func (s *Storage) UpdateUser(user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "update_user",
		Login:   user.Login,
		Role:    user.Role,
		Message: "updating user",
	})

	existingUser, err := s.GetUserByLogin(user.Login)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "update_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to get existing user: " + err.Error(),
		})
		return err
	}

	if existingUser == nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "update_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "user not found for update",
		})
		return errors.New("user not found")
	}

	notesJSON, err := json.Marshal(user.Notes)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "update_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to marshal notes: " + err.Error(),
		})
		return err
	}
	complaintsJSON, err := json.Marshal(user.Complaints)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "update_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to marshal complaints: " + err.Error(),
		})
		return err
	}

	query := `UPDATE users SET Name = ?,
	LastName = ?,
	Login = ?,
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
		user.Rating,
		user.Role,
		user.Class,
		string(notesJSON),
		string(complaintsJSON),
		user.ID,
	)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "update_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to update user: " + err.Error(),
		})
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "update_user",
		Login:   user.Login,
		Role:    user.Role,
		Message: "user updated",
	})
	return err
}

func (s *Storage) DeleteUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_user",
		Login:   user.Login,
		Role:    user.Role,
		Message: "deleting user",
	})

	query := `DELETE FROM users WHERE Id = ?`
	_, err := s.db.Exec(query, user.ID)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "delete_user",
			Login:   user.Login,
			Role:    user.Role,
			Message: "failed to delete user: " + err.Error(),
		})
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_user",
		Login:   user.Login,
		Role:    user.Role,
		Message: "user deleted",
	})
	return err
}

func (s *Storage) GetAllUsers() ([]models.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_users",
		Message: "getting all users",
	})

	rows, err := s.db.Query(`
		SELECT Id, Name, LastName, Login, Rating, Role, Class, Notes, Complaints
		FROM users
	`)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_users",
			Message: "failed to query users: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	users := make([]models.SafeUser, 0)
	for rows.Next() {
		var u models.SafeUser
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
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_users",
				Message: "failed to scan user: " + err.Error(),
			})
			return nil, err
		}

		if notesJSON != "" {
			if err := json.Unmarshal([]byte(notesJSON), &u.Notes); err != nil {
				writeLog(logger.LogEntry{
					Level:   "error",
					Action:  "get_all_users",
					Login:   u.Login,
					Role:    u.Role,
					Message: "failed to unmarshal notes: " + err.Error(),
				})
				return nil, err
			}
		}

		if complaintsJSON != "" {
			if err := json.Unmarshal([]byte(complaintsJSON), &u.Complaints); err != nil {
				writeLog(logger.LogEntry{
					Level:   "error",
					Action:  "get_all_users",
					Login:   u.Login,
					Role:    u.Role,
					Message: "failed to unmarshal complaints: " + err.Error(),
				})
				return nil, err
			}
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_users",
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return users, nil
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
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_user_by_login",
			Login:   login,
			Message: "failed to scan user: " + err.Error(),
		})
		return nil, err
	}

	if fullNameJSON != "" {
		if err := json.Unmarshal([]byte(fullNameJSON), &u.FullName); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_user_by_login",
				Login:   login,
				Role:    u.Role,
				Message: "failed to unmarshal full name: " + err.Error(),
			})
			return nil, err
		}
	}
	if notesJSON != "" {
		if err := json.Unmarshal([]byte(notesJSON), &u.Notes); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_user_by_login",
				Login:   login,
				Role:    u.Role,
				Message: "failed to unmarshal notes: " + err.Error(),
			})
			return nil, err
		}
	}
	if complaintsJSON != "" {
		if err := json.Unmarshal([]byte(complaintsJSON), &u.Complaints); err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_user_by_login",
				Login:   login,
				Role:    u.Role,
				Message: "failed to unmarshal complaints: " + err.Error(),
			})
			return nil, err
		}
	}

	return &u, nil
}


func (s *Storage) GetUsersByClass(class string) ([]models.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT Id, Name, LastName, Login, Rating, Role, Class, Notes, Complaints
		FROM users
		WHERE Class = ?
	`, class)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_users_by_class",
			Class:   class,
			Message: "failed to query users: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	users := make([]models.SafeUser, 0)
	for rows.Next() {
		var u models.SafeUser
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
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_users_by_class",
				Class:   class,
				Message: "failed to scan user: " + err.Error(),
			})
			return nil, err
		}

		if notesJSON != "" {
			if err := json.Unmarshal([]byte(notesJSON), &u.Notes); err != nil {
				writeLog(logger.LogEntry{
					Level:   "error",
					Action:  "get_users_by_class",
					Class:   class,
					Login:   u.Login,
					Role:    u.Role,
					Message: "failed to unmarshal notes: " + err.Error(),
				})
				return nil, err
			}
		}

		if complaintsJSON != "" {
			if err := json.Unmarshal([]byte(complaintsJSON), &u.Complaints); err != nil {
				writeLog(logger.LogEntry{
					Level:   "error",
					Action:  "get_users_by_class",
					Class:   class,
					Login:   u.Login,
					Role:    u.Role,
					Message: "failed to unmarshal complaints: " + err.Error(),
				})
				return nil, err
			}
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_users_by_class",
			Class:   class,
			Message: "row iteration failed: " + err.Error(),
		})
		return nil, err
	}

	return users, nil
}