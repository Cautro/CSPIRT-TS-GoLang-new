package handlertest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ratingModels "cspirt/internal/rating/models"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"

	"github.com/gin-gonic/gin"
)

const (
	Secret   = "test-secret"
	Password = "secret123"
)

type UsersFixture struct {
	Owner        *userModels.User
	Admin        *userModels.User
	Helper       *userModels.User
	Student      *userModels.User
	OtherStudent *userModels.User
}

func NewStorage(t testing.TB) *storage.Storage {
	t.Helper()

	st, err := storage.NewUserStorage(t.TempDir()+"/storage.db", Secret)
	if err != nil {
		t.Fatalf("new test storage returned error: %v", err)
	}

	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Fatalf("close test storage returned error: %v", err)
		}
	})

	return st
}

func SeedUsers(t testing.TB, st *storage.Storage) UsersFixture {
	t.Helper()

	return UsersFixture{
		Owner:        AddUser(t, st, "owner", string(ratingModels.RoleOwner), "", 1000),
		Admin:        AddUser(t, st, "admin", string(ratingModels.RoleAdmin), "", 1000),
		Helper:       AddUser(t, st, "helper", string(ratingModels.RoleHelper), "10A", 1000),
		Student:      AddUser(t, st, "student", string(ratingModels.RoleUser), "10A", 500),
		OtherStudent: AddUser(t, st, "otherstudent", string(ratingModels.RoleUser), "11B", 400),
	}
}

func AddUser(t testing.TB, st *storage.Storage, login string, role string, className string, rating int) *userModels.User {
	t.Helper()

	user := userModels.User{
		Name:     login,
		LastName: "Test",
		FullName: []userModels.FullName{{
			Name:     login,
			LastName: "Test",
		}},
		Login:    login,
		Password: Password,
		Rating:   rating,
		Role:     role,
		Class:    className,
	}

	if err := st.AddUser(user); err != nil {
		t.Fatalf("add user %q returned error: %v", login, err)
	}

	got, err := st.GetUserByLogin(login)
	if err != nil {
		t.Fatalf("get user %q returned error: %v", login, err)
	}
	if got == nil {
		t.Fatalf("user %q was not saved", login)
	}

	return got
}

func NewRouter(login string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	if login != "" {
		router.Use(func(c *gin.Context) {
			c.Set("Login", login)
			c.Next()
		})
	}

	return router
}

func JSONRequest(t testing.TB, method string, target string, body any) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode request body returned error: %v", err)
		}
	}

	request := httptest.NewRequest(method, target, &buf)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	return request
}

func Perform(router http.Handler, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func RequireStatus(t testing.TB, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()

	if recorder.Code != want {
		t.Fatalf("expected status %d, got %d: %s", want, recorder.Code, recorder.Body.String())
	}
}

func DecodeJSON(t testing.TB, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response body returned error: %v; body=%s", err, recorder.Body.String())
	}
}
