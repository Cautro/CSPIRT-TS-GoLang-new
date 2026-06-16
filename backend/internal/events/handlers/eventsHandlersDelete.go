package handlers

import (
	sr "cspirt/internal/events/service"
	"github.com/gin-gonic/gin"
	"cspirt/internal/storage"
	"cspirt/internal/logger"
	"cspirt/internal/utils"
	"strconv"
)

func DeleteEventParamsHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_event_params",
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
				Action:  "delete_event_params",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		if err := eventService.DeleteEventParams(eventID); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete event params"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_event_params",
			Login:   ctx.GetString("Login"),
			Message: "event params deleted successfully",
		})
		ctx.JSON(200, gin.H{"message": "Event params deleted successfully"})
	}
}

func DeleteEventHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_event",
				Login:   ctx.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
		eventID, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_event",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		if err := eventService.DeleteEvent(eventID); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete event"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_event",
			Login:   ctx.GetString("Login"),
			Message: "event deleted successfully",
		})
		ctx.JSON(200, gin.H{"message": "Event deleted successfully"})
	}
}

func DeletePlayersFromEvent(s *storage.Storage) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := utils.CheckUserRole(s, c.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_event_players",
				Login:   c.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
		eventID, err := strconv.Atoi(c.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_event_players",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			PlayerIDs []int `json:"playerIds"`
		}
		if err := c.BindJSON(&req); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_event_players",
				Login:   c.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.DeletePlayersFromEvent(eventID, req.PlayerIDs); err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete players from event"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_event_players",
			Login:   c.GetString("Login"),
			Message: "players deleted from event successfully",
		})
		c.JSON(200, gin.H{"message": "Players deleted from event successfully"})
	}
}
