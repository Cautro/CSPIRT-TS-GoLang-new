package main

import (
	"cspirt/internal/handlers"
	// "cspirt/internal/logger"
	"cspirt/internal/storage"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main()  {
	_ = godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET not set in environment")
	}

	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.Mkdir("data", 0o755); err != nil {
			slog.Error("create data dir", "error", err)
		}
	}

	dbPath := "data/storage.db"
	st, err := storage.NewStorage(dbPath)
	if err != nil {
		slog.Error("open sqlite storage", "error", err)
	}
	defer st.Close()

	r := gin.Default()
	r.GET("/health", handlers.HealthHandler)

	addr := ":8080"
	slog.Info("server listening", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("server failed", "error", err)
	}
}