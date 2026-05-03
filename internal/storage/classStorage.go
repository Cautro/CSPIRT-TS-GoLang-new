package storage

import (
	classModels "cspirt/internal/class/models"
	"cspirt/internal/logger"
	ratingModels "cspirt/internal/rating/models"
	userModels "cspirt/internal/users/models"
	"cspirt/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
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
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_classes_teacher_login ON classes(TeacherLogin);`); err != nil {
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

func (s *Storage) DeleteClassByID(classID int, login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID <= 0 {
		return errors.New("class id is required")
	}

	check, err := s.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil {
		return err
	}
	if !check {
		return errors.New("only owners can delete classes")
	}

	_, err = s.db.Exec(`
		DELETE FROM classes
		WHERE Id = ?
	`, classID)
	if err != nil {
		return err
	}

	return s.syncAllClassesLocked()
}

func (s *Storage) GetAllClassTeachers() ([]userModels.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT u.Id, u.Name, u.FullName, u.LastName, u.Login, u.Rating, u.Role, u.Class, u.ClassID
		FROM users u
		JOIN classes c ON c.TeacherLogin = u.Login
		ORDER BY u.LastName, u.Name, u.Login
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSafeUsers(rows)
}

func (s *Storage) AddClass(input classModels.ClassInput, login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := normalizeClassName(input.Name)
	if name == "" {
		return errors.New("class name is required")
	}

	check, err := s.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil {
		return err
	}
	if !check {
		return errors.New("only owners can add classes")
	}

	teacherLogin := normalizeLogin(input.TeacherLogin)

	if teacherLogin != "" {
		teacher, err := s.getUserByLoginLocked(teacherLogin)
		if err != nil {
			return err
		}
		if teacher == nil {
			return errors.New("teacher not found")
		}
		if !isTeacherCandidate(teacher.Role) {
			return errors.New("class teacher must have helper, admin or owner role")
		}
	}

	_, err = s.db.Exec(`
		INSERT INTO classes (Name, TeacherLogin)
		VALUES (?, ?)
	`, name, nullableString(teacherLogin))
	if err != nil {
		writeLog(logger.LogEntry{
			Level:   "error",
			Action:  "add_class",
			Message: "failed to insert class: " + err.Error(),
		})
		return err
	}

	return s.syncClassLocked(name)
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func (s *Storage) hasUserRoleLocked(login string, roles ...string) (bool, error) {
	user, err := s.getUserByLoginLocked(login)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("user not found")
	}

	userRole := strings.ToLower(strings.TrimSpace(user.Role))
	for _, role := range roles {
		if userRole == strings.ToLower(strings.TrimSpace(role)) {
			return true, nil
		}
	}

	return false, nil
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
	if !utils.IsSystemRole(teacher.Role) && teacher.ClassID != class.ID {
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

	if err := s.loadClassTeachersLocked(classes); err != nil {
		return nil, err
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

	if class.TeacherLogin == "" {
		return nil, nil
	}
	if class.Teacher != nil {
		return class.Teacher, nil
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
	if _, err := s.db.Exec(`
		INSERT OR IGNORE INTO classes (Name)
		SELECT DISTINCT UPPER(TRIM(Class))
		FROM users
		WHERE TRIM(Class) <> ''
	`); err != nil {
		return err
	}

	if _, err := s.db.Exec(`
		UPDATE users
		SET ClassID = (
			SELECT Id
			FROM classes
			WHERE classes.Name = UPPER(TRIM(users.Class))
		)
		WHERE (ClassID IS NULL OR ClassID = 0)
		AND TRIM(Class) <> ''
		AND EXISTS (
			SELECT 1
			FROM classes
			WHERE classes.Name = UPPER(TRIM(users.Class))
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

	rows, err := s.db.Query(`SELECT Id FROM classes`)
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

	if len(members) > 0 {
		for _, member := range members {
			totalRating += member.Rating
		}

		totalRating = totalRating / len(members)
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

func (s *Storage) loadClassTeachersLocked(classes []classModels.Class) error {
	logins := make([]string, 0, len(classes))
	seen := make(map[string]struct{}, len(classes))

	for i := range classes {
		login := classes[i].TeacherLogin
		if login == "" {
			continue
		}
		if _, ok := seen[login]; ok {
			continue
		}

		seen[login] = struct{}{}
		logins = append(logins, login)
	}
	if len(logins) == 0 {
		return nil
	}

	placeholders := make([]string, len(logins))
	args := make([]interface{}, len(logins))
	for i, login := range logins {
		placeholders[i] = "?"
		args[i] = login
	}

	rows, err := s.db.Query(`
		SELECT Id, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		WHERE Login IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	teachers, err := scanSafeUsers(rows)
	if err != nil {
		return err
	}

	teachersByLogin := make(map[string]*userModels.SafeUser, len(teachers))
	for i := range teachers {
		teachersByLogin[teachers[i].Login] = &teachers[i]
	}

	for i := range classes {
		if teacher, ok := teachersByLogin[classes[i].TeacherLogin]; ok {
			classes[i].Teacher = teacher
		}
	}

	return nil
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
