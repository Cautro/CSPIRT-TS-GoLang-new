package storage

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"cspirt/internal/logger"
	scheduleModels "cspirt/internal/schedule/models"
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

	base, err := s.getBaseSchedulesLocked(filter)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_schedules",
			Message: "failed to get base schedules: " + err.Error(),
		})
		return nil, err
	}

	exceptions := []scheduleModels.ScheduleException{}
	if filter.Date != "" {
		exceptions, err = s.getScheduleExceptionsLocked(filter)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "get_schedules",
				Message: "failed to get schedule exceptions: " + err.Error(),
			})
			return nil, err
		}
	}

	planned, err := s.getPlannedSchedulesLocked(filter)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_schedules",
			Message: "failed to get planned schedules: " + err.Error(),
		})
		return nil, err
	}

	return &scheduleModels.SchedulesResponse{
		Schedules:  buildScheduleViews(base, exceptions, planned, filter.Date),
		Base:       base,
		Exceptions: exceptions,
		Planned:    planned,
	}, nil
}

func (s *Storage) GetBaseScheduleByID(id int) (*scheduleModels.BaseSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return nil, errors.New("schedule id is required")
	}

	return s.getBaseScheduleByIDLocked(id)
}

func (s *Storage) UpsertBaseSchedule(schedule scheduleModels.BaseSchedule) (*scheduleModels.BaseSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalizeBaseScheduleStorage(&schedule)
	if err := s.validateBaseScheduleRefsLocked(schedule); err != nil {
		return nil, err
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "upsert_base_schedule",
		Message: "upserting base schedule",
	})

	if schedule.ID > 0 {
		result, err := s.db.Exec(`
			UPDATE schedules
			SET ClassID = ?, DayOfWeek = ?, LessonNumber = ?, WeekType = ?, Subject = ?,
				TeacherID = ?, Room = ?, StartTime = ?, EndTime = ?, Description = ?
			WHERE Id = ?
		`, schedule.ClassID, schedule.DayOfWeek, schedule.LessonNumber, schedule.WeekType,
			schedule.Subject, schedule.TeacherID, schedule.Room, schedule.StartTime,
			schedule.EndTime, schedule.Description, schedule.ID)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "upsert_base_schedule",
				Message: "failed to update base schedule: " + err.Error(),
			})
			return nil, err
		}
		if affected, err := result.RowsAffected(); err == nil && affected == 0 {
			return nil, errors.New("schedule not found")
		}

		return s.getBaseScheduleByIDLocked(schedule.ID)
	}

	_, err := s.db.Exec(`
		INSERT INTO schedules
			(ClassID, DayOfWeek, LessonNumber, WeekType, Subject, TeacherID, Room, StartTime, EndTime, Description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ClassID, DayOfWeek, LessonNumber, WeekType) DO UPDATE SET
			Subject = excluded.Subject,
			TeacherID = excluded.TeacherID,
			Room = excluded.Room,
			StartTime = excluded.StartTime,
			EndTime = excluded.EndTime,
			Description = excluded.Description
	`, schedule.ClassID, schedule.DayOfWeek, schedule.LessonNumber, schedule.WeekType,
		schedule.Subject, schedule.TeacherID, schedule.Room, schedule.StartTime,
		schedule.EndTime, schedule.Description)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "upsert_base_schedule",
			Message: "failed to insert base schedule: " + err.Error(),
		})
		return nil, err
	}

	return s.getBaseScheduleByUniqueKeyLocked(schedule.ClassID, schedule.DayOfWeek, schedule.LessonNumber, schedule.WeekType)
}

func (s *Storage) DeleteBaseSchedule(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return errors.New("schedule id is required")
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_base_schedule",
		Message: "deleting base schedule",
	})

	result, err := s.db.Exec(`DELETE FROM schedules WHERE Id = ?`, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_base_schedule",
			Message: "failed to delete base schedule: " + err.Error(),
		})
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("schedule not found")
	}

	return nil
}

func (s *Storage) GetScheduleExceptionByID(id int) (*scheduleModels.ScheduleException, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return nil, errors.New("exception id is required")
	}

	return s.getScheduleExceptionByIDLocked(id)
}

