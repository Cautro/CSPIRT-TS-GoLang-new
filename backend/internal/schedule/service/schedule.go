package service

import (
	"errors"
	"strings"

	"cspirt/internal/logger"
	scheduleModels "cspirt/internal/schedule/models"
	"cspirt/internal/schedule/repo"
)

type ScheduleService struct {
	schedules repo.ScheduleRepository
}

func NewScheduleService(schedules repo.ScheduleRepository, jwtSecret string) *ScheduleService {
	return &ScheduleService{
		schedules: schedules,
	}
}

func (s *ScheduleService) GetSchedules(filter scheduleModels.ScheduleFilter) (*scheduleModels.SchedulesResponse, error) {
	filter.Day = strings.TrimSpace(filter.Day)
	filter.Date = strings.TrimSpace(filter.Date)
	filter.WeekType = normalizeWeekType(filter.WeekType)

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
			Schedules:  []scheduleModels.ScheduleView{},
			Base:       []scheduleModels.BaseSchedule{},
			Exceptions: []scheduleModels.ScheduleException{},
			Planned:    []scheduleModels.PlannedSchedule{},
		}, nil
	}

	return result, nil
}

func (s *ScheduleService) UpdateSchedules(input scheduleModels.UpdateSchedulesInput) (*scheduleModels.UpdateSchedulesResult, error) {
	target := strings.ToLower(strings.TrimSpace(input.Target))
	action := strings.ToLower(strings.TrimSpace(input.Action))
	if action == "" {
		action = scheduleModels.ScheduleActionUpsert
	}
	if target == "" {
		target = inferScheduleTarget(input)
	}

	switch target {
	case scheduleModels.ScheduleTargetBase:
		return s.updateBaseSchedule(action, input)
	case scheduleModels.ScheduleTargetException:
		return s.updateScheduleException(action, input)
	case scheduleModels.ScheduleTargetPlanned:
		return s.updatePlannedSchedule(action, input)
	default:
		return nil, errors.New("invalid schedule target")
	}
}

func (s *ScheduleService) updateBaseSchedule(action string, input scheduleModels.UpdateSchedulesInput) (*scheduleModels.UpdateSchedulesResult, error) {
	if action == scheduleModels.ScheduleActionDelete {
		id := input.ID
		if id <= 0 && input.Schedule != nil {
			id = input.Schedule.ID
		}
		if id <= 0 {
			return nil, errors.New("schedule id is required")
		}
		if err := s.schedules.DeleteBaseSchedule(id); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_base_schedule",
				Message: "failed to delete base schedule: " + err.Error(),
			})
			return nil, err
		}
		return &scheduleModels.UpdateSchedulesResult{Target: scheduleModels.ScheduleTargetBase, Action: action}, nil
	}
	if action != scheduleModels.ScheduleActionUpsert {
		return nil, errors.New("invalid schedule action")
	}
	if input.Schedule == nil {
		return nil, errors.New("schedule payload is required")
	}

	schedule := *input.Schedule
	normalizeBaseSchedule(&schedule)
	if err := validateBaseSchedule(schedule); err != nil {
		return nil, err
	}

	result, err := s.schedules.UpsertBaseSchedule(schedule)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "upsert_base_schedule",
			Message: "failed to upsert base schedule: " + err.Error(),
		})
		return nil, err
	}

	return &scheduleModels.UpdateSchedulesResult{Target: scheduleModels.ScheduleTargetBase, Action: action, Schedule: result}, nil
}

func (s *ScheduleService) updateScheduleException(action string, input scheduleModels.UpdateSchedulesInput) (*scheduleModels.UpdateSchedulesResult, error) {
	if action == scheduleModels.ScheduleActionDelete {
		id := input.ID
		if id <= 0 && input.Exception != nil {
			id = input.Exception.ID
		}
		if id <= 0 {
			return nil, errors.New("exception id is required")
		}
		if err := s.schedules.DeleteScheduleException(id); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_schedule_exception",
				Message: "failed to delete schedule exception: " + err.Error(),
			})
			return nil, err
		}
		return &scheduleModels.UpdateSchedulesResult{Target: scheduleModels.ScheduleTargetException, Action: action}, nil
	}
	if action != scheduleModels.ScheduleActionUpsert {
		return nil, errors.New("invalid schedule action")
	}
	if input.Exception == nil {
		return nil, errors.New("exception payload is required")
	}

	exception := *input.Exception
	normalizeScheduleException(&exception)
	if err := validateScheduleException(exception); err != nil {
		return nil, err
	}

	result, err := s.schedules.UpsertScheduleException(exception)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "upsert_schedule_exception",
			Message: "failed to upsert schedule exception: " + err.Error(),
		})
		return nil, err
	}

	return &scheduleModels.UpdateSchedulesResult{Target: scheduleModels.ScheduleTargetException, Action: action, Exception: result}, nil
}

func (s *ScheduleService) updatePlannedSchedule(action string, input scheduleModels.UpdateSchedulesInput) (*scheduleModels.UpdateSchedulesResult, error) {
	if action == scheduleModels.ScheduleActionDelete {
		id := input.ID
		if id <= 0 && input.Planned != nil {
			id = input.Planned.ID
		}
		if id <= 0 {
			return nil, errors.New("planned schedule id is required")
		}
		if err := s.schedules.DeletePlannedSchedule(id); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_planned_schedule",
				Message: "failed to delete planned schedule: " + err.Error(),
			})
			return nil, err
		}
		return &scheduleModels.UpdateSchedulesResult{Target: scheduleModels.ScheduleTargetPlanned, Action: action}, nil
	}
	if action != scheduleModels.ScheduleActionUpsert {
		return nil, errors.New("invalid schedule action")
	}
	if input.Planned == nil {
		return nil, errors.New("planned schedule payload is required")
	}

	planned := *input.Planned
	normalizePlannedSchedule(&planned)
	if err := validatePlannedSchedule(planned); err != nil {
		return nil, err
	}

	result, err := s.schedules.UpsertPlannedSchedule(planned)
	if err != nil {
		logger.WriteSafe(logger.LogEntry{
			Level:   "error",
			Action:  "upsert_planned_schedule",
			Message: "failed to upsert planned schedule: " + err.Error(),
		})
		return nil, err
	}

	return &scheduleModels.UpdateSchedulesResult{Target: scheduleModels.ScheduleTargetPlanned, Action: action, Planned: result}, nil
}

