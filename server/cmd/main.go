package main

import (
	aHandlers "cspirt/internal/auth/handlers"
	clHandlers "cspirt/internal/class/handlers"
	cmHandlers "cspirt/internal/complaints/handlers"
	hHandlers "cspirt/internal/health/handlers"
	"cspirt/internal/logger"
	eHandlers "cspirt/internal/events/handlers"
	nHandlers "cspirt/internal/note/handlers"
	rHandlers "cspirt/internal/rating/handler"
	rs "cspirt/internal/rating/service"
	"cspirt/internal/storage"
	uHandlers "cspirt/internal/users/handlers"
	utils "cspirt/internal/utils"
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
	r.GET("/health", hHandlers.HealthHandler)           // Endpoint для проверки работоспособности сервера
	r.POST("/login", aHandlers.LoginHandler(s))         // Endpoint для входа и получения JWT
	r.POST("/api/refresh", aHandlers.RefreshHandler(s)) // Endpoint для обновления токена

	auth := r.Group("/api", utils.AuthMiddleware(jwtSecret))
	{
		// user handlers
		auth.GET("/users", uHandlers.GetUsersHandler(s))                // Получить всех пользователей или конкретного пользователя по ID (через Query параметр)
		auth.PATCH("/user/add", uHandlers.AddUserHandler(s))            // Добавление нового пользователя
		auth.DELETE("/user/delete/:id", uHandlers.DeleteUserHandler(s)) // Удаление пользователя по ID
		auth.GET("/me", uHandlers.GetMeHandler(s))                      // Получить информацию о текущем пользователе

		// Class handlers
		auth.GET("/classes", clHandlers.GetClassesHandler(s))                          // Получить все классы
		auth.GET("/classes/teacher", clHandlers.GetClassTeachersHandler(s))            // Получить всех классных руководителей
		auth.PATCH("/classes/add", clHandlers.AddClassHandler(s))                      // Добавить класс
		auth.DELETE("/classes/delete/:id", clHandlers.DeleteClassHandler(s))           // Удалить класс по ID
		auth.GET("/classes/:class_id/users", clHandlers.GetClassUsersHandler(s))       // Получить всех пользователей класса
		auth.GET("/classes/:class_id/teacher", clHandlers.GetClassTeacherHandler(s))   // Получить учителя
		auth.PATCH("/classes/:class_id/teacher", clHandlers.SetClassTeacherHandler(s)) // Установить учителя

		// Rating handlers
		auth.GET("/rating", rHandlers.GetRatingsHandler(s))                                                              // Получить рейтинг
		auth.PATCH("/rating/update", rHandlers.UpdateRatingsHandler(rs.NewRatingsService(s.RatingRepo, s, s.Secret), s)) // Обновить рейтинг

		// Notes handlers
		auth.GET("/notes", nHandlers.GetNotesHandler(s))                // Получить заметки, с возможностью фильтрации по классу Query параметром
		auth.PATCH("/note/add", nHandlers.AddNoteHandler(s))            // Добавить заметку
		auth.DELETE("/note/delete/:id", nHandlers.DeleteNoteHandler(s)) // Удалить заметку

		// Complaints handlers
		auth.GET("/complaints", cmHandlers.GetComplaintsHandler(s))                // Получить жалобы, с возможностью фильтрации по классу Query параметром
		auth.PATCH("/complaint/add", cmHandlers.AddcomplaintHandler(s))            // Добавить жалобу
		auth.DELETE("/complaint/delete/:id", cmHandlers.DeletecomplaintHandler(s)) // Удалить жалобу

		// Events handlers
		auth.GET("/events", eHandlers.GetEventsHandler(s))		                     		// Получить события, с возможностью фильтрации по классу Query параметром
		auth.PATCH("/event/add", eHandlers.AddEventHandler(s))                              // Добавить событие
		auth.DELETE("/event/delete/:id", eHandlers.DeleteEventHandler(s))                   // Удалить событие
		auth.PATCH("/event/:eventId/players/add", eHandlers.AddPlayersToEvent(s))           // Добавить игроков к событию
		auth.DELETE("/event/:eventId/players/delete", eHandlers.DeletePlayersFromEvent(s))  // Удалить игроков из события
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
