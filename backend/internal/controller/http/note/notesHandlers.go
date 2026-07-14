package handlers

import (
	srClass "cspirt/internal/usecase/class"
	"cspirt/pkg/logger"
	noteModels "cspirt/internal/domain/note"
	sr "cspirt/internal/usecase/note"
	permissionService "cspirt/internal/controller/permission/usecase"
	ratingModels "cspirt/internal/domain/rating"
	models "cspirt/internal/domain/user"
	usersvc "cspirt/internal/usecase/user"
	u "cspirt/internal/controller/utils"
	"errors"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetNotesHandler returns notes visible to the current user.
// @Summary List notes
// @Description Returns notes for the requested class or for the current user's class.
// @Tags notes
// @Produce json
// @Param class query int false "Class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notes [get]
func GetNotesHandler(noteService *sr.NoteUsecase, classService *srClass.ClassUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "get_notes")
		if !ok {
			return
		}

		err := perm.CheckUserRole(
			user.Login,
			string(ratingModels.RoleAdmin),
			string(ratingModels.RoleOwner),
			string(ratingModels.RoleHelper),
		)
		if err != nil {
			if errors.Is(err, permissionService.ErrAccessDenied) || errors.Is(err, permissionService.ErrUserNotFound) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		classIDStr := c.Query("class")
		if classIDStr != "" {
			classID, err := strconv.Atoi(classIDStr)
			if err != nil || classID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
				return
			}

			class, err := classService.GetClassByID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve class"})
				return
			}
			if class == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
				return
			}

			if !permissionService.CanReadClass(user, classID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this class"})
				return
			}

			result, err := noteService.GetNotesByClassID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"Notes": result})
			return
		}

		if !permissionService.CanManageClasses(user.Role) {
			if user.ClassID <= 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "User has no class"})
				return
			}

			result, err := noteService.GetNotesByClassID(user.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"Notes": result})
			return
		}

		result, err := noteService.GetAllNotes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"All_notes": result})
	}
}

// AddNoteHandler creates a new note.
// @Summary Create note
// @Description Creates a new note from the request body.
// @Tags notes
// @Accept json
// @Produce json
// @Param request body noteModels.AddNewNoteResponse true "Note payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/note/add [patch]
func AddNoteHandler(noteService *sr.NoteUsecase, users *usersvc.UsersUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		err := perm.CheckUserRole(login, string(ratingModels.RoleHelper), string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner))
		if err != nil {
			c.JSON(500, gin.H{"error": "You dont have permisions for that action"})
			return
		}

		user, err := users.GetUserByLogin(login)
		if err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "server error: " + err.Error(),
			})
			c.JSON(500, gin.H{"error": "Server error"})
			return
		}
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var in noteModels.AddNewNoteResponse
		if err := c.ShouldBindJSON(&in); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		needUser := &models.SafeUser{
			ID:       user.ID,
			Name:     user.Name,
			LastName: user.LastName,
			FullName: user.FullName,
			Login:    user.Login,
			Rating:   user.Rating,
			Role:     user.Role,
			Class:    user.Class,
			ClassID:  user.ClassID,
		}

		if err := noteService.AddNewNote(login, &in, needUser); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Login:   login,
			Class:   user.Class,
			Role:    user.Role,
			Action:  "added_note",
			Message: "Added new note",
		})
		c.JSON(200, gin.H{"message": "Note added"})
	}
}

// DeleteNoteHandler deletes a note by ID.
// @Summary Delete note
// @Description Deletes the note with the provided ID.
// @Tags notes
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/note/delete/{id} [delete]
func DeleteNoteHandler(noteService *sr.NoteUsecase, users *usersvc.UsersUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.Param("id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		foundUser, err := users.GetUserByLogin(login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}

		if foundUser == nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_note",
				Login:   login,
				Message: "note not found",
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "Server error"})
			return
		}

		if !permissionService.CanManageClasses(foundUser.Role) {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "delete_user",
				Login:   login,
				Role:    foundUser.Role,
				Class:   foundUser.Class,
				Message: "User without need roles trying to delete user",
			})
			c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
			return
		}

		needUser := u.UserToSafeUser(*foundUser)
		if err := noteService.DeleteNote(idInt, *needUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Note deleted"})
	}
}
