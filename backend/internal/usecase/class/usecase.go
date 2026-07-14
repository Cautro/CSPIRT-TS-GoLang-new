package usecase

import (
	classModels "cspirt/internal/domain/class"
	"cspirt/internal/domain/class/repo"
	userModels "cspirt/internal/domain/user"
	config "cspirt/internal/controller/http/class/config"
	"fmt"
)

type ClassUsecase struct { 
	classes repo.ClassRepository
}

func NewClassUsecase(classes repo.ClassRepository) *ClassUsecase {
	return &ClassUsecase{
		classes: classes,
	}
}

func (s *ClassUsecase) InitializeParallelsFromConfig(targetConfigs []config.ParallelConfig) error {
	existingParallels, err := s.GetParallelClasses()
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
	classes, err := s.classes.GetAllClasses()
	if err != nil {
		return nil, err
	}

	var ids []int
	for _, class := range classes {
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
	
	return s.classes.AddParallel(name, ids)
}

func (s *ClassUsecase) GetAllClassTeachers() ([]userModels.SafeUser, error) {
	return s.classes.GetAllClassTeachers()
} 

func (s *ClassUsecase) AddParallelClass(name string, classesIDs []int, login string) error {
	return s.classes.AddParallel(name, classesIDs)
}

func (s *ClassUsecase) GetParallelClasses() ([]classModels.ParallelClass, error) {
	parallelClasses, err := s.classes.GetParallelClasses()
	if err != nil {
		return nil, err
	}
	if parallelClasses == nil {
		return []classModels.ParallelClass{}, nil
	}

	return parallelClasses, nil
}

func (s *ClassUsecase) UpdateClass(classID int, input classModels.ClassInput, login string) error {
	return s.classes.UpdateClass(classID, input, login)
}

func (s *ClassUsecase) GetClassesInParallel(parallelID int) ([]classModels.Class, error) {
	return s.classes.GetClassesInParallel(parallelID)
}

func (s *ClassUsecase) GetBestClassInParallel(parallelID int) (*classModels.Class, error) {
	parallelClasses, err := s.classes.GetParallelClasses()
	if err != nil {
		return nil, err
	}
	for _, parallelClass := range parallelClasses {
		if parallelClass.ID == parallelID {
			return s.classes.GetClassByID(parallelClass.BestClassID)
		}
	}
	return nil, nil
}

func (s *ClassUsecase) YearComplete() ([]*classModels.Class, error) {
	return s.classes.YearComplete()
}

func (s *ClassUsecase) CompleteQuarter(parallelClassId int) ([]*classModels.Class, error) {
	return s.classes.QuarterComplete(parallelClassId)
}

func (s *ClassUsecase) DeleteParallelClass(parallelClassID int, login string) error {
	return s.classes.DeleteParallelClassByID(parallelClassID, login)
}

func (s *ClassUsecase) AddClass(input classModels.ClassInput, login string) error {
	return s.classes.AddClass(input, login)
}

func (s *ClassUsecase) DeleteClass(classID int, login string) error {
	return s.classes.DeleteClassByID(classID, login)
}

func (s *ClassUsecase) GetAllClasses() ([]classModels.Class, error) {
	classes, err := s.classes.GetAllClasses()
	if err != nil {
		return nil, err
	}
	if classes == nil {
		return []classModels.Class{}, nil
	}

	return classes, nil
}

func (s *ClassUsecase) GetClassByID(classID int) (*classModels.Class, error) {
	return s.classes.GetClassByID(classID)
}

func (s *ClassUsecase) GetUsersByClassID(classID int) ([]userModels.SafeUser, error) {
	users, err := s.classes.GetUsersByClassID(classID)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []userModels.SafeUser{}, nil
	}

	return users, nil
}

func (s *ClassUsecase) GetClassTeacher(classID int) (*userModels.SafeUser, error) {
	return s.classes.GetClassTeacherByID(classID)
}

func (s *ClassUsecase) SetClassTeacher(classID int, teacherLogin string) error {
	return s.classes.SaveClassTeacherByID(classID, teacherLogin)
}