package repo


type RatingRepository interface {
	UpdateRating(login string, rating int) error
}