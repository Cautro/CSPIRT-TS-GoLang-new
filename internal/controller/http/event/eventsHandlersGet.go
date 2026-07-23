package handlers

import (
	sr "cspirt/internal/usecase/event"
	"cspirt/pkg/logger"
	"net/http"
	"strconv"
	"context"
	"time"

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
func GetEventsHandler(eventService *sr.EventsUsecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		userIdStr := c.Query("user_id")
		classIdStr := c.Query("class_id")
		if classIdStr == "" {
			classIdStr = c.Query("class")
		}

		if userIdStr != "" {
			userID, err := strconv.Atoi(userIdStr)
			if err != nil {
				logger.WriteSafe(logger.LogEntry{
					Level:   "info",
					Action:  "get_events",
					Login:   c.GetString("Login"),
					Message: "invalid user id: " + err.Error(),
				})
				c.JSON(400, gin.H{"error": "Invalid user ID"})
				return
			}
			events, err := eventService.GetEventsByUserID(ctx, userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to get events"})
				return
			}
			c.JSON(200, events)
			return
		}
		if classIdStr != "" {
			classID, err := strconv.Atoi(classIdStr)
			if err != nil {
				logger.WriteSafe(logger.LogEntry{
					Level:   "info",
					Action:  "get_events",
					Login:   c.GetString("Login"),
					Message: "invalid class id: " + err.Error(),
				})
				c.JSON(400, gin.H{"error": "Invalid class ID"})
				return
			}
			events, err := eventService.GetEventsByClassID(ctx, classID)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to get events"})
				return
			}
			c.JSON(200, events)
			return
		}

		eventId := c.Query("event_id")
		if eventId != "" {
			eventIdInt, err := strconv.Atoi(eventId)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Event ID format"})
				return
			}

			event, err := eventService.GetEventsByEventID(ctx, eventIdInt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event"})
				return
			}

			if event == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
				return
			}

			c.JSON(http.StatusOK, event)
			return
		}

		events, err := eventService.GetEvents(ctx)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get events"})
			return
		}
		c.JSON(200, events)
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
func GetEventParamsHandler(eventService *sr.EventsUsecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		
		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}
		eventID, err := strconv.Atoi(c.Param("eventId"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		params, err := eventService.GetEventParams(ctx, eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event params"})
			return
		}

		c.JSON(200, params)
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
func GetEventPlayersHandler(eventService *sr.EventsUsecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		eventID, err := strconv.Atoi(c.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "get_event_players",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		players, err := eventService.GetEventPlayers(ctx, eventID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get event players"})
			return
		}

		c.JSON(200, players)
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
func GetEventPlayersCountHandler(eventService *sr.EventsUsecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		eventID, err := strconv.Atoi(c.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "get_event_players_count",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		count, err := eventService.GetEventPlayersCount(ctx, eventID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get event players count"})
			return
		}

		c.JSON(200, gin.H{"count": count})
	}
}
