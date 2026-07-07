package handlers

import (
	"cspirt/internal/events/models"
	sr "cspirt/internal/events/service"
	"cspirt/internal/logger"
	"cspirt/internal/storage"
	"cspirt/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UpdateEventParamsHandler updates event parameters.
// @Summary Update event params
// @Description Updates the parameters of the specified event.
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path int true "Event ID"
// @Param request body models.EventParams true "Updated event params"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/params/update [patch]
func UpdateEventParamsHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "update_event_params",
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
				Action:  "update_event_params",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}
		var req *models.EventParams
		if err := ctx.BindJSON(&req); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_event_params",
				Login:   ctx.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.UpdateEventParams(eventID, req); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update event parameters"})
			return
		}
		ctx.JSON(200, gin.H{"message": "Event parameters updated successfully"})
	}
}

// UpdateEventHandler updates an existing event.
// @Summary Update event
// @Description Updates the event with the provided ID.
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path int true "Event ID"
// @Param request body models.Event true "Updated event payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/update [patch]
func UpdateEventHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "update_event",
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
				Action:  "update_event",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}
		var req *models.Event
		if err := ctx.BindJSON(&req); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_event",
				Login:   ctx.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.UpdateEvent(eventID, req); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update event"})
			return
		}
		ctx.JSON(200, gin.H{"message": "Event updated successfully"})
	}
}
