package storage

import (
	models "cspirt/internal/domain/user"
	utils "cspirt/internal/utils"
	"database/sql"
	"encoding/json"
)

func SeedTestUsers(db *sql.DB) error {

	users := []models.User{
		{
			Name:     "Ivan",
			LastName: "Admin",
			FullName: []models.FullName{{Name: "Ivan", LastName: "Admin", MiddleName: "Ivanovich"}},
			Login:    "Owner",
			Password: "123456",
			Rating:   100,
			Role:     "Owner",
			Class:    "10A",
		},
		{
			Name:     "Petr",
			LastName: "Teacher",
			FullName: []models.FullName{{Name: "Petr", LastName: "Teacher", MiddleName: "Petrovich"}},
			Login:    "Admin",
			Password: "123456",
			Rating:   90,
			Role:     "Admin",
			Class:    "10A",
		},
		{
			Avatar:   sql.NullString{String: "Base64Test", Valid: true},
			Name:     "Olga",
			LastName: "Student",
			FullName: []models.FullName{{Name: "Olga", LastName: "Student", MiddleName: "Olgovna"}},
			Login:    "Helper",
			Password: "123456",
			Rating:   80,
			Role:     "Helper",
			Class:    "10A",
		},
		{
			Avatar:   sql.NullString{String: "Base64Test", Valid: true},
			Name:     "Sidr",
			LastName: "MrLoveSidr",
			FullName: []models.FullName{{Name: "Sidr", LastName: "MrLoveSidr", MiddleName: "Sidrovich"}},
			Login:    "User",
			Password: "123456",
			Rating:   70,
			Role:     "User",
			Class:    "10A",
		},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO users
		(Avatar, Name, FullName, LastName, Login, Password, Rating, Role, Class)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT DO NOTHING
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
			user.Avatar,
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

	return syncAllClasses(db)
}
