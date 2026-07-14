package repo

import (
	classConfig "cspirt/internal/controller/http/class/config"
	models "cspirt/internal/domain/user"
	entity "cspirt/internal/domain/class"
)

type ClassRepository interface {
	EnsureClass(name string) error
	SaveClassTeacherByID(classID int, teacherLogin string) error
	GetAllClasses() ([]entity.Class, error)
	GetClassByID(id int) (*entity.Class, error)
	GetClassTeacherByID(classID int) (*models.SafeUser, error)
	GetUsersByClassID(classID int) ([]models.SafeUser, error)
	DeleteClassByID(classID int, login string) error
	GetAllClassTeachers() ([]models.SafeUser, error) 
	AddClass(input entity.ClassInput, login string) error
	UpdateClass(classID int, input entity.ClassInput, login string) error
	YearComplete() ([]*entity.Class, error)

	// Parallel classes methods
	AddParallel(name string, classesIDs []int) error
	GetParallelClasses() ([]entity.ParallelClass, error)
	DeleteParallelClassByID(parallelClassID int, login string) error
	QuarterComplete(parallelClassID int) ([]*entity.Class, error)
	GetClassesInParallel(id int) ([]entity.Class, error)
	GetClassIDsByRange(minGrade, maxGrade int) ([]int, error)
	AddParallelByGradeRange(name string, minGrade, maxGrade int) error
	InitializeParallelsFromConfig(parallels []classConfig.ParallelConfig) error
}
