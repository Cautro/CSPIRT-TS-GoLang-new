package storage

import (
	"cspirt/internal/models"
	utils "cspirt/internal/utils/auth"
	"encoding/json"
)

func (s *Storage) SeedTestUsers() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users := []models.User{
		{
			Name:     "Ivan",
			LastName: "Admin",
			FullName: []models.FullName{{Name: "Ivan", LastName: "Admin"}},
			Login:    "Owner",
			Password: "123456",
			Rating:   100,
			Role:     "Owner",
			Class:    "10A",
			Notes:    []models.Note{},
			Complaints: []models.Complaint{},
		},
		{
			Name:     "Petr",
			LastName: "Teacher",
			FullName: []models.FullName{{Name: "Petr", LastName: "Teacher"}},
			Login:    "Admin",
			Password: "123456",
			Rating:   90,
			Role:     "Admin",
			Class:    "10A",
			Notes:    []models.Note{},
			Complaints: []models.Complaint{},
		},
		{
			Name:     "Olga",
			LastName: "Student",
			FullName: []models.FullName{{Name: "Olga", LastName: "Student"}},
			Login:    "Helper",
			Password: "123456",
			Rating:   80,
			Role:     "Helper",
			Class:    "10A",
			Notes:    []models.Note{},
			Complaints: []models.Complaint{},
		},
		{
			Name:     "Sidor",
			LastName: "Parent",
			FullName: []models.FullName{{Name: "Sidor", LastName: "Parent"}},
			Login:    "User",
			Password: "123456",
			Rating:   70,
			Role:     "User",
			Class:    "10A",
			Notes:    []models.Note{},
			Complaints: []models.Complaint{},
		},
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT OR IGNORE INTO users
		(Name, FullName, LastName, Login, Password, Rating, Role, Class, Notes, Complaints)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	for _, user := range users {
		fullNameJSON, err := json.Marshal(user.FullName)
		if err != nil {
			return err
		}

		notesJSON, err := json.Marshal(user.Notes)
		if err != nil {
			return err
		}

		complaintsJSON, err := json.Marshal(user.Complaints)
		if err != nil {
			return err
		}

		passwordHash, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			query,
			user.Name,
			string(fullNameJSON),
			user.LastName,
			user.Login,
			passwordHash,
			user.Rating,
			user.Role,
			user.Class,
			string(notesJSON),
			string(complaintsJSON),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}