package handlers

import (
	"cspirt/internal/storage"
	"cspirt/internal/events/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetEventsHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		userIdStr := ctx.Query("user_id")
		if userIdStr != "" {
			userID, err := strconv.Atoi(userIdStr)
			if err != nil {
				ctx.JSON(400, gin.H{"error": "Invalid user ID"})
				return
			}
			events, err := s.GetEventsByUserID(userID)
			if err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to get events"})
				return
			}
			ctx.JSON(200, events)
			return
		}

		events, err := s.GetEvents()
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to get events"})
			return
		}

		ctx.JSON(200, events)
	}
}

func AddEventHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var event models.Event
		if err := ctx.BindJSON(&event); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := s.AddEvent(event); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to add event"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Event added successfully"})
	}
}

func DeleteEventHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		if err := s.DeleteEvent(eventID); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete event"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Event deleted successfully"})
	}
}

func AddPlayersToEvent(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			PlayerIDs []int `json:"playerIds"`
		}

		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := s.AddPlayersToEvent(eventID, req.PlayerIDs); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to add players to event"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Players added to event successfully"})
	}
}

func DeletePlayersFromEvent(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			PlayerIDs []int `json:"playerIds"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := s.DeletePlayersFromEvent(eventID, req.PlayerIDs); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete players from event"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Players deleted from event successfully"})
	}
}