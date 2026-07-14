package usecase

import (
	classRepo "cspirt/internal/domain/class/repo"
	complaintsRepo "cspirt/internal/domain/complaint/repo"
	eventsRepo "cspirt/internal/domain/event/repo"
	notesRepo "cspirt/internal/domain/note/repo"
	userRepo "cspirt/internal/domain/user/repo"
	models "cspirt/internal/domain/user"
	"cspirt/internal/utils"
	eventDomain "cspirt/internal/domain/event"
)

type UsersUsecase struct {
	userRepo       userRepo.UserRepository
	notesRepo      notesRepo.NoteRepository
	complaintsRepo complaintsRepo.ComplaintRepository
	classRepo      classRepo.ClassRepository
	eventsRepo     eventsRepo.EventsRepository
}

func NewUsersUsecase(
	uRepo userRepo.UserRepository,
	nRepo notesRepo.NoteRepository,
	cRepo complaintsRepo.ComplaintRepository,
	clRepo classRepo.ClassRepository,
	eRepo eventsRepo.EventsRepository,
) *UsersUsecase {
	return &UsersUsecase{
		userRepo:       uRepo,
		notesRepo:      nRepo,
		complaintsRepo: cRepo,
		classRepo:      clRepo,
		eventsRepo:     eRepo,
	}
}

func (s *UsersUsecase) GetFullUserInfo(userID int) (models.UserWithFullInfo[eventDomain.Event], error) {
	emptyResponse := models.UserWithFullInfo[eventDomain.Event]{}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return emptyResponse, err
	}
	safeUser := utils.UserToSafeUser(*user)

	notes, err := s.notesRepo.GetNotesByUserId(safeUser.ID)
	if err != nil { return emptyResponse, err }

	complaints, err := s.complaintsRepo.GetComplaintsByUserId(safeUser.ID)
	if err != nil { return emptyResponse, err }

	var classTeacher *models.SafeUser
	if safeUser.ClassID > 0 {
		classTeacher, err = s.classRepo.GetClassTeacherByID(safeUser.ClassID)
		if err != nil { return emptyResponse, err }
	}

	events, err := s.eventsRepo.GetEventsByUserID(safeUser.ID)
	if err != nil { return emptyResponse, err }

	return models.UserWithFullInfo[eventDomain.Event]{
		User:         safeUser,
		Notes:        notes,
		Complaints:   complaints,
		ClassTeacher: classTeacher,
		Events:       events,
	}, nil
}

func (s *UsersUsecase) GetFullUserInfoByLogin(login string) (models.UserWithFullInfo[eventDomain.Event], error) {
	emptyResponse := models.UserWithFullInfo[eventDomain.Event]{}

	user, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		return emptyResponse, err
	}
	safeUser := utils.UserToSafeUser(*user)

	notes, err := s.notesRepo.GetNotesByUserId(safeUser.ID)
	if err != nil { return emptyResponse, err }

	complaints, err := s.complaintsRepo.GetComplaintsByUserId(safeUser.ID)
	if err != nil { return emptyResponse, err }

	var classTeacher *models.SafeUser
	if safeUser.ClassID > 0 {
		classTeacher, err = s.classRepo.GetClassTeacherByID(safeUser.ClassID)
		if err != nil { return emptyResponse, err }
	}

	events, err := s.eventsRepo.GetEventsByUserID(safeUser.ID)
	if err != nil { return emptyResponse, err }

	return models.UserWithFullInfo[eventDomain.Event]{
		User:         safeUser,
		Notes:        notes,
		Complaints:   complaints,
		ClassTeacher: classTeacher,
		Events:       events,
	}, nil
}

func (s *UsersUsecase) GetUserByLogin(login string) (*models.User, error) {
	output, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *UsersUsecase) GetUsersHandlerService() ([]models.SafeUser, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UsersUsecase) GetUsersByClassIDHandlerService(classID int) ([]models.SafeUser, error) {
	users, err := s.userRepo.GetUsersByClassID(classID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UsersUsecase) UpdateAvatar(input models.UpdateAvatarRequest, id int) error {
	err := s.userRepo.UpdateAvatar(input, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersUsecase) UpdateUserHandlerService(userID int, user models.SafeUser, login string) (error) {
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
	err := s.userRepo.UpdateUser(userID, NeedUser, login)
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersUsecase) DeleteRefreshToken(token string) error {
	return s.userRepo.DeleteRefreshToken(token)
}

func (s *UsersUsecase) GetOnlyStaffUsers() ([]models.SafeUser, error) {
	return s.userRepo.GetOnlyStaffUsers()
}