func (s *Storage) UpsertScheduleException(exception scheduleModels.ScheduleException) (*scheduleModels.ScheduleException, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalizeScheduleExceptionStorage(&exception)
	if err := s.validateScheduleExceptionRefsLocked(&exception); err != nil {
		return nil, err
	}
	if exception.CreatedAt == "" {
		exception.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "upsert_schedule_exception",
		Message: "upserting schedule exception",
	})

	if exception.ID > 0 {
		result, err := s.db.Exec(`
			UPDATE exceptions_schedules
			SET ScheduleID = ?, ClassID = ?, Date = ?, ChangeType = ?, Scope = ?, NewSubject = ?,
				NewLessonNumber = ?, NewTeacherID = ?, NewRoom = ?, NewStartTime = ?,
				NewEndTime = ?, NewDescription = ?, Reason = ?, CreatedAt = ?
			WHERE Id = ?
		`, nullableInt(exception.ScheduleID), exception.ClassID, exception.Date, string(exception.ChangeType),
			exception.Scope, nullableStringPtr(exception.NewSubject), nullableInt(exception.NewLessonNumber),
			nullableInt(exception.NewTeacherID), nullableInt(exception.NewRoom),
			nullableStringPtr(exception.NewStartTime), nullableStringPtr(exception.NewEndTime),
			nullableStringPtr(exception.NewDescription), exception.Reason, exception.CreatedAt, exception.ID)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "upsert_schedule_exception",
				Message: "failed to update schedule exception: " + err.Error(),
			})
			return nil, err
		}
		if affected, err := result.RowsAffected(); err == nil && affected == 0 {
			return nil, errors.New("exception not found")
		}

		return s.getScheduleExceptionByIDLocked(exception.ID)
	}

	result, err := s.db.Exec(`
		INSERT INTO exceptions_schedules
			(ScheduleID, ClassID, Date, ChangeType, Scope, NewSubject, NewLessonNumber,
			 NewTeacherID, NewRoom, NewStartTime, NewEndTime, NewDescription, Reason, CreatedAt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, nullableInt(exception.ScheduleID), exception.ClassID, exception.Date, string(exception.ChangeType),
		exception.Scope, nullableStringPtr(exception.NewSubject), nullableInt(exception.NewLessonNumber),
		nullableInt(exception.NewTeacherID), nullableInt(exception.NewRoom),
		nullableStringPtr(exception.NewStartTime), nullableStringPtr(exception.NewEndTime),
		nullableStringPtr(exception.NewDescription), exception.Reason, exception.CreatedAt)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "upsert_schedule_exception",
			Message: "failed to insert schedule exception: " + err.Error(),
		})
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.getScheduleExceptionByIDLocked(int(id))
}

func (s *Storage) DeleteScheduleException(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return errors.New("exception id is required")
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_schedule_exception",
		Message: "deleting schedule exception",
	})

	result, err := s.db.Exec(`DELETE FROM exceptions_schedules WHERE Id = ?`, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_schedule_exception",
			Message: "failed to delete schedule exception: " + err.Error(),
		})
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("exception not found")
	}

	return nil
}

func (s *Storage) GetPlannedScheduleByID(id int) (*scheduleModels.PlannedSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return nil, errors.New("planned schedule id is required")
	}

	return s.getPlannedScheduleByIDLocked(id)
}

func (s *Storage) UpsertPlannedSchedule(planned scheduleModels.PlannedSchedule) (*scheduleModels.PlannedSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalizePlannedScheduleStorage(&planned)
	if err := s.validatePlannedScheduleRefsLocked(&planned); err != nil {
		return nil, err
	}
	if planned.CreatedAt == "" {
		planned.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "upsert_planned_schedule",
		Message: "upserting planned schedule",
	})

	if planned.ID > 0 {
		result, err := s.db.Exec(`
			UPDATE planned_schedules
			SET BaseScheduleID = ?, ClassID = ?, Date = ?, LessonNumber = ?, Subject = ?,
				ChangeType = ?, Scope = ?, TeacherID = ?, Room = ?, StartTime = ?, EndTime = ?,
				Description = ?, Reason = ?, CreatedAt = ?
			WHERE Id = ?
		`, nullableInt(planned.BaseScheduleID), planned.ClassID, planned.Date, planned.LessonNumber,
			planned.Subject, string(planned.ChangeType), planned.Scope, planned.TeacherID,
			planned.Room, planned.StartTime, planned.EndTime, planned.Description,
			planned.Reason, planned.CreatedAt, planned.ID)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "upsert_planned_schedule",
				Message: "failed to update planned schedule: " + err.Error(),
			})
			return nil, err
		}
		if affected, err := result.RowsAffected(); err == nil && affected == 0 {
			return nil, errors.New("planned schedule not found")
		}

		return s.getPlannedScheduleByIDLocked(planned.ID)
	}

	result, err := s.db.Exec(`
		INSERT INTO planned_schedules
			(BaseScheduleID, ClassID, Date, LessonNumber, Subject, ChangeType, Scope,
			 TeacherID, Room, StartTime, EndTime, Description, Reason, CreatedAt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, nullableInt(planned.BaseScheduleID), planned.ClassID, planned.Date, planned.LessonNumber,
		planned.Subject, string(planned.ChangeType), planned.Scope, planned.TeacherID,
		planned.Room, planned.StartTime, planned.EndTime, planned.Description,
		planned.Reason, planned.CreatedAt)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "upsert_planned_schedule",
			Message: "failed to insert planned schedule: " + err.Error(),
		})
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.getPlannedScheduleByIDLocked(int(id))
}

