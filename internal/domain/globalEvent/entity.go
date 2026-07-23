package globalevent

// import (
// 	"github.com/google/uuid"
// )

type QuizOption struct {
	ID    int `json:"id"`
	Title string `json:"title"`
	Votes int    `json:"votes"`
}

type GlobalEventQuizEntity struct {
	ID          int    `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Options     []QuizOption `json:"options"` 
}

type GlobalEventInfoEntity struct {
	ID          int `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type QuizOptionDTO struct {
    Title       string   `json:"title" validate:"required"`
	Votes       int      `json:"votes" validate:"required"`
}

type GlobalEventQuizDTO struct {
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Option      []QuizOptionDTO `json:"option" validate:"required"`
}

type GlobalEventInfoDTO struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description" validate:"required"`
}

type VoteToPutinDTO struct {
	VoteItemId int `json:"voteItemId"`
}

// Output info from handlers
type GlobalEventOutput struct {
	InfoEvents []GlobalEventInfoEntity `json:"info_events"`
	Quizzes    []GlobalEventQuizEntity `json:"quizzes"`
}