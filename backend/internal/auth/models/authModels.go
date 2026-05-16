package models

type LoginInput struct {
	Login string `json:"Login" binding:"required"`
	Password string `json:"Password" binding:"required"`
}