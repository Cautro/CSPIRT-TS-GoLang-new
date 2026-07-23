package complaint

type AddNewComplaintResponse struct {
	ID        int    `json:"ID"`
    TargetID  int	 `json:"TargetID"`
    TargetName string `json:"TargetName"`
    Content   string `json:"Content"`
    CreatedAt string `json:"CreatedAt"`
}