func inferScheduleTarget(input scheduleModels.UpdateSchedulesInput) string {
	if input.Exception != nil {
		return scheduleModels.ScheduleTargetException
	}
	if input.Planned != nil {
		return scheduleModels.ScheduleTargetPlanned
	}
	return scheduleModels.ScheduleTargetBase
}

func normalizeBaseSchedule(schedule *scheduleModels.BaseSchedule) {
	schedule.DayOfWeek = strings.TrimSpace(schedule.DayOfWeek)
	schedule.WeekType = normalizeWeekType(schedule.WeekType)
	schedule.Subject = strings.TrimSpace(schedule.Subject)
	schedule.StartTime = strings.TrimSpace(schedule.StartTime)
	schedule.EndTime = strings.TrimSpace(schedule.EndTime)
	schedule.Description = strings.TrimSpace(schedule.Description)
}

func normalizeScheduleException(exception *scheduleModels.ScheduleException) {
	exception.Date = strings.TrimSpace(exception.Date)
	exception.Scope = strings.TrimSpace(exception.Scope)
	if exception.Scope == "" {
		exception.Scope = "lesson"
	}
	exception.ChangeType = scheduleModels.ChangeType(strings.ToLower(strings.TrimSpace(string(exception.ChangeType))))
	exception.Reason = strings.TrimSpace(exception.Reason)
	trimStringPtr(exception.NewSubject)
	trimStringPtr(exception.NewStartTime)
	trimStringPtr(exception.NewEndTime)
	trimStringPtr(exception.NewDescription)
}

func normalizePlannedSchedule(planned *scheduleModels.PlannedSchedule) {
	planned.Date = strings.TrimSpace(planned.Date)
	planned.Subject = strings.TrimSpace(planned.Subject)
	planned.ChangeType = scheduleModels.ChangeType(strings.ToLower(strings.TrimSpace(string(planned.ChangeType))))
	planned.Scope = strings.TrimSpace(planned.Scope)
	if planned.Scope == "" {
		planned.Scope = "lesson"
	}
	planned.StartTime = strings.TrimSpace(planned.StartTime)
	planned.EndTime = strings.TrimSpace(planned.EndTime)
	planned.Description = strings.TrimSpace(planned.Description)
	planned.Reason = strings.TrimSpace(planned.Reason)
}

func validateBaseSchedule(schedule scheduleModels.BaseSchedule) error {
	if schedule.ClassID <= 0 {
		return errors.New("class id is required")
	}
	if schedule.DayOfWeek == "" {
		return errors.New("day of week is required")
	}
	if schedule.LessonNumber <= 0 {
		return errors.New("lesson number is required")
	}
	if schedule.Subject == "" {
		return errors.New("subject is required")
	}
	if schedule.TeacherID <= 0 {
		return errors.New("teacher id is required")
	}
	if schedule.Room <= 0 {
		return errors.New("room is required")
	}
	if schedule.StartTime == "" || schedule.EndTime == "" {
		return errors.New("start time and end time are required")
	}

	return nil
}

func validateScheduleException(exception scheduleModels.ScheduleException) error {
	if exception.ScheduleID == nil && exception.ClassID <= 0 {
		return errors.New("class id or schedule id is required")
	}
	if exception.Date == "" {
		return errors.New("date is required")
	}
	if !isValidChangeType(exception.ChangeType) {
		return errors.New("invalid change type")
	}

	return nil
}

func validatePlannedSchedule(planned scheduleModels.PlannedSchedule) error {
	if planned.BaseScheduleID == nil && planned.ClassID <= 0 {
		return errors.New("class id is required")
	}
	if planned.Date == "" {
		return errors.New("date is required")
	}
	if planned.LessonNumber <= 0 {
		return errors.New("lesson number is required")
	}
	if planned.Subject == "" {
		return errors.New("subject is required")
	}
	if !isValidChangeType(planned.ChangeType) {
		return errors.New("invalid change type")
	}
	if planned.BaseScheduleID != nil {
		return nil
	}
	if planned.TeacherID <= 0 {
		return errors.New("teacher id is required")
	}
	if planned.Room <= 0 {
		return errors.New("room is required")
	}
	if planned.StartTime == "" || planned.EndTime == "" {
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

func isValidChangeType(changeType scheduleModels.ChangeType) bool {
	switch changeType {
	case scheduleModels.ChangeCancel,
		scheduleModels.ChangeReplace,
		scheduleModels.ChangeMove,
		scheduleModels.ChangeRoom,
		scheduleModels.ChangeTeacher,
		scheduleModels.ChangeUpdate,
		scheduleModels.ChangeAdd,
		scheduleModels.ChangeDayOff,
		scheduleModels.ChangeShortDay,
		scheduleModels.ChangeSwap:
		return true
	default:
		return false
	}
}

func trimStringPtr(value *string) {
	if value == nil {
		return
	}
	*value = strings.TrimSpace(*value)
}
