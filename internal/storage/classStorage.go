package storage

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
)

func (s *Storage) initClassStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS classes (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Name TEXT NOT NULL UNIQUE,
		TeacherLogin TEXT,
		Members TEXT NOT NULL DEFAULT '[]',
		TotalRating INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (TeacherLogin) REFERENCES users(Login) ON DELETE SET NULL
	);`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class ON users(Class);`); err != nil {
		return err
	}

	return s.syncAllClassesLocked()
}

func (s *Storage) EnsureClass(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = normalizeClassName(name)
	if name == "" {
		return errors.New("class is required")
	}

	if err := s.ensureClassLocked(name); err != nil {
		return err
	}

	return s.syncClassLocked(name)
}

func (s *Storage) SaveClassTeacher(name string, teacherLogin string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = normalizeClassName(name)
	teacherLogin = normalizeLogin(teacherLogin)

	if name == "" {
		return errors.New("class is required")
	}
	if teacherLogin == "" {
		return errors.New("teacher login is required")
	}

	teacher, err := s.getUserByLoginLocked(teacherLogin)
	if err != nil {
		return err
	}
	if teacher == nil {
		return errors.New("teacher not found")
	}
	if normalizeClassName(teacher.Class) != name {
		return errors.New("teacher must belong to this class")
	}
	if !isTeacherCandidate(teacher.Role) {
		return errors.New("class teacher must have helper, admin or owner role")
	}

	if err := s.ensureClassLocked(name); err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE classes
		SET TeacherLogin = ?
		WHERE Name = ?
	`, teacher.Login, name)
	if err != nil {
		return err
	}

	return s.syncClassLocked(name)
}

func (s *Storage) GetAllClasses() ([]models.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.syncAllClassesLocked(); err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_classes",
			Message: "failed to sync classes: " + err.Error(),
		})
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT Id, Name, TeacherLogin, Members, TotalRating
		FROM classes
		ORDER BY Name
	`)
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_classes",
			Message: "failed to query classes: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	classes := make([]models.Class, 0)
	for rows.Next() {
		class, err := s.scanClassRowsLocked(rows)
		if err != nil {
			writeLog(logger.LogEntry{
				Level:   "error",
				Action:  "get_all_classes",
				Message: "failed to scan class: " + err.Error(),
			})
			return nil, err
		}
		classes = append(classes, class)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for i := range classes {
		if classes[i].TeacherLogin == "" {
			continue
		}

		teacher, err := s.getSafeUserByLoginLocked(classes[i].TeacherLogin)
		if err != nil {
			return nil, err
		}
		classes[i].Teacher = teacher
	}

	return classes, nil
}

