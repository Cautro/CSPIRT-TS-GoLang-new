package storage

import (
	"cspirt/internal/users/models"
	utils "cspirt/internal/utils"
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
		},
		{
			Name:     "Sidr",
			LastName: "MrLoveSidr",
			FullName: []models.FullName{{Name: "Sidr", LastName: "MrLoveSidr"}},
			Login:    "User",
			Password: "123456",
			Rating:   70,
			Role:     "User",
			Class:    "10A",
		},
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT OR IGNORE INTO users
		(Name, FullName, LastName, Login, Password, Rating, Role, Class)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	for _, user := range users {
		fullNameJSON, err := json.Marshal(user.FullName)
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
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return s.syncAllClassesLocked()
}
