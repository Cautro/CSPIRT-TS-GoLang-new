package handlers

import (
	sr "cspirt/internal/events/service"
	ratingModels "cspirt/internal/rating/models"
	"cspirt/internal/events/models"
	"github.com/gin-gonic/gin"
	"cspirt/internal/storage"
	"cspirt/internal/logger"
	"cspirt/internal/utils"
	"strconv"
)

func AddEventHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event",
				Login:   ctx.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
		var event models.Event
		if err := ctx.BindJSON(&event); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event",
				Login:   ctx.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.AddEvent(event); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to add event"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_event",
			Login:   ctx.GetString("Login"),
			Message: "event added successfully",
		})
		ctx.JSON(200, gin.H{"message": "Event added successfully"})
	}
}

func AddEventParams(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_params",
				Login:   ctx.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_params",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var params models.EventParams
		if err := ctx.BindJSON(&params); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_params",
				Login:   ctx.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.AddEventParams(eventID, &params); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to add event params"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_event_params",
			Login:   ctx.GetString("Login"),
			Message: "event params added successfully",
		})

		ctx.JSON(200, gin.H{"message": "Event params added successfully"})
	}
}

func AddPlayersToEvent(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		login := ctx.GetString("Login")
		user, err := s.GetUserByLogin(login)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_players",
				Login:   login,
				Message: "failed to get user: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to get user"})
			return 
		}

		err = utils.CheckUserRole(s, ctx.GetString("Login"), string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_players",
				Login:   ctx.GetString("Login"),
				Role:    user.Role,
				Message: "failed to check user role: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_players",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			PlayerIDs []int `json:"playerIds"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_players",
				Login:   ctx.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.AddPlayersToEvent(eventID, req.PlayerIDs, ctx.GetString("Login")); err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_event_players",
			Login:   ctx.GetString("Login"),
			Message: "players added to event successfully",
		})
		ctx.JSON(200, gin.H{"message": "Players added to event successfully"})
	}
}