package repo

import "cspirt/internal/models"

type ClassRepository interface {
	EnsureClass(name string) error
	SaveClassTeacherByID(classID int, teacherLogin string) error
	GetAllClasses() ([]models.Class, error)
	GetClassByID(id int) (*models.Class, error)
	GetClassTeacherByID(classID int) (*models.SafeUser, error)
	GetUsersByClassID(classID int) ([]models.SafeUser, error)
}
