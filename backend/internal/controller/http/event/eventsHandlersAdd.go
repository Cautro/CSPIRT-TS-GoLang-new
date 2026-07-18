package handlers

import (
	models "cspirt/internal/domain/event"
	sr "cspirt/internal/usecase/event"
	permissionService "cspirt/internal/controller/permission/usecase"
	"cspirt/pkg/logger"
	ratingModels "cspirt/internal/domain/rating"
	usersvc "cspirt/internal/usecase/user"
	"strconv"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// AddEventHandler creates a new event.
// @Summary Create event
// @Description Creates a new event from the request body.
// @Tags events
// @Accept json
// @Produce json
// @Param request body models.Event true "Event payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/add [patch]
func AddEventHandler(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
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
				Action:  "add_event",
				Login:   c.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		var event models.Event
		if err := c.BindJSON(&event); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event",
				Login:   c.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.AddEvent(ctx, event); err != nil {
			c.JSON(500, gin.H{"error": "Failed to add event"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_event",
			Login:   c.GetString("Login"),
			Message: "event added successfully",
		})
		c.JSON(200, gin.H{"message": "Event added successfully"})
	}
}

// AddEventParams adds params to an existing event.
// @Summary Add event params
// @Description Adds parameters to the specified event.
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path int true "Event ID"
// @Param request body models.EventParams true "Event params payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/params/add [patch]
func AddEventParams(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
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
				Action:  "add_event_params",
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
				Action:  "add_event_params",
				Login:   c.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var params models.EventParams
		if err := c.BindJSON(&params); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_params",
				Login:   c.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.AddEventParams(ctx, eventID, &params); err != nil {
			c.JSON(500, gin.H{"error": "Failed to add event params"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_event_params",
			Login:   c.GetString("Login"),
			Message: "event params added successfully",
		})

		c.JSON(200, gin.H{"message": "Event params added successfully"})
	}
}

// AddPlayersToEvent adds players to an event.
// @Summary Add players to event
// @Description Adds one or more player IDs to the specified event.
// @Tags events
// @Accept json
// @Produce json
// @Param eventId path int true "Event ID"
// @Param request body object{playerIds=[]int} true "Player IDs payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/event/{eventId}/players/add [patch]
func AddPlayersToEvent(eventService *sr.EventsUsecase, users *usersvc.UsersUsecase, perm *permissionService.Usecase) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := eventService.ActivateDueEvents(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		login := c.GetString("Login")
		user, err := users.GetUserByLogin(ctx, login)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_players",
				Login:   login,
				Message: "failed to get user: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Failed to get user"})
			return
		}

		err = perm.CheckUserRole(ctx, c.GetString("Login"), string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_players",
				Login:   c.GetString("Login"),
				Role:    user.Role,
				Message: "failed to check user role: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventID, err := strconv.Atoi(c.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_players",
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
				Action:  "add_event_players",
				Login:   c.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.AddPlayersToEvent(ctx, eventID, req.PlayerIDs, c.GetString("Login")); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_event_players",
			Login:   c.GetString("Login"),
			Message: "players added to event successfully",
		})
		c.JSON(200, gin.H{"message": "Players added to event successfully"})
	}
}
