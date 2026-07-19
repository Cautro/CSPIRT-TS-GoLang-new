package router

import (
	// handlers
	authHandlers "cspirt/internal/controller/http/auth"
	healthHandlers "cspirt/internal/controller/http/checker"
	classHandlers "cspirt/internal/controller/http/class"
	complaintHandlers "cspirt/internal/controller/http/complaint"
	eventHandlers "cspirt/internal/controller/http/event"
	"cspirt/internal/controller/http/middleware-JWT"
	noteHandlers "cspirt/internal/controller/http/note"
	ratingHandlers "cspirt/internal/controller/http/rating"
	scheduleHandlers "cspirt/internal/controller/http/schedule"
	userHandlers "cspirt/internal/controller/http/user"

	// usecase
	permissionUsecase "cspirt/internal/controller/permission/usecase"
	cacheRepo "cspirt/internal/domain/cache/repo"
	authUsecase "cspirt/internal/usecase/auth"
	classUsecase "cspirt/internal/usecase/class"
	complaintUsecase "cspirt/internal/usecase/complaint"
	eventsUsecase "cspirt/internal/usecase/event"
	noteUsecase "cspirt/internal/usecase/note"
	ratingUsecase "cspirt/internal/usecase/rating"
	scheduleUsecase "cspirt/internal/usecase/schedule"
	usersUsecase "cspirt/internal/usecase/user"

	"database/sql"
	"os"

	"github.com/gin-gonic/gin"
)

type Usecases struct {
	Auth       *authUsecase.AuthUsecase
	Users      *usersUsecase.UsersUsecase
	Class      *classUsecase.ClassUsecase
	Note       *noteUsecase.NoteUsecase
	Complaint  *complaintUsecase.ComplaintUsecase
	Events     *eventsUsecase.EventsUsecase
	Rating     *ratingUsecase.RatingsUsecase
	Schedule   *scheduleUsecase.ScheduleUsecase
	Permission *permissionUsecase.Usecase
	Cache      cacheRepo.CacheRepository
	JWTSecret  string
	DB         *sql.DB
}

func NewRouter(s Usecases) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	if os.Getenv("PROFILE") == "1" {
		router.Use(DiagnosticsMiddleware(s.DB))
	}

	registerPublicRoutes(router, s)
	registerAuthenticatedRoutes(router, s)

	return router
}

func registerPublicRoutes(router *gin.Engine, s Usecases) {
	router.GET("/health", healthHandlers.HealthHandler)
	router.POST("/login", authHandlers.LoginHandler(s.Auth))
	router.POST("/api/refresh", authHandlers.RefreshHandler(s.Auth))
}

func registerAuthenticatedRoutes(router *gin.Engine, s Usecases) {
	auth := router.Group("/api", utils.AuthMiddleware(s.JWTSecret, s.Cache))

	registerUserRoutes(auth, s)
	registerClassRoutes(auth, s)
	registerRatingRoutes(auth, s)
	registerNoteRoutes(auth, s)
	registerComplaintRoutes(auth, s)
	registerEventRoutes(auth, s)
	registerScheduleRoutes(auth, s)
	registerNotificationRoutes(auth, s)
}

func registerNotificationRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.POST("device/register", authHandlers.RegisterDeviceHandler(s.Users))
	auth.POST("device/unregister", authHandlers.UnregisterDeviceHandler(s.Users))
}

func registerUserRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.GET("/users", userHandlers.GetUsersHandler(s.Users))
	auth.GET("/users/get/staff", userHandlers.GetStaffHandler(s.Users))
	auth.PATCH("/user/add", userHandlers.AddUserHandler(s.Users))
	auth.DELETE("/user/delete/:id", userHandlers.DeleteUserHandler(s.Users))
	auth.GET("/me", userHandlers.GetMeHandler(s.Users))
	auth.PATCH("/user/logout", userHandlers.LogoutHandler(s.Users, s.Auth))
	auth.PATCH("/user/update", userHandlers.UpdateUserHandler(s.Users, s.Permission))
	auth.PATCH("/user/update/avatar", userHandlers.UpdateAvatarHandler(s.Users))
}

func registerClassRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.GET("/classes", classHandlers.GetClassesHandler(s.Class))
	auth.GET("/classes/teacher", classHandlers.GetClassTeachersHandler(s.Class))
	auth.PATCH("/classes/add", classHandlers.AddClassHandler(s.Class))
	auth.DELETE("/classes/delete/:id", classHandlers.DeleteClassHandler(s.Class))
	auth.GET("/classes/:class_id/users", classHandlers.GetClassUsersHandler(s.Class))
	auth.GET("/classes/:class_id/teacher", classHandlers.GetClassTeacherHandler(s.Class))
	auth.PATCH("/classes/:class_id/teacher", classHandlers.SetClassTeacherHandler(s.Class, s.Permission))
	auth.PATCH("/classes/parallel/add", classHandlers.AddParallelClassHandler(s.Class))
	auth.GET("/classes/parallel", classHandlers.GetParallelClassesHandler(s.Class))
	auth.DELETE("/classes/parallel/delete", classHandlers.DeleteParallelClassHandler(s.Class))
	auth.GET("/classes/parallel/:parallel_class_id", classHandlers.GetParallelClassByIDHandler(s.Class))
	auth.GET("/classes/parallel/:parallel_class_id/users", classHandlers.GetParallelClassUsersHandler(s.Class))
	auth.PATCH("/classes/quarter/complete", classHandlers.CompleteQuarterHandler(s.Class))
	auth.GET("/classes/parallel/:parallel_class_id/best", classHandlers.GetBestClassInParallelHandler(s.Class))
	auth.GET("/classes/parallel/:parallel_class_id/classes", classHandlers.GetClassesInParallelHandler(s.Class))
	auth.PATCH("/classes/:class_id/update", classHandlers.UpdateClassHandler(s.Class))
	auth.PATCH("/classes/year/complete", classHandlers.YearComplete(s.Class))
}

func registerRatingRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.GET("/rating", ratingHandlers.GetRatingsHandler(s.Users))
	auth.PATCH("/rating/update", ratingHandlers.UpdateRatingsHandler(s.Rating, s.Users))
}

func registerNoteRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.GET("/notes", noteHandlers.GetNotesHandler(s.Note, s.Class, s.Permission))
	auth.PATCH("/note/add", noteHandlers.AddNoteHandler(s.Note, s.Users, s.Permission))
	auth.DELETE("/note/delete/:id", noteHandlers.DeleteNoteHandler(s.Note, s.Users))
}

func registerComplaintRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.GET("/complaints", complaintHandlers.GetComplaintsHandler(s.Complaint, s.Class, s.Permission))
	auth.PATCH("/complaint/add", complaintHandlers.AddcomplaintHandler(s.Complaint, s.Users, s.Permission))
	auth.DELETE("/complaint/delete/:id", complaintHandlers.DeletecomplaintHandler(s.Complaint, s.Users, s.Permission))
}

func registerEventRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.GET("/events", eventHandlers.GetEventsHandler(s.Events))
	auth.PATCH("/event/add", eventHandlers.AddEventHandler(s.Events, s.Permission))
	auth.DELETE("/event/delete/:id", eventHandlers.DeleteEventHandler(s.Events, s.Permission))
	auth.PATCH("/event/:eventId/players/add", eventHandlers.AddPlayersToEvent(s.Events, s.Users, s.Permission))
	auth.DELETE("/event/:eventId/players/delete", eventHandlers.DeletePlayersFromEvent(s.Events, s.Permission))
	auth.GET("/event/:eventId/players", eventHandlers.GetEventPlayersHandler(s.Events))
	auth.GET("/event/:eventId/players/count", eventHandlers.GetEventPlayersCountHandler(s.Events))
	auth.PATCH("/event/:eventId/complete", eventHandlers.EventComplete(s.Events, s.Permission))
	auth.PATCH("/event/:eventId/params/add", eventHandlers.AddEventParams(s.Events, s.Permission))
	auth.GET("/event/:eventId/params", eventHandlers.GetEventParamsHandler(s.Events))
	auth.DELETE("/event/:eventId/params/delete", eventHandlers.DeleteEventParamsHandler(s.Events, s.Permission))
	auth.PATCH("/event/:eventId/params/update", eventHandlers.UpdateEventParamsHandler(s.Events, s.Permission))
	auth.PATCH("/event/:eventId/update", eventHandlers.UpdateEventHandler(s.Events, s.Permission))
}

func registerScheduleRoutes(auth *gin.RouterGroup, s Usecases) {
	auth.GET("/schedules/teacher/current", scheduleHandlers.GetTeacherCurrentScheduleHandler(s.Schedule, s.Permission))
	auth.PATCH("/schedules/rollover", scheduleHandlers.RolloverSchedulesHandler(s.Schedule, s.Permission))
	auth.PATCH("/schedules/planned/reset", scheduleHandlers.ResetPlannedSchedulesHandler(s.Schedule, s.Permission))
	auth.GET("/schedules", scheduleHandlers.GetSchedulesHandler(s.Schedule, s.Class, s.Permission))
	auth.PATCH("/schedules/update", scheduleHandlers.UpdateSchedulesHandler(s.Schedule, s.Permission))
}
