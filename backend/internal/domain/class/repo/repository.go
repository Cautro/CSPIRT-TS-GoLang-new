package repo

import (
	classConfig "cspirt/internal/controller/http/class/config"
	models "cspirt/internal/domain/user"
	entity "cspirt/internal/domain/class"
	"context"
)

type ClassRepository interface {
	EnsureClass(ctx context.Context, name string) error
	SaveClassTeacherByID(ctx context.Context, classID int, teacherLogin string) error
	GetAllClasses(ctx context.Context) ([]entity.Class, error)
	GetClassByID(ctx context.Context, id int) (*entity.Class, error)
	GetClassTeacherByID(ctx context.Context, classID int) (*models.SafeUser, error)
	GetUsersByClassID(ctx context.Context, classID int) ([]models.SafeUser, error)
	DeleteClassByID(ctx context.Context, classID int, login string) error
	GetAllClassTeachers(ctx context.Context) ([]models.SafeUser, error) 
	AddClass(ctx context.Context, input entity.ClassInput, login string) error
	UpdateClass(ctx context.Context, classID int, input entity.ClassInput, login string) error
	YearComplete(ctx context.Context) ([]*entity.Class, error)

	// Parallel classes methods
	AddParallel(ctx context.Context, name string, classesIDs []int) error
	GetParallelClasses(ctx context.Context) ([]entity.ParallelClass, error)
	DeleteParallelClassByID(ctx context.Context, parallelClassID int, login string) error
	QuarterComplete(ctx context.Context, parallelClassID int) ([]*entity.Class, error)
	GetClassesInParallel(ctx context.Context, id int) ([]entity.Class, error)
	GetClassIDsByRange(ctx context.Context, minGrade, maxGrade int) ([]int, error)
	AddParallelByGradeRange(ctx context.Context, name string, minGrade, maxGrade int) error
	InitializeParallelsFromConfig(ctx context.Context, parallels []classConfig.ParallelConfig) error
}
