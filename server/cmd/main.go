package main

import (
	"cspirt/internal/handlers"
	"cspirt/internal/logger"
	rs "cspirt/internal/service/rating"
	"cspirt/internal/storage"
	utils "cspirt/internal/utils/auth"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("No .env file found")
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

	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.Mkdir("data", 0o755); err != nil {
			slog.Error("create data dir", "error", err)
		}
	}

	DB_PATH := os.Getenv("DB_PATH")
	if DB_PATH == "" {
		DB_PATH = "data/storage.db"
	}

	s, err := storage.NewUserStorage(DB_PATH, jwtSecret)
	if err != nil {
		slog.Error("open sqlite storage", "error", err)
		return
	}
	defer s.Close()

	if os.Getenv("SEED_TEST_USERS") == "1" {
		if err := s.SeedTestUsers(); err != nil {
			slog.Error("failed to seed test users", "error", err)
			return
		}
	}

	// Gin logic here
	r := gin.Default()
	r.GET("/health", handlers.HealthHandler)
	r.POST("/login", handlers.LoginHandler(s))
	r.POST("/api/refresh", handlers.RefreshHandler(s))

	auth := r.Group("/api", utils.AuthMiddleware(jwtSecret))
	{
		// user handlers
		auth.GET("/users", handlers.GetUsersHandler(s))
		auth.PATCH("/user/add", handlers.AddUserHandler(s))
		auth.DELETE("/user/delete/:id", handlers.DeleteUserHandler(s))
		auth.GET("/me", handlers.GetMeHandler(s))

		// Class handlers
		auth.GET("/classes", handlers.GetClassesHandler(s))
		auth.GET("/classes/:class/users", handlers.GetClassUsersHandler(s))
		auth.GET("/classes/:class/teacher", handlers.GetClassTeacherHandler(s)) // Получить учитель
		auth.PATCH("/classes/:class/teacher", handlers.SetClassTeacherHandler(s)) // Установить учителя

		// Rating handlers
		auth.GET("/rating", handlers.GetRatingsHandler(s))
		auth.PATCH("/rating/update", handlers.UpdateRatingsHandler(rs.NewRatingsService(s, s.Secret), s))

		// Notes handlers
		auth.GET("/notes", handlers.GetNotesHandler(s))
		auth.PATCH("/note/add", handlers.AddNoteHandler(s))
		auth.DELETE("/note/delete/:id", handlers.DeleteNoteHandler(s))

		// Complaints handlers
		auth.GET("/complaints", handlers.GetComplaintsHandler(s))
		auth.PATCH("/complaint/add", handlers.AddcomplaintHandler(s))
		auth.DELETE("/complaint/delete/:id", handlers.DeletecomplaintHandler(s))
	}

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}
	slog.Info("server listening", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("server failed", "error", err)
	}
}
