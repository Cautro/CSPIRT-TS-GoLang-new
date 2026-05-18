package handlers

import (
	"net/http"
	"strconv"
	"testing"

	"cspirt/internal/events/models"
	"cspirt/internal/handlertest"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"
)

func TestGetEventsHandlerReturnsEvents(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	seedEvent(t, st, []int{users.Student.ID}, []int{users.Student.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.GET("/api/events", GetEventsHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/events", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response []models.Event
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response) != 1 {
		t.Fatalf("expected 1 event, got %d", len(response))
	}
}

func TestAddEventHandlerAddsEvent(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/event/add", AddEventHandler(st))

	body := models.Event{
		Title:            "Tournament",
		Status:           models.EventStatusScheduled,
		Description:      "Class tournament",
		StartedAt:        "2999-01-01 10:00:00",
		Players:          []int{users.Student.ID},
		Classes:          []int{users.Student.ClassID},
		BaseRatingReward: 100,
	}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, "/api/event/add", body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	events, err := st.GetEvents()
	if err != nil {
		t.Fatalf("get events returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event after add, got %d", len(events))
	}
}

func TestDeleteEventHandlerDeletesEvent(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	event := seedEvent(t, st, []int{users.Student.ID}, []int{users.Student.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.DELETE("/api/event/delete/:id", DeleteEventHandler(st))

	target := "/api/event/delete/" + strconv.Itoa(event.ID)
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodDelete, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	events, err := st.GetEvents()
	if err != nil {
		t.Fatalf("get events returned error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected event to be deleted, got %d events", len(events))
	}
}

func TestEventParamsHandlersReturnAllParams(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	event := seedEvent(t, st, []int{users.Student.ID}, []int{users.Student.ClassID, users.OtherStudent.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/event/:eventId/params/add", AddEventParams(st))
	router.GET("/api/event/:eventId/params", GetEventParamsHandler(st))

	target := "/api/event/" + strconv.Itoa(event.ID) + "/params"
	first := models.EventParams{
		ExtraRatingReward: 560,
		Reason:            "Event win",
		ClassID:           users.Student.ClassID,
	}
	second := models.EventParams{
		ExtraRatingReward: 120,
		Reason:            "Active participation",
		ClassID:           users.OtherStudent.ClassID,
	}

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, target+"/add", first))
	handlertest.RequireStatus(t, recorder, http.StatusOK)
	recorder = handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, target+"/add", second))
	handlertest.RequireStatus(t, recorder, http.StatusOK)

	recorder = handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, target, nil))
	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response []models.EventParams
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response) != 2 {
		t.Fatalf("expected 2 event params, got %d: %+v", len(response), response)
	}
	if response[0].EventID != event.ID || response[0].ExtraRatingReward != 560 || response[0].Reason != first.Reason || response[0].ClassID != users.Student.ClassID {
		t.Fatalf("unexpected first event params: %+v", response[0])
	}
	if response[1].EventID != event.ID || response[1].ExtraRatingReward != 120 || response[1].Reason != second.Reason || response[1].ClassID != users.OtherStudent.ClassID {
		t.Fatalf("unexpected second event params: %+v", response[1])
	}
}

func TestAddPlayersToEventAddsPlayers(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	event := seedEvent(t, st, nil, []int{users.Student.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/event/:eventId/players/add", AddPlayersToEvent(st))

	target := "/api/event/" + strconv.Itoa(event.ID) + "/players/add"
	body := map[string][]int{"playerIds": []int{users.Student.ID}}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, target, body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	count, err := st.GetEventPlayersCount(event.ID)
	if err != nil {
		t.Fatalf("get event players count returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 player, got %d", count)
	}
}

func TestDeletePlayersFromEventDeletesPlayers(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	event := seedEvent(t, st, []int{users.Student.ID}, []int{users.Student.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.DELETE("/api/event/:eventId/players/delete", DeletePlayersFromEvent(st))

	target := "/api/event/" + strconv.Itoa(event.ID) + "/players/delete"
	body := map[string][]int{"playerIds": []int{users.Student.ID}}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodDelete, target, body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	count, err := st.GetEventPlayersCount(event.ID)
	if err != nil {
		t.Fatalf("get event players count returned error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 players, got %d", count)
	}
}

func TestGetEventPlayersHandlerReturnsPlayers(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	event := seedEvent(t, st, []int{users.Student.ID}, []int{users.Student.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.GET("/api/event/:eventId/players", GetEventPlayersHandler(st))

	target := "/api/event/" + strconv.Itoa(event.ID) + "/players"
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response []userModels.SafeUser
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response) != 1 || response[0].Login != users.Student.Login {
		t.Fatalf("unexpected players response: %+v", response)
	}
}

func TestGetEventPlayersCountHandlerReturnsCount(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	event := seedEvent(t, st, []int{users.Student.ID}, []int{users.Student.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.GET("/api/event/:eventId/players/count", GetEventPlayersCountHandler(st))

	target := "/api/event/" + strconv.Itoa(event.ID) + "/players/count"
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		Count int `json:"count"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if response.Count != 1 {
		t.Fatalf("expected count 1, got %d", response.Count)
	}
}

func TestEventCompleteCompletesEvent(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	event := seedEvent(t, st, []int{users.Student.ID}, []int{users.Student.ClassID})

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/event/:eventId/complete", EventComplete(st))

	target := "/api/event/" + strconv.Itoa(event.ID) + "/complete"
	body := map[string]int{"ratingReward": 100}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, target, body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	completed, err := st.GetEventsByID(event.ID)
	if err != nil {
		t.Fatalf("get completed event returned error: %v", err)
	}
	if completed == nil || completed.Status != models.EventStatusCompleted {
		t.Fatalf("event was not completed: %+v", completed)
	}

	student, err := st.GetUserByLogin(users.Student.Login)
	if err != nil {
		t.Fatalf("get student returned error: %v", err)
	}
	if student == nil || student.Rating != users.Student.Rating+100 {
		t.Fatalf("student rating was not updated: %+v", student)
	}
}

func seedEvent(t *testing.T, st *storage.Storage, players []int, classes []int) models.Event {
	t.Helper()

	if err := st.AddEvent(models.Event{
		Title:            "Seed event",
		Status:           models.EventStatusScheduled,
		Description:      "Seeded event",
		StartedAt:        "2999-01-01 10:00:00",
		Players:          players,
		Classes:          classes,
		BaseRatingReward: 50,
	}); err != nil {
		t.Fatalf("seed event returned error: %v", err)
	}

	events, err := st.GetEvents()
	if err != nil {
		t.Fatalf("get seeded events returned error: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("seeded event not found")
	}

	return events[len(events)-1]
}
