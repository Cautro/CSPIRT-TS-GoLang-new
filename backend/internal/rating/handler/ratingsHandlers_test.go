package handlers

import (
	"net/http"
	"testing"

	"cspirt/internal/handlertest"
	ratingModels "cspirt/internal/rating/models"
	ratingService "cspirt/internal/rating/service"
)

func TestGetRatingsHandlerReturnsRating(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Student.Login)
	router.GET("/api/rating", GetRatingsHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/rating", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		Rating int `json:"Rating"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if response.Rating != users.Student.Rating {
		t.Fatalf("expected rating %d, got %d", users.Student.Rating, response.Rating)
	}
}

func TestUpdateRatingsHandlerUpdatesRating(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	service := ratingService.NewRatingsService(st.RatingRepo, st, st.Secret)
	router.PATCH("/api/rating/update", UpdateRatingsHandler(service, st))

	body := ratingModels.RatingInput{
		TargetLogin: users.Student.Login,
		Rating:      150,
		Reason:      "handler test",
	}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, "/api/rating/update", body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	updated, err := st.GetUserByLogin(users.Student.Login)
	if err != nil {
		t.Fatalf("get updated user returned error: %v", err)
	}
	if updated == nil || updated.Rating != users.Student.Rating+150 {
		t.Fatalf("rating was not updated correctly: %+v", updated)
	}
}
