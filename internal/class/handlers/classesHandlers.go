package handlers

import (
	classModels "cspirt/internal/class/models"
	sr "cspirt/internal/class/service"
	"cspirt/internal/logger"
	ratingModels "cspirt/internal/rating/models"
	"cspirt/internal/storage"
	userModels "cspirt/internal/users/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClassTeachersHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		classService := sr.NewClassService(s, s.Secret)
		teachers, err := classService.GetAllClassTeachers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class teachers"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Teachers": teachers})
	}
}

func AddClassHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input classModels.ClassInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)

		if err := classService.AddClass(input, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(), // временно для отладки
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Class added"})
	}
}

func DeleteClassHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		classIdStr := c.Param("id")
		classId, err := strconv.Atoi(classIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		if err := classService.DeleteClass(classId, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete class"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Class deleted"})
	}
}

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

		var input classModels.ClassTeacherInput
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

func authenticatedUser(c *gin.Context, s *storage.Storage, action string) (*userModels.User, bool) {
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
	return strings.EqualFold(role, string(ratingModels.RoleAdmin)) ||
		strings.EqualFold(role, string(ratingModels.RoleOwner))
}

func canReadClass(user *userModels.User, classID int) bool {
	if canManageClasses(user.Role) {
		return true
	}

	return user.ClassID == classID
}
