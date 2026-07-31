package repo

import (
	"context"
	perm "cspirt/internal/controller/permission/usecase"
)

type RatingRepository interface {
	UpdateRating(login string, rating int) error
	UpdateClassRating(ctx context.Context, classId, userId, delta int, perm perm.Usecase) error
}