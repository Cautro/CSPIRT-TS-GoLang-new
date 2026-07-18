package usecase

import (
	classRepo "cspirt/internal/domain/class/repo"
	complaintsRepo "cspirt/internal/domain/complaint/repo"
	eventsRepo "cspirt/internal/domain/event/repo"
	notesRepo "cspirt/internal/domain/note/repo"
	userRepo "cspirt/internal/domain/user/repo"
	models "cspirt/internal/domain/user"
	"cspirt/internal/controller/utils"
	eventDomain "cspirt/internal/domain/event"

	redis "cspirt/internal/domain/cache/repo"
	"cspirt/pkg/profiler"

	"context"
	"sync"
	"fmt"
	"time"
)

type UsersUsecase struct {
	userRepo       userRepo.UserRepository
	notesRepo      notesRepo.NoteRepository
	complaintsRepo complaintsRepo.ComplaintRepository
	classRepo      classRepo.ClassRepository
	eventsRepo     eventsRepo.EventsRepository

	redis           redis.CacheRepository
}

func NewUsersUsecase(
	uRepo userRepo.UserRepository,
	nRepo notesRepo.NoteRepository,
	cRepo complaintsRepo.ComplaintRepository,
	clRepo classRepo.ClassRepository,
	eRepo eventsRepo.EventsRepository,
	rRepo redis.CacheRepository,
) *UsersUsecase {
	return &UsersUsecase{
		userRepo:       uRepo,
		notesRepo:      nRepo,
		complaintsRepo: cRepo,
		classRepo:      clRepo,
		eventsRepo:     eRepo,
		redis:          rRepo,
	}
}

func (s *UsersUsecase) GetFullUserInfo(ctx context.Context, userID int) (models.UserWithFullInfo[eventDomain.Event], error) {
    emptyResponse := models.UserWithFullInfo[eventDomain.Event]{}

    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return emptyResponse, err
    }
    safeUser := utils.UserToSafeUser(*user)

    var wg sync.WaitGroup

    var (
        notes        []models.Note
        complaints   []models.Complaint
        classTeacher *models.SafeUser
        events       []eventDomain.Event
    )

    goroutineErrChan := make(chan error, 4)

    wg.Add(1)
    go func() {
        defer wg.Done()
        err := profiler.Track(ctx, "repo.notes", func() (e error) {
            notes, e = s.notesRepo.GetNotesByUserId(ctx, safeUser.ID)
            return
        })
        if err != nil {
            goroutineErrChan <- err
        }
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        err := profiler.Track(ctx, "repo.complaints", func() (e error) {
            complaints, e = s.complaintsRepo.GetComplaintsByUserId(ctx, safeUser.ID)
            return
        })
        if err != nil {
            goroutineErrChan <- err
        }
    }()

    if safeUser.ClassID > 0 {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := profiler.Track(ctx, "repo.classTeacher", func() (e error) {
                classTeacher, e = s.classRepo.GetClassTeacherByID(ctx, safeUser.ClassID)
                return
            })
            if err != nil {
                goroutineErrChan <- err
            }
        }()
    }

    wg.Add(1)
    go func() {
        defer wg.Done()
        err := profiler.Track(ctx, "repo.events", func() (e error) {
            events, e = s.eventsRepo.GetEventsByUserID(ctx, safeUser.ID)
            return
        })
        if err != nil {
            goroutineErrChan <- err
        }
    }()

    wg.Wait()
    close(goroutineErrChan)

    for err := range goroutineErrChan {
        return emptyResponse, err
    }

    return models.UserWithFullInfo[eventDomain.Event]{
        User:         safeUser,
        Notes:        notes,
        Complaints:   complaints,
        ClassTeacher: classTeacher,
        Events:       events,
    }, nil
}

