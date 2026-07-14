package config

import (
	"errors"
	"os"
)

type Config struct {
	JWTSecret     string
	DBPath        string
	Port          string
	Parallels     string
	SeedTestUsers bool
}

func Load() (Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET not set in environment")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/storage.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	return Config{
		JWTSecret:     jwtSecret,
		DBPath:        dbPath,
		Port:          port,
		Parallels:     os.Getenv("PARALLELS"),
		SeedTestUsers: os.Getenv("SEED_TEST_USERS") == "1",
	}, nil
}
