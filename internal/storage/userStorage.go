package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/users/models"
	utils "cspirt/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

func (s *Storage) AddUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trimUserInput(&user)

	if err := validateNewUser(&user); err != nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: err.Error(),
		})
		return err
	}

	role, err := normalizeRole(user.Role)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: err.Error(),
		})
		return err
	}
	user.Role = role

	fullNameJSON, err := json.Marshal(user.FullName)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "failed to marshal full name: " + err.Error(),
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
			Class:   user.Class,
			Message: "failed to hash password: " + err.Error(),
		})
		return err
	}

	if err := s.resolveUserClassLocked(&user); err != nil {
		return err
	}

	query := `
		INSERT INTO users
		(Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		user.ClassID,
	)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "failed to insert user: " + err.Error(),
		})
		return err
	}

	if err := s.syncClassByIDLocked(user.ClassID); err != nil {
		return err
	}

	return nil
}

func (s *Storage) SaveUser(user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.Login = normalizeLogin(user.Login)
	user.Name = strings.TrimSpace(user.Name)
	user.LastName = strings.TrimSpace(user.LastName)
	user.Class = normalizeClassName(user.Class)

	role, err := normalizeRole(user.Role)
	if err != nil {
		return err
	}
	user.Role = role

	if user.ID <= 0 {
		return errors.New("user id is required")
	}
	if user.Login == "" {
		return errors.New("login is required")
	}
	if user.Name == "" {
		return errors.New("name is required")
	}
	if user.LastName == "" {
		return errors.New("last name is required")
	}
	user.Rating = clampRating(user.Rating)

	oldUser, err := s.getUserByIDLocked(user.ID)
	if err != nil {
		return err
	}
	if oldUser == nil {
		return errors.New("user not found")
	}
	if err := s.resolveSafeUserClassLocked(&user); err != nil {
		return err
	}

	fullNameJSON, err := json.Marshal(user.FullName)
	if err != nil {
		return err
	}
	if string(fullNameJSON) == "null" {
		fullNameJSON = []byte("[]")
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "save_user",
		Login:   user.Login,
		Role:    user.Role,
		Class:   user.Class,
		Message: "saving user",
	})

	query := `
		UPDATE users
		SET Name = ?, FullName = ?, LastName = ?, Login = ?, Rating = ?, Role = ?, Class = ?, ClassID = ?
		WHERE Id = ?
	`

	result, err := s.db.Exec(
		query,
		user.Name,
		string(fullNameJSON),
		user.LastName,
		user.Login,
		user.Rating,
		user.Role,
		user.Class,
		user.ClassID,
		user.ID,
	)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "save_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "failed to save user: " + err.Error(),
		})
		return err
	}

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("user not found")
	}

	if err := s.syncClassByIDLocked(user.ClassID); err != nil {
		return err
	}
	if oldUser.ClassID != 0 && oldUser.ClassID != user.ClassID {
		if err := s.syncClassByIDLocked(oldUser.ClassID); err != nil {
			return err
		}
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "save_user",
		Login:   user.Login,
		Role:    user.Role,
		Class:   user.Class,
		Message: "user saved",
	})
	return nil
}

func (s *Storage) UpdateUser(user models.SafeUser) error {
	return s.SaveUser(user)
}

func (s *Storage) UpdateRating(login string, rating int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	login = normalizeLogin(login)
	if login == "" {
		return errors.New("login is required")
	}

	user, err := s.getUserByLoginLocked(login)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	rating = clampRating(rating)
	_, err = s.db.Exec(`UPDATE users SET Rating = ? WHERE Login = ?`, rating, login)
	if err != nil {
		return err
	}

	return s.syncClassByIDLocked(user.ClassID)
}

func (s *Storage) DeleteUser(id int, user models.SafeUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_user",
		Login:   user.Login,
		Role:    user.Role,
		Class:   user.Class,
		Message: "deleting user",
	})

	deletedUser, err := s.getUserByIDLocked(id)
	if err != nil {
		return err
	}
	if deletedUser == nil {
		return errors.New("user not found")
	}

	query := `DELETE FROM users WHERE Id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "delete_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "failed to delete user: " + err.Error(),
		})
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("user not found")
	}

	if deletedUser.ClassID != 0 {
		if err := s.syncClassByIDLocked(deletedUser.ClassID); err != nil {
			return err
		}
	} else if err := s.syncClassLocked(deletedUser.Class); err != nil {
		return err
	}

	writeLog(logger.LogEntry{
		Level:   "info",
		Action:  "delete_user",
		Login:   user.Login,
		Role:    user.Role,
		Class:   user.Class,
		Message: "user deleted",
	})
	return nil
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
		SELECT Id, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		ORDER BY ClassID, LastName, Name, Login
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

	users, err := scanSafeUsers(rows)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_users",
			Message: "failed to scan users: " + err.Error(),
		})
		return nil, err
	}

	return users, nil
}

