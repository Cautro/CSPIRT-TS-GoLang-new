package repo

import (
	classConfig "cspirt/internal/controller/http/class/config"
	classModels "cspirt/internal/domain/class"
	classRepo "cspirt/internal/domain/class/repo"
	"cspirt/pkg/logger"
	ratingModels "cspirt/internal/domain/rating"
	userModels "cspirt/internal/domain/user"
	"cspirt/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) classRepo.ClassRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) EnsureClass(name string) error {
	name = normalizeClassName(name)
	if name == "" {
		return errors.New("class is required")
	}

	if err := r.ensureClassLocked(name); err != nil {
		return err
	}

	return r.syncClassLocked(name)
}

func (r *postgresRepository) AddParallel(name string, classesIDs []int) error {
	return r.addParallelInternal(name, classesIDs)
}

func (r *postgresRepository) GetParallelClasses() ([]classModels.ParallelClass, error) {
	rows, err := r.db.Query(`SELECT Id, Name, BestClassID, ClassTotalRating FROM parallels ORDER BY Name`)
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
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for i := range parallelClasses {
		classRows, err := r.db.Query(
			`SELECT ClassID FROM parallel_classes WHERE ParallelID = $1 ORDER BY ClassID`,
			parallelClasses[i].ID,
		)
		if err != nil {
			return nil, err
		}
		for classRows.Next() {
			var cid int
			if err := classRows.Scan(&cid); err != nil {
				classRows.Close()
				return nil, err
			}
			parallelClasses[i].ClassesIDs = append(parallelClasses[i].ClassesIDs, cid)
		}
		classRows.Close()
	}
	return parallelClasses, nil
}

func (r *postgresRepository) QuarterComplete(parallelClassID int) ([]*classModels.Class, error) {
	if parallelClassID <= 0 {
		return nil, errors.New("parallel class id is required")
	}

	parallelClass, err := r.getParallelClassByIDLocked(parallelClassID)
	if err != nil {
		return nil, err
	}
	if parallelClass == nil {
		return nil, errors.New("parallel class not found")
	}

	classes := make([]*classModels.Class, 0, len(parallelClass.ClassesIDs))
	for _, classID := range parallelClass.ClassesIDs {
		class, err := r.getClassByIDLocked(classID)
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
		_, err := r.db.Exec(`
			UPDATE classes
			SET FirstQuarterComplete = $1, SecondQuarterComplete = $2, ThirdQuarterComplete = $3
			WHERE id = $4
		`, class.FirstQuarterComplete, class.SecondQuarterComplete, class.ThirdQuarterComplete, class.ID)
		if err != nil {
			return nil, err
		}
	}

	return top3, nil
}

