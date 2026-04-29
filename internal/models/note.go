package models

type AddNewNoteResponse struct {
	ID        int    `json:"ID"`
    TargetID  int	 `json:"TargetID"`
    AuthorID  int	 `json:"AuthorID"`
    Content   string `json:"Content"`
    CreatedAt string `json:"CreatedAt"`
}