package models

type Class struct {
	ID           int        `json:"Id"`
	Name         string     `json:"Name"`
	TeacherLogin string     `json:"TeacherLogin,omitempty"`
	Teacher      *SafeUser  `json:"Teacher,omitempty"`
	Members      []SafeUser `json:"Members"`
	TotalRating  int        `json:"TotalRating"`
}

type ClassTeacherInput struct {
	TeacherLogin string `json:"TeacherLogin" binding:"required"`
}