func (s *Storage) DeletePlannedSchedule(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id <= 0 {
		return errors.New("planned schedule id is required")
	}

	logger.WriteSafe(logger.LogEntry{
		Level:   "info",
		Action:  "delete_planned_schedule",
		Message: "deleting planned schedule",
	})

	result, err := s.db.Exec(`DELETE FROM planned_schedules WHERE Id = ?`, id)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_planned_schedule",
			Message: "failed to delete planned schedule: " + err.Error(),
		})
		return err
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("planned schedule not found")
	}

	return nil
}

func (s *Storage) getBaseSchedulesLocked(filter scheduleModels.ScheduleFilter) ([]scheduleModels.BaseSchedule, error) {
	query := `
		SELECT s.Id, s.ClassID, c.Name, s.DayOfWeek, s.LessonNumber, s.WeekType,
			s.Subject, s.TeacherID, s.Room, s.StartTime, s.EndTime, s.Description
		FROM schedules s
		JOIN classes c ON c.Id = s.ClassID
		WHERE 1 = 1
	`
	args := make([]interface{}, 0)

	if filter.ClassID > 0 {
		query += ` AND s.ClassID = ?`
		args = append(args, filter.ClassID)
	}
	if filter.Day != "" {
		query += ` AND LOWER(s.DayOfWeek) = LOWER(?)`
		args = append(args, filter.Day)
	}
	if filter.WeekType != "" {
		if filter.WeekType == "all" {
			query += ` AND LOWER(s.WeekType) = 'all'`
		} else {
			query += ` AND LOWER(s.WeekType) IN ('all', LOWER(?))`
			args = append(args, filter.WeekType)
		}
	}

	query += `
		ORDER BY s.ClassID,
			CASE LOWER(s.DayOfWeek)
				WHEN 'monday' THEN 1 WHEN 'понедельник' THEN 1
				WHEN 'tuesday' THEN 2 WHEN 'вторник' THEN 2
				WHEN 'wednesday' THEN 3 WHEN 'среда' THEN 3
				WHEN 'thursday' THEN 4 WHEN 'четверг' THEN 4
				WHEN 'friday' THEN 5 WHEN 'пятница' THEN 5
				WHEN 'saturday' THEN 6 WHEN 'суббота' THEN 6
				WHEN 'sunday' THEN 7 WHEN 'воскресенье' THEN 7
				ELSE 8
			END,
			s.LessonNumber,
			s.Id
	`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBaseSchedules(rows)
}

