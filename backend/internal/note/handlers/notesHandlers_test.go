package handlers

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"cspirt/internal/handlertest"
	noteModels "cspirt/internal/note/models"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"
	"cspirt/internal/utils"
)

func TestGetNotesHandlerReturnsNotes(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	seedNote(t, st, users.Owner, users.Student)

	router := handlertest.NewRouter(users.Owner.Login)
	router.GET("/api/notes", GetNotesHandler(st))

	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodGet, "/api/notes", nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	var response struct {
		AllNotes []userModels.Note `json:"All_notes"`
	}
	handlertest.DecodeJSON(t, recorder, &response)
	if len(response.AllNotes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(response.AllNotes))
	}
}

func TestAddNoteHandlerAddsNote(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)

	router := handlertest.NewRouter(users.Owner.Login)
	router.PATCH("/api/note/add", AddNoteHandler(st))

	body := noteModels.AddNewNoteResponse{
		TargetID:   users.Student.ID,
		TargetName: users.Student.Name + " " + users.Student.LastName,
		Content:    "steady progress",
	}
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodPatch, "/api/note/add", body))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	notes, err := st.GetAllNotes()
	if err != nil {
		t.Fatalf("get notes returned error: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note after add, got %d", len(notes))
	}
}

func TestDeleteNoteHandlerDeletesNote(t *testing.T) {
	st := handlertest.NewStorage(t)
	users := handlertest.SeedUsers(t, st)
	noteID := seedNote(t, st, users.Owner, users.Student)

	router := handlertest.NewRouter(users.Owner.Login)
	router.DELETE("/api/note/delete/:id", DeleteNoteHandler(st))

	target := "/api/note/delete/" + strconv.Itoa(noteID)
	recorder := handlertest.Perform(router, handlertest.JSONRequest(t, http.MethodDelete, target, nil))

	handlertest.RequireStatus(t, recorder, http.StatusOK)

	notes, err := st.GetAllNotes()
	if err != nil {
		t.Fatalf("get notes returned error: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected notes to be deleted, got %d", len(notes))
	}
}

func seedNote(t *testing.T, st *storage.Storage, author *userModels.User, target *userModels.User) int {
	t.Helper()

	if err := st.AddNote(author.Login, userModels.Note{
		TargetID:   target.ID,
		TargetName: target.Name + " " + target.LastName,
		AuthorID:   author.ID,
		AuthorName: author.Name + " " + author.LastName,
		Content:    "seed note",
		CreatedAt:  time.Now(),
	}, *utils.UserToSafeUser(*author)); err != nil {
		t.Fatalf("seed note returned error: %v", err)
	}

	notes, err := st.GetAllNotes()
	if err != nil {
		t.Fatalf("get seeded notes returned error: %v", err)
	}
	if len(notes) == 0 {
		t.Fatal("seeded note not found")
	}

	return notes[len(notes)-1].ID
}
