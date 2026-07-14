package schedule

import (
	"errors"
	"strings"

	"cspirt/pkg/logger"
	scheduleModels "cspirt/internal/domain/schedule"
	repo "cspirt/internal/domain/schedule/repo"
)

type ScheduleUsecase struct {
	schedules repo.ScheduleRepository
}

func NewScheduleUsecase(schedules repo.ScheduleRepository) *ScheduleUsecase {
	return &ScheduleUsecase{
		schedules: schedules,
	}
}

func (s *ScheduleUsecase) GetSchedules(filter scheduleModels.ScheduleFilter) (*scheduleModels.SchedulesResponse, error) {
	filter = normalizeScheduleFilter(filter)

	result, err := s.schedules.GetSchedules(filter)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_schedules",
			Message: "failed to get schedules: " + err.Error(),
		})
		return nil, err
	}
	if result == nil {
		return &scheduleModels.SchedulesResponse{
			Schedules: []scheduleModels.ScheduleLesson{},
			Base:      []scheduleModels.ScheduleLesson{},
			Current:   []scheduleModels.ScheduleLesson{},
			Planned:   []scheduleModels.ScheduleLesson{},
		}, nil
	}

	return result, nil
}

func (s *ScheduleUsecase) GetCurrentScheduleForTeacher(teacherID int, filter scheduleModels.ScheduleFilter) ([]scheduleModels.ScheduleLesson, error) {
	if teacherID <= 0 {
		return nil, errors.New("teacher id is required")
	}

	filter = normalizeScheduleFilter(filter)
	filter.Type = scheduleModels.ScheduleTypeCurrent
	filter.TeacherID = teacherID

	lessons, err := s.schedules.GetCurrentScheduleForTeacher(teacherID, filter)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "get_teacher_current_schedule",
			Message: "failed to get teacher current schedule: " + err.Error(),
		})
		return nil, err
	}
	if lessons == nil {
		return []scheduleModels.ScheduleLesson{}, nil
	}

	return lessons, nil
}

func (s *ScheduleUsecase) UpdateSchedules(input scheduleModels.UpdateSchedulesInput) (*scheduleModels.UpdateSchedulesResult, error) {
	scheduleType, err := scheduleTypeFromInput(input)
	if err != nil {
		return nil, err
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	if action == "" {
		action = scheduleModels.ScheduleActionUpsert
	}

	switch action {
	case scheduleModels.ScheduleActionDelete:
		return s.deleteScheduleLesson(scheduleType, input)
	case scheduleModels.ScheduleActionUpsert:
		return s.upsertScheduleLesson(scheduleType, input)
	default:
		return nil, errors.New("invalid schedule action")
	}
}

func (s *ScheduleUsecase) RolloverSchedules(classID int) (*scheduleModels.ScheduleRolloverResult, error) {
	return s.schedules.RolloverSchedules(classID)
}

func (s *ScheduleUsecase) ResetPlannedSchedules(classID int) (*scheduleModels.ScheduleResetResult, error) {
	return s.schedules.ResetPlannedSchedules(classID)
}

func (s *ScheduleUsecase) deleteScheduleLesson(scheduleType string, input scheduleModels.UpdateSchedulesInput) (*scheduleModels.UpdateSchedulesResult, error) {
	id := input.ID
	if id <= 0 {
		if lesson := schedulePayload(input); lesson != nil {
			id = lesson.ID
		}
	}
	if id <= 0 {
		return nil, errors.New("schedule lesson id is required")
	}

	if err := s.schedules.DeleteScheduleLesson(scheduleType, id); err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "delete_schedule_lesson",
			Message: "failed to delete schedule lesson: " + err.Error(),
		})
		return nil, err
	}

	return &scheduleModels.UpdateSchedulesResult{Target: scheduleType, Action: scheduleModels.ScheduleActionDelete}, nil
}

func (s *ScheduleUsecase) upsertScheduleLesson(scheduleType string, input scheduleModels.UpdateSchedulesInput) (*scheduleModels.UpdateSchedulesResult, error) {
	payload := schedulePayload(input)
	if payload == nil {
		return nil, errors.New("schedule lesson payload is required")
	}

	lesson := *payload
	if lesson.ID <= 0 && input.ID > 0 {
		lesson.ID = input.ID
	}

	if lesson.ID > 0 {
		existing, err := s.schedules.GetScheduleLessonByID(scheduleType, lesson.ID)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			return nil, errors.New("schedule lesson not found")
		}
		lesson = mergeScheduleLesson(*existing, lesson)
	}

	lesson.Type = scheduleType
	normalizeScheduleLesson(&lesson)
	if err := validateScheduleLesson(scheduleType, lesson); err != nil {
		return nil, err
	}

	result, err := s.schedules.UpsertScheduleLesson(scheduleType, lesson)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "upsert_schedule_lesson",
			Message: "failed to upsert schedule lesson: " + err.Error(),
		})
		return nil, err
	}

	return &scheduleModels.UpdateSchedulesResult{
		Target:   scheduleType,
		Action:   scheduleModels.ScheduleActionUpsert,
		Lesson:   result,
		Schedule: result,
	}, nil
}

