package models

import "time"

type Event struct {
	ID          int       `json:"ID"`
	Title       string    `json:"Title"`
	Status      string    `json:"Status"`
	Description string    `json:"Description"`
	CreatedAt   time.Time `json:"CreatedAt"`
	StartedAt   string    `json:"StartedAt"`
	Players     []int     `json:"Players"` 
}