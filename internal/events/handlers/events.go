package handlers

import (
	"cspirt/internal/events/models"
	"cspirt/internal/storage"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetEventsHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		userIdStr := ctx.Query("user_id")
		classIdStr := ctx.Query("class_id")
		if classIdStr == "" {
			classIdStr = ctx.Query("class")
		}

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
		if classIdStr != "" {
			classID, err := strconv.Atoi(classIdStr)
			if err != nil {
				ctx.JSON(400, gin.H{"error": "Invalid class ID"})
				return
			}
			events, err := s.GetEventsByClassID(classID)
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

func GetEventPlayersHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		players, err := s.GetEventPlayers(eventID)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to get event players"})
			return
		}

		ctx.JSON(200, players)
	}
}

func GetEventPlayersCountHandler(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		count, err := s.GetEventPlayersCount(eventID)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to get event players count"})
			return
		}

		ctx.JSON(200, gin.H{"count": count})
	}
}

func EventComplete(s *storage.Storage) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var req struct {
			RatingReward int `json:"ratingReward"`
		}
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := s.EventComplete(eventID, req.RatingReward); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to complete event"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Event completed successfully"})
	}
}
