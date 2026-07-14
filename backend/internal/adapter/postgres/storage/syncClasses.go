package storage

import (
	userModels "cspirt/internal/domain/user"
	"cspirt/internal/controller/utils"
	"database/sql"
	"encoding/json"
)

// syncAllClasses reconciles the classes table against the users table
// (assigns ClassID by class name/teacher login, refreshes each class's
// cached Members/UserTotalRating/TeacherLogin snapshot). It runs once at
// schema init time and after seeding test fixtures — a bootstrap concern of
// the storage layer, not a per-request business operation, so it stays
// here rather than in a repository.
func syncAllClasses(db *sql.DB) error {
	if _, err := db.Exec(`
		INSERT INTO classes (Name)
		SELECT DISTINCT UPPER(TRIM(Class))
		FROM users
		WHERE TRIM(Class) <> ''
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
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

	if _, err := db.Exec(`
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

	rows, err := db.Query(`SELECT Id FROM classes`)
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
		if err := syncClassByIDLocked(db, classID); err != nil {
			return err
		}
	}

	return nil
}

func syncClassByIDLocked(db *sql.DB, classID int) error {
	if classID <= 0 {
		return nil
	}

	var exists bool
	if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM classes WHERE id = $1)`, classID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return nil
	}

	members, err := getUsersByClassIDLocked(db, classID)
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
	err = db.QueryRow(`SELECT TeacherLogin FROM classes WHERE id = $1`, classID).Scan(&teacherLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if teacherLogin.Valid && teacherLogin.String != "" {
		teacher, err := getUserByLoginLocked(db, teacherLogin.String)
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
		candidate, err := findTeacherCandidateLocked(db, classID)
		if err != nil {
			return err
		}
		if candidate != "" {
			teacherLogin = sql.NullString{String: candidate, Valid: true}
		}
	}

	_, err = db.Exec(`
		UPDATE classes
		SET Members = $1, UserTotalRating = $2, TeacherLogin = $3
		WHERE id = $4
	`, string(membersJSON), userTotalRating, teacherLogin, classID)
	return err
}

func getUsersByClassIDLocked(db *sql.DB, classID int) ([]userModels.SafeUser, error) {
	rows, err := db.Query(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		WHERE ClassID = $1
		ORDER BY LastName, Name, Login
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func getUserByLoginLocked(db *sql.DB, login string) (*userModels.User, error) {
	row := db.QueryRow(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Login = $1
	`, login)

	var user userModels.User
	var fullNameJSON sql.NullString

	err := row.Scan(
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

func findTeacherCandidateLocked(db *sql.DB, classID int) (string, error) {
	var login string
	err := db.QueryRow(`
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
