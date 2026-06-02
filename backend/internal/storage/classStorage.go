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
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func (s *Storage) initClassStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS classes (
		Id INTEGER PRIMARY KEY AUTOINCREMENT,
		Name TEXT NOT NULL UNIQUE,
		Grade INTEGER NOT NULL DEFAULT 0,
		Letter TEXT NOT NULL DEFAULT '',
		TeacherLogin TEXT,
		Members TEXT NOT NULL DEFAULT '[]',
		FirstQuarterComplete INTEGER NOT NULL DEFAULT 0,
		SecondQuarterComplete INTEGER NOT NULL DEFAULT 0,
		ThirdQuarterComplete INTEGER NOT NULL DEFAULT 0,
		QuarterComplete INTEGER NOT NULL DEFAULT 0,
		UserTotalRating INTEGER NOT NULL DEFAULT 0,
		ClassTotalRating INTEGER NOT NULL DEFAULT 0,
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

func (s *Storage) AddParallel(name string, classesIDs []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.addParallelInternal(name, classesIDs)
}

func (s *Storage) GetParallelClasses() ([]classModels.ParallelClass, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    rows, err := s.db.Query(`SELECT Id, Name, BestClassID, ClassTotalRating FROM parallels ORDER BY Name`)
    if err != nil {
        return nil, err
    }
    parallelClasses := make([]classModels.ParallelClass, 0)
    for rows.Next() {
        var pc classModels.ParallelClass
        if err := rows.Scan(&pc.ID, &pc.Name, &pc.BestClassID, &pc.ClassTotalRating); err != nil {
            rows.Close()
            return nil, err
        }
        parallelClasses = append(parallelClasses, pc)
    }
    if err := rows.Err(); err != nil { rows.Close(); return nil, err }
    rows.Close()

    for i := range parallelClasses {
        classRows, err := s.db.Query(
            `SELECT ClassID FROM parallel_classes WHERE ParallelID = ? ORDER BY ClassID`,
            parallelClasses[i].ID,
        )
        if err != nil { return nil, err }
        for classRows.Next() {
            var cid int
            if err := classRows.Scan(&cid); err != nil { classRows.Close(); return nil, err }
            parallelClasses[i].ClassesIDs = append(parallelClasses[i].ClassesIDs, cid)
        }
        classRows.Close()
    }
    return parallelClasses, nil
}

func (s *Storage) QuarterComplete(parallelClassID int) ([]*classModels.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if parallelClassID <= 0 {
		return nil, errors.New("parallel class id is required")
	}

	parallelClass, err := s.getParallelClassByIDLocked(parallelClassID)
	if err != nil {
		return nil, err
	}
	if parallelClass == nil {
		return nil, errors.New("parallel class not found")
	}

	classes := make([]*classModels.Class, 0, len(parallelClass.ClassesIDs))
	for _, classID := range parallelClass.ClassesIDs {
		class, err := s.getClassByIDLocked(classID)
		if err != nil {
			return nil, err
		}
		if class == nil {
			return nil, errors.New("class with id " + strconv.Itoa(classID) + " not found")
		}
		classes = append(classes, class)
	}

	sort.Slice(classes, func(i, j int) bool {
		if classes[i].ClassTotalRating == classes[j].ClassTotalRating {
			return classes[i].Name < classes[j].Name
		}
		return classes[i].ClassTotalRating > classes[j].ClassTotalRating
	})

	top3 := make([]*classModels.Class, 3)

	if len(classes) > 0 {
		classes[0].FirstQuarterComplete += 1
		top3[0] = classes[0]
	}
	if len(classes) > 1 {
		classes[1].SecondQuarterComplete += 1
		top3[1] = classes[1]
	}
	if len(classes) > 2 {
		classes[2].ThirdQuarterComplete += 1
		top3[2] = classes[2]
	}

	for _, class := range top3 {
		if class == nil {
			continue
		}
		_, err := s.db.Exec(`
			UPDATE classes
			SET FirstQuarterComplete = ?, SecondQuarterComplete = ?, ThirdQuarterComplete = ?
			WHERE Id = ?
		`, class.FirstQuarterComplete, class.SecondQuarterComplete, class.ThirdQuarterComplete, class.ID)
		if err != nil {
			return nil, err
		}
	}

	return top3, nil
}


func (s *Storage) YearComplete() ([]*classModels.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Открываем транзакцию СРАЗУ
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() 

	// 2. Делаем запрос чтением ВНУТРИ транзакции (tx.Query вместо s.db.Query)
	rows, err := tx.Query(`
		SELECT Id, Name, Grade, Letter, Members, TeacherLogin, 
		       FirstQuarterComplete, SecondQuarterComplete, ThirdQuarterComplete, 
		       ClassTotalRating, UserTotalRating, QuarterComplete
		FROM classes
		ORDER BY ClassTotalRating DESC
	`)
	if err != nil {
		return nil, err
	}
	// Мы закроем rows руками чуть ниже, но defer нужен на случай паники
	defer rows.Close() 

	classes := make([]*classModels.Class, 0)
	
	// Вспомогательная структура для хранения сырого JSON на этапе чтения
	type dbClass struct {
		class       *classModels.Class
		membersJSON string
	}
	var scannedClasses []dbClass

	// 3. Выкачиваем ВСЕ данные в оперативную память
	for rows.Next() {
		class := &classModels.Class{}
		var membersJSON string 

		if err := rows.Scan(
			&class.ID, &class.Name, &class.Grade, &class.Letter,
			&membersJSON, &class.TeacherLogin,
			&class.FirstQuarterComplete, &class.SecondQuarterComplete, &class.ThirdQuarterComplete,
			&class.ClassTotalRating, &class.UserTotalRating, &class.QuarterComplete,
		); err != nil {
			return nil, err
		}

		scannedClasses = append(scannedClasses, dbClass{class: class, membersJSON: membersJSON})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	// 4. КРИТИЧЕСКИ ВАЖНО: Закрываем чтение, чтобы освободить базу для записи!
	rows.Close()

	// 5. Теперь спокойно модифицируем данные и делаем UPDATE
	for _, item := range scannedClasses {
		class := item.class

		if err := json.Unmarshal([]byte(item.membersJSON), &class.Members); err != nil {
			return nil, err
		}

		class.QuarterComplete += 1
		class.Grade += 1
		class.ClassTotalRating = 0
		class.UserTotalRating = 0

		for i := range class.Members {
			class.Members[i].Rating = 0 
		}

		updatedMembersBytes, err := json.Marshal(class.Members)
		if err != nil {
			return nil, err
		}

		// Выполняем UPDATE в той же транзакции
		_, err = tx.Exec(`
			UPDATE classes 
			SET Grade = ?, QuarterComplete = ?, ClassTotalRating = ?, UserTotalRating = ?, Members = ?
			WHERE Id = ?
		`, class.Grade, class.QuarterComplete, class.ClassTotalRating, class.UserTotalRating, string(updatedMembersBytes), class.ID)
		
		if err != nil {
			return nil, err
		}

		classes = append(classes, class)
	}

	// 6. Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return classes, nil
}


func (s *Storage) getParallelClassByIDLocked(parallelClassID int) (*classModels.ParallelClass, error) {
	row := s.db.QueryRow(`
		SELECT Id, Name, BestClassID, ClassTotalRating
		FROM parallels
		WHERE Id = ?`, parallelClassID)
	var parallelClass classModels.ParallelClass
	if err := row.Scan(
		&parallelClass.ID,
		&parallelClass.Name,
		&parallelClass.BestClassID,
		&parallelClass.ClassTotalRating,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// load class ids
	rows, err := s.db.Query(`SELECT ClassID FROM parallel_classes WHERE ParallelID = ? ORDER BY ClassID`, parallelClass.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	classIDs := make([]int, 0)
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		classIDs = append(classIDs, cid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	parallelClass.ClassesIDs = classIDs

	return &parallelClass, nil
}

func (s *Storage) DeleteParallelClassByID(parallelClassID int, login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if parallelClassID <= 0 {
		return errors.New("parallel class id is required")
	}

	check, err := s.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil {
		return err
	}
	if !check {
		return errors.New("only owners can delete parallel classes")
	}

	_, err = s.db.Exec(`
		DELETE FROM parallels
		WHERE Id = ?
	`, parallelClassID)
	return err
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

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM users
		WHERE ClassID = ?
		AND LOWER(Role) IN ('user', 'helper');

		UPDATE users
		SET ClassID = 0,
			Class = ''
		WHERE ClassID = ?;
	`, classID)
	if err != nil {
		return err
	}

	res, err := tx.Exec(`
		DELETE FROM classes
		WHERE Id = ?
	`, classID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("class not found")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return s.syncAllClassesLocked()
}

func (s *Storage) GetAllClassTeachers() ([]userModels.SafeUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT u.Id, u.Avatar, u.Name, u.FullName, u.LastName, u.Login, u.Rating, u.Role, u.Class, u.ClassID
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

	check, err := s.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil || !check {
		return errors.New("no permission")
	}

	grade, letter, ok := ParseClass(input.Name)
	if !ok {
		return errors.New("invalid class name")
	}

	return s.saveClassInternal(input.Name, grade, letter, input.TeacherLogin)
}

func (s *Storage) AddParallelByGradeRange(name string, minGrade, maxGrade int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query("SELECT Id FROM classes WHERE Grade >= ? AND Grade <= ?", minGrade, maxGrade)
	if err != nil {
		return err
	}

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) == 0 {
		return errors.New("классы не найдены в этом диапазоне")
	}

	return s.addParallelInternal(name, ids)
}

func (s *Storage) addParallelInternal(name string, classesIDs []int) error {
	if len(classesIDs) == 0 {
		return errors.New("список классов пуст")
	}

	bestClassID := classesIDs[0]
	maxRating := -1

	for _, id := range classesIDs {
		var rating int
		err := s.db.QueryRow("SELECT ClassTotalRating FROM classes WHERE Id = ?", id).Scan(&rating)
		if err == nil && rating > maxRating {
			maxRating = rating
			bestClassID = id
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO parallels (Name, BestClassID, ClassTotalRating) VALUES (?, ?, ?)`,
		name, bestClassID, maxRating)
	if err != nil {
		return err
	}

	parallelID, _ := res.LastInsertId()

	for _, classID := range classesIDs {
		_, err := tx.Exec(`INSERT INTO parallel_classes (ParallelID, ClassID) VALUES (?, ?)`,
			parallelID, classID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) UpdateClass(classID int, input classModels.ClassInput, login string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	check, err := s.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil || !check {
		return errors.New("no permission")
	}

	grade, letter, ok := ParseClass(input.Name)
	if !ok {
		return errors.New("invalid class name")
	}

	return s.updateClassInternal(classID, input.Name, grade, letter, input.TeacherLogin)
}

func (s *Storage) updateClassInternal(classID int, name string, grade int, letter string, teacherLogin string) error {
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
	return s.db.QueryRow(`
		UPDATE classes
		SET Name = ?, Grade = ?, Letter = ?, TeacherLogin = ?
		WHERE Id = ?
	`, name, grade, letter, teacherLogin, classID).Err()
}

func (s *Storage) saveClassInternal(name string, grade int, letter string, teacherLogin string) error {
    res, err := s.db.Exec(`INSERT INTO classes (Name, Grade, Letter, TeacherLogin) VALUES (?, ?, ?, ?)`,
        name, grade, letter, teacherLogin)
    if err != nil { return err }
    classID, _ := res.LastInsertId()
    return s.autoAssignClassToParallelLocked(int(classID)) // ← internal, без Lock
}

func (s *Storage) GetClassIDsByRange(minGrade, maxGrade int) ([]int, error) {
	rows, err := s.db.Query("SELECT Id FROM classes WHERE Grade >= ? AND Grade <= ?", minGrade, maxGrade)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func ParseClass(s string) (int, string, bool) {
	var number int
	var letter string

	runes := []rune(s)

	for _, r := range runes {
		if unicode.IsDigit(r) {
			number = number*10 + int(r-'0')
		} else if unicode.IsLetter(r) {
			letter += string(r)
		}
	}

	if number == 0 || letter == "" {
		return 0, "", false
	}

	return number, letter, true
}

func (s *Storage) AddToParallelLocked(numberClass int) (int, error) {
	var parallelID int

	err := s.db.QueryRow(`SELECT Id FROM parallels WHERE MinGrade <= ? AND MaxGrade >= ? LIMIT 1`, numberClass, numberClass).Scan(&parallelID)
	if err == sql.ErrNoRows {
		name := strconv.Itoa(numberClass) + " параллель"
		res, err := s.db.Exec(`
			INSERT INTO parallels (Name, MinGrade, MaxGrade)
			VALUES (?, ?, ?)
		`, name, numberClass, numberClass)
		if err != nil {
			return 0, err
		}

		insertID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(insertID), nil
	}

	if err != nil {
		return 0, err
	}

	return parallelID, nil
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

func (s *Storage) GetClassesInParallel(id int) ([]classModels.Class, error) {
	s.mu.Lock()
    defer s.mu.Unlock()

    rows, err := s.db.Query(`
        SELECT c.Id, c.Name, c.Grade, c.Letter, c.TeacherLogin, c.Members,
               c.UserTotalRating, c.ClassTotalRating,
               c.FirstQuarterComplete, c.SecondQuarterComplete, c.ThirdQuarterComplete
        FROM classes c
        JOIN parallel_classes pc ON pc.ClassID = c.Id
        WHERE pc.ParallelID = ?
        ORDER BY c.Name
    `, id)
    if err != nil {
        logger.WriteSafe(logger.LogEntry{
            Level:   "error",
            Action:  "get_classes_in_parallel",
            Message: "failed to query classes in parallel: " + err.Error(),
        })
        return nil, err
    }
    defer rows.Close()

    classes := make([]classModels.Class, 0)
    for rows.Next() {
        class, err := s.scanClassRowsLocked(rows)
        if err != nil {
            // ...
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

func (s *Storage) GetAllClasses() ([]classModels.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT Id, Name, Grade, Letter, TeacherLogin, Members, UserTotalRating, ClassTotalRating,
		FirstQuarterComplete, SecondQuarterComplete, ThirdQuarterComplete
		FROM classes
		ORDER BY Name
	`)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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
			logger.WriteSafe(logger.LogEntry{
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

	userTotalRating := 0

	if len(members) > 0 {
		for _, member := range members {
			userTotalRating += member.Rating
		}

		userTotalRating = userTotalRating / len(members)
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

		if teacher == nil {
			teacherLogin = sql.NullString{}
		} else if !utils.IsSystemRole(teacher.Role) && teacher.ClassID != classID {
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
		SET Members = ?, UserTotalRating = ?, TeacherLogin = ?
		WHERE Id = ?
	`, string(membersJSON), userTotalRating, teacherLogin, classID)
	return err
}

func (s *Storage) AddClassRating(classID int, points int) error {
	_, err := s.db.Exec(`
		UPDATE classes
		SET ClassTotalRating = ClassTotalRating + ?
		WHERE Id = ?
	`, points, classID)

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
		SELECT Id, Name, Grade, Letter, TeacherLogin, Members, UserTotalRating, ClassTotalRating
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
		SELECT Id, Name, Grade, Letter, TeacherLogin, Members, UserTotalRating, ClassTotalRating,
		FirstQuarterComplete, SecondQuarterComplete, ThirdQuarterComplete
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
		SELECT Id, Avatar, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
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
		&class.Grade,
		&class.Letter,
		&teacherLogin,
		&membersJSON,
		&class.UserTotalRating,
		&class.ClassTotalRating,
		&class.FirstQuarterComplete,
		&class.SecondQuarterComplete,
		&class.ThirdQuarterComplete,
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
		SELECT Id, Avatar, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
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

// InitializeParallelsFromConfig инициализирует параллели из конфигурации при старте сервера
// Создает параллели, если они не существуют, и синхронизирует классы
func (s *Storage) InitializeParallelsFromConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.ParallelsConfig) == 0 {
		return nil
	}

	for _, parallelCfg := range s.ParallelsConfig {
		if err := s.initializeParallelLocked(parallelCfg.Name, parallelCfg.MinGrade, parallelCfg.MaxGrade); err != nil {
			return err
		}
	}

	return nil
}

// initializeParallelLocked создает одну параллель и добавляет в нее все соответствующие классы
func (s *Storage) initializeParallelLocked(name string, minGrade, maxGrade int) error {
	// Проверяем, существует ли уже такая параллель
	var parallelID int
	err := s.db.QueryRow(`
		SELECT Id FROM parallels 
		WHERE MinGrade = ? AND MaxGrade = ?
	`, minGrade, maxGrade).Scan(&parallelID)

	if err == nil {
		// Параллель уже существует, просто синхронизируем классы
		return s.syncParallelClassesLocked(parallelID, minGrade, maxGrade)
	}

	if err != sql.ErrNoRows {
		return err
	}

	// Параллель не существует, создаем ее
	// Сначала получаем все классы в этом диапазоне
	rows, err := s.db.Query(`
		SELECT Id, ClassTotalRating FROM classes 
		WHERE Grade >= ? AND Grade <= ?
		ORDER BY ClassTotalRating DESC
		LIMIT 1
	`, minGrade, maxGrade)
	if err != nil {
		return err
	}
	defer rows.Close()

	bestClassID := 0
	bestRating := 0
	if rows.Next() {
		var id, rating int
		if err := rows.Scan(&id, &rating); err != nil {
			return err
		}
		bestClassID = id
		bestRating = rating
	}

	// Если нет классов в диапазоне, создаем пустую параллель
	if bestClassID == 0 {
		bestClassID = 0
		bestRating = 0
	}

	// Вставляем новую параллель
	res, err := s.db.Exec(`
		INSERT INTO parallels (Name, MinGrade, MaxGrade, BestClassID, ClassTotalRating)
		VALUES (?, ?, ?, ?, ?)
	`, name, minGrade, maxGrade, bestClassID, bestRating)

	if err != nil {
		// Может быть UNIQUE constraint на Name, но это нормально
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil
		}
		return err
	}

	newParallelID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Добавляем все классы в этот диапазон в параллель
	return s.syncParallelClassesLocked(int(newParallelID), minGrade, maxGrade)
}

// syncParallelClassesLocked синхронизирует классы для параллели
// Удаляет старые связи и создает новые для всех классов в диапазоне
func (s *Storage) syncParallelClassesLocked(parallelID, minGrade, maxGrade int) error {
	// Получаем все классы в диапазоне
	rows, err := s.db.Query(`
		SELECT Id FROM classes 
		WHERE Grade >= ? AND Grade <= ?
	`, minGrade, maxGrade)
	if err != nil {
		return err
	}
	defer rows.Close()

	var classIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		classIDs = append(classIDs, id)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Удаляем старые связи для этой параллели
	_, err = s.db.Exec(`DELETE FROM parallel_classes WHERE ParallelID = ?`, parallelID)
	if err != nil {
		return err
	}

	// Добавляем новые связи
	for _, classID := range classIDs {
		_, err := s.db.Exec(`
			INSERT OR IGNORE INTO parallel_classes (ParallelID, ClassID) 
			VALUES (?, ?)
		`, parallelID, classID)
		if err != nil {
			return err
		}
	}

	// Обновляем BestClassID и ClassTotalRating для параллели
	var bestClassID int
	var bestRating int
	err = s.db.QueryRow(`
		SELECT c.Id, c.ClassTotalRating FROM classes c
		WHERE c.Grade >= ? AND c.Grade <= ?
		ORDER BY c.ClassTotalRating DESC
		LIMIT 1
	`, minGrade, maxGrade).Scan(&bestClassID, &bestRating)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if bestClassID == 0 {
		bestRating = 0
	}

	_, err = s.db.Exec(`
		UPDATE parallels 
		SET BestClassID = ?, ClassTotalRating = ?
		WHERE Id = ?
	`, bestClassID, bestRating, parallelID)

	return err
}

// AutoAssignClassToParallel автоматически добавляет класс в параллель на основе его grade
// Вызывается при добавлении нового класса
func (s *Storage) AutoAssignClassToParallel(classID int) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.autoAssignClassToParallelLocked(classID)
}

func (s *Storage) autoAssignClassToParallelLocked(classID int) error {
    var grade int
    err := s.db.QueryRow(`SELECT Grade FROM classes WHERE Id = ?`, classID).Scan(&grade)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil
        }
        return err
    }

    // Находим параллель, которой этот класс должен принадлежать
    var parallelID int
    err = s.db.QueryRow(`
        SELECT Id FROM parallels 
        WHERE MinGrade <= ? AND MaxGrade >= ?
        LIMIT 1
    `, grade, grade).Scan(&parallelID)

    if err != nil {
        if err == sql.ErrNoRows {
            // Нет подходящей параллели - это нормально
            return nil
        }
        return err
    }

    // Добавляем связь класса с параллелью
    _, err = s.db.Exec(`
        INSERT OR IGNORE INTO parallel_classes (ParallelID, ClassID)
        VALUES (?, ?)
    `, parallelID, classID)

    return err
}
