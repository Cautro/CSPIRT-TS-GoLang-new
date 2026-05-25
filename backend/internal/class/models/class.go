package models

import "cspirt/internal/users/models"

type Class struct {
	ID                int               `json:"Id"`
	Name              string            `json:"Name"`
	TeacherLogin      string            `json:"TeacherLogin,omitempty"`
	Teacher           *models.SafeUser  `json:"Teacher,omitempty"` 
	FirstQuarterComplete int               `json:"FirstQuarterComplete"`
	SecondQuarterComplete int               `json:"SecondQuarterComplete"`
	ThirdQuarterComplete int               `json:"ThirdQuarterComplete"`
	QuarterComplete int               `json:"QuarterComplete"`
	Members           []models.SafeUser `json:"Members"`
	Parallel          int	        	`json:"Parallel"`
	UserTotalRating   int               `json:"UserTotalRating"`
	ClassTotalRating  int               `json:"ClassTotalRating"`
}

type ParallelClass struct {
	ID           int        `json:"Id"`
	Name         string     `json:"Name"`
	BestClassID   int       `json:"BestClassId"`
	ClassesIDs    []int     `json:"ClassesIds"`
	ClassTotalRating  int   `json:"ClassTotalRating"`
}

type ClassTeacherInput struct {
	TeacherLogin string `json:"TeacherLogin" binding:"required"`
}

type ClassInput struct {
	Name         string     `json:"Name"`
	TeacherLogin string     `json:"TeacherLogin,omitempty"`
}
