package models

import "cspirt/internal/users/models"

type Class struct {
	ID           int        `json:"Id"`
	Name         string     `json:"Name"`
	TeacherLogin string     `json:"TeacherLogin,omitempty"`
	Teacher      *models.SafeUser  `json:"Teacher,omitempty"` 
	Members      []models.SafeUser `json:"Members"`
	UserTotalRating  int        `json:"UserTotalRating"`
	ClassTotalRating  int        `json:"ClassTotalRating"`
}

type ClassTeacherInput struct {
	TeacherLogin string `json:"TeacherLogin" binding:"required"`
}

type ClassInput struct {
	Name         string     `json:"Name"`
	TeacherLogin string     `json:"TeacherLogin,omitempty"`
}
