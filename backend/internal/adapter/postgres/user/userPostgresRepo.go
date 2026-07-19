package postgres

import (
	entity "cspirt/internal/domain/user"
	ratMod "cspirt/internal/domain/rating"
	"cspirt/internal/domain/user/repo" 
	"cspirt/internal/controller/utils"
	middleware "cspirt/internal/controller/http/middleware-JWT"
	"database/sql"
	"cspirt/pkg/logger"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"context"
)

type postgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) repo.UserRepository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) GetOnlyStaffUsers(ctx context.Context) ([]entity.SafeUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT Id, Avatar, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		WHERE LOWER(Role) IN ('admin', 'owner')
		ORDER BY LastName, Name, Login
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSafeUsers(rows)
}

func (r *postgresRepository) SaveDeviceToken(ctx context.Context, userID int64, token, platform string) error {
	query := `
		INSERT INTO user_devices (user_id, device_token, platform) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (device_token) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID, token, platform)
	return err
}

func (r *postgresRepository) DeleteToken(ctx context.Context, token string) error {
	query := `DELETE FROM user_devices WHERE device_token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *postgresRepository) GetTokensByUserID(ctx context.Context, userID int64) ([]string, error) {
	query := `SELECT device_token FROM user_devices WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (r *postgresRepository) UpdateAvatar(ctx context.Context, input entity.UpdateAvatarRequest, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.db.Exec(`
		UPDATE users
		SET Avatar = $1
		WHERE Id = $2
	`,
		input.Avatar,
		id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *postgresRepository) AddUser(ctx context.Context, user entity.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	trimUserInput(&user)

	role, err := normalizeRole(user.Role)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
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

	if err := validateNewUser(&user); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: err.Error(),
		})
		return err
	}

	fullNameJSON, err := json.Marshal(user.FullName)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "failed to marshal full name: " + err.Error(),
		})
		return err
	}

	passwordHash, err := middleware.HashPassword(user.Password)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "failed to hash password: " + err.Error(),
		})
		return err
	}

	if !utils.IsSystemRole(user.Role) {
		if err := r.resolveUserClassLocked(&user); err != nil {
			return err
		}
	}

	query := `
		INSERT INTO users
		(Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Exec(
		query,
		user.Avatar,
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
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "add_user",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "failed to insert user: " + err.Error(),
		})
		return err
	}

	if user.ClassID > 0 {
		if err := r.syncClassByIDLocked(user.ClassID); err != nil {
			return err
		}
	}

	return nil
}

func (r *postgresRepository) SaveUser(ctx context.Context, user entity.SafeUser) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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

	oldUser, err := r.getUserByIDLocked(ctx, user.ID)
	if err != nil {
		return err
	}
	if oldUser == nil {
		return errors.New("user not found")
	}
	if err := r.resolveSafeUserClassLocked(&user); err != nil {
		return err
	}

	fullNameJSON, err := json.Marshal(user.FullName)
	if err != nil {
		return err
	}
	if string(fullNameJSON) == "null" {
		fullNameJSON = []byte("[]")
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "save_user",
		Login:   user.Login,
		Role:    user.Role,
		Class:   user.Class,
		Message: "saving user",
	})

	query := `
		UPDATE users
		SET Avatar = $1, Name = $2, FullName = $3, LastName = $4, Login = $5, Rating = $6, Role = $7, Class = $8, ClassID = $9
		WHERE Id = $10
	`

	result, err := r.db.Exec(
		query,
		user.Avatar,
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
		logger.WriteSafe(logger.LogEntry{
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

	if err := r.syncClassByIDLocked(user.ClassID); err != nil {
		return err
	}
	if oldUser.ClassID != 0 && oldUser.ClassID != user.ClassID {
		if err := r.syncClassByIDLocked(oldUser.ClassID); err != nil {
			return err
		}
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "save_user",
		Login:   user.Login,
		Role:    user.Role,
		Class:   user.Class,
		Message: "user saved",
	})
	return nil
}

func (r *postgresRepository) UpdateUser(ctx context.Context, id int, req entity.UpdateUserRequest, login string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if id <= 0 {
		return errors.New("user id is required")
	}

	oldUser, err := r.getUserByIDLocked(ctx, id)
	if err != nil {
		return err
	}
	if oldUser == nil {
		return errors.New("user not found")
	}

	user := *oldUser

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return errors.New("name is required")
		}
		user.Name = name
	}

	if req.LastName != nil {
		lastName := strings.TrimSpace(*req.LastName)
		if lastName == "" {
			return errors.New("last name is required")
		}
		user.LastName = lastName
	}

	if req.Avatar != nil {
		trimmedValue := strings.TrimSpace(req.Avatar.String)

		user.Avatar = sql.NullString{
			String: trimmedValue,
			Valid:  true,
		}
	}

	if req.Login != nil {
		login := normalizeLogin(*req.Login)
		if login == "" {
			return errors.New("login is required")
		}
		user.Login = login
	}

	if req.Rating != nil {
		user.Rating = clampRating(*req.Rating)
	}

	if req.Role != nil {
		role, err := normalizeRole(*req.Role)
		if err != nil {
			return err
		}
		user.Role = role
	}

	if req.ClassID != nil {
		user.ClassID = *req.ClassID
	}

	if req.Class != nil {
		user.Class = normalizeClassName(*req.Class)
	}
	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	safeUser := entity.SafeUser{
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
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "update_user",
		Login:   safeUser.Login,
		Role:    safeUser.Role,
		Class:   safeUser.Class,
		Message: "updating user",
	})

	if err := r.resolveSafeUserClassLocked(&safeUser); err != nil {
		return err
	}

	fullNameJSON, err := json.Marshal(user.FullName)
	if err != nil {
		return err
	}
	if string(fullNameJSON) == "null" {
		fullNameJSON = []byte("[]")
	}

	_, err = r.db.Exec(`
		UPDATE users
		SET Avatar = $1, Name = $2, FullName = $3, LastName = $4, Login = $5, Rating = $6, Role = $7, Class = $8, ClassID = $9
		WHERE Id = $10
	`,
		user.Avatar,
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
		return err
	}

	return nil
}

