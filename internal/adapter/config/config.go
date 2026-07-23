package config

import (
	"errors"
	"os"
	"strconv"
	"cspirt/internal/adapter/redis"
)

type Config struct {
	Redis         redis.Config

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
		Redis:         loadRedisConfig(),
		JWTSecret:     jwtSecret,
		DBPath:        dbPath,
		Port:          port,
		Parallels:     os.Getenv("PARALLELS"),
		SeedTestUsers: os.Getenv("SEED_TEST_USERS") == "1",
	}, nil
}

func loadRedisConfig() redis.Config {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	db := 0
	if raw := os.Getenv("REDIS_DB"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			db = parsed
		}
	}

	return redis.Config{
		Host:     host,
		Port:     port,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
}