func (s *Storage) GetUserByLogin(login string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getUserByLoginLocked(login)
}

func (s *Storage) GetUserByID(id int) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getUserByIDLocked(id)
}

func validateNewUser(user *models.User) error {
	if user.Login == "" {
		return errors.New("login is required")
	}
	if strings.ContainsAny(user.Login, " \t\r\n") {
		return errors.New("login must not contain spaces")
	}
	if len(user.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if user.Role == "" {
		return errors.New("role is required")
	}
	if user.Class == "" && user.ClassID <= 0 {
		return errors.New("class is required")
	}
	if user.Rating < 0 {
		return errors.New("rating must be non-negative")
	}
	if user.Rating > 5000 {
		return errors.New("rating must be less than or equal to 5000")
	}
	if user.Rating == 0 {
		user.Rating = 500
	}
	if len(user.FullName) == 0 {
		return errors.New("full name is required")
	}
	if user.Name == "" {
		return errors.New("name is required")
	}
	if user.LastName == "" {
		return errors.New("last name is required")
	}

	return nil
}

func clampRating(rating int) int {
	if rating < 0 {
		return 0
	}
	if rating > 5000 {
		return 5000
	}
	return rating
}

func (s *Storage) resolveUserClassLocked(user *models.User) error {
	if user.ClassID > 0 {
		class, err := s.getClassByIDLocked(user.ClassID)
		if err != nil {
			return err
		}
		if class == nil {
			return errors.New("class not found")
		}
		user.Class = class.Name
		return nil
	}

	if user.Class == "" {
		return errors.New("class is required")
	}
	if err := s.ensureClassLocked(user.Class); err != nil {
		return err
	}

	classID, err := s.getClassIDByNameLocked(user.Class)
	if err != nil {
		return err
	}
	if classID == 0 {
		return errors.New("class not found")
	}

	user.ClassID = classID
	return nil
}

func (s *Storage) resolveSafeUserClassLocked(user *models.SafeUser) error {
	if user.ClassID > 0 {
		class, err := s.getClassByIDLocked(user.ClassID)
		if err != nil {
			return err
		}
		if class == nil {
			return errors.New("class not found")
		}
		user.Class = class.Name
		return nil
	}

	if user.Class == "" {
		return errors.New("class is required")
	}
	if err := s.ensureClassLocked(user.Class); err != nil {
		return err
	}

	classID, err := s.getClassIDByNameLocked(user.Class)
	if err != nil {
		return err
	}
	if classID == 0 {
		return errors.New("class not found")
	}

	user.ClassID = classID
	return nil
}

func (s *Storage) getSafeUserByLoginLocked(login string) (*models.SafeUser, error) {
	user, err := s.getUserByLoginLocked(login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &models.SafeUser{
		ID:       user.ID,
		Name:     user.Name,
		LastName: user.LastName,
		FullName: user.FullName,
		Login:    user.Login,
		Rating:   user.Rating,
		Role:     user.Role,
		Class:    user.Class,
		ClassID:  user.ClassID,
	}, nil
}

func (s *Storage) getUserByLoginLocked(login string) (*models.User, error) {
	row := s.db.QueryRow(`
		SELECT Id, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Login = ?
	`, normalizeLogin(login))

	return scanUser(row)
}

func (s *Storage) getUserByIDLocked(id int) (*models.User, error) {
	row := s.db.QueryRow(`
		SELECT Id, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Id = ?
	`, id)

	return scanUser(row)
}

type userScanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(scanner userScanner) (*models.User, error) {
	var user models.User
	var fullNameJSON sql.NullString

	err := scanner.Scan(
		&user.ID,
		&user.Name,
		&fullNameJSON,
		&user.LastName,
		&user.Login,
		&user.Password,
		&user.Rating,
		&user.Role,
		&user.Class,
		&user.ClassID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if fullNameJSON.Valid && fullNameJSON.String != "" {
		if err := json.Unmarshal([]byte(fullNameJSON.String), &user.FullName); err != nil {
			return nil, err
		}
	}
	if user.FullName == nil {
		user.FullName = []models.FullName{}
	}

	return &user, nil
}

func scanSafeUsers(rows *sql.Rows) ([]models.SafeUser, error) {
	users := make([]models.SafeUser, 0)

	for rows.Next() {
		var user models.SafeUser
		var fullNameJSON sql.NullString

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&fullNameJSON,
			&user.LastName,
			&user.Login,
			&user.Rating,
			&user.Role,
			&user.Class,
			&user.ClassID,
		); err != nil {
			return nil, err
		}

		if fullNameJSON.Valid && fullNameJSON.String != "" {
			if err := json.Unmarshal([]byte(fullNameJSON.String), &user.FullName); err != nil {
				return nil, err
			}
		}
		if user.FullName == nil {
			user.FullName = []models.FullName{}
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
