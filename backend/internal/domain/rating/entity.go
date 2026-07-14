package rating

type Role string

const (
    RolePublic Role = "public"
    RoleUser   Role = "User"
    RoleHelper Role = "Helper"
    RoleAdmin  Role = "Admin"
    RoleOwner  Role = "Owner"
)

type RatingInput struct {
	Rating int `json:"rating"`
	TargetLogin string `json:"target_login"`
	Reason string `json:"reason"`
}