func (s *Storage) getBaseScheduleByIDLocked(id int) (*scheduleModels.BaseSchedule, error) {
	row := s.db.QueryRow(`
		SELECT s.Id, s.ClassID, c.Name, s.DayOfWeek, s.LessonNumber, s.WeekType,
			s.Subject, s.TeacherID, s.Room, s.StartTime, s.EndTime, s.Description
		FROM schedules s
		JOIN classes c ON c.Id = s.ClassID
		WHERE s.Id = ?
	`, id)

	schedule, err := scanBaseSchedule(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &schedule, nil
}

func (s *Storage) getBaseScheduleByUniqueKeyLocked(classID int, day string, lessonNumber int, weekType string) (*scheduleModels.BaseSchedule, error) {
	row := s.db.QueryRow(`
		SELECT s.Id, s.ClassID, c.Name, s.DayOfWeek, s.LessonNumber, s.WeekType,
			s.Subject, s.TeacherID, s.Room, s.StartTime, s.EndTime, s.Description
		FROM schedules s
		JOIN classes c ON c.Id = s.ClassID
		WHERE s.ClassID = ?
			AND LOWER(s.DayOfWeek) = LOWER(?)
			AND s.LessonNumber = ?
			AND LOWER(s.WeekType) = LOWER(?)
	`, classID, day, lessonNumber, weekType)

	schedule, err := scanBaseSchedule(row)
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (s *Storage) getScheduleExceptionsLocked(filter scheduleModels.ScheduleFilter) ([]scheduleModels.ScheduleException, error) {
	query := `
		SELECT Id, ScheduleID, ClassID, Date, ChangeType, Scope, NewSubject,
			NewLessonNumber, NewTeacherID, NewRoom, NewStartTime, NewEndTime,
			NewDescription, Reason, CreatedAt
		FROM exceptions_schedules
		WHERE 1 = 1
	`
	args := make([]interface{}, 0)

	if filter.ClassID > 0 {
		query += ` AND ClassID = ?`
		args = append(args, filter.ClassID)
	}
	if filter.Date != "" {
		query += ` AND Date = ?`
		args = append(args, filter.Date)
	}

	query += ` ORDER BY Date, ClassID, Id`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanScheduleExceptions(rows)
}

func (s *Storage) getScheduleExceptionByIDLocked(id int) (*scheduleModels.ScheduleException, error) {
	row := s.db.QueryRow(`
		SELECT Id, ScheduleID, ClassID, Date, ChangeType, Scope, NewSubject,
			NewLessonNumber, NewTeacherID, NewRoom, NewStartTime, NewEndTime,
			NewDescription, Reason, CreatedAt
		FROM exceptions_schedules
		WHERE Id = ?
	`, id)

	exception, err := scanScheduleException(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &exception, nil
}

func (s *Storage) getPlannedSchedulesLocked(filter scheduleModels.ScheduleFilter) ([]scheduleModels.PlannedSchedule, error) {
	query := `
		SELECT Id, BaseScheduleID, ClassID, Date, LessonNumber, Subject, ChangeType,
			Scope, TeacherID, Room, StartTime, EndTime, Description, Reason, CreatedAt
		FROM planned_schedules
		WHERE 1 = 1
	`
	args := make([]interface{}, 0)

	if filter.ClassID > 0 {
		query += ` AND ClassID = ?`
		args = append(args, filter.ClassID)
	}
	if filter.Date != "" {
		query += ` AND Date = ?`
		args = append(args, filter.Date)
	}

	query += ` ORDER BY Date, ClassID, LessonNumber, Id`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPlannedSchedules(rows)
}

func (s *Storage) getPlannedScheduleByIDLocked(id int) (*scheduleModels.PlannedSchedule, error) {
	row := s.db.QueryRow(`
		SELECT Id, BaseScheduleID, ClassID, Date, LessonNumber, Subject, ChangeType,
			Scope, TeacherID, Room, StartTime, EndTime, Description, Reason, CreatedAt
		FROM planned_schedules
		WHERE Id = ?
	`, id)

	planned, err := scanPlannedSchedule(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &planned, nil
}

func (s *Storage) validateBaseScheduleRefsLocked(schedule scheduleModels.BaseSchedule) error {
	if schedule.ClassID <= 0 {
		return errors.New("class id is required")
	}
	class, err := s.getClassByIDLocked(schedule.ClassID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}

	return s.validateScheduleTeacherLocked(schedule.TeacherID)
}

func (s *Storage) validateScheduleExceptionRefsLocked(exception *scheduleModels.ScheduleException) error {
	if exception.ScheduleID != nil {
		base, err := s.getBaseScheduleByIDLocked(*exception.ScheduleID)
		if err != nil {
			return err
		}
		if base == nil {
			return errors.New("base schedule not found")
		}
		if exception.ClassID <= 0 {
			exception.ClassID = base.ClassID
		}
		if exception.ClassID != base.ClassID {
			return errors.New("exception class does not match base schedule class")
		}
	}

	if exception.ClassID <= 0 {
		return errors.New("class id is required")
	}
	class, err := s.getClassByIDLocked(exception.ClassID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}
	if exception.NewTeacherID != nil {
		if err := s.validateScheduleTeacherLocked(*exception.NewTeacherID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) validatePlannedScheduleRefsLocked(planned *scheduleModels.PlannedSchedule) error {
	if planned.BaseScheduleID != nil {
		base, err := s.getBaseScheduleByIDLocked(*planned.BaseScheduleID)
		if err != nil {
			return err
		}
		if base == nil {
			return errors.New("base schedule not found")
		}
		if planned.ClassID <= 0 {
			planned.ClassID = base.ClassID
		}
		if planned.ClassID != base.ClassID {
			return errors.New("planned schedule class does not match base schedule class")
		}
		if planned.LessonNumber <= 0 {
			planned.LessonNumber = base.LessonNumber
		}
		if planned.Subject == "" {
			planned.Subject = base.Subject
		}
		if planned.TeacherID <= 0 {
			planned.TeacherID = base.TeacherID
		}
		if planned.Room <= 0 {
			planned.Room = base.Room
		}
		if planned.StartTime == "" {
			planned.StartTime = base.StartTime
		}
		if planned.EndTime == "" {
			planned.EndTime = base.EndTime
		}
		if planned.Description == "" {
			planned.Description = base.Description
		}
	}

	if planned.ClassID <= 0 {
		return errors.New("class id is required")
	}
	class, err := s.getClassByIDLocked(planned.ClassID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}

	return s.validateScheduleTeacherLocked(planned.TeacherID)
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

type scheduleScanner interface {
	Scan(dest ...interface{}) error
}

func scanBaseSchedules(rows *sql.Rows) ([]scheduleModels.BaseSchedule, error) {
	schedules := make([]scheduleModels.BaseSchedule, 0)

	for rows.Next() {
		schedule, err := scanBaseSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

func scanBaseSchedule(scanner scheduleScanner) (scheduleModels.BaseSchedule, error) {
	var schedule scheduleModels.BaseSchedule
	if err := scanner.Scan(
		&schedule.ID,
		&schedule.ClassID,
		&schedule.Class,
		&schedule.DayOfWeek,
		&schedule.LessonNumber,
		&schedule.WeekType,
		&schedule.Subject,
		&schedule.TeacherID,
		&schedule.Room,
		&schedule.StartTime,
		&schedule.EndTime,
		&schedule.Description,
	); err != nil {
		return scheduleModels.BaseSchedule{}, err
	}

	return schedule, nil
}

func scanScheduleExceptions(rows *sql.Rows) ([]scheduleModels.ScheduleException, error) {
	exceptions := make([]scheduleModels.ScheduleException, 0)

	for rows.Next() {
		exception, err := scanScheduleException(rows)
		if err != nil {
			return nil, err
		}
		exceptions = append(exceptions, exception)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return exceptions, nil
}

func scanScheduleException(scanner scheduleScanner) (scheduleModels.ScheduleException, error) {
	var exception scheduleModels.ScheduleException
	var scheduleID sql.NullInt64
	var newSubject sql.NullString
	var newLessonNumber sql.NullInt64
	var newTeacherID sql.NullInt64
	var newRoom sql.NullInt64
	var newStartTime sql.NullString
	var newEndTime sql.NullString
	var newDescription sql.NullString

	if err := scanner.Scan(
		&exception.ID,
		&scheduleID,
		&exception.ClassID,
		&exception.Date,
		&exception.ChangeType,
		&exception.Scope,
		&newSubject,
		&newLessonNumber,
		&newTeacherID,
		&newRoom,
		&newStartTime,
		&newEndTime,
		&newDescription,
		&exception.Reason,
		&exception.CreatedAt,
	); err != nil {
		return scheduleModels.ScheduleException{}, err
	}

	exception.ScheduleID = intPtrFromNull(scheduleID)
	exception.NewSubject = stringPtrFromNull(newSubject)
	exception.NewLessonNumber = intPtrFromNull(newLessonNumber)
	exception.NewTeacherID = intPtrFromNull(newTeacherID)
	exception.NewRoom = intPtrFromNull(newRoom)
	exception.NewStartTime = stringPtrFromNull(newStartTime)
	exception.NewEndTime = stringPtrFromNull(newEndTime)
	exception.NewDescription = stringPtrFromNull(newDescription)

	return exception, nil
}

func scanPlannedSchedules(rows *sql.Rows) ([]scheduleModels.PlannedSchedule, error) {
	planned := make([]scheduleModels.PlannedSchedule, 0)

	for rows.Next() {
		item, err := scanPlannedSchedule(rows)
		if err != nil {
			return nil, err
		}
		planned = append(planned, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return planned, nil
}

func scanPlannedSchedule(scanner scheduleScanner) (scheduleModels.PlannedSchedule, error) {
	var planned scheduleModels.PlannedSchedule
	var baseScheduleID sql.NullInt64

	if err := scanner.Scan(
		&planned.ID,
		&baseScheduleID,
		&planned.ClassID,
		&planned.Date,
		&planned.LessonNumber,
		&planned.Subject,
		&planned.ChangeType,
		&planned.Scope,
		&planned.TeacherID,
		&planned.Room,
		&planned.StartTime,
		&planned.EndTime,
		&planned.Description,
		&planned.Reason,
		&planned.CreatedAt,
	); err != nil {
		return scheduleModels.PlannedSchedule{}, err
	}

	planned.BaseScheduleID = intPtrFromNull(baseScheduleID)
	return planned, nil
}

func buildScheduleViews(
	base []scheduleModels.BaseSchedule,
	exceptions []scheduleModels.ScheduleException,
	planned []scheduleModels.PlannedSchedule,
	date string,
) []scheduleModels.ScheduleView {
	views := make([]scheduleModels.ScheduleView, 0, len(base)+len(planned)+len(exceptions))
	indexByBaseID := make(map[int]int, len(base))

	for _, schedule := range base {
		view := baseScheduleToView(schedule, date)
		indexByBaseID[schedule.ID] = len(views)
		views = append(views, view)
	}
	if date == "" {
		return views
	}

	for _, item := range planned {
		if item.Date != date {
			continue
		}
		applyPlannedSchedule(&views, indexByBaseID, item)
	}

	for _, exception := range exceptions {
		if exception.Date != date {
			continue
		}
		applyScheduleException(&views, indexByBaseID, exception)
	}

	return views
}

func baseScheduleToView(schedule scheduleModels.BaseSchedule, date string) scheduleModels.ScheduleView {
	baseID := schedule.ID
	return scheduleModels.ScheduleView{
		ID:             schedule.ID,
		BaseScheduleID: &baseID,
		Source:         scheduleModels.ScheduleSourceBase,
		ClassID:        schedule.ClassID,
		Class:          schedule.Class,
		DayOfWeek:      schedule.DayOfWeek,
		Date:           date,
		LessonNumber:   schedule.LessonNumber,
		WeekType:       schedule.WeekType,
		Subject:        schedule.Subject,
		TeacherID:      schedule.TeacherID,
		Teacher:        schedule.Teacher,
		Room:           schedule.Room,
		StartTime:      schedule.StartTime,
		EndTime:        schedule.EndTime,
		Description:    schedule.Description,
	}
}

func applyPlannedSchedule(views *[]scheduleModels.ScheduleView, indexByBaseID map[int]int, planned scheduleModels.PlannedSchedule) {
	changeID := planned.ID
	if planned.BaseScheduleID != nil {
		if index, ok := indexByBaseID[*planned.BaseScheduleID]; ok {
			view := &(*views)[index]
			view.ChangeID = &changeID
			view.Source = scheduleModels.ScheduleSourcePlanned
			view.ChangeType = planned.ChangeType
			view.Date = planned.Date
			view.LessonNumber = planned.LessonNumber
			view.Subject = planned.Subject
			view.TeacherID = planned.TeacherID
			view.Teacher = nil
			view.Room = planned.Room
			view.StartTime = planned.StartTime
			view.EndTime = planned.EndTime
			view.Description = planned.Description
			view.Reason = planned.Reason
			view.IsCancelled = isCancelLikeChange(planned.ChangeType)
			return
		}
	}

	*views = append(*views, scheduleModels.ScheduleView{
		ID:           planned.ID,
		ChangeID:     &changeID,
		Source:       scheduleModels.ScheduleSourcePlanned,
		ChangeType:   planned.ChangeType,
		IsCancelled:  isCancelLikeChange(planned.ChangeType),
		ClassID:      planned.ClassID,
		Date:         planned.Date,
		LessonNumber: planned.LessonNumber,
		Subject:      planned.Subject,
		TeacherID:    planned.TeacherID,
		Room:         planned.Room,
		StartTime:    planned.StartTime,
		EndTime:      planned.EndTime,
		Description:  planned.Description,
		Reason:       planned.Reason,
	})
}

func applyScheduleException(views *[]scheduleModels.ScheduleView, indexByBaseID map[int]int, exception scheduleModels.ScheduleException) {
	changeID := exception.ID

	if exception.ChangeType == scheduleModels.ChangeDayOff && strings.EqualFold(exception.Scope, "class") {
		for index := range *views {
			if (*views)[index].ClassID == exception.ClassID {
				(*views)[index].ChangeID = &changeID
				(*views)[index].Source = scheduleModels.ScheduleSourceException
				(*views)[index].ChangeType = exception.ChangeType
				(*views)[index].Reason = exception.Reason
				(*views)[index].IsCancelled = true
			}
		}
		return
	}

	if exception.ScheduleID != nil {
		if index, ok := indexByBaseID[*exception.ScheduleID]; ok {
			applyExceptionToView(&(*views)[index], exception)
			return
		}
	}

	if exception.NewLessonNumber != nil {
		for index := range *views {
			if (*views)[index].ClassID == exception.ClassID && (*views)[index].LessonNumber == *exception.NewLessonNumber {
				applyExceptionToView(&(*views)[index], exception)
				return
			}
		}
	}

	if exception.ChangeType == scheduleModels.ChangeAdd {
		*views = append(*views, exceptionToAddedView(exception))
	}
}

func applyExceptionToView(view *scheduleModels.ScheduleView, exception scheduleModels.ScheduleException) {
	changeID := exception.ID
	view.ChangeID = &changeID
	view.Source = scheduleModels.ScheduleSourceException
	view.ChangeType = exception.ChangeType
	view.Date = exception.Date
	view.Reason = exception.Reason
	view.IsCancelled = isCancelLikeChange(exception.ChangeType)

	if exception.NewLessonNumber != nil {
		view.LessonNumber = *exception.NewLessonNumber
	}
	if exception.NewSubject != nil {
		view.Subject = *exception.NewSubject
	}
	if exception.NewTeacherID != nil {
		view.TeacherID = *exception.NewTeacherID
		view.Teacher = nil
	}
	if exception.NewRoom != nil {
		view.Room = *exception.NewRoom
	}
	if exception.NewStartTime != nil {
		view.StartTime = *exception.NewStartTime
	}
	if exception.NewEndTime != nil {
		view.EndTime = *exception.NewEndTime
	}
	if exception.NewDescription != nil {
		view.Description = *exception.NewDescription
	}
}

func exceptionToAddedView(exception scheduleModels.ScheduleException) scheduleModels.ScheduleView {
	changeID := exception.ID
	view := scheduleModels.ScheduleView{
		ID:          exception.ID,
		ChangeID:    &changeID,
		Source:      scheduleModels.ScheduleSourceException,
		ChangeType:  exception.ChangeType,
		Date:        exception.Date,
		ClassID:     exception.ClassID,
		Reason:      exception.Reason,
		IsCancelled: isCancelLikeChange(exception.ChangeType),
	}

	if exception.NewLessonNumber != nil {
		view.LessonNumber = *exception.NewLessonNumber
	}
	if exception.NewSubject != nil {
		view.Subject = *exception.NewSubject
	}
	if exception.NewTeacherID != nil {
		view.TeacherID = *exception.NewTeacherID
	}
	if exception.NewRoom != nil {
		view.Room = *exception.NewRoom
	}
	if exception.NewStartTime != nil {
		view.StartTime = *exception.NewStartTime
	}
	if exception.NewEndTime != nil {
		view.EndTime = *exception.NewEndTime
	}
	if exception.NewDescription != nil {
		view.Description = *exception.NewDescription
	}

	return view
}

func normalizeScheduleFilter(filter scheduleModels.ScheduleFilter) scheduleModels.ScheduleFilter {
	filter.Day = strings.TrimSpace(filter.Day)
	filter.Date = strings.TrimSpace(filter.Date)
	filter.WeekType = strings.ToLower(strings.TrimSpace(filter.WeekType))
	if filter.WeekType == "" {
		filter.WeekType = "all"
	}
	return filter
}

func normalizeBaseScheduleStorage(schedule *scheduleModels.BaseSchedule) {
	schedule.DayOfWeek = strings.TrimSpace(schedule.DayOfWeek)
	schedule.WeekType = strings.ToLower(strings.TrimSpace(schedule.WeekType))
	if schedule.WeekType == "" {
		schedule.WeekType = "all"
	}
	schedule.Subject = strings.TrimSpace(schedule.Subject)
	schedule.StartTime = strings.TrimSpace(schedule.StartTime)
	schedule.EndTime = strings.TrimSpace(schedule.EndTime)
	schedule.Description = strings.TrimSpace(schedule.Description)
}

func normalizeScheduleExceptionStorage(exception *scheduleModels.ScheduleException) {
	exception.Date = strings.TrimSpace(exception.Date)
	exception.ChangeType = scheduleModels.ChangeType(strings.ToLower(strings.TrimSpace(string(exception.ChangeType))))
	exception.Scope = strings.TrimSpace(exception.Scope)
	if exception.Scope == "" {
		exception.Scope = "lesson"
	}
	exception.Reason = strings.TrimSpace(exception.Reason)
}

func normalizePlannedScheduleStorage(planned *scheduleModels.PlannedSchedule) {
	planned.Date = strings.TrimSpace(planned.Date)
	planned.ChangeType = scheduleModels.ChangeType(strings.ToLower(strings.TrimSpace(string(planned.ChangeType))))
	planned.Scope = strings.TrimSpace(planned.Scope)
	if planned.Scope == "" {
		planned.Scope = "lesson"
	}
	planned.Subject = strings.TrimSpace(planned.Subject)
	planned.StartTime = strings.TrimSpace(planned.StartTime)
	planned.EndTime = strings.TrimSpace(planned.EndTime)
	planned.Description = strings.TrimSpace(planned.Description)
	planned.Reason = strings.TrimSpace(planned.Reason)
}

func isCancelLikeChange(changeType scheduleModels.ChangeType) bool {
	return changeType == scheduleModels.ChangeCancel ||
		changeType == scheduleModels.ChangeDayOff ||
		changeType == scheduleModels.ChangeShortDay
}

func intPtrFromNull(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	intValue := int(value.Int64)
	return &intValue
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	stringValue := value.String
	return &stringValue
}

func nullableInt(value *int) interface{} {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func nullableStringPtr(value *string) interface{} {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
