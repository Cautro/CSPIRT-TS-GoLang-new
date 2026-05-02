package models

type AddNewComplaintResponse struct {
	ID        int    `json:"ID"`
    TargetID  int	 `json:"TargetID"`
    TargetName string `json:"TargetName"`
    AuthorID  int	 `json:"AuthorID"`
    AuthorName string `json:"AuthorName"`
    Content   string `json:"Content"`
    CreatedAt string `json:"CreatedAt"`
}