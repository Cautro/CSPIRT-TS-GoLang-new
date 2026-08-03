package note

import "time"

type AddNewNoteResponse struct {
	ID         int    `json:"ID"`
	TargetID   int    `json:"TargetID"`
	TargetName string `json:"TargetName"`
	AuthorID   int    `json:"AuthorID"`
	AuthorName string `json:"AuthorName"`
	Content    string `json:"Content"`
	CreatedAt  string `json:"CreatedAt"`

	ModerateAt       time.Time `json:"ModerateAt"`
    ModeratorId      int `json:"ModeratorId"`
	ModerationStatus string `json:"ModerationStatus"`
}