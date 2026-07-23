package repo

import (
	"cspirt/internal/domain/schedule"
)

type ScheduleRepository interface {
	GetSchedules(filter schedule.ScheduleFilter) (*schedule.SchedulesResponse, error)
	GetCurrentScheduleForTeacher(teacherID int, filter schedule.ScheduleFilter) ([]schedule.ScheduleLesson, error)
	GetScheduleLessonByID(scheduleType string, id int) (*schedule.ScheduleLesson, error)
	UpsertScheduleLesson(scheduleType string, lessons schedule.ScheduleLesson) (*schedule.ScheduleLesson, error)
	DeleteScheduleLesson(scheduleType string, id int) error
	RolloverSchedules(classID int) (*schedule.ScheduleRolloverResult, error)
	ResetPlannedSchedules(classID int) (*schedule.ScheduleResetResult, error)
}
