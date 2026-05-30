package models

import (
	"time"
	"cspirt/internal/events/models"
)


type User struct {
	ID       int        `json:"Id"`
	Avatar   string     `json:"Avatar"`
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
	Avatar   string     `json:"Avatar"`
	Name     string     `json:"Name"`
	LastName string     `json:"LastName"`
	FullName []FullName `json:"FullName"`
	Login    string     `json:"Login"`
	Rating   int        `json:"Rating"`
	Role     string     `json:"Role"`
	Class    string     `json:"Class"`
	ClassID  int        `json:"ClassID"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
	FullName  *[]FullName `json:"full_name,omitempty"`
	Login     *string `json:"login,omitempty"`
	Rating    *int    `json:"rating,omitempty"`
	Role      *string `json:"role,omitempty"`
	Class     *string `json:"class,omitempty"`
	ClassID   *int    `json:"class_id,omitempty"`
}

type UserWithFullInfo struct {
	User       *SafeUser     `json:"User"`
	Notes      []Note        `json:"Notes"`
	Complaints []Complaint   `json:"Complaints"`
	ClassTeacher *SafeUser   `json:"ClassTeacher"`
	Events     []models.Event`json:"Events"`
}

type FullName struct {
	Name     string `json:"Name"`
	LastName string `json:"LastName"`
	MiddleName string `json:"MiddleName"`
}

type Note struct {
	ID        int      `json:"ID"`
	TargetID  int      `json:"TargetID"`
	TargetName string  `json:"TargetName"`
	AuthorID  int      `json:"AuthorID"`
	AuthorName string  `json:"AuthorName"`
	Content   string   `json:"Content"`
	CreatedAt time.Time`json:"CreatedAt"`
}

type Complaint struct {
	ID        int     `json:"ID"`
	TargetID  int     `json:"TargetID"`
	TargetName string `json:"TargetName"`
	AuthorID  int     `json:"AuthorID"`
	AuthorName string `json:"AuthorName"`
	Content   string  `json:"Content"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type RefreshToken struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
