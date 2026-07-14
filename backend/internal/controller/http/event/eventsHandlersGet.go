package handlers

import (
	sr "cspirt/internal/usecase/event"
	"cspirt/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetEventsHandler returns events filtered by user, class, or event ID.
// @Summary List events
// @Description Returns events, optionally filtered by user_id, class_id, or event_id query parameters.
// @Tags events
// @Produce json
// @Param user_id query int false "User ID"
// @Param class_id query int false "Class ID"
// @Param event_id query int false "Event ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/events [get]
func GetEventsHandler(eventService *sr.EventsUsecase) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		userIdStr := ctx.Query("user_id")
		classIdStr := ctx.Query("class_id")
		if classIdStr == "" {
			classIdStr = ctx.Query("class")
		}

		if userIdStr != "" {
			userID, err := strconv.Atoi(userIdStr)
			if err != nil {
				logger.WriteSafe(logger.LogEntry{
					Level:   "info",
					Action:  "get_events",
					Login:   ctx.GetString("Login"),
					Message: "invalid user id: " + err.Error(),
				})
				ctx.JSON(400, gin.H{"error": "Invalid user ID"})
				return
			}
			events, err := eventService.GetEventsByUserID(userID)
			if err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to get events"})
				return
			}
			ctx.JSON(200, events)
			return
		}
		if classIdStr != "" {
			classID, err := strconv.Atoi(classIdStr)
			if err != nil {
				logger.WriteSafe(logger.LogEntry{
					Level:   "info",
					Action:  "get_events",
					Login:   ctx.GetString("Login"),
					Message: "invalid class id: " + err.Error(),
				})
				ctx.JSON(400, gin.H{"error": "Invalid class ID"})
				return
			}
			events, err := eventService.GetEventsByClassID(classID)
			if err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to get events"})
				return
			}
			ctx.JSON(200, events)
			return
		}

		eventId := ctx.Query("event_id")
		if eventId != "" {
			eventIdInt, err := strconv.Atoi(eventId)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Event ID format"})
				return
			}

			event, err := eventService.GetEventsByEventID(eventIdInt)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event"})
				return
			}

			if event == nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
				return
			}

			ctx.JSON(http.StatusOK, event)
			return
		}

		events, err := eventService.GetEvents()
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to get events"})
			return
		}
		ctx.JSON(200, events)
	}
}

// GetEventParamsHandler returns params for an event.
// @Summary Get event params
// @Description Returns event parameters for the specified event.
// @Tags events
// @Produce json
// @Param eventId path int true "Event ID"
// @Success 200 {array} models.EventParams
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/params [get]
func GetEventParamsHandler(eventService *sr.EventsUsecase) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		params, err := eventService.GetEventParams(eventID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event params"})
			return
		}

		ctx.JSON(200, params)
	}
}

// GetEventPlayersHandler returns the players assigned to an event.
// @Summary Get event players
// @Description Returns players for the specified event.
// @Tags events
// @Produce json
// @Param eventId path int true "Event ID"
// @Success 200 {array} models.SafeUser
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/players [get]
func GetEventPlayersHandler(eventService *sr.EventsUsecase) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "get_event_players",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		players, err := eventService.GetEventPlayers(eventID)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to get event players"})
			return
		}

		ctx.JSON(200, players)
	}
}

// GetEventPlayersCountHandler returns the number of players in an event.
// @Summary Get event players count
// @Description Returns the count of players participating in the specified event.
// @Tags events
// @Produce json
// @Param eventId path int true "Event ID"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/players/count [get]
func GetEventPlayersCountHandler(eventService *sr.EventsUsecase) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "get_event_players_count",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		count, err := eventService.GetEventPlayersCount(eventID)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to get event players count"})
			return
		}

		ctx.JSON(200, gin.H{"count": count})
	}
}
