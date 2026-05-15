package handlers

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	complaintModels "cspirt/internal/complaints/models"
	"cspirt/internal/handlertest"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"
	"cspirt/internal/utils"
)

func TestGetComplaintsHandlerReturnsComplaints(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	seedComplaint(t, st, users.Owner, users.Student)

	router := handlertest.NewRouter(users.Owner.Login)
	router.GET("/api/complaints", GetComplaintsHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/complaints", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		AllComplaints []userModels.Complaint `json:"All_Complaints"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response.AllComplaints) != 1 {
		t.Fatalf("expected 1 complaint, got %d", len(response.AllComplaints))
	}
}

func TestAddComplaintHandlerAddsComplaint(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/complaint/add", AddcomplaintHandler(st))

	body := complaintModels.AddNewComplaintResponse{
		TargetID: users.Student.ID,
		Content:  "needs follow-up",
	}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, "/api/complaint/add", body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	complaints, err := st.GetAllComplaints()
	if err != nil {
		t.Fatalf("get complaints returned error: %v", err)
	}
	if len(complaints) != 1 {
		t.Fatalf("expected 1 complaint after add, got %d", len(complaints))
	}
}

func TestDeleteComplaintHandlerDeletesComplaint(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	complaintID := seedComplaint(t, st, users.Owner, users.Student)

	router := handlertest.NewRouter(users.Owner.Login)
	router.DELETE("/api/complaint/delete/:id", DeletecomplaintHandler(st))

	target := "/api/complaint/delete/" + strconv.Itoa(complaintID)
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodDelete, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	complaints, err := st.GetAllComplaints()
	if err != nil {
		t.Fatalf("get complaints returned error: %v", err)
	}
	if len(complaints) != 0 {
		t.Fatalf("expected complaints to be deleted, got %d", len(complaints))
	}
}

func seedComplaint(t *testing.T, st *storage.Storage, author *userModels.User, target *userModels.User) int {
	t.Helper()

	if err := st.AddComplaint(author.Login, userModels.Complaint{
		TargetID:  target.ID,
		Content:   "seed complaint",
		CreatedAt: time.Now(),
	}, *utils.UserToSafeUser(*author)); err != nil {
		t.Fatalf("seed complaint returned error: %v", err)
	}

	complaints, err := st.GetAllComplaints()
	if err != nil {
		t.Fatalf("get seeded complaints returned error: %v", err)
	}
	if len(complaints) == 0 {
		t.Fatal("seeded complaint not found")
	}

	return complaints[len(complaints)-1].ID
}
