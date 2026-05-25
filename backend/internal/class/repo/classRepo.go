package repo

import (
	"cspirt/internal/users/models"
	classModels "cspirt/internal/class/models"
)

type ClassRepository interface {
	EnsureClass(name string) error
	SaveClassTeacherByID(classID int, teacherLogin string) error
	GetAllClasses() ([]classModels.Class, error)
	GetClassByID(id int) (*classModels.Class, error)
	GetClassTeacherByID(classID int) (*models.SafeUser, error)
	GetUsersByClassID(classID int) ([]models.SafeUser, error) 
	DeleteClassByID(classID int, login string) error
	GetAllClassTeachers() ([]models.SafeUser, error)
	AddClass(input classModels.ClassInput, login string) error

	// Parallel classes methods
	AddParallel(name string, classesIDs []int) error
	GetParallelClasses() ([]classModels.ParallelClass, error)
	DeleteParallelClassByID(parallelClassID int, login string) error
	QuarterComplete(parallelClassID int) ([]*classModels.Class, error)
}