func (r *postgresRepository) YearComplete() ([]*classModels.Class, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

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
	defer rows.Close()

	type DBClass struct {
		class       *classModels.Class
		membersJSON string
	}
	var scannedClasses []DBClass

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
		scannedClasses = append(scannedClasses, DBClass{class: class, membersJSON: membersJSON})
	}
	rows.Close()

	classes := make([]*classModels.Class, 0)

	for _, item := range scannedClasses {
		class := item.class

		if err := json.Unmarshal([]byte(item.membersJSON), &class.Members); err != nil {
			return nil, err
		}

		class.QuarterComplete += 1
		class.Grade += 1
		if class.Grade > 11 {
			class.Grade = 11
		}
		class.ClassTotalRating = 0
		class.UserTotalRating = 0

		for i := range class.Members {
			class.Members[i].Rating = 0
		}

		updatedMembersBytes, err := json.Marshal(class.Members)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(`
			UPDATE classes
			SET Grade = $1, QuarterComplete = $2, ClassTotalRating = $3, UserTotalRating = $4, Members = $5
			WHERE id = $6
		`, class.Grade, class.QuarterComplete, class.ClassTotalRating, class.UserTotalRating, string(updatedMembersBytes), class.ID)
		if err != nil {
			return nil, err
		}
		var newParallelID int
		err = tx.QueryRow(`
			SELECT Id FROM parallels
			WHERE $1 >= MinGrade AND $2 <= MaxGrade
			LIMIT 1
		`, class.Grade, class.Grade).Scan(&newParallelID)

		if err == sql.ErrNoRows {
			_, err = tx.Exec(`DELETE FROM parallel_classes WHERE ClassID = $1`, class.ID)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		} else {
			_, err = tx.Exec(`DELETE FROM parallel_classes WHERE ClassID = $1`, class.ID)
			if err != nil {
				return nil, err
			}

			_, err = tx.Exec(`
				INSERT INTO parallel_classes (ParallelID, ClassID)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, newParallelID, class.ID)
			if err != nil {
				return nil, err
			}
		}

		classes = append(classes, class)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return classes, nil
}

func (r *postgresRepository) getParallelClassByIDLocked(parallelClassID int) (*classModels.ParallelClass, error) {
	row := r.db.QueryRow(`
		SELECT Id, Name, BestClassID, ClassTotalRating
		FROM parallels
		WHERE id = $1`, parallelClassID)
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

	rows, err := r.db.Query(`SELECT ClassID FROM parallel_classes WHERE ParallelID = $1 ORDER BY ClassID`, parallelClass.ID)
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

func (r *postgresRepository) DeleteParallelClassByID(parallelClassID int, login string) error {
	if parallelClassID <= 0 {
		return errors.New("parallel class id is required")
	}

	check, err := r.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil {
		return err
	}
	if !check {
		return errors.New("only owners can delete parallel classes")
	}

	_, err = r.db.Exec(`
		DELETE FROM parallels
		WHERE id = $1
	`, parallelClassID)
	return err
}

func (r *postgresRepository) DeleteClassByID(classID int, login string) error {
	if classID <= 0 {
		return errors.New("class id is required")
	}

	check, err := r.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil {
		return err
	}
	if !check {
		return errors.New("only owners can delete classes")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM users
		WHERE ClassID = $1
		AND LOWER(Role) IN ('user', 'helper')
	`, classID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		UPDATE users
		SET ClassID = 0,
			Class = ''
		WHERE ClassID = $1
	`, classID)
	if err != nil {
		return err
	}

	res, err := tx.Exec(`
		DELETE FROM classes
		WHERE id = $1
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

	return r.syncAllClassesLocked()
}

func (r *postgresRepository) GetAllClassTeachers() ([]userModels.SafeUser, error) {
	rows, err := r.db.Query(`
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

func (r *postgresRepository) AddClass(input classModels.ClassInput, login string) error {
	check, err := r.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil || !check {
		return errors.New("no permission")
	}

	grade, letter, ok := ParseClass(input.Name)
	if !ok {
		return errors.New("invalid class name")
	}

	return r.saveClassInternal(input.Name, grade, letter, input.TeacherLogin)
}

func (r *postgresRepository) AddParallelByGradeRange(name string, minGrade, maxGrade int) error {
	rows, err := r.db.Query("SELECT Id FROM classes WHERE Grade >= $1 AND Grade <= $2", minGrade, maxGrade)
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

	return r.addParallelInternal(name, ids)
}

func (r *postgresRepository) addParallelInternal(name string, classesIDs []int) error {
	if len(classesIDs) == 0 {
		return errors.New("список классов пуст")
	}

	bestClassID := classesIDs[0]
	maxRating := -1

	for _, id := range classesIDs {
		var rating int
		err := r.db.QueryRow("SELECT ClassTotalRating FROM classes WHERE id = $1", id).Scan(&rating)
		if err == nil && rating > maxRating {
			maxRating = rating
			bestClassID = id
		}
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var parallelID int64
	err = tx.QueryRow(`INSERT INTO parallels (Name, BestClassID, ClassTotalRating) VALUES ($1, $2, $3) RETURNING Id`,
		name, bestClassID, maxRating).Scan(&parallelID)
	if err != nil {
		return err
	}

	for _, classID := range classesIDs {
		_, err := tx.Exec(`INSERT INTO parallel_classes (ParallelID, ClassID) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			parallelID, classID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresRepository) UpdateClass(classID int, input classModels.ClassInput, login string) error {
	check, err := r.hasUserRoleLocked(login, string(ratingModels.RoleOwner))
	if err != nil || !check {
		return errors.New("no permission")
	}

	grade, letter, ok := ParseClass(input.Name)
	if !ok {
		return errors.New("invalid class name")
	}

	return r.updateClassInternal(classID, input.Name, grade, letter, input.TeacherLogin)
}

func (r *postgresRepository) updateClassInternal(classID int, name string, grade int, letter string, teacherLogin string) error {
	if classID <= 0 {
		return errors.New("class id is required")
	}
	class, err := r.getClassByIDLocked(classID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}
	return r.db.QueryRow(`
		UPDATE classes
		SET Name = $1, Grade = $2, Letter = $3, TeacherLogin = $4
		WHERE id = $5
	`, name, grade, letter, teacherLogin, classID).Err()
}

func (r *postgresRepository) saveClassInternal(name string, grade int, letter string, teacherLogin string) error {
	var classID int64
	err := r.db.QueryRow(`INSERT INTO classes (Name, Grade, Letter, TeacherLogin) VALUES ($1, $2, $3, $4) RETURNING Id`,
		name, grade, letter, teacherLogin).Scan(&classID)
	if err != nil {
		return err
	}
	return r.autoAssignClassToParallelLocked(int(classID))
}

func (r *postgresRepository) GetClassIDsByRange(minGrade, maxGrade int) ([]int, error) {
	rows, err := r.db.Query("SELECT Id FROM classes WHERE Grade >= $1 AND Grade <= $2", minGrade, maxGrade)
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

func (r *postgresRepository) hasUserRoleLocked(login string, roles ...string) (bool, error) {
	user, err := r.getUserByLoginLocked(login)
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

func (r *postgresRepository) saveClassTeacherLocked(name string, teacherLogin string) error {
	name = normalizeClassName(name)
	teacherLogin = normalizeLogin(teacherLogin)

	if name == "" {
		return errors.New("class is required")
	}
	if teacherLogin == "" {
		return errors.New("teacher login is required")
	}

	teacher, err := r.getUserByLoginLocked(teacherLogin)
	if err != nil {
		return err
	}
	if teacher == nil {
		return errors.New("teacher not found")
	}
	class, err := r.getClassByNameLocked(name)
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

	if err := r.ensureClassLocked(name); err != nil {
		return err
	}

	_, err = r.db.Exec(`
		UPDATE classes
		SET TeacherLogin = $1
		WHERE Name = $2
	`, teacher.Login, name)
	if err != nil {
		return err
	}

	return r.syncClassLocked(name)
}

func (r *postgresRepository) GetClassesInParallel(id int) ([]classModels.Class, error) {
	rows, err := r.db.Query(`
        SELECT c.Id, c.Name, c.Grade, c.Letter, c.TeacherLogin, c.Members,
               c.UserTotalRating, c.ClassTotalRating,
               c.FirstQuarterComplete, c.SecondQuarterComplete, c.ThirdQuarterComplete
        FROM classes c
        JOIN parallel_classes pc ON pc.ClassID = c.Id
        WHERE pc.ParallelID = $1
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
		class, err := r.scanClassRowsLocked(rows)
		if err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.loadClassTeachersLocked(classes); err != nil {
		return nil, err
	}

	return classes, nil
}

func (r *postgresRepository) GetAllClasses() ([]classModels.Class, error) {
	rows, err := r.db.Query(`
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
		class, err := r.scanClassRowsLocked(rows)
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
		return nil, err
	}

	if err := r.loadClassTeachersLocked(classes); err != nil {
		return nil, err
	}

	return classes, nil
}

func (r *postgresRepository) SaveClassTeacherByID(classID int, teacherLogin string) error {
	if classID <= 0 {
		return errors.New("class id is required")
	}

	class, err := r.getClassByIDLocked(classID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}

	return r.saveClassTeacherLocked(class.Name, teacherLogin)
}

func (r *postgresRepository) GetClassByID(id int) (*classModels.Class, error) {
	if id <= 0 {
		return nil, errors.New("class id is required")
	}

	return r.getClassByIDLocked(id)
}

func (r *postgresRepository) GetClassTeacherByID(classID int) (*userModels.SafeUser, error) {
	if classID <= 0 {
		return nil, errors.New("class id is required")
	}

	class, err := r.getClassByIDLocked(classID)
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

	return r.getSafeUserByLoginLocked(class.TeacherLogin)
}

func (r *postgresRepository) GetUsersByClassID(classID int) ([]userModels.SafeUser, error) {
	if classID <= 0 {
		return nil, errors.New("class id is required")
	}

	class, err := r.getClassByIDLocked(classID)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, nil
	}

	return r.getUsersByClassIDLocked(classID)
}

func (r *postgresRepository) syncAllClassesLocked() error {
	if _, err := r.db.Exec(`
		INSERT INTO classes (Name)
		SELECT DISTINCT UPPER(TRIM(Class))
		FROM users
		WHERE TRIM(Class) <> ''
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := r.db.Exec(`
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
		);
	`); err != nil {
		return err
	}

	if _, err := r.db.Exec(`
		UPDATE users
		SET ClassID = (
			SELECT Id
			FROM classes
			WHERE LOWER(classes.TeacherLogin) = LOWER(TRIM(users.Login))
		)
		WHERE EXISTS (
			SELECT 1
			FROM classes
			WHERE LOWER(classes.TeacherLogin) = LOWER(TRIM(users.Login))
		);
	`); err != nil {
		return err
	}

	rows, err := r.db.Query(`SELECT Id FROM classes`)
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
		if err := r.syncClassByIDLocked(classID); err != nil {
			return err
		}
	}

	return nil
}

func (r *postgresRepository) syncClassLocked(name string) error {
	name = normalizeClassName(name)
	if name == "" {
		return nil
	}

	if err := r.ensureClassLocked(name); err != nil {
		return err
	}

	classID, err := r.getClassIDByNameLocked(name)
	if err != nil {
		return err
	}
	if classID == 0 {
		return nil
	}

	return r.syncClassByIDLocked(classID)
}

func (r *postgresRepository) syncClassByIDLocked(classID int) error {
	if classID <= 0 {
		return nil
	}

	class, err := r.getClassByIDLocked(classID)
	if err != nil {
		return err
	}
	if class == nil {
		return nil
	}

	members, err := r.getUsersByClassIDLocked(classID)
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
	err = r.db.QueryRow(`SELECT TeacherLogin FROM classes WHERE id = $1`, classID).Scan(&teacherLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if teacherLogin.Valid && teacherLogin.String != "" {
		teacher, err := r.getUserByLoginLocked(teacherLogin.String)
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
		candidate, err := r.findTeacherCandidateLocked(classID)
		if err != nil {
			return err
		}
		if candidate != "" {
			teacherLogin = sql.NullString{String: candidate, Valid: true}
		}
	}

	_, err = r.db.Exec(`
		UPDATE classes
		SET Members = $1, UserTotalRating = $2, TeacherLogin = $3
		WHERE id = $4
	`, string(membersJSON), userTotalRating, teacherLogin, classID)
	return err
}

func (r *postgresRepository) ensureClassLocked(name string) error {
	name = normalizeClassName(name)
	if name == "" {
		return nil
	}

	_, err := r.db.Exec(`
		INSERT INTO classes (Name)
		VALUES ($1)
		ON CONFLICT DO NOTHING
	`, name)
	return err
}

func (r *postgresRepository) getClassByNameLocked(name string) (*classModels.Class, error) {
	row := r.db.QueryRow(`
		SELECT Id, Name, Grade, Letter, TeacherLogin, Members, UserTotalRating, ClassTotalRating
		FROM classes
		WHERE Name = $1
	`, name)

	class, err := r.scanClassScannerLocked(row, true)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &class, nil
}

func (r *postgresRepository) getClassByIDLocked(id int) (*classModels.Class, error) {
	row := r.db.QueryRow(`
		SELECT Id, Name, Grade, Letter, TeacherLogin, Members, UserTotalRating, ClassTotalRating,
		FirstQuarterComplete, SecondQuarterComplete, ThirdQuarterComplete
		FROM classes
		WHERE id = $1
	`, id)

	class, err := r.scanClassScannerLocked(row, true)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &class, nil
}

func (r *postgresRepository) getClassIDByNameLocked(name string) (int, error) {
	var classID int
	err := r.db.QueryRow(`SELECT Id FROM classes WHERE Name = $1`, normalizeClassName(name)).Scan(&classID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return classID, nil
}

func (r *postgresRepository) getUsersByClassIDLocked(classID int) ([]userModels.SafeUser, error) {
	rows, err := r.db.Query(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		WHERE ClassID = $1
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

func (r *postgresRepository) scanClassScannerLocked(scanner classScanner, loadTeacher bool) (classModels.Class, error) {
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
			teacher, err := r.getSafeUserByLoginLocked(teacherLogin.String)
			if err != nil {
				return classModels.Class{}, err
			}
			class.Teacher = teacher
		}
	}

	return class, nil
}

func (r *postgresRepository) scanClassRowsLocked(rows *sql.Rows) (classModels.Class, error) {
	return r.scanClassScannerLocked(rows, false)
}

func (r *postgresRepository) loadClassTeachersLocked(classes []classModels.Class) error {
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
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = login
	}

	rows, err := r.db.Query(`
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

func (r *postgresRepository) findTeacherCandidateLocked(classID int) (string, error) {
	var login string
	err := r.db.QueryRow(`
		SELECT Login
		FROM users
		WHERE ClassID = $1
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

func (r *postgresRepository) InitializeParallelsFromConfig(parallels []classConfig.ParallelConfig) error {
	if len(parallels) == 0 {
		return nil
	}

	for _, parallelCfg := range parallels {
		if err := r.initializeParallelLocked(parallelCfg.Name, parallelCfg.MinGrade, parallelCfg.MaxGrade); err != nil {
			return err
		}
	}

	return nil
}

func (r *postgresRepository) initializeParallelLocked(name string, minGrade, maxGrade int) error {
	var parallelID int
	err := r.db.QueryRow(`
		SELECT Id FROM parallels
		WHERE MinGrade = $1 AND MaxGrade = $2
	`, minGrade, maxGrade).Scan(&parallelID)

	if err == nil {
		return r.syncParallelClassesLocked(parallelID, minGrade, maxGrade)
	}

	if err != sql.ErrNoRows {
		return err
	}

	rows, err := r.db.Query(`
		SELECT Id, ClassTotalRating FROM classes
		WHERE Grade >= $1 AND Grade <= $2
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

	if bestClassID == 0 {
		bestClassID = 0
		bestRating = 0
	}

	var newParallelID int64
	err = r.db.QueryRow(`
		INSERT INTO parallels (Name, MinGrade, MaxGrade, BestClassID, ClassTotalRating)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING Id
	`, name, minGrade, maxGrade, bestClassID, bestRating).Scan(&newParallelID)

	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil
		}
		return err
	}

	return r.syncParallelClassesLocked(int(newParallelID), minGrade, maxGrade)
}

func (r *postgresRepository) syncParallelClassesLocked(parallelID, minGrade, maxGrade int) error {
	rows, err := r.db.Query(`
		SELECT Id FROM classes
		WHERE Grade >= $1 AND Grade <= $2
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

	_, err = r.db.Exec(`DELETE FROM parallel_classes WHERE ParallelID = $1`, parallelID)
	if err != nil {
		return err
	}

	for _, classID := range classIDs {
		_, err := r.db.Exec(`
			INSERT INTO parallel_classes (ParallelID, ClassID)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, parallelID, classID)
		if err != nil {
			return err
		}
	}

	var bestClassID int
	var bestRating int
	err = r.db.QueryRow(`
		SELECT c.Id, c.ClassTotalRating FROM classes c
		WHERE c.Grade >= $1 AND c.Grade <= $2
		ORDER BY c.ClassTotalRating DESC
		LIMIT 1
	`, minGrade, maxGrade).Scan(&bestClassID, &bestRating)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if bestClassID == 0 {
		bestRating = 0
	}

	_, err = r.db.Exec(`
		UPDATE parallels
		SET BestClassID = $1, ClassTotalRating = $2
		WHERE id = $3
	`, bestClassID, bestRating, parallelID)

	return err
}

func (r *postgresRepository) autoAssignClassToParallelLocked(classID int) error {
	var grade int
	err := r.db.QueryRow(`SELECT Grade FROM classes WHERE id = $1`, classID).Scan(&grade)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	var parallelID int
	err = r.db.QueryRow(`
        SELECT Id FROM parallels
        WHERE MinGrade <= $1 AND MaxGrade >= $2
        LIMIT 1
    `, grade, grade).Scan(&parallelID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	_, err = r.db.Exec(`
        INSERT INTO parallel_classes (ParallelID, ClassID)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, parallelID, classID)

	return err
}

func (r *postgresRepository) getUserByLoginLocked(login string) (*userModels.User, error) {
	row := r.db.QueryRow(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Login = $1
	`, normalizeLogin(login))

	return scanUser(row)
}

func (r *postgresRepository) getSafeUserByLoginLocked(login string) (*userModels.SafeUser, error) {
	user, err := r.getUserByLoginLocked(login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &userModels.SafeUser{
		ID:       user.ID,
		Avatar:   user.Avatar,
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

type userScanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(scanner userScanner) (*userModels.User, error) {
	var user userModels.User
	var fullNameJSON sql.NullString

	err := scanner.Scan(
		&user.ID,
		&user.Avatar,
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
		user.FullName = []userModels.FullName{}
	}

	return &user, nil
}

func scanSafeUsers(rows *sql.Rows) ([]userModels.SafeUser, error) {
	users := make([]userModels.SafeUser, 0)

	for rows.Next() {
		var user userModels.SafeUser
		var fullNameJSON sql.NullString

		if err := rows.Scan(
			&user.ID,
			&user.Avatar,
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
			user.FullName = []userModels.FullName{}
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func normalizeLogin(login string) string {
	return strings.TrimSpace(login)
}

func normalizeClassName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

func isTeacherCandidate(role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	return role == strings.ToLower(string(ratingModels.RoleAdmin)) ||
		role == strings.ToLower(string(ratingModels.RoleOwner)) ||
		role == strings.ToLower(string(ratingModels.RoleHelper))
}
