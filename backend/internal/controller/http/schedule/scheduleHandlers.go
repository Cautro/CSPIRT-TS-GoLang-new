package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	classService "cspirt/internal/usecase/class"
	"cspirt/pkg/logger"
	permissionService "cspirt/internal/controller/permission/usecase"
	ratingModels "cspirt/internal/domain/rating"
	scheduleModels "cspirt/internal/domain/schedule"
	ScheduleUsecase "cspirt/internal/usecase/schedule"
)

// GetSchedulesHandler returns schedule data for the current user or a requested class.
// @Summary List schedules
// @Description Returns schedules filtered by type, class, day, and week type.
// @Tags schedules
// @Produce json
// @Param type query string false "Schedule type"
// @Param class_id query int false "Class ID"
// @Param day query string false "Day"
// @Param week_type query string false "Week type"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/schedules [get]
func GetSchedulesHandler(service *ScheduleUsecase.ScheduleUsecase, classes *classService.ClassUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "get_schedules")
		if !ok {
			return
		}

		scheduleType, err := scheduleTypeFromQuery(c, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule type"})
			return
		}

		classID, err := optionalIntQuery(c, "class_id", "classID", "class")
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "get_schedules",
				Login:   user.Login,
				Message: "invalid class id: " + err.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
			return
		}

		if classID <= 0 && !isTeacherScheduleViewer(user.Role) {
			classID = user.ClassID
		}
		if scheduleType == scheduleModels.ScheduleTypeAll && classID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Class ID is required for all schedule types"})
			return
		}
		if classID > 0 {
			class, err := classes.GetClassByID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
				return
			}
			if class == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
				return
			}
		}

		if status, message := authorizeScheduleRead(user.Role, user.ClassID, classID, scheduleType); status != http.StatusOK {
			c.JSON(status, gin.H{"error": message})
			return
		}

		result, err := service.GetSchedules(scheduleModels.ScheduleFilter{
			Type:     scheduleType,
			ClassID:  classID,
			Day:      c.Query("day"),
			WeekType: c.DefaultQuery("week_type", "all"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve schedules"})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "get_schedules",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "schedules retrieved",
		})
		c.JSON(http.StatusOK, result)
	}
}

// GetTeacherCurrentScheduleHandler returns the current schedule for a teacher.
// @Summary Get teacher current schedule
// @Description Returns the current schedule for the authenticated teacher or a specified teacher ID.
// @Tags schedules
// @Produce json
// @Param teacher_id query int false "Teacher ID"
// @Param day query string false "Day"
// @Param week_type query string false "Week type"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/schedules/teacher/current [get]
func GetTeacherCurrentScheduleHandler(service *ScheduleUsecase.ScheduleUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "get_teacher_current_schedule")
		if !ok {
			return
		}
		if !isTeacherScheduleViewer(user.Role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
			return
		}

		teacherID, err := optionalIntQuery(c, "teacher_id", "teacherID", "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid teacher ID"})
			return
		}
		if teacherID <= 0 {
			teacherID = user.ID
		}
		if !isOwner(user.Role) && teacherID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Teachers can retrieve only their own current schedule"})
			return
		}

		lessons, err := service.GetCurrentScheduleForTeacher(teacherID, scheduleModels.ScheduleFilter{
			Day:      c.Query("day"),
			WeekType: c.DefaultQuery("week_type", "all"),
		})
		if err != nil {
			status := http.StatusBadRequest
			if strings.Contains(strings.ToLower(err.Error()), "failed") {
				status = http.StatusInternalServerError
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Schedules": lessons})
	}
}

