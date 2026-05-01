package repo

import "cspirt/internal/models"

type ClassRepository interface {
	EnsureClass(name string) error
	SaveClassTeacher(name string, teacherLogin string) error
	GetAllClasses() ([]models.Class, error)
	GetClassByName(name string) (*models.Class, error)
	GetClassTeacher(name string) (*models.SafeUser, error)
	GetUsersByClass(name string) ([]models.SafeUser, error)
}