func (r *postgresRepository) DeleteUser(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_user",
		Message: "deleting user",
	})

	deletedUser, err := r.getUserByIDLocked(ctx, id)
	if err != nil {
		return err
	}
	if deletedUser == nil {
		return errors.New("user not found")
	}

	query := `DELETE FROM users WHERE Id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_user",
			Message: "failed to delete user: " + err.Error(),
		})
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("user not found")
	}

	if deletedUser.ClassID != 0 {
		if err := r.syncClassByIDLocked(deletedUser.ClassID); err != nil {
			return err
		}
	} else if err := r.syncClassLocked(deletedUser.Class); err != nil {
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_user",
		Message: "user deleted",
	})
	return nil
}

func (r *postgresRepository) GetAllUsers(ctx context.Context) ([]entity.SafeUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_all_users",
		Message: "getting all users",
	})

	rows, err := r.db.QueryContext(ctx, `
		SELECT Id, Avatar, Name, FullName, LastName, Login, Rating, Role, Class, ClassID
		FROM users
		ORDER BY ClassID, LastName, Name, Login
	`)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_users",
			Message: "failed to query users: " + err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	users, err := scanSafeUsers(rows)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_all_users",
			Message: "failed to scan users: " + err.Error(),
		})
		return nil, err
	}

	return users, nil
}

func (r *postgresRepository) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.getUserByLoginLocked(ctx, login)
}

func (r *postgresRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.getUserByIDLocked(ctx, id)
}

func (r *postgresRepository) GetUsersByClassID(ctx context.Context, classID int) ([]entity.SafeUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if classID <= 0 {
		return nil, errors.New("class id is required")
	}

	return r.getUsersByClassIDLocked(ctx, classID)
}

func (r *postgresRepository) SaveRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, token, expiresAt)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "save_refresh_token",
			Message: "failed to save refresh token: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "save_refresh_token",
		Message: "refresh token saved",
	})
	return nil
}

func (r *postgresRepository) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`, token)

	var rt entity.RefreshToken

	err := row.Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_refresh_token",
			Message: "failed to get refresh token: " + err.Error(),
		})
		return nil, err
	}

	return &rt, nil
}

