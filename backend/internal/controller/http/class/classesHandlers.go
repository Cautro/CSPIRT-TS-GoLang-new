package handlers

import (
	classModels "cspirt/internal/domain/class"
	sr "cspirt/internal/usecase/class"
	"cspirt/pkg/logger"
	permissionService "cspirt/internal/controller/permission/usecase"
	userModels "cspirt/internal/domain/user"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetClassTeachersHandler returns the list of class teachers.
// @Summary Get class teachers
// @Description Returns all class teachers.
// @Tags classes
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/classes/teacher [get]
func GetClassTeachersHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		teachers, err := ClassUsecase.GetAllClassTeachers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class teachers"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Teachers": teachers})
	}
}

// AddParallelClassHandler creates a parallel class from a grade range or explicit class IDs.
// @Summary Create parallel class
// @Description Creates a parallel class from the provided payload.
// @Tags classes
// @Accept json
// @Produce json
// @Param request body classModels.AddParallelRequest true "Parallel class payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/parallel/add [patch]
func AddParallelClassHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input classModels.AddParallelRequest

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		classesIDs := input.ClassIDs

		if input.MinGrade > 0 && input.MaxGrade > 0 {
			ids, err := ClassUsecase.GetClassIDsByRange(input.MinGrade, input.MaxGrade)
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
			err := ClassUsecase.AddParallelByGradeRange(input.Name, input.MinGrade, input.MaxGrade)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if err := ClassUsecase.AddParallelClass(input.Name, classesIDs, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Parallel created"})
	}
}

// UpdateClassHandler updates an existing class.
// @Summary Update class
// @Description Updates a class by its ID.
// @Tags classes
// @Accept json
// @Produce json
// @Param class_id path int true "Class ID"
// @Param request body classModels.ClassInput true "Updated class payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/{class_id}/update [patch]
func UpdateClassHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
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

		if err := ClassUsecase.UpdateClass(classId, input, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Class updated"})
	}
}

// GetParallelClassesHandler returns all parallel classes.
// @Summary List parallel classes
// @Description Returns a list of parallel classes.
// @Tags classes
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/classes/parallel [get]
func GetParallelClassesHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClasses, err := ClassUsecase.GetParallelClass()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve parallel classes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ParallelClasses": parallelClasses})
	}
}

// DeleteParallelClassHandler deletes a parallel class by ID.
// @Summary Delete parallel class
// @Description Deletes a parallel class by query parameter.
// @Tags classes
// @Produce json
// @Param parallel_class_id query int true "Parallel class ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/parallel/delete [delete]
func DeleteParallelClassHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Query("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		if err := ClassUsecase.DeleteParallelClass(parallelClassId, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete parallel class"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Parallel class deleted"})
	}
}

// GetParallelClassByIDHandler returns a single parallel class by its ID.
// @Summary Get parallel class by ID
// @Description Returns the parallel class with the specified ID.
// @Tags classes
// @Produce json
// @Param parallel_class_id path int true "Parallel class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/parallel/{parallel_class_id} [get]
func GetParallelClassByIDHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		parallelClasses, err := ClassUsecase.GetParallelClass()
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

// GetParallelClassUsersHandler returns all users belonging to a parallel class.
// @Summary Get users of parallel class
// @Description Returns all users from classes included in the specified parallel class.
// @Tags classes
// @Produce json
// @Param parallel_class_id path int true "Parallel class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/parallel/{parallel_class_id}/users [get]
func GetParallelClassUsersHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		parallelClasses, err := ClassUsecase.GetParallelClass()
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
			users, err := ClassUsecase.GetUsersByClassID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users for class ID " + strconv.Itoa(classID)})
				return
			}
			allUsers = append(allUsers, users...)
		}

		c.JSON(http.StatusOK, gin.H{"Users": allUsers})
	}
}

// CompleteQuarterHandler completes quarter results for a parallel class.
// @Summary Complete quarter
// @Description Completes the quarter for the specified parallel class.
// @Tags classes
// @Produce json
// @Param parallel_class_id query int true "Parallel class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/quarter/complete [patch]
func CompleteQuarterHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
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

		login := c.GetString("Login"); if login == "" { c.JSON(400, gin.H{"Error": "Error to take you're login"}); return }
		classes, err := ClassUsecase.CompleteQuarter(parallelClassId, login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete quarter"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Quarter completed", "1st": classes[0], "2nd": classes[1], "3rd": classes[2]})
	}
}

