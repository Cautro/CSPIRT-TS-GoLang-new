package handlers

import (
	"net/http"
	"strconv"
	"testing"

	classModels "cspirt/internal/class/models"
	"cspirt/internal/handlertest"
	userModels "cspirt/internal/users/models"
)

func TestGetClassTeachersHandlerReturnsTeachers(t *testing.T) {
	st := handlertest.NewStorage(t)
	handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter("")
	router.GET("/api/classes/teacher", GetClassTeachersHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/classes/teacher", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		Teachers []userModels.SafeUser `json:"Teachers"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response.Teachers) == 0 {
		t.Fatal("expected at least one class teacher")
	}
}

func TestAddClassHandlerAddsClass(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/classes/add", AddClassHandler(st))

	body := classModels.ClassInput{
		Name:         "12C",
		TeacherLogin: users.Owner.Login,
	}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, "/api/classes/add", body))

	handlertest.RequireStatus(t, recorder, http.StatusCreated)

	if findClassByName(t, st, "12C") == nil {
		t.Fatal("added class was not persisted")
	}
}

func TestDeleteClassHandlerDeletesClass(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	if err := st.AddClass(classModels.ClassInput{Name: "12C", TeacherLogin: users.Owner.Login}, users.Owner.Login); err != nil {
		t.Fatalf("seed class returned error: %v", err)
	}
	class := findClassByName(t, st, "12C")
	if class == nil {
		t.Fatal("seeded class not found")
	}

	router := handlertest.NewRouter(users.Owner.Login)
	router.DELETE("/api/classes/delete/:id", DeleteClassHandler(st))

	target := "/api/classes/delete/" + strconv.Itoa(class.ID)
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodDelete, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	if findClassByName(t, st, "12C") != nil {
		t.Fatal("class was not deleted")
	}
}

func TestGetClassesHandlerReturnsClasses(t *testing.T) {
	st := handlertest.NewStorage(t)
	handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter("")
	router.GET("/api/classes", GetClassesHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/classes", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		Classes []classModels.Class `json:"Classes"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response.Classes) == 0 {
		t.Fatal("expected classes in response")
	}
}

func TestGetClassUsersHandlerReturnsClassUsers(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Helper.Login)
	router.GET("/api/classes/:class_id/users", GetClassUsersHandler(st))

	target := "/api/classes/" + strconv.Itoa(users.Helper.ClassID) + "/users"
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		Users []userModels.SafeUser `json:"Users"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response.Users) != 2 {
		t.Fatalf("expected helper and student in class, got %d", len(response.Users))
	}
}

func TestGetClassTeacherHandlerReturnsTeacher(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Student.Login)
	router.GET("/api/classes/:class_id/teacher", GetClassTeacherHandler(st))

	target := "/api/classes/" + strconv.Itoa(users.Student.ClassID) + "/teacher"
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		Teacher *userModels.SafeUser `json:"Teacher"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if response.Teacher == nil || response.Teacher.Login != users.Helper.Login {
		t.Fatalf("unexpected teacher response: %+v", response.Teacher)
	}
}

func TestSetClassTeacherHandlerUpdatesTeacher(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/classes/:class_id/teacher", SetClassTeacherHandler(st))

	target := "/api/classes/" + strconv.Itoa(users.Student.ClassID) + "/teacher"
	body := classModels.ClassTeacherInput{TeacherLogin: users.Owner.Login}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, target, body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	teacher, err := st.GetClassTeacherByID(users.Student.ClassID)
	if err != nil {
		t.Fatalf("get class teacher returned error: %v", err)
	}
	if teacher == nil || teacher.Login != users.Owner.Login {
		t.Fatalf("teacher was not updated: %+v", teacher)
	}
}

func findClassByName(t *testing.T, st interface {
	GetAllClasses() ([]classModels.Class, error)
}, name string) *classModels.Class {
	t.Helper()

	classes, err := st.GetAllClasses()
	if err != nil {
		t.Fatalf("get classes returned error: %v", err)
	}
	for i := range classes {
		if classes[i].Name == name {
			return &classes[i]
		}
	}

	return nil
}
