package usecase

import (
	config "cspirt/internal/controller/http/class/config"
	permission "cspirt/internal/controller/permission/usecase"
	classModels "cspirt/internal/domain/class"
	"cspirt/internal/domain/class/repo"
	userModels "cspirt/internal/domain/user"
	userRepo "cspirt/internal/domain/user/repo"
	"fmt"
)

type ClassUsecase struct { 
	classRepo repo.ClassRepository
	userRepo  userRepo.UserRepository
}

func NewClassUsecase(classRepo repo.ClassRepository, user userRepo.UserRepository) *ClassUsecase {
	return &ClassUsecase{
		classRepo: classRepo,
		userRepo: user,
	}
}

func (s *ClassUsecase) InitializeParallelsFromConfig(targetConfigs []config.ParallelConfig) error {
	existingParallels, err := s.GetParallelClass()
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

		classIDs, err := s.GetClassIDsByRange(pc.MinGrade, pc.MaxGrade)
		if err != nil {
			return fmt.Errorf("failed to get class IDs for range %d-%d: %w", pc.MinGrade, pc.MaxGrade, err)
		}

		if len(classIDs) == 0 {
			continue
		}

		err = s.AddParallelClass(pc.Name, classIDs, "system")
		if err != nil {
			return fmt.Errorf("failed to auto-create parallel %s: %w", pc.Name, err)
		}
	}

	return nil
}

func (s *ClassUsecase) GetClassIDsByRange(minGrade, maxGrade int) ([]int, error) {
	classRepo, err := s.classRepo.GetAllClasses()
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

func (s *ClassUsecase) AddParallelByGradeRange(name string, minGrade, maxGrade int) error {
	ids, err := s.GetClassIDsByRange(minGrade, maxGrade)
	if err != nil {
		return err
	}
	
	return s.classRepo.AddParallel(name, ids)
}

func (s *ClassUsecase) GetAllClassTeachers() ([]userModels.SafeUser, error) {
	return s.classRepo.GetAllClassTeachers()
} 

func (s *ClassUsecase) AddParallelClass(name string, classRepoIDs []int, login string) error {
	user, err := s.userRepo.GetUserByLogin(login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }

	return s.classRepo.AddParallel(name, classRepoIDs)
}

func (s *ClassUsecase) GetParallelClass() ([]classModels.ParallelClass, error) {
	parallelclassRepo, err := s.classRepo.GetParallelClasses()
	if err != nil {
		return nil, err
	}
	if parallelclassRepo == nil {
		return []classModels.ParallelClass{}, nil
	}

	return parallelclassRepo, nil
}

func (s *ClassUsecase) UpdateClass(classID int, input classModels.ClassInput, login string) error {
	user, err := s.userRepo.GetUserByLogin(login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }
	
	return s.classRepo.UpdateClass(classID, input, login)
}

func (s *ClassUsecase) GetClassInParallel(parallelID int) ([]classModels.Class, error) {
	return s.classRepo.GetClassesInParallel(parallelID)
}

func (s *ClassUsecase) GetBestClassInParallel(parallelID int) (*classModels.Class, error) {
	parallelclassRepo, err := s.classRepo.GetParallelClasses()
	if err != nil {
		return nil, err
	}
	for _, parallelClass := range parallelclassRepo {
		if parallelClass.ID == parallelID {
			return s.classRepo.GetClassByID(parallelClass.BestClassID)
		}
	}
	return nil, nil
}

func (s *ClassUsecase) YearComplete(login string) ([]*classModels.Class, error) {
	user, err := s.userRepo.GetUserByLogin(login); if err != nil { return []*classModels.Class{}, err }
	check := permission.CanManageClasses(user.Role); if !check { return []*classModels.Class{}, err }

	return s.classRepo.YearComplete()
}

func (s *ClassUsecase) CompleteQuarter(parallelClassId int, login string) ([]*classModels.Class, error) {
	user, err := s.userRepo.GetUserByLogin(login); if err != nil { return []*classModels.Class{}, err }
	check := permission.CanManageClasses(user.Role); if !check { return []*classModels.Class{}, err }

	return s.classRepo.QuarterComplete(parallelClassId)
}

func (s *ClassUsecase) DeleteParallelClass(parallelClassID int, login string) error {
	user, err := s.userRepo.GetUserByLogin(login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }


	return s.classRepo.DeleteParallelClassByID(parallelClassID, login)
}

func (s *ClassUsecase) AddClass(input classModels.ClassInput, login string) error {
	user, err := s.userRepo.GetUserByLogin(login); if err != nil { return err }
	check := permission.CanManageClasses(user.Role); if !check { return err }

	return s.classRepo.AddClass(input, login)
}

func (s *ClassUsecase) DeleteClass(classID int, login string) error {
	return s.classRepo.DeleteClassByID(classID, login)
}

func (s *ClassUsecase) GetAllClasses() ([]classModels.Class, error) {
	classRepo, err := s.classRepo.GetAllClasses()
	if err != nil {
		return nil, err
	}
	if classRepo == nil {
		return []classModels.Class{}, nil
	}

	return classRepo, nil
}

func (s *ClassUsecase) GetClassByID(classID int) (*classModels.Class, error) {
	return s.classRepo.GetClassByID(classID)
}

func (s *ClassUsecase) GetUsersByClassID(classID int) ([]userModels.SafeUser, error) {
	users, err := s.classRepo.GetUsersByClassID(classID)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []userModels.SafeUser{}, nil
	}

	return users, nil
}

func (s *ClassUsecase) GetClassTeacher(classID int) (*userModels.SafeUser, error) {
	return s.classRepo.GetClassTeacherByID(classID)
}

func (s *ClassUsecase) SetClassTeacher(classID int, teacherLogin string) error {
	return s.classRepo.SaveClassTeacherByID(classID, teacherLogin)
}