// UpdateSchedulesHandler updates one or more schedule entries.
// @Summary Update schedules
// @Description Updates schedules using the request body payload.
// @Tags schedules
// @Accept json
// @Produce json
// @Param request body scheduleModels.UpdateSchedulesInput true "Schedule update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/schedules/update [patch]
func UpdateSchedulesHandler(service *ScheduleUsecase.ScheduleUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "update_schedules")
		if !ok {
			return
		}
		if !canManageSchedules(user.Role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
			return
		}

		var input scheduleModels.UpdateSchedulesInput
		if err := c.ShouldBindJSON(&input); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_schedules",
				Login:   user.Login,
				Role:    user.Role,
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		result, err := service.UpdateSchedules(input)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "update_schedules",
				Login:   user.Login,
				Role:    user.Role,
				Class:   user.Class,
				Message: "failed to update schedules: " + err.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_schedules",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "schedules updated",
		})
		c.JSON(http.StatusOK, result)
	}
}

// RolloverSchedulesHandler rolls over schedules for a class.
// @Summary Rollover schedules
// @Description Rolls over schedules for the specified class.
// @Tags schedules
// @Produce json
// @Param class_id query int false "Class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/schedules/rollover [patch]
func RolloverSchedulesHandler(service *ScheduleUsecase.ScheduleUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "rollover_schedules")
		if !ok {
			return
		}
		if !canManageSchedules(user.Role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
			return
		}

		classID, err := optionalIntQuery(c, "class_id", "classID", "class")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
			return
		}

		result, err := service.RolloverSchedules(classID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// ResetPlannedSchedulesHandler resets planned schedules for a class.
// @Summary Reset planned schedules
// @Description Resets planned schedules for the specified class.
// @Tags schedules
// @Produce json
// @Param class_id query int false "Class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/schedules/planned/reset [patch]
func ResetPlannedSchedulesHandler(service *ScheduleUsecase.ScheduleUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "reset_planned_schedules")
		if !ok {
			return
		}
		if !canManageSchedules(user.Role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
			return
		}

		classID, err := optionalIntQuery(c, "class_id", "classID", "class")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
			return
		}

		result, err := service.ResetPlannedSchedules(classID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func optionalIntQuery(c *gin.Context, names ...string) (int, error) {
	for _, name := range names {
		value := strings.TrimSpace(c.Query(name))
		if value == "" {
			continue
		}

		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return 0, errors.New("invalid integer")
		}
		return parsed, nil
	}

	return 0, nil
}

func scheduleTypeFromQuery(c *gin.Context, allowAll bool) (string, error) {
	value := strings.ToLower(strings.TrimSpace(c.Query("type")))
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(c.Query("target")))
	}
	if value == "" {
		value = scheduleModels.ScheduleTypeCurrent
	}

	switch value {
	case scheduleModels.ScheduleTypeBase,
		scheduleModels.ScheduleTypeCurrent,
		scheduleModels.ScheduleTypePlanned:
		return value, nil
	case scheduleModels.ScheduleTypeAll:
		if allowAll {
			return value, nil
		}
	}

	return "", errors.New("invalid schedule type")
}

func authorizeScheduleRead(role string, userClassID int, classID int, scheduleType string) (int, string) {
	if isOwner(role) {
		return http.StatusOK, ""
	}

	if scheduleType == scheduleModels.ScheduleTypeBase || scheduleType == scheduleModels.ScheduleTypeAll {
		return http.StatusForbidden, "Only owner can view base schedule"
	}

	if isTeacherScheduleViewer(role) {
		return http.StatusOK, ""
	}

	if scheduleType != scheduleModels.ScheduleTypeCurrent {
		return http.StatusForbidden, "Students can view only current schedule"
	}

	return http.StatusOK, ""
}

func canManageSchedules(role string) bool {
	return isOwner(role)
}

func isOwner(role string) bool {
	return strings.EqualFold(role, string(ratingModels.RoleOwner))
}

func isTeacherScheduleViewer(role string) bool {
	return strings.EqualFold(role, string(ratingModels.RoleHelper)) ||
		strings.EqualFold(role, string(ratingModels.RoleAdmin)) ||
		strings.EqualFold(role, string(ratingModels.RoleOwner))
}
