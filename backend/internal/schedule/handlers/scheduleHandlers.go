package handlers

import (
	"cspirt/internal/events/models"
	sr "cspirt/internal/events/service"
	"cspirt/internal/logger"
	"cspirt/internal/storage"
	"strconv"
	"cspirt/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
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

func AddPlayersToEvent(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := s.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := utils.CheckUserRole(s, ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_players",
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

		if err := eventService.AddPlayersToEvent(eventID, req.PlayerIDs); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to add players to event"})
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

		if err := eventService.EventComplete(eventID, req.RatingReward); err != nil {
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
