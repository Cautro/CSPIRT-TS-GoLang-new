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
	"errors"

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

func AddParallelClassHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input classModels.AddParallelRequest

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		classesIDs := input.ClassIDs

		if input.MinGrade > 0 && input.MaxGrade > 0 {
			ids, err := s.GetClassIDsByRange(input.MinGrade, input.MaxGrade)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classes by range"})
				return
			}
			classesIDs = ids
		}

		if input.Name == "" {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Name is required",
            })
            return
        }
        
        hasGradeRange := input.MinGrade > 0 && input.MaxGrade > 0
        hasClassIDs := len(input.ClassIDs) > 0
        
        if !hasGradeRange && !hasClassIDs {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "ClassIDs or grade range required",
            })
            return
        }

		if input.MinGrade != 0 {
            err := s.AddParallelByGradeRange(input.Name, input.MinGrade, input.MaxGrade)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
        }
		
		classService := sr.NewClassService(s, s.Secret) 
		if err := classService.AddParallelClass(input.Name, classesIDs, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Parallel created"})
	}
}

func UpdateClassHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		classIdStr := c.Param("class_id")
		classId, err := strconv.Atoi(classIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		var input classModels.ClassInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := checkInputClass(input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "update_class",
			Login:   c.GetString("Login"),
			Message: "Update class input: " + input.Name + ", " + input.TeacherLogin,
		})

		classService := sr.NewClassService(s, s.Secret)

		if err := classService.UpdateClass(classId, input, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Class updated"})
	}
}

func GetParallelClassesHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		classService := sr.NewClassService(s, s.Secret)
		parallelClasses, err := classService.GetParallelClasses()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve parallel classes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ParallelClasses": parallelClasses})
	}
}

func DeleteParallelClassHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Query("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		if err := classService.DeleteParallelClass(parallelClassId, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete parallel class"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Parallel class deleted"})
	}
}

func GetParallelClassByIDHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		parallelClasses, err := classService.GetParallelClasses()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve parallel classes"})
			return
		}

		var parallelClass *classModels.ParallelClass
		for _, pc := range parallelClasses {
			if pc.ID == parallelClassId {
				parallelClass = &pc
				break
			}
		}

		if parallelClass == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parallel class not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ParallelClass": parallelClass})
	}
}

func GetParallelClassUsersHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		parallelClasses, err := classService.GetParallelClasses()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve parallel classes"})
			return
		}

		var parallelClass *classModels.ParallelClass
		for _, pc := range parallelClasses {
			if pc.ID == parallelClassId {
				parallelClass = &pc
				break
			}
		}

		if parallelClass == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parallel class not found"})
			return
		}

		var allUsers []userModels.SafeUser
		for _, classID := range parallelClass.ClassesIDs {
			users, err := classService.GetUsersByClassID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users for class ID " + strconv.Itoa(classID)})
				return
			}
			allUsers = append(allUsers, users...)
		}

		c.JSON(http.StatusOK, gin.H{"Users": allUsers})
	}
}

func CompleteQuarterHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Query("parallel_class_id")
		if parallelClassIdStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parallel_class_id"})
			return
		}
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		classes, err := classService.CompleteQuarter(parallelClassId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete quarter"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Quarter completed", "1st": classes[0], "2nd": classes[1], "3rd": classes[2]})
	}
}

func GetBestClassInParallelHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		bestClass, err := classService.GetBestClassInParallel(parallelClassId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve best class in parallel"})
			return
		}
		if bestClass == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Best class in parallel not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"BestClass": bestClass})
	}
}

func GetClassesInParallelHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		classService := sr.NewClassService(s, s.Secret)
		classes, err := classService.GetClassesInParallel(parallelClassId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve classes in parallel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Classes": classes})
	}
}

func AddClassHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input classModels.ClassInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := checkInputClass(input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Action:  "add_class",
			Login:   c.GetString("Login"),
			Message: "Add class input: " + input.Name + ", " + input.TeacherLogin,
		})

		classService := sr.NewClassService(s, s.Secret)

		if err := classService.AddClass(input, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Class added"})
	}
}

func checkInputClass(input classModels.ClassInput) error {
	if input.Name == "" {
		return errors.New("class name is required")
	}
	if input.TeacherLogin == "" {
		return errors.New("class teacher login is required")
	}
	return nil
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

		classId := c.Query("class_id")
		if classId != "" {
			classIdInt, err := strconv.Atoi(classId)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
				return
			}

			var filteredClasses []classModels.Class
			for _, class := range classes {
				if class.ID == classIdInt {
					filteredClasses = append(filteredClasses, class)
					break
				}
			}
			classes = filteredClasses
		}

		if len(classes) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No classes found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Classes": classes})
	}
}

func GetClassUsersHandler(s *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		logger.WriteSafe(logger.LogEntry{
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
		logger.WriteSafe(logger.LogEntry{
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