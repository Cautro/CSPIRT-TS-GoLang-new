package handlers

import (
	sr "cspirt/internal/events/service"
	"cspirt/internal/logger"
	ratingModels "cspirt/internal/rating/models"
	"cspirt/internal/storage"
	"cspirt/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteEventParamsHandler deletes params for an event.
// @Summary Delete event params
// @Description Deletes the params for the specified event.
// @Tags events
// @Produce json
// @Param eventId path int true "Event ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/params/delete [delete]
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

// DeleteEventHandler deletes an event by ID.
// @Summary Delete event
// @Description Deletes the event with the provided ID.
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/delete/{id} [delete]
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

// DeletePlayersFromEvent removes players from an event.
// @Summary Delete players from event
// @Description Removes one or more player IDs from the specified event.
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path int true "Event ID"
// @Param request body object{playerIds=[]int} true "Player IDs payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/players/delete [delete]
func DeletePlayersFromEvent(s *storage.Storage) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := utils.CheckUserRole(s, c.GetString("Login"), string(ratingModels.RoleOwner), string(ratingModels.RoleAdmin))
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
