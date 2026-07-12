package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cspirt/internal/logger"
	scheduleModels "cspirt/internal/schedule/models"
	userModels "cspirt/internal/users/models"
)

func (s *Storage) GetSchedules(filter scheduleModels.ScheduleFilter) (*scheduleModels.SchedulesResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filter = normalizeScheduleFilter(filter)

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "get_schedules",
		Message: "getting schedules",
	})

	result := &scheduleModels.SchedulesResponse{
		Schedules: []scheduleModels.ScheduleLesson{},
		Base:      []scheduleModels.ScheduleLesson{},
		Current:   []scheduleModels.ScheduleLesson{},
		Planned:   []scheduleModels.ScheduleLesson{},
	}

	load := func(scheduleType string) ([]scheduleModels.ScheduleLesson, error) {
		return s.getScheduleLessonsLocked(scheduleType, filter, 0)
	}

	switch filter.Type {
	case scheduleModels.ScheduleTypeAll:
		base, err := load(scheduleModels.ScheduleTypeBase)
		if err != nil {
			return nil, err
		}
		current, err := load(scheduleModels.ScheduleTypeCurrent)
		if err != nil {
			return nil, err
		}
		planned, err := load(scheduleModels.ScheduleTypePlanned)
		if err != nil {
			return nil, err
		}
		result.Base = base
		result.Current = current
		result.Planned = planned
		result.Schedules = append(result.Schedules, base...)
		result.Schedules = append(result.Schedules, current...)
		result.Schedules = append(result.Schedules, planned...)
	case scheduleModels.ScheduleTypeBase:
		base, err := load(scheduleModels.ScheduleTypeBase)
		if err != nil {
			return nil, err
		}
		result.Base = base
		result.Schedules = base
	case scheduleModels.ScheduleTypePlanned:
		planned, err := load(scheduleModels.ScheduleTypePlanned)
		if err != nil {
			return nil, err
		}
		result.Planned = planned
		result.Schedules = planned
	default:
		current, err := load(scheduleModels.ScheduleTypeCurrent)
		if err != nil {
			return nil, err
		}
		result.Current = current
		result.Schedules = current
	}

	return result, nil
}

func (s *Storage) GetCurrentScheduleForTeacher(teacherID int, filter scheduleModels.ScheduleFilter) ([]scheduleModels.ScheduleLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateScheduleTeacherLocked(teacherID); err != nil {
		return nil, err
	}

	filter = normalizeScheduleFilter(filter)
	filter.Type = scheduleModels.ScheduleTypeCurrent
	filter.TeacherID = teacherID

	return s.getScheduleLessonsLocked(scheduleModels.ScheduleTypeCurrent, filter, 0)
}

func (s *Storage) GetScheduleLessonByID(scheduleType string, id int) (*scheduleModels.ScheduleLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduleType, err := normalizeScheduleType(scheduleType)
	if err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, errors.New("schedule lesson id is required")
	}

	return s.getScheduleLessonByIDLocked(scheduleType, id)
}

func (s *Storage) UpsertScheduleLesson(scheduleType string, lesson scheduleModels.ScheduleLesson) (*scheduleModels.ScheduleLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduleType, err := normalizeScheduleType(scheduleType)
	if err != nil {
		return nil, err
	}

	normalizeScheduleLessonStorage(&lesson)
	if err := s.resolveScheduleLessonRefsLocked(scheduleType, &lesson); err != nil {
		return nil, err
	}
	if err := s.validateScheduleLessonRefsLocked(lesson); err != nil {
		return nil, err
	}

	if lesson.ID <= 0 {
		if scheduleType == scheduleModels.ScheduleTypePlanned && lesson.BaseScheduleID != nil {
			existing, err := s.getPlannedScheduleByBaseIDLocked(*lesson.BaseScheduleID)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				lesson.ID = existing.ID
			}
		}
	}
	if lesson.ID <= 0 {
		existing, err := s.getScheduleLessonByUniqueKeyLocked(scheduleType, lesson.ClassID, lesson.DayOfWeek, lesson.LessonNumber, lesson.WeekType)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			lesson.ID = existing.ID
		}
	}

	if lesson.ID > 0 {
		if err := s.updateScheduleLessonLocked(scheduleType, lesson); err != nil {
			return nil, err
		}
		return s.getScheduleLessonByIDLocked(scheduleType, lesson.ID)
	}

	id, err := s.insertScheduleLessonLocked(scheduleType, lesson)
	if err != nil {
		return nil, err
	}

	return s.getScheduleLessonByIDLocked(scheduleType, id)
}

