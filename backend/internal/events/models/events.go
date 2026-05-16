package models

import "time"

type Event struct {
	ID           int       `json:"ID"`
	Title        string    `json:"Title"`
	Status       string    `json:"Status"`
	BaseRatingReward int       `json:"RatingReward"`
	Description  string    `json:"Description"`
	CreatedAt    time.Time `json:"CreatedAt"`
	StartedAt    string    `json:"StartedAt"`
	Players      []int     `json:"Players"`
	Classes      []int     `json:"Classes"`
}

type EventParams struct {
	ExtraRatingReward int `json:"ExtraRatingReward"`
	Reason 	 string `json:"Reason"`
	EventID int `json:"EventID"`
	ClassID int `json:"ClassID"`
}

const (
	EventStatusScheduled   = "scheduled"
	EventStatusActive    = "active"
	EventStatusCompleted = "completed"
)