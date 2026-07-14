package handlers

import (
	models "cspirt/internal/domain/event"
	sr "cspirt/internal/usecase/event"
	permissionService "cspirt/internal/controller/permission/usecase"
	"cspirt/pkg/logger"
	ratingModels "cspirt/internal/domain/rating"
	usersvc "cspirt/internal/usecase/user"
	"strconv"

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
func AddEventHandler(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := perm.CheckUserRole(ctx.GetString("Login"), "owner")
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
func AddEventParams(eventService *sr.EventsUsecase, perm *permissionService.Usecase) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		err := perm.CheckUserRole(ctx.GetString("Login"), "owner")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_params",
				Login:   ctx.GetString("Login"),
				Message: "failed to check user role: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

		eventID, err := strconv.Atoi(ctx.Param("eventId"))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_params",
				Login:   ctx.GetString("Login"),
				Message: "invalid event id: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid event ID"})
			return
		}

		var params models.EventParams
		if err := ctx.BindJSON(&params); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "add_event_params",
				Login:   ctx.GetString("Login"),
				Message: "invalid request body: " + err.Error(),
			})
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if err := eventService.AddEventParams(eventID, &params); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to add event params"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_event_params",
			Login:   ctx.GetString("Login"),
			Message: "event params added successfully",
		})

		ctx.JSON(200, gin.H{"message": "Event params added successfully"})
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
func AddPlayersToEvent(eventService *sr.EventsUsecase, users *usersvc.UsersUsecase, perm *permissionService.Usecase) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if err := eventService.ActivateDueEvents(); err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to activate due events"})
			return
		}

		login := ctx.GetString("Login")
		user, err := users.GetUserByLogin(login)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_players",
				Login:   login,
				Message: "failed to get user: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to get user"})
			return
		}

		err = perm.CheckUserRole(ctx.GetString("Login"), string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner))
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "error",
				Action:  "add_event_players",
				Login:   ctx.GetString("Login"),
				Role:    user.Role,
				Message: "failed to check user role: " + err.Error(),
			})
			ctx.JSON(500, gin.H{"error": "Failed to check user role"})
			return
		}

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

		if err := eventService.AddPlayersToEvent(eventID, req.PlayerIDs, ctx.GetString("Login")); err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
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
