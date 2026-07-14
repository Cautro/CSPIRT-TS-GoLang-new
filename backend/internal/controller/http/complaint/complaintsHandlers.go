package handlers

import (
	srClass "cspirt/internal/usecase/class"
	complaintModels "cspirt/internal/domain/complaint"
	complaintsservice "cspirt/internal/usecase/complaint"
	"cspirt/pkg/logger"
	permissionService "cspirt/internal/controller/permission/usecase"
	ratingModels "cspirt/internal/domain/rating"
	usersvc "cspirt/internal/usecase/user"
	"cspirt/internal/utils"
	"errors"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetComplaintsHandler returns complaints visible to the current user.
// @Summary List complaints
// @Description Returns complaints for a class or for the current user depending on permissions.
// @Tags complaints
// @Produce json
// @Param class query int false "Class ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/complaints [get]
func GetComplaintsHandler(complaintService *complaintsservice.ComplaintUsecase, classService *srClass.ClassUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := perm.AuthenticatedUser(c, "get_notes")
		if !ok {
			return
		}

		err := perm.CheckUserRole(user.Login, string(ratingModels.RoleAdmin), string(ratingModels.RoleOwner), string(ratingModels.RoleHelper))
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

			result, err := complaintService.GetComplaintsByClassID(classID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"Complaints": result})
			return
		}

		if !permissionService.CanManageClasses(user.Role) {
			if user.ClassID <= 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "User has no class"})
				return
			}

			result, err := complaintService.GetComplaintsByClassID(user.ClassID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"Complaints": result})
			return
		}

		result, err := complaintService.GetAllComplaints()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"All_Complaints": result})
	}
}

// AddcomplaintHandler creates a complaint.
// @Summary Create complaint
// @Description Creates a new complaint from the request body.
// @Tags complaints
// @Accept json
// @Produce json
// @Param request body complaintModels.AddNewComplaintResponse true "Complaint payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/complaint/add [patch]
func AddcomplaintHandler(complaintService *complaintsservice.ComplaintUsecase, users *usersvc.UsersUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		if login == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		check := perm.CheckPublicRole(login)
		if check != nil {
			if errors.Is(check, permissionService.ErrAccessDenied) || errors.Is(check, permissionService.ErrUserNotFound) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
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

		var in complaintModels.AddNewComplaintResponse
		if err := c.ShouldBindJSON(&in); err != nil {
			logger.WriteSafe(logger.LogEntry{
				Level:   "info",
				Action:  "notes",
				Message: "invalid input: " + err.Error(),
			})
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		needUser := utils.UserToSafeUser(*user)
		if err := complaintService.AddNewComplaint(login, &in, needUser); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logger.WriteSafe(logger.LogEntry{
			Level:   "info",
			Login:   login,
			Class:   user.Class,
			Role:    user.Role,
			Action:  "added_complaint",
			Message: "Added new complaint",
		})
		c.JSON(200, gin.H{"message": "Complaint added"})
	}
}

// DeletecomplaintHandler deletes a complaint by ID.
// @Summary Delete complaint
// @Description Deletes the complaint with the provided ID.
// @Tags complaints
// @Produce json
// @Param id path int true "Complaint ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/complaint/delete/{id} [delete]
func DeletecomplaintHandler(complaintService *complaintsservice.ComplaintUsecase, users *usersvc.UsersUsecase, perm *permissionService.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.Param("id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		check := perm.CheckPublicRole(login)
		if check != nil {
			if errors.Is(check, permissionService.ErrAccessDenied) || errors.Is(check, permissionService.ErrUserNotFound) {
				c.JSON(http.StatusForbidden, gin.H{"error": "You dont have permissions for this action"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
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

		needUser := utils.UserToSafeUser(*foundUser)
		if err := complaintService.DeleteComplaint(idInt, *needUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Complaint deleted"})
	}
}
