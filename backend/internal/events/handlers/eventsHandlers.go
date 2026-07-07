package handlers

import (
	sr "cspirt/internal/events/service"
	"cspirt/internal/logger"
	"cspirt/internal/storage"
	"cspirt/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// EventComplete marks an event as completed.
// @Summary Complete event
// @Description Marks an event as completed with rating and class rewards.
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path int true "Event ID"
// @Param request body object{ratingReward=int,classReward=int} true "Completion payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/complete [patch]
func EventComplete(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "complete_event",
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
				Action:  "complete_event",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			RatingReward int `json:"ratingReward"`
			ClassReward  int `json:"classReward"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "complete_event",
				Login:   ctx.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.EventComplete(eventID, req.RatingReward, req.ClassReward); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to complete event"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "complete_event",
			Login:   ctx.GetString("Login"),
			Message: "event completed successfully",
		})
		ctx.JSON(200, gin.H{"message": "Event completed successfully"})
	}
}
