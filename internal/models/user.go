package models

type User struct {
	ID 	 int           `json:"Id"`
	Name string        `json:"Name"`
	LastName string    `json:"LastName"`
	Login string       `json:"Login"`
	Password string    `json:"Password"`
	Rating int         `json:"Rating"`
	Role string        `json:"Role"`
	Class string       `json:"Class"`
	Notes []Note       `json:"Notes"`
	Complaints []Complaint `json:"Complaints"`
}

type FullName struct {
	Name string     `json:"Name"`
	LastName string `json:"LastName"`
}

type Note struct {
	ID int `json:"Id"`
	UserID int `json:"UserId"`
	Content string `json:"Content"`
}

type Complaint struct {
	ID int `json:"Id"`
	UserID int `json:"UserId"`
	Content string `json:"Content"`
}
