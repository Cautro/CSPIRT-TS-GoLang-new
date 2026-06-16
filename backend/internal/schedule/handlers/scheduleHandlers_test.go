package handlers

import (
	"net/http"
	"strconv"
	"testing"

	"cspirt/internal/handlertest"
	scheduleModels "cspirt/internal/schedule/models"
	"cspirt/internal/storage"
)

func TestGetSchedulesHandlerReturnsCurrentSchedules(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	seedCurrentSchedule(t, st, users)

	router := handlertest.NewRouter(users.Student.Login)
	router.GET("/api/schedules", GetSchedulesHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/schedules?type=current", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response scheduleModels.SchedulesResponse
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response.Current) != 1 {
		t.Fatalf("expected 1 current schedule, got %d", len(response.Current))
	}
}

func TestGetTeacherCurrentScheduleHandlerReturnsSchedules(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	seedCurrentSchedule(t, st, users)

	router := handlertest.NewRouter(users.Helper.Login)
	router.GET("/api/schedules/teacher/current", GetTeacherCurrentScheduleHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/schedules/teacher/current", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		Schedules []scheduleModels.ScheduleLesson `json:"Schedules"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response.Schedules) != 1 {
		t.Fatalf("expected 1 teacher schedule, got %d", len(response.Schedules))
	}
}

func TestUpdateSchedulesHandlerUpsertsSchedule(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/schedules/update", UpdateSchedulesHandler(st))

	lesson := scheduleLesson(users)
	body := scheduleModels.UpdateSchedulesInput{
		Type:   scheduleModels.ScheduleTypeBase,
		Action: scheduleModels.ScheduleActionUpsert,
		Lesson: &lesson,
	}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, "/api/schedules/update", body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response scheduleModels.UpdateSchedulesResult
	handlertest.DecodeJSON(t, recorder, &response)
	if response.Lesson == nil || response.Lesson.ID <= 0 {
		t.Fatalf("unexpected update response: %+v", response)
	}
}

func TestRolloverSchedulesHandlerPromotesBaseSchedule(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	seedBaseSchedule(t, st, users)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/schedules/rollover", RolloverSchedulesHandler(st))

	target := "/api/schedules/rollover?class_id=" + strconv.Itoa(users.Student.ClassID)
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response scheduleModels.ScheduleRolloverResult
	handlertest.DecodeJSON(t, recorder, &response)
	if response.CurrentCount != 1 {
		t.Fatalf("expected 1 current schedule after rollover, got %+v", response)
	}
}

func TestResetPlannedSchedulesHandlerCopiesBaseSchedule(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	seedBaseSchedule(t, st, users)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/schedules/planned/reset", ResetPlannedSchedulesHandler(st))

	target := "/api/schedules/planned/reset?class_id=" + strconv.Itoa(users.Student.ClassID)
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response scheduleModels.ScheduleResetResult
	handlertest.DecodeJSON(t, recorder, &response)
	if response.PlannedCount != 1 {
		t.Fatalf("expected 1 planned schedule after reset, got %+v", response)
	}
}

func seedBaseSchedule(t *testing.T, st *storage.Storage, users handlertest.UsersFixture) *scheduleModels.BaseSchedule {
	t.Helper()

	lesson := scheduleLesson(users)
	result, err := st.UpsertBaseSchedule(lesson)
	if err != nil {
		t.Fatalf("seed base schedule returned error: %v", err)
	}

	return result
}

func seedCurrentSchedule(t *testing.T, st *storage.Storage, users handlertest.UsersFixture) *scheduleModels.CurrentSchedule {
	t.Helper()

	lesson := scheduleLesson(users)
	result, err := st.UpsertCurrentSchedule(lesson)
	if err != nil {
		t.Fatalf("seed current schedule returned error: %v", err)
	}

	return result
}

func scheduleLesson(users handlertest.UsersFixture) scheduleModels.ScheduleLesson {
	return scheduleModels.ScheduleLesson{
		ClassID:      users.Student.ClassID,
		DayOfWeek:    "monday",
		LessonNumber: 1,
		WeekType:     "all",
		Subject:      "Math",
		TeacherID:    users.Helper.ID,
		Room:         101,
		StartTime:    "08:30",
		EndTime:      "09:15",
		Description:  "handler test",
	}
}