func (s *UsersUsecase) GetFullUserInfoByLogin(ctx context.Context, login string) (models.UserWithFullInfo[eventDomain.Event], error) {
    key := userFullInfoKey(login)

    if s.redis != nil {
        var cached models.UserWithFullInfo[eventDomain.Event]
        err := profiler.Track(ctx, "cache.get", func() error {
            return s.redis.Get(ctx, key, &cached)
        })
        if err == nil {
            return cached, nil
        }
    }

    emptyResponse := models.UserWithFullInfo[eventDomain.Event]{}

    user, err := s.userRepo.GetUserByLogin(ctx, login)
    if err != nil {
        return emptyResponse, err
    }
    if user == nil {
        return emptyResponse, fmt.Errorf("user not found")
    }
    safeUser := utils.UserToSafeUser(*user)

    var wg sync.WaitGroup

    var (
        notes        []models.Note
        complaints   []models.Complaint
        classTeacher *models.SafeUser
        events       []eventDomain.Event
    )

    goroutineErrChan := make(chan error, 4)

    wg.Add(1)
    go func() {
        defer wg.Done()
        err := profiler.Track(ctx, "repo.notes", func() (e error) {
            notes, e = s.notesRepo.GetNotesByUserId(ctx, safeUser.ID)
            return
        })
        if err != nil {
            goroutineErrChan <- err
        }
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        err := profiler.Track(ctx, "repo.complaints", func() (e error) {
            complaints, e = s.complaintsRepo.GetComplaintsByUserId(ctx, safeUser.ID)
            return
        })
        if err != nil {
            goroutineErrChan <- err
        }
    }()

    if safeUser.ClassID > 0 {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := profiler.Track(ctx, "repo.classTeacher", func() (e error) {
                classTeacher, e = s.classRepo.GetClassTeacherByID(ctx, safeUser.ClassID)
                return
            })
            if err != nil {
                goroutineErrChan <- err
            }
        }()
    }

	wg.Add(1)
    go func() {
        defer wg.Done()
        err := profiler.Track(ctx, "repo.events", func() (e error) {
            events, e = s.eventsRepo.GetEventsByUserID(ctx, safeUser.ID)
            return
        })
        if err != nil {
            goroutineErrChan <- err
        }
    }()

    wg.Wait()
    close(goroutineErrChan)

    if len(goroutineErrChan) > 0 {
        return emptyResponse, <-goroutineErrChan
    }

    output := models.UserWithFullInfo[eventDomain.Event]{
        User:         safeUser,
        Notes:        notes,
        Complaints:   complaints,
        ClassTeacher: classTeacher,
        Events:       events,
    }

    if s.redis != nil {
        _ = s.redis.Set(ctx, key, output, 10*time.Minute)
    }

    return output, nil
}

// userFullInfoKey is the Redis key for a user's /me aggregate. Kept in one place
// so reads and invalidation (InvalidateUserFullInfo) never drift apart.
func userFullInfoKey(login string) string {
    return fmt.Sprintf("user:full:%s", login)
}

// InvalidateUserFullInfo drops the cached /me aggregate for login. Call it after
// any mutation that changes what /me returns (profile update, avatar, delete,
// logout). No-op when caching is disabled.
func (s *UsersUsecase) InvalidateUserFullInfo(ctx context.Context, login string) {
    if s.redis == nil || login == "" {
        return
    }
    _ = s.redis.Delete(ctx, userFullInfoKey(login))
}

// invalidateUserByID resolves the user's login and drops its /me cache. Used by
// mutations that only have the numeric id (e.g. avatar update).
func (s *UsersUsecase) invalidateUserByID(ctx context.Context, id int) {
    if s.redis == nil || id <= 0 {
        return
    }
    if u, err := s.userRepo.GetUserByID(ctx, id); err == nil && u != nil {
        s.InvalidateUserFullInfo(ctx, u.Login)
    }
}

func (s *UsersUsecase) GetUserByLogin(ctx context.Context, 	login string) (*models.User, error) {
	output, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *UsersUsecase) GetUsersHandlerService(ctx context.Context) ([]models.SafeUser, error) {
	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UsersUsecase) GetUsersByClassIDHandlerService(ctx context.Context, classID int) ([]models.SafeUser, error) {
	users, err := s.userRepo.GetUsersByClassID(ctx, classID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UsersUsecase) UpdateAvatar(ctx context.Context, input models.UpdateAvatarRequest, id int) error {
	err := s.userRepo.UpdateAvatar(ctx, input, id)
	if err != nil {
		return err
	}
	s.invalidateUserByID(ctx, id)
	return nil
}

func (s *UsersUsecase) UpdateUserHandlerService(ctx context.Context, userID int, user models.SafeUser, login string) (error) {
	// Capture the target's current login before the update so we can drop its
	// stale /me cache even if the update changes the login.
	var oldLogin string
	if existing, err := s.userRepo.GetUserByID(ctx, userID); err == nil && existing != nil {
		oldLogin = existing.Login
	}

	NeedUser := models.UpdateUserRequest{
		Name: &user.Name,
		LastName: &user.LastName,
		Avatar: &user.Avatar,
		FullName: &user.FullName,
		Login: &user.Login,
		Rating: &user.Rating,
		Role: &user.Role,
		Class: &user.Class,
		ClassID: &user.ClassID,
	}
	err := s.userRepo.UpdateUser(ctx, userID, NeedUser, login)
	if err != nil {
		return err
	}
	s.InvalidateUserFullInfo(ctx, oldLogin)
	s.InvalidateUserFullInfo(ctx, user.Login)
	return nil
}

func (s *UsersUsecase) DeleteRefreshToken(ctx context.Context, token string) error {
	return s.userRepo.DeleteRefreshToken(ctx, token)
}

func (s *UsersUsecase) GetOnlyStaffUsers(ctx context.Context) ([]models.SafeUser, error) {
	return s.userRepo.GetOnlyStaffUsers(ctx)
}