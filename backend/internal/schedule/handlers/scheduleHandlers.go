package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"cspirt/internal/logger"
	ratingModels "cspirt/internal/rating/models"
	scheduleModels "cspirt/internal/schedule/models"
	scheduleService "cspirt/internal/schedule/service"
	"cspirt/internal/storage"
	"cspirt/internal/utils"
)

func GetSchedulesHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := utils.AuthenticatedUser(c, s, "get_schedules")
		if !ok {
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
		if classID <= 0 && !utils.CanManageClasses(user.Role) {
			classID = user.ClassID
		}
		if classID > 0 {
			class, err := s.GetClassByID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
				return
			}
			if class == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
				return
			}
			if !utils.CanReadClass(user, classID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this class"})
				return
			}
		}

		service := scheduleService.NewScheduleService(s, s.Secret)
		result, err := service.GetSchedules(scheduleModels.ScheduleFilter{
			ClassID:  classID,
			Day:      c.Query("day"),
			Date:     c.Query("date"),
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

func UpdateSchedulesHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := utils.AuthenticatedUser(c, s, "update_schedules")
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

		classID, err := affectedScheduleClassID(s, input)
		if err != nil {
			status := http.StatusBadRequest
			if strings.Contains(strings.ToLower(err.Error()), "failed") {
				status = http.StatusInternalServerError
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		if !utils.CanManageClasses(user.Role) {
			if classID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Class ID is required"})
				return
			}
			if user.ClassID != classID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this class"})
				return
			}
		}

		service := scheduleService.NewScheduleService(s, s.Secret)
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

func canManageSchedules(role string) bool {
	return strings.EqualFold(role, string(ratingModels.RoleHelper)) ||
		strings.EqualFold(role, string(ratingModels.RoleAdmin)) ||
		strings.EqualFold(role, string(ratingModels.RoleOwner))
}

func affectedScheduleClassID(s *storage.Storage, input scheduleModels.UpdateSchedulesInput) (int, error) {
	target := strings.ToLower(strings.TrimSpace(input.Target))
	if target == "" {
		if input.Exception != nil {
			target = scheduleModels.ScheduleTargetException
		} else if input.Planned != nil {
			target = scheduleModels.ScheduleTargetPlanned
		} else {
			target = scheduleModels.ScheduleTargetBase
		}
	}

	switch target {
	case scheduleModels.ScheduleTargetBase:
		return affectedBaseScheduleClassID(s, input)
	case scheduleModels.ScheduleTargetException:
		return affectedExceptionClassID(s, input)
	case scheduleModels.ScheduleTargetPlanned:
		return affectedPlannedClassID(s, input)
	default:
		return 0, errors.New("invalid schedule target")
	}
}

func affectedBaseScheduleClassID(s *storage.Storage, input scheduleModels.UpdateSchedulesInput) (int, error) {
	if input.Schedule != nil && input.Schedule.ClassID > 0 {
		return input.Schedule.ClassID, nil
	}

	id := input.ID
	if id <= 0 && input.Schedule != nil {
		id = input.Schedule.ID
	}
	if id <= 0 {
		return 0, nil
	}

	schedule, err := s.GetBaseScheduleByID(id)
	if err != nil {
		return 0, errors.New("failed to retrieve schedule")
	}
	if schedule == nil {
		return 0, errors.New("schedule not found")
	}

	return schedule.ClassID, nil
}

func affectedExceptionClassID(s *storage.Storage, input scheduleModels.UpdateSchedulesInput) (int, error) {
	if input.Exception != nil {
		if input.Exception.ClassID > 0 {
			return input.Exception.ClassID, nil
		}
		if input.Exception.ScheduleID != nil {
			schedule, err := s.GetBaseScheduleByID(*input.Exception.ScheduleID)
			if err != nil {
				return 0, errors.New("failed to retrieve base schedule")
			}
			if schedule == nil {
				return 0, errors.New("base schedule not found")
			}
			return schedule.ClassID, nil
		}
	}

	id := input.ID
	if id <= 0 && input.Exception != nil {
		id = input.Exception.ID
	}
	if id <= 0 {
		return 0, nil
	}

	exception, err := s.GetScheduleExceptionByID(id)
	if err != nil {
		return 0, errors.New("failed to retrieve schedule exception")
	}
	if exception == nil {
		return 0, errors.New("schedule exception not found")
	}

	return exception.ClassID, nil
}

func affectedPlannedClassID(s *storage.Storage, input scheduleModels.UpdateSchedulesInput) (int, error) {
	if input.Planned != nil && input.Planned.ClassID > 0 {
		return input.Planned.ClassID, nil
	}
	if input.Planned != nil && input.Planned.BaseScheduleID != nil {
		schedule, err := s.GetBaseScheduleByID(*input.Planned.BaseScheduleID)
		if err != nil {
			return 0, errors.New("failed to retrieve base schedule")
		}
		if schedule == nil {
			return 0, errors.New("base schedule not found")
		}
		return schedule.ClassID, nil
	}

	id := input.ID
	if id <= 0 && input.Planned != nil {
		id = input.Planned.ID
	}
	if id <= 0 {
		return 0, nil
	}

	planned, err := s.GetPlannedScheduleByID(id)
	if err != nil {
		return 0, errors.New("failed to retrieve planned schedule")
	}
	if planned == nil {
		return 0, errors.New("planned schedule not found")
	}

	return planned.ClassID, nil
}