func (s *Storage) GetClassByName(name string) (*models.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = normalizeClassName(name)
	if name == "" {
		return nil, errors.New("class is required")
	}

	exists, err := s.classExistsLocked(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	if err := s.syncClassLocked(name); err != nil {
		return nil, err
	}

	class, err := s.getClassByNameLocked(name)
	if err != nil {
		return nil, err
	}

	return class, nil
}

func (s *Storage) GetClassTeacher(name string) (*models.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = normalizeClassName(name)
	if name == "" {
		return nil, errors.New("class is required")
	}

	exists, err := s.classExistsLocked(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	if err := s.syncClassLocked(name); err != nil {
		return nil, err
	}

	var teacherLogin sql.NullString
	err = s.db.QueryRow(`SELECT TeacherLogin FROM classes WHERE Name = ?`, name).Scan(&teacherLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !teacherLogin.Valid || teacherLogin.String == "" {
		return nil, nil
	}

	return s.getSafeUserByLoginLocked(teacherLogin.String)
}

func (s *Storage) syncAllClasses() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.syncAllClassesLocked()
}

func (s *Storage) syncAllClassesLocked() error {
	rows, err := s.db.Query(`
		SELECT DISTINCT Class
		FROM users
		WHERE TRIM(Class) <> ''
	`)
	if err != nil {
		return err
	}

	classNames := make([]string, 0)
	for rows.Next() {
		var className string
		if err := rows.Scan(&className); err != nil {
			rows.Close()
			return err
		}
		className = normalizeClassName(className)
		if className != "" {
			classNames = append(classNames, className)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, className := range classNames {
		if err := s.ensureClassLocked(className); err != nil {
			return err
		}
	}

	rows, err = s.db.Query(`SELECT Name FROM classes`)
	if err != nil {
		return err
	}

	allClassNames := make([]string, 0)
	for rows.Next() {
		var className string
		if err := rows.Scan(&className); err != nil {
			rows.Close()
			return err
		}
		allClassNames = append(allClassNames, className)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, className := range allClassNames {
		if err := s.syncClassLocked(className); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) syncClassLocked(name string) error {
	name = normalizeClassName(name)
	if name == "" {
		return nil
	}

	members, err := s.getUsersByClassLocked(name)
	if err != nil {
		return err
	}

	totalRating := 0
	for _, member := range members {
		totalRating += member.Rating
	}

	membersJSON, err := json.Marshal(members)
	if err != nil {
		return err
	}

	var teacherLogin sql.NullString
	err = s.db.QueryRow(`SELECT TeacherLogin FROM classes WHERE Name = ?`, name).Scan(&teacherLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			if err := s.ensureClassLocked(name); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if teacherLogin.Valid && teacherLogin.String != "" {
		teacher, err := s.getUserByLoginLocked(teacherLogin.String)
		if err != nil {
			return err
		}
		if teacher == nil || normalizeClassName(teacher.Class) != name {
			teacherLogin = sql.NullString{}
		}
	}

	if !teacherLogin.Valid || teacherLogin.String == "" {
		candidate, err := s.findTeacherCandidateLocked(name)
		if err != nil {
			return err
		}
		if candidate != "" {
			teacherLogin = sql.NullString{String: candidate, Valid: true}
		}
	}

	_, err = s.db.Exec(`
		UPDATE classes
		SET Members = ?, TotalRating = ?, TeacherLogin = ?
		WHERE Name = ?
	`, string(membersJSON), totalRating, teacherLogin, name)
	return err
}

func (s *Storage) ensureClassLocked(name string) error {
	name = normalizeClassName(name)
	if name == "" {
		return nil
	}

	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO classes (Name)
		VALUES (?)
	`, name)
	return err
}

func (s *Storage) classExistsLocked(name string) (bool, error) {
	var exists int
	err := s.db.QueryRow(`SELECT 1 FROM classes WHERE Name = ?`, name).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *Storage) getClassByNameLocked(name string) (*models.Class, error) {
	row := s.db.QueryRow(`
		SELECT Id, Name, TeacherLogin, Members, TotalRating
		FROM classes
		WHERE Name = ?
	`, name)

	class, err := s.scanClassScannerLocked(row, true)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &class, nil
}

type classScanner interface {
	Scan(dest ...interface{}) error
}

func (s *Storage) scanClassScannerLocked(scanner classScanner, loadTeacher bool) (models.Class, error) {
	var class models.Class
	var teacherLogin sql.NullString
	var membersJSON string

	if err := scanner.Scan(
		&class.ID,
		&class.Name,
		&teacherLogin,
		&membersJSON,
		&class.TotalRating,
	); err != nil {
		return models.Class{}, err
	}

	if membersJSON != "" {
		if err := json.Unmarshal([]byte(membersJSON), &class.Members); err != nil {
			return models.Class{}, err
		}
	}
	if class.Members == nil {
		class.Members = []models.SafeUser{}
	}

	if teacherLogin.Valid && teacherLogin.String != "" {
		class.TeacherLogin = teacherLogin.String
		if loadTeacher {
			teacher, err := s.getSafeUserByLoginLocked(teacherLogin.String)
			if err != nil {
				return models.Class{}, err
			}
			class.Teacher = teacher
		}
	}

	return class, nil
}

func (s *Storage) scanClassRowsLocked(rows *sql.Rows) (models.Class, error) {
	return s.scanClassScannerLocked(rows, false)
}

func (s *Storage) findTeacherCandidateLocked(name string) (string, error) {
	var login string
	err := s.db.QueryRow(`
		SELECT Login
		FROM users
		WHERE Class = ?
		AND LOWER(Role) IN ('admin', 'owner', 'helper')
		ORDER BY
			CASE LOWER(Role)
				WHEN 'admin' THEN 0
				WHEN 'owner' THEN 1
				WHEN 'helper' THEN 2
				ELSE 3
			END,
			Id
		LIMIT 1
	`, name).Scan(&login)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return login, nil
}
