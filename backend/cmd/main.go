package main

import (
	"cspirt/internal/app"
	"cspirt/internal/logger"
	"cspirt/internal/storage"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			slog.Error("flush logger", "error", err)
		}
	}()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET not set in environment")
		return
	}

	if err := os.MkdirAll("data", 0o755); err != nil {
		slog.Error("create data dir", "error", err)
		return
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/storage.db"
	}

	store, err := storage.NewUserStorage(dbPath, jwtSecret)
	if err != nil {
		slog.Error("open sqlite storage", "error", err)
		return
	}
	defer store.Close()

	if os.Getenv("SEED_TEST_USERS") == "1" {
		if err := store.SeedTestUsers(); err != nil {
			slog.Error("failed to seed test users", "error", err)
			return
		}
	}

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	router := app.NewRouter(store, jwtSecret)
	slog.Info("server listening", "addr", addr)
	if err := router.Run(addr); err != nil {
		slog.Error("server failed", "error", err)
	}
}
