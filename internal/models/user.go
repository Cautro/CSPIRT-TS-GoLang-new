package models

import "time"

type User struct {
	ID       int        `json:"Id"`
	Name     string     `json:"Name"`
	LastName string     `json:"LastName"`
	FullName []FullName `json:"FullName"`
	Login    string     `json:"Login"`
	Password string     `json:"Password"`
	Rating   int        `json:"Rating"`
	Role     string     `json:"Role"`
	Class    string     `json:"Class"`
	ClassID  int        `json:"ClassID"`
}

type SafeUser struct {
	ID       int        `json:"Id"`
	Name     string     `json:"Name"`
	LastName string     `json:"LastName"`
	FullName []FullName `json:"FullName"`
	Login    string     `json:"Login"`
	Rating   int        `json:"Rating"`
	Role     string     `json:"Role"`
	Class    string     `json:"Class"`
	ClassID  int        `json:"ClassID"`
}

type UserWithFullInfo struct {
	User       *SafeUser
	Notes      []Note
	Complaints []Complaint
	ClassTeacher *SafeUser
	Events     []Event
}

type Event struct {
	ID          int
	Title       string
	Description string
	CreatedAt   string
	startedAt   string
	Players     []int
}

type FullName struct {
	Name     string `json:"Name"`
	LastName string `json:"LastName"`
}

type Note struct {
	ID        int    `json:"ID"`
	TargetID  int    `json:"TargetID"`
	AuthorID  int    `json:"AuthorID"`
	Content   string `json:"Content"`
	CreatedAt string `json:"CreatedAt"`
}

type Complaint struct {
	ID        int    `json:"ID"`
	TargetID  int    `json:"TargetID"`
	AuthorID  int    `json:"AuthorID"`
	Content   string `json:"Content"`
	CreatedAt string `json:"CreatedAt"`
}

type RefreshToken struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
