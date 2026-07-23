package schedule

import userModels "cspirt/internal/domain/user"

const (
	ScheduleTypeBase    = "base"
	ScheduleTypeCurrent = "current"
	ScheduleTypePlanned = "planned"
	ScheduleTypeAll     = "all"

	ScheduleActionUpsert = "upsert"
	ScheduleActionDelete = "delete"
)

type ScheduleFilter struct {
	Type      string
	ClassID   int
	TeacherID int
	Day       string
	WeekType  string
}

type ScheduleLesson struct {
	ID             int                  `json:"Id"`
	Type           string               `json:"Type,omitempty"`
	BaseScheduleID *int                 `json:"BaseScheduleID,omitempty"`
	ClassID        int                  `json:"ClassID"`
	Class          string               `json:"Class,omitempty"`
	DayOfWeek      string               `json:"DayOfWeek"`
	LessonNumber   int                  `json:"LessonNumber"`
	WeekType       string               `json:"WeekType"`
	Subject        string               `json:"Subject"`
	TeacherID      int                  `json:"TeacherID"`
	Teacher        *userModels.SafeUser `json:"Teacher,omitempty"`
	Room           int                  `json:"Room"`
	StartTime      string               `json:"StartTime"`
	EndTime        string               `json:"EndTime"`
	Description    string               `json:"Description"`
	CreatedAt      string               `json:"CreatedAt,omitempty"`
}

type BaseSchedule = ScheduleLesson
type CurrentSchedule = ScheduleLesson
type PlannedSchedule = ScheduleLesson

type SchedulesResponse struct {
	Schedules []ScheduleLesson `json:"Schedules"`
	Base      []ScheduleLesson `json:"Base"`
	Current   []ScheduleLesson `json:"Current"`
	Planned   []ScheduleLesson `json:"Planned"`
}

type UpdateSchedulesInput struct {
	Target string `json:"Target"`
	Type   string `json:"Type"`
	Action string `json:"Action"`
	ID     int    `json:"Id"`

	Lesson   *ScheduleLesson `json:"Lesson,omitempty"`
	Schedule *ScheduleLesson `json:"Schedule,omitempty"`
}

type UpdateSchedulesResult struct {
	Target string          `json:"Target"`
	Action string          `json:"Action"`
	Lesson *ScheduleLesson `json:"Lesson,omitempty"`

	Schedule *ScheduleLesson `json:"Schedule,omitempty"`
}

type ScheduleRolloverResult struct {
	Source         string `json:"Source"`
	ClassID        int    `json:"ClassID,omitempty"`
	CurrentCount   int    `json:"CurrentCount"`
	PlannedCleared int    `json:"PlannedCleared"`
}

type ScheduleResetResult struct {
	Source       string `json:"Source"`
	ClassID      int    `json:"ClassID,omitempty"`
	PlannedCount int    `json:"PlannedCount"`
}
