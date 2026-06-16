package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ratingModels "cspirt/internal/rating/models"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"
	"cspirt/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	handlerTestPassword = "secret123"
	handlerTestSecret   = "test-secret"
)

func TestLoginHandlerSetsAccessAndRefreshTokenCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := newHandlerTestStorage(t)
	addHandlerTestUser(t, st)

	router := gin.New()
	router.POST("/login", LoginHandler(st))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"Login":"cookie-user","Password":"secret123"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	accessCookie := findCookie(recorder.Result().Cookies(), utils.AccessTokenCookieName)
	if accessCookie == nil {
		t.Fatal("access token cookie was not set")
	}
	if accessCookie.Value == "" {
		t.Fatal("access token cookie is empty")
	}
	if accessCookie.Path != "/api" {
		t.Fatalf("expected access cookie path /api, got %q", accessCookie.Path)
	}
	if !accessCookie.HttpOnly {
		t.Fatal("access token cookie must be HttpOnly")
	}
	if accessCookie.MaxAge != utils.AccessTokenCookieMaxAge {
		t.Fatalf("expected access cookie max age %d, got %d", utils.AccessTokenCookieMaxAge, accessCookie.MaxAge)
	}

	refreshCookie := findCookie(recorder.Result().Cookies(), utils.RefreshTokenCookieName)
	if refreshCookie == nil {
		t.Fatal("refresh token cookie was not set")
	}
	if refreshCookie.Value == "" {
		t.Fatal("refresh token cookie is empty")
	}
	if refreshCookie.Path != "/api/refresh" {
		t.Fatalf("expected refresh cookie path /api/refresh, got %q", refreshCookie.Path)
	}
	if !refreshCookie.HttpOnly {
		t.Fatal("refresh token cookie must be HttpOnly")
	}
}

func TestRefreshHandlerSetsNewAccessTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st := newHandlerTestStorage(t)
	addHandlerTestUser(t, st)

	router := gin.New()
	router.POST("/login", LoginHandler(st))
	router.POST("/api/refresh", RefreshHandler(st))

	loginRecorder := httptest.NewRecorder()
	loginRequest := httptest.NewRequest(
		http.MethodPost,
		"/login",
		strings.NewReader(`{"Login":"cookie-user","Password":"secret123"}`),
	)
	loginRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(loginRecorder, loginRequest)

	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d: %s", http.StatusOK, loginRecorder.Code, loginRecorder.Body.String())
	}

	refreshCookie := findCookie(loginRecorder.Result().Cookies(), utils.RefreshTokenCookieName)
	if refreshCookie == nil {
		t.Fatal("refresh token cookie was not set on login")
	}

	refreshRecorder := httptest.NewRecorder()
	refreshRequest := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	refreshRequest.AddCookie(refreshCookie)

	router.ServeHTTP(refreshRecorder, refreshRequest)

	if refreshRecorder.Code != http.StatusOK {
		t.Fatalf("expected refresh status %d, got %d: %s", http.StatusOK, refreshRecorder.Code, refreshRecorder.Body.String())
	}

	accessCookie := findCookie(refreshRecorder.Result().Cookies(), utils.AccessTokenCookieName)
	if accessCookie == nil {
		t.Fatal("access token cookie was not refreshed")
	}
	if accessCookie.Value == "" {
		t.Fatal("refreshed access token cookie is empty")
	}
	if accessCookie.Path != "/api" {
		t.Fatalf("expected access cookie path /api, got %q", accessCookie.Path)
	}
	if !accessCookie.HttpOnly {
		t.Fatal("access token cookie must be HttpOnly")
	}
}

func newHandlerTestStorage(t *testing.T) *storage.Storage {
	t.Helper()

	st, err := storage.NewUserStorage(t.TempDir()+"/storage.db", handlerTestSecret)
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

func addHandlerTestUser(t *testing.T, st *storage.Storage) {
	t.Helper()

	user := userModels.User{
		Name:     "Cookie",
		LastName: "User",
		FullName: []userModels.FullName{{
			Name:     "Cookie",
			LastName: "User",
		}},
		Login:    "cookie-user",
		Password: handlerTestPassword,
		Role:     string(ratingModels.RoleOwner),
	}

	if err := st.AddUser(user); err != nil {
		t.Fatalf("add test user returned error: %v", err)
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}
