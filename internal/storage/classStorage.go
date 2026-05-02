package storage

import (
	"cspirt/internal/logger"
	userModels "cspirt/internal/users/models"
	classModels "cspirt/internal/class/models"
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
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_class_id ON users(ClassID);`); err != nil {
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

func (s *Storage) saveClassTeacherLocked(name string, teacherLogin string) error {
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
	class, err := s.getClassByNameLocked(name)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}
	if teacher.ClassID != class.ID {
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

func (s *Storage) GetAllClasses() ([]classModels.Class, error) {
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

	classes := make([]classModels.Class, 0)
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

func (s *Storage) SaveClassTeacherByID(classID int, teacherLogin string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID <= 0 {
		return errors.New("class id is required")
	}

	class, err := s.getClassByIDLocked(classID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}

	return s.saveClassTeacherLocked(class.Name, teacherLogin)
}

func (s *Storage) GetClassByID(id int) (*classModels.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return nil, errors.New("class id is required")
	}

	class, err := s.getClassByIDLocked(id)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, nil
	}

	if err := s.syncClassByIDLocked(id); err != nil {
		return nil, err
	}

	return s.getClassByIDLocked(id)
}

func (s *Storage) GetClassTeacherByID(classID int) (*userModels.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID <= 0 {
		return nil, errors.New("class id is required")
	}

	class, err := s.getClassByIDLocked(classID)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, nil
	}

	if err := s.syncClassByIDLocked(classID); err != nil {
		return nil, err
	}
	class, err = s.getClassByIDLocked(classID)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, nil
	}

	if class.TeacherLogin == "" {
		return nil, nil
	}

	return s.getSafeUserByLoginLocked(class.TeacherLogin)
}

func (s *Storage) GetUsersByClassID(classID int) ([]userModels.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID <= 0 {
		return nil, errors.New("class id is required")
	}

	class, err := s.getClassByIDLocked(classID)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, nil
	}

	return s.getUsersByClassIDLocked(classID)
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

	if _, err := s.db.Exec(`
		UPDATE users
		SET ClassID = (
			SELECT Id
			FROM classes
			WHERE classes.Name = users.Class
		)
		WHERE (ClassID IS NULL OR ClassID = 0)
		  AND TRIM(Class) <> ''
		  AND EXISTS (
			SELECT 1
			FROM classes
			WHERE classes.Name = users.Class
		  )
	`); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
		UPDATE users
		SET Class = (
			SELECT Name
			FROM classes
			WHERE classes.Id = users.ClassID
		)
		WHERE ClassID IS NOT NULL
		  AND ClassID > 0
		  AND EXISTS (
			SELECT 1
			FROM classes
			WHERE classes.Id = users.ClassID
		  )
	`); err != nil {
		return err
	}

	rows, err = s.db.Query(`SELECT Id FROM classes`)
	if err != nil {
		return err
	}

	allClassIDs := make([]int, 0)
	for rows.Next() {
		var classID int
		if err := rows.Scan(&classID); err != nil {
			rows.Close()
			return err
		}
		allClassIDs = append(allClassIDs, classID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, classID := range allClassIDs {
		if err := s.syncClassByIDLocked(classID); err != nil {
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

	if err := s.ensureClassLocked(name); err != nil {
		return err
	}

	classID, err := s.getClassIDByNameLocked(name)
	if err != nil {
		return err
	}
	if classID == 0 {
		return nil
	}

	return s.syncClassByIDLocked(classID)
}

func (s *Storage) syncClassByIDLocked(classID int) error {
	if classID <= 0 {
		return nil
	}

	class, err := s.getClassByIDLocked(classID)
	if err != nil {
		return err
	}
	if class == nil {
		return nil
	}

	members, err := s.getUsersByClassIDLocked(classID)
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
	err = s.db.QueryRow(`SELECT TeacherLogin FROM classes WHERE Id = ?`, classID).Scan(&teacherLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else {
			return err
		}
	}

	if teacherLogin.Valid && teacherLogin.String != "" {
		teacher, err := s.getUserByLoginLocked(teacherLogin.String)
		if err != nil {
			return err
		}
		if teacher == nil || teacher.ClassID != classID {
			teacherLogin = sql.NullString{}
		}
	}

	if !teacherLogin.Valid || teacherLogin.String == "" {
		candidate, err := s.findTeacherCandidateLocked(classID)
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
		WHERE Id = ?
	`, string(membersJSON), totalRating, teacherLogin, classID)
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

func (s *Storage) getClassByNameLocked(name string) (*classModels.Class, error) {
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

func (s *Storage) getClassByIDLocked(id int) (*classModels.Class, error) {
	row := s.db.QueryRow(`
		SELECT Id, Name, TeacherLogin, Members, TotalRating
		FROM classes
		WHERE Id = ?
	`, id)

	class, err := s.scanClassScannerLocked(row, true)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &class, nil
}

func (s *Storage) getClassIDByNameLocked(name string) (int, error) {
	var classID int
	err := s.db.QueryRow(`SELECT Id FROM classes WHERE Name = ?`, normalizeClassName(name)).Scan(&classID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return classID, nil
}

func (s *Storage) getUsersByClassIDLocked(classID int) ([]userModels.SafeUser, error) {
	rows, err := s.db.Query(`
		SELECT Id, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		WHERE ClassID = ?
		ORDER BY LastName, Name, Login
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSafeUsers(rows)
}

type classScanner interface {
	Scan(dest ...interface{}) error
}

func (s *Storage) scanClassScannerLocked(scanner classScanner, loadTeacher bool) (classModels.Class, error) {
	var class classModels.Class
	var teacherLogin sql.NullString
	var membersJSON string

	if err := scanner.Scan(
		&class.ID,
		&class.Name,
		&teacherLogin,
		&membersJSON,
		&class.TotalRating,
	); err != nil {
		return classModels.Class{}, err
	}

	if membersJSON != "" {
		if err := json.Unmarshal([]byte(membersJSON), &class.Members); err != nil {
			return classModels.Class{}, err
		}
	}
	if class.Members == nil {
		class.Members = []userModels.SafeUser{}
	}

	if teacherLogin.Valid && teacherLogin.String != "" {
		class.TeacherLogin = teacherLogin.String
		if loadTeacher {
			teacher, err := s.getSafeUserByLoginLocked(teacherLogin.String)
			if err != nil {
				return classModels.Class{}, err
			}
			class.Teacher = teacher
		}
	}

	return class, nil
}

func (s *Storage) scanClassRowsLocked(rows *sql.Rows) (classModels.Class, error) {
	return s.scanClassScannerLocked(rows, false)
}

func (s *Storage) findTeacherCandidateLocked(classID int) (string, error) {
	var login string
	err := s.db.QueryRow(`
		SELECT Login
		FROM users
		WHERE ClassID = ?
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
	`, classID).Scan(&login)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return login, nil
}
