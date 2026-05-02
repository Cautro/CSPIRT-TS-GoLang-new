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
	r.GET("/health", handlers.HealthHandler) // Endpoint для проверки работоспособности сервера
	r.POST("/login", handlers.LoginHandler(s)) // Endpoint для входа и получения JWT
	r.POST("/api/refresh", handlers.RefreshHandler(s)) // Endpoint для обновления токена

	auth := r.Group("/api", utils.AuthMiddleware(jwtSecret))
	{
		// user handlers
		auth.GET("/users", handlers.GetUsersHandler(s)) // Получить всех пользователей или конкретного пользователя по ID (через Query параметр)
		auth.PATCH("/user/add", handlers.AddUserHandler(s)) // Добавление нового пользователя
		auth.DELETE("/user/delete/:id", handlers.DeleteUserHandler(s)) // Удаление пользователя по ID
		auth.GET("/me", handlers.GetMeHandler(s)) // Получить информацию о текущем пользователе

		// Class handlers
		auth.GET("/classes", handlers.GetClassesHandler(s)) // Получить все классы
		auth.GET("/classes/:class_id/users", handlers.GetClassUsersHandler(s)) // Получить всех пользователей класса
		auth.GET("/classes/:class_id/teacher", handlers.GetClassTeacherHandler(s))   // Получить учителя
		auth.PATCH("/classes/:class_id/teacher", handlers.SetClassTeacherHandler(s)) // Установить учителя

		// Rating handlers
		auth.GET("/rating", handlers.GetRatingsHandler(s)) // Получить рейтинг
		auth.PATCH("/rating/update", handlers.UpdateRatingsHandler(rs.NewRatingsService(s, s.Secret), s)) // Обновить рейтинг

		// Notes handlers
		auth.GET("/notes", handlers.GetNotesHandler(s)) // Получить заметки, с возможностью фильтрации по классу Query параметром
		auth.PATCH("/note/add", handlers.AddNoteHandler(s)) // Добавить заметку
		auth.DELETE("/note/delete/:id", handlers.DeleteNoteHandler(s)) // Удалить заметку

		// Complaints handlers
		auth.GET("/complaints", handlers.GetComplaintsHandler(s)) // Получить жалобы, с возможностью фильтрации по классу Query параметром
		auth.PATCH("/complaint/add", handlers.AddcomplaintHandler(s)) // Добавить жалобу
		auth.DELETE("/complaint/delete/:id", handlers.DeletecomplaintHandler(s)) // Удалить жалобу
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
