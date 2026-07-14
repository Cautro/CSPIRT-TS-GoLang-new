package repo

import (
	userModels "cspirt/internal/domain/user"
	"cspirt/internal/domain/rating/repo"
	"cspirt/internal/controller/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.RatingRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) UpdateRating(login string, rating int) error {
	login = strings.TrimSpace(login)
	if login == "" {
		return errors.New("login is required")
	}

	var role string
	var classID int
	err := r.db.QueryRow(`SELECT Role, ClassID FROM users WHERE Login = $1`, login).Scan(&role, &classID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}

	if utils.IsSystemRole(role) {
		return errors.New("cannot update rating for system user")
	}

	rating = clampRating(rating)
	if _, err := r.db.Exec(`UPDATE users SET Rating = $1 WHERE Login = $2`, rating, login); err != nil {
		return err
	}

	return r.syncClassByIDLocked(classID)
}

func clampRating(rating int) int {
	if rating < 0 {
		return 0
	}
	return rating
}

// syncClassByIDLocked recomputes the class's cached Members/UserTotalRating/
// TeacherLogin snapshot after a rating change — mirrors users/repo's
// syncClassByIDLocked exactly, since repos don't depend on one another.
func (r *postgresRepository) syncClassByIDLocked(classID int) error {
	if classID <= 0 {
		return nil
	}

	var exists bool
	if err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM classes WHERE id = $1)`, classID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
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

func (r *postgresRepository) getUserByLoginLocked(login string) (*userModels.User, error) {
	row := r.db.QueryRow(`
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Login = $1
	`, strings.TrimSpace(login))

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
