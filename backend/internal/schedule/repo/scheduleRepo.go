package repo

import scheduleModels "cspirt/internal/schedule/models"

type ScheduleRepository interface {
	GetSchedules(filter scheduleModels.ScheduleFilter) (*scheduleModels.SchedulesResponse, error)
	UpsertBaseSchedule(schedule scheduleModels.BaseSchedule) (*scheduleModels.BaseSchedule, error)
	DeleteBaseSchedule(id int) error
	UpsertScheduleException(exception scheduleModels.ScheduleException) (*scheduleModels.ScheduleException, error)
	DeleteScheduleException(id int) error
	UpsertPlannedSchedule(planned scheduleModels.PlannedSchedule) (*scheduleModels.PlannedSchedule, error)
	DeletePlannedSchedule(id int) error
	GetBaseScheduleByID(id int) (*scheduleModels.BaseSchedule, error)
	GetScheduleExceptionByID(id int) (*scheduleModels.ScheduleException, error)
	GetPlannedScheduleByID(id int) (*scheduleModels.PlannedSchedule, error)
}
