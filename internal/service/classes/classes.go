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

func (s *ClassService) GetUsersByClass(name string) ([]models.SafeUser, error) {
	users, err := s.classes.GetUsersByClass(name)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []models.SafeUser{}, nil
	}

	return users, nil
}

func (s *ClassService) GetClassTeacher(name string) (*models.SafeUser, error) {
	return s.classes.GetClassTeacher(name)
}

func (s *ClassService) SetClassTeacher(name string, teacherLogin string) error {
	return s.classes.SaveClassTeacher(name, teacherLogin)
}
