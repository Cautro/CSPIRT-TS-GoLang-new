package handlers

import (
	models "cspirt/internal/domain/event"
	sr "cspirt/internal/usecase/event"
	permissionService "cspirt/internal/controller/permission/usecase"
	"cspirt/pkg/logger"
	"strconv"
	"context"
	"time"

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
func UpdateEventParamsHandler(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := perm.CheckUserRole(ctx, c.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "update_event_params",
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
				Action:  "update_event_params",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}
		var req *models.EventParams
		if err := c.BindJSON(&req); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_event_params",
				Login:   c.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.UpdateEventParams(ctx, eventID, req); err != nil {
			c.JSON(500, gin.H{"error": "Failed to update event parameters"})
			return
		}
		c.JSON(200, gin.H{"message": "Event parameters updated successfully"})
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
func UpdateEventHandler(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := perm.CheckUserRole(ctx, c.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "update_event",
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
				Action:  "update_event",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}
		var req *models.Event
		if err := c.BindJSON(&req); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_event",
				Login:   c.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.UpdateEvent(ctx, eventID, req); err != nil {
			c.JSON(500, gin.H{"error": "Failed to update event"})
			return
		}
		c.JSON(200, gin.H{"message": "Event updated successfully"})
	}
}
