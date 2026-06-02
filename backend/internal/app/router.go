package app

import (
	authHandlers "cspirt/internal/auth/handlers"
	classHandlers "cspirt/internal/class/handlers"
	complaintHandlers "cspirt/internal/complaints/handlers"
	eventHandlers "cspirt/internal/events/handlers"
	healthHandlers "cspirt/internal/health/handlers"
	noteHandlers "cspirt/internal/note/handlers"
	ratingHandlers "cspirt/internal/rating/handler"
	ratingService "cspirt/internal/rating/service"
	scheduleHandlers "cspirt/internal/schedule/handlers"
	"cspirt/internal/storage"
	userHandlers "cspirt/internal/users/handlers"
	"cspirt/internal/utils"

	"github.com/gin-gonic/gin"
)

func NewRouter(s *storage.Storage, jwtSecret string) *gin.Engine {
	router := gin.Default()

	registerPublicRoutes(router, s)
	registerAuthenticatedRoutes(router, s, jwtSecret)

	return router
}

func registerPublicRoutes(router *gin.Engine, s *storage.Storage) {
	router.GET("/health", healthHandlers.HealthHandler)
	router.POST("/login", authHandlers.LoginHandler(s))
	router.POST("/api/refresh", authHandlers.RefreshHandler(s))
}

func registerAuthenticatedRoutes(router *gin.Engine, s *storage.Storage, jwtSecret string) {
	auth := router.Group("/api", utils.AuthMiddleware(jwtSecret))

	registerUserRoutes(auth, s)
	registerClassRoutes(auth, s)
	registerRatingRoutes(auth, s)
	registerNoteRoutes(auth, s)
	registerComplaintRoutes(auth, s)
	registerEventRoutes(auth, s)
	registerScheduleRoutes(auth, s)
}

func registerUserRoutes(auth *gin.RouterGroup, s *storage.Storage) {
	auth.GET("/users", userHandlers.GetUsersHandler(s))
	auth.GET("/users/get/staff", userHandlers.GetStaffHandler(s))
	auth.PATCH("/user/add", userHandlers.AddUserHandler(s))
	auth.DELETE("/user/delete/:id", userHandlers.DeleteUserHandler(s))
	auth.GET("/me", userHandlers.GetMeHandler(s))
	auth.PATCH("/user/logout", userHandlers.LogoutHandler(s))
	auth.PATCH("/user/update", userHandlers.UpdateUserHandler(s))
}

func registerClassRoutes(auth *gin.RouterGroup, s *storage.Storage) {
	auth.GET("/classes", classHandlers.GetClassesHandler(s))
	auth.GET("/classes/teacher", classHandlers.GetClassTeachersHandler(s))
	auth.PATCH("/classes/add", classHandlers.AddClassHandler(s))
	auth.DELETE("/classes/delete/:id", classHandlers.DeleteClassHandler(s))
	auth.GET("/classes/:class_id/users", classHandlers.GetClassUsersHandler(s))
	auth.GET("/classes/:class_id/teacher", classHandlers.GetClassTeacherHandler(s))
	auth.PATCH("/classes/:class_id/teacher", classHandlers.SetClassTeacherHandler(s))
	auth.PATCH("/classes/parallel/add", classHandlers.AddParallelClassHandler(s))
	auth.GET("/classes/parallel", classHandlers.GetParallelClassesHandler(s))
	auth.DELETE("/classes/parallel/delete", classHandlers.DeleteParallelClassHandler(s))
	auth.GET("/classes/parallel/:parallel_class_id", classHandlers.GetParallelClassByIDHandler(s))
	auth.GET("/classes/parallel/:parallel_class_id/users", classHandlers.GetParallelClassUsersHandler(s))
	auth.PATCH("/classes/quarter/complete", classHandlers.CompleteQuarterHandler(s))
	auth.GET("/classes/parallel/:parallel_class_id/best", classHandlers.GetBestClassInParallelHandler(s))
	auth.GET("/classes/parallel/:parallel_class_id/classes", classHandlers.GetClassesInParallelHandler(s))
	auth.PATCH("/classes/:class_id/update", classHandlers.UpdateClassHandler(s))
	auth.PATCH("/classes/year/complete", classHandlers.YearComplete(s))
}

func registerRatingRoutes(auth *gin.RouterGroup, s *storage.Storage) {
	ratings := ratingService.NewRatingsService(s.RatingRepo, s, s.Secret)

	auth.GET("/rating", ratingHandlers.GetRatingsHandler(s))
	auth.PATCH("/rating/update", ratingHandlers.UpdateRatingsHandler(ratings, s))
}

func registerNoteRoutes(auth *gin.RouterGroup, s *storage.Storage) {
	auth.GET("/notes", noteHandlers.GetNotesHandler(s))
	auth.PATCH("/note/add", noteHandlers.AddNoteHandler(s))
	auth.DELETE("/note/delete/:id", noteHandlers.DeleteNoteHandler(s))
}

func registerComplaintRoutes(auth *gin.RouterGroup, s *storage.Storage) {
	auth.GET("/complaints", complaintHandlers.GetComplaintsHandler(s))
	auth.PATCH("/complaint/add", complaintHandlers.AddcomplaintHandler(s))
	auth.DELETE("/complaint/delete/:id", complaintHandlers.DeletecomplaintHandler(s))
}

func registerEventRoutes(auth *gin.RouterGroup, s *storage.Storage) {
	auth.GET("/events", eventHandlers.GetEventsHandler(s))
	auth.PATCH("/event/add", eventHandlers.AddEventHandler(s))
	auth.DELETE("/event/delete/:id", eventHandlers.DeleteEventHandler(s))
	auth.PATCH("/event/:eventId/players/add", eventHandlers.AddPlayersToEvent(s))
	auth.DELETE("/event/:eventId/players/delete", eventHandlers.DeletePlayersFromEvent(s))
	auth.GET("/event/:eventId/players", eventHandlers.GetEventPlayersHandler(s))
	auth.GET("/event/:eventId/players/count", eventHandlers.GetEventPlayersCountHandler(s))
	auth.PATCH("/event/:eventId/complete", eventHandlers.EventComplete(s))
	auth.PATCH("/event/:eventId/params/add", eventHandlers.AddEventParams(s))
	auth.GET("/event/:eventId/params", eventHandlers.GetEventParamsHandler(s))
	auth.DELETE("/event/:eventId/params/delete", eventHandlers.DeleteEventParamsHandler(s))
	auth.PATCH("/event/:eventId/params/update", eventHandlers.UpdateEventParamsHandler(s))
	auth.PATCH("/event/:eventId/update", eventHandlers.UpdateEventHandler(s))
}

func registerScheduleRoutes(auth *gin.RouterGroup, s *storage.Storage) {
	auth.GET("/schedules/teacher/current", scheduleHandlers.GetTeacherCurrentScheduleHandler(s))
	auth.PATCH("/schedules/rollover", scheduleHandlers.RolloverSchedulesHandler(s))
	auth.PATCH("/schedules/planned/reset", scheduleHandlers.ResetPlannedSchedulesHandler(s))
	auth.GET("/schedules", scheduleHandlers.GetSchedulesHandler(s))
	auth.PATCH("/schedules/update", scheduleHandlers.UpdateSchedulesHandler(s))
}