func (s *Storage) DeleteScheduleLesson(scheduleType string, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduleType, err := normalizeScheduleType(scheduleType)
	if err != nil {
		return err
	}
	if id <= 0 {
		return errors.New("schedule lesson id is required")
	}

	table, err := scheduleTable(scheduleType)
	if err != nil {
		return err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_schedule_lesson",
		Message: "deleting schedule lesson",
	})

	result, err := s.db.Exec(`DELETE FROM `+table+` WHERE Id = $1`, id)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("schedule lesson not found")
	}

	return nil
}

func (s *Storage) RolloverSchedules(classID int) (*scheduleModels.ScheduleRolloverResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID > 0 {
		class, err := s.getClassByIDLocked(classID)
		if err != nil {
			return nil, err
		}
		if class == nil {
			return nil, errors.New("class not found")
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	plannedCount, err := countScheduleRowsTx(tx, "planned_schedules", classID, true)
	if err != nil {
		return nil, err
	}

	source := scheduleModels.ScheduleTypeBase
	sourceTable := "schedules"
	if plannedCount > 0 {
		source = scheduleModels.ScheduleTypePlanned
		sourceTable = "planned_schedules"
	}

	if err := deleteCurrentSchedulesTx(tx, classID); err != nil {
		return nil, err
	}

	currentCount, err := copyLessonsToCurrentTx(tx, sourceTable, classID, source == scheduleModels.ScheduleTypePlanned)
	if err != nil {
		return nil, err
	}

	plannedCleared, err := deletePlannedSchedulesTx(tx, classID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &scheduleModels.ScheduleRolloverResult{
		Source:         source,
		ClassID:        classID,
		CurrentCount:   currentCount,
		PlannedCleared: plannedCleared,
	}, nil
}

func (s *Storage) ResetPlannedSchedules(classID int) (*scheduleModels.ScheduleResetResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if classID > 0 {
		class, err := s.getClassByIDLocked(classID)
		if err != nil {
			return nil, err
		}
		if class == nil {
			return nil, errors.New("class not found")
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := deletePlannedSchedulesTx(tx, classID); err != nil {
		return nil, err
	}

	plannedCount, err := copyBaseLessonsToPlannedTx(tx, classID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &scheduleModels.ScheduleResetResult{
		Source:       scheduleModels.ScheduleTypeBase,
		ClassID:      classID,
		PlannedCount: plannedCount,
	}, nil
}

func (s *Storage) ensureCurrentSchedulesSeeded() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM current_schedules`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err := s.db.Exec(`
		INSERT INTO current_schedules
			(ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description)
		SELECT ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description
		FROM schedules
	`)
	return err
}

func (s *Storage) UpsertBaseSchedule(lesson scheduleModels.BaseSchedule) (*scheduleModels.BaseSchedule, error) {
	return s.UpsertScheduleLesson(scheduleModels.ScheduleTypeBase, lesson)
}

func (s *Storage) UpsertCurrentSchedule(lesson scheduleModels.CurrentSchedule) (*scheduleModels.CurrentSchedule, error) {
	return s.UpsertScheduleLesson(scheduleModels.ScheduleTypeCurrent, lesson)
}

func (s *Storage) UpsertPlannedSchedule(lesson scheduleModels.PlannedSchedule) (*scheduleModels.PlannedSchedule, error) {
	return s.UpsertScheduleLesson(scheduleModels.ScheduleTypePlanned, lesson)
}

func (s *Storage) DeleteBaseSchedule(id int) error {
	return s.DeleteScheduleLesson(scheduleModels.ScheduleTypeBase, id)
}

func (s *Storage) DeleteCurrentSchedule(id int) error {
	return s.DeleteScheduleLesson(scheduleModels.ScheduleTypeCurrent, id)
}

func (s *Storage) DeletePlannedSchedule(id int) error {
	return s.DeleteScheduleLesson(scheduleModels.ScheduleTypePlanned, id)
}

func (s *Storage) GetBaseScheduleByID(id int) (*scheduleModels.BaseSchedule, error) {
	return s.GetScheduleLessonByID(scheduleModels.ScheduleTypeBase, id)
}

func (s *Storage) GetCurrentScheduleByID(id int) (*scheduleModels.CurrentSchedule, error) {
	return s.GetScheduleLessonByID(scheduleModels.ScheduleTypeCurrent, id)
}

func (s *Storage) GetPlannedScheduleByID(id int) (*scheduleModels.PlannedSchedule, error) {
	return s.GetScheduleLessonByID(scheduleModels.ScheduleTypePlanned, id)
}

func (s *Storage) getScheduleLessonsLocked(scheduleType string, filter scheduleModels.ScheduleFilter, id int) ([]scheduleModels.ScheduleLesson, error) {
	table, err := scheduleTable(scheduleType)
	if err != nil {
		return nil, err
	}

	baseIDExpr := `NULL AS BaseScheduleID`
	createdAtExpr := `'' AS CreatedAt`
	if scheduleType == scheduleModels.ScheduleTypePlanned {
		baseIDExpr = `l.BaseScheduleID`
		createdAtExpr = `l.CreatedAt`
	}

	query := `
		SELECT l.Id, l.ClassID, c.Name, l.DayOfWeek, l.LessonNumber, l.WeekType,
			l.Subject, l.TeacherID,
			u.Id, u.Name, u.FullName, u.LastName, u.Login, u.Rating, u.Role, u.Class, u.ClassID,
			l.Room, l.StartTime, l.EndTime, l.Description,
			` + baseIDExpr + `, ` + createdAtExpr + `
		FROM ` + table + ` l
		JOIN classes c ON c.Id = l.ClassID
		JOIN users u ON u.Id = l.TeacherID
		WHERE 1 = 1
	`
	args := make([]interface{}, 0)

	if id > 0 {
		args = append(args, id)
		query += fmt.Sprintf(` AND l.Id = $%d`, len(args))
	}
	if filter.ClassID > 0 {
		args = append(args, filter.ClassID)
		query += fmt.Sprintf(` AND l.ClassID = $%d`, len(args))
	}
	if filter.TeacherID > 0 {
		args = append(args, filter.TeacherID)
		query += fmt.Sprintf(` AND l.TeacherID = $%d`, len(args))
	}
	if filter.Day != "" {
		args = append(args, filter.Day)
		query += fmt.Sprintf(` AND LOWER(l.DayOfWeek) = LOWER($%d)`, len(args))
	}
	if filter.WeekType != "" {
		if filter.WeekType == "all" {
			query += ` AND LOWER(l.WeekType) = 'all'`
		} else {
			args = append(args, filter.WeekType)
			query += fmt.Sprintf(` AND LOWER(l.WeekType) IN ('all', LOWER($%d))`, len(args))
		}
	}
	if scheduleType == scheduleModels.ScheduleTypePlanned {
		query += ` AND TRIM(l.DayOfWeek) <> ''`
	}

	query += `
		ORDER BY l.ClassID,
			CASE LOWER(l.DayOfWeek)
				WHEN 'monday' THEN 1
				WHEN 'tuesday' THEN 2
				WHEN 'wednesday' THEN 3
				WHEN 'thursday' THEN 4
				WHEN 'friday' THEN 5
				WHEN 'saturday' THEN 6
				WHEN 'sunday' THEN 7
				ELSE 8
			END,
			l.LessonNumber,
			l.StartTime,
			l.Id
	`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanScheduleLessons(rows, scheduleType)
}

func (s *Storage) getScheduleLessonByIDLocked(scheduleType string, id int) (*scheduleModels.ScheduleLesson, error) {
	lessons, err := s.getScheduleLessonsLocked(scheduleType, scheduleModels.ScheduleFilter{}, id)
	if err != nil {
		return nil, err
	}
	if len(lessons) == 0 {
		return nil, nil
	}

	return &lessons[0], nil
}

func (s *Storage) getScheduleLessonByUniqueKeyLocked(scheduleType string, classID int, day string, lessonNumber int, weekType string) (*scheduleModels.ScheduleLesson, error) {
	table, err := scheduleTable(scheduleType)
	if err != nil {
		return nil, err
	}

	var id int
	err = s.db.QueryRow(`
		SELECT Id
		FROM `+table+`
		WHERE ClassID = $1
			AND LOWER(DayOfWeek) = LOWER($2)
			AND LessonNumber = $3
			AND LOWER(WeekType) = LOWER($4)
		LIMIT 1
	`, classID, day, lessonNumber, weekType).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return s.getScheduleLessonByIDLocked(scheduleType, id)
}

func (s *Storage) getPlannedScheduleByBaseIDLocked(baseScheduleID int) (*scheduleModels.ScheduleLesson, error) {
	if baseScheduleID <= 0 {
		return nil, nil
	}

	var id int
	err := s.db.QueryRow(`
		SELECT Id
		FROM planned_schedules
		WHERE BaseScheduleID = $1
		LIMIT 1
	`, baseScheduleID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return s.getScheduleLessonByIDLocked(scheduleModels.ScheduleTypePlanned, id)
}

func (s *Storage) insertScheduleLessonLocked(scheduleType string, lesson scheduleModels.ScheduleLesson) (int, error) {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "insert_schedule_lesson",
		Message: "inserting schedule lesson",
	})

	if scheduleType == scheduleModels.ScheduleTypePlanned {
		var id int64
		err := s.db.QueryRow(`
			INSERT INTO planned_schedules
				(BaseScheduleID, ClassID, Date, DayOfWeek, LessonNumber, WeekType, Subject,
				 ChangeType, Scope, TeacherID, Room, StartTime, EndTime, Description, Reason, CreatedAt)
			VALUES ($1, $2, '', $3, $4, $5, $6, 'update', 'lesson', $7, $8, $9, $10, $11, '', $12)
			RETURNING Id
		`, nullableScheduleInt(lesson.BaseScheduleID), lesson.ClassID, lesson.DayOfWeek, lesson.LessonNumber,
			lesson.WeekType, lesson.Subject, lesson.TeacherID, lesson.Room, lesson.StartTime,
			lesson.EndTime, lesson.Description, time.Now().UTC().Format(time.RFC3339)).Scan(&id)
		return int(id), err
	}

	table, err := scheduleTable(scheduleType)
	if err != nil {
		return 0, err
	}

	var id int64
	err = s.db.QueryRow(`
		INSERT INTO `+table+`
			(ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING Id
	`, lesson.ClassID, lesson.DayOfWeek, lesson.LessonNumber, lesson.WeekType, lesson.Subject,
		lesson.TeacherID, lesson.Room, lesson.StartTime, lesson.EndTime, lesson.Description).Scan(&id)
	return int(id), err
}

func (s *Storage) updateScheduleLessonLocked(scheduleType string, lesson scheduleModels.ScheduleLesson) error {
	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "update_schedule_lesson",
		Message: "updating schedule lesson",
	})

	if scheduleType == scheduleModels.ScheduleTypePlanned {
		result, err := s.db.Exec(`
			UPDATE planned_schedules
			SET BaseScheduleID = $1, ClassID = $2, DayOfWeek = $3, LessonNumber = $4, WeekType = $5,
				Subject = $6, TeacherID = $7, Room = $8, StartTime = $9, EndTime = $10, Description = $11
			WHERE Id = $12
		`, nullableScheduleInt(lesson.BaseScheduleID), lesson.ClassID, lesson.DayOfWeek, lesson.LessonNumber,
			lesson.WeekType, lesson.Subject, lesson.TeacherID, lesson.Room, lesson.StartTime,
			lesson.EndTime, lesson.Description, lesson.ID)
		if err != nil {
			return err
		}
		if affected, err := result.RowsAffected(); err == nil && affected == 0 {
			return errors.New("schedule lesson not found")
		}
		return nil
	}

	table, err := scheduleTable(scheduleType)
	if err != nil {
		return err
	}

	result, err := s.db.Exec(`
		UPDATE `+table+`
		SET ClassID = $1, DayOfWeek = $2, LessonNumber = $3, WeekType = $4, Subject = $5,
			TeacherID = $6, Room = $7, StartTime = $8, EndTime = $9, Description = $10
		WHERE Id = $11
	`, lesson.ClassID, lesson.DayOfWeek, lesson.LessonNumber, lesson.WeekType, lesson.Subject,
		lesson.TeacherID, lesson.Room, lesson.StartTime, lesson.EndTime, lesson.Description, lesson.ID)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("schedule lesson not found")
	}

	return nil
}

func (s *Storage) resolveScheduleLessonRefsLocked(scheduleType string, lesson *scheduleModels.ScheduleLesson) error {
	if scheduleType != scheduleModels.ScheduleTypePlanned || lesson.BaseScheduleID == nil {
		return nil
	}

	base, err := s.getScheduleLessonByIDLocked(scheduleModels.ScheduleTypeBase, *lesson.BaseScheduleID)
	if err != nil {
		return err
	}
	if base == nil {
		return errors.New("base schedule not found")
	}

	if lesson.ClassID <= 0 {
		lesson.ClassID = base.ClassID
	}
	if lesson.ClassID != base.ClassID {
		return errors.New("planned schedule class does not match base schedule class")
	}
	if lesson.DayOfWeek == "" {
		lesson.DayOfWeek = base.DayOfWeek
	}
	if lesson.LessonNumber <= 0 {
		lesson.LessonNumber = base.LessonNumber
	}
	if lesson.WeekType == "" {
		lesson.WeekType = base.WeekType
	}
	if lesson.Subject == "" {
		lesson.Subject = base.Subject
	}
	if lesson.TeacherID <= 0 {
		lesson.TeacherID = base.TeacherID
	}
	if lesson.Room <= 0 {
		lesson.Room = base.Room
	}
	if lesson.StartTime == "" {
		lesson.StartTime = base.StartTime
	}
	if lesson.EndTime == "" {
		lesson.EndTime = base.EndTime
	}
	if lesson.Description == "" {
		lesson.Description = base.Description
	}

	return nil
}

func (s *Storage) validateScheduleLessonRefsLocked(lesson scheduleModels.ScheduleLesson) error {
	if lesson.ClassID <= 0 {
		return errors.New("class id is required")
	}
	class, err := s.getClassByIDLocked(lesson.ClassID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}
	if lesson.DayOfWeek == "" {
		return errors.New("day of week is required")
	}
	if lesson.LessonNumber <= 0 {
		return errors.New("lesson number is required")
	}
	if lesson.Subject == "" {
		return errors.New("subject is required")
	}
	if lesson.Room <= 0 {
		return errors.New("room is required")
	}
	if lesson.StartTime == "" || lesson.EndTime == "" {
		return errors.New("start time and end time are required")
	}

	return s.validateScheduleTeacherLocked(lesson.TeacherID)
}

func (s *Storage) validateScheduleTeacherLocked(teacherID int) error {
	if teacherID <= 0 {
		return errors.New("teacher id is required")
	}

	teacher, err := s.getUserByIDLocked(teacherID)
	if err != nil {
		return err
	}
	if teacher == nil {
		return errors.New("teacher not found")
	}
	if !isTeacherCandidate(teacher.Role) {
		return errors.New("teacher must have helper, admin or owner role")
	}

	return nil
}

func scanScheduleLessons(rows *sql.Rows, scheduleType string) ([]scheduleModels.ScheduleLesson, error) {
	lessons := make([]scheduleModels.ScheduleLesson, 0)

	for rows.Next() {
		lesson, err := scanScheduleLesson(rows, scheduleType)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lessons, nil
}

func scanScheduleLesson(scanner scheduleScanner, scheduleType string) (scheduleModels.ScheduleLesson, error) {
	var lesson scheduleModels.ScheduleLesson
	var teacher userModels.SafeUser
	var teacherFullNameJSON sql.NullString
	var baseScheduleID sql.NullInt64
	var createdAt sql.NullString

	if err := scanner.Scan(
		&lesson.ID,
		&lesson.ClassID,
		&lesson.Class,
		&lesson.DayOfWeek,
		&lesson.LessonNumber,
		&lesson.WeekType,
		&lesson.Subject,
		&lesson.TeacherID,
		&teacher.ID,
		&teacher.Name,
		&teacherFullNameJSON,
		&teacher.LastName,
		&teacher.Login,
		&teacher.Rating,
		&teacher.Role,
		&teacher.Class,
		&teacher.ClassID,
		&lesson.Room,
		&lesson.StartTime,
		&lesson.EndTime,
		&lesson.Description,
		&baseScheduleID,
		&createdAt,
	); err != nil {
		return scheduleModels.ScheduleLesson{}, err
	}

	if teacherFullNameJSON.Valid && teacherFullNameJSON.String != "" {
		if err := json.Unmarshal([]byte(teacherFullNameJSON.String), &teacher.FullName); err != nil {
			return scheduleModels.ScheduleLesson{}, err
		}
	}
	if teacher.FullName == nil {
		teacher.FullName = []userModels.FullName{}
	}

	lesson.Type = scheduleType
	lesson.Teacher = &teacher
	lesson.BaseScheduleID = intPtrFromNull(baseScheduleID)
	if createdAt.Valid {
		lesson.CreatedAt = createdAt.String
	}

	return lesson, nil
}

type scheduleScanner interface {
	Scan(dest ...interface{}) error
}

func normalizeScheduleFilter(filter scheduleModels.ScheduleFilter) scheduleModels.ScheduleFilter {
	filter.Type = strings.ToLower(strings.TrimSpace(filter.Type))
	if filter.Type == "" {
		filter.Type = scheduleModels.ScheduleTypeCurrent
	}
	filter.Day = strings.TrimSpace(filter.Day)
	filter.WeekType = normalizeScheduleWeekType(filter.WeekType)
	return filter
}

func normalizeScheduleLessonStorage(lesson *scheduleModels.ScheduleLesson) {
	lesson.Type = strings.ToLower(strings.TrimSpace(lesson.Type))
	lesson.DayOfWeek = strings.TrimSpace(lesson.DayOfWeek)
	lesson.WeekType = normalizeScheduleWeekType(lesson.WeekType)
	lesson.Subject = strings.TrimSpace(lesson.Subject)
	lesson.StartTime = strings.TrimSpace(lesson.StartTime)
	lesson.EndTime = strings.TrimSpace(lesson.EndTime)
	lesson.Description = strings.TrimSpace(lesson.Description)
}

func normalizeScheduleWeekType(weekType string) string {
	weekType = strings.ToLower(strings.TrimSpace(weekType))
	if weekType == "" {
		return "all"
	}
	return weekType
}

func normalizeScheduleType(scheduleType string) (string, error) {
	scheduleType = strings.ToLower(strings.TrimSpace(scheduleType))
	switch scheduleType {
	case "", scheduleModels.ScheduleTypeCurrent:
		return scheduleModels.ScheduleTypeCurrent, nil
	case scheduleModels.ScheduleTypeBase:
		return scheduleModels.ScheduleTypeBase, nil
	case scheduleModels.ScheduleTypePlanned:
		return scheduleModels.ScheduleTypePlanned, nil
	default:
		return "", errors.New("invalid schedule type")
	}
}

func scheduleTable(scheduleType string) (string, error) {
	switch scheduleType {
	case scheduleModels.ScheduleTypeBase:
		return "schedules", nil
	case scheduleModels.ScheduleTypeCurrent:
		return "current_schedules", nil
	case scheduleModels.ScheduleTypePlanned:
		return "planned_schedules", nil
	default:
		return "", errors.New("invalid schedule type")
	}
}

func countScheduleRowsTx(tx *sql.Tx, table string, classID int, onlyLessons bool) (int, error) {
	query := `SELECT COUNT(*) FROM ` + table + ` WHERE 1 = 1`
	args := make([]interface{}, 0)

	if classID > 0 {
		args = append(args, classID)
		query += fmt.Sprintf(` AND ClassID = $%d`, len(args))
	}
	if onlyLessons {
		query += ` AND TRIM(DayOfWeek) <> ''`
	}

	var count int
	err := tx.QueryRow(query, args...).Scan(&count)
	return count, err
}

func deleteCurrentSchedulesTx(tx *sql.Tx, classID int) error {
	query := `DELETE FROM current_schedules`
	args := make([]interface{}, 0)
	if classID > 0 {
		args = append(args, classID)
		query += fmt.Sprintf(` WHERE ClassID = $%d`, len(args))
	}

	_, err := tx.Exec(query, args...)
	return err
}

func deletePlannedSchedulesTx(tx *sql.Tx, classID int) (int, error) {
	query := `DELETE FROM planned_schedules`
	args := make([]interface{}, 0)
	if classID > 0 {
		args = append(args, classID)
		query += fmt.Sprintf(` WHERE ClassID = $%d`, len(args))
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}

func copyLessonsToCurrentTx(tx *sql.Tx, sourceTable string, classID int, onlyValidLessons bool) (int, error) {
	query := `
		INSERT INTO current_schedules
			(ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description)
		SELECT ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description
		FROM ` + sourceTable + `
		WHERE 1 = 1
	`
	args := make([]interface{}, 0)

	if classID > 0 {
		args = append(args, classID)
		query += fmt.Sprintf(` AND ClassID = $%d`, len(args))
	}
	if onlyValidLessons {
		query += ` AND TRIM(DayOfWeek) <> ''`
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}

func copyBaseLessonsToPlannedTx(tx *sql.Tx, classID int) (int, error) {
	query := `
		INSERT INTO planned_schedules
			(BaseScheduleID, ClassID, Date, DayOfWeek, LessonNumber, WeekType, Subject,
			 ChangeType, Scope, TeacherID, Room, StartTime, EndTime, Description, Reason, CreatedAt)
		SELECT Id, ClassID, '', DayOfWeek, LessonNumber, WeekType, Subject,
			'update', 'lesson', TeacherID, Room, StartTime, EndTime, Description, '', $1
		FROM schedules
		WHERE 1 = 1
	`
	args := []interface{}{time.Now().UTC().Format(time.RFC3339)}

	if classID > 0 {
		args = append(args, classID)
		query += fmt.Sprintf(` AND ClassID = $%d`, len(args))
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}

func nullableScheduleInt(value *int) interface{} {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func intPtrFromNull(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	intValue := int(value.Int64)
	return &intValue
}
