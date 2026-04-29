package models

type User struct {
	ID 	 int              `json:"Id"`
	Name string           `json:"Name"`
	LastName string       `json:"LastName"`
	FullName []FullName   `json:"FullName"`
	Login string          `json:"Login"`
	Password string       `json:"Password"`
	Rating int            `json:"Rating"`
	Role string           `json:"Role"`
	Class string          `json:"Class"`
}

type SafeUser struct {
	ID         int         `json:"Id"`
	Name       string      `json:"Name"`
	LastName   string      `json:"LastName"`
	FullName   []FullName  `json:"FullName"`
	Login      string      `json:"Login"`
	Rating     int         `json:"Rating"`
	Role       string      `json:"Role"`
	Class      string      `json:"Class"`
}

type FullName struct {
	Name string     `json:"Name"`
	LastName string `json:"LastName"`
}

type Note struct {
    ID        int    `json:"ID"`
    TargetID  int	 `json:"TargetID"`
    AuthorID  int	 `json:"AuthorID"`
    Content   string `json:"Content"`
    CreatedAt string `json:"CreatedAt"`
}

type Complaint struct {
    ID        int    `json:"ID"`
    TargetID  int	 `json:"TargetID"`
    AuthorID  int	 `json:"AuthorID"`
    Content   string `json:"Content"`
    CreatedAt string `json:"CreatedAt"`
}
