package modarete

import (
	perm "cspirt/internal/controller/permission/usecase"
	entity "cspirt/internal/domain/moderate"
	usecase "cspirt/internal/usecase/moderate"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllWaitNotes(usecase usecase.ModerateUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		output, err := usecase.GetAllWaitNotes(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return
		}

		c.JSON(200, output)
	}
}

func GetAllWaitComplaints(usecase usecase.ModerateUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		output, err := usecase.GetAllWaitComplaints(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return
		}

		c.JSON(200, output)
	}
}

func UpdateNoteModerateWait(u usecase.ModerateUsecase, p perm.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.GetString("Id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return
		}
		
		var input entity.ModerateDTO
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		if err := u.UpdateNoteModerateWait(c.Request.Context(), login, id, input, p); err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}
	}
}

func UpdateComplaintModerateWait(u usecase.ModerateUsecase, p perm.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		login := c.GetString("Login")
		idStr := c.GetString("Id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return
		}
		
		var input entity.ModerateDTO
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}

		if err := u.UpdateComplaintModerateWait(c.Request.Context(), login, id, input, p); err != nil {
			c.JSON(500, gin.H{"error":"Server error"})
			return 
		}
	}
}