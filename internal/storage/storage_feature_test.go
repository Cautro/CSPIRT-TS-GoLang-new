package storage_test

import (
	"reflect"
	"testing"
	"time"

	authModels "cspirt/internal/auth/models"
	authService "cspirt/internal/auth/service/auth"
	classModels "cspirt/internal/class/models"
	eventModels "cspirt/internal/events/models"
	ratingModels "cspirt/internal/rating/models"
	ratingService "cspirt/internal/rating/service"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"
	"cspirt/internal/utils"
)

const testSecret = "test-secret"

func TestAuthFeatureLoginAndRefresh(t *testing.T) {
	st := newTestStorage(t)
	owner := addTestUser(t, st, "owner", string(ratingModels.RoleOwner), "10A", 1000)

	auth := authService.NewAuthService(st, testSecret)

	badLogin, err := auth.Login(authModels.LoginInput{
		Login:    owner.Login,
		Password: "wrong-password",
	})
	if err != nil {
		t.Fatalf("login with wrong password returned error: %v", err)
	}
	if badLogin.Token != "" || badLogin.RefreshToken != "" {
		t.Fatalf("wrong password should not return tokens: %+v", badLogin)
	}

	result, err := auth.Login(authModels.LoginInput{
		Login:    owner.Login,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if result.Token == "" {
		t.Fatal("login did not return access token")
	}
	if result.RefreshToken == "" {
		t.Fatal("login did not return refresh token")
	}

	session, err := st.GetRefreshToken(result.RefreshToken)
	if err != nil {
		t.Fatalf("get refresh token returned error: %v", err)
	}
	if session == nil || session.UserID != owner.ID {
		t.Fatalf("refresh token was not persisted for owner: %+v", session)
	}

	refreshed, err := auth.Refresh(result.RefreshToken)
	if err != nil {
		t.Fatalf("refresh returned error: %v", err)
	}
	if refreshed.Token == "" {
		t.Fatal("refresh did not return new access token")
	}

	expiredToken := "expired-refresh-token"
	if err := st.SaveRefreshToken(owner.ID, expiredToken, time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("save expired refresh token returned error: %v", err)
	}
	if _, err := auth.Refresh(expiredToken); err == nil {
		t.Fatal("expired refresh token should return error")
	}
	deletedSession, err := st.GetRefreshToken(expiredToken)
	if err != nil {
		t.Fatalf("get expired refresh token returned error: %v", err)
	}
	if deletedSession != nil {
		t.Fatalf("expired refresh token should be deleted, got %+v", deletedSession)
	}
}

func TestUsersAndClassesFeature(t *testing.T) {
	st := newTestStorage(t)
	owner := addTestUser(t, st, "owner", string(ratingModels.RoleOwner), "10A", 1000)
	helper := addTestUser(t, st, "helper", string(ratingModels.RoleHelper), "10A", 500)
	student := addTestUser(t, st, "student", string(ratingModels.RoleUser), "11B", 300)

	users, err := st.GetAllUsers()
	if err != nil {
		t.Fatalf("get all users returned error: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}

	if owner.ClassID <= 0 || helper.ClassID != owner.ClassID || student.ClassID == owner.ClassID {
		t.Fatalf("class IDs were not assigned correctly: owner=%d helper=%d student=%d", owner.ClassID, helper.ClassID, student.ClassID)
	}

	classUsers, err := st.GetUsersByClassID(owner.ClassID)
	if err != nil {
		t.Fatalf("get users by class returned error: %v", err)
	}
	if len(classUsers) != 2 {
		t.Fatalf("expected 2 users in class 10A, got %d", len(classUsers))
	}

	class, err := st.GetClassByID(owner.ClassID)
	if err != nil {
		t.Fatalf("get class returned error: %v", err)
	}
	if class == nil {
		t.Fatal("class 10A not found")
	}
	if class.Name != "10A" {
		t.Fatalf("class name was not normalized: %q", class.Name)
	}
	if class.TotalRating != 750 {
		t.Fatalf("expected class average rating 750, got %d", class.TotalRating)
	}
	if len(class.Members) != 2 {
		t.Fatalf("expected class members to be synced, got %d", len(class.Members))
	}

	teacher, err := st.GetClassTeacherByID(owner.ClassID)
	if err != nil {
		t.Fatalf("get class teacher returned error: %v", err)
	}
	if teacher == nil || teacher.Login != owner.Login {
		t.Fatalf("expected owner to be auto-selected as teacher, got %+v", teacher)
	}

	if err := st.SaveClassTeacherByID(owner.ClassID, helper.Login); err != nil {
		t.Fatalf("save class teacher returned error: %v", err)
	}
	teacher, err = st.GetClassTeacherByID(owner.ClassID)
	if err != nil {
		t.Fatalf("get class teacher after update returned error: %v", err)
	}
	if teacher == nil || teacher.Login != helper.Login {
		t.Fatalf("expected helper teacher, got %+v", teacher)
	}

	updatedOwner := utils.UserToSafeUser(*owner)
	updatedOwner.Rating = 1200
	if err := st.SaveUser(*updatedOwner); err != nil {
		t.Fatalf("save user returned error: %v", err)
	}
	owner, err = st.GetUserByLogin(owner.Login)
	if err != nil {
		t.Fatalf("get owner after save returned error: %v", err)
	}
	if owner.Rating != 1200 {
		t.Fatalf("expected updated owner rating 1200, got %d", owner.Rating)
	}

	if err := st.AddClass(classModels.ClassInput{Name: "12c"}); err != nil {
		t.Fatalf("add class returned error: %v", err)
	}
	classes, err := st.GetAllClasses()
	if err != nil {
		t.Fatalf("get all classes returned error: %v", err)
	}
	if !hasClass(classes, "12C") {
		t.Fatalf("added class 12C not found: %+v", classes)
	}

	if err := st.DeleteUser(student.ID, *utils.UserToSafeUser(*owner)); err != nil {
		t.Fatalf("delete user returned error: %v", err)
	}
	deleted, err := st.GetUserByLogin(student.Login)
	if err != nil {
		t.Fatalf("get deleted user returned error: %v", err)
	}
	if deleted != nil {
		t.Fatalf("deleted user should not exist: %+v", deleted)
	}
}

func TestNotesFeature(t *testing.T) {
	st := newTestStorage(t)
	author := addTestUser(t, st, "owner", string(ratingModels.RoleOwner), "10A", 1000)
	target := addTestUser(t, st, "student", string(ratingModels.RoleUser), "10A", 500)
	authorSafe := *utils.UserToSafeUser(*author)

	if err := st.AddNote(author.Login, userModels.Note{
		TargetID:   target.ID,
		TargetName: target.Name + " " + target.LastName,
		AuthorID:   author.ID,
		AuthorName: author.Name + " " + author.LastName,
		Content:    "solid progress",
		CreatedAt:  time.Now(),
	}, authorSafe); err != nil {
		t.Fatalf("add note returned error: %v", err)
	}

	if err := st.AddNote(author.Login, userModels.Note{
		TargetID:  target.ID,
		AuthorID:  author.ID,
		Content:   "   ",
		CreatedAt: time.Now(),
	}, authorSafe); err == nil {
		t.Fatal("empty note content should return error")
	}

	allNotes, err := st.GetAllNotes()
	if err != nil {
		t.Fatalf("get all notes returned error: %v", err)
	}
	if len(allNotes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(allNotes))
	}

	userNotes, err := st.GetNotesByUserId(target.ID)
	if err != nil {
		t.Fatalf("get notes by user returned error: %v", err)
	}
	if len(userNotes) != 1 || userNotes[0].Content != "solid progress" {
		t.Fatalf("unexpected user notes: %+v", userNotes)
	}

	classNotes, err := st.GetNotesByClassID(target.ClassID)
	if err != nil {
		t.Fatalf("get notes by class returned error: %v", err)
	}
	if len(classNotes) != 1 {
		t.Fatalf("expected 1 class note, got %d", len(classNotes))
	}

	if err := st.DeleteNote(allNotes[0].ID, authorSafe); err != nil {
		t.Fatalf("delete note returned error: %v", err)
	}
	allNotes, err = st.GetAllNotes()
	if err != nil {
		t.Fatalf("get notes after delete returned error: %v", err)
	}
	if len(allNotes) != 0 {
		t.Fatalf("expected notes to be deleted, got %+v", allNotes)
	}
}

func TestComplaintsFeature(t *testing.T) {
	st := newTestStorage(t)
	author := addTestUser(t, st, "owner", string(ratingModels.RoleOwner), "10A", 1000)
	target := addTestUser(t, st, "student", string(ratingModels.RoleUser), "10A", 500)
	authorSafe := *utils.UserToSafeUser(*author)

	if err := st.AddComplaint(author.Login, userModels.Complaint{
		TargetID:   target.ID,
		TargetName: target.Name + " " + target.LastName,
		AuthorID:   author.ID,
		AuthorName: author.Name + " " + author.LastName,
		Content:    "needs follow-up",
		CreatedAt:  time.Now(),
	}, authorSafe); err != nil {
		t.Fatalf("add complaint returned error: %v", err)
	}

	if err := st.AddComplaint(author.Login, userModels.Complaint{
		TargetID:  target.ID,
		AuthorID:  author.ID,
		Content:   "",
		CreatedAt: time.Now(),
	}, authorSafe); err == nil {
		t.Fatal("empty complaint content should return error")
	}

	allComplaints, err := st.GetAllComplaints()
	if err != nil {
		t.Fatalf("get all complaints returned error: %v", err)
	}
	if len(allComplaints) != 1 {
		t.Fatalf("expected 1 complaint, got %d", len(allComplaints))
	}

	userComplaints, err := st.GetComplaintsByUserId(target.ID)
	if err != nil {
		t.Fatalf("get complaints by user returned error: %v", err)
	}
	if len(userComplaints) != 1 || userComplaints[0].Content != "needs follow-up" {
		t.Fatalf("unexpected user complaints: %+v", userComplaints)
	}

	classComplaints, err := st.GetComplaintsByClassID(target.ClassID)
	if err != nil {
		t.Fatalf("get complaints by class returned error: %v", err)
	}
	if len(classComplaints) != 1 {
		t.Fatalf("expected 1 class complaint, got %d", len(classComplaints))
	}

	if err := st.DeleteComplaint(allComplaints[0].ID, authorSafe); err != nil {
		t.Fatalf("delete complaint returned error: %v", err)
	}
	allComplaints, err = st.GetAllComplaints()
	if err != nil {
		t.Fatalf("get complaints after delete returned error: %v", err)
	}
	if len(allComplaints) != 0 {
		t.Fatalf("expected complaints to be deleted, got %+v", allComplaints)
	}
}

func TestRatingFeature(t *testing.T) {
	st := newTestStorage(t)
	owner := addTestUser(t, st, "owner", string(ratingModels.RoleOwner), "10A", 1000)
	target := addTestUser(t, st, "student", string(ratingModels.RoleUser), "10A", 500)

	ratings := ratingService.NewRatingsService(st.RatingRepo, st, testSecret)
	ownerSafe := utils.UserToSafeUser(*owner)

	if err := ratings.UpdateRating(owner.Login, &ratingModels.RatingInput{
		TargetLogin: target.Login,
		Rating:      250,
		Reason:      "helpful",
	}, ownerSafe); err != nil {
		t.Fatalf("update rating returned error: %v", err)
	}

	target, err := st.GetUserByLogin(target.Login)
	if err != nil {
		t.Fatalf("get target after rating returned error: %v", err)
	}
	if target.Rating != 750 {
		t.Fatalf("expected target rating 750, got %d", target.Rating)
	}

	if err := ratings.UpdateRating(owner.Login, &ratingModels.RatingInput{
		TargetLogin: target.Login,
		Rating:      5000,
		Reason:      "cap",
	}, ownerSafe); err != nil {
		t.Fatalf("update rating to cap returned error: %v", err)
	}
	target, err = st.GetUserByLogin(target.Login)
	if err != nil {
		t.Fatalf("get target after cap returned error: %v", err)
	}
	if target.Rating != 5000 {
		t.Fatalf("expected target rating to be clamped to 5000, got %d", target.Rating)
	}

	targetSafe := utils.UserToSafeUser(*target)
	if err := ratings.UpdateRating(target.Login, &ratingModels.RatingInput{
		TargetLogin: owner.Login,
		Rating:      100,
		Reason:      "not allowed",
	}, targetSafe); err == nil {
		t.Fatal("regular user should not update rating")
	}
}

func TestEventsFeature(t *testing.T) {
	st := newTestStorage(t)
	owner := addTestUser(t, st, "owner", string(ratingModels.RoleOwner), "10A", 1000)
	student := addTestUser(t, st, "student", string(ratingModels.RoleUser), "10A", 500)

	if err := st.AddEvent(eventModels.Event{
		Title:       "Tournament",
		Status:      "planned",
		Description: "Class tournament",
		CreatedAt:   time.Now().UTC().Truncate(time.Second),
		StartedAt:   "2026-05-03T10:00:00Z",
		Players:     []int{owner.ID},
	}); err != nil {
		t.Fatalf("add event returned error: %v", err)
	}

	events, err := st.GetEvents()
	if err != nil {
		t.Fatalf("get events returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !reflect.DeepEqual(events[0].Players, []int{owner.ID}) {
		t.Fatalf("unexpected event players: %+v", events[0].Players)
	}

	ownerEvents, err := st.GetEventsByUserID(owner.ID)
	if err != nil {
		t.Fatalf("get events by owner returned error: %v", err)
	}
	if len(ownerEvents) != 1 {
		t.Fatalf("expected owner to have 1 event, got %d", len(ownerEvents))
	}

	if err := st.AddPlayersToEvent(events[0].ID, []int{student.ID, owner.ID}); err != nil {
		t.Fatalf("add players to event returned error: %v", err)
	}
	studentEvents, err := st.GetEventsByUserID(student.ID)
	if err != nil {
		t.Fatalf("get events by student returned error: %v", err)
	}
	if len(studentEvents) != 1 {
		t.Fatalf("expected student to have 1 event, got %d", len(studentEvents))
	}
	if !reflect.DeepEqual(studentEvents[0].Players, []int{owner.ID, student.ID}) {
		t.Fatalf("unexpected players after add: %+v", studentEvents[0].Players)
	}

	if err := st.DeletePlayersFromEvent(events[0].ID, []int{owner.ID}); err != nil {
		t.Fatalf("delete players from event returned error: %v", err)
	}
	ownerEvents, err = st.GetEventsByUserID(owner.ID)
	if err != nil {
		t.Fatalf("get owner events after delete player returned error: %v", err)
	}
	if len(ownerEvents) != 0 {
		t.Fatalf("expected owner events to be empty after removal, got %+v", ownerEvents)
	}

	if err := st.DeleteEvent(events[0].ID); err != nil {
		t.Fatalf("delete event returned error: %v", err)
	}
	events, err = st.GetEvents()
	if err != nil {
		t.Fatalf("get events after delete returned error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected events to be deleted, got %+v", events)
	}
}

const testPassword = "secret123"

func newTestStorage(t *testing.T) *storage.Storage {
	t.Helper()

	st, err := storage.NewUserStorage(t.TempDir()+"/storage.db", testSecret)
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

func addTestUser(t *testing.T, st *storage.Storage, login string, role string, className string, rating int) *userModels.User {
	t.Helper()

	user := userModels.User{
		Name:     login,
		LastName: "Test",
		FullName: []userModels.FullName{{
			Name:     login,
			LastName: "Test",
		}},
		Login:    login,
		Password: testPassword,
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

func hasClass(classes []classModels.Class, name string) bool {
	for _, class := range classes {
		if class.Name == name {
			return true
		}
	}

	return false
}