func schedulePayload(input scheduleModels.UpdateSchedulesInput) *scheduleModels.ScheduleLesson {
	if input.Lesson != nil {
		return input.Lesson
	}
	return input.Schedule
}

func scheduleTypeFromInput(input scheduleModels.UpdateSchedulesInput) (string, error) {
	raw := strings.TrimSpace(input.Type)
	if raw == "" {
		raw = strings.TrimSpace(input.Target)
	}
	if raw == "" {
		if lesson := schedulePayload(input); lesson != nil {
			raw = lesson.Type
		}
	}
	if raw == "" {
		raw = scheduleModels.ScheduleTypeCurrent
	}

	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case scheduleModels.ScheduleTypeBase:
		return scheduleModels.ScheduleTypeBase, nil
	case scheduleModels.ScheduleTypeCurrent:
		return scheduleModels.ScheduleTypeCurrent, nil
	case scheduleModels.ScheduleTypePlanned:
		return scheduleModels.ScheduleTypePlanned, nil
	default:
		return "", errors.New("invalid schedule type")
	}
}

func mergeScheduleLesson(existing scheduleModels.ScheduleLesson, patch scheduleModels.ScheduleLesson) scheduleModels.ScheduleLesson {
	if patch.BaseScheduleID != nil {
		existing.BaseScheduleID = patch.BaseScheduleID
	}
	if patch.ClassID > 0 {
		existing.ClassID = patch.ClassID
	}
	if strings.TrimSpace(patch.DayOfWeek) != "" {
		existing.DayOfWeek = patch.DayOfWeek
	}
	if patch.LessonNumber > 0 {
		existing.LessonNumber = patch.LessonNumber
	}
	if strings.TrimSpace(patch.WeekType) != "" {
		existing.WeekType = patch.WeekType
	}
	if strings.TrimSpace(patch.Subject) != "" {
		existing.Subject = patch.Subject
	}
	if patch.TeacherID > 0 {
		existing.TeacherID = patch.TeacherID
	}
	if patch.Room > 0 {
		existing.Room = patch.Room
	}
	if strings.TrimSpace(patch.StartTime) != "" {
		existing.StartTime = patch.StartTime
	}
	if strings.TrimSpace(patch.EndTime) != "" {
		existing.EndTime = patch.EndTime
	}
	if strings.TrimSpace(patch.Description) != "" {
		existing.Description = patch.Description
	}

	return existing
}

func normalizeScheduleFilter(filter scheduleModels.ScheduleFilter) scheduleModels.ScheduleFilter {
	filter.Type = strings.ToLower(strings.TrimSpace(filter.Type))
	if filter.Type == "" {
		filter.Type = scheduleModels.ScheduleTypeCurrent
	}
	filter.Day = strings.TrimSpace(filter.Day)
	filter.WeekType = normalizeWeekType(filter.WeekType)
	return filter
}

func normalizeScheduleLesson(lesson *scheduleModels.ScheduleLesson) {
	lesson.Type = strings.ToLower(strings.TrimSpace(lesson.Type))
	lesson.DayOfWeek = strings.TrimSpace(lesson.DayOfWeek)
	lesson.WeekType = normalizeWeekType(lesson.WeekType)
	lesson.Subject = strings.TrimSpace(lesson.Subject)
	lesson.StartTime = strings.TrimSpace(lesson.StartTime)
	lesson.EndTime = strings.TrimSpace(lesson.EndTime)
	lesson.Description = strings.TrimSpace(lesson.Description)
}

func validateScheduleLesson(scheduleType string, lesson scheduleModels.ScheduleLesson) error {
	if scheduleType == scheduleModels.ScheduleTypePlanned && lesson.BaseScheduleID != nil && lesson.ID <= 0 {
		if *lesson.BaseScheduleID <= 0 {
			return errors.New("base schedule id is required")
		}
		return nil
	}

	if lesson.ClassID <= 0 {
		return errors.New("class id is required")
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
	if lesson.TeacherID <= 0 {
		return errors.New("teacher id is required")
	}
	if lesson.Room <= 0 {
		return errors.New("room is required")
	}
	if lesson.StartTime == "" || lesson.EndTime == "" {
		return errors.New("start time and end time are required")
	}

	return nil
}

func normalizeWeekType(weekType string) string {
	weekType = strings.ToLower(strings.TrimSpace(weekType))
	if weekType == "" {
		return "all"
	}
	return weekType
}