func (r *postgresRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`, token)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_refresh_token",
			Message: "failed to delete refresh token: " + err.Error(),
		})
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_refresh_token",
		Message: "refresh token deleted",
	})
	return nil
}

func validateNewUser(user *entity.User) error {
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
	if !utils.IsSystemRole(user.Role) && user.Class == "" && user.ClassID <= 0 && user.Role != string(ratMod.RolePublic) {
		return errors.New("class is required")
	}
	if user.Rating < 0 {
		return errors.New("rating must be non-negative")
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
	return rating
}

func (r *postgresRepository) resolveUserClassLocked(user *entity.User) error {
	if user.ClassID > 0 {
		name, ok, err := r.classNameByIDLocked(user.ClassID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("class not found")
		}
		user.Class = name
		return nil
	}

	if user.Class == "" && user.Role != string(ratMod.RolePublic) {
		return errors.New("class is required")
	}
	if err := r.ensureClassLocked(user.Class); err != nil {
		return err
	}

	classID, err := r.getClassIDByNameLocked(user.Class)
	if err != nil {
		return err
	}
	if classID == 0 && user.Role != string(ratMod.RolePublic) {
		return errors.New("class not found")
	}

	user.ClassID = classID
	return nil
}

func (r *postgresRepository) resolveSafeUserClassLocked(user *entity.SafeUser) error {
	if user.ClassID > 0 {
		name, ok, err := r.classNameByIDLocked(user.ClassID)
		if err != nil {
			return err
		}
		if !ok && user.Role != string(ratMod.RolePublic) {
			return errors.New("class not found")
		}
		user.Class = name
		return nil
	}

	if user.Class == "" && user.Role != string(ratMod.RolePublic) {
		return errors.New("class is required")
	}
	if err := r.ensureClassLocked(user.Class); err != nil {
		return err
	}

	classID, err := r.getClassIDByNameLocked(user.Class)
	if err != nil {
		return err
	}
	if classID == 0 && user.Role != string(ratMod.RolePublic) {
		return errors.New("class not found")
	}

	user.ClassID = classID
	return nil
}

func (r *postgresRepository) classNameByIDLocked(id int) (string, bool, error) {
	var name string
	err := r.db.QueryRow(`SELECT Name FROM classes WHERE id = $1`, id).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}
	return name, true, nil
}

func (r *postgresRepository) classExistsByIDLocked(id int) (bool, error) {
	_, ok, err := r.classNameByIDLocked(id)
	return ok, err
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

	exists, err := r.classExistsByIDLocked(classID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	members, err := r.getUsersByClassIDLocked(context.Background(), classID)
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
		teacher, err := r.getUserByLoginLocked(context.Background(), teacherLogin.String)
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

func (r *postgresRepository) getSafeUserByLoginLocked(login string) (*entity.SafeUser, error) {
	user, err := r.getUserByLoginLocked(context.Background(), login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &entity.SafeUser{
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

func (r *postgresRepository) getUserByLoginLocked(ctx context.Context, login string) (*entity.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Login = $1
	`, normalizeLogin(login))

	return scanUser(row)
}

func (r *postgresRepository) getUserByIDLocked(ctx context.Context, id int) (*entity.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT Id, Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class, ClassID
		FROM users
		WHERE Id = $1
	`, id)

	return scanUser(row)
}

func (r *postgresRepository) getUsersByClassIDLocked(ctx context.Context, classID int) ([]entity.SafeUser, error) {
	rows, err := r.db.QueryContext(ctx, `
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

type userScanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(scanner userScanner) (*entity.User, error) {
	var user entity.User
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
		user.FullName = []entity.FullName{}
	}

	return &user, nil
}

func scanSafeUsers(rows *sql.Rows) ([]entity.SafeUser, error) {
	users := make([]entity.SafeUser, 0)

	for rows.Next() {
		var user entity.SafeUser
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
			user.FullName = []entity.FullName{}
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

func normalizeRole(role string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case strings.ToLower(string(ratMod.RoleAdmin)):
		return string(ratMod.RoleAdmin), nil
	case strings.ToLower(string(ratMod.RoleUser)):
		return string(ratMod.RoleUser), nil
	case strings.ToLower(string(ratMod.RoleHelper)):
		return string(ratMod.RoleHelper), nil
	case strings.ToLower(string(ratMod.RoleOwner)):
		return string(ratMod.RoleOwner), nil
	case strings.ToLower(string(ratMod.RolePublic)):
		return string(ratMod.RolePublic), nil
	default:
		return "", errors.New("invalid role")
	}
}

func trimUserInput(user *entity.User) {
	user.Name = strings.TrimSpace(user.Name)
	user.LastName = strings.TrimSpace(user.LastName)
	user.Login = normalizeLogin(user.Login)
	user.Class = normalizeClassName(user.Class)

	for i := range user.FullName {
		user.FullName[i].Name = strings.TrimSpace(user.FullName[i].Name)
		user.FullName[i].LastName = strings.TrimSpace(user.FullName[i].LastName)
	}
}