// YearComplete completes the year for all classes.
// @Summary Complete year
// @Description Completes the year for all classes.
// @Tags classes
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/classes/year/complete [patch]
func YearComplete(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var classes []*classModels.Class
		var err error

		login := c.GetString("Login"); if login == "" { c.JSON(400, gin.H{"Error": "Error to take you're login"}); return }
		if classes, err = ClassUsecase.YearComplete(login); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete year"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Year completed", "Classes": classes})
	}
}

// GetBestClassInParallelHandler returns the best class in a parallel class.
// @Summary Get best class in parallel
// @Description Returns the best class from the specified parallel class.
// @Tags classes
// @Produce json
// @Param parallel_class_id path int true "Parallel class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/parallel/{parallel_class_id}/best [get]
func GetBestClassInParallelHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		bestClass, err := ClassUsecase.GetBestClassInParallel(parallelClassId)
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

// GetClassesInParallelHandler returns the list of classes inside a parallel class.
// @Summary Get classes in parallel
// @Description Returns classes belonging to the given parallel class.
// @Tags classes
// @Produce json
// @Param parallel_class_id path int true "Parallel class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/parallel/{parallel_class_id}/classes [get]
func GetClassesInParallelHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		parallelClassIdStr := c.Param("parallel_class_id")
		parallelClassId, err := strconv.Atoi(parallelClassIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Parallel Class ID format"})
			return
		}

		classes, err := ClassUsecase.GetClassInParallel(parallelClassId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve classes in parallel"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Classes": classes})
	}
}

// AddClassHandler creates a new class.
// @Summary Create class
// @Description Creates a new class from the provided payload.
// @Tags classes
// @Accept json
// @Produce json
// @Param request body classModels.ClassInput true "Class payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/add [patch]
func AddClassHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
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

		if err := ClassUsecase.AddClass(input, c.GetString("Login")); err != nil {
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

// DeleteClassHandler deletes a class by ID.
// @Summary Delete class
// @Description Deletes a class by its ID.
// @Tags classes
// @Produce json
// @Param id path int true "Class ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/delete/{id} [delete]
func DeleteClassHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		classIdStr := c.Param("id")
		classId, err := strconv.Atoi(classIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		if err := ClassUsecase.DeleteClass(classId, c.GetString("Login")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete class"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Class deleted"})
	}
}

// GetClassesHandler returns classes, optionally filtered by class_id.
// @Summary List classes
// @Description Returns classes, optionally filtered by the given class_id query parameter.
// @Tags classes
// @Produce json
// @Param class_id query int false "Class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes [get]
func GetClassesHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		classes, err := ClassUsecase.GetAllClasses()
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

// GetClassUsersHandler returns all users in a class.
// @Summary Get class users
// @Description Returns all users belonging to the specified class.
// @Tags classes
// @Produce json
// @Param class_id path int true "Class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/{class_id}/users [get]
func GetClassUsersHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		classIdStr := c.Param("class_id")
		classId, err := strconv.Atoi(classIdStr)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		class, err := ClassUsecase.GetClassByID(classId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
			return
		}
		if class == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		users, err := ClassUsecase.GetUsersByClassID(classId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Users": users})
	}
}

// GetClassTeacherHandler returns the teacher of a class.
// @Summary Get class teacher
// @Description Returns the teacher assigned to the specified class.
// @Tags classes
// @Produce json
// @Param class_id path int true "Class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/classes/{class_id}/teacher [get]
func GetClassTeacherHandler(ClassUsecase *sr.ClassUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		classIdStr := c.Param("class_id")
		classId, err := strconv.Atoi(classIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Class ID format"})
			return
		}

		class, err := ClassUsecase.GetClassByID(classId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
			return
		}
		if class == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		teacher, err := ClassUsecase.GetClassTeacher(classId)
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

// SetClassTeacherHandler assigns a teacher to a class.
// @Summary Set class teacher
// @Description Assigns a teacher login to the specified class.
// @Tags classes
// @Accept json
// @Produce json
// @Param class_id path int true "Class ID"
// @Param request body classModels.ClassTeacherInput true "Teacher assignment payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/classes/{class_id}/teacher [patch]
func SetClassTeacherHandler(ClassUsecase *sr.ClassUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "set_class_teacher")
		if !ok {
			return
		}
		if !permissionService.CanManageClasses(user.Role) {
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

		if err := ClassUsecase.SetClassTeacher(classId, input.TeacherLogin); err != nil {
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
