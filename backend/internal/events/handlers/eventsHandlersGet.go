package handlers

import (
	sr "cspirt/internal/events/service"
	"github.com/gin-gonic/gin"
	"cspirt/internal/storage"
	"cspirt/internal/logger"
	"strconv"
	"net/http"
)

func GetEventsHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
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

func GetEventParamsHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
		params, err := eventService.GetEventParams(eventID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event params"})
			return
		}

		ctx.JSON(200, params)
	}
}

func GetEventPlayersHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
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

func GetEventPlayersCountHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		eventService := sr.NewEventsService(s, s.Secret)
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