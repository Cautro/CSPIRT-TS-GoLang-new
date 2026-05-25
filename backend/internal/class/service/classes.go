package classes

import (
	classModels "cspirt/internal/class/models"
	"cspirt/internal/class/repo"
	userModels "cspirt/internal/users/models"
)

type ClassService struct { 
	classes repo.ClassRepository
}

func NewClassService(classes repo.ClassRepository, jwtSecret string) *ClassService {
	return &ClassService{
		classes: classes,
	}
}

func (s *ClassService) GetAllClassTeachers() ([]userModels.SafeUser, error) {
	return s.classes.GetAllClassTeachers()
} 

func (s *ClassService) AddParallelClass(name string, classesIDs []int, login string) error {
	return s.classes.AddParallel(name, classesIDs)
}

func (s *ClassService) GetParallelClasses() ([]classModels.ParallelClass, error) {
	parallelClasses, err := s.classes.GetParallelClasses()
	if err != nil {
		return nil, err
	}
	if parallelClasses == nil {
		return []classModels.ParallelClass{}, nil
	}

	return parallelClasses, nil
}

func (s *ClassService) GetBestClassInParallel(parallelID int) (*classModels.Class, error) {
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

func (s *ClassService) CompleteQuarter(parallelClassId int) ([]*classModels.Class, error) {
	return s.classes.QuarterComplete(parallelClassId)
}

func (s *ClassService) DeleteParallelClass(parallelClassID int, login string) error {
	return s.classes.DeleteParallelClassByID(parallelClassID, login)
}

func (s *ClassService) AddClass(input classModels.ClassInput, login string) error {
	return s.classes.AddClass(input, login)
}

func (s *ClassService) DeleteClass(classID int, login string) error {
	return s.classes.DeleteClassByID(classID, login)
}

func (s *ClassService) GetAllClasses() ([]classModels.Class, error) {
	classes, err := s.classes.GetAllClasses()
	if err != nil {
		return nil, err
	}
	if classes == nil {
		return []classModels.Class{}, nil
	}

	return classes, nil
}

func (s *ClassService) GetClassByID(classID int) (*classModels.Class, error) {
	return s.classes.GetClassByID(classID)
}

func (s *ClassService) GetUsersByClassID(classID int) ([]userModels.SafeUser, error) {
	users, err := s.classes.GetUsersByClassID(classID)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []userModels.SafeUser{}, nil
	}

	return users, nil
}

func (s *ClassService) GetClassTeacher(classID int) (*userModels.SafeUser, error) {
	return s.classes.GetClassTeacherByID(classID)
}

func (s *ClassService) SetClassTeacher(classID int, teacherLogin string) error {
	return s.classes.SaveClassTeacherByID(classID, teacherLogin)
}