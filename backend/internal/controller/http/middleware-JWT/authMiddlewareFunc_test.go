package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareAcceptsAccessTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const secret = "test-secret"

	token, err := GenerateToken(12, "cookie-user", "test", secret)
	if err != nil {
		t.Fatalf("generate token returned error: %v", err)
	}

	router := gin.New()
	router.GET("/api/private", AuthMiddleware(secret), func(c *gin.Context) {
		login, ok := c.Get("Login")
		if !ok {
			t.Fatal("login was not set in context")
		}

		c.JSON(http.StatusOK, gin.H{"login": login})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/private", nil)
	request.AddCookie(&http.Cookie{
		Name:  AccessTokenCookieName,
		Value: token,
	})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != `{"login":"cookie-user"}` {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestAuthMiddlewareRequiresBearerOrAccessTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/private", AuthMiddleware("test-secret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/private", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}
