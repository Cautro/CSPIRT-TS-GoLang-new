package handlers

import (
	"context"
	"strconv"
	"time"

	sr "cspirt/internal/usecase/event"
	permissionService "cspirt/internal/controller/permission/usecase"
	"cspirt/pkg/logger"
	ratingModels "cspirt/internal/domain/rating"

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
func DeleteEventParamsHandler(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := perm.CheckUserRole(ctx, c.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_event_params",
				Login:   c.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventID, err := strconv.Atoi(c.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_event_params",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		if err := eventService.DeleteEventParams(ctx, eventID); err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete event params"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_event_params",
			Login:   c.GetString("Login"),
			Message: "event params deleted successfully",
		})
		c.JSON(200, gin.H{"message": "Event params deleted successfully"})
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
func DeleteEventHandler(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := perm.CheckUserRole(ctx, c.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "delete_event",
				Login:   c.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_event",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		if err := eventService.DeleteEvent(ctx, eventID); err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete event"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "delete_event",
			Login:   c.GetString("Login"),
			Message: "event deleted successfully",
		})
		c.JSON(200, gin.H{"message": "Event deleted successfully"})
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
func DeletePlayersFromEvent(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := perm.CheckUserRole(ctx, c.GetString("Login"), string(ratingModels.RoleOwner), string(ratingModels.RoleAdmin))
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

		if err := eventService.DeletePlayersFromEvent(ctx, eventID, req.PlayerIDs); err != nil {
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
