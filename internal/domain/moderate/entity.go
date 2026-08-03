package moderate

import "time"

type Moderate struct {
	Id                int    `json:"id"`
	TargetId          int    `json:"targetId"`
	TargetStatus      string `json:"targetStatus"`
	TargetModeratorId int    `json:"targetModeratorId"`
	CreatedAt         time.Time `json:"createdAt"`
	ModerateAt        time.Time `json:"moderateAt"`
}

type ModerateDTO struct {
	NewStatus      string `json:"newStatus"`
	ModeratorId    int    `json:"moderatorId"`
}
