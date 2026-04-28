package models

type Role string

const (
    RoleUser   Role = "user"
    RoleHelper Role = "helper"
    RoleAdmin  Role = "admin"
    RoleOwner  Role = "owner"
)

type RatingInput struct {
			Rating int `json:"rating"`
			TargetLogin string `json:"target_login"`
			Reason string `json:"reason"`
			Anonymous bool `json:"anonymous"`
		}