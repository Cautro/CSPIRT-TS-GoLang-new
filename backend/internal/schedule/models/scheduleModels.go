package models

import userModels "cspirt/internal/users/models"

type ChangeType string

const (
	ChangeCancel   ChangeType = "cancel"
	ChangeReplace  ChangeType = "replace"
	ChangeMove     ChangeType = "move"
	ChangeRoom     ChangeType = "room_change"
	ChangeTeacher  ChangeType = "teacher_change"
	ChangeUpdate   ChangeType = "update"
	ChangeAdd      ChangeType = "add"
	ChangeDayOff   ChangeType = "day_off"
	ChangeShortDay ChangeType = "short_day"
	ChangeSwap     ChangeType = "swap"
)

const (
	ScheduleTargetBase      = "base"
	ScheduleTargetException = "exception"
	ScheduleTargetPlanned   = "planned"

	ScheduleActionUpsert = "upsert"
	ScheduleActionDelete = "delete"

	ScheduleSourceBase      = "base"
	ScheduleSourceException = "exception"
	ScheduleSourcePlanned   = "planned"
)

type ScheduleFilter struct {
	ClassID  int
	Day      string
	Date     string
	WeekType string
}

type BaseSchedule struct {
	ID           int                  `json:"Id"`
	ClassID      int                  `json:"ClassID"`
	Class        string               `json:"Class,omitempty"`
	DayOfWeek    string               `json:"DayOfWeek"`
	LessonNumber int                  `json:"LessonNumber"`
	WeekType     string               `json:"WeekType"`
	Subject      string               `json:"Subject"`
	TeacherID    int                  `json:"TeacherID"`
	Teacher      *userModels.SafeUser `json:"Teacher,omitempty"`
	Room         int                  `json:"Room"`
	StartTime    string               `json:"StartTime"`
	EndTime      string               `json:"EndTime"`
	Description  string               `json:"Description"`
}

type ScheduleException struct {
	ID              int        `json:"Id"`
	ScheduleID      *int       `json:"ScheduleID,omitempty"`
	ClassID         int        `json:"ClassID"`
	Date            string     `json:"Date"`
	ChangeType      ChangeType `json:"ChangeType"`
	Scope           string     `json:"Scope"`
	NewSubject      *string    `json:"NewSubject,omitempty"`
	NewLessonNumber *int       `json:"NewLessonNumber,omitempty"`
	NewTeacherID    *int       `json:"NewTeacherID,omitempty"`
	NewRoom         *int       `json:"NewRoom,omitempty"`
	NewStartTime    *string    `json:"NewStartTime,omitempty"`
	NewEndTime      *string    `json:"NewEndTime,omitempty"`
	NewDescription  *string    `json:"NewDescription,omitempty"`
	Reason          string     `json:"Reason"`
	CreatedAt       string     `json:"CreatedAt"`
}

type PlannedSchedule struct {
	ID             int        `json:"Id"`
	BaseScheduleID *int       `json:"BaseScheduleID,omitempty"`
	ClassID        int        `json:"ClassID"`
	Date           string     `json:"Date"`
	LessonNumber   int        `json:"LessonNumber"`
	Subject        string     `json:"Subject"`
	ChangeType     ChangeType `json:"ChangeType"`
	Scope          string     `json:"Scope"`
	TeacherID      int        `json:"TeacherID"`
	Room           int        `json:"Room"`
	StartTime      string     `json:"StartTime"`
	EndTime        string     `json:"EndTime"`
	Description    string     `json:"Description"`
	Reason         string     `json:"Reason"`
	CreatedAt      string     `json:"CreatedAt"`
}

type ScheduleView struct {
	ID             int                  `json:"Id"`
	BaseScheduleID *int                 `json:"BaseScheduleID,omitempty"`
	ChangeID       *int                 `json:"ChangeID,omitempty"`
	Source         string               `json:"Source"`
	ChangeType     ChangeType           `json:"ChangeType,omitempty"`
	IsCancelled    bool                 `json:"IsCancelled"`
	ClassID        int                  `json:"ClassID"`
	Class          string               `json:"Class,omitempty"`
	DayOfWeek      string               `json:"DayOfWeek"`
	Date           string               `json:"Date,omitempty"`
	LessonNumber   int                  `json:"LessonNumber"`
	WeekType       string               `json:"WeekType"`
	Subject        string               `json:"Subject"`
	TeacherID      int                  `json:"TeacherID"`
	Teacher        *userModels.SafeUser `json:"Teacher,omitempty"`
	Room           int                  `json:"Room"`
	StartTime      string               `json:"StartTime"`
	EndTime        string               `json:"EndTime"`
	Description    string               `json:"Description"`
	Reason         string               `json:"Reason,omitempty"`
}

type SchedulesResponse struct {
	Schedules  []ScheduleView      `json:"Schedules"`
	Base       []BaseSchedule      `json:"Base"`
	Exceptions []ScheduleException `json:"Exceptions"`
	Planned    []PlannedSchedule   `json:"Planned"`
}

type UpdateSchedulesInput struct {
	Target    string             `json:"Target"`
	Action    string             `json:"Action"`
	ID        int                `json:"Id"`
	Schedule  *BaseSchedule      `json:"Schedule,omitempty"`
	Exception *ScheduleException `json:"Exception,omitempty"`
	Planned   *PlannedSchedule   `json:"Planned,omitempty"`
}

type UpdateSchedulesResult struct {
	Target    string             `json:"Target"`
	Action    string             `json:"Action"`
	Schedule  *BaseSchedule      `json:"Schedule,omitempty"`
	Exception *ScheduleException `json:"Exception,omitempty"`
	Planned   *PlannedSchedule   `json:"Planned,omitempty"`
}
