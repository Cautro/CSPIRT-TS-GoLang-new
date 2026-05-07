package repo

import scheduleModels "cspirt/internal/schedule/models"

type ScheduleRepository interface {
	GetSchedules(filter scheduleModels.ScheduleFilter) (*scheduleModels.SchedulesResponse, error)
	GetCurrentScheduleForTeacher(teacherID int, filter scheduleModels.ScheduleFilter) ([]scheduleModels.ScheduleLesson, error)
	GetScheduleLessonByID(scheduleType string, id int) (*scheduleModels.ScheduleLesson, error)
	UpsertScheduleLesson(scheduleType string, lesson scheduleModels.ScheduleLesson) (*scheduleModels.ScheduleLesson, error)
	DeleteScheduleLesson(scheduleType string, id int) error
	RolloverSchedules(classID int) (*scheduleModels.ScheduleRolloverResult, error)
	ResetPlannedSchedules(classID int) (*scheduleModels.ScheduleResetResult, error)
}
