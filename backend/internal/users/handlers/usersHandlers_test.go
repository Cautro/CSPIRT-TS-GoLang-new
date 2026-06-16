package handlers

import (
	"net/http"
	"strconv"
	"testing"

	"cspirt/internal/handlertest"
	userModels "cspirt/internal/users/models"
)

func TestGetUsersHandlerReturnsUsers(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.GET("/api/users", GetUsersHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/users", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response []userModels.SafeUser
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response) < 5 {
		t.Fatalf("expected seeded users, got %d", len(response))
	}
}

func TestAddUserHandlerAddsUser(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/user/add", AddUserHandler(st))

	body := userModels.User{
		Name:     "newstudent",
		LastName: "Test",
		FullName: []userModels.FullName{{
			Name:     "newstudent",
			LastName: "Test",
		}},
		Login:    "newstudent",
		Password: handlertest.Password,
		Rating:   700,
		Role:     "User",
		Class:    "10A",
	}

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, "/api/user/add", body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	got, err := st.GetUserByLogin("newstudent")
	if err != nil {
		t.Fatalf("get added user returned error: %v", err)
	}
	if got == nil {
		t.Fatal("added user was not persisted")
	}
}

func TestDeleteUserHandlerDeletesUser(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.DELETE("/api/user/delete/:id", DeleteUserHandler(st))

	target := "/api/user/delete/" + strconv.Itoa(users.OtherStudent.ID)
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodDelete, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	deleted, err := st.GetUserByLogin(users.OtherStudent.Login)
	if err != nil {
		t.Fatalf("get deleted user returned error: %v", err)
	}
	if deleted != nil {
		t.Fatalf("expected user to be deleted, got %+v", deleted)
	}
}

func TestGetMeHandlerReturnsCurrentUser(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Student.Login)
	router.GET("/api/me", GetMeHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/me", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response userModels.UserWithFullInfo
	handlertest.DecodeJSON(t, recorder, &response)
	if response.User == nil || response.User.Login != users.Student.Login {
		t.Fatalf("unexpected current user response: %+v", response.User)
	}
}

func TestGetStaffHandlerReturnsStaff(t *testing.T) {
	st := handlertest.NewStorage(t)
	handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter("")
	router.GET("/api/users/get/staff", GetStaffHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/users/get/staff", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response []userModels.SafeUser
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response) != 2 {
		t.Fatalf("expected owner and admin staff users, got %d", len(response))
	}
}
