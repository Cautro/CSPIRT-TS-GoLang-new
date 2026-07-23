package usecase

import (
	"context"
	config "cspirt/internal/controller/http/class/config"
	permission "cspirt/internal/controller/permission/usecase"
	classModels "cspirt/internal/domain/class"
	"cspirt/internal/domain/class/repo"
	repoNotification "cspirt/internal/domain/notification/repo"
	userModels "cspirt/internal/domain/user"
	userRepo "cspirt/internal/domain/user/repo"
	logger "cspirt/pkg/logger"
	"errors"
	"fmt"
	//"strconv"
)

type ClassUsecase struct { 
	classRepo repo.ClassRepository
	userRepo  userRepo.UserRepository
	notifService  repoNotification.NotificationService
}

func NewClassUsecase(classRepo repo.ClassRepository, user userRepo.UserRepository, notif repoNotification.NotificationService) *ClassUsecase {
	return &ClassUsecase{
		classRepo: classRepo,
		userRepo: user,
		notifService: notif,
	}
}

func (s *ClassUsecase) InitializeParallelsFromConfig(ctx context.Context, targetConfigs []config.ParallelConfig) error {
	existingParallels, err := s.GetParallelClass(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch existing parallels: %w", err)
	}

	existingMap := make(map[string]bool)
	for _, p := range existingParallels {
		existingMap[p.Name] = true
	}

	for _, pc := range targetConfigs {
		if existingMap[pc.Name] {
			continue
		}

		classIDs, err := s.GetClassIDsByRange(ctx, pc.MinGrade, pc.MaxGrade)
		if err != nil {
			return fmt.Errorf("failed to get class IDs for range %d-%d: %w", pc.MinGrade, pc.MaxGrade, err)
		}

		if len(classIDs) == 0 {
			continue
		}

		err = s.AddParallelClass(ctx, pc.Name, classIDs, "system")
		if err != nil {
			return fmt.Errorf("failed to auto-create parallel %s: %w", pc.Name, err)
		}
	}

	return nil
}

func (s *ClassUsecase) GetClassIDsByRange(ctx context.Context, minGrade, maxGrade int) ([]int, error) {
	classRepo, err := s.classRepo.GetAllClasses(ctx)
	if err != nil {
		return nil, err
	}

	var ids []int
	for _, class := range classRepo {
		if class.Grade >= minGrade && class.Grade <= maxGrade {
			ids = append(ids, class.ID)
		}
	}
	return ids, nil
}

func (s *ClassUsecase) AddParallelByGradeRange(ctx context.Context, name string, minGrade, maxGrade int) error {
	ids, err := s.GetClassIDsByRange(ctx, minGrade, maxGrade)
	if err != nil {
		return err
	}
	
	return s.classRepo.AddParallel(ctx, name, ids)
}

func (s *ClassUsecase) GetAllClassTeachers(ctx context.Context) ([]userModels.SafeUser, error) {
	return s.classRepo.GetAllClassTeachers(ctx)
} 

func (s *ClassUsecase) AddParallelClass(ctx context.Context, name string, classRepoIDs []int, login string) error {
	user, err := s.userRepo.GetUserByLogin(ctx, login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }

	return s.classRepo.AddParallel(ctx, name, classRepoIDs)
}

func (s *ClassUsecase) GetParallelClass(ctx context.Context) ([]classModels.ParallelClass, error) {
	parallelclassRepo, err := s.classRepo.GetParallelClasses(ctx)
	if err != nil {
		return nil, err
	}
	if parallelclassRepo == nil {
		return []classModels.ParallelClass{}, nil
	}

	return parallelclassRepo, nil
}

func (s *ClassUsecase) UpdateClass(ctx context.Context, classID int, input classModels.ClassInput, login string) error {
	user, err := s.userRepo.GetUserByLogin(ctx, login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }
	
	return s.classRepo.UpdateClass(ctx, classID, input, login)
}

func (s *ClassUsecase) GetClassInParallel(ctx context.Context, parallelID int) ([]classModels.Class, error) {
	return s.classRepo.GetClassesInParallel(ctx, parallelID)
}

func (s *ClassUsecase) GetBestClassInParallel(ctx context.Context, parallelID int) (*classModels.Class, error) {
	parallelclassRepo, err := s.classRepo.GetParallelClasses(ctx)
	if err != nil {
		return nil, err
	}
	for _, parallelClass := range parallelclassRepo {
		if parallelClass.ID == parallelID {
			return s.classRepo.GetClassByID(ctx, parallelClass.BestClassID)
		}
	}
	return nil, nil
}

