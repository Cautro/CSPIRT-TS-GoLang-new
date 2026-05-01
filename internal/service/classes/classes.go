package classes

import (
	"cspirt/internal/models"
	"cspirt/internal/repo"
)

type ClassService struct {
	classes repo.ClassRepository
}

func NewClassService(classes repo.ClassRepository, jwtSecret string) *ClassService {
	return &ClassService{
		classes: classes,
	}
}

func (s *ClassService) GetAllClasses() ([]models.Class, error) {
	classes, err := s.classes.GetAllClasses()
	if err != nil {
		return nil, err
	}
	if classes == nil {
		return []models.Class{}, nil
	}

	return classes, nil
}

func (s *ClassService) GetClassByID(classID int) (*models.Class, error) {
	return s.classes.GetClassByID(classID)
}

func (s *ClassService) GetUsersByClassID(classID int) ([]models.SafeUser, error) {
	users, err := s.classes.GetUsersByClassID(classID)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []models.SafeUser{}, nil
	}

	return users, nil
}

func (s *ClassService) GetClassTeacher(classID int) (*models.SafeUser, error) {
	return s.classes.GetClassTeacherByID(classID)
}

func (s *ClassService) SetClassTeacher(classID int, teacherLogin string) error {
	return s.classes.SaveClassTeacherByID(classID, teacherLogin)
}
