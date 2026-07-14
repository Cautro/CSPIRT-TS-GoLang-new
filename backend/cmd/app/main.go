package main

import (
	authService "cspirt/internal/usecase/auth"
	classConfig "cspirt/internal/controller/http/class/config"
	classService "cspirt/internal/usecase/class"
	complaintService "cspirt/internal/usecase/complaint"
	"cspirt/internal/adapter/config"
	router "cspirt/internal/controller/http"
	eventsService "cspirt/internal/usecase/event"
	noteService "cspirt/internal/usecase/note"
	permissionService "cspirt/internal/controller/permission/usecase"
	ratingService "cspirt/internal/usecase/rating"
	scheduleService "cspirt/internal/usecase/schedule"
	"cspirt/internal/adapter/postgres/storage"
	usersService "cspirt/internal/usecase/user"
	"cspirt/pkg/logger"
	"log/slog"
	"os"

	userPostgres "cspirt/internal/adapter/postgres/user"
	classPostgres "cspirt/internal/adapter/postgres/class"
	notePostgres "cspirt/internal/adapter/postgres/note"
	complaintPostgres "cspirt/internal/adapter/postgres/complaint"
	eventPostgres "cspirt/internal/adapter/postgres/event"
	ratingPostgres "cspirt/internal/adapter/postgres/rating"
	schedulePostgres "cspirt/internal/adapter/postgres/schedule"

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

	cfg, err := config.Load()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if err := os.MkdirAll("data", 0o755); err != nil {
		slog.Error("create data dir", "error", err)
		return
	}

	store, err := storage.New(cfg.DBPath)
	if err != nil {
		slog.Error("open storage", "error", err)
		return
	}
	defer store.Close()

	uRepo := userPostgres.New(store.DB)
	clRepo := classPostgres.New(store.DB)
	nRepo := notePostgres.New(store.DB)
	cRepo := complaintPostgres.New(store.DB)
	eRepo := eventPostgres.New(store.DB)
	rRepo := ratingPostgres.New(store.DB)
	schRepo := schedulePostgres.New(store.DB)

	usersSvc := usersService.NewUsersUsecase(uRepo, nRepo, cRepo, clRepo, eRepo)
	classSvc := classService.NewClassUsecase(clRepo)
	noteSvc := noteService.NewNoteUsecase(nRepo)
	complaintSvc := complaintService.NewComplaintsUsecase(cRepo)
	eventsSvc := eventsService.NewEventsUsecase(eRepo)
	ratingSvc := ratingService.NewRatingsUsecase(rRepo, uRepo)
	scheduleSvc := scheduleService.NewScheduleUsecase(schRepo)
	permSvc := permissionService.New(uRepo)
	authSvc := authService.NewAuthService(uRepo, cfg.JWTSecret)

	if cfg.Parallels != "" {
		parallelsConfig, err := classConfig.ParseParallelsConfig(cfg.Parallels)
		if err != nil {
			slog.Error("failed to parse PARALLELS config", "error", err)
			return
		}
		if err := classSvc.InitializeParallelsFromConfig(parallelsConfig); err != nil {
			slog.Error("failed to initialize parallels", "error", err)
			return
		}
		slog.Info("Parallels initialized", "count", len(parallelsConfig))
	}

	if cfg.SeedTestUsers {
		if err := storage.SeedTestUsers(store.DB); err != nil {
			slog.Error("failed to seed test users", "error", err)
			return
		}
	}

	r := router.NewRouter(router.Usecases{
		Auth:       authSvc,
		Users:      usersSvc,
		Class:      classSvc,
		Note:       noteSvc,
		Complaint:  complaintSvc,
		Events:     eventsSvc,
		Rating:     ratingSvc,
		Schedule:   scheduleSvc,
		Permission: permSvc,
		JWTSecret:  cfg.JWTSecret,
	})

	slog.Info("server listening", "addr", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
	}
}