func (s *ClassUsecase) YearComplete(ctx context.Context, login string) ([]*classModels.Class, error) {
    user, err := s.userRepo.GetUserByLogin(ctx, login)
    if err != nil { 
        return []*classModels.Class{}, err 
    }

    if !permission.CanManageClasses(user.Role) { 
        return []*classModels.Class{}, errors.New("permission denied") 
    }

    classes, err := s.classRepo.YearComplete(ctx)
    if err != nil {
        return []*classModels.Class{}, err 
    }

    users, err := s.userRepo.GetAllUsers(ctx)
    if err != nil { 
        logger.WriteSafe(logger.LogEntry{
            Level:   "error",
            Action:  "get_users_for_notifications",
            Message: "Year completed in DB, but failed to fetch users for pushes: " + err.Error(),
        })
        return classes, nil 
    }

    go func(usersList []userModels.SafeUser) {
        bgCtx := context.Background()
        
        for _, u := range usersList {
            if err := s.notifService.Send(bgCtx, int64(u.ID), "Год был завершен", "Заходи, чтобы узнать итоги года!"); err != nil {
                logger.WriteSafe(logger.LogEntry{
                    Level:   "error",
                    Action:  "send_push_notification_bulk",
                    Message: fmt.Sprintf("Failed to send FCM to user %d: %s", u.ID, err.Error()),
                })
            }
        }
    }(users)

    return classes, nil
}

func (s *ClassUsecase) CompleteQuarter(ctx context.Context, parallelClassId int, login string) ([]*classModels.Class, error) {
	user, err := s.userRepo.GetUserByLogin(ctx, login); if err != nil { return []*classModels.Class{}, err }
	check := permission.CanManageClasses(user.Role); if !check { return []*classModels.Class{}, errors.New("permission denied") }

	classes, err := s.classRepo.QuarterComplete(ctx, parallelClassId); if err != nil { return []*classModels.Class{}, err }
	
	users, err := s.userRepo.GetAllUsers(ctx)
    if err != nil { 
        logger.WriteSafe(logger.LogEntry{
            Level:   "error",
            Action:  "get_users_for_notifications",
            Message: "Year completed in DB, but failed to fetch users for pushes: " + err.Error(),
        })
        return classes, nil 
    }

    go func(usersList []userModels.SafeUser) {
        bgCtx := context.Background()
        
        for _, u := range usersList {
            if err := s.notifService.Send(bgCtx, int64(u.ID), "Четверть была завершена", "Заходи, чтобы узнать итоги!"); err != nil {
                logger.WriteSafe(logger.LogEntry{
                    Level:   "error",
                    Action:  "send_push_notification_bulk",
                    Message: fmt.Sprintf("Failed to send FCM to user %d: %s", u.ID, err.Error()),
                })
            }
        }
    }(users)

	return classes, nil
}

func (s *ClassUsecase) DeleteParallelClass(ctx context.Context, parallelClassID int, login string) error {
	user, err := s.userRepo.GetUserByLogin(ctx, login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }


	return s.classRepo.DeleteParallelClassByID(ctx, parallelClassID, login)
}

func (s *ClassUsecase) AddClass(ctx context.Context, input classModels.ClassInput, login string) error {
	user, err := s.userRepo.GetUserByLogin(ctx, login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }

	return s.classRepo.AddClass(ctx, input, login)
}

func (s *ClassUsecase) DeleteClass(ctx context.Context, classID int, login string) error {
	return s.classRepo.DeleteClassByID(ctx, classID, login)
}

func (s *ClassUsecase) GetAllClasses(ctx context.Context) ([]classModels.Class, error) {
	classRepo, err := s.classRepo.GetAllClasses(ctx)
	if err != nil {
		return nil, err
	}
	if classRepo == nil {
		return []classModels.Class{}, nil
	}

	return classRepo, nil
}

func (s *ClassUsecase) GetClassByID(ctx context.Context, classID int) (*classModels.Class, error) {
	return s.classRepo.GetClassByID(ctx, classID)
}

func (s *ClassUsecase) GetUsersByClassID(ctx context.Context, classID int) ([]userModels.SafeUser, error) {
	users, err := s.classRepo.GetUsersByClassID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []userModels.SafeUser{}, nil
	}

	return users, nil
}

func (s *ClassUsecase) GetClassTeacher(ctx context.Context, classID int) (*userModels.SafeUser, error) {
	return s.classRepo.GetClassTeacherByID(ctx, classID)
}

func (s *ClassUsecase) SetClassTeacher(ctx context.Context, classID int, teacherLogin string) error {
	return s.classRepo.SaveClassTeacherByID(ctx, classID, teacherLogin)
}