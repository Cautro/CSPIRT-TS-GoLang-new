package handlers

import (
	"cspirt/internal/logger"
	"cspirt/internal/models"
	sr "cspirt/internal/service/classes"
	"cspirt/internal/storage"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClassesHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		classService := sr.NewClassService(s, s.Secret)
		classes, err := classService.GetAllClasses()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve classes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Classes": classes})
	}
}

func GetClassUsersHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticatedUser(c, s, "get_class_users")
		if !ok {
			return
		}

		classIdStr := c.Param("class_id")
		classId, err := strconv.Atoi(classIdStr)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		class, err := classService.GetClassByID(classId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
			return
		}
		if class == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		if !canReadClass(user, classId) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this class"})
			return
		}

		users, err := classService.GetUsersByClassID(classId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Users": users})
	}
}

func GetClassTeacherHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticatedUser(c, s, "get_class_teacher")
		if !ok {
			return
		}

		classIdStr := c.Param("class_id")
		classId, err := strconv.Atoi(classIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		class, err := classService.GetClassByID(classId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
			return
		}
		if class == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		if !canReadClass(user, classId) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this class"})
			return
		}

		teacher, err := classService.GetClassTeacher(classId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class teacher"})
			return
		}
		if teacher == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class teacher not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Teacher": teacher})
	}
}

func SetClassTeacherHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authenticatedUser(c, s, "set_class_teacher")
		if !ok {
			return
		}
		if !canManageClasses(user.Role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
			return
		}

		classIdStr := c.Param("class_id")
		classId, err := strconv.Atoi(classIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		var input models.ClassTeacherInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		if err := classService.SetClassTeacher(classId, input.TeacherLogin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  "set_class_teacher",
			Login:   user.Login,
			Role:    user.Role,
			Class:   user.Class,
			Message: "class teacher updated",
		})

		c.JSON(http.StatusOK, gin.H{"message": "Class teacher updated"})
	}
}

func authenticatedUser(c *gin.Context, s *storage.Storage, action string) (*models.User, bool) {
	login := c.GetString("Login")
	if login == "" {
		writeLog(logger.LogEntry{
			Level:   "info",
			Action:  action,
			Message: "invalid login or token",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}

	user, err := s.GetUserByLogin(login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return nil, false
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return nil, false
	}

	return user, true
}

func canManageClasses(role string) bool {
	return strings.EqualFold(role, string(models.RoleAdmin)) ||
		strings.EqualFold(role, string(models.RoleOwner))
}

func canReadClass(user *models.User, classID int) bool {
	if canManageClasses(user.Role) {
		return true
	}

	return user.ClassID == classID
}
